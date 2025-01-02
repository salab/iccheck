// Package fleccs contains a golang re-implementation of the tool "FLeCCS" - "Fragment Level Similar Co-Change Suggester"
// by Manishankar Mondal, et al., on ICPC 2021.
// https://ieeexplore.ieee.org/document/9463009
//
// Details may differ.
package fleccs

import (
	"bytes"
	"context"
	"github.com/cespare/xxhash"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/salab/iccheck/pkg/utils/files"
	"github.com/salab/iccheck/pkg/utils/strs"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"runtime"
)

type Source struct {
	Filename     string
	StartL, EndL int
}

type Candidate struct {
	Filename   string
	StartLine  int
	EndLine    int
	Similarity float64
	// Source indicates from which query this co-change candidate was detected
	Source Source
}

type Query struct {
	Filename     string
	StartL, EndL int

	enlargedContext    int
	contextStartLine   int
	contextEndLine     int
	contextLineLengths []int
	contextBigrams     []strs.BigramSet
	hash               uint64
}

func (q *Query) toSource() Source {
	return Source{
		Filename: q.Filename,
		StartL:   q.StartL,
		EndL:     q.EndL,
	}
}

func (q *Query) calculateContextLines(c *config, queryTree domain.Searcher) error {
	file, err := queryTree.Open(q.Filename)
	if err != nil {
		return errors.Wrapf(err, "opening file %v", q.Filename)
	}
	queryFileContent, err := file.Content()
	if err != nil {
		return errors.Wrapf(err, "reading file contents %v", q.Filename)
	}
	queryFileStrLines := files.Lines(queryFileContent)
	queryFileLines := ds.Map(queryFileStrLines, func(l string) []byte { return []byte(l) })

	q.enlargedContext = c.contextLines
	q.contextStartLine = max(1, q.StartL-c.contextLines)               // inclusive, 1-indexed
	q.contextEndLine = min(len(queryFileLines), q.EndL+c.contextLines) // inclusive, 1-indexed

	contextLines := queryFileLines[q.contextStartLine-1 : q.contextEndLine]
	q.contextLineLengths = ds.Map(contextLines, func(line []byte) int { return len(line) })
	q.contextBigrams = ds.Map(contextLines, func(line []byte) strs.BigramSet {
		return strs.Bigram(line)
	})

	q.hash = xxhash.Sum64(bytes.Join(contextLines, nil))

	return nil
}

func (q *Query) accountForContextLines(c *Candidate) *Candidate {
	// The detected candidate lines are enlarged due to the context area
	// Shrink the enlarged area to get the true clone region
	contextStartDiff := q.StartL - q.contextStartLine
	contextEndDiff := q.contextEndLine - q.EndL

	c.StartLine += contextStartDiff
	c.EndLine -= contextEndDiff

	return c
}

// disc returns Dice-Sørensen Coefficient.
func disc(bigram1, bigram2 strs.BigramSet) float64 {
	totalSetLen := len(bigram1) + len(bigram2)
	if totalSetLen == 0 {
		return 0
	}
	intersection := strs.BigramIntersectionCount(bigram1, bigram2)
	return 2 * float64(intersection) / float64(totalSetLen)
}

// waDiSC returns Weighted-Average DiSC (Dice-Sørensen Coefficient).
func waDiSC(lengths1, lengths2 []int, bigrams1, bigrams2 []strs.BigramSet) float64 {
	totalLength := lo.Sum(lengths1) + lo.Sum(lengths2)
	if totalLength == 0 {
		return 0
	}

	var ret float64
	for i := range bigrams1 {
		discV := disc(bigrams1[i], bigrams2[i])
		l1 := lengths1[i]
		l2 := lengths2[i]
		weight := float64(l1 + l2)
		ret += discV * weight
	}
	return ret / float64(totalLength)
}

func findCandidates(
	q *Query,
	searchFilename string,
	searchFileLineLengths []int,
	searchFileBigrams []strs.BigramSet,
	ignoreRule *domain.IgnoreLineRule,
	similarityThreshold float64,
) []*Candidate {
	var candidates []*Candidate
	windowSize := len(q.contextBigrams)
	orgWindowSize := windowSize - 2*q.enlargedContext

	if len(searchFileBigrams) < windowSize {
		// If the search target file is shorter than the query lines (including context)
		// TODO: compare once?
	}

	for i := 0; i < len(searchFileBigrams)-windowSize+1; i++ {
		startLine := i            // 0-indexed, inclusive
		endLine := i + windowSize // 0-indexed, exclusive

		// Check if this (original) region can be skipped
		if canSkip, skipUntil := ignoreRule.CanSkip(i+q.enlargedContext, orgWindowSize); canSkip {
			i = skipUntil - q.enlargedContext
			continue
		}

		fileCmpLengths := searchFileLineLengths[startLine:endLine]
		fileCmpLines := searchFileBigrams[startLine:endLine]
		similarity := waDiSC(q.contextLineLengths, fileCmpLengths, q.contextBigrams, fileCmpLines)
		if similarity >= similarityThreshold {
			candidates = append(candidates, &Candidate{
				Filename:   searchFilename,
				StartLine:  startLine + 1, // 1-indexed, inclusive
				EndLine:    endLine,       // 1-indexed, inclusive
				Similarity: similarity,
				Source:     q.toSource(),
			})
			i += windowSize - 1 // Proceed the search window
		}
	}

	return candidates
}

func fileSearch(
	ctx context.Context,
	queries []*Query,
	searchTree domain.Searcher,
	searchFilename string,
	ignore domain.IgnoreRules,
	similarityThreshold float64,
) ([]*Candidate, error) {
	if ctx.Err() != nil { // check for deadline
		return nil, ctx.Err()
	}

	searchFile, err := searchTree.Open(searchFilename)
	if err != nil {
		return nil, errors.Wrapf(err, "opening search target file %v", searchFilename)
	}

	// Skip binary file search because it is rarely needed and consumes cpu
	isBinary, err := isBinaryCache.Get(context.WithValue(ctx, searchFileCtx, searchFile), searchFilename)
	if err != nil {
		return nil, errors.Wrapf(err, "calculating binary status of search target file %v", searchFilename)
	}
	if isBinary {
		return nil, nil
	}

	fileContent, err := searchFile.Content()
	if err != nil {
		return nil, errors.Wrapf(err, "reading search target file %v", searchFilename)
	}

	// Check ignore settings
	skipEntireFile, ignoreRule := ignore.Match(searchFilename, fileContent)
	if skipEntireFile {
		return nil, nil
	}

	fileHash := xxhash.Sum64(fileContent)
	fileStrLines := files.Lines(fileContent)
	fileLines := ds.Map(fileStrLines, func(l string) []byte { return []byte(l) })

	fileLineLengths := ds.Map(fileLines, func(line []byte) int { return len(line) })
	fileLineBigrams := ds.Map(fileLines, func(line []byte) strs.BigramSet {
		return strs.Bigram(line)
	})

	var candidates []*Candidate
	// For each query, extract candidates
	for _, q := range queries {
		if ctx.Err() != nil { // check for deadline
			return nil, ctx.Err()
		}

		qCandidates := getFromCacheOrCalcCandidates(q.hash, fileHash, func() []*Candidate {
			qCandidates := findCandidates(q, searchFilename, fileLineLengths, fileLineBigrams, ignoreRule, similarityThreshold)
			// Fix found candidate lines not to include the enlarged context lines
			return ds.Map(qCandidates, func(c *Candidate) *Candidate { return q.accountForContextLines(c) })
		})
		candidates = append(candidates, qCandidates...)
	}

	return candidates, nil
}

func Search(
	ctx context.Context,
	queriesTree domain.Searcher,
	queries []*Query,
	searchTree domain.Searcher,
	ignore domain.IgnoreRules,
	options ...ConfigFunc,
) ([]*Candidate, error) {
	// Calculate config
	c := applyConfig(options...)

	// Pre-calculate query line bi-grams
	for _, q := range queries {
		err := q.calculateContextLines(c, queriesTree)
		if err != nil {
			return nil, errors.Wrapf(err, "calculating context lines for query %v", q.Filename)
		}
	}

	// List all file names from search root directory
	searchFiles, err := searchTree.Files()
	if err != nil {
		return nil, errors.Wrap(err, "listing search tree files")
	}

	// Search for co-change candidates!
	p := pool.NewWithResults[[]*Candidate]().
		WithMaxGoroutines(runtime.NumCPU()).
		WithErrors().
		WithContext(ctx).
		WithCancelOnError().
		WithFirstError()
	for _, searchFile := range searchFiles {
		p.Go(func(ctx context.Context) ([]*Candidate, error) {
			return fileSearch(ctx, queries, searchTree, searchFile, ignore, c.similarityThreshold)
		})
	}
	candidates, err := p.Wait()
	if err != nil {
		return nil, err
	}

	return lo.Flatten(candidates), nil
}

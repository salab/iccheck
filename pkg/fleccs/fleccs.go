// Package fleccs contains a golang re-implementation of the tool "FLeCCS" - "Fragment Level Similar Co-Change Suggester"
// by Manishankar Mondal, et al., on ICPC 2021.
// https://ieeexplore.ieee.org/document/9463009
//
// Details may differ.
package fleccs

import (
	"bytes"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/salab/iccheck/pkg/utils/files"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/salab/iccheck/pkg/utils/strs"
)

type Candidate struct {
	Filename   string
	StartLine  int
	EndLine    int
	Similarity float64
}

func lines(b []byte) [][]byte {
	return bytes.Split(bytes.Trim(b, "\n"), []byte("\n"))
}

// disc returns Dice-Sørensen Coefficient.
func disc(bigram1, bigram2 strs.Set) float64 {
	intersection := strs.IntersectionCount(bigram1, bigram2)
	return 2 * float64(intersection) / float64(len(bigram1)+len(bigram2))
}

// waDiSC returns Weighted-Average DiSC (Dice-Sørensen Coefficient).
func waDiSC(lengths1, lengths2 []int, bigrams1, bigrams2 []strs.Set) float64 {
	discs := lo.Map(bigrams1, func(bigram1 strs.Set, idx int) float64 {
		return disc(bigram1, bigrams2[idx])
	})
	totalLength := lo.Sum(lengths1) + lo.Sum(lengths2)
	weightedDiscs := lo.Map(discs, func(disc float64, idx int) float64 {
		l1 := lengths1[idx]
		l2 := lengths2[idx]
		weight := float64(l1+l2) / float64(totalLength)
		return disc * weight
	})
	return lo.Sum(weightedDiscs)
}

func findCandidates(
	q *Query,
	searchFilename string,
	searchFileLineLengths []int,
	searchFileBigrams []strs.Set,
	similarityThreshold float64,
) []*Candidate {
	var candidates []*Candidate

	if len(searchFileBigrams) < len(q.contextBigrams) {
		// If the search target file is shorter than the query lines (including context)
		// TODO: compare once?
	}

	for i := 0; i < len(searchFileBigrams)-len(q.contextBigrams)+1; i++ {
		startLine := i                       // 0-indexed, inclusive
		endLine := i + len(q.contextBigrams) // 0-indexed, exclusive

		fileCmpLengths := searchFileLineLengths[startLine:endLine]
		fileCmpLines := searchFileBigrams[startLine:endLine]
		similarity := waDiSC(q.contextLineLengths, fileCmpLengths, q.contextBigrams, fileCmpLines)
		if similarity >= similarityThreshold {
			candidates = append(candidates, &Candidate{
				Filename:   searchFilename,
				StartLine:  startLine + 1, // 1-indexed, inclusive
				EndLine:    endLine,       // 1-indexed, inclusive
				Similarity: similarity,
			})
			i += len(q.contextBigrams) - 1 // Proceed the search window
		}
	}

	return candidates
}

func fileSearch(queries []*Query, searchRoot, searchFilename string, similarityThreshold float64) []*Candidate {
	fileBytes := lo.Must(os.ReadFile(filepath.Join(searchRoot, searchFilename)))
	fileLines := lines(fileBytes)
	fileLineLengths := ds.Map(fileLines, func(line []byte) int { return len(line) })
	fileLineBigrams := ds.Map(fileLines, func(line []byte) strs.Set {
		return strs.NGram(2, line)
	})

	var candidates []*Candidate
	// For each query, extract candidates
	for _, q := range queries {
		qCandidates := findCandidates(q, searchFilename, fileLineLengths, fileLineBigrams, similarityThreshold)
		// Fix found candidate lines not to include the enlarged context lines
		qCandidates = ds.Map(qCandidates, func(c *Candidate) *Candidate { return q.accountForContextLines(c) })
		candidates = append(candidates, qCandidates...)
	}

	return candidates
}

func FileTreeDistance(path1, path2 string) int {
	dirs1 := strings.Split(path1, string(os.PathSeparator))
	dirs2 := strings.Split(path2, string(os.PathSeparator))

	matchingLeadingPaths := 0
	for i := 0; i < min(len(dirs1), len(dirs2)); i++ {
		if dirs1[i] == dirs2[i] {
			matchingLeadingPaths++
		} else {
			break
		}
	}

	path1Dist := len(dirs1) - matchingLeadingPaths
	path2Dist := len(dirs2) - matchingLeadingPaths
	return path1Dist + path2Dist
}

type Query struct {
	Filename     string
	StartL, EndL int

	contextStartLine   int
	contextEndLine     int
	contextLineLengths []int
	contextBigrams     []strs.Set
}

func (q *Query) calculateContextLines(c *config, basePath string) {
	queryFullPath := lo.Must(os.ReadFile(filepath.Join(basePath, q.Filename)))
	queryFileLines := lines(queryFullPath)

	q.contextStartLine = max(1, q.StartL-c.contextLines)               // inclusive, 1-indexed
	q.contextEndLine = min(len(queryFileLines), q.EndL+c.contextLines) // inclusive, 1-indexed

	contextLines := queryFileLines[q.contextStartLine-1 : q.contextEndLine]
	q.contextLineLengths = ds.Map(contextLines, func(line []byte) int { return len(line) })
	q.contextBigrams = ds.Map(contextLines, func(line []byte) strs.Set {
		return strs.NGram(2, line)
	})
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

func Search(
	queryBasePath string,
	queries []*Query,
	searchRoot string,
	options ...ConfigFunc,
) []*Candidate {
	// Calculate config
	c := applyConfig(options...)

	// Pre-calculate query line bi-grams
	for _, q := range queries {
		q.calculateContextLines(c, queryBasePath)
	}

	// List all file names from search root directory
	searchFiles := files.WalkAllFilenames(searchRoot)

	// Search for co-change candidates!
	p := pool.NewWithResults[[]*Candidate]().WithMaxGoroutines(runtime.NumCPU())
	for _, filename := range searchFiles {
		p.Go(func() []*Candidate {
			return fileSearch(queries, searchRoot, filename, c.similarityThreshold)
		})
	}
	var candidates []*Candidate
	for _, fileCandidates := range p.Wait() {
		candidates = append(candidates, fileCandidates...)
	}

	return candidates
}

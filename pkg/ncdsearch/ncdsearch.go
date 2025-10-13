// Package ncdsearch contains a basic golang re-implementation of the tool "NCDSearch"
// by Takashi Ishio, et al.
// https://github.com/takashi-ishio/NCDSearch
//
// Details may differ.
package ncdsearch

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"slices"
	"sort"

	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"

	"github.com/salab/iccheck/pkg/utils/files"
	"github.com/salab/iccheck/pkg/utils/strs"
)

func overlap(n int, q, f []byte) float64 {
	nGramQ := strs.NGram(n, q)
	nGramF := strs.NGram(n, f)
	intersection := strs.IntersectionCount(nGramQ, nGramF)
	return float64(intersection) / float64(len(nGramQ))
}

type lzSet = strs.Set

func extractLZSet(b []byte) lzSet {
	set := make(lzSet)
	start, end := 0, 1
	for end <= len(b) {
		bs := string(b[start:end])
		if _, ok := set[bs]; !ok {
			set[bs] = struct{}{}
			start = end
		}
		end++
	}
	return set
}

func compareLZJD(b []byte, pos []int, queryLZSet lzSet) (kBest int, distance float64) {
	s := make(map[string]struct{})
	start, end := 0, 1
	intersection := 0
	distance = math.MaxFloat64

	for k := 0; k < len(pos); k++ {
		for end <= pos[k] {
			bs := string(b[start:end])
			if _, ok := s[bs]; !ok {
				s[bs] = struct{}{}
				start = end
				if _, ok := queryLZSet[bs]; ok {
					intersection++
				}
			}
			end++
		}

		lzjd := 1.0 - float64(intersection)/float64(len(queryLZSet)+len(s)-intersection)
		if lzjd < distance {
			kBest = k
			distance = lzjd
		}
	}

	return
}

func tokenIndices(s []byte, tokenize TokenizeFunc) []int {
	tokens := tokenize(s)
	indices := make([]int, len(tokens)+1)
	// make cumulative sum
	for i := 1; i < len(indices); i++ {
		indices[i] = indices[i-1] + len(tokens[i-1])
	}
	return indices
}

type Clone struct {
	Filename  string
	StartLine int
	EndLine   int
	KBest     int
	Distance  float64
}

func codeSearch(
	q []byte,
	windowSize int,
	threshold float64,
	file []byte,
	filename string,
	tokenize TokenizeFunc,
	ignoreRule *domain.IgnoreLineRule,
) (clones []*Clone) {
	lzSetQ := extractLZSet(q)

	tokenIndices := tokenIndices(file, tokenize)
	lineIndices := files.LineStartIndices(file)
	getLine := func(pos int) int {
		found := sort.SearchInts(lineIndices, pos)
		if found < len(lineIndices) && lineIndices[found] == pos {
			return found + 1
		}
		return found
	}

	end := len(tokenIndices) - windowSize
	for tokenIdx := 0; tokenIdx < end; tokenIdx++ {
		tokenStartIdx := tokenIndices[tokenIdx]
		tokenEndIdx := tokenIndices[tokenIdx+windowSize]

		b := file[tokenStartIdx:tokenEndIdx]
		pos := make([]int, windowSize)
		for i := 0; i < len(pos); i++ {
			pos[i] = tokenIndices[tokenIdx+i] - tokenStartIdx
		}

		kBest, distance := compareLZJD(b, pos, lzSetQ)
		if distance < threshold {
			startLine := getLine(tokenStartIdx)
			if canSkip, _ := ignoreRule.CanSkip(startLine, windowSize); canSkip {
				continue
			}
			clones = append(clones, &Clone{
				Filename:  filename,
				StartLine: startLine,
				EndLine:   getLine(tokenEndIdx), // TODO: get accurate end line
				KBest:     kBest,
				Distance:  distance,
			})
		}
	}

	return
}

func fileSearch(
	ctx context.Context,
	c *config,
	queries []*Query,
	searchTree domain.Searcher,
	searchFilename string,
	matcher *domain.MatcherRules,
) ([]*Clone, error) {
	if ctx.Err() != nil { // check for deadline
		return nil, ctx.Err()
	}

	searchFile, err := searchTree.Open(searchFilename)
	if err != nil {
		return nil, errors.Wrapf(err, "opening search target file %v", searchFilename)
	}

	// Skip binary file search because it is rarely needed and consumes cpu
	isBinary, err := searchFile.IsBinary()
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
	skipEntireFile, ignoreRule := matcher.Match(searchFilename, fileContent)
	if skipEntireFile {
		return nil, nil
	}

	var clones []*Clone
	for _, q := range queries {
		if ctx.Err() != nil { // check for deadline
			return nil, ctx.Err()
		}

		filterSim := overlap(c.overlapNGram, q.contents, fileContent)
		if filterSim < c.filterThreshold {
			return nil, nil
		}
		result := codeSearch(
			q.contents,
			q.windowSize,
			c.searchThreshold,
			fileContent,
			searchFilename,
			c.tokenize,
			ignoreRule,
		)
		clones = append(clones, result...)
	}

	return clones, nil
}

type Query struct {
	Filename     string
	StartL, EndL int

	contents   []byte
	lzSet      lzSet
	windowSize int
}

func (q *Query) readContents(c *config, queryTree domain.Searcher) error {
	f, err := queryTree.Open(q.Filename)
	if err != nil {
		return err
	}
	fileContents, err := f.Content()
	if err != nil {
		return err
	}
	lineIndices := files.LineStartIndices(fileContents)
	if len(lineIndices) < q.EndL {
		return fmt.Errorf("unexpected too short file %v", q.Filename)
	}

	// StartL and EndL are 1-indexed
	if len(lineIndices) == q.EndL {
		q.contents = fileContents[lineIndices[q.StartL-1]:]
	} else {
		q.contents = fileContents[lineIndices[q.StartL-1]:lineIndices[q.EndL-1]]
	}

	q.lzSet = extractLZSet(q.contents)
	q.windowSize = int(math.Floor(c.windowSizeMult * float64(len(c.tokenize(q.contents)))))

	return nil
}

func Search(
	ctx context.Context,
	queriesTree domain.Searcher,
	queries []*Query,
	searchTree domain.Searcher,
	matcher *domain.MatcherRules,
	options ...ConfigFunc,
) ([]*Clone, error) {
	c := applyConfig(options...)

	// Read actual bytes of the queries and calculate lz set
	for _, q := range queries {
		err := q.readContents(c, queriesTree)
		if err != nil {
			return nil, errors.Wrapf(err, "reading contents of query")
		}
	}

	// List all file names from search root directory
	searchFiles, err := searchTree.Files()
	if err != nil {
		return nil, errors.Wrap(err, "listing search tree files")
	}

	p := pool.NewWithResults[[]*Clone]().
		WithMaxGoroutines(runtime.NumCPU()).
		WithErrors().
		WithContext(ctx).
		WithCancelOnError().
		WithFirstError()
	for _, searchFile := range searchFiles {
		p.Go(func(ctx context.Context) ([]*Clone, error) {
			return fileSearch(ctx, c, queries, searchTree, searchFile, matcher)
		})
	}
	results, err := p.Wait()
	if err != nil {
		return nil, err
	}
	clones := lo.Flatten(results)

	slices.SortFunc(clones, ds.SortDesc(func(c *Clone) float64 { return c.Distance }))
	return clones, nil
}

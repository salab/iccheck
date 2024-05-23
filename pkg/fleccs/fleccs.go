// Package fleccs contains a golang re-implementation of the tool "FLeCCS" - "Fragment Level Similar Co-Change Suggester"
// by Manishankar Mondal, et al., on ICPC 2021.
// https://ieeexplore.ieee.org/document/9463009
//
// Details may differ.
package fleccs

import (
	"bytes"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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
func disc(line1, line2 []byte) float64 {
	bigrams1 := strs.NGram(2, line1)
	bigrams2 := strs.NGram(2, line2)
	intersection := strs.IntersectionCount(bigrams1, bigrams2)
	return 2 * float64(intersection) / float64(len(bigrams1)+len(bigrams2))
}

// waDiSC returns Weighted-Average DiSC (Dice-Sørensen Coefficient).
func waDiSC(lines1, lines2 [][]byte) float64 {
	discs := lo.Map(lines1, func(line1 []byte, idx int) float64 {
		return disc(line1, lines2[idx])
	})
	totalLength :=
		lo.SumBy(lines1, func(l []byte) int { return len(l) }) +
			lo.SumBy(lines2, func(l []byte) int { return len(l) })
	weightedDiscs := lo.Map(discs, func(disc float64, idx int) float64 {
		l1 := len(lines1[idx])
		l2 := len(lines2[idx])
		weight := float64(l1+l2) / float64(totalLength)
		return disc * weight
	})
	return lo.Sum(weightedDiscs)
}

func fileSearch(contextLines [][]byte, searchRoot, searchFilename string, similarityThreshold float64) []*Candidate {
	fileBytes := lo.Must(os.ReadFile(filepath.Join(searchRoot, searchFilename)))
	fileLines := lines(fileBytes)

	var candidates []*Candidate
	if len(fileLines) < len(contextLines) {
		// TODO: compare once?
	}

	for i := 0; i < len(fileLines)-len(contextLines)+1; i++ {
		startLine := i                   // 0-indexed, inclusive
		endLine := i + len(contextLines) // 0-indexed, exclusive

		fileCmpLines := fileLines[startLine:endLine]
		similarity := waDiSC(contextLines, fileCmpLines)
		if similarity >= similarityThreshold {
			candidates = append(candidates, &Candidate{
				Filename:   searchFilename,
				StartLine:  startLine + 1, // 1-indexed, inclusive
				EndLine:    endLine,       // 1-indexed, inclusive
				Similarity: similarity,
			})
			i += len(contextLines) - 1
		}
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

func Search(
	basePath, queryFilename string,
	queryStartLine, queryEndLine int,
	searchRoot string,
	options ...ConfigFunc,
) []*Candidate {
	c := applyConfig(options...)

	// Calculate query context lines
	queryFullPath := lo.Must(os.ReadFile(filepath.Join(basePath, queryFilename)))
	queryFileLines := lines(queryFullPath)

	contextStartLine := max(1, queryStartLine-c.contextLines)               // inclusive, 1-indexed
	contextEndLine := min(len(queryFileLines), queryEndLine+c.contextLines) // inclusive, 1-indexed
	contextLines := queryFileLines[contextStartLine-1 : contextEndLine]

	searchFiles := make([]string, 0)
	lo.Must0(filepath.WalkDir(searchRoot, func(fullPath string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		relPath := lo.Must(filepath.Rel(searchRoot, fullPath))
		searchFiles = append(searchFiles, relPath)
		return nil
	}))

	// Search for co-change candidates!
	p := pool.NewWithResults[[]*Candidate]().WithMaxGoroutines(runtime.NumCPU())
	for _, filename := range searchFiles {
		p.Go(func() []*Candidate {
			return fileSearch(contextLines, searchRoot, filename, c.similarityThreshold)
		})
	}
	var candidates []*Candidate
	for _, fileCandidates := range p.Wait() {
		candidates = append(candidates, fileCandidates...)
	}

	// Account for enlarged context area
	contextStartDiff := queryStartLine - contextStartLine
	contextEndDiff := contextEndLine - queryEndLine
	for _, candidate := range candidates {
		candidate.StartLine += contextStartDiff
		candidate.EndLine -= contextEndDiff
	}

	// File proximity ranking
	// 1. Asc sort by file proximity on file tree
	// 2. Desc sort by similarity (as a fallback)
	slices.SortFunc(candidates, ds.SortCompose(
		ds.SortAsc(func(e *Candidate) int {
			return FileTreeDistance(queryFilename, e.Filename)
		}),
		ds.SortDesc(func(e *Candidate) float64 {
			return e.Similarity
		}),
	))

	return candidates
}

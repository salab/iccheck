package files

import (
	"bytes"
	"github.com/salab/iccheck/pkg/utils/strs"
	"io"
	"os"
	"strings"
)

// FileTreeDistance calculates distance in file tree according to FLeCCS ranking
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

// LineStartIndices returns start indices of lines in the given bytes slice.
func LineStartIndices(s []byte) []int {
	currentIdx := 0
	indices := make([]int, 0, 32)

	var line []byte
	var found bool
	for {
		line, s, found = bytes.Cut(s, []byte{'\n'})
		indices = append(indices, currentIdx)
		currentIdx += len(line) + 1

		if !found {
			break
		}
	}

	return indices
}

func ReadAll(reader io.ReadCloser, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func LengthsAndBigrams(content []byte, startLine, endLine int) (lengths []int, bigrams []strs.BigramSet) {
	indices := LineStartIndices(content)
	nLines := len(indices)
	indices = append(indices, len(content)) // to make code below cleaner

	if endLine <= 0 {
		endLine = nLines
	}
	lengths = make([]int, 0, endLine-startLine+1)
	bigrams = make([]strs.BigramSet, 0, endLine-startLine+1)

	for i := startLine - 1; i < endLine; i++ {
		startIdx := indices[i]
		endIdx := indices[i+1]
		lengths = append(lengths, endIdx-startIdx)

		line := content[startIdx:endIdx]
		bigrams = append(bigrams, strs.Bigram(line))
	}

	bigrams = strs.CompactBigrams(bigrams) // For performance
	return
}

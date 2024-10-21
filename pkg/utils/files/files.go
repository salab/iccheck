package files

import (
	"bytes"
	"github.com/pkg/errors"
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

// LineIndices returns start indices of lines in the given bytes slice.
func LineIndices(s []byte) []int {
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

func ReadFileLines(filename string, startLine, endLine int) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "reading file contents %v", filename)
	}
	indices := LineIndices(content)

	var startIdx, endIdx int
	startIdx = indices[startLine]
	if endLine == len(indices)-1 {
		endIdx = len(content)
	} else {
		endIdx = indices[endLine+1]
	}

	return content[startIdx:endIdx], nil
}

func ReadAll(reader io.ReadCloser, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

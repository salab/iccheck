package files

import (
	"bytes"
	"github.com/samber/lo"
	"io/fs"
	"os"
	"path/filepath"
)

func WalkAllFilenames(root string) []string {
	filenames := make([]string, 0)
	lo.Must0(filepath.WalkDir(root, func(fullPath string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		relPath := lo.Must(filepath.Rel(root, fullPath))
		filenames = append(filenames, relPath)
		return nil
	}))
	return filenames
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

func ReadFileLines(filename string, startLine, endLine int) []byte {
	content := lo.Must(os.ReadFile(filename))
	indices := LineIndices(content)

	var startIdx, endIdx int
	startIdx = indices[startLine]
	if endLine == len(indices)-1 {
		endIdx = len(content)
	} else {
		endIdx = indices[endLine+1]
	}

	return content[startIdx:endIdx]
}

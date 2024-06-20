package domain

import (
	"github.com/go-git/go-git/v5/utils/binary"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"
	"github.com/salab/iccheck/pkg/utils/ds"
	"os"
	"strings"
)

// Searcher is an interface for use with clone search backend implementations.
type Searcher interface {
	Files() ([]string, error)
	Open(name string) (SearcherFile, error)
}

type SearcherFile interface {
	Name() string
	IsBinary() (bool, error)
	Lines() ([]string, error)
}

func NewSearcherFromTree(tree Tree) Searcher {
	return &diffTreeSearcher{tree: tree}
}

type diffTreeSearcher struct {
	tree Tree
}

func (d *diffTreeSearcher) Files() ([]string, error) {
	node, err := d.tree.Noder()
	if err != nil {
		return nil, err
	}

	rootChildren, err := node.Children()
	if err != nil {
		return nil, err
	}
	return ds.FlatMapError(rootChildren, func(rootChild noder.Noder) ([]string, error) {
		return listFilesFromNoder(rootChild, nil)
	})
}

func listFilesFromNoder(node noder.Noder, path []string) ([]string, error) {
	if !node.IsDir() {
		return []string{strings.Join(append(path, node.Name()), string(os.PathSeparator))}, nil
	}

	thisName := node.Name()
	children, err := node.Children()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, child := range children {
		childFiles, err := listFilesFromNoder(child, append(path, thisName))
		if err != nil {
			return nil, err
		}
		files = append(files, childFiles...)
	}
	return files, nil
}

func (d *diffTreeSearcher) Open(name string) (SearcherFile, error) {
	content, err := d.tree.ReadFile(name)
	if err != nil {
		return nil, err
	}

	// Check binary
	isBinary, err := binary.IsBinary(strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	return &inMemoryFile{
		name:     name,
		content:  content,
		isBinary: isBinary,
	}, nil
}

type inMemoryFile struct {
	name     string
	content  string
	isBinary bool
}

func (i *inMemoryFile) Name() string {
	return i.name
}

func (i *inMemoryFile) IsBinary() (bool, error) {
	return i.isBinary, nil
}

func (i *inMemoryFile) Lines() ([]string, error) {
	return strings.Split(i.content, "\n"), nil
}

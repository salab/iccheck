package domain

import (
	"bytes"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/utils/binary"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/salab/iccheck/pkg/utils/files"
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
	Content() ([]byte, error)
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

func nodeIsSubmodule(node noder.Noder) bool {
	h := node.Hash()
	return len(h) == 24 && bytes.Equal(h[20:24], filemode.Submodule.Bytes())
}

func listFilesFromNoder(node noder.Noder, path []string) ([]string, error) {
	if !node.IsDir() {
		if nodeIsSubmodule(node) {
			return nil, nil // Skip detection of submodule entirely, for now
		}
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
	return &inMemoryFile{
		d:    d,
		name: name,
	}, nil
}

type inMemoryFile struct {
	d    *diffTreeSearcher
	name string

	isBinary *bool
	content  *[]byte
}

func (i *inMemoryFile) Name() string {
	return i.name
}

func (i *inMemoryFile) IsBinary() (bool, error) {
	if i.isBinary != nil {
		return *i.isBinary, nil
	}

	reader, err := i.d.tree.Reader(i.name)
	if err != nil {
		return false, err
	}
	isBinary, err := binary.IsBinary(reader)
	if err != nil {
		return false, err
	}
	err = reader.Close()
	if err != nil {
		return false, err
	}

	i.isBinary = &isBinary
	return isBinary, nil
}

func (i *inMemoryFile) Content() ([]byte, error) {
	if i.content != nil {
		return *i.content, nil
	}

	b, err := files.ReadAll(i.d.tree.Reader(i.name))
	if err != nil {
		return nil, err
	}

	i.content = &b
	return b, nil
}

package domain

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"io"
	"sync"
)

type Tree interface {
	Files() []string
	Open(name string) (File, error)
}

type File interface {
	Name() string
	IsBinary() (bool, error)
	Lines() ([]string, error)
}

type goGitTree struct {
	tree  *object.Tree
	files []string

	// NOTE: opening files via go-git's Open() is not thread-safe for some reason...
	l sync.Mutex
}

func NewTreeWalkerImplGoGit(tree *object.Tree) (Tree, error) {
	files, err := listGoGitFiles(tree)
	if err != nil {
		return nil, err
	}
	return &goGitTree{
		tree:  tree,
		files: files,
	}, nil
}

func listGoGitFiles(tree *object.Tree) ([]string, error) {
	var files []string
	w := object.NewTreeWalker(tree, true, nil)
	for {
		name, entry, err := w.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "error iterating tree")
		}

		// Skip directory
		if !entry.Mode.IsFile() {
			continue
		}
		files = append(files, name)
	}
	w.Close()
	return files, nil
}

func (g *goGitTree) Files() []string {
	return g.files
}

func (g *goGitTree) Open(name string) (File, error) {
	g.l.Lock()
	defer g.l.Unlock()

	file, err := g.tree.File(name)
	if err != nil {
		return nil, err
	}
	return &goGitFile{file: file}, nil
}

type goGitFile struct {
	file *object.File
}

func (f *goGitFile) Name() string {
	return f.file.Name
}

func (f *goGitFile) IsBinary() (bool, error) {
	return f.file.IsBinary()
}

func (f *goGitFile) Lines() ([]string, error) {
	return f.file.Lines()
}

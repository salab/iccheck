package domain

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/go-git/go-git/v5/utils/merkletrie/filesystem"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"
	"github.com/pkg/errors"
	"io"
	"strings"
)

type goGitCommitTree struct {
	commit *object.Commit
}

func NewGoGitCommitTree(commit *object.Commit) Tree {
	return &goGitCommitTree{commit: commit}
}

func (g *goGitCommitTree) String() string {
	return g.commit.String()
}

func (g *goGitCommitTree) Tree() (t *object.Tree, err error, ok bool) {
	t, err = g.commit.Tree()
	if err != nil {
		return nil, errors.Wrap(err, "resolving commit tree"), false
	}
	return t, nil, true
}

func (g *goGitCommitTree) Noder() (noder.Noder, error) {
	cTree, err := g.commit.Tree()
	if err != nil {
		return nil, errors.Wrap(err, "resolving commit tree")
	}
	node := object.NewTreeRootNode(cTree)
	return node, nil
}

func (g *goGitCommitTree) FilterIgnoredChanges(changes merkletrie.Changes) merkletrie.Changes {
	return changes
}

func (g *goGitCommitTree) ReadFile(path string) (string, error) {
	file, err := g.commit.File(path)
	if err != nil {
		return "", errors.Wrapf(err, "resolving file %v", path)
	}
	content, err := file.Contents()
	if err != nil {
		return "", errors.Wrapf(err, "reading file contents %v", file)
	}
	return content, nil
}

type goGitIndexTree struct{} // TODO?

type goGitWorktree struct {
	worktree *git.Worktree
	fs       billy.Filesystem
}

func NewGoGitWorkTree(worktree *git.Worktree) Tree {
	return &goGitWorktree{worktree: worktree, fs: worktree.Filesystem}
}

func newGoGitWorkTreeWithFS(worktree *git.Worktree, fs billy.Filesystem) *goGitWorktree {
	return &goGitWorktree{worktree: worktree, fs: fs}
}

func (g *goGitWorktree) String() string {
	return "WORKTREE"
}

func (g *goGitWorktree) Tree() (t *object.Tree, err error, ok bool) {
	return nil, nil, false
}

func (g *goGitWorktree) Noder() (noder.Noder, error) {
	submodules, err := getSubmodulesStatus(g.worktree)
	if err != nil {
		return nil, errors.Wrap(err, "getting submodules status")
	}
	node := filesystem.NewRootNode(g.fs, submodules)
	return node, nil
}

func (g *goGitWorktree) FilterIgnoredChanges(changes merkletrie.Changes) merkletrie.Changes {
	return excludeIgnoredChanges(g.worktree, changes)
}

func (g *goGitWorktree) ReadFile(path string) (string, error) {
	file, err := g.fs.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "resolving file %v", path)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		return "", errors.Wrapf(err, "reading file contents %v", path)
	}
	err = file.Close()
	if err != nil {
		return "", errors.Wrapf(err, "closing file %v", path)
	}
	return string(content), nil
}

type goGitWorktreeWithOverlay struct {
	*goGitWorktree
}

// billyFSOverlay intercepts Open() calls to billy.Filesystem
// for use with filesystem.NewRootNode and Tree.ReadFile methods.
type billyFSOverlay struct {
	billy.Filesystem
	overlay map[string]string
}

func (o *billyFSOverlay) Open(path string) (billy.File, error) {
	if content, ok := o.overlay[path]; ok {
		return &billyInMemoryFile{
			name:   path,
			Reader: strings.NewReader(content),
		}, nil
	}
	return o.Filesystem.Open(path)
}

type billyInMemoryFile struct {
	name string
	*strings.Reader
	nullBillyFile // Nest field to avoid "ambiguous field selector"
}

type nullBillyFile struct {
	billy.File // Intentionally leave to null to panic if called on unimplemented methods
}

func (f *billyInMemoryFile) Name() string {
	return f.name
}

func NewGoGitWorktreeWithOverlay(worktree *git.Worktree, overlay map[string]string) Tree {
	fs := &billyFSOverlay{Filesystem: worktree.Filesystem, overlay: overlay}
	return &goGitWorktreeWithOverlay{
		goGitWorktree: newGoGitWorkTreeWithFS(worktree, fs),
	}
}

func (g *goGitWorktreeWithOverlay) String() string {
	return "WORKTREE+Override"
}

type fileSystemTree struct{} // TODO?

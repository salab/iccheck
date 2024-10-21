package domain

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/go-git/go-git/v5/utils/merkletrie/filesystem"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
	"io"
	"strings"
	"sync"
)

type goGitCommitTree struct {
	commit *object.Commit

	// go-git's (*object).Commit does not allow concurrent read through File() for some reason
	l sync.Mutex
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

func (g *goGitCommitTree) Reader(path string) (io.ReadCloser, error) {
	g.l.Lock()
	defer g.l.Unlock()

	file, err := g.commit.File(path)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving file %v", path)
	}
	return file.Reader()
}

type goGitIndexTree struct{} // TODO?

type goGitWorktree struct {
	worktree *git.Worktree
	fs       billy.Filesystem

	ignoreMatcher gitignore.Matcher
}

func newGoGitWorkTree(worktree *git.Worktree, fs billy.Filesystem) (*goGitWorktree, error) {
	patterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil {
		return nil, err
	}
	patterns = append(patterns, worktree.Excludes...)
	m := gitignore.NewMatcher(patterns)

	return &goGitWorktree{
		worktree: worktree,
		fs:       fs,

		ignoreMatcher: m,
	}, nil
}

func NewGoGitWorkTree(worktree *git.Worktree) (Tree, error) {
	return newGoGitWorkTree(worktree, worktree.Filesystem)
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
	return &filesystemGitignoreOverlay{
		ignoreMatcher: g.ignoreMatcher,
		pathPrefix:    nil,
		Noder:         node,
	}, nil
}

func (g *goGitWorktree) FilterIgnoredChanges(changes merkletrie.Changes) merkletrie.Changes {
	return excludeIgnoredChanges(g.ignoreMatcher, changes)
}

func (g *goGitWorktree) Reader(path string) (io.ReadCloser, error) {
	file, err := g.fs.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving file %v", path)
	}
	return file, nil
}

// filesystemGitignoreOverlay intercepts Children() (and NumChildren()) calls and
// wraps their return values in order to prevent gitignore-d files and directories
// from being listed in noder.Noder methods.
//
// Preventing gitignore-d files from being listed before being compared allows faster
// comparison, and prevents unnecessary file accesses from occurring in case
// there are files that should not be accessed in gitignore-d directories.
type filesystemGitignoreOverlay struct {
	ignoreMatcher gitignore.Matcher
	pathPrefix    []string
	noder.Noder
}

func (f *filesystemGitignoreOverlay) Children() ([]noder.Noder, error) {
	children, err := f.Noder.Children()
	if err != nil {
		return nil, err
	}
	filtered := lo.Filter(children, func(n noder.Noder, _ int) bool {
		return !f.ignoreMatcher.Match(append(f.pathPrefix, n.Name()), n.IsDir())
	})
	wrapped := ds.Map(filtered, func(n noder.Noder) noder.Noder {
		return &filesystemGitignoreOverlay{
			ignoreMatcher: f.ignoreMatcher,
			pathPrefix:    append(ds.Copy(f.pathPrefix), n.Name()),
			Noder:         n,
		}
	})
	return wrapped, nil
}

func (f *filesystemGitignoreOverlay) NumChildren() (int, error) {
	children, err := f.Children()
	if err != nil {
		return 0, err
	}
	return len(children), nil
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

func (f *billyInMemoryFile) Close() error {
	return nil
}

func NewGoGitWorktreeWithOverlay(worktree *git.Worktree, overlay map[string]string) (Tree, error) {
	fs := &billyFSOverlay{Filesystem: worktree.Filesystem, overlay: overlay}
	tree, err := newGoGitWorkTree(worktree, fs)
	if err != nil {
		return nil, err
	}
	return &goGitWorktreeWithOverlay{
		goGitWorktree: tree,
	}, nil
}

func (g *goGitWorktreeWithOverlay) String() string {
	return "WORKTREE+Override"
}

type fileSystemTree struct{} // TODO?

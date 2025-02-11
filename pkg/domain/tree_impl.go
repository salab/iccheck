package domain

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie/filesystem"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
	"io"
	"os"
	"runtime"
	"strings"
)

type goGitCommitTree struct {
	commit *object.Commit
	ref    string
}

func NewGoGitCommitTree(commit *object.Commit, ref string) Tree {
	g := &goGitCommitTree{commit: commit, ref: ref}
	return g
}

func (g *goGitCommitTree) String() string {
	return fmt.Sprintf("%s (%s)", g.ref, g.commit.Hash.String())
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

func (g *goGitCommitTree) Reader(path string) (io.ReadCloser, error) {
	file, err := g.commit.File(path)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving path %v", path)
	}
	return file.Reader()
}

type goGitIndexTree struct{} // TODO?

// billyFSGitignore intercepts billy.Filesystem.Readdir() calls to filter out gitignore-d files.
type billyFSGitignore struct {
	billy.Filesystem

	m gitignore.Matcher
	// NOTE: Ignoring worktree directly by gitignore patterns results in invalid diff
	// - that is, if git-tracked file is present in a gitignore-d directory and is checked out,
	// ignoring that file by overlay will result in a 'deleted' diff.
	// To cope with this problem, if there exists a tracked file or directory in a commit tree we're comparing with,
	// do NOT mark that file or directory and return that as it is.
	// If the file is to be ignored, it will be filtered after the tree diffing in Status() method and such.
	commitTree *object.Tree
}

func ReadSystemGitignore() ([]gitignore.Pattern, error) {
	rootFs := osfs.New("/")
	system, err := gitignore.LoadSystemPatterns(rootFs)
	if err != nil {
		return nil, err
	}
	user, err := gitignore.LoadGlobalPatterns(rootFs)
	// Ignoring error if windows here, because go-git cannot load user home .gitignore for some reason
	// example error: "open \\C:\\Users\\moto\\.gitconfig: The specified path is invalid."
	if err != nil && runtime.GOOS != "windows" {
		return nil, err
	}
	return append(system, user...), nil
}

func appendSystemPatterns(fs billy.Filesystem) ([]gitignore.Pattern, error) {
	systemPatterns, err := ReadSystemGitignore()
	if err != nil {
		return nil, err
	}
	repoPatterns, err := gitignore.ReadPatterns(fs, nil)
	if err != nil {
		return nil, err
	}
	return append(systemPatterns, repoPatterns...), nil
}

func NewBillyFSGitignore(fs billy.Filesystem, commitTree *object.Tree) (billy.Filesystem, error) {
	patterns, err := appendSystemPatterns(fs)
	if err != nil {
		return nil, err
	}
	m := gitignore.NewMatcher(patterns)
	return &billyFSGitignore{Filesystem: fs, m: m, commitTree: commitTree}, nil
}

func (b *billyFSGitignore) ReadDir(path string) ([]os.FileInfo, error) {
	files, err := b.Filesystem.ReadDir(path)
	if err != nil {
		return nil, err
	}
	elms := strings.Split(path, string(os.PathSeparator))
	if path == "" {
		elms = nil
	}
	files = lo.Reject(files, func(f os.FileInfo, _ int) bool {
		ignoreMatch := b.m.Match(append(elms, f.Name()), f.IsDir())
		if !ignoreMatch {
			return false
		}
		if _, err := b.commitTree.FindEntry(path + string(os.PathSeparator) + f.Name()); err == nil {
			// File exists in HEAD tree we're differencing against - do NOT mark this as ignored.
			return false
		}
		return true
	})
	return files, nil
}

func getHeadTree(repo *git.Repository) (*object.Tree, error) {
	headHash, err := repo.ResolveRevision("HEAD")
	if err != nil {
		return nil, errors.Wrap(err, "resolving HEAD ref")
	}
	headCommit, err := repo.CommitObject(*headHash)
	if err != nil {
		return nil, errors.Wrap(err, "resolving HEAD commit")
	}
	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, errors.Wrap(err, "resolving HEAD tree")
	}
	return headTree, nil
}

type GoGitWorktree struct {
	worktree *git.Worktree
}

func NewGoGitWorkTree(repo *git.Repository) (*GoGitWorktree, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	headTree, err := getHeadTree(repo)
	if err != nil {
		return nil, err
	}

	worktree.Excludes, err = ReadSystemGitignore()
	if err != nil {
		return nil, err
	}

	fs, err := NewBillyFSGitignore(worktree.Filesystem, headTree)
	if err != nil {
		return nil, err
	}
	worktree.Filesystem = fs
	return &GoGitWorktree{worktree: worktree}, nil
}

func (g *GoGitWorktree) String() string {
	return "WORKTREE"
}

func (g *GoGitWorktree) Tree() (t *object.Tree, err error, ok bool) {
	return nil, nil, false
}

func (g *GoGitWorktree) Noder() (noder.Noder, error) {
	submodules, err := getSubmodulesStatus(g.worktree)
	if err != nil {
		return nil, errors.Wrap(err, "getting submodules status")
	}
	node := filesystem.NewRootNode(g.worktree.Filesystem, submodules)
	return node, nil
}

func (g *GoGitWorktree) Reader(path string) (io.ReadCloser, error) {
	file, err := g.worktree.Filesystem.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving file %v", path)
	}
	return file, nil
}

func (g *GoGitWorktree) Status() (git.Status, error) {
	return g.worktree.Status()
}

type goGitWorktreeWithOverlay struct {
	*GoGitWorktree
}

// billyFSOverlay intercepts Open() calls to billy.Filesystem
// for use with filesystem.NewRootNode and Tree.Reader methods.
type billyFSOverlay struct {
	billy.Filesystem
	overlay *ds.SyncMap[string, string]
}

func (o *billyFSOverlay) Open(path string) (billy.File, error) {
	if content, ok := o.overlay.Load(path); ok {
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

func NewGoGitWorktreeWithOverlay(repo *git.Repository, overlay *ds.SyncMap[string, string]) (Tree, error) {
	tree, err := NewGoGitWorkTree(repo)
	if err != nil {
		return nil, err
	}
	fs := &billyFSOverlay{Filesystem: tree.worktree.Filesystem, overlay: overlay}
	tree.worktree.Filesystem = fs
	return &goGitWorktreeWithOverlay{
		GoGitWorktree: tree,
	}, nil
}

func (g *goGitWorktreeWithOverlay) String() string {
	return "WORKTREE+Override"
}

type fileSystemTree struct{} // TODO?

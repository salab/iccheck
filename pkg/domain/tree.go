package domain

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/diff"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/salab/iccheck/pkg/utils/files"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io"
	"log/slog"
)

// Tree represents a tree-like filesystem with various backends.
// Tree interface is used for comparison between two filesystem snapshots without depending on
// a specific backend implementation.
type Tree interface {
	// String should return a short human-readable representation of this tree.
	String() string
	// Tree MAY return concrete go-git tree implementation, for use with direct comparison from the other tree
	// and with the go-git embedded rename detection function.
	//
	// If Tree() returns false, comparison will fallback to Noder() diffing.
	Tree() (t *object.Tree, err error, ok bool)
	// Noder returns noder.Noder for diffing changes from the other noder.Noder instance.
	Noder() (noder.Noder, error)
	// FilterIgnoredChanges SHOULD filter ignored changes (such as paths specified in .gitignore), if any.
	FilterIgnoredChanges(changes merkletrie.Changes) merkletrie.Changes
	// Reader returns io.ReadCloser to the file contents.
	Reader(path string) (io.ReadCloser, error)
}

// --- adapters below

// diffFile is an adapter to diff.File.
type diffFile struct {
	path string
}

func (f *diffFile) Hash() plumbing.Hash {
	panic("not implemented")
}

func (f *diffFile) Mode() filemode.FileMode {
	panic("not implemented")
}

func (f *diffFile) Path() string {
	return f.path
}

// worktreeChunk is an adapter for diff.Chunk.
type worktreeChunk struct {
	content string
	typ     fdiff.Operation
}

func (f *worktreeChunk) Content() string {
	return f.content
}

func (f *worktreeChunk) Type() fdiff.Operation {
	return f.typ
}

// worktreeFilePatch is an adapter for diff.FilePatch.
type worktreeFilePatch struct {
	from, to fdiff.File
	chunks   []fdiff.Chunk
}

var _ fdiff.FilePatch = &worktreeFilePatch{}

func (fp *worktreeFilePatch) IsBinary() bool {
	return len(fp.chunks) == 0
}

func (fp *worktreeFilePatch) Files() (from, to fdiff.File) {
	return fp.from, fp.to
}

func (fp *worktreeFilePatch) Chunks() []fdiff.Chunk {
	return fp.chunks
}

// --- adapters above

// DiffTrees calculates diffs between two trees.
//
// Diff implementations is referred from go-git's (*git.Worktree).Status() method.
func DiffTrees(base, target Tree) ([]fdiff.FilePatch, error) {
	baseTree, err, baseOk := base.Tree()
	if err != nil {
		return nil, errors.Wrap(err, "getting base git tree")
	}
	targetTree, err, targetOk := target.Tree()
	if err != nil {
		return nil, errors.Wrap(err, "getting target git tree")
	}

	// Special path: direct comparison between commits with go-git implementation
	// This also supports detecting renames
	if baseOk && targetOk {
		changes, err := object.DiffTreeWithOptions(context.TODO(), baseTree, targetTree, object.DefaultDiffTreeOptions)
		if err != nil {
			return nil, errors.Wrap(err, "diffing base to target commit tree")
		}
		slog.Info("File changes detected", "files", len(changes))

		filePatches, err := ds.FlatMapError(changes, func(c *object.Change) ([]fdiff.FilePatch, error) {
			p, err := c.Patch()
			if err != nil {
				return nil, err
			}
			return p.FilePatches(), nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "resolving patch from change")
		}
		return filePatches, nil
	}

	// Normal path: general comparison between noder.Noder
	baseNode, err := base.Noder()
	if err != nil {
		return nil, errors.Wrap(err, "getting base noder")
	}
	targetNode, err := target.Noder()
	if err != nil {
		return nil, errors.Wrap(err, "getting target noder")
	}

	changes, err := merkletrie.DiffTree(baseNode, targetNode, diffTreeIsEquals)
	if err != nil {
		return nil, errors.Wrap(err, "diffing nodes")
	}

	changes = base.FilterIgnoredChanges(changes) // maybe exclusion by base tree is not necessary
	changes = target.FilterIgnoredChanges(changes)

	slog.Info("File changes detected", "files", len(changes))

	filePatches, err := ds.MapError(changes, func(c merkletrie.Change) (fdiff.FilePatch, error) {
		return changeToFilePatch(base, target, c)
	})
	if err != nil {
		return nil, errors.Wrap(err, "converting change to file patch")
	}
	return filePatches, nil
}

func changeToFilePatch(base, target Tree, c merkletrie.Change) (fdiff.FilePatch, error) {
	action, err := c.Action()
	if err != nil {
		return nil, errors.Wrap(err, "retrieving action for a change")
	}

	var fp worktreeFilePatch

	if action == merkletrie.Delete || action == merkletrie.Modify {
		fromFilePath := c.From.String()
		fp.from = &diffFile{fromFilePath}
	}

	if action == merkletrie.Insert || action == merkletrie.Modify {
		toFilePath := c.To.String()
		fp.to = &diffFile{toFilePath}
	}

	if action == merkletrie.Modify {
		fromFilePath := c.From.String()
		fromContent, err := files.ReadAll(base.Reader(fromFilePath))
		if err != nil {
			return nil, errors.Wrapf(err, "reading base file contents %v", fromFilePath)
		}

		toFilePath := c.To.String()
		toContent, err := files.ReadAll(target.Reader(toFilePath))
		if err != nil {
			return nil, errors.Wrapf(err, "reading target file contents %v", toFilePath)
		}

		diffs := diff.Do(string(fromContent), string(toContent))
		fp.chunks = ds.Map(diffs, func(d diffmatchpatch.Diff) fdiff.Chunk {
			var typ fdiff.Operation
			switch d.Type {
			case diffmatchpatch.DiffDelete:
				typ = fdiff.Delete
			case diffmatchpatch.DiffEqual:
				typ = fdiff.Equal
			case diffmatchpatch.DiffInsert:
				typ = fdiff.Add
			default:
				panic(fmt.Sprintf("unknown diff type: %v", d.Type))
			}
			return &worktreeChunk{d.Text, typ}
		})
	}

	return &fp, nil
}

// This file exists as a workaround to get diff chunks between a commit and a worktree.
// go-git does not support retrieving such diffs unlike 'git diff',
// due to its internal implementation difference between commit tree representation (*object.Tree)
// and worktree representation (*git.Worktree).
// Related issue: https://github.com/go-git/go-git/issues/561
//
// This file's implementation does not support detecting renames between a commit and a worktree,
// as go-git's object.DetectRename function is dependent on object.Changes instance which requires
// git tree object (storer.EncodedObjectStorer) backend.

package search

import (
	"bytes"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/diff"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/go-git/go-git/v5/utils/merkletrie/filesystem"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"
	"github.com/samber/lo"
)

// ----- The following are unexported functions from github.com/go-git/go-git/v5@v5.12.0

func getSubmodulesStatus(w *git.Worktree) (map[string]plumbing.Hash, error) {
	o := map[string]plumbing.Hash{}

	sub, err := w.Submodules()
	if err != nil {
		return nil, err
	}

	status, err := sub.Status()
	if err != nil {
		return nil, err
	}

	for _, s := range status {
		if s.Current.IsZero() {
			o[s.Path] = s.Expected
			continue
		}

		o[s.Path] = s.Current
	}

	return o, nil
}

var emptyNoderHash = make([]byte, 24)

func diffTreeIsEquals(a, b noder.Hasher) bool {
	hashA := a.Hash()
	hashB := b.Hash()

	if bytes.Equal(hashA, emptyNoderHash) || bytes.Equal(hashB, emptyNoderHash) {
		return false
	}

	return bytes.Equal(hashA, hashB)
}

func excludeIgnoredChanges(w *git.Worktree, changes merkletrie.Changes) merkletrie.Changes {
	patterns, err := gitignore.ReadPatterns(w.Filesystem, nil)
	if err != nil {
		return changes
	}

	patterns = append(patterns, w.Excludes...)

	if len(patterns) == 0 {
		return changes
	}

	m := gitignore.NewMatcher(patterns)

	var res merkletrie.Changes
	for _, ch := range changes {
		var path []string
		for _, n := range ch.To {
			path = append(path, n.Name())
		}
		if len(path) == 0 {
			for _, n := range ch.From {
				path = append(path, n.Name())
			}
		}
		if len(path) != 0 {
			isDir := (len(ch.To) > 0 && ch.To.IsDir()) || (len(ch.From) > 0 && ch.From.IsDir())
			if m.Match(path, isDir) {
				if len(ch.From) == 0 {
					continue
				}
			}
		}
		res = append(res, ch)
	}
	return res
}

// ----- Unexported functions from go-git above

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

func diffCommitToWorktree(commit *object.Commit, workTree *git.Worktree) merkletrie.Changes {
	// git diff commit WORKTREE
	fromTree := lo.Must(commit.Tree())
	from := object.NewTreeRootNode(fromTree)

	submodules := lo.Must(getSubmodulesStatus(workTree))
	to := filesystem.NewRootNode(workTree.Filesystem, submodules)

	changes := lo.Must(merkletrie.DiffTree(from, to, diffTreeIsEquals))
	changes = excludeIgnoredChanges(workTree, changes)

	return changes
}

func diffChunksToWorktree(c merkletrie.Change, commit *object.Commit, workTree *git.Worktree) fdiff.FilePatch {
	action := lo.Must(c.Action())
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
		fromFile := lo.Must(commit.File(fromFilePath))
		fromContent := lo.Must(fromFile.Contents())

		toFilePath := c.To.String()
		toFile := lo.Must(workTree.Filesystem.Open(toFilePath))
		toContent := string(lo.Must(io.ReadAll(toFile)))
		lo.Must0(toFile.Close())

		diffs := diff.Do(fromContent, toContent)
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

	return &fp
}

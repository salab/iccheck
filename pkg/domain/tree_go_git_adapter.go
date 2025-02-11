// This file exists as a workaround to get diff chunks between a commit and a worktree.
// go-git does not support retrieving such diffs unlike 'git diff',
// due to its internal implementation difference between commit tree representation (*object.Tree)
// and worktree representation (*git.Worktree).
// Related issue: https://github.com/go-git/go-git/issues/561
//
// This file's implementation does not support detecting renames between a commit and a worktree,
// as go-git's object.DetectRename function is dependent on object.Changes instance which requires
// git tree object (storer.EncodedObjectStorer) backend.

package domain

import (
	"bytes"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/utils/merkletrie/noder"
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

// ----- Unexported functions from go-git above

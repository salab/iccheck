package search

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/samber/lo"
	"testing"
)

func resolveTree(repo *git.Repository, ref string) (domain.Tree, error) {
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, errors.Wrapf(err, "resolving hash revision from %v", ref)
	}
	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving commit from hash %v", *hash)
	}
	return domain.NewGoGitCommitTree(commit), nil
}

func TestSearch(t *testing.T) {
	repoDir := "/home/moto/projects/salab/iccheck/data/NeoShowcase"
	fromRef := "e79b8d658dde4c68186c4dfdf887183db0093430~"
	toRef := "e79b8d658dde4c68186c4dfdf887183db0093430"

	// Resolve
	repo := lo.Must(git.PlainOpen(repoDir))
	fromTree := lo.Must(resolveTree(repo, fromRef))
	toTree := lo.Must(resolveTree(repo, toRef))

	// Search
	_, err := Search(fromTree, toTree)
	if err != nil {
		t.Error(err)
	}
}

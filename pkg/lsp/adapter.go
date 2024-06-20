package lsp

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/search"
	"github.com/sourcegraph/go-lsp"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func getGitRoot(root string, filename string) ([]string, bool) {
	path := strings.Split(filename, string(os.PathSeparator))
	for i := len(path) - 1; i >= 0; i-- {
		gitPath := []string{root}
		gitPath = append(gitPath, path[0:i]...)
		gitPath = append(gitPath, ".git")
		gitFullPath := filepath.Join(gitPath...)
		info, err := os.Stat(gitFullPath)
		if err == nil && info.IsDir() {
			return gitPath[1 : len(gitPath)-1], true
		}
	}
	return nil, false
}

func (h *handler) analyzeFile(ctx context.Context, filename string) (diagnostics []lsp.Diagnostic, err error) {
	diagnostics = make([]lsp.Diagnostic, 0) // Should not be "nil" because nil slice gets marshalled into "null"

	// Check .git directory
	gitPath, ok := getGitRoot(h.rootPath, filename)
	if !ok {
		slog.Warn(fmt.Sprintf("no parent of %s/%s contains git directory: skipping analysis", h.rootPath, filename))
		return diagnostics, nil
	}

	// Calculate
	cloneSets, err := h.calcCache.Get(ctx, filepath.Join(gitPath...))
	if err != nil {
		return diagnostics, err
	}

	// Transform
	content := h.files[filename]
	lines := strings.Split(content, "\n")
	for _, cs := range cloneSets {
		for _, c := range cs.Missing {
			detectedPath := filepath.Join(append(gitPath, c.Filename)...)
			if detectedPath == filename {
				diagnostics = append(diagnostics, lsp.Diagnostic{
					Range: lsp.Range{
						Start: lsp.Position{Line: c.StartL - 1, Character: 0},
						End:   lsp.Position{Line: c.EndL - 1, Character: len(lines[c.EndL-1])},
					},
					Severity: lsp.Warning,
					Code:     "",
					Source:   "",
					Message:  "Possibly missing a consistent change",
				})
			}
		}
	}

	return diagnostics, nil
}

func (h *handler) getCloneSets(_ context.Context, gitPath string) ([]*domain.CloneSet, error) {
	slog.Info("analyzing", "gitPath", gitPath)

	// Open repository
	gitFullPath := filepath.Join(append([]string{h.rootPath}, gitPath)...)
	repo, err := git.PlainOpen(gitFullPath)
	if err != nil {
		return nil, errors.Wrap(err, "opening git directory")
	}

	// Get head tree
	headHash, err := repo.ResolveRevision("HEAD")
	if err != nil {
		return nil, errors.Wrapf(err, "resolving hash revision from %v", headHash)
	}
	headCommit, err := repo.CommitObject(*headHash)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving commit from hash %v", *headHash)
	}
	headTree := domain.NewGoGitCommitTree(headCommit)

	// Get overlay tree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "resolving worktree")
	}
	targetTree := domain.NewGoGitWorktreeWithOverlay(worktree, h.files)

	// Calculate
	cloneSets, err := search.Search(headTree, targetTree)
	if err != nil {
		return nil, err
	}
	return cloneSets, nil
}

package lsp

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/search"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
	"github.com/sourcegraph/go-lsp"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const analyzeSourceName = "ICCheck"
const analyzeCodeName = "Consistency check"

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

func toLSPRange(c *domain.Clone, lines []string) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: c.StartL - 1, Character: 0},
		End:   lsp.Position{Line: c.EndL - 1, Character: len(lines[c.EndL-1])},
	}
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
		missingPaths := make(map[string]struct{}, len(cs.Missing))

		// For all missing parts, display warnings
		for _, c := range cs.Missing {
			missingPaths[c.Filename] = struct{}{}
			detectedPath := filepath.Join(append(gitPath, c.Filename)...)
			if detectedPath == filename {
				message := fmt.Sprintf("Missing a change? (%d out of %d clones changed)", len(cs.Changed), len(cs.Changed)+len(cs.Missing))
				diagnostics = append(diagnostics, lsp.Diagnostic{
					Range:    toLSPRange(c, lines),
					Severity: lsp.Warning,
					Code:     analyzeCodeName,
					Source:   analyzeSourceName,
					Message:  message,
				})
			}
		}

		// Also display warnings to changed lines, if no missing changes are nearby (in the same file)
		for _, c := range cs.Changed {
			detectedPath := filepath.Join(append(gitPath, c.Filename)...)
			_, hasMissingWarning := missingPaths[c.Filename]
			if detectedPath == filename && !hasMissingWarning {
				const missingFilepathDisplayLimit = 3
				missingPathList := lo.Keys(missingPaths)
				message := fmt.Sprintf(
					"Missing a change in other files? (%s%s)",
					strings.Join(ds.Limit(missingPathList, missingFilepathDisplayLimit), ", "),
					lo.Ternary(len(missingPathList) > missingFilepathDisplayLimit, ", ...", ""),
				)
				diagnostics = append(diagnostics, lsp.Diagnostic{
					Range:    toLSPRange(c, lines),
					Severity: lsp.Warning,
					Code:     analyzeCodeName,
					Source:   analyzeSourceName,
					Message:  message,
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

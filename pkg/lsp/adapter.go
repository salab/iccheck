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
	"time"
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

type lspPublishDiagnosticsParams struct {
	URI         lsp.DocumentURI   `json:"uri"`
	Diagnostics []*lsp.Diagnostic `json:"diagnostics"`
}

func (h *handler) analyzePath(ctx context.Context, gitPath string) (struct{}, error) {
	start := time.Now()
	defer func() {
		dur := time.Since(start)
		slog.Info(fmt.Sprintf("Analysis took %v", dur), "gitPath", gitPath)
		h.limiterLock.Lock() // Rate limit calculation must be serialized
		toAdd := dur.Milliseconds()
		added := h.limiter.Add(toAdd)
		if added < toAdd {
			sleepFor := time.Duration(float64(toAdd-added)/targetUtilization) * time.Millisecond
			slog.Warn(fmt.Sprintf("Analyze rate limit reached, sleeping for %v ...", sleepFor), "gitPath", gitPath)
			time.Sleep(sleepFor)
		}
		h.limiterLock.Unlock()
	}()
	slog.Debug("Analyzing ...", "gitPath", gitPath)

	diagnostics := make(map[string][]*lsp.Diagnostic)

	// Calculate
	cloneSets, err := h.getCloneSets(ctx, gitPath)
	if err != nil {
		return struct{}{}, err
	}

	// Transform
	for _, cs := range cloneSets {
		missingPaths := make(map[string]struct{}, len(cs.Missing))

		// For all missing parts, display warnings
		for _, c := range cs.Missing {
			missingPaths[c.Filename] = struct{}{}

			detectedPath := filepath.Join(append([]string{gitPath}, c.Filename)...)
			lines, err := h.filesCache.Get(ctx, detectedPath)
			if err != nil {
				return struct{}{}, err
			}

			message := fmt.Sprintf("Missing a change? (%d out of %d clones changed)", len(cs.Changed), len(cs.Changed)+len(cs.Missing))
			diagnostics[detectedPath] = append(diagnostics[detectedPath], &lsp.Diagnostic{
				Range:    toLSPRange(c, lines),
				Severity: lsp.Warning,
				Code:     analyzeCodeName,
				Source:   analyzeSourceName,
				Message:  message,
			})
		}

		// Also display warnings to changed lines, if no missing changes are nearby (in the same file)
		for _, c := range cs.Changed {
			detectedPath := filepath.Join(append([]string{gitPath}, c.Filename)...)

			_, hasMissingWarning := missingPaths[c.Filename]
			if !hasMissingWarning {
				lines, err := h.filesCache.Get(ctx, detectedPath)
				if err != nil {
					return struct{}{}, err
				}

				const missingFilepathDisplayLimit = 3
				missingPathList := lo.Keys(missingPaths)
				message := fmt.Sprintf(
					"Missing a change in other files? (%s%s)",
					strings.Join(ds.Limit(missingPathList, missingFilepathDisplayLimit), ", "),
					lo.Ternary(len(missingPathList) > missingFilepathDisplayLimit, ", ...", ""),
				)
				diagnostics[detectedPath] = append(diagnostics[detectedPath], &lsp.Diagnostic{
					Range:    toLSPRange(c, lines),
					Severity: lsp.Warning,
					Code:     analyzeCodeName,
					Source:   analyzeSourceName,
					Message:  message,
				})
			}
		}
	}

	// Publish diagnostics
	// TODO: remove old warnings when there are 0 diagnostics remaining in the file
	for filename, d := range diagnostics {
		err = h.conn.Notify(ctx, "textDocument/publishDiagnostics", lspPublishDiagnosticsParams{
			URI:         h.appendFilePrefix(filename),
			Diagnostics: d,
		})
		if err != nil {
			slog.Warn("failed to publish diagnostics", "file", filename, "error", err)
		}
	}
	// Remove old warnings when there are 0 diagnostics remaining in the file
	prevPaths, _ := h.previousDiagnostics.Load(gitPath)
	for _, prevPath := range prevPaths {
		if _, ok := diagnostics[prevPath]; !ok {
			err = h.conn.Notify(ctx, "textDocument/publishDiagnostics", lspPublishDiagnosticsParams{
				URI:         h.appendFilePrefix(prevPath),
				Diagnostics: make([]*lsp.Diagnostic, 0),
			})
			if err != nil {
				slog.Warn("failed to clear diagnostics", "file", prevPaths, "error", err)
			}
		}
	}

	// Store current diagnostic paths
	h.previousDiagnostics.Store(gitPath, lo.Keys(diagnostics))

	return struct{}{}, nil
}

func (h *handler) getCloneSets(ctx context.Context, gitPath string) ([]*domain.CloneSet, error) {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

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
	targetTree, err := domain.NewGoGitWorktreeWithOverlay(worktree, h.openFiles.Copy())
	if err != nil {
		return nil, errors.Wrap(err, "creating domain tree")
	}

	// Calculate
	cloneSets, err := search.Search(ctx, headTree, targetTree)
	if err != nil {
		return nil, err
	}
	return cloneSets, nil
}

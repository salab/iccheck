package lsp

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/search"
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

func getGitRoot(filename string) ([]string, bool) {
	path := strings.Split(filename, lspPathSeparator)
	for i := len(path) - 1; i >= 0; i-- {
		var parts []string
		parts = append(parts, path[0:i]...)
		parts = append(parts, ".git")
		info, err := os.Stat(filepath.Join(parts...))
		if err == nil && info.IsDir() {
			return parts[:len(parts)-1], true
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
		const filepathDisplayLimit = 3

		// For all missing parts, display warnings
		for _, c := range cs.Missing {
			detectedPath := filepath.Join(gitPath, c.Filename)
			lines, err := h.filesCache.Get(ctx, detectedPath)
			if err != nil {
				return struct{}{}, errors.Wrapf(err, "getting file contents for %s", detectedPath)
			}

			message := fmt.Sprintf(
				"Missing a change here? (%d out of %d clones changed: %s)",
				len(cs.Changed), len(cs.Changed)+len(cs.Missing),
				readablePaths(c.Filename, cs.Changed, filepathDisplayLimit),
			)
			lspPath := h.osToLSPFilepath(detectedPath)
			diagnostics[lspPath] = append(diagnostics[lspPath], &lsp.Diagnostic{
				Range:    toLSPRange(c, lines),
				Severity: lsp.Warning,
				Code:     analyzeCodeName,
				Source:   analyzeSourceName,
				Message:  message,
			})
		}

		// Also display warnings to changed lines, if no missing changes are nearby (in the same file)
		for _, c := range cs.Changed {
			detectedPath := filepath.Join(gitPath, c.Filename)
			lines, err := h.filesCache.Get(ctx, detectedPath)
			if err != nil {
				return struct{}{}, errors.Wrapf(err, "getting file contents for %s", detectedPath)
			}

			var message string
			var severity lsp.DiagnosticSeverity
			if len(cs.Missing) > 0 {
				// A change is missing.
				message = fmt.Sprintf(
					"Missing %s in other %d %s? (%s) (%d out of %d clones changed)",
					lo.Ternary(len(cs.Missing) == 1, "a change", "changes"),
					len(cs.Missing),
					lo.Ternary(len(cs.Missing) == 1, "location", "locations"),
					readablePaths(c.Filename, cs.Missing, filepathDisplayLimit),
					len(cs.Changed),
					len(cs.Changed)+len(cs.Missing),
				)
				severity = lsp.Warning
			} else {
				// No change is missing in this clone set, but still display "info" line to signify
				// that the user is editing a clone set.
				message = fmt.Sprintf(
					"Clone of size %d detected (%s)",
					len(cs.Changed)+len(cs.Missing),
					readablePaths(c.Filename, cs.Changed, filepathDisplayLimit),
				)
				severity = lsp.Info
			}
			lspPath := h.osToLSPFilepath(detectedPath)
			diagnostics[lspPath] = append(diagnostics[lspPath], &lsp.Diagnostic{
				Range:    toLSPRange(c, lines),
				Severity: severity,
				Code:     analyzeCodeName,
				Source:   analyzeSourceName,
				Message:  message,
			})
		}
	}

	// Publish diagnostics
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
	prevDiagnostics, _ := h.previousDiagnostics.Load(gitPath)
	for prevPath := range prevDiagnostics {
		if _, ok := diagnostics[prevPath]; !ok {
			err = h.conn.Notify(ctx, "textDocument/publishDiagnostics", lspPublishDiagnosticsParams{
				URI:         h.appendFilePrefix(prevPath),
				Diagnostics: make([]*lsp.Diagnostic, 0),
			})
			if err != nil {
				slog.Warn("failed to clear diagnostics", "file", prevPath, "error", err)
			}
		}
	}

	// Store current analysis results and diagnostic paths
	h.previousAnalysis.Store(gitPath, cloneSets)
	h.previousDiagnostics.Store(gitPath, diagnostics)

	return struct{}{}, nil
}

func (h *handler) getCloneSets(ctx context.Context, gitPath string) ([]*domain.CloneSet, error) {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Open repository
	repo, err := git.PlainOpen(gitPath)
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
	headTree := domain.NewGoGitCommitTree(headCommit, "HEAD")

	// Get overlay tree
	worktree, err := domain.NewGoGitWorktreeWithOverlay(repo, &h.openFiles)
	if err != nil {
		return nil, errors.Wrap(err, "creating domain tree")
	}

	// Read search config
	searchConf, err := h.searchConfCache.Get(ctx, gitPath)
	if err != nil {
		return nil, errors.Wrapf(err, "getting search config for %v", gitPath)
	}

	// Calculate
	queries, changedFiles, err := search.DiffTrees(ctx, headTree, worktree)
	if err != nil {
		return nil, errors.Wrapf(err, "diffing tree")
	}
	slog.Info(fmt.Sprintf("%d changed text chunk(s) were found within %d changed file(s).", len(queries), changedFiles))
	cloneSets, err := search.Search(ctx, h.algorithm, queries, worktree, searchConf)
	if err != nil {
		return nil, errors.Wrapf(err, "searching clone sets")
	}
	return cloneSets, nil
}

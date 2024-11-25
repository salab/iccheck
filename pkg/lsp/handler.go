package lsp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/utils/ds"
)

const fileURIPrefix = "file://"

func (h *handler) trimFilePrefix(uri lsp.DocumentURI) string {
	fullPath := strings.TrimPrefix(string(uri), fileURIPrefix)
	return strings.TrimPrefix(fullPath, h.rootPath+string(os.PathSeparator))
}

func (h *handler) appendFilePrefix(relPathInProject string) lsp.DocumentURI {
	fullPath := filepath.Join(h.rootPath, relPathInProject)
	return lsp.DocumentURI(fileURIPrefix + fullPath)
}

func (h *handler) handleNop(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (any, error) {
	return nil, nil
}

type initializeResult struct {
	Capabilities serverCapabilities `json:"capabilities"`
}

type serverCapabilities struct {
	TextDocumentSync   lsp.TextDocumentSyncOptions `json:"textDocumentSync"`
	DiagnosticProvider diagnosticProvider          `json:"diagnosticProvider"`
	ReferencesProvider bool                        `json:"referencesProvider"`
}

type diagnosticProvider struct {
	InterFileDependencies bool `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool `json:"workspaceDiagnostics"`
}

func (h *handler) handleInitialize(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params lsp.InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	h.rootPath = strings.TrimPrefix(string(params.RootURI), fileURIPrefix)

	return initializeResult{
		Capabilities: serverCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    lsp.TDSKFull,
			},
			ReferencesProvider: true,
		},
	}, nil
}

func (h *handler) handleTextDocumentDidOpen(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	filePath := h.trimFilePrefix(params.TextDocument.URI)
	h.openFiles.Store(filePath, params.TextDocument.Text)
	h.filesCache.Forget(filePath)

	// Update calculation cache
	h.notifyAnalysisForPath(filePath)

	return nil, nil
}

func (h *handler) handleTextDocumentDidChange(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	filePath := h.trimFilePrefix(params.TextDocument.URI)
	h.openFiles.Store(filePath, params.ContentChanges[0].Text)
	h.filesCache.Forget(filePath)

	// Update calculation cache
	h.notifyAnalysisForPath(filePath)

	return nil, nil
}

func (h *handler) handleTextDocumentDidClose(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params lsp.DidCloseTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	filePath := h.trimFilePrefix(params.TextDocument.URI)
	h.openFiles.Delete(filePath)
	h.filesCache.Forget(filePath)

	return nil, nil
}

func (h *handler) notifyAnalysisForPath(filePath string) {
	gitPath, ok := getGitRoot(h.rootPath, filePath)
	if ok {
		gitFullPath := strings.Join(gitPath, string(os.PathSeparator))
		h.debouncedAnalyze(gitFullPath)
	}
}

type textDocumentDiagnosticParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type textDocumentDiagnosticReport struct {
	Kind  string            `json:"kind"`
	Items []*lsp.Diagnostic `json:"items"`
}

func (h *handler) handleTextDocumentDiagnostic(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params textDocumentDiagnosticParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	// For some reason, VSCode tries to retrieve diagnostics via textDocument/diagnostic request, even though
	// we do not reply with corresponding server capabilities.
	// Just a plumbing implementation so there would be no noisy errors on screen for VSCode.
	// Returning 0 items here is okay because diagnostics pushed by textDocument/publishDiagnostics is displayed
	// together in VSCode.
	return textDocumentDiagnosticReport{
		Kind:  "full",
		Items: make([]*lsp.Diagnostic, 0),
	}, nil
}

func (h *handler) handleTextDocumentReferences(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) ([]*lsp.Location, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params lsp.ReferenceParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	locations := make([]*lsp.Location, 0)
	filePath := h.trimFilePrefix(params.TextDocument.URI)
	gitPathElements, ok := getGitRoot(h.rootPath, filePath)
	if !ok {
		return locations, nil
	}

	gitPath := strings.Join(gitPathElements, string(os.PathSeparator))
	cloneSets, ok := h.previousAnalysis.Load(gitPath)
	if !ok {
		return locations, nil
	}

	// NOTE: LSP location's line is 0-indexed, while our clone data is recorded in 1-indexed manner.
	targetLine := params.Position.Line + 1

	toLSPLocation := func(c *domain.Clone) (*lsp.Location, error) { return h.toLSPLocation(gitPath, c) }
	// Check if target location is part of any clone set
	for _, cs := range cloneSets {
		clones := append(ds.Copy(cs.Missing), cs.Changed...)
		for _, c := range clones {
			cloneFullPath := filepath.Join(gitPath, c.Filename)
			if filePath == cloneFullPath && c.StartL <= targetLine && targetLine <= c.EndL {
				// Target location is inside a clone in this clone set
				// => Display all clone locations in this set
				locations, err := ds.MapError(clones, toLSPLocation)
				if err != nil {
					return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: err.Error()}
				}
				return locations, nil
			}
		}
	}

	// Target location was not part of any clone set
	return locations, nil
}

func (h *handler) toLSPLocation(gitPath string, clone *domain.Clone) (*lsp.Location, error) {
	detectedPath := filepath.Join(gitPath, clone.Filename)
	lines, err := h.filesCache.Get(context.Background(), detectedPath)
	if err != nil {
		return nil, err
	}
	return &lsp.Location{
		URI:   h.appendFilePrefix(detectedPath),
		Range: toLSPRange(clone, lines),
	}, nil
}

func (h *handler) handleTextDocumentCodeAction(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) (any, error) {
	// IntelliJ clients send textDocument/codeAction requests for some reason,
	// even though we have not announced codeAction capabilities.
	// For now, in order to suppress error logs, just return 'null' which is also a valid response.
	return nil, nil
}

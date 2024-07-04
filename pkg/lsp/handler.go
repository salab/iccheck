package lsp

import (
	"context"
	"encoding/json"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	h.openFiles[filePath] = params.TextDocument.Text
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
	h.openFiles[filePath] = params.ContentChanges[0].Text
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
	delete(h.openFiles, filePath)
	h.filesCache.Forget(filePath)

	return nil, nil
}

func (h *handler) notifyAnalysisForPath(filePath string) {
	gitPath, ok := getGitRoot(h.rootPath, filePath)
	if ok {
		gitFullPath := strings.Join(gitPath, string(os.PathSeparator))
		h.analyzeCache.Forget(gitFullPath)
		go func() {
			_, err := h.analyzeCache.Get(context.Background(), gitFullPath)
			if err != nil {
				slog.Warn("failed to analyze path", "path", gitFullPath, "error", err)
			}
		}()
	}
}

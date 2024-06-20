package lsp

import (
	"context"
	"encoding/json"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	"os"
	"strings"
)

const fileURIPrefix = "file://"

func (h *handler) trimFilePrefix(uri lsp.DocumentURI) string {
	fullPath := strings.TrimPrefix(string(uri), fileURIPrefix)
	return strings.TrimPrefix(fullPath, h.rootPath+string(os.PathSeparator))
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
			DiagnosticProvider: diagnosticProvider{
				InterFileDependencies: true,
				WorkspaceDiagnostics:  false,
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

	h.files[h.trimFilePrefix(params.TextDocument.URI)] = params.TextDocument.Text

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

	filename := h.trimFilePrefix(params.TextDocument.URI)
	h.files[filename] = params.ContentChanges[0].Text

	// Update calculation cache
	gitPath, ok := getGitRoot(h.rootPath, filename)
	if ok {
		h.calcCache.Forget(strings.Join(gitPath, string(os.PathSeparator)))
	}

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

	delete(h.files, h.trimFilePrefix(params.TextDocument.URI))

	return nil, nil
}

type textDocumentDiagnosticParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type textDocumentDiagnosticReport struct {
	Kind  string           `json:"kind"`
	Items []lsp.Diagnostic `json:"items"`
}

func (h *handler) handleTextDocumentDiagnostic(ctx context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	var params textDocumentDiagnosticParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	items, err := h.analyzeFile(ctx, h.trimFilePrefix(lsp.DocumentURI(params.TextDocument.URI)))
	if err != nil {
		return nil, err
	}
	return textDocumentDiagnosticReport{
		Kind:  "full",
		Items: items,
	}, nil
}

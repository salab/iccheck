// Package lsp contains JSON-RPC 2.0 implementation for Language Server Protocol.
// Code is partially copied from https://github.com/vito/bass/blob/main/cmd/bass/lsp.go.
package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type handler struct{}

func NewHandler() jsonrpc2.Handler {
	h := &handler{}
	return jsonrpc2.HandlerWithError(h.handle)
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, conn, req)
	case "textDocument/didChange":
		return h.handleTextDocumentDidChange(ctx, conn, req)
	}

	return nil, &jsonrpc2.Error{
		Code:    jsonrpc2.CodeMethodNotFound,
		Message: fmt.Sprintf("method not supported: %s", req.Method),
	}
}

type initializeResult struct {
	Capabilities serverCapabilities `json:"capabilities"`
}

type serverCapabilities struct {
	DiagnosticProvider struct{} `json:"diagnosticProvider"`
}

func (h *handler) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params lsp.InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	return initializeResult{
		Capabilities: serverCapabilities{
			DiagnosticProvider: struct{}{},
		},
	}, nil
}

func (h *handler) handleTextDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result any, err error) {
	panic("implement me")
}

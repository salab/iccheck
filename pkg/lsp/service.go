// Package lsp contains JSON-RPC 2.0 implementation for Language Server Protocol.
// Code is partially copied from https://github.com/vito/bass/blob/main/cmd/bass/lsp.go.
package lsp

import (
	"context"
	"fmt"
	"github.com/sourcegraph/jsonrpc2"
	"log/slog"
)

type handler struct {
	conn *jsonrpc2.Conn

	files map[string]string
}

func NewHandler() jsonrpc2.Handler {
	h := &handler{
		files: make(map[string]string),
	}
	return jsonrpc2.HandlerWithError(h.handle)
}

func (h *handler) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
	if h.conn == nil {
		h.conn = conn
	}

	slog.Info(fmt.Sprintf("handle(): method: %v\n", req.Method))
	if req.Params != nil {
		slog.Info(fmt.Sprintf("handle(): params: %v\n", string(*req.Params)))
	}

	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, conn, req)
	case "initialized":
		return h.handleNop(ctx, conn, req)
	case "textDocument/didOpen":
		return h.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didChange":
		return h.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/didClose":
		return h.handleNop(ctx, conn, req)
	case "textDocument/diagnostic":
		return h.handleTextDocumentDiagnostic(ctx, conn, req)
	}

	return nil, &jsonrpc2.Error{
		Code:    jsonrpc2.CodeMethodNotFound,
		Message: fmt.Sprintf("method not supported: %s", req.Method),
	}
}

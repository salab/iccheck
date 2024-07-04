package lsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

func (h *handler) readFile(_ context.Context, path string) ([]string, error) {
	if content, ok := h.openFiles[path]; ok {
		return strings.Split(content, "\n"), nil
	}

	b, err := os.ReadFile(filepath.Join(h.rootPath, path))
	if err != nil {
		return nil, err
	}
	return strings.Split(string(b), "\n"), nil
}

package lsp

import (
	"context"
	"github.com/sourcegraph/go-lsp"
	"regexp"
	"strings"
)

var testWarnRegexp = regexp.MustCompile("bad-content")

func (h *handler) analyzeFile(ctx context.Context, filename string) ([]lsp.Diagnostic, error) {
	content := h.files[filename]
	lines := strings.Split(content, "\n")

	// TODO: compute actual diagnostics
	diagnostics := make([]lsp.Diagnostic, 0)
	for i, line := range lines {
		indices := testWarnRegexp.FindAllStringIndex(line, -1)
		for _, match := range indices {
			diagnostics = append(diagnostics, lsp.Diagnostic{
				Range: lsp.Range{
					Start: lsp.Position{Line: i, Character: match[0]},
					End:   lsp.Position{Line: i, Character: match[1]},
				},
				Severity: lsp.Warning,
				Code:     "",
				Source:   "",
				Message:  "Test diagnostic",
			})
		}
	}

	return diagnostics, nil
}

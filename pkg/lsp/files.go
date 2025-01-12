package lsp

import (
	"context"
	"fmt"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
	"os"
	"path/filepath"
	"strings"
)

func (h *handler) readFile(_ context.Context, path string) ([]string, error) {
	if content, ok := h.openFiles.Load(path); ok {
		return strings.Split(content, "\n"), nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(b), "\n"), nil
}

func readablePaths(pwd string, clones []*domain.Clone, limit int) string {
	return strings.Join(
		ds.Map(ds.Limit(clones, limit), func(c *domain.Clone) string {
			relPath, err := filepath.Rel(pwd, c.Filename)
			if err != nil {
				relPath = c.Filename
			}
			prefix := lo.Ternary(relPath == ".", "", relPath+"#")
			if c.StartL == c.EndL {
				return fmt.Sprintf("%sL%d", prefix, c.StartL)
			}
			return fmt.Sprintf("%sL%d-L%d", prefix, c.StartL, c.EndL)
		}),
		", ",
	) + lo.Ternary(len(clones) > limit, ", ...", "")
}

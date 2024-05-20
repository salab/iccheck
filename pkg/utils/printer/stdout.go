package printer

import (
	"fmt"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/samber/lo"
	"log/slog"
	"path/filepath"
)

type stdoutPrinter struct{}

func NewStdoutPrinter() Printer {
	return &stdoutPrinter{}
}

func (s *stdoutPrinter) PrintClones(repoDir string, clones []domain.Clone) {
	if len(clones) == 0 {
		slog.Info(fmt.Sprintf("No clones are missing inconsistent changes."))
		return
	}

	slog.Info(fmt.Sprintf("%d clone(s) are likely missing a consistent change.", len(clones)))
	for _, c := range clones {
		path := lo.Must(filepath.Abs(filepath.Join(repoDir, c.Filename)))
		slog.Info(fmt.Sprintf("Clone %s:%d (L%d-L%d) is likely missing a consistent change.", path, c.StartL, c.StartL, c.EndL))
	}
}

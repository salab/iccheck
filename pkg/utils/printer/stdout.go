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
	for _, c := range clones {
		path := lo.Must(filepath.Abs(filepath.Join(repoDir, c.Filename)))
		slog.Info(fmt.Sprintf("Clone %s:%d (L%d-L%d) is likely missing a consistent change.", path, c.StartL, c.StartL, c.EndL))
	}
}

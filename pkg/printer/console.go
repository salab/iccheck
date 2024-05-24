package printer

import (
	"bytes"
	"fmt"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/samber/lo"
	"path/filepath"
)

type consolePrinter struct{}

func NewConsolePrinter() Printer {
	return &consolePrinter{}
}

func (s *consolePrinter) PrintClones(repoDir string, clones []domain.Clone) []byte {
	var buf bytes.Buffer
	for _, c := range clones {
		path := lo.Must(filepath.Abs(filepath.Join(repoDir, c.Filename)))
		buf.WriteString(
			fmt.Sprintf("Clone %s:%d (L%d-L%d, distance %f) is likely missing a consistent change.\n", path, c.StartL, c.StartL, c.EndL, c.Distance),
		)
	}
	return buf.Bytes()
}

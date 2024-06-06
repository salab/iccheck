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

func (s *consolePrinter) cloneToStr(repoDir string, c *domain.Clone) string {
	path := lo.Must(filepath.Abs(filepath.Join(repoDir, c.Filename)))
	return fmt.Sprintf("%s:%d (L%d-L%d)", path, c.StartL, c.StartL, c.EndL)
}

func (s *consolePrinter) PrintClones(repoDir string, sets []*domain.CloneSet) []byte {
	var buf bytes.Buffer
	for i, set := range sets {
		buf.WriteString("\n")
		buf.WriteString(
			fmt.Sprintf("Clone set #%d - %d out of %d clones are likely missing consistent change(s).\n", i, len(set.Missing), len(set.Missing)+len(set.Changed)),
		)
		buf.WriteString(fmt.Sprintf("  Missing changes (%d):\n", len(set.Missing)))
		for _, c := range set.Missing {
			buf.WriteString("    - " + s.cloneToStr(repoDir, c) + "\n")
		}
		buf.WriteString(fmt.Sprintf("  Changed clones (%d):\n", len(set.Changed)))
		for _, c := range set.Changed {
			buf.WriteString("    - " + s.cloneToStr(repoDir, c) + "\n")
		}
	}
	return buf.Bytes()
}

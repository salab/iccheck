package printer

import (
	"bytes"
	"fmt"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/utils/ds"
	"strings"
)

const githubPrintLimit = 3

// githubPrinter prints output in GitHub annotations compatible format.
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
type githubPrinter struct{}

func NewGitHubPrinter() Printer {
	return &githubPrinter{}
}

func (g *githubPrinter) PrintClones(_ string, clones []domain.Clone) []byte {
	var buf bytes.Buffer
	exceedLimit := len(clones) > githubPrintLimit
	if exceedLimit {
		buf.WriteString(fmt.Sprintf("Warn: Many (%d) inconsistent changes detected. Only displaying the top %d.\n", len(clones), githubPrintLimit))
	}
	for _, c := range ds.FirstN(clones, githubPrintLimit) {
		buf.WriteString(
			fmt.Sprintf("::notice file=%s,line=%d,endLine=%d,title=%s::%s\n",
				c.Filename,
				c.StartL,
				c.EndL,
				"Possibly missing change",
				fmt.Sprintf(
					"Possibly missing a consistent change here (L%d - L%d, distance %f) (%s)",
					c.StartL, c.EndL, c.Distance,
					// NOTE: is there any way to output multiline annotation?
					strings.Join(
						ds.Map(c.Sources, func(s *domain.Source) string {
							return fmt.Sprintf("deduced from change at %s:%d (L%d - L%d)", s.Filename, s.StartL, s.StartL, s.EndL)
						}),
						",",
					),
				),
			),
		)
	}
	return buf.Bytes()
}

package printer

import (
	"bytes"
	"fmt"
	"github.com/salab/iccheck/pkg/domain"
)

// githubPrinter prints output in GitHub annotations compatible format.
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
type githubPrinter struct{}

func NewGitHubPrinter() Printer {
	return &githubPrinter{}
}

func (g *githubPrinter) PrintClones(sets []*domain.CloneSet) []byte {
	var buf bytes.Buffer

	for _, set := range sets {
		for _, c := range set.Missing {
			buf.WriteString(
				fmt.Sprintf("::notice file=%s,line=%d,endLine=%d,title=%s::%s\n",
					c.Filename,
					c.StartL,
					c.EndL,
					"Possibly missing change",
					fmt.Sprintf(
						"Possibly missing a consistent change here (L%d - L%d) (%d / %d clone(s) in this clone set were changed)",
						c.StartL, c.EndL,
						len(set.Changed), len(set.Changed)+len(set.Missing),
					),
				),
			)
		}
	}

	return buf.Bytes()
}

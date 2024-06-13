package printer

import (
	"bytes"
	"fmt"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/utils/ds"
)

const githubPrintLimit = 3

// githubPrinter prints output in GitHub annotations compatible format.
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
type githubPrinter struct{}

func NewGitHubPrinter() Printer {
	return &githubPrinter{}
}

func (g *githubPrinter) PrintClones(_ string, sets []*domain.CloneSet) []byte {
	var buf bytes.Buffer

	missingChanges := ds.FlatMap(sets, func(cs *domain.CloneSet) []*domain.Clone { return cs.Missing })
	exceedLimit := len(missingChanges) > githubPrintLimit
	if exceedLimit {
		buf.WriteString(fmt.Sprintf("Warn: Many (%d) inconsistent changes detected. Only displaying the top %d.\n", len(missingChanges), githubPrintLimit))
	}

	var limit int
outer:
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

			limit++
			if limit >= githubPrintLimit {
				break outer
			}
		}
	}
	return buf.Bytes()
}

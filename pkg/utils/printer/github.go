package printer

import (
	"fmt"
	"github.com/salab/iccheck/pkg/domain"
	"log/slog"
)

const githubPrintLimit = 3

type githubPrinter struct{}

func NewGitHubPrinter() Printer {
	return &githubPrinter{}
}

func (g *githubPrinter) PrintClones(_ string, clones []domain.Clone) {
	exceedLimit := len(clones) > githubPrintLimit
	if exceedLimit {
		slog.Warn(fmt.Sprintf("Many (%d) inconsistent changes detected. Only displaying the top %d.", len(clones), githubPrintLimit))
	}
	for _, c := range clones[:githubPrintLimit] {
		fmt.Printf("::notice file=%s,line=%d,endLine=%d,title=%s::%s\n",
			c.Filename,
			c.StartL,
			c.EndL,
			"Possibly missing change",
			fmt.Sprintf("Possibly missing a consistent change here (L%d - L%d, distance %f)", c.StartL, c.EndL, c.Distance),
		)
	}
}

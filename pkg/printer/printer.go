package printer

import "github.com/salab/iccheck/pkg/domain"

type Printer interface {
	PrintClones(repoDir string, clones []domain.Clone) []byte
}

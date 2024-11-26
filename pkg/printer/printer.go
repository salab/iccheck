package printer

import "github.com/salab/iccheck/pkg/domain"

type Printer interface {
	PrintClones(sets []*domain.CloneSet) []byte
}

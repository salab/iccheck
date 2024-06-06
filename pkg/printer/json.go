package printer

import (
	"bytes"
	"encoding/json"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
)

type jsonPrinter struct{}

func NewJsonPrinter() Printer {
	return &jsonPrinter{}
}

type jsonSource struct {
	Filename string `json:"filename"`
	StartL   int    `json:"start_l"`
	EndL     int    `json:"end_l"`
}

type jsonClone struct {
	BaseDir  string        `json:"base_dir"`
	Filename string        `json:"filename"`
	StartL   int           `json:"start_l"`
	EndL     int           `json:"end_l"`
	Distance float64       `json:"distance"`
	Sources  []*jsonSource `json:"sources"`
}

type jsonCloneSet struct {
	Missing []*jsonClone `json:"missing"`
	Changed []*jsonClone `json:"changed"`
}

func (j *jsonPrinter) formatClone(repoDir string, c *domain.Clone) *jsonClone {
	return &jsonClone{
		BaseDir:  repoDir,
		Filename: c.Filename,
		StartL:   c.StartL,
		EndL:     c.EndL,
		Distance: c.Distance,
		Sources: ds.Map(c.Sources, func(s *domain.Source) *jsonSource {
			return &jsonSource{
				Filename: s.Filename,
				StartL:   s.StartL,
				EndL:     s.EndL,
			}
		}),
	}
}

func (j *jsonPrinter) formatCloneSet(repoDir string, set *domain.CloneSet) *jsonCloneSet {
	return &jsonCloneSet{
		Missing: ds.Map(set.Missing, func(c *domain.Clone) *jsonClone { return j.formatClone(repoDir, c) }),
		Changed: ds.Map(set.Changed, func(c *domain.Clone) *jsonClone { return j.formatClone(repoDir, c) }),
	}
}

func (j *jsonPrinter) PrintClones(repoDir string, sets []*domain.CloneSet) []byte {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	for _, set := range sets {
		obj := j.formatCloneSet(repoDir, set)
		lo.Must0(encoder.Encode(obj))
	}
	return buf.Bytes()
}

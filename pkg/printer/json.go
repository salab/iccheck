package printer

import (
	"bytes"
	"encoding/json"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/samber/lo"
)

type jsonPrinter struct{}

func NewJsonPrinter() Printer {
	return &jsonPrinter{}
}

type jsonOutput struct {
	BaseDir  string  `json:"base_dir"`
	Filename string  `json:"filename"`
	StartL   int     `json:"start_l"`
	EndL     int     `json:"end_l"`
	Distance float64 `json:"distance"`
}

func (j *jsonPrinter) format(repoDir string, c domain.Clone) jsonOutput {
	return jsonOutput{
		BaseDir:  repoDir,
		Filename: c.Filename,
		StartL:   c.StartL,
		EndL:     c.EndL,
		Distance: c.Distance,
	}
}

func (j *jsonPrinter) PrintClones(repoDir string, clones []domain.Clone) []byte {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	for _, c := range clones {
		obj := j.format(repoDir, c)
		lo.Must0(encoder.Encode(obj))
	}
	return buf.Bytes()
}

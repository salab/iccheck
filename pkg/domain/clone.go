package domain

import (
	"fmt"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/salab/iccheck/pkg/utils/files"
	"github.com/samber/lo"
	"slices"
)

type Source struct {
	Filename string
	StartL   int
	EndL     int
}

func (s Source) Key() string {
	return fmt.Sprintf("%s-%d-%d", s.Filename, s.StartL, s.EndL)
}

// SlideCut cuts this source into multiple sources, sliding by the specified window size.
func (s Source) SlideCut(window int) []*Source {
	startL, endL := s.StartL, s.EndL
	lines := endL - startL + 1
	if lines <= window {
		return []*Source{
			{s.Filename, startL, endL},
		}
	}
	ret := make([]*Source, 0, lines-1)
	for i := 0; i < lines-(window-1); i++ {
		ret = append(ret, &Source{
			Filename: s.Filename,
			StartL:   startL + i,
			EndL:     startL + i + (window - 1),
		})
	}
	return ret
}

type Clone struct {
	Filename string
	StartL   int
	EndL     int
	Distance float64
	// Sources indicate from which queries this co-change candidate was detected
	Sources []*Source
}

func (c Clone) Key() string {
	return fmt.Sprintf("%v-%d-%d", c.Filename, c.StartL, c.EndL)
}

type CloneSet struct {
	Changed []*Clone
	Missing []*Clone
}

func (cs *CloneSet) Sort() {
	// Use file proximity ranking from FLeCCS
	// NOTE: Maybe we can utilize simple clone tracking to improve suggestion accuracy?
	patchPaths := ds.Map(cs.Changed, func(c *Clone) string { return c.Filename })
	slices.SortFunc(cs.Missing, ds.SortCompose(
		ds.SortAsc(func(c *Clone) int {
			distances := ds.Map(patchPaths, func(path string) int { return files.FileTreeDistance(path, c.Filename) })
			return lo.Min(distances)
		}),
		ds.SortAsc(func(c *Clone) float64 { return c.Distance }),
	))
}

func (cs *CloneSet) ChangedProportion() float64 {
	changed := len(cs.Changed)
	missing := len(cs.Missing)
	return float64(changed) / float64(missing+changed)
}

func SortCloneSets(sets []*CloneSet) {
	// Sort from sets that is most likely missing consistent changes
	slices.SortFunc(sets, ds.SortAsc(func(cs *CloneSet) int { return len(cs.Missing) }))
}

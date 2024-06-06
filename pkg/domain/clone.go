package domain

import (
	"fmt"
	"github.com/salab/iccheck/pkg/fleccs"
	"github.com/salab/iccheck/pkg/utils/ds"
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
			distances := ds.Map(patchPaths, func(path string) int { return fleccs.FileTreeDistance(path, c.Filename) })
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
	slices.SortFunc(sets, ds.SortDesc((*CloneSet).ChangedProportion))
}

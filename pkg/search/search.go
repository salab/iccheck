package search

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/pkg/errors"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
	"github.com/theodesp/unionfind"
	"log/slog"
	"slices"
	"strings"
)

type changeKind string

var (
	changeKindEqual        = changeKind("equal")
	changeKindAddition     = changeKind("addition")
	changeKindDeletion     = changeKind("deletion")
	changeKindModification = changeKind("modification")
)

type chunk struct {
	beforeFilename string
	afterFilename  string

	kind changeKind

	beforeStartL int
	beforeEndL   int
	afterStartL  int
	afterEndL    int
}

func (c *chunk) searchQuery() *domain.Source {
	return &domain.Source{
		Filename: c.afterFilename,
		StartL:   c.afterStartL,
		EndL:     c.afterEndL,
	}
}

type chunkTracker struct {
	beforeFilename string
	afterFilename  string

	chunks []*chunk
}

func (t *chunkTracker) recordChunk(kind changeKind, beforeStartL, beforeEndL, afterStartL, afterEndL int) {
	t.chunks = append(t.chunks, &chunk{
		beforeFilename: t.beforeFilename,
		afterFilename:  t.afterFilename,

		kind: kind,

		beforeStartL: beforeStartL,
		beforeEndL:   beforeEndL,
		afterStartL:  afterStartL,
		afterEndL:    afterEndL,
	})
}

// patches give all recorded chunks that is one of addition, deletion, or modification
func (t *chunkTracker) patches() []*chunk {
	return lo.Filter(t.chunks, func(c *chunk, _ int) bool {
		return c.kind != changeKindEqual
	})
}

// dedupe overlapping search result
func dedupeDetectedClones(clones []*domain.Clone) (deduped []*domain.Clone) {
	slices.SortFunc(clones, ds.SortCompose(
		ds.SortAsc(func(c *domain.Clone) string { return c.Filename }),
		ds.SortAsc(func(c *domain.Clone) int { return c.StartL }),
	))

	// First, coalesce overlapping detected clones, to make mutually exclusive 'clone' zones
	deduped = make([]*domain.Clone, 0, len(clones))
	{
		startIdx := 0
		for i := 0; i < len(clones); i++ {
			nextCoalesce := i+1 < len(clones) && // is not last in loop
				clones[i].Filename == clones[i+1].Filename && // in the same file
				clones[i+1].StartL <= clones[i].EndL // is overlapping

			c := clones[i]
			if !nextCoalesce {
				// record coalesced clone
				distanceSum := lo.SumBy(clones[startIdx:i+1], func(c *domain.Clone) float64 { return c.Distance })
				cloneSources := lo.UniqBy(
					ds.FlatMap(clones[startIdx:i+1], func(c *domain.Clone) []*domain.Source { return c.Sources }),
					func(s *domain.Source) string { return s.Key() },
				)
				clone := &domain.Clone{
					Filename: c.Filename,
					StartL:   clones[startIdx].StartL,
					EndL:     c.EndL,
					Distance: distanceSum / float64(i-startIdx+1),
					Sources:  cloneSources,
				}

				deduped = append(deduped, clone)

				startIdx = i + 1
			}
		}
	}

	return
}

func findCloneSets(clones []*domain.Clone, sources []*domain.Source) (sets [][]*domain.Clone) {
	// Now that we have deduped clones,
	// we can convert all source (query) locations to match the granularity of cloned lines.
	matchedSources := make(map[string][]*domain.Source, len(sources))
	{
		clonesByFilename := lo.GroupBy(clones, func(c *domain.Clone) string { return c.Filename })
		for _, src := range sources {
			clonesInFile := clonesByFilename[src.Filename]
			for _, cl := range clonesInFile {
				// If they overlap, they 'match'.
				hasOverlap := cl.StartL <= src.EndL && src.StartL <= cl.EndL
				if hasOverlap {
					matchedSources[src.Key()] = append(matchedSources[src.Key()], &domain.Source{
						Filename: cl.Filename,
						StartL:   cl.StartL,
						EndL:     cl.EndL,
					})
				}
			}
		}
	}

	// We can start converting sources to clone granularity, and start building a Union-Find tree to find clone-sets.
	cloneKeyToIdx := make(map[string]int, len(clones))
	{
		for i, c := range clones {
			cloneKeyToIdx[c.Key()] = i
		}
	}
	uf := unionfind.New(len(clones))
	for i, c := range clones {
		for _, src := range c.Sources {
			matched := matchedSources[src.Key()]
			for _, m := range matched {
				j := cloneKeyToIdx[m.Key()]
				uf.Union(i, j)
			}
		}
	}
	setByRootID := make(map[int][]*domain.Clone)
	for i, clone := range clones {
		root := uf.Root(i)
		setByRootID[root] = append(setByRootID[root], clone)
	}

	return lo.Values(setByRootID)
}

func filterMissingChanges(cloneSets [][]*domain.Clone, actualPatches []*chunk) []*domain.CloneSet {
	perFile := lo.GroupBy(actualPatches, func(c *chunk) string { return c.searchQuery().Filename })
	isChanged := func(c *domain.Clone) bool {
		patchChunks := perFile[c.Filename]
		// Is the clone changed by any of the patch chunks?
		for _, p := range patchChunks {
			q := p.searchQuery()
			hasOverlap := c.StartL <= q.EndL && q.StartL <= c.EndL
			if hasOverlap {
				return true
			}
		}
		// If not, the clone is missing consistent change
		return false
	}
	return ds.Map(cloneSets, func(set []*domain.Clone) *domain.CloneSet {
		var cs domain.CloneSet
		for _, c := range set {
			if isChanged(c) {
				cs.Changed = append(cs.Changed, c)
			} else {
				cs.Missing = append(cs.Missing, c)
			}
		}
		return &cs
	})
}

func Search(
	ctx context.Context,
	algorithmName string,
	fromTree, toTree domain.Tree,
	ignore domain.IgnoreRules,
) ([]*domain.CloneSet, error) {
	// Compare the trees
	filePatches, err := domain.DiffTrees(fromTree, toTree)
	if err != nil {
		return nil, errors.Wrap(err, "diffing trees")
	}

	// Calculate the chunks
	var patchChunks []*chunk
	for _, filePatch := range filePatches {
		from, to := filePatch.Files()
		slog.Debug("File patch", "from", from, "to", to)

		// Only handle non-binary patches
		if filePatch.IsBinary() {
			continue
		}

		// Only handle file modifications (or renames) and additions
		// When handling file additions, target lines are the added lines itself
		if to == nil {
			// File was deleted
			continue
		}

		tracker := &chunkTracker{afterFilename: to.Path()}
		if from != nil {
			tracker.beforeFilename = from.Path()
		}

		// Categorize file patch chunks (equal, add, delete) into
		// patch categories (add, delete, modification)
		chunks := filePatch.Chunks()
		fromFileL, toFileL := 1, 1
		for i := 0; i < len(chunks); i++ {
			// Did the chunk undergo a modification?
			if i+1 < len(chunks) &&
				((chunks[i].Type() == diff.Add && chunks[i+1].Type() == diff.Delete) ||
					(chunks[i].Type() == diff.Delete && chunks[i+1].Type() == diff.Add)) {
				addition := lo.Ternary(chunks[i].Type() == diff.Add, chunks[i], chunks[i+1])
				deletion := lo.Ternary(chunks[i].Type() == diff.Add, chunks[i+1], chunks[i])

				additionLines := strings.Count(addition.Content(), "\n")
				deletionLines := strings.Count(deletion.Content(), "\n")

				tracker.recordChunk(
					changeKindModification,
					fromFileL, fromFileL+deletionLines-1,
					toFileL, toFileL+additionLines-1,
				)

				fromFileL += deletionLines
				toFileL += additionLines

				i++ // Advance i by 2 in next loop
				continue
			}

			chunk := chunks[i]
			lines := strings.Count(chunk.Content(), "\n")
			switch chunk.Type() {
			case diff.Equal:
				tracker.recordChunk(
					changeKindEqual,
					fromFileL, fromFileL+lines-1,
					toFileL, toFileL+lines-1,
				)

				fromFileL += lines
				toFileL += lines
			case diff.Add:
				tracker.recordChunk(
					changeKindAddition,
					fromFileL-1, fromFileL-1,
					toFileL, toFileL+lines-1,
				)

				toFileL += lines
			case diff.Delete:
				tracker.recordChunk(
					changeKindDeletion,
					fromFileL, fromFileL+lines-1,
					toFileL-1, toFileL-1,
				)

				fromFileL += lines
			default:
				panic(fmt.Sprintf("unknown chunk type: %v", chunk.Type()))
			}
		}

		patchChunks = append(patchChunks, tracker.patches()...)
	}

	// Prepare queries
	queries := ds.Map(patchChunks, (*chunk).searchQuery)
	// Cut diffs into 3 lines for detecting micro-clones
	/*
		queries := ds.FlatMap(queries, func(c *domain.Source) []*domain.Source {
			return c.SlideCut(3)
		})
	*/

	// Search for clones
	slog.Info(fmt.Sprintf("%d change chunk(s) within %d file(s) found.", len(queries), len(filePatches)), "from", fromTree, "to", toTree)
	for i, q := range queries {
		slog.Debug(fmt.Sprintf("Query#%d", i), "query", q)
	}
	algorithmFn, ok := algorithms[algorithmName]
	if !ok {
		return nil, fmt.Errorf("invalid algorithm name: %v", algorithmName)
	}
	toSearcher := domain.NewSearcherFromTree(toTree)
	toClones, err := algorithmFn(ctx, toSearcher, queries, toSearcher, ignore)
	if err != nil {
		return nil, errors.Wrap(err, "searching for clones")
	}

	// Deduplicate overlapping clones
	toClones = dedupeDetectedClones(toClones)

	// Calculate clone sets
	rawCloneSets := findCloneSets(toClones, queries)

	// Calculate inconsistent changes by listing clones not modified by this patch
	cloneSets := filterMissingChanges(rawCloneSets, patchChunks)

	// Filter size 1 "clone sets" - this is included in the calculation result of this algorithm, but not really "clone sets"
	cloneSets = lo.Filter(cloneSets, func(cs *domain.CloneSet, _ int) bool { return len(cs.Missing)+len(cs.Changed) > 1 })

	// Sort
	domain.SortCloneSets(cloneSets)
	for _, set := range cloneSets {
		set.Sort()
	}

	// Return the inconsistent changes found
	return cloneSets, nil
}

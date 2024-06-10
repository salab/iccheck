package search

import (
	"context"
	"fmt"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
	"github.com/theodesp/unionfind"
	"log/slog"
	"os"
	"slices"
	"strings"
)

func newTmpOSFS(repo *git.Repository) (wt *git.Worktree, path string, clean func()) {
	tmpDir := lo.Must(os.MkdirTemp("", "osfs-"))
	wt = lo.Must(repo.Worktree())
	wt.Filesystem = osfs.New(tmpDir)
	return wt, tmpDir, func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			slog.Error("failed to delete tmp osfs", "err", err)
		}
	}
}

type changeKind string

var (
	changeKindEqual        = changeKind("equal")
	changeKindAddition     = changeKind("addition")
	changeKindDeletion     = changeKind("deletion")
	changeKindModification = changeKind("modification")
)

type chunk struct {
	filename string
	kind     changeKind

	beforeStartL int
	beforeEndL   int
	afterStartL  int
	afterEndL    int
}

func (c *chunk) searchQueryLines() (startL, endL int) {
	// These 'before' fields are recorded so that they match the search query lines
	return c.beforeStartL, c.beforeEndL
}

type chunkTracker struct {
	filename string
	chunks   []*chunk
}

func (t *chunkTracker) recordChunk(kind changeKind, beforeStartL, beforeEndL, afterStartL, afterEndL int) {
	t.chunks = append(t.chunks, &chunk{
		filename: t.filename,
		kind:     kind,

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
					lo.FlatMap(clones[startIdx:i+1], func(c *domain.Clone, _ int) []*domain.Source { return c.Sources }),
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
	perFile := lo.GroupBy(actualPatches, func(p *chunk) string { return p.filename })
	isChanged := func(c *domain.Clone) bool {
		filePatches := perFile[c.Filename]
		// Is the clone changed by any of the patch chunks?
		for _, p := range filePatches {
			startL, endL := p.searchQueryLines()
			hasOverlap := c.StartL <= endL && startL <= c.EndL
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

const worktreeRef = "WORKTREE"

func Search(repoDir, fromRef, toRef string) ([]*domain.CloneSet, error) {
	slog.Info("Searching for inconsistent changes...", "repository", repoDir, "from", fromRef, "to", toRef)

	// Prepare repository
	repo := lo.Must(git.PlainOpen(repoDir))

	// Get diff chunks from source to target
	fromHash := lo.Must(repo.ResolveRevision(plumbing.Revision(fromRef)))
	fromCommit := lo.Must(repo.CommitObject(*fromHash))
	slog.Info(fmt.Sprintf("Git ref (from): %v", fromCommit))

	var filePatches []diff.FilePatch
	if toRef == worktreeRef {
		// Special handling when to-ref is set to worktree
		// as go-git does not allow getting diff from commit to worktree
		slog.Info(fmt.Sprintf("Git ref (to): %s", toRef))
		workTree := lo.Must(repo.Worktree())
		changes := diffCommitToWorktree(fromCommit, workTree)
		slog.Info("File changes detected", "files", len(changes))

		filePatches = ds.Map(changes, func(c merkletrie.Change) diff.FilePatch {
			return diffChunksToWorktree(c, fromCommit, workTree)
		})
	} else {
		// Normal diff (and rename) detection using go-git's diff feature
		toHash := lo.Must(repo.ResolveRevision(plumbing.Revision(toRef)))
		toCommit := lo.Must(repo.CommitObject(*toHash))
		slog.Info(fmt.Sprintf("Git ref (to): %v", toCommit))

		fromTree := lo.Must(fromCommit.Tree())
		toTree := lo.Must(toCommit.Tree())

		changes := lo.Must(object.DiffTreeWithOptions(context.TODO(), fromTree, toTree, object.DefaultDiffTreeOptions))
		slog.Info("File changes detected", "files", len(changes))

		filePatches = lo.FlatMap(changes, func(c *object.Change, index int) []diff.FilePatch {
			return lo.Must(c.Patch()).FilePatches()
		})
	}

	chunkTrackers := make(map[string]*chunkTracker)
	for _, filePatch := range filePatches {
		from, to := filePatch.Files()
		// Only handle file modifications (or renames)
		// TODO: handle file deletion (target lines are deleted lines)
		// TODO: handle file addition (target line to be determined)
		if from == nil || to == nil {
			continue
		}

		// Categorize file patch chunks (equal, add, delete) into
		// patch categories (add, delete, modification)
		chunks := filePatch.Chunks()
		chunkTrackers[from.Path()] = &chunkTracker{filename: from.Path()}
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

				chunkTrackers[from.Path()].recordChunk(
					changeKindModification,
					fromFileL, fromFileL+deletionLines-1,
					toFileL, toFileL+additionLines-1,
				)

				fromFileL += deletionLines
				toFileL += additionLines

				i++ // Increase i by 2
				continue
			}

			chunk := chunks[i]
			lines := strings.Count(chunk.Content(), "\n")
			switch chunk.Type() {
			case diff.Equal:
				chunkTrackers[from.Path()].recordChunk(
					changeKindEqual,
					fromFileL, fromFileL+lines-1,
					toFileL, toFileL+lines-1,
				)

				fromFileL += lines
				toFileL += lines
			case diff.Add:
				chunkTrackers[to.Path()].recordChunk(
					changeKindAddition,
					fromFileL-1, fromFileL-1,
					toFileL, toFileL+lines-1,
				)

				toFileL += lines
			case diff.Delete:
				chunkTrackers[from.Path()].recordChunk(
					changeKindDeletion,
					fromFileL, fromFileL+lines-1,
					toFileL-1, toFileL-1,
				)

				fromFileL += lines
			default:
				panic(fmt.Sprintf("unknown chunk type: %v", chunk.Type()))
			}
		}
	}

	patchChunks := lo.FlatMap(lo.Values(chunkTrackers), func(c *chunkTracker, _ int) []*chunk {
		return c.patches()
	})
	slog.Info("Detected patch chunks", "patches", len(patchChunks))
	for _, patch := range patchChunks {
		slog.Debug(fmt.Sprintf("%v", patch))
	}

	// Cut diffs into 3 lines for detecting micro-clones
	/*
		// queryPatch represents query source lines
		type queryPatch struct {
			filename string
			startL   int
			endL     int
		}
		microPatches := lo.FlatMap(patchChunks, func(c *chunk, _ int) []*queryPatch {
			const chunkLines = 3
			startL, endL := c.searchQueryLines()
			lines := endL - startL + 1
			if lines <= chunkLines {
				return []*queryPatch{
					{c.filename, startL, endL},
				}
			}
			ret := make([]*queryPatch, 0, lines-1)
			for i := 0; i < lines-(chunkLines-1); i++ {
				ret = append(ret, &queryPatch{
					filename: c.filename,
					startL:   startL + i,
					endL:     startL + i + (chunkLines - 1),
				})
			}
			return ret
		})
	*/

	// Checkout
	// TODO: fix the problem that original worktree is disturbed after the operation
	origWT := lo.Must(repo.Worktree())
	origHeadCommit := lo.Must(repo.ResolveRevision("HEAD"))
	fromWT, fromWTPath, clean1 := newTmpOSFS(repo)
	defer clean1()
	lo.Must0(fromWT.Reset(&git.ResetOptions{Commit: *fromHash, Mode: git.HardReset}))
	//toWT, toWTPath, clean2 := newTmpOSFS(repo)
	//defer clean2()
	//lo.Must0(toWT.Reset(&git.ResetOptions{Commit: *toHash, Mode: git.HardReset}))
	defer func() {
		err := origWT.Reset(&git.ResetOptions{Commit: *origHeadCommit, Mode: git.MixedReset})
		if err != nil {
			slog.Error("Failed to reset original worktree to original HEAD", "err", err)
		}
	}()

	// Search for clones including the diff, in each snapshot
	slog.Info(fmt.Sprintf("Searching for clones corresponding to %d chunks...", len(patchChunks)))
	queries := ds.Map(patchChunks, func(c *chunk) *domain.Source {
		return &domain.Source{
			Filename: c.filename,
			StartL:   c.beforeStartL,
			EndL:     c.beforeEndL,
		}
	})
	fromClones := fleccsSearchMulti(fromWTPath, queries, fromWTPath)

	// Deduplicate overlapping clones
	fromClones = dedupeDetectedClones(fromClones)

	// Calculate clone sets
	rawCloneSets := findCloneSets(fromClones, queries)

	// Calculate inconsistent changes by listing clones not modified by this patch
	cloneSets := filterMissingChanges(rawCloneSets, patchChunks)
	// If all clones in a set went through some changes, no need to notify
	cloneSets = lo.Filter(cloneSets, func(cs *domain.CloneSet, _ int) bool { return len(cs.Missing) > 0 })

	// Sort
	domain.SortCloneSets(cloneSets)
	for _, set := range cloneSets {
		set.Sort()
	}

	// Return the inconsistent changes found
	return cloneSets, nil
}

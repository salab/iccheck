package search

import (
	"fmt"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/salab/iccheck/pkg/domain"
	"github.com/salab/iccheck/pkg/fleccs"
	"github.com/salab/iccheck/pkg/utils/ds"
	"github.com/samber/lo"
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
func dedupeDetectedClones(clones []domain.Clone) []domain.Clone {
	slices.SortFunc(clones, ds.SortCompose(
		ds.SortAsc(func(c domain.Clone) string { return c.Filename }),
		ds.SortAsc(func(c domain.Clone) int { return c.StartL }),
	))

	deduped := make([]domain.Clone, 0, len(clones))
	startIdx := 0
	for i := 0; i < len(clones); i++ {
		nextCoalesce := i+1 < len(clones) && // is not last in loop
			clones[i].Filename == clones[i+1].Filename && // in the same file
			clones[i+1].StartL <= clones[i].EndL // is overlapping

		c := clones[i]
		if !nextCoalesce {
			// record coalesced clone
			distanceSum := lo.SumBy(clones[startIdx:i+1], func(c domain.Clone) float64 { return c.Distance })
			deduped = append(deduped, domain.Clone{
				Filename: c.Filename,
				StartL:   clones[startIdx].StartL,
				EndL:     c.EndL,
				Distance: distanceSum / float64(i-startIdx+1),
			})

			startIdx = i + 1
		}
	}
	return deduped
}

func filterMissingChanges(clones []domain.Clone, actualPatches []*chunk) []domain.Clone {
	perFile := lo.GroupBy(actualPatches, func(p *chunk) string { return p.filename })
	return lo.Filter(clones, func(c domain.Clone, _ int) bool {
		filePatches := perFile[c.Filename]
		// Is the clone changed by any of the patch chunks?
		for _, p := range filePatches {
			startL, endL := p.searchQueryLines()
			hasOverlap := c.StartL <= endL && startL <= c.EndL
			if hasOverlap {
				return false
			}
		}
		// If not, the clone is missing consistent change
		return true
	})
}

func Search(repoDir, fromRef, toRef string) ([]domain.Clone, error) {
	slog.Info("Searching for inconsistent changes...", "repository", repoDir, "from", fromRef, "to", toRef)

	// Prepare repository
	repo := lo.Must(git.PlainOpen(repoDir))

	// Get diff chunks from source to target
	fromHash := lo.Must(repo.ResolveRevision(plumbing.Revision(fromRef)))
	toHash := lo.Must(repo.ResolveRevision(plumbing.Revision(toRef)))

	fromCommit := lo.Must(repo.CommitObject(*fromHash))
	toCommit := lo.Must(repo.CommitObject(*toHash))
	slog.Info("Base git ref (from)")
	fmt.Println(fromCommit)
	slog.Info("Base git ref (to)")
	fmt.Println(toCommit)

	// Calculate diff
	fromTree := lo.Must(repo.TreeObject(fromCommit.TreeHash))
	toTree := lo.Must(repo.TreeObject(toCommit.TreeHash))

	diffs := lo.Must(fromTree.Diff(toTree))
	slog.Info("Diffs detected", "files", len(diffs))
	filePatches := lo.FlatMap(diffs, func(diff *object.Change, index int) []diff.FilePatch {
		return lo.Must(diff.Patch()).FilePatches()
	})
	// Or equally,
	// filePatches := lo.Must(fromTree.Patch(toTree)).FilePatches()

	// TODO: check if go-git's diff handles file renames
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

	// queryPatch represents query source lines
	type queryPatch struct {
		filename string
		startL   int
		endL     int
	}
	// Cut diffs into 3 lines for detecting micro-clones
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

	// Checkout
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
	var fromClones []domain.Clone
	for _, patch := range microPatches {
		slog.Debug("Searching clones", "patch", patch)
		//clones := ncdSearchOriginal(patch, basePath)
		//clones := ncdSearchReImpl(patch, basePath)
		clones := fleccsSearch(fromWTPath, patch.filename, patch.startL, patch.endL)

		slog.Debug(fmt.Sprintf("%d clones detected", len(clones)))
		fromClones = append(fromClones, clones...)
	}

	// Deduplicate overlapping clones
	fromClones = dedupeDetectedClones(fromClones)

	// Calculate inconsistent changes by listing clones not modified by this patch
	missingClones := filterMissingChanges(fromClones, patchChunks)

	// Use file proximity ranking from FLeCCS
	// NOTE: Maybe we can utilize simple clone tracking to improve suggestion accuracy?
	patchPaths := ds.Map(patchChunks, func(c *chunk) string { return c.filename })
	slices.SortFunc(missingClones, ds.SortCompose(
		ds.SortAsc(func(c domain.Clone) int {
			distances := ds.Map(patchPaths, func(path string) int { return fleccs.FileTreeDistance(path, c.Filename) })
			return lo.Min(distances)
		}),
		ds.SortAsc(func(c domain.Clone) float64 { return c.Distance }),
	))

	// Return the inconsistent changes found
	return missingClones, nil
}

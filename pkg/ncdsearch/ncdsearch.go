// Package ncdsearch contains a basic golang re-implementation of the tool "NCDSearch"
// by Takashi Ishio, et al.
// https://github.com/takashi-ishio/NCDSearch
//
// Details may differ.
package ncdsearch

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"

	"github.com/salab/iccheck/pkg/utils/files"
	"github.com/salab/iccheck/pkg/utils/strs"
)

func overlap(n int, q, f []byte) float64 {
	nGramQ := strs.NGram(n, q)
	nGramF := strs.NGram(n, f)
	intersection := strs.IntersectionCount(nGramQ, nGramF)
	return float64(intersection) / float64(len(nGramQ))
}

type lzSet = strs.Set

func extractLZSet(b []byte) lzSet {
	set := make(lzSet)
	start, end := 0, 1
	for end <= len(b) {
		bs := string(b[start:end])
		if _, ok := set[bs]; !ok {
			set[bs] = struct{}{}
			start = end
		}
		end++
	}
	return set
}

func compareLZJD(b []byte, pos []int, queryLZSet lzSet) (kBest int, distance float64) {
	s := make(map[string]struct{})
	start, end := 0, 1
	intersection := 0
	distance = math.MaxFloat64

	for k := 0; k < len(pos); k++ {
		for end <= pos[k] {
			bs := string(b[start:end])
			if _, ok := s[bs]; !ok {
				s[bs] = struct{}{}
				start = end
				if _, ok := queryLZSet[bs]; ok {
					intersection++
				}
			}
			end++
		}

		lzjd := 1.0 - float64(intersection)/float64(len(queryLZSet)+len(s)-intersection)
		if lzjd < distance {
			kBest = k
			distance = lzjd
		}
	}

	return
}

func tokenIndices(s []byte, tokenize TokenizeFunc) []int {
	tokens := tokenize(s)
	indices := make([]int, len(tokens)+1)
	// make cumulative sum
	for i := 1; i < len(indices); i++ {
		indices[i] = indices[i-1] + len(tokens[i-1])
	}
	return indices
}

type Clone struct {
	Filename  string
	StartLine int
	EndLine   int
	KBest     int
	Distance  float64
}

func codeSearch(q []byte, windowSize int, threshold float64, file []byte, filename string, tokenize TokenizeFunc) (clones []Clone) {
	clones = make([]Clone, 0)
	lzSetQ := extractLZSet(q)

	tokenIndices := tokenIndices(file, tokenize)
	lineIndices := files.LineIndices(file)
	getLine := func(pos int) int {
		found := sort.SearchInts(lineIndices, pos)
		if found < len(lineIndices) && lineIndices[found] == pos {
			return found + 1
		}
		return found
	}

	end := len(tokenIndices) - windowSize
	for tokenIdx := 0; tokenIdx < end; tokenIdx++ {
		tokenStartIdx := tokenIndices[tokenIdx]
		tokenEndIdx := tokenIndices[tokenIdx+windowSize]

		b := file[tokenStartIdx:tokenEndIdx]
		pos := make([]int, windowSize)
		for i := 0; i < len(pos); i++ {
			pos[i] = tokenIndices[tokenIdx+i] - tokenStartIdx
		}

		kBest, distance := compareLZJD(b, pos, lzSetQ)
		if distance < threshold {
			clones = append(clones, Clone{
				Filename:  filename,
				StartLine: getLine(tokenStartIdx),
				EndLine:   getLine(tokenEndIdx), // TODO: get accurate end line
				KBest:     kBest,
				Distance:  distance,
			})
		}
	}

	return clones
}

func Search(query []byte, searchRoot string, options ...ConfigFunc) []Clone {
	c := applyConfig(options...)

	lzSetQ := extractLZSet(query)
	windowSize := int(math.Floor(c.windowSizeMult * float64(len(c.tokenize(query)))))
	fmt.Println("lzSetQ", lzSetQ)
	fmt.Println("windowSize", windowSize)

	files := make([]string, 0)
	lo.Must0(filepath.WalkDir(searchRoot, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	}))

	var clones []Clone
	cloneCh := make(chan []Clone)
	go func() {
		for detected := range cloneCh {
			clones = append(clones, detected...)
		}
	}()
	p := pool.New().WithMaxGoroutines(runtime.NumCPU())
	for _, filename := range files {
		p.Go(func() {
			file := lo.Must(os.ReadFile(filename))
			filterSim := overlap(c.overlapNGram, query, file)
			//fmt.Printf("%v overlap: %v\n", filename, filterSim)
			if filterSim < c.filterThreshold {
				return
			}
			cloneCh <- codeSearch(query, windowSize, c.searchThreshold, file, filename, c.tokenize)
		})
	}
	p.Wait()
	close(cloneCh)

	slices.SortFunc(clones, func(a, b Clone) int {
		if a.Distance < b.Distance {
			return 1
		}
		if b.Distance < a.Distance {
			return -1
		}
		return 0
	})
	return clones
}

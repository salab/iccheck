package strs

import (
	"github.com/salab/iccheck/pkg/utils/ds"
	"slices"
)

type Set = map[string]struct{}

func NGram(n int, s []byte) Set {
	end := len(s) - n + 1
	set := make(Set, end)
	for i := 0; i < end; i++ {
		ss := s[i : i+n]
		set[string(ss)] = struct{}{}
	}
	return set
}

func IntersectionCount(s1, s2 Set) int {
	cnt := 0
	for elt := range s1 {
		if _, ok := s2[elt]; ok {
			cnt++
		}
	}
	return cnt
}

type BigramSet = []uint16

func Bigram(s []byte) BigramSet {
	end := max(0, len(s)-1)
	set := make([]uint16, end)
	for i := 0; i < end; i++ {
		s0, s1 := uint16(s[i]), uint16(s[i+1])
		set[i] = s0 + (s1 << 8)
	}
	slices.Sort(set)
	return ds.Uniq(set)
}

func BigramIntersectionCount(s1, s2 BigramSet) int {
	cnt := 0
	for idx1, idx2 := 0, 0; idx1 < len(s1) && idx2 < len(s2); {
		e1, e2 := s1[idx1], s2[idx2]
		if e1 < e2 {
			idx1++
		} else if e2 < e1 {
			idx2++
		} else {
			cnt++
			idx1++
			idx2++
		}
	}
	return cnt
}

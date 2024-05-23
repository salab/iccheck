package strs

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

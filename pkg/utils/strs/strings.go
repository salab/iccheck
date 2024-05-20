package strs

type Set = map[string]struct{}

func NGram(n int, s []byte) Set {
	set := make(Set)
	end := len(s) - n + 1
	for i := 0; i < end; i++ {
		ss := s[i : i+n]
		set[string(ss)] = struct{}{}
	}
	return set
}

func Intersection(s1, s2 Set) Set {
	i := make(Set)
	for elt := range s1 {
		if _, ok := s2[elt]; ok {
			i[elt] = struct{}{}
		}
	}
	return i
}

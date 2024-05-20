package ds

import "golang.org/x/exp/constraints"

func Map[T any, R any](arr []T, iteratee func(T) R) []R {
	ret := make([]R, len(arr))
	for i, elt := range arr {
		ret[i] = iteratee(elt)
	}
	return ret
}

func SortCompose[E any](comparators ...func(e1, e2 E) int) func(e1, e2 E) int {
	return func(e1, e2 E) int {
		for _, cmp := range comparators {
			res := cmp(e1, e2)
			if res < 0 {
				return -1
			} else if res > 0 {
				return 1
			}
		}
		return 0
	}
}

func SortAsc[E any, K constraints.Ordered](key func(e E) K) func(e1, e2 E) int {
	return func(e1, e2 E) int {
		k1, k2 := key(e1), key(e2)
		if k1 < k2 {
			return -1
		} else if k1 == k2 {
			return 0
		} else {
			return 1
		}
	}
}

func SortDesc[E any, K constraints.Ordered](key func(e E) K) func(e1, e2 E) int {
	return func(e1, e2 E) int {
		k1, k2 := key(e1), key(e2)
		if k1 < k2 {
			return 1
		} else if k1 == k2 {
			return 0
		} else {
			return -1
		}
	}
}

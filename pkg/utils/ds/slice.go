package ds

import "golang.org/x/exp/constraints"

func Copy[T any, ST ~[]T](arr ST) ST {
	res := make(ST, len(arr))
	copy(res, arr)
	return res
}

// Limit limits array to given limit length, if the slice is larger than the given limit.
func Limit[T any, ST ~[]T](arr ST, limit int) ST {
	if len(arr) <= limit {
		return arr
	}
	return arr[:limit]
}

func Map[T any, R any, ST ~[]T](arr ST, iteratee func(T) R) []R {
	ret := make([]R, len(arr))
	for i, elt := range arr {
		ret[i] = iteratee(elt)
	}
	return ret
}

func MapError[T any, R any, ST ~[]T](arr ST, iteratee func(T) (R, error)) ([]R, error) {
	ret := make([]R, len(arr))
	var err error
	for i, elt := range arr {
		ret[i], err = iteratee(elt)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func FlatMap[T any, R any, ST ~[]T, SR ~[]R](arr ST, iteratee func(T) SR) SR {
	ret := make([]R, 0, len(arr))
	for _, elt := range arr {
		values := iteratee(elt)
		ret = append(ret, values...)
	}
	return ret
}

func FlatMapError[T any, R any, ST ~[]T, SR ~[]R](arr ST, iteratee func(T) (SR, error)) (SR, error) {
	ret := make([]R, 0, len(arr))
	for _, elt := range arr {
		values, err := iteratee(elt)
		if err != nil {
			return nil, err
		}
		ret = append(ret, values...)
	}
	return ret, nil
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

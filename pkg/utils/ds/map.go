package ds

import (
	"sync"
)

// SyncMap is a simple wrapper around sync.Map.
// https://www.reddit.com/r/golang/comments/twucb0/comment/j4x7xbx/
type SyncMap[K comparable, V any] struct {
	m sync.Map
}

func (m *SyncMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return value, ok
	}
	return v.(V), ok
}

func (m *SyncMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, loaded := m.m.LoadAndDelete(key)
	if !loaded {
		return value, loaded
	}
	return v.(V), loaded
}

func (m *SyncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	a, loaded := m.m.LoadOrStore(key, value)
	return a.(V), loaded
}

func (m *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

func (m *SyncMap[K, V]) Copy() map[K]V {
	res := make(map[K]V)
	m.Range(func(key K, value V) bool {
		res[key] = value
		return true
	})
	return res
}

func (m *SyncMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func MergeMap[K comparable, V any, M ~map[K]V](m1, m2 M) M {
	ret := make(M, len(m1)+len(m2))
	for k, v := range m1 {
		ret[k] = v
	}
	for k, v := range m2 {
		ret[k] = v
	}
	return ret
}

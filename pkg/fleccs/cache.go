package fleccs

import (
	"github.com/dgraph-io/ristretto"
	"github.com/samber/lo"
	"unsafe"
)

// cache kicks in only when searching for multiple times in the same process.
// Currently, LSP server takes advantage of this.
var cache = lo.Must(ristretto.NewCache(&ristretto.Config[uint64, []*Candidate]{
	NumCounters: 1e7,
	MaxCost:     100 * 1e6, // 100MB
	BufferItems: 64,
}))

var candidateStructBaseSize = int64(unsafe.Sizeof(Candidate{}))

func getFromCacheOrCalcCandidates(hash1, hash2 uint64, fn func() []*Candidate) []*Candidate {
	key := (hash1 << 32) | uint64(uint32(hash2))
	if v, ok := cache.Get(key); ok {
		return v
	}

	v := fn()
	cost := lo.SumBy(v, func(c *Candidate) int64 {
		return candidateStructBaseSize + int64(len(c.Filename)) + int64(len(c.Source.Filename))
	}) + 24 // nil (8 bytes?) or slice overhead
	cache.Set(key, v, cost)

	return v
}

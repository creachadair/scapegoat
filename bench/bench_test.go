package bench_test

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"

	"bitbucket.org/creachadair/scapegoat/bench"
)

const benchSeed = 1471808909908695897

// Trial values of β for load-testing tree operations.
var balances = []int{0, 50, 100, 150, 200, 225, 250, 275, 500, 800, 1000}

func randomTree(b *testing.B, β int) (*bench.Tree, []bench.KV) {
	rng := rand.New(rand.NewSource(benchSeed))
	values := make([]bench.KV, b.N)
	for i := range values {
		values[i].Key = rng.Intn(math.MaxInt32)
	}
	return bench.NewKeys(β, values...), values
}

func BenchmarkNewKeys(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			randomTree(b, β)
		})
	}
}

func BenchmarkInsertRandom(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			_, values := randomTree(b, β)
			b.ResetTimer()
			tree := bench.New(β)
			for _, v := range values {
				tree.Insert(v.Key, v.Value)
			}
		})
	}
}

func BenchmarkInsertOrdered(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree := bench.New(β)
			for i := 1; i <= b.N; i++ {
				tree.Insert(i, nil)
			}
		})
	}
}

func BenchmarkRemoveRandom(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree, values := randomTree(b, β)
			b.ResetTimer()
			for _, v := range values {
				tree.Remove(v.Key)
			}
		})
	}
}

func BenchmarkRemoveOrdered(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree, values := randomTree(b, β)
			sort.Sort(kvSlice(values))
			b.ResetTimer()
			for _, v := range values {
				tree.Remove(v.Key)
			}
		})
	}
}

func BenchmarkLookup(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree, values := randomTree(b, β)
			b.ResetTimer()
			for _, v := range values {
				tree.Lookup(v.Key)
			}
		})
	}
}

type kvSlice []bench.KV

func (s kvSlice) Len() int           { return len(s) }
func (s kvSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s kvSlice) Less(i, j int) bool { return s[i].Key < s[j].Key }

package scapegoat

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
)

const benchSeed = 1471808909908695897

// Trial values of β for load-testing tree operations.
var balances = []int{0, 50, 100, 150, 200, 225, 250, 275, 500, 800, 1000}

func randomTree(b *testing.B, β int) (*Tree, []Key) {
	rng := rand.New(rand.NewSource(benchSeed))
	values := make([]Key, b.N)
	for i := range values {
		values[i] = Z(rng.Intn(math.MaxInt32))
	}
	return NewKeys(β, values...), values
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
			tree := New(β)
			for _, v := range values {
				tree.Insert(v)
			}
		})
	}
}

func BenchmarkInsertOrdered(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree := New(β)
			for i := 1; i <= b.N; i++ {
				tree.Insert(Z(i))
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
				tree.Remove(v)
			}
		})
	}
}

func BenchmarkRemoveOrdered(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree, values := randomTree(b, β)
			sort.Sort(keySlice(values))
			b.ResetTimer()
			for _, v := range values {
				tree.Remove(v)
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
				tree.Lookup(v)
			}
		})
	}
}

type keySlice []Key

func (s keySlice) Len() int           { return len(s) }
func (s keySlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s keySlice) Less(i, j int) bool { return s[i].Less(s[j]) }

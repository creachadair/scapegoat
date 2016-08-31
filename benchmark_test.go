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
var balances = []int{0, 50, 100, 150, 200, 250, 500, 800, 1000}

func benchTree(β int) (*Tree, *rand.Rand) {
	return New(β), rand.New(rand.NewSource(benchSeed))
}

func randomTree(b *testing.B, β int) (*Tree, []int) {
	tree, rng := benchTree(β)
	values := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		values[i] = rng.Intn(math.MaxInt32)
		tree.Insert(Z(values[i]))
	}
	return tree, values
}

func BenchmarkInsertRandom(b *testing.B) {
	b.Run("β", func(b *testing.B) {
		for _, β := range balances {
			b.Run(fmt.Sprint(β), func(b *testing.B) {
				randomTree(b, β)
			})
		}
	})
}

func BenchmarkInsertOrdered(b *testing.B) {
	b.Run("β", func(b *testing.B) {
		for _, β := range balances {
			b.Run(fmt.Sprint(β), func(b *testing.B) {
				tree := New(β)
				for i := 1; i <= b.N; i++ {
					tree.Insert(Z(i))
				}
			})
		}
	})
}

func BenchmarkRemoveRandom(b *testing.B) {
	b.Run("β", func(b *testing.B) {
		for _, β := range balances {
			b.Run(fmt.Sprint(β), func(b *testing.B) {
				tree, values := randomTree(b, β)
				b.ResetTimer()
				for _, v := range values {
					tree.Remove(Z(v))
				}
			})
		}
	})
}

func BenchmarkRemoveOrdered(b *testing.B) {
	b.Run("β", func(b *testing.B) {
		for _, β := range balances {
			b.Run(fmt.Sprint(β), func(b *testing.B) {
				tree, values := randomTree(b, β)
				sort.Ints(values)
				b.ResetTimer()
				for _, v := range values {
					tree.Remove(Z(v))
				}
			})
		}
	})
}

func BenchmarkLookup(b *testing.B) {
	b.Run("β", func(b *testing.B) {
		for _, β := range balances {
			b.Run(fmt.Sprint(β), func(b *testing.B) {
				tree, values := randomTree(b, β)
				sort.Ints(values)
				b.ResetTimer()
				for _, v := range values {
					tree.Lookup(Z(v))
				}
			})
		}
	})
}

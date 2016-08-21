package scapegoat

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

const benchSeed = 1471808909908695897

// Trial values of β for load-testing tree operations.
var balances = []int{0, 50, 100, 150, 200, 250, 500, 800, 1000}

func benchTree(β int) (*Tree, *rand.Rand) {
	return New(β), rand.New(rand.NewSource(benchSeed))
}

func benchInsert(b *testing.B, β int) *Tree {
	tree, rng := benchTree(β)
	for i := 0; i < b.N; i++ {
		tree.Insert(Z(rng.Intn(math.MaxInt32)))
	}
	return tree
}

func BenchmarkInsert(b *testing.B) {
	b.Run("balance", func(b *testing.B) {
		for _, β := range balances {
			b.Run(fmt.Sprint(β), func(b *testing.B) { benchInsert(b, β) })
		}
	})
}

func fullTree(b *testing.B, β int) (*Tree, []int) {
	tree, rng := benchTree(β)
	values := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		values[i] = rng.Intn(math.MaxInt32)
		tree.Insert(Z(values[i]))
	}
	return tree, values
}

func benchRemove(b *testing.B, β int) {
	tree, values := fullTree(b, β)
	b.ResetTimer()
	for _, v := range values {
		tree.Remove(Z(v))
	}
}

func BenchmarkRemove(b *testing.B) {
	b.Run("balance", func(b *testing.B) {
		for _, β := range balances {
			b.Run(fmt.Sprint(β), func(b *testing.B) { benchRemove(b, β) })
		}
	})
}

func benchLookup(b *testing.B, β int) {
	tree, values := fullTree(b, β)
	b.ResetTimer()
	for _, v := range values {
		tree.Lookup(Z(v))
	}
}

func BenchmarkLookup(b *testing.B) {
	b.Run("balance", func(b *testing.B) {
		for _, β := range balances {
			b.Run(fmt.Sprint(β), func(b *testing.B) { benchLookup(b, β) })
		}
	})
}

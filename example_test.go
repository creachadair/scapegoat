package scapegoat

import "fmt"

type W string

func (w W) Less(key Key) bool { return w < key.(W) }

type Z int

func (z Z) Less(key Key) bool { return z < key.(Z) }

func ExampleTree_Insert() {
	// type W string
	// func (w W) Less(key Key) bool { return w < key.(W) }

	tree := New(200)
	tree.Insert(W("never"))
	tree.Insert(W("say"))
	tree.Insert(W("never"))
	fmt.Println("tree.Len() =", tree.Len())
	// Output:
	// tree.Len() = 2
}

func ExampleTree_Remove() {
	key := W("Aloysius")
	tree := New(1)
	fmt.Println("inserted:", tree.Insert(key))
	fmt.Println("removed:", tree.Remove(key))
	fmt.Println("re-removed:", tree.Remove(key))
	// Output:
	// inserted: true
	// removed: true
	// re-removed: false
}

func ExampleTree_Lookup() {
	tree := NewKeys(1, W("mom"))
	fmt.Println("hit:", tree.Lookup(W("mom")))
	fmt.Println("miss:", tree.Lookup(W("dad")))
	// Output:
	// hit: mom
	// miss: <nil>
}

func ExampleTree_Inorder() {
	tree := NewKeys(15, W("freaking"), W("eat"), W("those"), W("vegetables"))
	tree.Inorder(func(key Key) bool {
		fmt.Println(key)
		return true
	})
	// Output:
	// eat
	// freaking
	// those
	// vegetables
}

func ExampleTree_Min() {
	tree := NewKeys(50, Z(1814), Z(1956), Z(955), Z(1066), Z(2016))
	fmt.Println("len:", tree.Len())
	fmt.Println("min:", tree.Min())
	fmt.Println("max:", tree.Max())
	// Output:
	// len: 5
	// min: 955
	// max: 2016
}

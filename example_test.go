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
	tree.Insert(W("never"), nil)
	tree.Insert(W("say"), nil)
	tree.Insert(W("never"), nil)
	fmt.Println("tree.Len() =", tree.Len())
	// Output:
	// tree.Len() = 2
}

func ExampleTree_Remove() {
	key := W("Aloysius")
	tree := New(1)
	fmt.Println("inserted:", tree.Insert(key, nil))
	fmt.Println("removed:", tree.Remove(key))
	fmt.Println("re-removed:", tree.Remove(key))
	// Output:
	// inserted: true
	// removed: true
	// re-removed: false
}

func ExampleTree_Lookup() {
	tree := NewKeys(1, KV{Key: W("mom")})
	hit, ok := tree.Lookup(W("mom"))
	fmt.Printf("%v, %v\n", hit, ok)
	miss, ok := tree.Lookup(W("dad"))
	fmt.Printf("%v, %v\n", miss, ok)
	// Output:
	// mom, true
	// <nil>, false
}

func ExampleTree_Inorder() {
	tree := NewKeys(15,
		KV{Key: W("eat")},
		KV{Key: W("those")},
		KV{Key: W("bloody")},
		KV{Key: W("vegetables")},
	)
	tree.Inorder(func(kv KV) bool {
		fmt.Println(kv.Key)
		return true
	})
	// Output:
	// bloody
	// eat
	// those
	// vegetables
}

func ExampleTree_Min() {
	tree := NewKeys(50,
		KV{Key: Z(1814)},
		KV{Key: Z(1956)},
		KV{Key: Z(955)},
		KV{Key: Z(1066)},
		KV{Key: Z(2016)},
	)
	fmt.Println("len:", tree.Len())
	fmt.Println("min:", tree.Min().Key)
	fmt.Println("max:", tree.Max().Key)
	// Output:
	// len: 5
	// min: 955
	// max: 2016
}

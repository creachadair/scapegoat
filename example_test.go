package scapegoat

import "fmt"

func ExampleTree_Insert() {
	tree := New(200)
	tree.Insert("never", nil)
	tree.Insert("say", nil)
	tree.Insert("never", nil)
	fmt.Println("tree.Len() =", tree.Len())
	// Output:
	// tree.Len() = 2
}

func ExampleTree_Remove() {
	const key = "Aloysius"
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
	tree := NewKeys(1, KV{Key: "mom"})
	hit, ok := tree.Lookup("mom")
	fmt.Printf("%v, %v\n", hit, ok)
	miss, ok := tree.Lookup("dad")
	fmt.Printf("%v, %v\n", miss, ok)
	// Output:
	// mom, true
	// <nil>, false
}

func ExampleTree_Inorder() {
	tree := NewKeys(15,
		KV{Key: "eat"},
		KV{Key: "those"},
		KV{Key: "bloody"},
		KV{Key: "vegetables"},
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
		KV{Key: "1814"},
		KV{Key: "1956"},
		KV{Key: "0955"},
		KV{Key: "1066"},
		KV{Key: "2016"},
	)
	fmt.Println("len:", tree.Len())
	fmt.Println("min:", tree.Min().Key)
	fmt.Println("max:", tree.Max().Key)
	// Output:
	// len: 5
	// min: 0955
	// max: 2016
}

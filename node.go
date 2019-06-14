package scapegoat

import "fmt"

type node struct {
	key         Key
	value       Value
	left, right *node
}

// size reports the number of nodes contained in the tree rooted at n.
// If n == nil, this is defined as 0.
func (n *node) size() int {
	if n == nil {
		return 0
	}
	return 1 + n.left.size() + n.right.size()
}

// flatten extracts the nodes rooted at n into a slice in order, and returns
// the resulting slice. The results are appended to into, thus allowing the
// caller to preallocate storage:
//
// Example:
//   into := n.flatten(make([]*node, 0, n.size()))
//
// If cap(into) ≥ n.size(), this method does not allocate on the heap.
func (n *node) flatten(into []*node) []*node {
	if n != nil {
		into = n.left.flatten(into)
		into = append(into, n)
		into = n.right.flatten(into)
	}
	return into
}

// extract constructs a balanced tree from the given nodes and returns the root
// of the tree. The child pointers of the resulting nodes are updated in place.
// This function does not allocate on the heap.
func extract(nodes []*node) *node {
	if len(nodes) == 0 {
		return nil
	}
	mid := (len(nodes) - 1) / 2
	root := nodes[mid]
	root.left = extract(nodes[:mid])
	root.right = extract(nodes[mid+1:])
	return root
}

// rewrite composes flatten and extract, returning the rewritten root.
// Costs a single size-element array allocation, plus O(lg size) stack space,
// but does no other allocation.
func rewrite(root *node, size int) *node {
	nodes := root.flatten(make([]*node, 0, size))
	if len(nodes) != size {
		panic(fmt.Sprintf("len(nodes) = %d but size = %d", len(nodes), size))
	}
	return extract(nodes)
}

// popMinRight removes the smallest node from the right subtree of root,
// modifying the tree in-place and returning the node removed.
// This function panics if root == nil or root.right == nil.
func popMinRight(root *node) *node {
	par, goat := root, root.right
	for goat.left != nil {
		par, goat = goat, goat.left
	}
	if par == root {
		root.right = goat.right
	} else {
		par.left = goat.right
	}
	goat.left = nil
	goat.right = nil
	return goat
}

// inorder visits the subtree under n inorder, calling f until f returns false.
func (n *node) inorder(f func(KV) bool) bool {
	if n == nil {
		return true
	} else if ok := n.left.inorder(f); !ok {
		return false
	} else if ok := f(KV{Key: n.key, Value: n.value}); !ok {
		return false
	}
	return n.right.inorder(f)
}

// pathTo returns the sequence of nodes beginning at n leading to key, if key
// is present. If key was found, its node is the last element of the path.
func (n *node) pathTo(key Key) []*node {
	var path []*node
	cur := n
	for cur != nil {
		path = append(path, cur)
		if keyLess(key, cur.key) {
			cur = cur.left
		} else if keyLess(cur.key, key) {
			cur = cur.right
		} else {
			break
		}
	}
	return path
}

// inorderAfter visits the elements of the subtree under n not less than key
// inorder, calling f for each until f returns false.
func (n *node) inorderAfter(key Key, f func(KV) bool) {
	// Find the path from the root to key. Any nodes greater than or equal to
	// key must be on or to the right of this path.
	path := n.pathTo(key)
	for i := len(path) - 1; i >= 0; i-- {
		cur := path[i]
		if keyLess(cur.key, key) {
			continue
		} else if ok := f(KV{Key: cur.key, Value: cur.value}); !ok {
			return
		} else if ok := cur.right.inorder(f); !ok {
			return
		}
	}
}

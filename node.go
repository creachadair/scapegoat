package scapegoat

import "fmt"

type node struct {
	key         Key
	left, right *node
}

// size reports the number of nodes contained in the tree rooted at n.
// If n == 0, this is defined as 0.
func (n *node) size() int {
	if n == nil {
		return 0
	}
	return 1 + n.left.size() + n.right.size()
}

// findLeast returns the node containing the smallest key not less than key.
// The flag reports whether it is equal.
func (n *node) findLeast(key Key) (*node, bool) {
	if n == nil {
		return nil, false
	}

	var next *node
	before := key.Less(n.key)
	if before {
		next = n.left
	} else if n.key.Less(key) {
		next = n.right
	} else {
		return n, true // exact match
	}
	if match, ok := next.findLeast(key); match != nil {
		return match, ok
	}

	// If we reach here, the subtree where key would exist is not present.  The
	// key we want, if it exists, is the first one along the path up to the
	// root that is after key.

	if before {
		return n, false // it's me!
	}
	return nil, false // keep looking
}

// flatten extracts the nodes rooted at n into a slice in order, and returns
// the resulting slice. The he results are appended to it, thus allowing the
// caller to preallocate storage:
//
// Example:
//   into := make([]*node, 0, n.size())
//   n.flatten(into)
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

// inorder visits the subtree under root inorder, calling f until f returns false.
func (n *node) inorder(f func(Key) bool) bool {
	if n == nil {
		return true
	} else if ok := n.left.inorder(f); !ok {
		return false
	} else if ok := f(n.key); !ok {
		return false
	}
	return n.right.inorder(f)
}

// Package scapegoat implements a Scapegoat Tree, as described in the paper
//
//  I. Galperin, R. Rivest: "Scapegoat Trees"
//  https://people.csail.mit.edu/rivest/pubs/GR93.pdf
//
// A scapegoat tree is an approximately-balanced binary search tree structure
// with worst-case O(lg n) lookup and amortized O(lg n) insert and delete.  The
// worst-case cost of a single insert or delete is O(n).
//
// It is also relatively memory-efficient, as interior nodes do not require any
// ancillary metadata for balancing purposes, and the tree itself costs only a
// few words of bookkeeping overhead beyond the nodes. A rebalancing operation
// requires only a single contiguous vector allocation.
//
package scapegoat

import (
	"math"
	"sort"
)

// A KV combines a key with a value. Values are not interpreted, and may be nil
// if the key records all the information of interest.
type KV struct {
	Key   Key
	Value Value
}

func (kv KV) node() *node { return &node{key: kv.Key, value: kv.Value} }

const (
	maxBalance = 1000
	fracLimit  = 2 * maxBalance
)

// New returns *Tree with the given balancing factor 0 ≤ β ≤ 1000 and keys.
// The balancing factor represents how unbalanced the tree is permitted to be,
// with 0 being strictest (as near as possible to 50% weight balance) and 1000
// being loosest (no rebalancing).
//
// New panics if β < 0 or β > 1000.
func New(β int, kvs ...KV) *Tree {
	if β < 0 || β > maxBalance {
		panic("β out of range")
	}
	tree := &Tree{
		β:     β,
		limit: limitFunc(β),
		size:  len(kvs),
		max:   len(kvs),
	}
	if len(kvs) != 0 {
		nodes := make([]*node, len(kvs))
		for i, kv := range kvs {
			nodes[i] = kv.node()
		}
		sort.Slice(nodes, func(i, j int) bool {
			return keyLess(nodes[i].key, nodes[j].key)
		})
		tree.root = extract(nodes)
	}
	return tree
}

// A Tree is the root of a scapegoat tree. A *Tree is not safe for concurrent
// use without external synchronization.
type Tree struct {
	root *node

	// β identifies a point on the interval [maxBalance,fracLimit], and we
	// compute the balance fraction as β/fracLimit. This permits breakpoint
	// computations to use only fixed-point integer arithmetic and only
	// requires one floating-point operation per insertion to recompute the
	// depth limit.

	β     int             // balancing factor
	limit func(n int) int // depth limit for size n
	size  int             // cache of root.size()
	max   int             // max of size since last rebuild of root
}

func toFraction(β int) float64 { return (float64(β) + maxBalance) / fracLimit }

// limitFunc returns a function that computes the depth limit for a tree of
// size n given the balance factor β.
func limitFunc(β int) func(int) int {
	inv := 1 / toFraction(β)
	if inv == 1 { // int(+Inf) ⇒ undefined
		return func(n int) int { return n + 1 }
	}
	base := math.Log(inv)
	return func(n int) int { return int(math.Log(float64(n)) / base) }
}

// Insert adds key into the tree if it is not already present, and reports
// whether a new node was added.
func (t *Tree) Insert(key Key, value Value) bool {
	// We don't yet know whether the insertion will add mass to the tree; we
	// conservatively assume it might for purposes of choosing a depth limit.
	ins, ok, _, _ := t.insert(&KV{Key: key, Value: value}, false, t.root, t.limit(t.size+1))
	t.incSize(ok)
	t.root = ins
	return ok
}

// Replace adds key to the tree, updating an existing key if it is already
// present. Reports whether a new node was added.
func (t *Tree) Replace(key Key, value Value) bool {
	ins, ok, _, _ := t.insert(&KV{Key: key, Value: value}, true, t.root, t.limit(t.size+1))
	t.incSize(ok)
	t.root = ins
	return ok
}

// incSize increments t.size and updates t.max if inserted is true.
func (t *Tree) incSize(inserted bool) {
	if inserted {
		t.size++
		if t.size > t.max {
			t.max = t.size
		}
	}
}

// insert key in order under root, with the given depth limit.
//
// If replace is true and an existing node has an equivalent key, it is updated
// with the given key; otherwise, inserting an existing key is a no-op.
//
// Returns the modified tree, and reports whether a new node was added and the
// height of the returned node above the point of insertion.
// If the insertion did not exceed the depth limit, size == 0.
// Otherwise, size == ins.size() meaning a scapegoat is needed.
func (t *Tree) insert(kv *KV, replace bool, root *node, limit int) (ins *node, added bool, size, height int) {
	// Descending phase: Insert the key into the tree structure.
	var sib *node
	if root == nil {
		if limit < 0 {
			size = 1
		}
		return kv.node(), true, size, 0
	} else if keyLess(kv.Key, root.key) {
		ins, added, size, height = t.insert(kv, replace, root.left, limit-1)
		root.left = ins
		sib = root.right
		height++
	} else if keyLess(root.key, kv.Key) {
		ins, added, size, height = t.insert(kv, replace, root.right, limit-1)
		root.right = ins
		sib = root.left
		height++
	} else {
		// Replacing an existing node. This cannot introduce a violation, so we
		// can return immediately without triggering a goat search.
		if replace {
			root.value = kv.Value
		}
		return root, false, 0, 0
	}

	// Ascending phase, a.k.a., goat rodeo.
	// Uses the selection strategy from section 4.6 of Galperin & Rivest .

	// If size != 0, we exceeded the depth limit and are looking for a goat.
	// Note: size == ins.size() not root.size() at this point.
	if size > 0 {
		sibSize := sib.size()          // size of sibling subtree
		rootSize := sibSize + 1 + size // new size of root

		if bw := t.limit(rootSize); height <= bw {
			size = rootSize // not the goat you're looking for; move along
		} else {
			// root is the goat; rewrite it and signal the activations above us
			// to stop looking by setting size to 0.
			root = rewrite(root, rootSize)
			size = 0
		}
	}
	return root, added, size, height
}

// Remove key from the tree and report whether it was present.
func (t *Tree) Remove(key Key) bool {
	del, ok := t.root.remove(key)
	t.root = del
	if ok {
		t.size--
		if bw := (t.max*t.β + maxBalance) / fracLimit; t.size < bw {
			t.root = rewrite(t.root, t.size)
			t.max = t.size
		}
	}
	return ok
}

// remove key from the subtree under n, returning the modified tree reporting
// whether the mass of the tree was decreased.
func (n *node) remove(key Key) (_ *node, ok bool) {
	if n == nil {
		return nil, false // nothing to do
	} else if keyLess(key, n.key) {
		n.left, ok = n.left.remove(key)
		return n, ok
	} else if keyLess(n.key, key) {
		n.right, ok = n.right.remove(key)
		return n, ok
	} else if n.left == nil {
		return n.right, true
	} else if n.right == nil {
		return n.left, true
	}

	// At this point we need to remove n, but it has two children.
	// Do the usual trick.
	goat := popMinRight(n)
	n.key = goat.key
	return n, true
}

// Len reports the number of elements stored in the tree.
func (t *Tree) Len() int { return t.size }

// Lookup reports whether key is present in the tree, and returns the value
// associated with that key, or nil if the key is not present.
func (t *Tree) Lookup(key Key) (v Value, ok bool) {
	cur := t.root
	for cur != nil {
		if keyLess(key, cur.key) {
			cur = cur.left
		} else if keyLess(cur.key, key) {
			cur = cur.right
		} else {
			v, ok = cur.value, true
			return
		}
	}
	return
}

// Inorder traverses t inorder and invokes f for each key until either f
// returns false or no further keys are available.
func (t *Tree) Inorder(f func(KV) bool) { t.root.inorder(f) }

// InorderAfter traverses t inorder, considering only keys equal to or after
// key, and invokes f for each key until either f returns false or no further
// keys are available.
func (t *Tree) InorderAfter(key Key, f func(KV) bool) { t.root.inorderAfter(key, f) }

// Min returns the key/value pair in the tree with the minimum key, or nil if
// the tree is empty.
func (t *Tree) Min() *KV {
	if t.root == nil {
		return nil
	}
	cur := t.root
	for cur.left != nil {
		cur = cur.left
	}
	return &KV{Key: cur.key, Value: cur.value}
}

// Max returns the key/value pair in the tree with the maximum key, or nil if
// the tree is empty.
func (t *Tree) Max() *KV {
	if t.root == nil {
		return nil
	}
	cur := t.root
	for cur.right != nil {
		cur = cur.right
	}
	return &KV{Key: cur.key, Value: cur.value}
}

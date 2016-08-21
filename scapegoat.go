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
// few words of bookkeeping overhead beyond the nodes.
//
package scapegoat

import "math"

// A Key represents an entry in the tree, defined by having a ordering
// relationship with other keys. An implementation of this interface may carry
// additional data that does not affect comparison.
type Key interface {
	// Less reports whether the receiver is ordered prior to the argument.
	// If a.Less(b) == b.Less(a) == false, a and b will be assumed equal.
	//
	// If the receiver and argument are not comparable, the implementation
	// should panic.
	Less(Key) bool
}

const (
	maxBalance = 1000
	fracLimit  = 2 * maxBalance
)

// New returns an empty *Tree with the given balancing factor 0 ≤ β ≤ 1000.
// The balancing factor represents how unbalanced the tree is permitted to be,
// with 0 being strictest (50% balance) and 1000 being loosest (0% balance).
//
// New panics if β < 0 or β > 1000.
func New(β int) *Tree {
	if β < 0 || β > maxBalance {
		panic("β out of range")
	}
	return &Tree{β: β, limit: limitFunc(β)}
}

// NewKeys constructs a *Tree with the given balancing factor and keys.
// See New for a description of β.
func NewKeys(β int, keys ...Key) *Tree {
	tree := New(β)
	for _, key := range keys {
		tree.Insert(key)
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
	base := math.Log(inv)
	return func(n int) int { return int(math.Log(float64(n)) / base) }
}

// breakpoint computes the balance value breakpoint of a tree of n nodes.
func (t *Tree) breakpoint(n int) int {
	if bw := (n*t.β + maxBalance) / fracLimit; bw > 0 {
		return bw
	}
	return 1
}

// Insert adds key into the tree if it is not already present, and reports
// whether a new node was added.
func (t *Tree) Insert(key Key) bool {
	// We don't yet know whether the insertion will add mass to the tree; we
	// conservatively assume it might for purposes of choosing a depth limit.
	ins, ok, _, _ := t.insert(key, false, t.root, t.limit(t.size+1))
	t.incSize(ok)
	t.root = ins
	return ok
}

// Replace adds key to the tree, updating an existing key if it is already
// present. Reports whether a new node was added.
func (t *Tree) Replace(key Key) bool {
	ins, ok, _, _ := t.insert(key, true, t.root, t.limit(t.size+1))
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
func (t *Tree) insert(key Key, replace bool, root *node, limit int) (ins *node, added bool, size, height int) {
	// Descending phase: Insert the key into the tree structure.
	var sib *node
	if root == nil {
		if limit < 0 {
			size = 1
		}
		return &node{key: key}, true, size, 0
	} else if key.Less(root.key) {
		ins, added, size, height = t.insert(key, replace, root.left, limit-1)
		root.left = ins
		sib = root.right
		height++
	} else if root.key.Less(key) {
		ins, added, size, height = t.insert(key, replace, root.right, limit-1)
		root.right = ins
		sib = root.left
		height++
	} else {
		// Replacing an existing node. This cannot introduce a violation, so we
		// can return immediately without triggering a goat search.
		if replace {
			root.key = key
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
		if bw := t.breakpoint(t.max); t.size < bw {
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
	} else if key.Less(n.key) {
		n.left, ok = n.left.remove(key)
		return n, ok
	} else if n.key.Less(key) {
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

// Lookup returns the matching key from the tree, or nil if absent.
func (t *Tree) Lookup(key Key) Key {
	n, ok := t.root.findLeast(key)
	if n != nil && ok {
		return n.key
	}
	return nil
}

// Inorder traverses t inorder and invokes f for each key until either f
// returns false or no further keys are available.
func (t *Tree) Inorder(f func(Key) bool) { inorder(t.root, f) }

// Min returns the smallest key in the tree, or nil if the tree is empty.
func (t *Tree) Min() Key {
	if t.root == nil {
		return nil
	}
	cur := t.root
	for cur.left != nil {
		cur = cur.left
	}
	return cur.key
}

// Max returns the maximum key in the tree, or nil if the tree is empty.
func (t *Tree) Max() Key {
	if t.root == nil {
		return nil
	}
	cur := t.root
	for cur.right != nil {
		cur = cur.right
	}
	return cur.key
}

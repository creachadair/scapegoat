package scapegoat

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"bitbucket.org/creachadair/stringset"

	"github.com/kylelemons/godebug/pretty"
)

var (
	strictness = flag.Int("balance", 100, "Balancing factor")
	dotFile    = flag.String("dot", "", "Emit DOT output to this file")
)

func (n *node) height() int {
	if n == nil {
		return 0
	}
	h := n.left.height()
	if r := n.right.height(); r > h {
		h = r
	}
	return h + 1
}

// Construct a tree with the words from input, returning the finished tree and
// the original words as split by strings.Fields.
func makeTree(β int, input string) (*Tree, []string) {
	tree := New(β)
	words := strings.Fields(input)
	for _, w := range words {
		tree.Insert(W(w))
	}
	return tree, words
}

// Export all the words in tree in their stored order.
func allWords(tree *Tree) []string {
	var got []string
	tree.Inorder(func(key Key) bool {
		got = append(got, string(key.(W)))
		return true
	})
	return got
}

// If an output file is specified, dump a DOT graph of tree.
func dumpTree(tree *Tree) {
	if *dotFile == "" {
		return
	}
	f, err := os.Create(*dotFile)
	if err != nil {
		log.Fatalf("Unable to create DOT output: %v", err)
	}
	dotTree(f, tree.root)
	if err := f.Close(); err != nil {
		log.Fatalf("Unable to close output: %v", err)
	}
}

// Render tree to a GraphViz graph.
func dotTree(w io.Writer, root *node) {
	fmt.Fprintln(w, "digraph Tree {")

	i := 0
	next := func() int {
		i++
		return i
	}

	var ptree func(*node) int
	ptree = func(root *node) int {
		if root == nil {
			return 0
		}
		id := next()
		fmt.Fprintf(w, "\tN%04d [label=\"%s\"]\n", id, root.key)
		if lc := ptree(root.left); lc != 0 {
			fmt.Fprintf(w, "\tN%04d -> N%04d\n", id, lc)
		}
		if rc := ptree(root.right); rc != 0 {
			fmt.Fprintf(w, "\tN%04d -> N%04d\n", id, rc)
		}
		return id
	}
	ptree(root)
	fmt.Fprintln(w, "}")
}

func TestNewKeys(t *testing.T) {
	tree := NewKeys(200, W("please"), W("fetch"), W("your"), W("slippers"))
	got := allWords(tree)
	want := []string{"fetch", "please", "slippers", "your"}
	if diff := pretty.Compare(got, want); diff != "" {
		t.Errorf("NewTree produced unexpected output (-got, +want)\n%s", diff)
	}
}

func TestBasicProperties(t *testing.T) {
	// http://www.gutenberg.org/files/1063/1063-h/1063-h.htm
	tree, words := makeTree(*strictness, `
The thousand injuries of Fortunato I had borne as I best could but when he
ventured upon insult I vowed revenge You who so well know the nature of my soul
will not suppose however that gave utterance to a threat At length I would be
avenged this was a point definitely settled but the very definitiveness with
which it was resolved precluded the idea of risk I must not only punish but
punish with impunity A wrong is unredressed when retribution overtakes its
redresser It is equally unredressed when the avenger fails to make himself felt
as such to him who has done the wrong

It must be understood that neither by word nor deed had I given Fortunato cause
to doubt my good will I continued as was my in to smile in his face and he did
not perceive that my to smile now was at the thought of his immolation.

He had a weak point this Fortunato although in other regards he was a man to be
respected and even feared He prided himself on his connoisseurship in wine Few
Italians have the true virtuoso spirit For the most part their enthusiasm is
adopted to suit the time and opportunity to practise imposture upon the British
and Austrian millionaires In painting and gemmary Fortunato like his countrymen
was a quack but in the matter of old wines he was sincere In this respect I did
not differ from him materially I was skilful in the Italian vintages myself and
bought largely whenever I could`)

	t.Logf("Final tree has size %d; height %d", tree.Len(), tree.root.height())
	dumpTree(tree)

	got := allWords(tree)
	want := stringset.New(words...).Elements()
	if diff := pretty.Compare(got, want); diff != "" {
		t.Errorf("Inorder produced unexpected output (-got, +want)\n%s", diff)
	}
}

func TestRemoval(t *testing.T) {
	tree, words := makeTree(0, `a foolish consistency is the hobgoblin of little minds`)

	got := allWords(tree)
	if diff := pretty.Compare(got, stringset.New(words...).Elements()); diff != "" {
		t.Errorf("Original input differs from expected (-got, +want)\n%s", diff)
	}

	drop := stringset.New("a", "is", "of", "the")
	for w := range drop {
		if !tree.Remove(W(w)) {
			t.Errorf("Remove(%q) returned false, wanted true", w)
		}
	}

	got = allWords(tree)
	want := stringset.New(words...).Diff(drop).Elements()
	if diff := pretty.Compare(got, want); diff != "" {
		t.Errorf("Tree after removal is incorrect (-got, +want)\n%s", diff)
	}
}

func TestInorderAfter(t *testing.T) {
	keys := []Key{Z(8), Z(6), Z(7), Z(5), Z(3), Z(0), Z(9)}
	tree := NewKeys(0, keys...)
	tests := []struct {
		key  Z
		want []int
	}{
		{10, nil},
		{9, []int{9}},
		{8, []int{8, 9}},
		{7, []int{7, 8, 9}},
		{6, []int{6, 7, 8, 9}},
		{5, []int{5, 6, 7, 8, 9}},
		{4, []int{5, 6, 7, 8, 9}},
		{3, []int{3, 5, 6, 7, 8, 9}},
		{2, []int{3, 5, 6, 7, 8, 9}},
		{1, []int{3, 5, 6, 7, 8, 9}},
		{0, []int{0, 3, 5, 6, 7, 8, 9}},
		{-1, []int{0, 3, 5, 6, 7, 8, 9}},
	}
	for _, test := range tests {
		var got []int
		tree.InorderAfter(test.key, func(key Key) bool {
			got = append(got, int(key.(Z)))
			return true
		})
		if diff := pretty.Compare(got, test.want); diff != "" {
			t.Errorf("InorderAfter(%v) result differed from expected\n%s", test.key, diff)
		}
	}
}

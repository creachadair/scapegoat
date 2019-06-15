package scapegoat

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"testing"

	"bitbucket.org/creachadair/stringset"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	strictness = flag.Int("balance", 100, "Balancing factor")
	dotFile    = flag.String("dot", "", "Emit DOT output to this file")
	sortWords  = flag.Bool("sort", false, "Sort input words before insertion")
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
	if *sortWords {
		sort.Strings(words)
	}
	for i, w := range words {
		tree.Insert(w, i+1)
	}
	return tree, words
}

// Export all the words in tree in their stored order.
func allWords(tree *Tree) []string {
	var got []string
	tree.Inorder(func(kv KV) bool {
		got = append(got, kv.Key)
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
	tree := NewKeys(200,
		KV{Key: "please"},
		KV{Key: "fetch"},
		KV{Key: "your"},
		KV{Key: "slippers"},
	)
	got := allWords(tree)
	want := []string{"fetch", "please", "slippers", "your"}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("NewTree produced unexpected output (-got, +want)\n%s", diff)
	}
}

func TestBasicProperties(t *testing.T) {
	// http://www.gutenberg.org/files/1063/1063-h/1063-h.htm
	text, err := ioutil.ReadFile("cask.txt")
	if err != nil {
		t.Fatalf("Reading text: %v", err)
	}
	tree, words := makeTree(*strictness, string(text))
	t.Logf("Final tree has size %d; height %d", tree.Len(), tree.root.height())
	dumpTree(tree)

	got := allWords(tree)
	want := stringset.New(words...).Elements()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Inorder produced unexpected output (-want, +got)\n%s", diff)
	}

	// Verify that the values are of the correct type.
	for _, word := range want {
		v, ok := tree.Lookup(word)
		if !ok {
			t.Errorf("Word %q not found", word)
		} else if _, ok := v.(int); !ok {
			t.Errorf("Word %q value is %T, want int", word, v)
		}
	}
}

func TestRemoval(t *testing.T) {
	tree, words := makeTree(0, `a foolish consistency is the hobgoblin of little minds`)

	got := allWords(tree)
	if diff := cmp.Diff(stringset.New(words...).Elements(), got); diff != "" {
		t.Errorf("Original input differs from expected (-want, +got)\n%s", diff)
	}

	drop := stringset.New("a", "is", "of", "the")
	for w := range drop {
		if !tree.Remove(w) {
			t.Errorf("Remove(%q) returned false, wanted true", w)
		}
	}

	got = allWords(tree)
	want := stringset.New(words...).Diff(drop).Elements()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Tree after removal is incorrect (-want, +got)\n%s", diff)
	}
}

func TestInorderAfter(t *testing.T) {
	keys := []KV{
		{Key: "8"}, {Key: "6"}, {Key: "7"}, {Key: "5"},
		{Key: "3"}, {Key: "0"}, {Key: "9"},
	}
	tree := NewKeys(0, keys...)
	tests := []struct {
		key  Key
		want string
	}{
		{"A", ""},
		{"9", "9"},
		{"8", "8 9"},
		{"7", "7 8 9"},
		{"6", "6 7 8 9"},
		{"5", "5 6 7 8 9"},
		{"4", "5 6 7 8 9"},
		{"3", "3 5 6 7 8 9"},
		{"2", "3 5 6 7 8 9"},
		{"1", "3 5 6 7 8 9"},
		{"0", "0 3 5 6 7 8 9"},
		{"", "0 3 5 6 7 8 9"},
	}
	for _, test := range tests {
		want := strings.Fields(test.want)
		var got []string
		tree.InorderAfter(test.key, func(kv KV) bool {
			got = append(got, kv.Key)
			return true
		})
		if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("InorderAfter(%v) result differed from expected\n%s", test.key, diff)
		}
	}
}

# scapegoat

http://godoc.org/github.com/creachadair/scapegoat

This repository provides an implementation of Scapegoat Trees, as described in
https://people.csail.mit.edu/rivest/pubs/GR93.pdf

## Generated Code

The top-level `scapegoat` package implements a tree with `string` keys and
arbitrary values.  However, the implementation will work with any ordered key
type, and you can substitute a more specific value type. To generate a package
for your own type, use `go generate`:

```shell
mkdir pairtree
gofmt > pairtree/key.go <<EOF
package pairtree

//go:generate go run github.com/creachadair/scapegoat/mktree -p pairtree

// A Key is a pair of string values, ordered lexicographically.
type Key struct {
  A, B string
}

func keyLess(a, b Key) bool {
  return a.A < b.A || (a.A == b.A && a.B < b.B)
}

// A Value is an integer weight.
type Value int
EOF
go generate ./pairtree
```

As shown, you must provide a definition for the `Key` and `Value` types, as
well as a comparison function `keyLess(a, b Key) bool` to compare values of the
`Key` type.  The rest of the package is a straightforward copy of the main
package, apart from changing the name in the package clause.

## Visualization

One of the unit tests supports writing its output to a Graphviz `.dot` file so
that you can see what the output looks like for different weighting conditions.
To use this, include the `-dot` flag when running the tests, e.g.,

```shell
$ for w in 1 100 200 300 400 500 800 1000 ; do
     go test -dot w"$w".dot -balance $w
     dot -Tpng -o w"$w".png w"$w".dot
done
```

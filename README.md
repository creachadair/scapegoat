# scapegoat

http://godoc.org/bitbucket.org/creachadair/scapegoat

This repository provides an implementation of Scapegoat Trees, as described in
https://people.csail.mit.edu/rivest/pubs/GR93.pdf

## Visualization

One of the unit tests supports writing its output to a Graphviz `.dot` file so
that you can see what the output looks like for different weighting conditions.
To use this, include the `-dot` flag when running the tests, e.g.,

```shell
$ go test -dot w200.dot -balance 200
$ dot -Tpng -o w200.png w200.dot
```

# scapegoat

http://godoc.org/bitbucket.org/creachadair/scapegoat

This repository provides an implementation of Scapegoat Trees, as described in
https://people.csail.mit.edu/rivest/pubs/GR93.pdf

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

package scapegoat

// Key defines a string key for a scapegoat tree.
type Key = string

// keyLess reports whether a is ordered prior to b.
func keyLess(a, b Key) bool { return a < b }

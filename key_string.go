package scapegoat

// Key defines a string key for a scapegoat tree. This is the default key type
// for the base package in the module. Use the mktree tool to generate packages
// for other key types.
type Key = string

// keyLess reports whether a is ordered prior to b.
func keyLess(a, b Key) bool { return a < b }

package bench

//go:generate go run bitbucket.org/creachadair/scapegoat/mktree -p bench

// Key defines an int as a key for a scapegoat tree.
type Key = int

func keyLess(a, b int) bool { return a < b }

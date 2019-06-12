// Program mktree copies the scapegoat tree implementation source into a new
// package with the specified name. This is intended to be invoked from a go
// generate rule to fill in a package that provides a definition of a Key type
// and a keyLess function.
//
// See the bench subdirectory for an example of use.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	packageName = flag.String("p", "", "Output package name (required)")
	outputDir   = flag.String("output", "", "Output directory (default is '.')")
)

const thisPackage = "github.com/creachadair/scapegoat"

func main() {
	flag.Parse()
	if *packageName == "" {
		log.Fatal("You must provide a non-empty -package name")
	}

	// Load the package to find the source files to copy.
	pkgs, err := packages.Load(nil, thisPackage)
	if err != nil {
		log.Fatalf("Cannot find source package: %v", err)
	} else if len(pkgs) != 1 {
		log.Fatalf("No unique source package: %v", pkgs)
	}

	// If necessary, create the output directory.
	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Fatalf("Creating output directory: %v", err)
		}
	}

	// Copy the implementation sources, updating the name in the package clause
	// and the documentation comment.
	pkg := pkgs[0]
	rep := strings.NewReplacer(
		fmt.Sprintf("// Package %s", pkg.Name), fmt.Sprintf("// Package %s", *packageName),
		fmt.Sprintf("package %s\n", pkg.Name), fmt.Sprintf("package %s\n", *packageName),
	)
	for _, src := range pkg.GoFiles {
		base := filepath.Base(src)
		if strings.HasPrefix(base, "key_") {
			continue // skip the built-in default.
		}
		data, err := ioutil.ReadFile(src)
		if err != nil {
			log.Fatalf("Reading source failed: %v", err)
		}
		s := rep.Replace(string(data))
		out := filepath.Join(*outputDir, base)
		if err := ioutil.WriteFile(out, []byte(s), 0644); err != nil {
			log.Fatalf("Writing source failed: %v", err)
		}
	}
}

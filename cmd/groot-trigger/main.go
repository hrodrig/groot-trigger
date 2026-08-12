// Package main is the groot-trigger entrypoint.
// Full HTTP → Job behavior is implemented via GSD against docs/SPECIFICATIONS.md.
package main

import (
	"fmt"
	"os"
)

// Set via -ldflags at build time (Makefile / GoReleaser).
var (
	version   = "dev"
	commit    = "unknown"
	branch    = "unknown"
	buildDate = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-V") {
		fmt.Printf("groot-trigger %s commit=%s branch=%s date=%s\n", version, commit, branch, buildDate)
		return
	}

	fmt.Fprintf(os.Stderr, "groot-trigger %s — application not implemented yet\n", version)
	fmt.Fprintf(os.Stderr, "See docs/SPECIFICATIONS.md; implement with GSD.\n")
	os.Exit(2)
}

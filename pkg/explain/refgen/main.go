// Command refgen generates the committed markdown reference for the cloud
// kind catalog from the descriptors compiled into the binary, via the
// explain engine's markdown renderer: one reference.md co-located with each
// kind's protos, plus the catalog-level files (per-provider indexes, the
// root index, the foreign-key graph, the commons page).
//
// Run through `make generate-reference`. Generation is always whole-catalog:
// a kind's Referenced By section and every catalog-level file depend on
// every other kind's schema, so a scoped run could never leave the
// committed tree consistent. Output is deterministic: identical descriptors
// produce byte-identical files, and the drift test in this package holds
// committed files to that promise.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Refuse ANY argument: generation is always whole-catalog and writes
	// hundreds of committed files, and an exploratory invocation (a
	// `--help` probe, a guessed scoping flag) must never trigger it --
	// silently ignoring arguments has caused accidental whole-tree
	// regens on dirty checkouts twice.
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "refgen takes no arguments (got %v); it regenerates the ENTIRE committed reference tree from the current descriptors. Run through `make generate-reference` on a settled tree.\n", os.Args[1:])
		os.Exit(2)
	}

	repoRoot, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	summary, err := Generate(repoRoot)
	if err != nil {
		fatal(err)
	}

	for _, path := range summary.SortedPaths() {
		if err := os.WriteFile(filepath.Join(repoRoot, path), []byte(summary.Files[path]), 0644); err != nil {
			fatal(err)
		}
	}
	fmt.Printf("generated %d reference file(s)\n", len(summary.Files))

	if len(summary.MissingManifests) > 0 {
		fmt.Printf("\n%d kind(s) have no e2e/manifest.yaml (page rendered without an Example):\n", len(summary.MissingManifests))
		for _, kind := range summary.MissingManifests {
			fmt.Printf("  - %s\n", kind)
		}
	}
	if len(summary.InvalidManifests) > 0 {
		fmt.Printf("\n%d hack manifest(s) FAILED validation -- these are catalog bugs, fix the manifests:\n", len(summary.InvalidManifests))
		for _, invalid := range summary.InvalidManifests {
			fmt.Printf("  - %s: %v\n", invalid.Kind, invalid.Err)
		}
		os.Exit(1)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "refgen:", err)
	os.Exit(1)
}

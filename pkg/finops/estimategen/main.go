// Command estimategen writes the generated per-kind cost estimates.
// Run through `make generate-cost-estimates`. Generation is always
// whole-tree: the dead-price sweep needs every model's references, so a
// scoped run could never prove the price books clean. See generate.go for
// the join and the coherence checks.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	summary, err := Generate(repoRoot)
	if err != nil {
		fatal(err)
	}

	if len(summary.Problems) > 0 {
		fmt.Fprintf(os.Stderr, "%d coherence problem(s) between models, price books, and cost profiles -- nothing written:\n", len(summary.Problems))
		for _, problem := range summary.Problems {
			fmt.Fprintf(os.Stderr, "  - %s\n", problem)
		}
		os.Exit(1)
	}

	for _, path := range summary.SortedPaths() {
		if err := os.WriteFile(filepath.Join(repoRoot, path), []byte(summary.Files[path]), 0644); err != nil {
			fatal(err)
		}
	}
	fmt.Printf("generated %d cost estimate(s)\n", len(summary.Files))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "estimategen:", err)
	os.Exit(1)
}

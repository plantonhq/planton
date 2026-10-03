// Command defspack validates the skills/, agents/, and automations/
// definition trees and packages a definitions release: one deterministic zip
// per skill, each agent's instructions file, each automation's definition
// file, a definitions-manifest.json carrying the release version,
// per-artifact SHA-256 checksums, and the compatibility floor from
// skills/compat.yaml -- plus a definitions-browse.json describing the
// release as a per-file tree (see browse.go), which the release workflow
// verifies the browsable exploded layout against before uploading it.
//
// Run `go run ./pkg/skills/defspack` to validate only (the lint gate), or
// add -version and -out to package a release (the release workflow):
//
//	go run ./pkg/skills/defspack -version v0.4.0 -out build/definitions
//
// Validation is deliberately strict in both directions: every reference
// file SKILL.md cites must exist and be non-empty, and every file present
// under references/ must be cited. An empty or orphaned reference is
// content rot an agent would silently load at runtime, so it fails here
// instead.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	version := flag.String("version", "", "release version (vX.Y.Z) to stamp into the manifest; required with -out")
	outDir := flag.String("out", "", "directory to write release artifacts into; omit to validate only")
	root := flag.String("root", "", "repository root (defaults to the working directory)")
	// The release lanes pass this flag: the catalog skill ships
	// self-contained, its ~18MB assembled pack inside the archive. Skill
	// artifacts above the engines' 10MB inline gRPC lane ride the
	// capability-URL transfer lane (stigmer/stigmer#675, fixed upstream in
	// v3.12.3), so pushing an embedded artifact needs Stigmer SDK >=
	// v3.12.3 on the pushing side -- this repo's go.mod pins one, and any
	// consumer engine older than v3.12.3 refuses the push loudly rather
	// than mounting a truncated skill. The default stays false so a plain
	// packaging run needs no catalog tree; loading and VALIDATION always
	// assemble the pack either way, so a pull request that breaks it fails
	// the lint gate regardless of the flag.
	embedPack := flag.Bool("embed-catalog-pack", false, "package the catalog skill self-contained (its kinds/ reference pack inside the archive)")
	flag.Parse()

	repoRoot := *root
	if repoRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			fatal(err)
		}
		repoRoot = wd
	}

	tree, err := LoadTree(repoRoot)
	if err != nil {
		fatal(err)
	}
	if errs := Validate(tree); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "defspack: INVALID:", e)
		}
		os.Exit(1)
	}
	fmt.Printf("validated %d skill(s), %d agent(s), and %d automation(s)\n", len(tree.Skills), len(tree.Agents), len(tree.Automations))

	if *outDir == "" {
		return
	}
	if *version == "" {
		fatal(fmt.Errorf("-version is required when packaging with -out"))
	}
	if !*embedPack {
		StripPackFiles(tree)
	}
	manifest, err := PackageRelease(tree, *version, *outDir)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("packaged definitions release %s into %s (%d skill archive(s), %d agent file(s), %d automation file(s))\n",
		manifest.Version, filepath.Clean(*outDir), len(manifest.Skills), len(manifest.Agents), len(manifest.Automations))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "defspack:", err)
	os.Exit(1)
}

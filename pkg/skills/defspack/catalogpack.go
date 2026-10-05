package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// The multi-cloud-catalog skill ships SELF-CONTAINED: its archive carries
// the kind reference pack -- the generated reference pages, the
// indexes, the cross-reference graph, the commons page, the authored
// GUIDE.md / patterns wisdom layer, and the verified fact-sheet layer
// (per-kind cost/controls/permissions sidecars, the generated
// per-preset cost estimates, and the central compliance catalog with its
// framework crosswalks) -- assembled from the repository's catalog/ tree at
// package time. The facts travel with the skill to every engine, so a
// conversation never depends on a repo checkout or a network fetch to
// ground a schema claim or a price, and "one pack version per answer" holds
// by construction: the mounted skill and its pack are one artifact.
//
// The git tree stays single-sourced: catalog/ remains the only home of the
// pack's files, and nothing is copied into skills/. Assembly happens only
// here, inside the deterministic packager, so the CDN release, the hosted
// engine, and every local engine carry byte-identical content.
const catalogPackSkillSlug = "multi-cloud-catalog"

// packDirName is the directory inside the skill archive that receives the
// assembled pack. Paths below it are catalog-relative, preserving the
// repository's own layout (providers at the top, _docs/ for the commons,
// indexes and graph, _patterns/ for the pattern library) so relative links
// between pack pages keep resolving after extraction.
const packDirName = "kinds"

// Engine ceilings for one skill artifact (the serving engine refuses a push
// beyond 100MB compressed or 10,000 files). Validation enforces margins
// well inside them so catalog growth trips a lint failure with a clear
// message instead of a runtime push refusal.
const (
	maxSkillArchiveFiles = 9000
	maxSkillArchiveBytes = 80 << 20 // 80MB of stored (uncompressed) content
)

// packFileNames is the pack's frozen public contract, mirrored from the
// release content lane's reference-pack selection: files are selected by
// NAME, never by version-segment path, so the selection survives
// api-version directory renames and a contributor-edited file is the same
// file an agent reads. The fact-sheet sidecar names (cost.yaml,
// controls.yaml, permissions.yaml) are anatomy-enforced and unique to the
// fact-sheet layer, so name selection stays collision-free.
var packFileNames = map[string]bool{
	"reference.md":         true,
	"GUIDE.md":             true,
	"reference-index.md":   true,
	"reference-graph.yaml": true,
	"reference-commons.md": true,
	"cost.yaml":            true,
	"controls.yaml":        true,
	"permissions.yaml":     true,
}

// packPathPrefixes selects the pack's central trees by catalog-relative
// path prefix (the _patterns/ precedent): the generated per-preset cost
// estimates and the compliance catalog + framework crosswalks. Only .yaml
// documents ride from these trees -- their siblings (quantity models,
// derivation rules, price books) are pricing-pipeline machinery, not agent
// reading material, and the estimates already carry every consumable
// dollar with its source and verification date inline.
var packPathPrefixes = []string{
	"_pricing/estimates/",
	"_compliance/",
}

// StripPackFiles drops every skill's assembled pack from the tree -- what
// packaging does when -embed-catalog-pack is not passed (the release lanes
// pass it; a bare run packages pack-free artifacts). Validation runs BEFORE
// the strip, so the assembly stays continuously proven either way.
func StripPackFiles(tree *Tree) {
	for i := range tree.Skills {
		tree.Skills[i].PackFiles = nil
	}
}

// loadCatalogPack collects the reference pack from catalog/ under root,
// keyed by archive path (packDirName + "/" + catalog-relative path).
// A missing catalog/ tree yields an empty map -- like the rest of loading,
// shape problems surface in Validate, not here -- but real I/O errors fail
// immediately.
func loadCatalogPack(root string) (map[string][]byte, error) {
	catalogRoot := filepath.Join(root, "catalog")
	if _, err := os.Stat(catalogRoot); os.IsNotExist(err) {
		return map[string][]byte{}, nil
	}

	pack := map[string][]byte{}
	err := filepath.WalkDir(catalogRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(catalogRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		// Test fixtures are never part of the shipped pack.
		if strings.HasPrefix(rel, "_test/") || strings.Contains(rel, "/_test/") {
			return nil
		}
		isPattern := strings.HasPrefix(rel, "_patterns/") && strings.HasSuffix(rel, ".md")
		if !packFileNames[d.Name()] && !isPattern && !isPackCentralDoc(rel) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading pack file %s: %w", rel, err)
		}
		pack[packDirName+"/"+rel] = content
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking catalog tree: %w", err)
	}
	return pack, nil
}

// isPackCentralDoc reports whether a catalog-relative path is a .yaml
// document in one of the pack's central trees (packPathPrefixes).
func isPackCentralDoc(rel string) bool {
	if !strings.HasSuffix(rel, ".yaml") {
		return false
	}
	for _, prefix := range packPathPrefixes {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	return false
}

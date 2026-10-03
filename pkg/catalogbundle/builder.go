package catalogbundle

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"
)

// BuildInput names everything a bundle build consumes. The descriptor set is
// produced by `buf build` (the proto toolchain owns descriptor production;
// this builder owns assembly and self-verification).
type BuildInput struct {
	// DescriptorSetPath is the buf-built FileDescriptorSet (descriptors.binpb).
	DescriptorSetPath string
	// CatalogDir is the catalog tree holding conversions/ and presets/
	// per kind.
	CatalogDir string
	// ReleaseTag stamps the manifest (empty for local builds).
	ReleaseTag string
	// OutputPath is the bundle zip to write.
	OutputPath string
}

// Build assembles a catalog bundle zip from the inputs and returns the
// manifest it wrote. Every entry is checksummed; an empty descriptor set or
// a catalog tree with no presets fails the build -- a bundle missing its
// cargo must never ship.
func Build(input BuildInput) (*Manifest, error) {
	descriptors, err := os.ReadFile(input.DescriptorSetPath)
	if err != nil {
		return nil, fmt.Errorf("reading descriptor set: %w", err)
	}
	if len(descriptors) == 0 {
		return nil, fmt.Errorf("descriptor set %s is empty -- refusing to build a bundle without schemas", input.DescriptorSetPath)
	}

	entries := map[string][]byte{
		"descriptors.binpb": descriptors,
	}

	// Conversions and presets both live beside the kind
	// (<provider>/<kind>/conversions/*, <provider>/<kind>/presets/*): the
	// living kind is version-agnostic, so bundle keys carry no version.
	if err := collectFiles(entries, input.CatalogDir, "*/*/conversions/*", 4,
		func(parts []string) string {
			return "conversions/" + parts[0] + "/" + parts[1] + "/" + parts[3]
		}); err != nil {
		return nil, err
	}
	if err := collectFiles(entries, input.CatalogDir, "*/*/presets/*", 4,
		func(parts []string) string {
			return "presets/" + parts[0] + "/" + parts[1] + "/" + parts[3]
		}); err != nil {
		return nil, err
	}
	presetCount := 0
	for name := range entries {
		if strings.HasPrefix(name, "presets/") {
			presetCount++
		}
	}
	if presetCount == 0 {
		return nil, fmt.Errorf("no presets found under %s -- refusing to build a bundle without its preset cargo", input.CatalogDir)
	}

	// Fact-sheet cargo (cost/controls/permissions sidecars, estimates, and
	// their central documents) packs the exact bytes it parses -- the
	// parsed forms feed the entry summaries below, so summary and cargo
	// can never disagree within one build.
	cargo, err := collectCargo(input.CatalogDir, entries)
	if err != nil {
		return nil, err
	}

	// Catalog entries are projected from the tree plus the compiled kind
	// registry -- the projection itself refuses missing kinds, dead
	// module directories, and an empty result.
	entryFiles, err := projectEntries(input.CatalogDir, cargo)
	if err != nil {
		return nil, err
	}
	for name, content := range entryFiles {
		entries[name] = content
	}

	manifest := &Manifest{
		FormatVersion: FormatVersion,
		ReleaseTag:    input.ReleaseTag,
		Checksums:     map[string]string{},
	}
	for name, content := range entries {
		sum := sha256.Sum256(content)
		manifest.Checksums[name] = hex.EncodeToString(sum[:])
	}
	manifestYAML, err := yaml.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	out, err := os.Create(input.OutputPath)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	zw := zip.NewWriter(out)

	names := make([]string, 0, len(entries)+1)
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)

	if err := writeZipEntry(zw, "manifest.yaml", manifestYAML); err != nil {
		return nil, err
	}
	for _, name := range names {
		if err := writeZipEntry(zw, name, entries[name]); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return manifest, nil
}

// collectFiles gathers files matching the glob (relative to catalogDir)
// into bundle entries, naming each through entryName over the rel-path
// segments. Directories (e.g. conversions/testdata) never match the
// fixed-depth globs -- fixtures gate engines, they are not runtime cargo.
func collectFiles(entries map[string][]byte, catalogDir, glob string, depth int, entryName func(parts []string) string) error {
	matches, err := filepath.Glob(filepath.Join(catalogDir, glob))
	if err != nil {
		return err
	}
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return err
		}
		if info.IsDir() {
			continue
		}
		rel, err := filepath.Rel(catalogDir, match)
		if err != nil {
			return err
		}
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) != depth {
			continue
		}
		content, err := os.ReadFile(match)
		if err != nil {
			return err
		}
		entries[entryName(parts)] = content
	}
	return nil
}

func writeZipEntry(zw *zip.Writer, name string, content []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, strings.NewReader(string(content)))
	return err
}

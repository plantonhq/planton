package catalogbundle

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"sigs.k8s.io/yaml"
)

// Bundle is a loaded, checksum-verified catalog bundle.
type Bundle struct {
	Manifest Manifest
	// Files resolves every schema the bundle carries (kind messages,
	// validation rules in options, the kind registry enum).
	Files *protoregistry.Files
	// entries holds the raw bundle contents by entry name (conversions/**,
	// presets/**, entries/**, the fact-sheet cargo trees costs/**,
	// controls/**, permissions/**, estimates/**, compliance/**,
	// pricebooks/**, and descriptors.binpb).
	entries map[string][]byte
	// catalogEntries are the decoded entries/** documents, sorted by kind.
	catalogEntries []CatalogEntry
}

// Load opens a bundle zip, verifies every entry against the manifest's
// checksums, and parses the descriptor set. A bundle that fails ANY check is
// refused whole -- consumers never run on partially-verified data.
func Load(path string) (*Bundle, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("opening catalog bundle %s: %w", path, err)
	}
	defer reader.Close()

	entries := map[string][]byte{}
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("reading bundle entry %s: %w", file.Name, err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("reading bundle entry %s: %w", file.Name, err)
		}
		entries[file.Name] = content
	}

	manifestRaw, ok := entries["manifest.yaml"]
	if !ok {
		return nil, fmt.Errorf("catalog bundle %s has no manifest.yaml -- not a bundle", path)
	}
	delete(entries, "manifest.yaml")
	var manifest Manifest
	if err := yaml.UnmarshalStrict(manifestRaw, &manifest); err != nil {
		return nil, fmt.Errorf("catalog bundle manifest is malformed: %w", err)
	}
	if manifest.FormatVersion != FormatVersion {
		return nil, fmt.Errorf("catalog bundle format %q is not understood by this consumer (wants %q) -- upgrade the consumer or rebuild the bundle", manifest.FormatVersion, FormatVersion)
	}

	// Self-verification: every entry must match its recorded checksum, and
	// every recorded checksum must have its entry.
	for name, content := range entries {
		want, recorded := manifest.Checksums[name]
		if !recorded {
			return nil, fmt.Errorf("bundle entry %s is not in the manifest -- the bundle was tampered with or mis-built", name)
		}
		sum := sha256.Sum256(content)
		if hex.EncodeToString(sum[:]) != want {
			return nil, fmt.Errorf("bundle entry %s fails its checksum -- the bundle is corrupt", name)
		}
	}
	for name := range manifest.Checksums {
		if _, present := entries[name]; !present {
			return nil, fmt.Errorf("the manifest records %s but the bundle does not contain it -- the bundle is incomplete", name)
		}
	}

	var fds descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(entries["descriptors.binpb"], &fds); err != nil {
		return nil, fmt.Errorf("parsing the bundle's descriptor set: %w", err)
	}
	files, err := protodesc.NewFiles(&fds)
	if err != nil {
		return nil, fmt.Errorf("building a registry from the bundle's descriptors: %w", err)
	}

	catalogEntries, err := parseCatalogEntries(entries)
	if err != nil {
		return nil, err
	}

	return &Bundle{Manifest: manifest, Files: files, entries: entries, catalogEntries: catalogEntries}, nil
}

// ConversionSpecs returns the bundle's conversion spec contents keyed by
// entry name (conversions/<provider>/<kind>/<file>.yaml), sorted.
func (b *Bundle) ConversionSpecs() map[string][]byte {
	return b.subtree("conversions/")
}

// Presets returns the bundle's preset contents keyed by entry name
// (presets/<provider>/<kind>/<file>), sorted.
func (b *Bundle) Presets() map[string][]byte {
	return b.subtree("presets/")
}

// CatalogEntries returns the bundle's catalog entries sorted by kind name,
// decoded and shape-checked at load. Whether the set agrees with the kind
// registry is the conformance gate's verdict, not the loader's.
func (b *Bundle) CatalogEntries() []CatalogEntry {
	return b.catalogEntries
}

// CostProfiles returns the bundle's catalog kind cost profiles keyed by entry
// name (costs/<provider>/<kind>.yaml), sorted. Like presets, the contents
// are raw bytes -- consumers parse what they read, and the conformance gate
// has already proven every document parses against its schema.
func (b *Bundle) CostProfiles() map[string][]byte {
	return b.subtree(costsPrefix)
}

// ControlProfiles returns the bundle's catalog kind control profiles keyed by
// entry name (controls/<provider>/<kind>.yaml), sorted.
func (b *Bundle) ControlProfiles() map[string][]byte {
	return b.subtree(controlsPrefix)
}

// Permissions returns the bundle's kind permission manifests keyed by
// entry name (permissions/<provider>/<kind>.yaml), sorted.
func (b *Bundle) Permissions() map[string][]byte {
	return b.subtree(permissionsPrefix)
}

// CostEstimates returns the bundle's generated per-preset cost estimates
// keyed by entry name (estimates/<provider>/<kind>.yaml), sorted.
func (b *Bundle) CostEstimates() map[string][]byte {
	return b.subtree(estimatesPrefix)
}

// CostDerivations returns the bundle's machine-executable cost derivations
// keyed by entry name (derivations/<provider>/<kind>.yaml), sorted. A
// server-side estimator evaluates these rules against live manifests with
// the aboard price books -- zero external calls; kinds without a
// derivation aboard are estimated at their preset ranges only.
func (b *Bundle) CostDerivations() map[string][]byte {
	return b.subtree(derivationsPrefix)
}

// Compliance returns the bundle's central compliance documents keyed by
// entry name (compliance/controls-catalog.yaml and
// compliance/frameworks/<framework>.yaml), sorted.
func (b *Bundle) Compliance() map[string][]byte {
	return b.subtree(compliancePrefix)
}

// PriceBooks returns the bundle's pinned per-provider price books keyed by
// entry name (pricebooks/<provider>.yaml), sorted.
func (b *Bundle) PriceBooks() map[string][]byte {
	return b.subtree(pricebooksPrefix)
}

func (b *Bundle) subtree(prefix string) map[string][]byte {
	out := map[string][]byte{}
	names := make([]string, 0, len(b.entries))
	for name := range b.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			out[name] = b.entries[name]
		}
	}
	return out
}

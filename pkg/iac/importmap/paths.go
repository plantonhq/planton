package importmap

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
)

const (
	catalogRoot = "catalog"

	// ProviderCatalogRelPath is the provider import catalog's path relative to
	// the provider directory (e.g. aws/aa_import/catalog.yaml).
	ProviderCatalogRelPath = "aa_import/catalog.yaml"
)

// ProviderCatalogPath returns the absolute path to a provider's import catalog.
func ProviderCatalogPath(repoRoot, provider string) string {
	return filepath.Join(repoRoot, catalogRoot, provider, ProviderCatalogRelPath)
}

// CatalogKindImportMapPath returns the absolute path to a kind's import
// map (e.g. awss3bucket/iac/import-map.yaml — it sits next to the module it
// maps, at the kind root). The kind is validated against the kind
// registry, so an unregistered kind name fails here rather than
// composing a path that does not exist.
func CatalogKindImportMapPath(repoRoot, provider, kindDir string) (string, error) {
	if _, err := catalogkindreflect.KindVersionDir(kindDir); err != nil {
		return "", errors.Wrapf(err, "cannot locate the import map for %s/%s", provider, kindDir)
	}
	return filepath.Join(repoRoot, catalogRoot, provider, kindDir, "iac", "import-map.yaml"), nil
}

// SplitAttributePath splits a declared attribute sub-path (a provider
// catalog config-only sub-path or a kind map import-normalized
// sub-path) into its segments. Plain segments are dot-separated:
// "spec.update_strategy" → ["spec", "update_strategy"]. A segment may
// also be written bracket-quoted — `data["password.db"]` → ["data",
// "password.db"] — for map keys that themselves contain dots
// (Kubernetes Secret data keys are the canonical case), which a plain
// dotted path cannot express. An unterminated bracket segment is kept
// verbatim so the mismatch fails loud downstream instead of silently
// tolerating the wrong key.
func SplitAttributePath(path string) []string {
	var segments []string
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			segments = append(segments, current.String())
			current.Reset()
		}
	}
	for i := 0; i < len(path); {
		switch {
		case strings.HasPrefix(path[i:], `["`):
			flush()
			end := strings.Index(path[i+2:], `"]`)
			if end < 0 {
				current.WriteString(path[i:])
				i = len(path)
				continue
			}
			segments = append(segments, path[i+2:i+2+end])
			i += 2 + end + 2
		case path[i] == '.':
			flush()
			i++
		default:
			current.WriteByte(path[i])
			i++
		}
	}
	flush()
	return segments
}

package importmap

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

// DiscoverCatalogKindImportMaps walks the provider tree and returns every
// kind that ships an import map, keyed by provider directory name
// (e.g. "aws" -> ["awss3bucket", "awsvpc", ...]), with both providers and
// kinds sorted for deterministic iteration.
//
// File presence IS enrollment: a kind with a
// `v1/iac/import-map.yaml` is validated by the offline conformance guard,
// picked up by the E2E import round-trip gate, and bundled into the
// platform's catalog data for the import wizard — all from this one signal.
// There is deliberately no separate allowlist to keep in sync; an authored
// map that would not be checked, or a checked map that would not ship, is a
// divergence this single signal makes impossible.
func DiscoverCatalogKindImportMaps(repoRoot string) (map[string][]string, error) {
	providersDir := filepath.Join(repoRoot, catalogRoot)
	providerEntries, err := os.ReadDir(providersDir)
	if err != nil {
		return nil, errors.Wrapf(err, "reading providers directory %s", providersDir)
	}

	discovered := map[string][]string{}
	for _, providerEntry := range providerEntries {
		if !providerEntry.IsDir() {
			continue
		}
		provider := providerEntry.Name()
		kindEntries, err := os.ReadDir(filepath.Join(providersDir, provider))
		if err != nil {
			return nil, errors.Wrapf(err, "reading provider directory %s", provider)
		}
		for _, kindEntry := range kindEntries {
			if !kindEntry.IsDir() {
				continue
			}
			kindDir := kindEntry.Name()
			if HasCatalogKindImportMap(repoRoot, provider, kindDir) {
				discovered[provider] = append(discovered[provider], kindDir)
			}
		}
	}
	for provider := range discovered {
		sort.Strings(discovered[provider])
	}
	return discovered, nil
}

package importmap

import (
	"os"

	"github.com/pkg/errors"

	kindv1 "github.com/plantonhq/planton/iac/catalogkindimportmap/v1"
	providerv1 "github.com/plantonhq/planton/iac/providerimportcatalog/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// LoadProviderCatalog reads and parses a provider's import catalog from disk.
func LoadProviderCatalog(repoRoot, provider string) (*providerv1.ProviderImportCatalog, error) {
	catalog := &providerv1.ProviderImportCatalog{}
	if err := protobufyaml.Load(ProviderCatalogPath(repoRoot, provider), catalog); err != nil {
		return nil, errors.Wrapf(err, "loading provider import catalog for %s", provider)
	}
	return catalog, nil
}

// LoadCatalogKindImportMap reads and parses a kind's import map from disk.
func LoadCatalogKindImportMap(repoRoot, provider, kindDir string) (*kindv1.CatalogKindImportMap, error) {
	path, err := CatalogKindImportMapPath(repoRoot, provider, kindDir)
	if err != nil {
		return nil, err
	}
	m := &kindv1.CatalogKindImportMap{}
	if err := protobufyaml.Load(path, m); err != nil {
		return nil, errors.Wrapf(err, "loading catalog kind import map for %s/%s", provider, kindDir)
	}
	return m, nil
}

// HasCatalogKindImportMap reports whether a kind ships an import map --
// the "is this kind mapped?" check callers use before offering derived import.
// A name that does not resolve to a registered kind has no import map.
func HasCatalogKindImportMap(repoRoot, provider, kindDir string) bool {
	path, err := CatalogKindImportMapPath(repoRoot, provider, kindDir)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

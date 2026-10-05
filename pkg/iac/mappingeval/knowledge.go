package mappingeval

import (
	"github.com/pkg/errors"
	kindv1 "github.com/plantonhq/planton/iac/catalogkindimportmap/v1"
	providerv1 "github.com/plantonhq/planton/iac/providerimportcatalog/v1"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/importmap"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// ScoreOptionsFromCatalog derives the scorer's declared knowledge from the
// artifacts the import subsystem already maintains -- the provider import
// catalog and the kinds' import maps. Nothing is authored per suite:
// when a catalog entry gains a config-only attribute or a kind's
// recipe changes, the scorer's expectations follow automatically.
func ScoreOptionsFromCatalog(repoRoot, provider string, kindDirs []string) (ScoreOptions, error) {
	catalog, err := importmap.LoadProviderCatalog(repoRoot, provider)
	if err != nil {
		return ScoreOptions{}, err
	}

	// Config-only attributes can never round-trip through a scan (they
	// exist only in IaC configuration), so their spec fields leave the spec
	// axis. The catalog declares them per resource type; the union is safe
	// because exclusion is by field NAME and the naming convention keeps
	// spec fields aligned with the module attributes they drive.
	excluded := map[string]bool{}
	for _, rt := range catalog.GetSpec().GetResourceTypes() {
		for _, attr := range rt.GetConfigOnlyAttributes() {
			excluded[attr] = true
		}
	}

	nameDerived, err := nameDerivedIdentity(repoRoot, provider, kindDirs, catalog)
	if err != nil {
		return ScoreOptions{}, err
	}

	return ScoreOptions{
		ExcludedSpecFields:  excluded,
		NameDerivedIdentity: nameDerived,
	}, nil
}

// nameDerivedIdentity finds, per kind, the Cloud Control type
// whose claimed identifier must equal metadata.name: the kind's import
// map derives a placeholder from_metadata_name, and a catalog resource type
// with a declared scan-side name imports by exactly that placeholder.
func nameDerivedIdentity(repoRoot, provider string, kindDirs []string, catalog *providerv1.ProviderImportCatalog) (map[catalogkind.CatalogKind]string, error) {
	result := map[catalogkind.CatalogKind]string{}
	for _, kindDir := range kindDirs {
		if !importmap.HasCatalogKindImportMap(repoRoot, provider, kindDir) {
			continue
		}
		m, err := importmap.LoadCatalogKindImportMap(repoRoot, provider, kindDir)
		if err != nil {
			return nil, errors.Wrapf(err, "import map for %s", kindDir)
		}
		nameDerivedValues := map[string]bool{}
		for _, v := range m.GetSpec().GetValues() {
			for _, d := range v.GetDerivations() {
				if _, ok := d.GetSource().(*kindv1.ImportValueDerivation_FromMetadataName); ok {
					nameDerivedValues[v.GetName()] = true
				}
			}
		}
		if len(nameDerivedValues) == 0 {
			continue
		}
		kind := catalogkindreflect.KindFromString(kindDir)
		for _, rt := range catalog.GetSpec().GetResourceTypes() {
			if rt.GetCloudControlTypeName() == "" {
				continue
			}
			placeholders := importmap.Placeholders(rt.GetIdFormat())
			if len(placeholders) == 1 && nameDerivedValues[placeholders[0]] {
				result[kind] = rt.GetCloudControlTypeName()
				break
			}
		}
	}
	return result, nil
}

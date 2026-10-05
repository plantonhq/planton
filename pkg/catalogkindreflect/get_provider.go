package catalogkindreflect

import (
	"github.com/plantonhq/planton/shared/catalogkind"
)

// GetProvider returns the catalog **provider** recorded in the
// (provider) enum‑value option of the given CatalogKind.
//
// If the kind is unknown or the option is absent, the function returns the
// “unspecified” sentinel.
func GetProvider(
	kind catalogkind.CatalogKind,
) catalogkind.CatalogProvider {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		return catalogkind.CatalogProvider_catalog_provider_unspecified
	}
	return kindMeta.Provider
}

package catalogkindreflect

import (
	"strings"

	"github.com/plantonhq/planton/shared/catalogkind"
)

// ProviderDirName returns the catalog directory name of one cloud provider:
// the lowercase enum name with separators stripped (digital_ocean ->
// digitalocean). This is the canonical enum-name -> directory rule the
// catalog tree is laid out by; every tool that composes a
// catalog/<provider>/ path from the enum must go through it, because enum
// names may carry underscores while catalog directories never do.
//
// The kind-map code generator applies the same rule when deriving provider
// package paths; it carries its own copy because the codegen build tag
// exists precisely so the generator never imports this package.
func ProviderDirName(provider catalogkind.CatalogProvider) string {
	return strings.ToLower(strings.ReplaceAll(provider.String(), "_", ""))
}

// ProviderFromString resolves a provider name to its enum value, accepting
// the enum name (digital_ocean), the catalog directory name (digitalocean),
// and hyphenated forms (digital-ocean) -- the same normalization
// KindFromString applies to kind names. Returns
// catalog_provider_unspecified when nothing matches.
func ProviderFromString(providerString string) catalogkind.CatalogProvider {
	want := normalizeProviderName(providerString)
	if want == "" {
		return catalogkind.CatalogProvider_catalog_provider_unspecified
	}
	for name, value := range catalogkind.CatalogProvider_value {
		if value == 0 {
			continue
		}
		if normalizeProviderName(name) == want {
			return catalogkind.CatalogProvider(value)
		}
	}
	return catalogkind.CatalogProvider_catalog_provider_unspecified
}

func normalizeProviderName(s string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return strings.ToLower(s)
}

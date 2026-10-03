package catalogkindreflect

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

// The catalog tree never carries underscores in provider directory names,
// so the derivation must strip them (digital_ocean -> digitalocean) and be
// a no-op for providers whose enum name already is the directory name.
func TestProviderDirName(t *testing.T) {
	cases := map[catalogkind.CatalogProvider]string{
		catalogkind.CatalogProvider_gcp:           "gcp",
		catalogkind.CatalogProvider_aws:           "aws",
		catalogkind.CatalogProvider_digital_ocean: "digitalocean",
	}
	for provider, want := range cases {
		if got := ProviderDirName(provider); got != want {
			t.Errorf("ProviderDirName(%s) = %q, want %q", provider, got, want)
		}
	}
	for name, value := range catalogkind.CatalogProvider_value {
		if value == 0 {
			continue
		}
		dir := ProviderDirName(catalogkind.CatalogProvider(value))
		if strings.ContainsAny(dir, "_-") || dir != strings.ToLower(dir) {
			t.Errorf("ProviderDirName(%s) = %q -- directory names are lowercase with no separators", name, dir)
		}
	}
}

// Resolution accepts the enum name, the directory name, and hyphenated
// forms, and refuses everything else -- including the empty string, which
// must never resolve to a provider.
func TestProviderFromString(t *testing.T) {
	cases := map[string]catalogkind.CatalogProvider{
		"gcp":           catalogkind.CatalogProvider_gcp,
		"digital_ocean": catalogkind.CatalogProvider_digital_ocean,
		"digitalocean":  catalogkind.CatalogProvider_digital_ocean,
		"digital-ocean": catalogkind.CatalogProvider_digital_ocean,
		"DigitalOcean":  catalogkind.CatalogProvider_digital_ocean,
		"":              catalogkind.CatalogProvider_catalog_provider_unspecified,
		"not-a-cloud":   catalogkind.CatalogProvider_catalog_provider_unspecified,
	}
	for input, want := range cases {
		if got := ProviderFromString(input); got != want {
			t.Errorf("ProviderFromString(%q) = %s, want %s", input, got, want)
		}
	}
}

// Round trip: every real provider's directory name resolves back to the
// provider, so the two helpers can never drift apart.
func TestProviderDirNameRoundTrip(t *testing.T) {
	for name, value := range catalogkind.CatalogProvider_value {
		if value == 0 {
			continue
		}
		provider := catalogkind.CatalogProvider(value)
		if got := ProviderFromString(ProviderDirName(provider)); got != provider {
			t.Errorf("ProviderFromString(ProviderDirName(%s)) = %s -- the round trip must be lossless", name, got)
		}
	}
}

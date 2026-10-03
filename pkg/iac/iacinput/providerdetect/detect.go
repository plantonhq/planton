package providerdetect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// DetectionResult contains the results of provider detection from a manifest.
type DetectionResult struct {
	// Kind is the detected CatalogKind from the manifest
	Kind catalogkind.CatalogKind
	// Provider is the provider required for this kind
	Provider catalogkind.CatalogProvider
	// KindName is the human-readable name of the kind (e.g., "GkeCluster")
	KindName string
	// ProviderName is the human-readable name of the provider (e.g., "gcp")
	ProviderName string
	// RequiresProviderConfig indicates if this provider needs credentials
	RequiresProviderConfig bool
}

// DetectFromManifest extracts the CatalogKind and required provider from manifest YAML.
// This is the primary entry point for provider detection.
func DetectFromManifest(manifestYaml []byte) (*DetectionResult, error) {
	// Extract CatalogKind from manifest
	kind, err := catalogkindreflect.ExtractKindFromYaml(manifestYaml)
	if err != nil {
		return nil, errors.Wrap(err, "failed to detect catalog kind from manifest")
	}

	// Get provider for this kind
	provider := catalogkindreflect.GetProvider(kind)

	// Build result
	result := &DetectionResult{
		Kind:                   kind,
		Provider:               provider,
		KindName:               kind.String(),
		ProviderName:           provider.String(),
		RequiresProviderConfig: requiresProviderConfig(provider),
	}

	return result, nil
}

// requiresProviderConfig returns true if the provider requires credential configuration.
func requiresProviderConfig(provider catalogkind.CatalogProvider) bool {
	switch provider {
	case catalogkind.CatalogProvider_catalog_provider_unspecified:
		return false
	case catalogkind.CatalogProvider__test:
		return false
	default:
		// All other providers require credentials
		return true
	}
}

// ProviderDisplayName returns a human-friendly display name for the provider.
func ProviderDisplayName(provider catalogkind.CatalogProvider) string {
	switch provider {
	case catalogkind.CatalogProvider_auth0:
		return "Auth0"
	case catalogkind.CatalogProvider_aws:
		return "AWS"
	case catalogkind.CatalogProvider_azure:
		return "Azure"
	case catalogkind.CatalogProvider_cloudflare:
		return "Cloudflare"
	case catalogkind.CatalogProvider_digital_ocean:
		return "DigitalOcean"
	case catalogkind.CatalogProvider_gcp:
		return "GCP"
	case catalogkind.CatalogProvider_kubernetes:
		return "Kubernetes"
	case catalogkind.CatalogProvider_openfga:
		return "OpenFGA"
	case catalogkind.CatalogProvider_stripe:
		return "Stripe"
	default:
		return provider.String()
	}
}

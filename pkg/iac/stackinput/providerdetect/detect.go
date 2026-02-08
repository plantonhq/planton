package providerdetect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
)

// DetectionResult contains the results of provider detection from a manifest.
type DetectionResult struct {
	// Kind is the detected CloudResourceKind from the manifest
	Kind cloudresourcekind.CloudResourceKind
	// Provider is the provider required for this kind
	Provider cloudresourcekind.CloudResourceProvider
	// KindName is the human-readable name of the kind (e.g., "GkeCluster")
	KindName string
	// ProviderName is the human-readable name of the provider (e.g., "gcp")
	ProviderName string
	// RequiresProviderConfig indicates if this provider needs credentials
	RequiresProviderConfig bool
}

// DetectFromManifest extracts the CloudResourceKind and required provider from manifest YAML.
// This is the primary entry point for provider detection.
func DetectFromManifest(manifestYaml []byte) (*DetectionResult, error) {
	// Extract CloudResourceKind from manifest
	kind, err := crkreflect.ExtractKindFromYaml(manifestYaml)
	if err != nil {
		return nil, errors.Wrap(err, "failed to detect cloud resource kind from manifest")
	}

	// Get provider for this kind
	provider := crkreflect.GetProvider(kind)

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
func requiresProviderConfig(provider cloudresourcekind.CloudResourceProvider) bool {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified:
		return false
	case cloudresourcekind.CloudResourceProvider__test:
		return false
	default:
		// All other providers require credentials
		return true
	}
}

// ProviderDisplayName returns a human-friendly display name for the provider.
func ProviderDisplayName(provider cloudresourcekind.CloudResourceProvider) string {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_atlas:
		return "MongoDB Atlas"
	case cloudresourcekind.CloudResourceProvider_auth0:
		return "Auth0"
	case cloudresourcekind.CloudResourceProvider_aws:
		return "AWS"
	case cloudresourcekind.CloudResourceProvider_azure:
		return "Azure"
	case cloudresourcekind.CloudResourceProvider_civo:
		return "Civo"
	case cloudresourcekind.CloudResourceProvider_cloudflare:
		return "Cloudflare"
	case cloudresourcekind.CloudResourceProvider_confluent:
		return "Confluent"
	case cloudresourcekind.CloudResourceProvider_digital_ocean:
		return "DigitalOcean"
	case cloudresourcekind.CloudResourceProvider_gcp:
		return "GCP"
	case cloudresourcekind.CloudResourceProvider_kubernetes:
		return "Kubernetes"
	case cloudresourcekind.CloudResourceProvider_open_fga:
		return "OpenFGA"
	case cloudresourcekind.CloudResourceProvider_snowflake:
		return "Snowflake"
	case cloudresourcekind.CloudResourceProvider_openstack:
		return "OpenStack"
	default:
		return provider.String()
	}
}

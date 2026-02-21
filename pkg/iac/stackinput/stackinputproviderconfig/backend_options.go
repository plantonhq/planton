package stackinputproviderconfig

import (
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

// BuildFromProto creates a ProviderConfig from a proto message by writing it to a temporary file.
// This is used by backend services that receive credentials via API rather than from a file.
// Returns the ProviderConfig, a cleanup function to remove the temp file, and any error.
func BuildFromProto(
	credentialProto proto.Message,
	provider cloudresourcekind.CloudResourceProvider,
) (*ProviderConfig, func(), error) {
	if credentialProto == nil {
		return &ProviderConfig{
			Path:     "",
			Provider: provider,
		}, func() {}, nil
	}

	// Create temp file for the provider config
	tmpFile, err := os.CreateTemp("", "provider-config-*.yaml")
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create temp file for provider config")
	}

	// Marshal proto to YAML
	protoYaml, err := marshalProtoToYaml(credentialProto)
	if err != nil {
		os.Remove(tmpFile.Name())
		return nil, nil, errors.Wrap(err, "failed to marshal provider config proto to YAML")
	}

	// Write to temp file
	if _, err := tmpFile.Write(protoYaml); err != nil {
		os.Remove(tmpFile.Name())
		return nil, nil, errors.Wrap(err, "failed to write provider config to temp file")
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return nil, nil, errors.Wrap(err, "failed to close temp file")
	}

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return &ProviderConfig{
		Path:     tmpFile.Name(),
		Provider: provider,
	}, cleanup, nil
}

// ValidateProviderConfig validates that the provider config has a path set
// for providers that require credentials.
func ValidateProviderConfig(
	provider cloudresourcekind.CloudResourceProvider,
	providerConfig *ProviderConfig,
	resourceName string,
) error {
	if providerConfig == nil || providerConfig.Path == "" {
		// Check if this provider requires credentials
		switch provider {
		case cloudresourcekind.CloudResourceProvider_aws:
			return errors.Errorf(
				"AWS credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_gcp:
			return errors.Errorf(
				"GCP credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_azure:
			return errors.Errorf(
				"Azure credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_atlas:
			return errors.Errorf(
				"Atlas credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_auth0:
			return errors.Errorf(
				"Auth0 credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_cloudflare:
			return errors.Errorf(
				"Cloudflare credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_confluent:
			return errors.Errorf(
				"Confluent credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_snowflake:
			return errors.Errorf(
				"Snowflake credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_kubernetes:
			return errors.Errorf(
				"Kubernetes credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_openfga:
			return errors.Errorf(
				"OpenFGA credentials required for resource '%s'. Provide credentials via provider_config in API request",
				resourceName,
			)
		case cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified:
			// No credentials needed for unspecified provider
			return nil
		default:
			// For other providers (civo, digitalocean, etc.), credentials are optional
			return nil
		}
	}
	return nil
}

// marshalProtoToYaml marshals a proto message to YAML format using JSON field names.
func marshalProtoToYaml(msg proto.Message) ([]byte, error) {
	// Use protojson to get the proper JSON field names, then convert to YAML
	// This ensures consistency with how the rest of the codebase handles proto->YAML conversion
	jsonBytes, err := protojson.Marshal(msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal proto to JSON")
	}

	// Convert JSON to YAML
	var data interface{}
	if err := yaml.Unmarshal(jsonBytes, &data); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON")
	}

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal to YAML")
	}

	return yamlBytes, nil
}

package providerenvvars

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
	"gopkg.in/yaml.v3"
)

// ProviderConfigKey is the key used to store provider configuration in stack input YAML.
// All providers use this same key - the correct provider is determined by the target's api_version/kind.
const ProviderConfigKey = "provider_config"

// GetEnvVars takes stack input YAML and returns provider-specific environment variables.
// It extracts the CloudResourceKind from target, determines the provider using crkreflect,
// and loads the provider_config into the correct proto type.
//
// This function is IaC-agnostic - it can be used by both Pulumi and Tofu.
func GetEnvVars(stackInputYaml string) (map[string]string, error) {
	return GetEnvVarsWithOptions(stackInputYaml, Options{})
}

// Options contains optional parameters for GetEnvVars.
type Options struct {
	// FileCacheLoc is the directory where temporary files (like kubeconfig) can be written.
	// Required for Kubernetes provider.
	FileCacheLoc string
}

// GetEnvVarsWithOptions takes stack input YAML and options, returns provider-specific environment variables.
func GetEnvVarsWithOptions(stackInputYaml string, opts Options) (map[string]string, error) {
	// 1. Parse stack input YAML
	stackInputMap := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(stackInputYaml), &stackInputMap); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal stack input yaml")
	}

	// 2. Extract target YAML to determine the CloudResourceKind
	targetYaml, err := extractTargetYaml(stackInputMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract target from stack input")
	}

	// 3. Use crkreflect to get CloudResourceKind from target YAML
	kind, err := crkreflect.ExtractKindFromYaml(targetYaml)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract cloud resource kind from target")
	}

	// 4. Use crkreflect to get CloudResourceProvider from kind
	provider := crkreflect.GetProvider(kind)
	if provider == cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified {
		// No provider config needed for unspecified provider
		return map[string]string{}, nil
	}

	// 5. Check if provider_config exists in stack input
	providerConfigYaml, exists := extractProviderConfigYaml(stackInputMap)
	if !exists {
		// No provider_config in stack input - return empty map
		return map[string]string{}, nil
	}

	// 6. Load provider_config and convert to env vars based on provider
	return loadProviderEnvVars(providerConfigYaml, provider, opts)
}

// extractTargetYaml extracts the target field from stack input and marshals it back to YAML bytes.
func extractTargetYaml(stackInputMap map[string]interface{}) ([]byte, error) {
	target, ok := stackInputMap["target"]
	if !ok {
		return nil, errors.New("stack input does not contain 'target' field")
	}

	targetYaml, err := yaml.Marshal(target)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal target to yaml")
	}

	return targetYaml, nil
}

// extractProviderConfigYaml extracts the provider_config field from stack input and marshals it to YAML bytes.
func extractProviderConfigYaml(stackInputMap map[string]interface{}) ([]byte, bool) {
	providerConfig, ok := stackInputMap[ProviderConfigKey]
	if !ok {
		return nil, false
	}

	providerConfigYaml, err := yaml.Marshal(providerConfig)
	if err != nil {
		return nil, false
	}

	return providerConfigYaml, true
}

// loadProviderEnvVars loads the provider config YAML and returns environment variables based on the provider type.
func loadProviderEnvVars(providerConfigYaml []byte, provider cloudresourcekind.CloudResourceProvider, opts Options) (map[string]string, error) {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_open_fga:
		return loadOpenFgaEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_gcp:
		return loadGcpEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_azure:
		return loadAzureEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_aws:
		return loadAwsEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_atlas:
		return loadAtlasEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_auth0:
		return loadAuth0EnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_confluent:
		return loadConfluentEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_kubernetes:
		return loadKubernetesEnvVars(providerConfigYaml, opts.FileCacheLoc)
	case cloudresourcekind.CloudResourceProvider_snowflake:
		return loadSnowflakeEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_openstack:
		return loadOpenStackEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_scaleway:
		return loadScalewayEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_alicloud:
		return loadAlicloudEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_oci:
		return loadOciEnvVars(providerConfigYaml)
	default:
		// Unknown or unspecified provider - no env vars needed
		return map[string]string{}, nil
	}
}

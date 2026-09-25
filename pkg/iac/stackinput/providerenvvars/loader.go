package providerenvvars

import (
	"os"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/provider/aws/awswebidentity"
	"github.com/plantonhq/planton/pkg/kubernetes/kubeconfig"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"gopkg.in/yaml.v3"
)

// runningInCluster answers whether this process is a pod with its own
// ServiceAccount identity; a variable so a test can stand in a pod.
var runningInCluster = kubeconfig.RunningInCluster

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

	// ResolveAwsWebIdentity, when true, makes the AWS loader perform the STS
	// AssumeRoleWithWebIdentity exchange for keyless (oidc) provider
	// configs and emit the resulting short-lived credentials as AWS_* env vars. The
	// tofu/terraform execution path sets this (its HCL `provider "aws" {}` block is empty, so
	// credentials must arrive via env vars); the pulumi path leaves it false because its
	// in-program builder owns provider auth -- resolving here would trigger a wasteful STS
	// call whose output is shadowed by the state-backend keys anyway.
	ResolveAwsWebIdentity bool

	// KubeContext selects the kubeconfig context for a Kubernetes deploy (the
	// --kube-context flag, else the manifest's context label). It is exported
	// as KUBE_CTX, the name the Terraform kubernetes and helm providers and
	// the Pulumi provider getter all read. Empty means the kubeconfig's
	// current context.
	KubeContext string
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

	// 5. Read provider_config (may be absent for ambient-credential runs).
	providerConfigYaml, hasProviderConfig := extractProviderConfigYaml(stackInputMap)

	// 6. AWS is handled here -- NOT in loadProviderEnvVars -- because the AWS tofu modules ship
	//    an empty `provider "aws" {}` block, so both region and credentials are injection-driven:
	//    AWS_REGION is a RESOURCE property that must be emitted even when there is no
	//    provider_config (the standalone-CLI ambient case), and keyless connections require an
	//    STS exchange. Every other provider keeps the simple provider_config -> env-var mapping.
	if provider == cloudresourcekind.CloudResourceProvider_aws {
		resourceRegion := extractTargetSpecRegion(stackInputMap)
		return loadAwsEnvVars(providerConfigYaml, hasProviderConfig, resourceRegion, opts,
			awswebidentity.ResolveCredentials)
	}

	// 7. Kubernetes without a provider_config is the ambient workflow, and it
	//    has two homes. On a laptop it is the operator's own kubeconfig, the
	//    way kubectl reads it: the Terraform kubernetes and helm providers do
	//    not read KUBECONFIG on their own (they read KUBE_CONFIG_PATH), so
	//    left alone they would fall back to in-cluster auth and fail with a
	//    connection refused to localhost; the ambient branch hands them the
	//    host kubeconfig and explains a missing one in three parts before any
	//    engine starts. Inside a pod with no kubeconfig named -- the in-cluster
	//    runner deploying through a runner-mode connection -- the pod's own
	//    ServiceAccount IS the credential and the cluster it lives in IS the
	//    target, so nothing is exported and every engine takes its in-cluster
	//    path. A kubeconfig named in the environment wins even inside a pod.
	if provider == cloudresourcekind.CloudResourceProvider_kubernetes && !hasProviderConfig {
		if os.Getenv("KUBECONFIG") == "" && runningInCluster() {
			return map[string]string{}, nil
		}
		return loadHostKubernetesEnvVars(opts.KubeContext)
	}

	if !hasProviderConfig {
		// No provider_config in stack input - return empty map
		return map[string]string{}, nil
	}

	// 8. Load provider_config and convert to env vars based on provider
	envVars, err := loadProviderEnvVars(providerConfigYaml, provider, opts)
	if err != nil {
		return nil, err
	}
	if provider == cloudresourcekind.CloudResourceProvider_kubernetes && opts.KubeContext != "" {
		envVars[kubeconfig.KubeContextEnvVar] = opts.KubeContext
	}
	return envVars, nil
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

// extractTargetSpecRegion reads target.spec.region from the parsed stack input. AWS region is a
// resource property, so it is sourced from the target manifest (the connection region carried in
// provider_config is only a fallback). Returns "" when absent.
func extractTargetSpecRegion(stackInputMap map[string]interface{}) string {
	target, ok := stackInputMap["target"].(map[string]interface{})
	if !ok {
		return ""
	}
	spec, ok := target["spec"].(map[string]interface{})
	if !ok {
		return ""
	}
	region, _ := spec["region"].(string)
	return region
}

// putIfSet adds key to env only when value is non-empty. An empty variable is not an unset one:
// the runner appends these after its own environment and the last duplicate wins, so an empty
// value would erase what the runner's machine holds (its ambient identity, in runner mode).
func putIfSet(env map[string]string, key, value string) {
	if value != "" {
		env[key] = value
	}
}

// loadProviderEnvVars loads the provider config YAML and returns environment variables based on the provider type.
// AWS is intentionally absent here -- it is handled in GetEnvVarsWithOptions (region injection + STS exchange).
func loadProviderEnvVars(providerConfigYaml []byte, provider cloudresourcekind.CloudResourceProvider, opts Options) (map[string]string, error) {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_openfga:
		return loadOpenFgaEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_gcp:
		return loadGcpEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_azure:
		return loadAzureEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_auth0:
		return loadAuth0EnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_kubernetes:
		return loadKubernetesEnvVars(providerConfigYaml, opts.FileCacheLoc)
	case cloudresourcekind.CloudResourceProvider_cloudflare:
		return loadCloudflareEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_digital_ocean:
		return loadDigitalOceanEnvVars(providerConfigYaml)
	default:
		// Unknown or unspecified provider - no env vars needed
		return map[string]string{}, nil
	}
}

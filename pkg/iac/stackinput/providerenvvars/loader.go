package providerenvvars

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/provider/aws/awswebidentity"
	"github.com/plantonhq/planton/pkg/iac/provider/gcp/gcpwebidentity"
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

	// Engine names the IaC engine these variables are for. It decides only what a keyless
	// (web-identity) configuration becomes, and it has no default: see Engine.
	Engine Engine

	// KubeContext selects the kubeconfig context for a Kubernetes deploy (the
	// --kube-context flag, else the manifest's context label). It is exported
	// as KUBE_CTX, the name the Terraform kubernetes and helm providers and
	// the Pulumi provider getter all read. Empty means the kubeconfig's
	// current context.
	KubeContext string
}

// Engine is how the IaC engine consuming these variables authenticates its providers.
//
// It matters for keyless (web-identity) configurations alone, and there it is the whole
// difference. An engine that reads credentials only from its environment must be handed the
// keyless credential in that form, while an engine that builds its providers in the program owns
// the exchange itself and must be handed nothing, or the job pays for an exchange whose result is
// shadowed. The zero value is deliberately neither: a keyless configuration reaching a caller that
// never said which engine it is gets a refusal, never a deploy left to whatever ambient identity
// the machine holds, which is exactly the harm keyless connections exist to prevent. Stored-key
// and runner-mode configurations read the same on every engine and ignore it.
type Engine int

const (
	engineUnset Engine = iota
	// EngineReadsEnvironment is OpenTofu and Terraform: the catalog modules' provider blocks are
	// empty, so a keyless configuration becomes the credential the provider reads -- a short-lived
	// credential exchanged here (AWS, Google Cloud) or the federated token the provider exchanges
	// itself (Azure).
	EngineReadsEnvironment
	// EngineBuildsProviders is Pulumi: the program's provider builders read provider_config and
	// perform every keyless exchange, so a keyless configuration adds no credential variable here.
	EngineBuildsProviders
)

// keylessResolvers are the exchanges an environment-reading engine needs, injected so a test never
// reaches a cloud.
type keylessResolvers struct {
	aws awswebidentity.CredentialResolver
	gcp gcpwebidentity.TokenResolver
}

var cloudResolvers = keylessResolvers{aws: awswebidentity.ResolveCredentials, gcp: gcpwebidentity.ResolveAccessToken}

// keylessExchangeTimeout bounds each keyless exchange done for an environment-reading engine. The
// exchange runs once, before any engine command; the ceiling protects the stack job from a hung
// token endpoint. A fresh context.Background() is used rather than a caller's so the public
// providerenvvars and tofumodule signatures stay stable; the minted token's own short lifetime
// bounds the credential independently.
const keylessExchangeTimeout = 60 * time.Second

// errEngineUnset refuses a keyless configuration whose caller did not name its engine.
func errEngineUnset(provider string) error {
	return errors.Errorf("a keyless %s configuration reached the provider environment without naming its engine: "+
		"set Options.Engine (EngineReadsEnvironment for OpenTofu and Terraform, EngineBuildsProviders for Pulumi); "+
		"a keyless credential is never left to the machine's ambient identity", provider)
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
	//    provider_config (the standalone-CLI ambient case), and a keyless connection's STS
	//    exchange runs in that region. Every other provider keeps the provider_config -> env-var
	//    mapping of loadProviderEnvVars.
	if provider == cloudresourcekind.CloudResourceProvider_aws {
		resourceRegion := extractTargetSpecRegion(stackInputMap)
		return loadAwsEnvVars(providerConfigYaml, hasProviderConfig, resourceRegion, opts, cloudResolvers.aws)
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
	envVars, err := loadProviderEnvVars(providerConfigYaml, provider, opts, cloudResolvers)
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
func loadProviderEnvVars(providerConfigYaml []byte, provider cloudresourcekind.CloudResourceProvider, opts Options,
	resolvers keylessResolvers) (map[string]string, error) {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_openfga:
		return loadOpenFgaEnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_gcp:
		return loadGcpEnvVars(providerConfigYaml, opts, resolvers.gcp)
	case cloudresourcekind.CloudResourceProvider_azure:
		return loadAzureEnvVars(providerConfigYaml, opts)
	case cloudresourcekind.CloudResourceProvider_auth0:
		return loadAuth0EnvVars(providerConfigYaml)
	case cloudresourcekind.CloudResourceProvider_stripe:
		return loadStripeEnvVars(providerConfigYaml)
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

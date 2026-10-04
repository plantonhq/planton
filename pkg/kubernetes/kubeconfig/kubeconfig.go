// Package kubeconfig renders the kubeconfig every Planton deploy engine uses to reach
// a Kubernetes cluster. It is the ONE builder shared by the tofu env-var path and the
// pulumi provider getter, so the two engines cannot drift on how a provider's
// credentials are wired.
//
// Providers whose tokens are short-lived (EKS, GKE, AKS) get an exec entry pointing
// back at the engine-spawning Planton binary (the ExecCredential protocol -- see
// pkg/kubernetes/execcredential); DOKS and self-managed clusters ship a long-lived
// kubeconfig and pass through unchanged. Kubeconfigs are rendered from typed structs,
// never string templates, so field values can never inject YAML structure.
package kubeconfig

import (
	"strings"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/kubernetes/execcredential"
	"sigs.k8s.io/yaml"
)

// entryName names the single cluster/user/context entry in every rendered kubeconfig.
// One connection = one kubeconfig; nothing ever merges these files.
const entryName = "planton"

// Build renders the kubeconfig for the given provider config. credentialCommand is
// the absolute path of the binary serving the ExecCredential subcommand -- the host
// itself when it renders in-process (os.Executable()), or the value advertised via
// execcredential.CommandPathEnvVar when rendering happens inside a module process.
// Arms that need no exec credential (DOKS) ignore it.
func Build(config *kubernetesprovider.KubernetesProviderConfig, credentialCommand string) (string, error) {
	if config == nil {
		return "", errors.New("kubernetes provider config is nil")
	}

	switch config.Provider {
	case kubernetesprovider.KubernetesProvider_aws_eks:
		return buildAwsEks(config.AwsEks, credentialCommand)
	case kubernetesprovider.KubernetesProvider_gcp_gke:
		return buildGcpGke(config.GcpGke, credentialCommand)
	case kubernetesprovider.KubernetesProvider_azure_aks:
		return buildAzureAks(config.AzureAks, credentialCommand)
	case kubernetesprovider.KubernetesProvider_digital_ocean_doks:
		return buildDigitalOceanDoks(config.DigitalOceanDoks)
	case kubernetesprovider.KubernetesProvider_self_managed:
		return buildSelfManaged(config.SelfManaged)
	default:
		return "", errors.Errorf("unsupported kubernetes provider: %v", config.Provider)
	}
}

func buildAwsEks(c *kubernetesprovider.KubernetesProviderConfigAwsEks, credentialCommand string) (string, error) {
	if c == nil {
		return "", errors.New("aws_eks provider config is nil")
	}

	env := []execEnvVar{
		{Name: execcredential.ProviderEnvVar, Value: execcredential.ProviderAwsEks},
		{Name: execcredential.EksClusterNameEnvVar, Value: c.ClusterName},
		{Name: execcredential.EksRegionEnvVar, Value: c.Region},
	}
	// Static keys ride the SDK's standard names (see execcredential); omitted entirely
	// in ambient-chain mode so the helper's own environment is never poisoned with
	// empty values.
	if c.AccessKeyId != "" {
		env = append(env,
			execEnvVar{Name: execcredential.AwsAccessKeyIDEnvVar, Value: c.AccessKeyId},
			execEnvVar{Name: execcredential.AwsSecretAccessKeyEnvVar, Value: c.SecretAccessKey},
		)
		if c.SessionToken != "" {
			env = append(env, execEnvVar{Name: execcredential.AwsSessionTokenEnvVar, Value: c.SessionToken})
		}
	}

	return renderExecKubeconfig(c.ClusterEndpoint, c.ClusterCaData, credentialCommand, env)
}

func buildGcpGke(c *kubernetesprovider.KubernetesProviderConfigGcpGke, credentialCommand string) (string, error) {
	if c == nil {
		return "", errors.New("gcp_gke provider config is nil")
	}

	env := []execEnvVar{
		{Name: execcredential.ProviderEnvVar, Value: execcredential.ProviderGcpGke},
	}
	// The service-account key is omitted entirely in ambient mode so the helper's own
	// environment is never poisoned with empty values (same discipline as the EKS/AKS
	// arms); the minter then falls back to GOOGLE_OAUTH_ACCESS_TOKEN or ADC.
	if c.ServiceAccountKey != "" {
		env = append(env, execEnvVar{Name: execcredential.GkeServiceAccountKeyEnvVar, Value: c.ServiceAccountKey})
	}

	return renderExecKubeconfig(c.ClusterEndpoint, c.ClusterCaData, credentialCommand, env)
}

func buildAzureAks(c *kubernetesprovider.KubernetesProviderConfigAzureAks, credentialCommand string) (string, error) {
	if c == nil {
		return "", errors.New("azure_aks provider config is nil")
	}

	env := []execEnvVar{
		{Name: execcredential.ProviderEnvVar, Value: execcredential.ProviderAzureAks},
	}
	// The tenant is an identity coordinate, not credential material -- it travels
	// whenever set, independent of the credential arm: in ambient mode it steers
	// the minter's chain to the right tenant on multi-tenant hosts.
	if c.TenantId != "" {
		env = append(env, execEnvVar{Name: execcredential.AksTenantIdEnvVar, Value: c.TenantId})
	}
	// The service-principal credential rides dedicated env entries; omitted entirely
	// in ambient-chain mode so the helper's own environment is never poisoned with
	// empty values (same discipline as the EKS arm).
	if c.ClientSecret != "" {
		env = append(env,
			execEnvVar{Name: execcredential.AksClientIdEnvVar, Value: c.ClientId},
			execEnvVar{Name: execcredential.AksClientSecretEnvVar, Value: c.ClientSecret},
		)
	}

	return renderExecKubeconfig(c.ClusterEndpoint, c.ClusterCaData, credentialCommand, env)
}

func buildDigitalOceanDoks(c *kubernetesprovider.KubernetesProviderConfigDigitalOceanDoks) (string, error) {
	if c == nil {
		return "", errors.New("digital_ocean_doks provider config is nil")
	}
	// DOKS hands out a complete long-lived kubeconfig; there is no token to refresh
	// and therefore nothing for this builder to construct.
	return c.KubeConfig, nil
}

func buildSelfManaged(c *kubernetesprovider.KubernetesProviderConfigSelfManaged) (string, error) {
	if c == nil {
		return "", errors.New("self_managed provider config is nil")
	}
	// Self-managed clusters (kind, on-prem, vSphere, and any kubeconfig-only
	// cloud) authenticate through the kubeconfig itself -- pass-through, exactly
	// like DOKS.
	return c.KubeConfig, nil
}

// renderExecKubeconfig assembles the one-cluster exec-credential kubeconfig every
// short-lived-token provider shares; only the exec env entries differ per provider.
func renderExecKubeconfig(endpoint, caData, credentialCommand string, env []execEnvVar) (string, error) {
	if credentialCommand == "" {
		return "", errors.Errorf("credential command path is required to render an exec-credential "+
			"kubeconfig; engine-spawning hosts advertise it via %s", execcredential.CommandPathEnvVar)
	}

	rendered, err := yaml.Marshal(kubeconfigFile{
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: entryName,
		Clusters: []namedCluster{{
			Name: entryName,
			Cluster: cluster{
				Server:                   normalizeServerURL(endpoint),
				CertificateAuthorityData: caData,
			},
		}},
		Contexts: []namedContext{{
			Name:    entryName,
			Context: contextEntry{Cluster: entryName, User: entryName},
		}},
		Users: []namedUser{{
			Name: entryName,
			User: user{
				Exec: execConfig{
					APIVersion: "client.authentication.k8s.io/v1",
					Command:    credentialCommand,
					// The subcommand name is the ONLY argument; credentials travel in
					// env entries, never argv (argv is world-readable via ps).
					Args:            []string{execcredential.Command.Use},
					Env:             env,
					InteractiveMode: "Never",
				},
			},
		}},
	})
	if err != nil {
		return "", errors.Wrap(err, "marshaling kubeconfig")
	}
	return string(rendered), nil
}

// normalizeServerURL tolerates both endpoint shapes seen across providers: EKS
// outputs export a full https:// URL while GKE exports a bare endpoint IP.
func normalizeServerURL(endpoint string) string {
	if strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	return "https://" + endpoint
}

// Minimal kubeconfig v1 document model -- just the fields Planton renders. Field tags
// are json because sigs.k8s.io/yaml round-trips through JSON; client-go's own types
// are not used to keep client-go out of the dependency tree.
type kubeconfigFile struct {
	APIVersion     string         `json:"apiVersion"`
	Kind           string         `json:"kind"`
	CurrentContext string         `json:"current-context"`
	Clusters       []namedCluster `json:"clusters"`
	Contexts       []namedContext `json:"contexts"`
	Users          []namedUser    `json:"users"`
}

type namedCluster struct {
	Name    string  `json:"name"`
	Cluster cluster `json:"cluster"`
}

type cluster struct {
	Server                   string `json:"server"`
	CertificateAuthorityData string `json:"certificate-authority-data"`
}

type namedContext struct {
	Name    string       `json:"name"`
	Context contextEntry `json:"context"`
}

type contextEntry struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

type namedUser struct {
	Name string `json:"name"`
	User user   `json:"user"`
}

type user struct {
	Exec execConfig `json:"exec"`
}

type execConfig struct {
	APIVersion      string       `json:"apiVersion"`
	Command         string       `json:"command"`
	Args            []string     `json:"args,omitempty"`
	Env             []execEnvVar `json:"env,omitempty"`
	InteractiveMode string       `json:"interactiveMode"`
}

type execEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

package kubeconfig

import (
	"testing"

	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/kubernetes/execcredential"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

const (
	testCredentialCommand = "/usr/local/bin/planton"
	testSecretAccessKey   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	testSessionToken      = "test-session-token"
	testServiceAccountKey = `{"type":"service_account","private_key":"-----BEGIN RSA PRIVATE KEY-----"}`
)

func eksConfig() *kubernetesprovider.KubernetesProviderConfig {
	return &kubernetesprovider.KubernetesProviderConfig{
		Provider: kubernetesprovider.KubernetesProvider_aws_eks,
		AwsEks: &kubernetesprovider.KubernetesProviderConfigAwsEks{
			ClusterName:     "demo-cluster",
			ClusterEndpoint: "https://ABC123.gr7.us-east-1.eks.amazonaws.com",
			ClusterCaData:   "dGVzdC1jYS1kYXRh",
			Region:          "us-east-1",
			AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey: testSecretAccessKey,
			SessionToken:    testSessionToken,
		},
	}
}

func parseKubeconfig(t *testing.T, rendered string) kubeconfigFile {
	t.Helper()
	var doc kubeconfigFile
	require.NoError(t, yaml.Unmarshal([]byte(rendered), &doc))
	return doc
}

func execEnvMap(t *testing.T, doc kubeconfigFile) map[string]string {
	t.Helper()
	require.Len(t, doc.Users, 1)
	env := map[string]string{}
	for _, e := range doc.Users[0].User.Exec.Env {
		env[e.Name] = e.Value
	}
	return env
}

// TestBuild_AwsEksExecShape guards the full EKS kubeconfig contract: cluster identity,
// v1 exec entry pointing at the credential command, and the env contract the
// ExecCredential command reads on the other end.
func TestBuild_AwsEksExecShape(t *testing.T) {
	rendered, err := Build(eksConfig(), testCredentialCommand)
	require.NoError(t, err)

	doc := parseKubeconfig(t, rendered)
	require.Len(t, doc.Clusters, 1)
	assert.Equal(t, "https://ABC123.gr7.us-east-1.eks.amazonaws.com", doc.Clusters[0].Cluster.Server,
		"an endpoint already carrying a scheme must pass through unchanged")
	assert.Equal(t, "dGVzdC1jYS1kYXRh", doc.Clusters[0].Cluster.CertificateAuthorityData)
	assert.Equal(t, doc.CurrentContext, doc.Contexts[0].Name)

	exec := doc.Users[0].User.Exec
	assert.Equal(t, "client.authentication.k8s.io/v1", exec.APIVersion)
	assert.Equal(t, testCredentialCommand, exec.Command)
	assert.Equal(t, []string{execcredential.Command.Use}, exec.Args)
	assert.Equal(t, "Never", exec.InteractiveMode)

	env := execEnvMap(t, doc)
	assert.Equal(t, execcredential.ProviderAwsEks, env[execcredential.ProviderEnvVar])
	assert.Equal(t, "demo-cluster", env[execcredential.EksClusterNameEnvVar])
	assert.Equal(t, "us-east-1", env[execcredential.EksRegionEnvVar])
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", env[execcredential.AwsAccessKeyIDEnvVar])
	assert.Equal(t, testSecretAccessKey, env[execcredential.AwsSecretAccessKeyEnvVar])
	assert.Equal(t, testSessionToken, env[execcredential.AwsSessionTokenEnvVar])
}

// TestBuild_NoSecretsInArgv is the security invariant of the whole seam: credential
// material travels ONLY in exec env entries, never in the command arguments (argv is
// world-readable via ps).
func TestBuild_NoSecretsInArgv(t *testing.T) {
	rendered, err := Build(eksConfig(), testCredentialCommand)
	require.NoError(t, err)
	doc := parseKubeconfig(t, rendered)

	for _, arg := range doc.Users[0].User.Exec.Args {
		assert.NotContains(t, arg, testSecretAccessKey)
		assert.NotContains(t, arg, testSessionToken)
		assert.NotContains(t, arg, "AKIA")
	}

	gkeRendered, err := Build(&kubernetesprovider.KubernetesProviderConfig{
		Provider: kubernetesprovider.KubernetesProvider_gcp_gke,
		GcpGke: &kubernetesprovider.KubernetesProviderConfigGcpGke{
			ClusterEndpoint:   "34.100.155.147",
			ClusterCaData:     "dGVzdC1jYS1kYXRh",
			ServiceAccountKey: testServiceAccountKey,
		},
	}, testCredentialCommand)
	require.NoError(t, err)
	gkeDoc := parseKubeconfig(t, gkeRendered)

	for _, arg := range gkeDoc.Users[0].User.Exec.Args {
		assert.NotContains(t, arg, "service_account")
	}
}

// TestBuild_AwsEksAmbientMode: with no static keys on the connection, the kubeconfig
// must not emit ANY AWS credential env entries -- empty values would poison the
// helper's ambient credential chain.
func TestBuild_AwsEksAmbientMode(t *testing.T) {
	config := eksConfig()
	config.AwsEks.AccessKeyId = ""
	config.AwsEks.SecretAccessKey = ""
	config.AwsEks.SessionToken = ""

	rendered, err := Build(config, testCredentialCommand)
	require.NoError(t, err)

	env := execEnvMap(t, parseKubeconfig(t, rendered))
	assert.NotContains(t, env, execcredential.AwsAccessKeyIDEnvVar)
	assert.NotContains(t, env, execcredential.AwsSecretAccessKeyEnvVar)
	assert.NotContains(t, env, execcredential.AwsSessionTokenEnvVar)
}

// TestBuild_GcpGkeExecShape: GKE rides the same exec seam with its own env contract,
// and a bare endpoint IP (the GKE output shape) gains the https:// scheme.
func TestBuild_GcpGkeExecShape(t *testing.T) {
	rendered, err := Build(&kubernetesprovider.KubernetesProviderConfig{
		Provider: kubernetesprovider.KubernetesProvider_gcp_gke,
		GcpGke: &kubernetesprovider.KubernetesProviderConfigGcpGke{
			ClusterEndpoint:   "34.100.155.147",
			ClusterCaData:     "dGVzdC1jYS1kYXRh",
			ServiceAccountKey: testServiceAccountKey,
		},
	}, testCredentialCommand)
	require.NoError(t, err)

	doc := parseKubeconfig(t, rendered)
	assert.Equal(t, "https://34.100.155.147", doc.Clusters[0].Cluster.Server)

	env := execEnvMap(t, doc)
	assert.Equal(t, execcredential.ProviderGcpGke, env[execcredential.ProviderEnvVar])
	assert.Equal(t, testServiceAccountKey, env[execcredential.GkeServiceAccountKeyEnvVar])
}

// TestBuild_GcpGkeAmbientMode: with no service-account key on the connection, the
// kubeconfig must not emit the key env entry -- an empty value would poison the
// helper's ambient credential chain (GOOGLE_OAUTH_ACCESS_TOKEN / ADC).
func TestBuild_GcpGkeAmbientMode(t *testing.T) {
	rendered, err := Build(&kubernetesprovider.KubernetesProviderConfig{
		Provider: kubernetesprovider.KubernetesProvider_gcp_gke,
		GcpGke: &kubernetesprovider.KubernetesProviderConfigGcpGke{
			ClusterEndpoint: "34.100.155.147",
			ClusterCaData:   "dGVzdC1jYS1kYXRh",
		},
	}, testCredentialCommand)
	require.NoError(t, err)

	env := execEnvMap(t, parseKubeconfig(t, rendered))
	assert.Equal(t, execcredential.ProviderGcpGke, env[execcredential.ProviderEnvVar])
	assert.NotContains(t, env, execcredential.GkeServiceAccountKeyEnvVar)
}

// TestBuild_DigitalOceanDoksPassthrough: DOKS hands out a complete long-lived
// kubeconfig; the builder must return it byte-for-byte.
func TestBuild_DigitalOceanDoksPassthrough(t *testing.T) {
	rawKubeconfig := "apiVersion: v1\nkind: Config\n"

	rendered, err := Build(&kubernetesprovider.KubernetesProviderConfig{
		Provider:         kubernetesprovider.KubernetesProvider_digital_ocean_doks,
		DigitalOceanDoks: &kubernetesprovider.KubernetesProviderConfigDigitalOceanDoks{KubeConfig: rawKubeconfig},
	}, "")
	require.NoError(t, err)
	assert.Equal(t, rawKubeconfig, rendered)
}

func aksConfig() *kubernetesprovider.KubernetesProviderConfig {
	return &kubernetesprovider.KubernetesProviderConfig{
		Provider: kubernetesprovider.KubernetesProvider_azure_aks,
		AzureAks: &kubernetesprovider.KubernetesProviderConfigAzureAks{
			ClusterEndpoint: "https://demo-aks.hcp.eastus.azmk8s.io",
			ClusterCaData:   "dGVzdC1jYS1kYXRh",
			TenantId:        "11111111-2222-3333-4444-555555555555",
			ClientId:        "test-client-id",
			ClientSecret:    "test-client-secret",
		},
	}
}

// TestBuild_AzureAksExecShape: AKS rides the same exec seam as EKS/GKE with the
// Entra service-principal env contract the ExecCredential command reads.
func TestBuild_AzureAksExecShape(t *testing.T) {
	rendered, err := Build(aksConfig(), testCredentialCommand)
	require.NoError(t, err)

	doc := parseKubeconfig(t, rendered)
	assert.Equal(t, "https://demo-aks.hcp.eastus.azmk8s.io", doc.Clusters[0].Cluster.Server)
	assert.Equal(t, "dGVzdC1jYS1kYXRh", doc.Clusters[0].Cluster.CertificateAuthorityData)
	assert.Equal(t, "client.authentication.k8s.io/v1", doc.Users[0].User.Exec.APIVersion)

	env := execEnvMap(t, doc)
	assert.Equal(t, execcredential.ProviderAzureAks, env[execcredential.ProviderEnvVar])
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", env[execcredential.AksTenantIdEnvVar])
	assert.Equal(t, "test-client-id", env[execcredential.AksClientIdEnvVar])
	assert.Equal(t, "test-client-secret", env[execcredential.AksClientSecretEnvVar])

	for _, arg := range doc.Users[0].User.Exec.Args {
		assert.NotContains(t, arg, "test-client-secret",
			"credential material travels ONLY in exec env entries, never argv")
	}
}

// TestBuild_AzureAksAmbientMode: with no client secret on the connection, the
// kubeconfig must not emit ANY Entra credential env entries -- empty values would
// poison the helper's ambient credential chain.
func TestBuild_AzureAksAmbientMode(t *testing.T) {
	config := aksConfig()
	config.AzureAks.TenantId = ""
	config.AzureAks.ClientId = ""
	config.AzureAks.ClientSecret = ""

	rendered, err := Build(config, testCredentialCommand)
	require.NoError(t, err)

	env := execEnvMap(t, parseKubeconfig(t, rendered))
	assert.NotContains(t, env, execcredential.AksTenantIdEnvVar)
	assert.NotContains(t, env, execcredential.AksClientIdEnvVar)
	assert.NotContains(t, env, execcredential.AksClientSecretEnvVar)
}

// TestBuild_AzureAksAmbientModeWithTenant: the tenant is an identity coordinate,
// not credential material -- it must travel in ambient mode (steering the minter's
// chain on multi-tenant hosts) while the credential entries stay omitted.
func TestBuild_AzureAksAmbientModeWithTenant(t *testing.T) {
	config := aksConfig()
	config.AzureAks.ClientId = ""
	config.AzureAks.ClientSecret = ""

	rendered, err := Build(config, testCredentialCommand)
	require.NoError(t, err)

	env := execEnvMap(t, parseKubeconfig(t, rendered))
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", env[execcredential.AksTenantIdEnvVar])
	assert.NotContains(t, env, execcredential.AksClientIdEnvVar)
	assert.NotContains(t, env, execcredential.AksClientSecretEnvVar)
}

// TestBuild_SelfManagedPassthrough: self-managed clusters (kind, on-prem, vSphere,
// kubeconfig-only clouds) authenticate through the kubeconfig itself; the builder
// must return it byte-for-byte.
func TestBuild_SelfManagedPassthrough(t *testing.T) {
	rawKubeconfig := "apiVersion: v1\nkind: Config\n"

	rendered, err := Build(&kubernetesprovider.KubernetesProviderConfig{
		Provider:    kubernetesprovider.KubernetesProvider_self_managed,
		SelfManaged: &kubernetesprovider.KubernetesProviderConfigSelfManaged{KubeConfig: rawKubeconfig},
	}, "")
	require.NoError(t, err)
	assert.Equal(t, rawKubeconfig, rendered)
}

// TestBuild_ExecArmsRequireCredentialCommand: exec-credential providers must reject an
// empty command path with an error naming the host contract, never render a kubeconfig
// whose exec entry points nowhere.
func TestBuild_ExecArmsRequireCredentialCommand(t *testing.T) {
	_, err := Build(eksConfig(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), execcredential.CommandPathEnvVar)
}

func TestBuild_NilConfigsRejected(t *testing.T) {
	_, err := Build(nil, testCredentialCommand)
	assert.Error(t, err)

	_, err = Build(&kubernetesprovider.KubernetesProviderConfig{
		Provider: kubernetesprovider.KubernetesProvider_aws_eks,
	}, testCredentialCommand)
	assert.Error(t, err, "a provider selector without its sub-message must be rejected")
}

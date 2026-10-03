package providerenvvars

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/plantonhq/planton/pkg/failure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadKubernetesEnvVars_ExportsBothKubeconfigEnvVars guards the Tofu/Pulumi parity
// fix: the generated kubeconfig must be advertised under BOTH KUBECONFIG (Pulumi) and
// KUBE_CONFIG_PATH (Terraform/OpenTofu hashicorp/kubernetes + helm). Dropping
// KUBE_CONFIG_PATH makes the tofu provider silently fall back to in-cluster auth.
func TestLoadKubernetesEnvVars_ExportsBothKubeconfigEnvVars(t *testing.T) {
	// Minimal valid GCP GKE provider config. protojson accepts the proto field names
	// (snake_case) and the enum value name as a string.
	providerConfigYaml := []byte(`
provider: gcp_gke
gcp_gke:
  cluster_endpoint: "34.100.155.147"
  cluster_ca_data: "dGVzdC1jYS1kYXRh"
  service_account_key: "{\"type\":\"service_account\"}"
`)

	cacheDir := t.TempDir()

	envVars, err := loadKubernetesEnvVars(providerConfigYaml, cacheDir)
	require.NoError(t, err)

	kubeconfig, ok := envVars["KUBECONFIG"]
	assert.True(t, ok, "KUBECONFIG must be set for the Pulumi kubernetes provider")

	kubeConfigPath, ok := envVars["KUBE_CONFIG_PATH"]
	assert.True(t, ok, "KUBE_CONFIG_PATH must be set for the Terraform/OpenTofu kubernetes provider")

	assert.Equal(t, kubeconfig, kubeConfigPath,
		"both env vars must point at the same generated kubeconfig file")

	// The kubeconfig file the env vars reference must actually exist on disk.
	_, statErr := os.Stat(kubeConfigPath)
	assert.NoError(t, statErr, "the kubeconfig file referenced by the env vars should exist")
}

// TestLoadKubernetesEnvVars_AwsEks proves the EKS arm end to end on this path: the
// rendered kubeconfig lands on disk (0600 -- it can carry credentials in exec env
// entries) with an exec entry pointing back at this very binary.
func TestLoadKubernetesEnvVars_AwsEks(t *testing.T) {
	providerConfigYaml := []byte(`
provider: aws_eks
aws_eks:
  cluster_name: demo-cluster
  cluster_endpoint: "https://ABC123.gr7.us-east-1.eks.amazonaws.com"
  cluster_ca_data: "dGVzdC1jYS1kYXRh"
  region: us-east-1
`)

	envVars, err := loadKubernetesEnvVars(providerConfigYaml, t.TempDir())
	require.NoError(t, err)

	kubeConfigPath := envVars["KUBECONFIG"]
	require.NotEmpty(t, kubeConfigPath)
	assert.Equal(t, kubeConfigPath, envVars["KUBE_CONFIG_PATH"])

	info, err := os.Stat(kubeConfigPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	content, err := os.ReadFile(kubeConfigPath)
	require.NoError(t, err)

	executable, err := os.Executable()
	require.NoError(t, err)
	assert.Contains(t, string(content), executable,
		"the exec credential command must be the engine-spawning binary itself")
	assert.Contains(t, string(content), "client.authentication.k8s.io/v1")
}

// TestLoadKubernetesEnvVars_DigitalOceanDoks: the DOKS kubeconfig passes through to
// disk unchanged and is advertised under both engine env vars.
func TestLoadKubernetesEnvVars_DigitalOceanDoks(t *testing.T) {
	providerConfigYaml := []byte(`
provider: digital_ocean_doks
digital_ocean_doks:
  kube_config: |
    apiVersion: v1
    kind: Config
`)

	envVars, err := loadKubernetesEnvVars(providerConfigYaml, t.TempDir())
	require.NoError(t, err)

	content, err := os.ReadFile(envVars["KUBE_CONFIG_PATH"])
	require.NoError(t, err)
	assert.Contains(t, string(content), "kind: Config")
}

// The local workflow: a Kubernetes kind with NO provider_config uses the
// operator's own kubeconfig, exported under every name an engine reads. The
// Terraform kubernetes and helm providers never read KUBECONFIG themselves, so
// KUBE_CONFIG_PATH is the load-bearing key; without it the providers fall back
// to in-cluster auth and fail with a connection refused to localhost.
func TestGetEnvVars_KubernetesWithoutProviderConfig_UsesHostKubeconfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	one := filepath.Join(home, "one.yaml")
	two := filepath.Join(home, "two.yaml")

	iacInputYaml := `
target:
  apiVersion: kubernetes.planton.dev/v1alpha1
  kind: KubernetesNamespace
  metadata:
    name: demo
  spec:
    name: demo
`

	t.Run("one file: KUBE_CONFIG_PATH", func(t *testing.T) {
		t.Setenv("KUBECONFIG", one)
		envVars, err := GetEnvVarsWithOptions(iacInputYaml, Options{KubeContext: "kind-planton-e2e"})
		require.NoError(t, err)
		assert.Equal(t, one, envVars["KUBECONFIG"])
		assert.Equal(t, one, envVars["KUBE_CONFIG_PATH"])
		assert.NotContains(t, envVars, "KUBE_CONFIG_PATHS")
		assert.Equal(t, "kind-planton-e2e", envVars["KUBE_CTX"])
	})

	t.Run("a list: KUBE_CONFIG_PATHS, and no KUBE_CTX when none was chosen", func(t *testing.T) {
		list := one + string(os.PathListSeparator) + two
		t.Setenv("KUBECONFIG", list)
		envVars, err := GetEnvVarsWithOptions(iacInputYaml, Options{})
		require.NoError(t, err)
		assert.Equal(t, list, envVars["KUBECONFIG"])
		assert.Equal(t, list, envVars["KUBE_CONFIG_PATHS"])
		assert.NotContains(t, envVars, "KUBE_CONFIG_PATH")
		assert.NotContains(t, envVars, "KUBE_CTX")
	})

	t.Run("inside a pod with no kubeconfig named: nothing exported, the engines take their in-cluster path", func(t *testing.T) {
		// The in-cluster runner deploying through a runner-mode connection:
		// the pod's own ServiceAccount is the credential and the cluster it
		// lives in is the target, so a demand for a host kubeconfig here
		// would refuse every such deploy.
		t.Setenv("KUBECONFIG", "")
		previous := runningInCluster
		runningInCluster = func() bool { return true }
		t.Cleanup(func() { runningInCluster = previous })
		envVars, err := GetEnvVarsWithOptions(iacInputYaml, Options{})
		require.NoError(t, err)
		assert.Empty(t, envVars)
	})

	t.Run("inside a pod but a kubeconfig is named: the named files win", func(t *testing.T) {
		t.Setenv("KUBECONFIG", one)
		previous := runningInCluster
		runningInCluster = func() bool { return true }
		t.Cleanup(func() { runningInCluster = previous })
		envVars, err := GetEnvVarsWithOptions(iacInputYaml, Options{})
		require.NoError(t, err)
		assert.Equal(t, one, envVars["KUBE_CONFIG_PATH"])
	})

	t.Run("no kubeconfig anywhere and not a pod: the three-part refusal, before any engine runs", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "")
		previous := runningInCluster
		runningInCluster = func() bool { return false }
		t.Cleanup(func() { runningInCluster = previous })
		_, err := GetEnvVarsWithOptions(iacInputYaml, Options{})
		require.Error(t, err)
		var f *failure.Failure
		require.True(t, errors.As(err, &f), "want a *failure.Failure, got %T: %v", err, err)
		assert.Contains(t, f.Observed, "KUBECONFIG")
		assert.Contains(t, f.NextStep, "--kube-context")
	})
}

// A connection's rendered kubeconfig still honours an explicit context choice.
func TestGetEnvVars_KubernetesWithProviderConfig_ExportsKubeContext(t *testing.T) {
	iacInputYaml := `
target:
  apiVersion: kubernetes.planton.dev/v1alpha1
  kind: KubernetesNamespace
  metadata:
    name: demo
  spec:
    name: demo
provider_config:
  provider: gcp_gke
  gcp_gke:
    cluster_endpoint: "34.100.155.147"
    cluster_ca_data: "dGVzdC1jYS1kYXRh"
    service_account_key: "{\"type\":\"service_account\"}"
`
	envVars, err := GetEnvVarsWithOptions(iacInputYaml, Options{FileCacheLoc: t.TempDir(), KubeContext: "chosen"})
	require.NoError(t, err)
	assert.NotEmpty(t, envVars["KUBE_CONFIG_PATH"])
	assert.Equal(t, "chosen", envVars["KUBE_CTX"])
}

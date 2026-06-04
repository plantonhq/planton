package providerenvvars

import (
	"os"
	"testing"

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

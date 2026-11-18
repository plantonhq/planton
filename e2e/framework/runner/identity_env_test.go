package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// A declared lane's Pulumi environment carries the lane's own credential and
// nothing else, so a default provider in the program cannot fall back to the
// harness's credential; a lane that declares no identity adds nothing.
func TestLaneEngineEnvCarriesOnlyTheLaneCredential(t *testing.T) {
	if os.Getenv("TEST_WORKSPACE") != "" {
		t.Skip("repo-reading test; skipped under Bazel")
	}
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(root, "catalog", "gcp", "gcptagkey", "e2e", "scenarios", "minimal.yaml")
	t.Setenv("PLANTON_E2E_GCP_PROJECT_ID", "planton-e2e-unit") // the harness exports it at Setup
	expanded, err := ExpandManifestTokens(manifest, "0a1b2c3d-p", "minimal", LaneClock())
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(t.TempDir(), "provider-config.yaml")
	if err := os.WriteFile(configPath, []byte("serviceAccountKey: '{\"type\": \"service_account\", \"client_email\": \"pe2e-x@p.iam.gserviceaccount.com\"}'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	iacInput, err := BuildIacInput(expanded, &iacinputproviderconfig.ProviderConfig{Path: configPath, Provider: catalogkind.CatalogProvider_gcp})
	if err != nil {
		t.Fatal(err)
	}

	tc := &provider.KindTestContext{IacInputFilePath: iacInput}
	env, err := laneEngineEnv(tc)
	if err != nil || env != nil {
		t.Fatalf("a lane with no identity got %v (%v), want nothing", env, err)
	}

	tc.IdentityProviderConfig = configPath
	env, err = laneEngineEnv(tc)
	if err != nil {
		t.Fatal(err)
	}
	if len(env) != 1 || !strings.HasPrefix(env[0], "GOOGLE_CREDENTIALS=") || !strings.Contains(env[0], "pe2e-x@") {
		t.Fatalf("lane environment = %v, want the lane's key as GOOGLE_CREDENTIALS alone", env)
	}
}

// A Kubernetes declared lane's Pulumi environment points KUBECONFIG at the
// lane identity's own kubeconfig, the file its Terraform arm reads too.
func TestLaneEngineEnvForAKubernetesIdentity(t *testing.T) {
	if os.Getenv("TEST_WORKSPACE") != "" {
		t.Skip("repo-reading test; skipped under Bazel")
	}
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(root, "catalog", "kubernetes", "kubernetesnamespace", "e2e", "scenarios", "minimal.yaml")
	expanded, err := ExpandManifestTokens(manifest, "0a1b2c3d-p", "minimal", LaneClock())
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(t.TempDir(), "provider-config.yaml")
	kubeconfig := "apiVersion: v1\nkind: Config\nclusters: [{name: c, cluster: {server: 'https://127.0.0.1:6443'}}]\nusers: [{name: u, user: {token: lane}}]\ncontexts: [{name: x, context: {cluster: c, user: u}}]\ncurrent-context: x\n"
	if err := os.WriteFile(configPath, []byte("provider: self_managed\nself_managed:\n  kube_config: |\n"+indent(kubeconfig, "    ")), 0o600); err != nil {
		t.Fatal(err)
	}
	iacInput, err := BuildIacInput(expanded, &iacinputproviderconfig.ProviderConfig{Path: configPath, Provider: catalogkind.CatalogProvider_kubernetes})
	if err != nil {
		t.Fatal(err)
	}
	env, err := laneEngineEnv(&provider.KindTestContext{IacInputFilePath: iacInput, IdentityProviderConfig: configPath})
	if err != nil {
		t.Fatal(err)
	}
	var kubeconfigPath string
	for _, entry := range env {
		if strings.HasPrefix(entry, "KUBECONFIG=") {
			kubeconfigPath = strings.TrimPrefix(entry, "KUBECONFIG=")
		}
	}
	written, err := os.ReadFile(kubeconfigPath)
	if err != nil || !strings.Contains(string(written), "lane") {
		t.Fatalf("KUBECONFIG = %q holds %q (%v), want the lane's kubeconfig", kubeconfigPath, written, err)
	}
}

func indent(text, prefix string) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n") + "\n"
}

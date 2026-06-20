package runner

import (
	"os"
	"path/filepath"
	"testing"
)

// writeManifest creates a placeholder manifest file (with parent dirs) under a
// fake repo root and returns its absolute path.
func writeManifest(t *testing.T, repoRoot, relPath string) string {
	t.Helper()
	full := filepath.Join(repoRoot, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte("apiVersion: kubernetes.openmcf.org/v1\nkind: Placeholder\n"), 0o600); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
	return full
}

const (
	gwCrdsPrereqRel  = "apis/org/openmcf/provider/kubernetes/kubernetesgatewayapicrds/v1/e2e/prerequisite.yaml"
	gwCrdsMinimalRel = "apis/org/openmcf/provider/kubernetes/kubernetesgatewayapicrds/v1/e2e/scenarios/minimal.yaml"
)

func TestResolveDependencies_RegistryPrerequisite(t *testing.T) {
	repoRoot := t.TempDir()
	want := writeManifest(t, repoRoot, gwCrdsPrereqRel)

	deps, err := ResolveDependencies(repoRoot, "kubernetes", "kuberneteshttproute")
	if err != nil {
		t.Fatalf("ResolveDependencies: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d: %+v", len(deps), deps)
	}
	got := deps[0]
	if got.KindSlug != "kubernetesgatewayapicrds" {
		t.Errorf("kind slug = %q, want kubernetesgatewayapicrds", got.KindSlug)
	}
	if got.ManifestPath != want {
		t.Errorf("manifest path = %q, want %q", got.ManifestPath, want)
	}
}

func TestResolveDependencies_FallbackToMinimalScenario(t *testing.T) {
	repoRoot := t.TempDir()
	want := writeManifest(t, repoRoot, gwCrdsMinimalRel)

	deps, err := ResolveDependencies(repoRoot, "kubernetes", "kuberneteshttproute")
	if err != nil {
		t.Fatalf("ResolveDependencies: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(deps))
	}
	if deps[0].ManifestPath != want {
		t.Errorf("manifest path = %q, want fallback %q", deps[0].ManifestPath, want)
	}
}

func TestResolveDependencies_PrerequisiteYamlWinsOverMinimal(t *testing.T) {
	repoRoot := t.TempDir()
	prereq := writeManifest(t, repoRoot, gwCrdsPrereqRel)
	writeManifest(t, repoRoot, gwCrdsMinimalRel)

	deps, err := ResolveDependencies(repoRoot, "kubernetes", "kuberneteshttproute")
	if err != nil {
		t.Fatalf("ResolveDependencies: %v", err)
	}
	if len(deps) != 1 || deps[0].ManifestPath != prereq {
		t.Fatalf("expected prerequisite.yaml to win, got %+v", deps)
	}
}

func TestResolveDependencies_NoPrerequisites(t *testing.T) {
	repoRoot := t.TempDir()
	deps, err := ResolveDependencies(repoRoot, "kubernetes", "kubernetesnamespace")
	if err != nil {
		t.Fatalf("ResolveDependencies: %v", err)
	}
	if len(deps) != 0 {
		t.Fatalf("expected no dependencies, got %d: %+v", len(deps), deps)
	}
}

func TestResolveDependencies_MissingInstallManifestErrors(t *testing.T) {
	repoRoot := t.TempDir()
	// httproute has a registry prereq but we create no install manifest for it.
	if _, err := ResolveDependencies(repoRoot, "kubernetes", "kuberneteshttproute"); err == nil {
		t.Fatal("expected an error when the prerequisite install manifest is missing, got nil")
	}
}

// TestResolveDependencies_TransitiveDeployOrder guards the deep-composition
// ordering DeployDependencies relies on: AwsNatGateway -> [AwsSubnet, AwsElasticIp]
// with AwsSubnet -> [AwsVpc] must resolve to [AwsVpc, AwsSubnet, AwsElasticIp], so
// the VPC's outputs are accumulated before the Subnet whose vpc_id references them.
func TestResolveDependencies_TransitiveDeployOrder(t *testing.T) {
	repoRoot := t.TempDir()
	writeManifest(t, repoRoot, "apis/org/openmcf/provider/aws/awsvpc/v1/e2e/prerequisite.yaml")
	writeManifest(t, repoRoot, "apis/org/openmcf/provider/aws/awssubnet/v1/e2e/scenarios/minimal.yaml")
	writeManifest(t, repoRoot, "apis/org/openmcf/provider/aws/awselasticip/v1/e2e/prerequisite.yaml")

	deps, err := ResolveDependencies(repoRoot, "aws", "awsnatgateway")
	if err != nil {
		t.Fatalf("ResolveDependencies: %v", err)
	}

	got := make([]string, len(deps))
	for i, d := range deps {
		got[i] = d.KindSlug
	}
	want := []string{"awsvpc", "awssubnet", "awselasticip"}
	if len(got) != len(want) {
		t.Fatalf("dependency count = %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("dependency order = %v, want %v (VPC must precede Subnet)", got, want)
		}
	}
}

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
	if got.Source != SourceRegistry {
		t.Errorf("source = %q, want %q", got.Source, SourceRegistry)
	}
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

func TestResolveDependencies_FixtureOverridesRegistry(t *testing.T) {
	repoRoot := t.TempDir()
	// KubernetesPostgres declares KubernetesZalandoPostgresOperator as a registry
	// prerequisite. A fixture supplying the same kind must override it (and the
	// registry entry must be deduped -- note we deliberately do NOT create an
	// operator install manifest, proving registry resolution was skipped).
	fixture := writeManifest(t, repoRoot, "apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1/e2e/fixtures/01-kuberneteszalandopostgresoperator.yaml")

	deps, err := ResolveDependencies(repoRoot, "kubernetes", "kubernetespostgres")
	if err != nil {
		t.Fatalf("ResolveDependencies: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency (fixture only), got %d: %+v", len(deps), deps)
	}
	got := deps[0]
	if got.Source != SourceFixture {
		t.Errorf("source = %q, want %q", got.Source, SourceFixture)
	}
	if got.KindSlug != "kuberneteszalandopostgresoperator" {
		t.Errorf("kind slug = %q, want kuberneteszalandopostgresoperator", got.KindSlug)
	}
	if got.ManifestPath != fixture {
		t.Errorf("manifest path = %q, want %q", got.ManifestPath, fixture)
	}
}

func TestResolveDependencies_RegistryBeforeFixtureOrdering(t *testing.T) {
	repoRoot := t.TempDir()
	// httproute's registry prereq (gateway api crds) plus an unrelated fixture
	// (kubernetesnamespace). Expect registry first, then fixture.
	writeManifest(t, repoRoot, gwCrdsPrereqRel)
	writeManifest(t, repoRoot, "apis/org/openmcf/provider/kubernetes/kuberneteshttproute/v1/e2e/fixtures/01-kubernetesnamespace.yaml")

	deps, err := ResolveDependencies(repoRoot, "kubernetes", "kuberneteshttproute")
	if err != nil {
		t.Fatalf("ResolveDependencies: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 dependencies, got %d: %+v", len(deps), deps)
	}
	if deps[0].Source != SourceRegistry || deps[0].KindSlug != "kubernetesgatewayapicrds" {
		t.Errorf("first dep = %+v, want registry kubernetesgatewayapicrds", deps[0])
	}
	if deps[1].Source != SourceFixture || deps[1].KindSlug != "kubernetesnamespace" {
		t.Errorf("second dep = %+v, want fixture kubernetesnamespace", deps[1])
	}
}

func TestResolveDependencies_NoPrerequisitesNoFixtures(t *testing.T) {
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

func TestKindSlugFromFixtureFilename(t *testing.T) {
	cases := map[string]string{
		"01-kuberneteszalandopostgresoperator": "kuberneteszalandopostgresoperator",
		"02-kubernetesnamespace":               "kubernetesnamespace",
		"kubernetesredis":                      "kubernetesredis",
	}
	for in, want := range cases {
		if got := kindSlugFromFixtureFilename(in); got != want {
			t.Errorf("kindSlugFromFixtureFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

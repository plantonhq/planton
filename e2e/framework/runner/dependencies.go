package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/e2e/framework/provider"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
)

// DependencySource identifies where a test dependency was declared.
type DependencySource string

const (
	// SourceRegistry means the dependency came from the component's
	// CloudResourceKindMeta.prerequisites graph in the proto registry. Declaring
	// `prerequisites: [X]` on a kind is enough for the harness to install X first.
	SourceRegistry DependencySource = "registry"

	// SourceFixture means the dependency came from a hand-authored manifest in the
	// component's v1/e2e/fixtures/ directory -- an explicit per-consumer override.
	SourceFixture DependencySource = "fixture"
)

// Dependency is a single prerequisite deployment that must exist before a
// component's own scenario is applied.
type Dependency struct {
	// Source records whether this came from the registry or a fixture override.
	Source DependencySource

	// KindSlug is the lowercase component directory name of the dependency
	// (e.g. "kubernetesgatewayapicrds").
	KindSlug string

	// ManifestPath is the absolute path to the KRM manifest deployed for it.
	ManifestPath string
}

// DependencyState tracks a deployed dependency so it can be torn down later.
type DependencyState struct {
	Dependency     Dependency
	ModuleDir      string
	StackName      string
	BackendURL     string
	StackInputPath string
}

// ResolveDependencies builds the ordered, deduplicated list of prerequisite
// deployments for a component. Two sources are merged:
//
//  1. Registry prerequisites -- the component's CloudResourceKindMeta.prerequisites
//     graph (resolved transitively, deploy-first order) from the proto registry.
//     This is the source of truth: declaring `prerequisites: [X]` on a kind makes
//     the harness install X before the component, with no per-component wiring.
//  2. Fixture overrides -- hand-authored manifests in <component>/v1/e2e/fixtures/.
//     A fixture lets a consumer pin a dependency's exact config when the flat
//     registry list cannot express it.
//
// When both supply the same kind, the fixture wins (it is the explicit override)
// and the registry entry is dropped. Registry prerequisites are ordered first
// (most foundational); fixtures follow. Teardown runs in reverse, so foundational
// dependencies are removed last.
//
// The install manifest for a registry prerequisite is, in order of preference:
//   - <dep>/v1/e2e/prerequisite.yaml      (the dependency's published install profile)
//   - <dep>/v1/e2e/scenarios/minimal.yaml (fallback to its minimal scenario)
func ResolveDependencies(repoRoot, componentProvider, component string) ([]Dependency, error) {
	fixtures, err := resolveFixtureDependencies(repoRoot, componentProvider, component)
	if err != nil {
		return nil, err
	}

	fixtureKinds := make(map[string]bool, len(fixtures))
	for _, f := range fixtures {
		fixtureKinds[f.KindSlug] = true
	}

	registry, err := resolveRegistryDependencies(repoRoot, componentProvider, component, fixtureKinds)
	if err != nil {
		return nil, err
	}

	// Registry prerequisites first (most foundational), fixtures after.
	return append(registry, fixtures...), nil
}

// resolveRegistryDependencies maps a component to its transitive registry
// prerequisites, skipping any kind already supplied by a fixture override.
func resolveRegistryDependencies(repoRoot, componentProvider, component string, skip map[string]bool) ([]Dependency, error) {
	kind := crkreflect.KindFromString(component)
	if kind == cloudresourcekind.CloudResourceKind_unspecified {
		// Not a registered kind (or an alias mismatch); no registry prerequisites.
		return nil, nil
	}

	prereqs, err := crkreflect.TransitivePrerequisites(kind)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving prerequisites for %s", component)
	}

	var deps []Dependency
	for _, p := range prereqs {
		slug := strings.ToLower(p.String())
		if skip[slug] {
			continue // a fixture overrides this dependency
		}
		manifestPath, err := prerequisiteManifestPath(repoRoot, componentProvider, slug)
		if err != nil {
			return nil, err
		}
		deps = append(deps, Dependency{
			Source:       SourceRegistry,
			KindSlug:     slug,
			ManifestPath: manifestPath,
		})
	}
	return deps, nil
}

// prerequisiteManifestPath returns the manifest used to install a registry
// prerequisite: its published prerequisite.yaml if present, else its minimal
// scenario. Errors if neither exists, so a missing install profile fails loudly
// rather than silently skipping a required dependency.
func prerequisiteManifestPath(repoRoot, componentProvider, slug string) (string, error) {
	base := filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", componentProvider, slug, "v1", "e2e")
	prereq := filepath.Join(base, "prerequisite.yaml")
	if pathExists(prereq) {
		return prereq, nil
	}
	minimal := filepath.Join(base, "scenarios", "minimal.yaml")
	if pathExists(minimal) {
		return minimal, nil
	}
	return "", errors.Errorf("no install manifest for prerequisite %q: expected %s or %s", slug, prereq, minimal)
}

// resolveFixtureDependencies discovers fixture override manifests in numeric order.
func resolveFixtureDependencies(repoRoot, componentProvider, component string) ([]Dependency, error) {
	fixturesDir := filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", componentProvider, component, "v1", "e2e", "fixtures")

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "reading fixtures directory %s", fixturesDir)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	var deps []Dependency
	for _, filename := range files {
		fixtureName := strings.TrimSuffix(strings.TrimSuffix(filename, ".yaml"), ".yml")
		deps = append(deps, Dependency{
			Source:       SourceFixture,
			KindSlug:     kindSlugFromFixtureFilename(fixtureName),
			ManifestPath: filepath.Join(fixturesDir, filename),
		})
	}
	return deps, nil
}

// DeployDependencies resolves and deploys all prerequisite deployments for a
// component (registry prerequisites + fixture overrides) in order, via Pulumi.
// Returns the deployed states (needed for teardown) and any error. On the first
// failure it stops and returns whatever was already deployed so the caller can
// tear it down.
func DeployDependencies(ctx context.Context, repoRoot, componentProvider, component, backendURL, runID string, harness provider.Harness) ([]DependencyState, error) {
	deps, err := ResolveDependencies(repoRoot, componentProvider, component)
	if err != nil {
		return nil, err
	}
	if len(deps) == 0 {
		return nil, nil
	}

	fmt.Printf("  [deps] Deploying %d dependencies for %s\n", len(deps), component)

	var deployed []DependencyState
	for _, dep := range deps {
		state, err := deployDependency(ctx, repoRoot, componentProvider, dep, backendURL, runID, harness)
		// A non-empty stack name means Pulumi created resources we must track for
		// teardown, even if verification afterwards failed.
		if state.StackName != "" {
			deployed = append(deployed, state)
		}
		if err != nil {
			return deployed, err
		}
	}
	return deployed, nil
}

// deployDependency builds the stack input, runs `pulumi up`, and verifies the
// dependency is present. The dependency's own pulumi module is always used
// (dependencies deploy via Pulumi even when the component under test uses
// Terraform, matching the historical fixture behavior).
func deployDependency(ctx context.Context, repoRoot, componentProvider string, dep Dependency, backendURL, runID string, harness provider.Harness) (DependencyState, error) {
	moduleDir := filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", componentProvider, dep.KindSlug, "v1", "iac", "pulumi")
	if !pathExists(moduleDir) {
		return DependencyState{}, errors.Errorf("dependency %q (%s) pulumi module not found at %s", dep.KindSlug, dep.Source, moduleDir)
	}

	stackName := GenerateStackName("dep-"+dep.KindSlug, runID)
	if len(stackName) > 50 {
		stackName = stackName[:50]
	}

	fmt.Printf("  [deps] Deploying %s dependency %s...\n", dep.Source, dep.KindSlug)
	start := time.Now()

	stackInputPath, err := BuildStackInput(dep.ManifestPath, moduleDir)
	if err != nil {
		return DependencyState{}, errors.Wrapf(err, "failed to build stack input for dependency %q", dep.KindSlug)
	}

	if _, err := PulumiDeploy(moduleDir, stackName, backendURL, stackInputPath); err != nil {
		return DependencyState{}, errors.Wrapf(err, "failed to deploy dependency %q", dep.KindSlug)
	}

	state := DependencyState{
		Dependency:     dep,
		ModuleDir:      moduleDir,
		StackName:      stackName,
		BackendURL:     backendURL,
		StackInputPath: stackInputPath,
	}

	verifyCtx := context.WithValue(ctx, provider.ManifestPathKey{}, dep.ManifestPath)
	if err := harness.VerifyDeployed(verifyCtx, dep.KindSlug, nil); err != nil {
		return state, errors.Wrapf(err, "dependency %q deployed but verification failed", dep.KindSlug)
	}

	fmt.Printf("  [deps] Dependency %s deployed and verified in %s\n", dep.KindSlug, time.Since(start).Round(time.Second))
	return state, nil
}

// TeardownDependencies destroys deployed dependencies in reverse order.
func TeardownDependencies(deployed []DependencyState) {
	for i := len(deployed) - 1; i >= 0; i-- {
		dep := deployed[i]
		fmt.Printf("  [deps] Destroying dependency %s...\n", dep.Dependency.KindSlug)

		if _, err := PulumiDestroy(dep.ModuleDir, dep.StackName, dep.BackendURL, dep.StackInputPath); err != nil {
			fmt.Printf("  [WARN] dependency %s destroy failed: %v\n", dep.Dependency.KindSlug, err)
			continue
		}
		if err := PulumiRemoveStack(dep.ModuleDir, dep.StackName, dep.BackendURL); err != nil {
			fmt.Printf("  [WARN] dependency %s stack removal failed: %v\n", dep.Dependency.KindSlug, err)
		}
	}
}

// kindSlugFromFixtureFilename derives the component directory name from a fixture
// filename: "01-kuberneteszalandopostgresoperator" -> "kuberneteszalandopostgresoperator".
func kindSlugFromFixtureFilename(fixtureName string) string {
	parts := strings.SplitN(fixtureName, "-", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return fixtureName
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

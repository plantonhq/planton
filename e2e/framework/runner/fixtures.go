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
	"github.com/plantonhq/openmcf/e2e/framework/provider"
)

// FixtureState tracks a deployed fixture so it can be torn down later.
type FixtureState struct {
	Name           string
	ManifestPath   string
	ModuleDir      string
	StackName      string
	BackendURL     string
	StackInputPath string
}

// DeployFixtures discovers and deploys all fixture YAML files in numeric order.
// Returns the deployed fixture states (needed for teardown) and any error.
func DeployFixtures(ctx context.Context, repoRoot string, componentProvider string, component string, backendURL string, runID string, harness provider.Harness) ([]FixtureState, error) {
	fixturesDir := filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", componentProvider, component, "v1", "e2e", "fixtures")

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to read fixtures directory %s", fixturesDir)
	}

	var fixtureFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			fixtureFiles = append(fixtureFiles, name)
		}
	}

	sort.Strings(fixtureFiles)

	if len(fixtureFiles) == 0 {
		return nil, nil
	}

	fmt.Printf("  [fixtures] Deploying %d fixtures for %s\n", len(fixtureFiles), component)

	var deployed []FixtureState
	for _, filename := range fixtureFiles {
		manifestPath := filepath.Join(fixturesDir, filename)
		fixtureName := strings.TrimSuffix(strings.TrimSuffix(filename, ".yaml"), ".yml")

		fixtureKind := extractFixtureKind(fixtureName)

		moduleDir := filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", componentProvider, fixtureKind, "v1", "iac", "pulumi")
		if _, err := os.Stat(moduleDir); err != nil {
			return deployed, errors.Wrapf(err, "fixture %q pulumi module not found at %s", fixtureName, moduleDir)
		}

		stackName := GenerateStackName("fix-"+fixtureName, runID)
		if len(stackName) > 50 {
			stackName = stackName[:50]
		}

		fmt.Printf("  [fixtures] Deploying fixture %s...\n", fixtureName)
		start := time.Now()

		stackInputPath, err := BuildStackInput(manifestPath, moduleDir)
		if err != nil {
			return deployed, errors.Wrapf(err, "failed to build stack input for fixture %q", fixtureName)
		}

		_, err = PulumiDeploy(moduleDir, stackName, backendURL, stackInputPath)
		if err != nil {
			return deployed, errors.Wrapf(err, "failed to deploy fixture %q", fixtureName)
		}

		state := FixtureState{
			Name:           fixtureName,
			ManifestPath:   manifestPath,
			ModuleDir:      moduleDir,
			StackName:      stackName,
			BackendURL:     backendURL,
			StackInputPath: stackInputPath,
		}
		deployed = append(deployed, state)

		// Verify fixture is running via harness
		verifyCtx := context.WithValue(ctx, provider.ManifestPathKey{}, manifestPath)
		if err := harness.VerifyDeployed(verifyCtx, fixtureKind, nil); err != nil {
			return deployed, errors.Wrapf(err, "fixture %q deployed but verification failed", fixtureName)
		}

		fmt.Printf("  [fixtures] Fixture %s deployed and verified in %s\n", fixtureName, time.Since(start).Round(time.Second))
	}

	return deployed, nil
}

// TeardownFixtures destroys all deployed fixtures in reverse order.
func TeardownFixtures(deployed []FixtureState) {
	for i := len(deployed) - 1; i >= 0; i-- {
		fix := deployed[i]
		fmt.Printf("  [fixtures] Destroying fixture %s...\n", fix.Name)

		_, err := PulumiDestroy(fix.ModuleDir, fix.StackName, fix.BackendURL, fix.StackInputPath)
		if err != nil {
			fmt.Printf("  [WARN] fixture %s destroy failed: %v\n", fix.Name, err)
			continue
		}

		if err := PulumiRemoveStack(fix.ModuleDir, fix.StackName, fix.BackendURL); err != nil {
			fmt.Printf("  [WARN] fixture %s stack removal failed: %v\n", fix.Name, err)
		}
	}
}

// extractFixtureKind derives the component directory name from a fixture filename.
// "01-kuberneteszalandopostgresoperator" -> "kuberneteszalandopostgresoperator"
func extractFixtureKind(fixtureName string) string {
	parts := strings.SplitN(fixtureName, "-", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return fixtureName
}

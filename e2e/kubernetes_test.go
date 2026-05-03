//go:build e2e

package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/plantonhq/openmcf/e2e/framework/provider"
	"github.com/plantonhq/openmcf/e2e/framework/runner"
)

func TestKubernetesNamespace_Pulumi(t *testing.T) {
	runKubernetesComponentTest(t, "kubernetesnamespace")
}

func TestKubernetesDeployment_Pulumi(t *testing.T) {
	runKubernetesComponentTest(t, "kubernetesdeployment")
}

func TestKubernetesSecret_Pulumi(t *testing.T) {
	runKubernetesComponentTest(t, "kubernetessecret")
}

func TestKubernetesService_Pulumi(t *testing.T) {
	runKubernetesComponentTest(t, "kubernetesservice")
}

func TestKubernetesStatefulSet_Pulumi(t *testing.T) {
	runKubernetesComponentTest(t, "kubernetesstatefulset")
}

func runKubernetesComponentTest(t *testing.T, component string) {
	t.Helper()

	moduleDir := filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", "kubernetes", component, "v1", "iac", "pulumi")
	manifestPath := filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", "kubernetes", component, "v1", "iac", "hack", "manifest.yaml")

	// Verify the component exists
	if !fileExists(moduleDir) {
		t.Skipf("component %s pulumi module not found at %s", component, moduleDir)
	}
	if !fileExists(manifestPath) {
		t.Skipf("component %s manifest not found at %s", component, manifestPath)
	}

	stackName := runner.GenerateStackName(component, runID)

	tc := &provider.ComponentTestContext{
		Component:    component,
		Provider:     "kubernetes",
		Engine:       "pulumi",
		ModuleDir:    moduleDir,
		ManifestPath: manifestPath,
		StackName:    stackName,
		BackendURL:   pulumiBackendURL,
	}

	ctx := context.Background()
	result := runner.RunComponentTest(ctx, tc, testHarness)

	// Report results
	for _, phase := range result.Phases {
		status := "PASS"
		if !phase.Passed {
			status = "FAIL"
		}
		t.Logf("  %s: %s (%s)", phase.Phase, status, phase.Duration)
		if phase.Error != nil {
			t.Logf("    Error: %v", phase.Error)
		}
	}

	if !result.Passed {
		t.Fatalf("E2E test for %s failed (total: %s)", component, result.Duration)
	}

	t.Logf("E2E test for %s passed (total: %s)", component, result.Duration)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

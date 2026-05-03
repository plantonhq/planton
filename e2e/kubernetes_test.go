//go:build e2e

package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/plantonhq/openmcf/e2e/framework/discovery"
	"github.com/plantonhq/openmcf/e2e/framework/provider"
	"github.com/plantonhq/openmcf/e2e/framework/runner"
)

// Kubernetes Tier 1 components that have Pulumi modules.
var kubernetesTier1Components = []string{
	"kubernetesnamespace",
	"kubernetesdeployment",
	"kubernetesstatefulset",
	"kubernetessecret",
	"kubernetesservice",
}

func TestKubernetesNamespace_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesnamespace")
}

func TestKubernetesDeployment_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesdeployment")
}

func TestKubernetesStatefulSet_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesstatefulset")
}

func TestKubernetesSecret_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetessecret")
}

func TestKubernetesService_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesservice")
}

func runAllScenariosForComponent(t *testing.T, component string) {
	t.Helper()

	moduleDir := filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", "kubernetes", component, "v1", "iac", "pulumi")
	if !fileExists(moduleDir) {
		t.Skipf("component %s pulumi module not found at %s", component, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "kubernetes", component)
	if err != nil {
		t.Fatalf("failed to discover test scenarios for %s: %v", component, err)
	}

	if len(scenarios) == 0 {
		t.Skipf("no test scenarios found for %s in e2e/testdata/kubernetes/%s/", component, component)
	}

	t.Logf("Discovered %d scenarios for %s", len(scenarios), component)

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Name, func(t *testing.T) {
			runSingleScenario(t, component, moduleDir, scenario)
		})
	}
}

func runSingleScenario(t *testing.T, component, moduleDir string, scenario discovery.TestScenario) {
	t.Helper()

	stackName := runner.GenerateStackName(component+"-"+scenario.Name, runID)
	// Pulumi stack names have a max length; truncate if needed
	if len(stackName) > 50 {
		stackName = stackName[:50]
	}

	tc := &provider.ComponentTestContext{
		Component:    component,
		Provider:     "kubernetes",
		Engine:       "pulumi",
		ModuleDir:    moduleDir,
		ManifestPath: scenario.ManifestPath,
		StackName:    stackName,
		BackendURL:   pulumiBackendURL,
	}

	ctx := context.Background()
	result := runner.RunComponentTest(ctx, tc, testHarness)

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
		t.Fatalf("scenario %s/%s failed (total: %s)", component, scenario.Name, result.Duration)
	}

	t.Logf("scenario %s/%s passed (total: %s)", component, scenario.Name, result.Duration)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

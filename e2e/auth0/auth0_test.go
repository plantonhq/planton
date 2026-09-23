//go:build e2e

package auth0

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	auth0e2e "github.com/plantonhq/planton/catalog/auth0/aa_e2e"
	"github.com/plantonhq/planton/e2e/framework/discovery"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	profilepkg "github.com/plantonhq/planton/pkg/e2e/profile"
	componentv1 "github.com/plantonhq/planton/qa/componente2eprofile/v1"
)

var (
	testHarness      *auth0e2e.Harness
	repoRoot         string
	runID            string
	pulumiBackendURL string
)

func TestMain(m *testing.M) {
	var err error
	repoRoot, err = filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve repo root: %v\n", err)
		os.Exit(1)
	}

	runID = uuid.New().String()[:8]

	backendDir, err := os.MkdirTemp("", "planton-e2e-auth0-pulumi-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp backend dir: %v\n", err)
		os.Exit(1)
	}
	pulumiBackendURL = "file://" + backendDir
	defer os.RemoveAll(backendDir)

	if err := runner.PulumiLogin(pulumiBackendURL); err != nil {
		fmt.Fprintf(os.Stderr, "failed to login to pulumi backend: %v\n", err)
		os.Exit(1)
	}

	testHarness = auth0e2e.NewHarness()
	ctx := context.Background()
	if err := testHarness.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup Auth0 harness: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testHarness.Teardown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to teardown Auth0 harness: %v\n", err)
	}

	os.Exit(code)
}

// --- Auth0 Client ---

func TestAuth0Client_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "auth0client", "pulumi") }
func TestAuth0Client_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "auth0client", "terraform")
}

// --- Auth0 Connection ---

func TestAuth0Connection_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "auth0connection", "pulumi")
}
func TestAuth0Connection_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "auth0connection", "terraform")
}

// --- Auth0 Resource Server ---

func TestAuth0ResourceServer_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "auth0resourceserver", "pulumi")
}
func TestAuth0ResourceServer_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "auth0resourceserver", "terraform")
}

// --- Auth0 Action ---

func TestAuth0Action_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "auth0action", "pulumi") }
func TestAuth0Action_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "auth0action", "terraform")
}

// --- Auth0 Event Stream ---

func TestAuth0EventStream_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "auth0eventstream", "pulumi")
}
func TestAuth0EventStream_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "auth0eventstream", "terraform")
}

// --- Auth0 Role ---

func TestAuth0Role_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "auth0role", "pulumi") }
func TestAuth0Role_Terraform(t *testing.T) { runAllScenariosForComponent(t, "auth0role", "terraform") }

// --- Auth0 User ---

func TestAuth0User_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "auth0user", "pulumi") }
func TestAuth0User_Terraform(t *testing.T) { runAllScenariosForComponent(t, "auth0user", "terraform") }

// runAllScenariosForComponent discovers and runs all E2E scenarios for an Auth0 component.
func runAllScenariosForComponent(t *testing.T, component, engine string) {
	t.Helper()

	if cp, err := profilepkg.LoadComponentProfile(repoRoot, "auth0", component); err == nil && cp.Spec != nil {
		switch cp.Spec.Status {
		case componentv1.ComponentE2EProfileSpec_deferred,
			componentv1.ComponentE2EProfileSpec_skip,
			componentv1.ComponentE2EProfileSpec_stub:
			reason := cp.Spec.DeferredReason
			if reason == "" {
				reason = cp.Spec.Status.String()
			}
			t.Skipf("component %s E2E profile status is %s: %s", component, cp.Spec.Status, reason)
		}
	}

	moduleDir, err := discovery.ModuleDir(repoRoot, "auth0", component, engine)
	if err != nil {
		t.Fatalf("failed to locate %s %s module: %v", component, engine, err)
	}

	if !fileExists(moduleDir) {
		t.Skipf("component %s %s module not found at %s", component, engine, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "auth0", component)
	if err != nil {
		t.Fatalf("failed to discover test scenarios for %s: %v", component, err)
	}

	if len(scenarios) == 0 {
		t.Skipf("no test scenarios found for %s", component)
	}

	t.Logf("Discovered %d scenarios for %s [%s]", len(scenarios), component, engine)

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Name, func(t *testing.T) {
			runSingleScenario(t, component, moduleDir, engine, scenario)
		})
	}
}

func runSingleScenario(t *testing.T, component, moduleDir, engine string, scenario discovery.TestScenario) {
	t.Helper()

	tc := &provider.ComponentTestContext{
		Component:    component,
		Provider:     "auth0",
		Engine:       engine,
		ModuleDir:    moduleDir,
		ManifestPath: scenario.ManifestPath,
		RepoRoot:     repoRoot,
		RunID:        runID,
		T:            t,
		// Dependencies always deploy via Pulumi — even for Terraform
		// scenarios — so the backend URL must be set unconditionally.
		// Leaving it empty makes the dependency stacks fall back to the
		// machine's ambient `pulumi login` backend, coupling the run to
		// stale developer state.
		BackendURL: pulumiBackendURL,
	}

	if engine == "pulumi" {
		// GenerateStackName enforces the length cap uniqueness-preservingly
		// (blind truncation here would collide long kind names' scenarios).
		tc.StackName = runner.GenerateStackName(component+"-"+scenario.Name, runID)
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
		t.Fatalf("scenario %s/%s [%s] failed (total: %s)", component, scenario.Name, engine, result.Duration)
	}

	t.Logf("scenario %s/%s [%s] passed (total: %s)", component, scenario.Name, engine, result.Duration)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

//go:build e2e

// Package aws contains end-to-end tests that provision real AWS resources via
// OpenMCF IaC modules and verify them through the AWS SDK. Credentials come from
// the ambient chain (local AWS SSO or GitHub Actions OIDC -- never a stored
// secret); see the aa_e2e harness package.
//
// Run with: go test -tags=e2e -timeout=30m -v ./e2e/aws/...
package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	awse2e "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/aa_e2e"
	"github.com/plantonhq/openmcf/e2e/framework/discovery"
	"github.com/plantonhq/openmcf/e2e/framework/provider"
	"github.com/plantonhq/openmcf/e2e/framework/runner"
)

var (
	testHarness      *awse2e.Harness
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

	backendDir, err := os.MkdirTemp("", "openmcf-e2e-aws-pulumi-*")
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

	testHarness = awse2e.NewHarness()
	ctx := context.Background()
	if err := testHarness.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup AWS harness: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testHarness.Teardown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to teardown AWS harness: %v\n", err)
	}

	os.Exit(code)
}

// --- AWS S3 Bucket (walking skeleton) ---

func TestAwsS3Bucket_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "awss3bucket", "pulumi") }
func TestAwsS3Bucket_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awss3bucket", "terraform")
}

// --- AWS VPC (thin root of the networking graph) ---

func TestAwsVpc_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "awsvpc", "pulumi") }
func TestAwsVpc_Terraform(t *testing.T) { runAllScenariosForComponent(t, "awsvpc", "terraform") }

// --- AWS Subnet (first composed topology: deploys an AwsVpc prerequisite) ---

func TestAwsSubnet_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "awssubnet", "pulumi") }
func TestAwsSubnet_Terraform(t *testing.T) { runAllScenariosForComponent(t, "awssubnet", "terraform") }

// --- AWS NAT Gateway (deep composed topology: AwsVpc -> AwsSubnet -> AwsElasticIp) ---

func TestAwsNatGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsnatgateway", "pulumi")
}
func TestAwsNatGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsnatgateway", "terraform")
}

// --- AWS Internet Gateway (attaches to a gateway-free AwsVpc prerequisite) ---

func TestAwsInternetGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsinternetgateway", "pulumi")
}
func TestAwsInternetGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsinternetgateway", "terraform")
}

// --- AWS Egress-Only Internet Gateway (IPv6 outbound-only; attaches to an AwsVpc prerequisite) ---

func TestAwsEgressOnlyInternetGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsegressonlyinternetgateway", "pulumi")
}
func TestAwsEgressOnlyInternetGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsegressonlyinternetgateway", "terraform")
}

// runAllScenariosForComponent discovers and runs all E2E scenarios for an AWS component.
func runAllScenariosForComponent(t *testing.T, component, engine string) {
	t.Helper()

	var moduleDir string
	switch engine {
	case "pulumi":
		moduleDir = filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", "aws", component, "v1", "iac", "pulumi")
	case "terraform":
		moduleDir = filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", "aws", component, "v1", "iac", "tf")
	default:
		t.Fatalf("unsupported engine: %s", engine)
	}

	if !fileExists(moduleDir) {
		t.Skipf("component %s %s module not found at %s", component, engine, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "aws", component)
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
		Provider:     "aws",
		Engine:       engine,
		ModuleDir:    moduleDir,
		ManifestPath: scenario.ManifestPath,
		RepoRoot:     repoRoot,
		RunID:        runID,
		T:            t,
	}

	if engine == "pulumi" {
		stackName := runner.GenerateStackName(component+"-"+scenario.Name, runID)
		if len(stackName) > 50 {
			stackName = stackName[:50]
		}
		tc.StackName = stackName
		tc.BackendURL = pulumiBackendURL
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

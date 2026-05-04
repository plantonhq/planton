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

// Kubernetes Tier 1 components: native K8s resources, zero dependencies.
var kubernetesTier1Components = []string{
	"kubernetesnamespace",
	"kubernetesdeployment",
	"kubernetesstatefulset",
	"kubernetessecret",
	"kubernetesservice",
}

// Kubernetes Tier 3 components: operator-dependent, need fixtures deployed first.
// The fixture system (DD-007) automatically deploys prerequisite operators
// by reading CloudResourceKindMeta.prerequisites from proto options and
// deploying the numbered YAML files in each component's e2e/fixtures/ directory.
var kubernetesTier3Components = []string{
	"kubernetespostgres",
	"kuberneteskafka",
	"kuberneteselasticsearch",
	"kubernetesmongodb",
	"kubernetessolr",
	"kubernetesclickhouse",
}

// Kubernetes Tier 2 components: Helm-based, self-contained chart installs.
var kubernetesTier2Components = []string{
	"kubernetesredis",
	"kubernetesgrafana",
	"kubernetesopenbao",
	"kubernetesargocd",
	"kuberneteslocust",
	"kubernetesnats",
	"kubernetesneo4j",
	"kubernetesjenkins",
	"kubernetessolroperator",
	"kubernetesperconamongooperator",
	"kubernetesperconamysqloperator",
	"kubernetesperconapostgresoperator",
	"kubernetesgitlab",
}

// --- Tier 1 test entry points ---

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

// --- Tier 2 test entry points (Helm-based) ---

func TestKubernetesRedis_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesredis")
}

func TestKubernetesGrafana_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesgrafana")
}

func TestKubernetesOpenBao_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesopenbao")
}

func TestKubernetesArgoCD_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesargocd")
}

func TestKubernetesLocust_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kuberneteslocust")
}

func TestKubernetesNats_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesnats")
}

func TestKubernetesNeo4j_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesneo4j")
}

func TestKubernetesJenkins_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesjenkins")
}

func TestKubernetesSolrOperator_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetessolroperator")
}

func TestKubernetesPerconaMongoOperator_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesperconamongooperator")
}

func TestKubernetesPerconaMysqlOperator_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesperconamysqloperator")
}

func TestKubernetesPerconaPostgresOperator_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesperconapostgresoperator")
}

func TestKubernetesGitlab_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesgitlab")
}

// --- Tier 3 test entry points (operator-dependent, fixtures deployed automatically) ---

func TestKubernetesPostgres_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetespostgres")
}

func TestKubernetesKafka_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kuberneteskafka")
}

func TestKubernetesElasticsearch_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kuberneteselasticsearch")
}

func TestKubernetesMongodb_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesmongodb")
}

func TestKubernetesSolr_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetessolr")
}

func TestKubernetesClickHouse_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "kubernetesclickhouse")
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
		t.Skipf("no test scenarios found for %s in %s/v1/e2e/", component, component)
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
		RepoRoot:     repoRoot,
		RunID:        runID,
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

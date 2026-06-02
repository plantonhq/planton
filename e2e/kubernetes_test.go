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
	"kubernetescronjob",
	"kubernetesjob",
	"kubernetesdaemonset",
	"kubernetesmanifest",
}

// Kubernetes Tier 3 components: operator-dependent. Each declares its operator
// as a registry prerequisite (CloudResourceKindMeta.prerequisites) AND ships an
// explicit e2e/fixtures/ override that pins the operator's exact config; the
// override wins, so the fixture is what actually deploys here. Either way the
// harness installs the operator before the test and tears it down after
// (see e2e/framework/runner/dependencies.go -- ResolveDependencies).
var kubernetesTier3Components = []string{
	"kubernetespostgres",
	"kuberneteskafka",
	"kuberneteselasticsearch",
	"kubernetesmongodb",
	"kubernetessolr",
	"kubernetesclickhouse",
}

// Kubernetes Tier 4 components: operators, addons, and cluster-level infrastructure.
// Includes operators that were previously only exercised as Tier 3 fixtures,
// plus new components tested in session 010.
var kubernetesTier4Components = []string{
	// Operators already proven as Tier 3 fixtures -- now standalone
	"kuberneteszalandopostgresoperator",
	"kubernetesstrimzikafkaoperator",
	"kuberneteselasticoperator",
	"kubernetesaltinityoperator",
	// New Tier 4 (session 010)
	"kubernetesgatewayapicrds",
	"kubernetesgharunnerscalesetcontroller",
	"kubernetesrookcephoperator",
	"kubernetesexternalsecrets",
	"kubernetesingressnginx",
	"kubernetestekton",
	"kubernetestektonoperator",
	"kubernetesistio",
	// Istio base CRDs installer (868). The CRDs-only prerequisite for the typed
	// Istio API components; analog of kubernetesgatewayapicrds.
	"kubernetesistiobasecrds",
	// Gateway API deployment components (854-860). Each declares
	// KubernetesGatewayApiCrds as a registry prerequisite, which the harness
	// installs (experimental v1.5.1) before applying the route/gateway scenario.
	"kubernetesgatewayclass",
	"kubernetesgateway",
	"kuberneteshttproute",
	"kubernetesgrpcroute",
	"kubernetestcproute",
	"kubernetestlsroute",
	"kubernetesreferencegrant",
	// Istio API deployment components (861-867). Each declares
	// KubernetesIstioBaseCrds as a registry prerequisite, which the harness
	// installs (istio/base CRDs, no istiod) before applying the scenario.
	// Verification asserts the typed Istio CR exists.
	"kubernetespeerauthentication",
	"kubernetesrequestauthentication",
	"kubernetesauthorizationpolicy",
	"kubernetesserviceentry",
	"kubernetesdestinationrule",
	"kubernetesenvoyfilter",
	"kubernetestelemetry",
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
	"kubernetestemporal",
	"kubernetessignoz",
}

// ─── Tier 1 Pulumi ──────────────────────────────────────────────────────────

func TestKubernetesNamespace_Pulumi(t *testing.T)  { runAllScenariosForComponent(t, "kubernetesnamespace", "pulumi") }
func TestKubernetesDeployment_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetesdeployment", "pulumi") }
func TestKubernetesStatefulSet_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetesstatefulset", "pulumi") }
func TestKubernetesSecret_Pulumi(t *testing.T)     { runAllScenariosForComponent(t, "kubernetessecret", "pulumi") }
func TestKubernetesService_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "kubernetesservice", "pulumi") }
func TestKubernetesCronJob_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "kubernetescronjob", "pulumi") }
func TestKubernetesJob_Pulumi(t *testing.T)        { runAllScenariosForComponent(t, "kubernetesjob", "pulumi") }
func TestKubernetesDaemonSet_Pulumi(t *testing.T)  { runAllScenariosForComponent(t, "kubernetesdaemonset", "pulumi") }
func TestKubernetesManifest_Pulumi(t *testing.T)   { runAllScenariosForComponent(t, "kubernetesmanifest", "pulumi") }

// ─── Tier 1 Terraform ───────────────────────────────────────────────────────

func TestKubernetesNamespace_Terraform(t *testing.T)  { runAllScenariosForComponent(t, "kubernetesnamespace", "terraform") }
func TestKubernetesDeployment_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetesdeployment", "terraform") }
func TestKubernetesStatefulSet_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetesstatefulset", "terraform") }
func TestKubernetesSecret_Terraform(t *testing.T)     { runAllScenariosForComponent(t, "kubernetessecret", "terraform") }
func TestKubernetesService_Terraform(t *testing.T)    { runAllScenariosForComponent(t, "kubernetesservice", "terraform") }
func TestKubernetesCronJob_Terraform(t *testing.T)    { runAllScenariosForComponent(t, "kubernetescronjob", "terraform") }
func TestKubernetesJob_Terraform(t *testing.T)        { runAllScenariosForComponent(t, "kubernetesjob", "terraform") }
func TestKubernetesDaemonSet_Terraform(t *testing.T)  { runAllScenariosForComponent(t, "kubernetesdaemonset", "terraform") }
func TestKubernetesManifest_Terraform(t *testing.T)   { runAllScenariosForComponent(t, "kubernetesmanifest", "terraform") }

// ─── Tier 2 Pulumi (Helm-based) ─────────────────────────────────────────────

func TestKubernetesRedis_Pulumi(t *testing.T)                  { runAllScenariosForComponent(t, "kubernetesredis", "pulumi") }
func TestKubernetesGrafana_Pulumi(t *testing.T)                { runAllScenariosForComponent(t, "kubernetesgrafana", "pulumi") }
func TestKubernetesOpenBao_Pulumi(t *testing.T)                { runAllScenariosForComponent(t, "kubernetesopenbao", "pulumi") }
func TestKubernetesArgoCD_Pulumi(t *testing.T)                 { runAllScenariosForComponent(t, "kubernetesargocd", "pulumi") }
func TestKubernetesLocust_Pulumi(t *testing.T)                 { runAllScenariosForComponent(t, "kuberneteslocust", "pulumi") }
func TestKubernetesNats_Pulumi(t *testing.T)                   { runAllScenariosForComponent(t, "kubernetesnats", "pulumi") }
func TestKubernetesNeo4j_Pulumi(t *testing.T)                  { runAllScenariosForComponent(t, "kubernetesneo4j", "pulumi") }
func TestKubernetesJenkins_Pulumi(t *testing.T)                { runAllScenariosForComponent(t, "kubernetesjenkins", "pulumi") }
func TestKubernetesSolrOperator_Pulumi(t *testing.T)           { runAllScenariosForComponent(t, "kubernetessolroperator", "pulumi") }
func TestKubernetesPerconaMongoOperator_Pulumi(t *testing.T)   { runAllScenariosForComponent(t, "kubernetesperconamongooperator", "pulumi") }
func TestKubernetesPerconaMysqlOperator_Pulumi(t *testing.T)   { runAllScenariosForComponent(t, "kubernetesperconamysqloperator", "pulumi") }
func TestKubernetesPerconaPostgresOperator_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetesperconapostgresoperator", "pulumi") }
func TestKubernetesGitlab_Pulumi(t *testing.T)                 { runAllScenariosForComponent(t, "kubernetesgitlab", "pulumi") }
func TestKubernetesTemporal_Pulumi(t *testing.T)               { runAllScenariosForComponent(t, "kubernetestemporal", "pulumi") }
func TestKubernetesSignoz_Pulumi(t *testing.T)                 { runAllScenariosForComponent(t, "kubernetessignoz", "pulumi") }

// ─── Tier 2 Terraform (Helm-based) ──────────────────────────────────────────

func TestKubernetesRedis_Terraform(t *testing.T)                  { runAllScenariosForComponent(t, "kubernetesredis", "terraform") }
func TestKubernetesGrafana_Terraform(t *testing.T)                { runAllScenariosForComponent(t, "kubernetesgrafana", "terraform") }
func TestKubernetesArgoCD_Terraform(t *testing.T)                 { runAllScenariosForComponent(t, "kubernetesargocd", "terraform") }
func TestKubernetesLocust_Terraform(t *testing.T)                 { runAllScenariosForComponent(t, "kuberneteslocust", "terraform") }
func TestKubernetesNats_Terraform(t *testing.T)                   { runAllScenariosForComponent(t, "kubernetesnats", "terraform") }
func TestKubernetesSolrOperator_Terraform(t *testing.T)           { runAllScenariosForComponent(t, "kubernetessolroperator", "terraform") }
func TestKubernetesPerconaMongoOperator_Terraform(t *testing.T)   { runAllScenariosForComponent(t, "kubernetesperconamongooperator", "terraform") }
func TestKubernetesPerconaMysqlOperator_Terraform(t *testing.T)   { runAllScenariosForComponent(t, "kubernetesperconamysqloperator", "terraform") }
func TestKubernetesPerconaPostgresOperator_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetesperconapostgresoperator", "terraform") }
func TestKubernetesTemporal_Terraform(t *testing.T)                { runAllScenariosForComponent(t, "kubernetestemporal", "terraform") }
func TestKubernetesSignoz_Terraform(t *testing.T)                  { runAllScenariosForComponent(t, "kubernetessignoz", "terraform") }

// ─── Tier 3 Pulumi (operator-dependent) ─────────────────────────────────────

func TestKubernetesPostgres_Pulumi(t *testing.T)       { runAllScenariosForComponent(t, "kubernetespostgres", "pulumi") }
func TestKubernetesKafka_Pulumi(t *testing.T)          { runAllScenariosForComponent(t, "kuberneteskafka", "pulumi") }
func TestKubernetesElasticsearch_Pulumi(t *testing.T)  { runAllScenariosForComponent(t, "kuberneteselasticsearch", "pulumi") }
func TestKubernetesMongodb_Pulumi(t *testing.T)        { runAllScenariosForComponent(t, "kubernetesmongodb", "pulumi") }
func TestKubernetesSolr_Pulumi(t *testing.T)           { runAllScenariosForComponent(t, "kubernetessolr", "pulumi") }
func TestKubernetesClickHouse_Pulumi(t *testing.T)     { runAllScenariosForComponent(t, "kubernetesclickhouse", "pulumi") }

// ─── Tier 3 Terraform (operator-dependent) ──────────────────────────────────

func TestKubernetesPostgres_Terraform(t *testing.T)       { runAllScenariosForComponent(t, "kubernetespostgres", "terraform") }
func TestKubernetesKafka_Terraform(t *testing.T)          { runAllScenariosForComponent(t, "kuberneteskafka", "terraform") }
func TestKubernetesElasticsearch_Terraform(t *testing.T)  { runAllScenariosForComponent(t, "kuberneteselasticsearch", "terraform") }
func TestKubernetesMongodb_Terraform(t *testing.T)        { runAllScenariosForComponent(t, "kubernetesmongodb", "terraform") }
func TestKubernetesSolr_Terraform(t *testing.T)           { runAllScenariosForComponent(t, "kubernetessolr", "terraform") }
func TestKubernetesClickHouse_Terraform(t *testing.T)     { runAllScenariosForComponent(t, "kubernetesclickhouse", "terraform") }

// ─── Tier 4 Pulumi (operators, addons) ──────────────────────────────────────

func TestKubernetesZalandoPostgresOperator_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "kuberneteszalandopostgresoperator", "pulumi") }
func TestKubernetesStrimziKafkaOperator_Pulumi(t *testing.T)       { runAllScenariosForComponent(t, "kubernetesstrimzikafkaoperator", "pulumi") }
func TestKubernetesElasticOperator_Pulumi(t *testing.T)            { runAllScenariosForComponent(t, "kuberneteselasticoperator", "pulumi") }
func TestKubernetesAltinityOperator_Pulumi(t *testing.T)           { runAllScenariosForComponent(t, "kubernetesaltinityoperator", "pulumi") }
func TestKubernetesGatewayApiCrds_Pulumi(t *testing.T)             { runAllScenariosForComponent(t, "kubernetesgatewayapicrds", "pulumi") }
func TestKubernetesGhaRunnerScaleSetController_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetesgharunnerscalesetcontroller", "pulumi") }
func TestKubernetesRookCephOperator_Pulumi(t *testing.T)           { runAllScenariosForComponent(t, "kubernetesrookcephoperator", "pulumi") }
func TestKubernetesExternalSecrets_Pulumi(t *testing.T)            { runAllScenariosForComponent(t, "kubernetesexternalsecrets", "pulumi") }
func TestKubernetesIngressNginx_Pulumi(t *testing.T)               { runAllScenariosForComponent(t, "kubernetesingressnginx", "pulumi") }
func TestKubernetesTekton_Pulumi(t *testing.T)                     { runAllScenariosForComponent(t, "kubernetestekton", "pulumi") }
func TestKubernetesTektonOperator_Pulumi(t *testing.T)             { runAllScenariosForComponent(t, "kubernetestektonoperator", "pulumi") }
func TestKubernetesIstio_Pulumi(t *testing.T)                      { runAllScenariosForComponent(t, "kubernetesistio", "pulumi") }
func TestKubernetesIstioBaseCrds_Pulumi(t *testing.T)             { runAllScenariosForComponent(t, "kubernetesistiobasecrds", "pulumi") }

// ─── Tier 4 Terraform (operators, addons) ───────────────────────────────────

func TestKubernetesZalandoPostgresOperator_Terraform(t *testing.T)    { runAllScenariosForComponent(t, "kuberneteszalandopostgresoperator", "terraform") }
func TestKubernetesStrimziKafkaOperator_Terraform(t *testing.T)       { runAllScenariosForComponent(t, "kubernetesstrimzikafkaoperator", "terraform") }
func TestKubernetesElasticOperator_Terraform(t *testing.T)            { runAllScenariosForComponent(t, "kuberneteselasticoperator", "terraform") }
func TestKubernetesAltinityOperator_Terraform(t *testing.T)           { runAllScenariosForComponent(t, "kubernetesaltinityoperator", "terraform") }
func TestKubernetesGatewayApiCrds_Terraform(t *testing.T)             { runAllScenariosForComponent(t, "kubernetesgatewayapicrds", "terraform") }
func TestKubernetesGhaRunnerScaleSetController_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetesgharunnerscalesetcontroller", "terraform") }
func TestKubernetesRookCephOperator_Terraform(t *testing.T)           { runAllScenariosForComponent(t, "kubernetesrookcephoperator", "terraform") }
func TestKubernetesExternalSecrets_Terraform(t *testing.T)            { runAllScenariosForComponent(t, "kubernetesexternalsecrets", "terraform") }
func TestKubernetesTekton_Terraform(t *testing.T)                     { runAllScenariosForComponent(t, "kubernetestekton", "terraform") }
func TestKubernetesIstioBaseCrds_Terraform(t *testing.T)             { runAllScenariosForComponent(t, "kubernetesistiobasecrds", "terraform") }

// ─── Gateway API Pulumi (854-860) ───────────────────────────────────────────
// Each kind declares KubernetesGatewayApiCrds as a registry prerequisite, which
// the harness installs before the scenario applies. Verification asserts the CR
// exists (controller-free: applies succeed once the CRDs are present).

func TestKubernetesGatewayClass_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "kubernetesgatewayclass", "pulumi") }
func TestKubernetesGateway_Pulumi(t *testing.T)         { runAllScenariosForComponent(t, "kubernetesgateway", "pulumi") }
func TestKubernetesHttpRoute_Pulumi(t *testing.T)       { runAllScenariosForComponent(t, "kuberneteshttproute", "pulumi") }
func TestKubernetesGrpcRoute_Pulumi(t *testing.T)       { runAllScenariosForComponent(t, "kubernetesgrpcroute", "pulumi") }
func TestKubernetesTcpRoute_Pulumi(t *testing.T)        { runAllScenariosForComponent(t, "kubernetestcproute", "pulumi") }
func TestKubernetesTlsRoute_Pulumi(t *testing.T)        { runAllScenariosForComponent(t, "kubernetestlsroute", "pulumi") }
func TestKubernetesReferenceGrant_Pulumi(t *testing.T)  { runAllScenariosForComponent(t, "kubernetesreferencegrant", "pulumi") }

// ─── Gateway API Terraform (854-860) ────────────────────────────────────────

func TestKubernetesGatewayClass_Terraform(t *testing.T)    { runAllScenariosForComponent(t, "kubernetesgatewayclass", "terraform") }
func TestKubernetesGateway_Terraform(t *testing.T)         { runAllScenariosForComponent(t, "kubernetesgateway", "terraform") }
func TestKubernetesHttpRoute_Terraform(t *testing.T)       { runAllScenariosForComponent(t, "kuberneteshttproute", "terraform") }
func TestKubernetesGrpcRoute_Terraform(t *testing.T)       { runAllScenariosForComponent(t, "kubernetesgrpcroute", "terraform") }
func TestKubernetesTcpRoute_Terraform(t *testing.T)        { runAllScenariosForComponent(t, "kubernetestcproute", "terraform") }
func TestKubernetesTlsRoute_Terraform(t *testing.T)        { runAllScenariosForComponent(t, "kubernetestlsroute", "terraform") }
func TestKubernetesReferenceGrant_Terraform(t *testing.T)  { runAllScenariosForComponent(t, "kubernetesreferencegrant", "terraform") }

// ─── Istio API Pulumi (861-867) ─────────────────────────────────────────────
// Each kind declares KubernetesIstioBaseCrds as a registry prerequisite, which
// the harness installs (istio/base CRDs, no istiod) before the scenario applies.
// Verification asserts the typed Istio CR exists.

func TestKubernetesPeerAuthentication_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetespeerauthentication", "pulumi") }

func TestKubernetesRequestAuthentication_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetesrequestauthentication", "pulumi") }

func TestKubernetesAuthorizationPolicy_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetesauthorizationpolicy", "pulumi") }

func TestKubernetesServiceEntry_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetesserviceentry", "pulumi") }

func TestKubernetesDestinationRule_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetesdestinationrule", "pulumi") }

func TestKubernetesEnvoyFilter_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetesenvoyfilter", "pulumi") }

func TestKubernetesTelemetry_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "kubernetestelemetry", "pulumi") }

// ─── Istio API Terraform (861-867) ──────────────────────────────────────────

func TestKubernetesPeerAuthentication_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetespeerauthentication", "terraform") }

func TestKubernetesRequestAuthentication_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetesrequestauthentication", "terraform") }

func TestKubernetesAuthorizationPolicy_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetesauthorizationpolicy", "terraform") }

func TestKubernetesServiceEntry_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetesserviceentry", "terraform") }

func TestKubernetesDestinationRule_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetesdestinationrule", "terraform") }

func TestKubernetesEnvoyFilter_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetesenvoyfilter", "terraform") }

func TestKubernetesTelemetry_Terraform(t *testing.T) { runAllScenariosForComponent(t, "kubernetestelemetry", "terraform") }

// runAllScenariosForComponent discovers and runs all E2E scenarios for a component
// using the specified IaC engine ("pulumi" or "terraform").
func runAllScenariosForComponent(t *testing.T, component, engine string) {
	t.Helper()

	var moduleDir string
	switch engine {
	case "pulumi":
		moduleDir = filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", "kubernetes", component, "v1", "iac", "pulumi")
	case "terraform":
		moduleDir = filepath.Join(repoRoot, "apis", "org", "openmcf", "provider", "kubernetes", component, "v1", "iac", "tf")
	default:
		t.Fatalf("unsupported engine: %s", engine)
	}

	if !fileExists(moduleDir) {
		t.Skipf("component %s %s module not found at %s", component, engine, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "kubernetes", component)
	if err != nil {
		t.Fatalf("failed to discover test scenarios for %s: %v", component, err)
	}

	if len(scenarios) == 0 {
		t.Skipf("no test scenarios found for %s in %s/v1/e2e/", component, component)
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
		Provider:     "kubernetes",
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

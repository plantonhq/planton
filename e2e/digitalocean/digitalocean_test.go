//go:build e2e

// Package digitalocean contains end-to-end tests that provision real
// DigitalOcean resources via Planton IaC modules and verify them through the
// DigitalOcean REST API. Credentials come from the environment
// (DIGITALOCEAN_TOKEN, plus SPACES_ACCESS_KEY_ID / SPACES_SECRET_ACCESS_KEY
// for bucket lanes -- never a stored secret); see the aa_e2e harness package.
//
// Run with: go test -tags=e2e -timeout=60m -v ./e2e/digitalocean/...
package digitalocean

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	digitaloceane2e "github.com/plantonhq/planton/catalog/digitalocean/aa_e2e"
	"github.com/plantonhq/planton/e2e/framework/discovery"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	profilepkg "github.com/plantonhq/planton/pkg/e2e/profile"
	componentv1 "github.com/plantonhq/planton/qa/componente2eprofile/v1"
)

var (
	testHarness      *digitaloceane2e.Harness
	repoRoot         string
	runID            string
	pulumiBackendURL string
	// assertApplyIdempotency mirrors the provider profile's
	// assert_apply_idempotency switch into every scenario's test context. The
	// profile is the single place the gate is armed; a test file that does
	// not read it leaves the switch inert, and every lane silently skips the
	// IDEMPOTENCY phase while the profile claims it runs.
	assertApplyIdempotency bool
)

func TestMain(m *testing.M) {
	var err error
	repoRoot, err = filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve repo root: %v\n", err)
		os.Exit(1)
	}

	runID = uuid.New().String()[:8]

	backendDir, err := os.MkdirTemp("", "planton-e2e-digitalocean-pulumi-*")
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

	providerProfile, err := profilepkg.LoadProviderProfile(repoRoot, "digitalocean")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load DigitalOcean provider E2E profile: %v\n", err)
		os.Exit(1)
	}
	assertApplyIdempotency = providerProfile.GetSpec().GetAssertApplyIdempotency()

	testHarness = digitaloceane2e.NewHarness()
	ctx := context.Background()
	if err := testHarness.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup DigitalOcean harness: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testHarness.Teardown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to teardown DigitalOcean harness: %v\n", err)
	}

	os.Exit(code)
}

// --- DigitalOcean VPC (root of the FK graph; the shared network fixture) ---

func TestDigitalOceanVpc_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanvpc", "pulumi")
}
func TestDigitalOceanVpc_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanvpc", "terraform")
}

// --- DigitalOcean Droplet (composed topology: deploys the Vpc prerequisite) ---

func TestDigitalOceanDroplet_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandroplet", "pulumi")
}
func TestDigitalOceanDroplet_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandroplet", "terraform")
}

// --- DigitalOcean Volume (standalone block storage) ---

func TestDigitalOceanVolume_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanvolume", "pulumi")
}
func TestDigitalOceanVolume_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanvolume", "terraform")
}

// --- DigitalOcean Firewall (standalone; droplet attachment arms ride scenarios) ---

func TestDigitalOceanFirewall_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanfirewall", "pulumi")
}
func TestDigitalOceanFirewall_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanfirewall", "terraform")
}

// --- DigitalOcean Load Balancer (composed topology: deploys the Vpc prerequisite) ---

func TestDigitalOceanLoadBalancer_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanloadbalancer", "pulumi")
}
func TestDigitalOceanLoadBalancer_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanloadbalancer", "terraform")
}

// --- DigitalOcean Database Cluster (slow lane: ~5 min creates, billed hourly) ---

func TestDigitalOceanDatabaseCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasecluster", "pulumi")
}
func TestDigitalOceanDatabaseCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasecluster", "terraform")
}

// --- DigitalOcean Database User (satellite: rides the cluster prerequisite) ---

func TestDigitalOceanDatabaseUser_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabaseuser", "pulumi")
}
func TestDigitalOceanDatabaseUser_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabaseuser", "terraform")
}

// --- DigitalOcean Database Db (satellite: rides the cluster prerequisite) ---

func TestDigitalOceanDatabaseDb_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasedb", "pulumi")
}
func TestDigitalOceanDatabaseDb_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasedb", "terraform")
}

// --- DigitalOcean Database Connection Pool (satellite: rides the cluster prerequisite) ---

func TestDigitalOceanDatabaseConnectionPool_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabaseconnectionpool", "pulumi")
}
func TestDigitalOceanDatabaseConnectionPool_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabaseconnectionpool", "terraform")
}

// --- DigitalOcean Database Firewall (satellite: rides the cluster prerequisite) ---

func TestDigitalOceanDatabaseFirewall_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasefirewall", "pulumi")
}
func TestDigitalOceanDatabaseFirewall_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasefirewall", "terraform")
}

// --- DigitalOcean Database Replica (satellite; its own billing class: a second cluster-sized node) ---

func TestDigitalOceanDatabaseReplica_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasereplica", "pulumi")
}
func TestDigitalOceanDatabaseReplica_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasereplica", "terraform")
}

// --- DigitalOcean Kubernetes Cluster (slow lane: ~5-10 min creates; Vpc prerequisite) ---

func TestDigitalOceanKubernetesCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceankubernetescluster", "pulumi")
}
func TestDigitalOceanKubernetesCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceankubernetescluster", "terraform")
}

// --- DigitalOcean Kubernetes Node Pool (composed topology: deploys the cluster prerequisite) ---

func TestDigitalOceanKubernetesNodePool_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceankubernetesnodepool", "pulumi")
}
func TestDigitalOceanKubernetesNodePool_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceankubernetesnodepool", "terraform")
}

// --- DigitalOcean DNS Zone (a domain; cheap and instant) ---

func TestDigitalOceanDnsZone_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandnszone", "pulumi")
}
func TestDigitalOceanDnsZone_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandnszone", "terraform")
}

// --- DigitalOcean DNS Record (composed topology: deploys the DnsZone prerequisite) ---

func TestDigitalOceanDnsRecord_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandnsrecord", "pulumi")
}
func TestDigitalOceanDnsRecord_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandnsrecord", "terraform")
}

// --- DigitalOcean Certificate (lets_encrypt arms need a delegated domain -- environmental gate) ---

func TestDigitalOceanCertificate_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceancertificate", "pulumi")
}
func TestDigitalOceanCertificate_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceancertificate", "terraform")
}

// --- DigitalOcean Container Registry (one per account -- lanes must not run concurrently) ---

func TestDigitalOceanContainerRegistry_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceancontainerregistry", "pulumi")
}
func TestDigitalOceanContainerRegistry_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceancontainerregistry", "terraform")
}

// --- DigitalOcean Bucket (Spaces; needs SPACES_ACCESS_KEY_ID / SPACES_SECRET_ACCESS_KEY) ---

func TestDigitalOceanBucket_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanbucket", "pulumi")
}
func TestDigitalOceanBucket_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanbucket", "terraform")
}

// --- DigitalOcean App (git-source deploys run a real build -- slow lane) ---

func TestDigitalOceanApp_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanapp", "pulumi")
}
func TestDigitalOceanApp_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanapp", "terraform")
}

// --- DigitalOcean Function (deploys an App Platform app carrying a functions section) ---

func TestDigitalOceanFunction_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanfunction", "pulumi")
}
func TestDigitalOceanFunction_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanfunction", "terraform")
}

// --- DigitalOcean Project ---

func TestDigitalOceanProject_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanproject", "pulumi")
}
func TestDigitalOceanProject_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanproject", "terraform")
}

// --- DigitalOcean SSH Key ---

func TestDigitalOceanSshKey_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceansshkey", "pulumi")
}
func TestDigitalOceanSshKey_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceansshkey", "terraform")
}

// --- DigitalOcean Monitor Alert ---

func TestDigitalOceanMonitorAlert_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanmonitoralert", "pulumi")
}
func TestDigitalOceanMonitorAlert_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanmonitoralert", "terraform")
}

// --- DigitalOcean Uptime Check (composes its alert rules) ---

func TestDigitalOceanUptimeCheck_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanuptimecheck", "pulumi")
}
func TestDigitalOceanUptimeCheck_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanuptimecheck", "terraform")
}

// --- DigitalOcean Database Kafka Topic (shared Kafka fixture; run serially with the schema kind) ---

func TestDigitalOceanDatabaseKafkaTopic_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasekafkatopic", "pulumi")
}
func TestDigitalOceanDatabaseKafkaTopic_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasekafkatopic", "terraform")
}

// --- DigitalOcean Database Kafka Schema (shared Kafka fixture; run serially with the topic kind) ---

func TestDigitalOceanDatabaseKafkaSchema_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasekafkaschema", "pulumi")
}
func TestDigitalOceanDatabaseKafkaSchema_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandatabasekafkaschema", "terraform")
}

// --- DigitalOcean Reserved IP (unassigned IPv4 BILLS -- zero-orphan sweep is doubly binding) ---

func TestDigitalOceanReservedIp_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanreservedip", "pulumi")
}
func TestDigitalOceanReservedIp_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanreservedip", "terraform")
}

// --- DigitalOcean VPC Peering (dual VPC fixtures via the scenario's e2e-prerequisites annotation) ---

func TestDigitalOceanVpcPeering_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanvpcpeering", "pulumi")
}
func TestDigitalOceanVpcPeering_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanvpcpeering", "terraform")
}

// --- DigitalOcean Spaces Key (per-bucket scenario rides the Bucket fixture; the bucket lane needs the Spaces key pair env) ---

func TestDigitalOceanSpacesKey_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanspaceskey", "pulumi")
}
func TestDigitalOceanSpacesKey_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceanspaceskey", "terraform")
}

// --- DigitalOcean CDN (bucket-origin fixture via the scenario's e2e-prerequisites annotation) ---

func TestDigitalOceanCdn_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceancdn", "pulumi")
}
func TestDigitalOceanCdn_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceancdn", "terraform")
}

// --- DigitalOcean Droplet Autoscale Pool (SshKey registry prerequisite; member droplets BILL and destroy destroys them) ---

func TestDigitalOceanDropletAutoscalePool_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandropletautoscalepool", "pulumi")
}
func TestDigitalOceanDropletAutoscalePool_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "digitaloceandropletautoscalepool", "terraform")
}

// runAllScenariosForComponent discovers and runs all E2E scenarios for a
// DigitalOcean component.
func runAllScenariosForComponent(t *testing.T, component, engine string) {
	t.Helper()

	if cp, err := profilepkg.LoadComponentProfile(repoRoot, "digitalocean", component); err == nil && cp.Spec != nil {
		switch cp.Spec.Status {
		case componentv1.ComponentE2EProfileSpec_deferred,
			componentv1.ComponentE2EProfileSpec_skip,
			componentv1.ComponentE2EProfileSpec_stub,
			// pending_proof: fully authored, offline-validated, awaiting its
			// first live proof. The proving session flips the profile to green
			// immediately before executing the lanes; until then a sweep must
			// never run it.
			componentv1.ComponentE2EProfileSpec_pending_proof:
			reason := cp.Spec.DeferredReason
			if reason == "" {
				reason = cp.Spec.Status.String()
			}
			t.Skipf("component %s E2E profile status is %s: %s", component, cp.Spec.Status, reason)
		}
	}

	moduleDir, err := discovery.ModuleDir(repoRoot, "digitalocean", component, engine)
	if err != nil {
		t.Fatalf("failed to locate %s %s module: %v", component, engine, err)
	}

	if !fileExists(moduleDir) {
		t.Skipf("component %s %s module not found at %s", component, engine, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "digitalocean", component)
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

	// Scenarios needing owner-arranged external context (the
	// e2e-required-env annotation -- for DigitalOcean, typically a domain
	// delegated to the account's DNS, injected as
	// ${E2E_ENV:PLANTON_E2E_DIGITALOCEAN_DELEGATED_DOMAIN} for the Let's
	// Encrypt certificate arm) skip honestly where the environment does not
	// carry the arrangement -- unset tokens would otherwise fail expansion
	// loudly, turning a recorded deferral into a false failure.
	if missing, err := runner.ScenarioMissingRequiredEnv(scenario.ManifestPath); err != nil {
		t.Fatalf("reading required-env declaration for scenario %s/%s: %v", component, scenario.Name, err)
	} else if len(missing) > 0 {
		t.Skipf("scenario %s/%s needs owner-arranged environment variables that are unset: %s (per %s)",
			component, scenario.Name, strings.Join(missing, ", "), runner.ScenarioRequiredEnvAnnotation)
	}

	tc := &provider.ComponentTestContext{
		Component: component,
		// The provider is ALWAYS the catalog directory slug. Deriving it from
		// the registry enum would yield "digital_ocean", and the import
		// round-trip's map lookup would silently skip instead of running.
		Provider:     "digitalocean",
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
		BackendURL:             pulumiBackendURL,
		AssertApplyIdempotency: assertApplyIdempotency,
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

//go:build e2e

package stripe

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	stripee2e "github.com/plantonhq/planton/catalog/stripe/aa_e2e"
	"github.com/plantonhq/planton/e2e/framework/discovery"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	profilepkg "github.com/plantonhq/planton/pkg/e2e/profile"
	kindv1 "github.com/plantonhq/planton/qa/catalogkinde2eprofile/v1"
)

// tofuEngine is the framework's engine for every HCL lane. It runs the tofu binary by default,
// and the runner refuses any other binary for a kind that declares OpenTofu only.
const tofuEngine = "terraform"

var (
	testHarness            *stripee2e.Harness
	repoRoot               string
	runID                  string
	assertApplyIdempotency bool
)

// TestMain proves the key belongs to the dedicated test sandbox before any lane runs: the
// harness's Setup refuses a live key, and a refusal exits here. Stripe kinds run on OpenTofu
// only, and so do their prerequisites (each deploys on its own kind's engine), so no Pulumi
// backend is prepared.
func TestMain(m *testing.M) {
	var err error
	repoRoot, err = filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve repo root: %v\n", err)
		os.Exit(1)
	}

	runID = uuid.New().String()[:8]

	providerProfile, err := profilepkg.LoadProviderProfile(repoRoot, "stripe")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load Stripe provider E2E profile: %v\n", err)
		os.Exit(1)
	}
	assertApplyIdempotency = providerProfile.GetSpec().GetAssertApplyIdempotency()

	testHarness = stripee2e.NewHarness()
	ctx := context.Background()
	if err := testHarness.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up the Stripe harness: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testHarness.Teardown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to tear down the Stripe harness: %v\n", err)
	}

	os.Exit(code)
}

// One entrypoint per Stripe kind, named Test + registry kind name + _Tofu: the matrix builder
// derives the suffix from the profile's validated_provisioners, which for these kinds is tofu.

func TestStripeWebhookEndpoint_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripewebhookendpoint")
}

func TestStripeBillingPortalConfiguration_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripebillingportalconfiguration")
}

func TestStripePaymentMethodConfiguration_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripepaymentmethodconfiguration")
}

func TestStripeEventDestination_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripeeventdestination")
}

func TestStripePaymentMethodDomain_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripepaymentmethoddomain")
}

func TestStripeRadarValueList_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "striperadarvaluelist")
}

func TestStripeProduct_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripeproduct")
}

func TestStripePrice_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripeprice")
}

func TestStripeEntitlementFeature_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripeentitlementfeature")
}

func TestStripeCoupon_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripecoupon")
}

func TestStripePromotionCode_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripepromotioncode")
}

func TestStripeShippingRate_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripeshippingrate")
}

func TestStripeTaxRate_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripetaxrate")
}

func TestStripeTaxRegistration_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripetaxregistration")
}

func TestStripeBillingMeter_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripebillingmeter")
}

func TestStripePaymentLink_Tofu(t *testing.T) {
	runAllScenariosForKind(t, "stripepaymentlink")
}

// runAllScenariosForKind discovers and runs every E2E scenario of a Stripe kind.
func runAllScenariosForKind(t *testing.T, kindDir string) {
	t.Helper()

	if cp, err := profilepkg.LoadKindProfile(repoRoot, "stripe", kindDir); err == nil && cp.Spec != nil {
		switch cp.Spec.Status {
		case kindv1.CatalogKindE2EProfileSpec_deferred,
			kindv1.CatalogKindE2EProfileSpec_skip,
			kindv1.CatalogKindE2EProfileSpec_stub,
			// pending_proof: fully authored, offline-validated, awaiting its first live proof.
			// The proving session flips the profile to green immediately before running the
			// lanes; until then a sweep must never run it.
			kindv1.CatalogKindE2EProfileSpec_pending_proof,
			// real_cluster has no meaning for a SaaS provider; a profile carrying it is a
			// mistake that must skip loudly, never run.
			kindv1.CatalogKindE2EProfileSpec_real_cluster:
			reason := cp.Spec.DeferredReason
			if reason == "" {
				reason = cp.Spec.Status.String()
			}
			t.Skipf("kind %s E2E profile status is %s: %s", kindDir, cp.Spec.Status, reason)
		}
	}

	moduleDir, err := discovery.ModuleDir(repoRoot, "stripe", kindDir, tofuEngine)
	if err != nil {
		t.Fatalf("failed to locate the %s OpenTofu module: %v", kindDir, err)
	}
	if !fileExists(moduleDir) {
		t.Skipf("kind %s OpenTofu module not found at %s", kindDir, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "stripe", kindDir)
	if err != nil {
		t.Fatalf("failed to discover test scenarios for %s: %v", kindDir, err)
	}
	if len(scenarios) == 0 {
		t.Skipf("no test scenarios found for %s", kindDir)
	}

	t.Logf("Discovered %d scenarios for %s [tofu]", len(scenarios), kindDir)
	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Name, func(t *testing.T) {
			runSingleScenario(t, kindDir, moduleDir, scenario)
		})
	}
}

func runSingleScenario(t *testing.T, kindDir, moduleDir string, scenario discovery.TestScenario) {
	t.Helper()

	// A scenario that needs an owner-arranged sandbox object declares it with the required-env
	// annotation, and skips honestly where the environment does not carry it.
	if missing, err := runner.ScenarioMissingRequiredEnv(scenario.ManifestPath); err != nil {
		t.Fatalf("reading required-env declaration for scenario %s/%s: %v", kindDir, scenario.Name, err)
	} else if len(missing) > 0 {
		t.Skipf("scenario %s/%s needs owner-arranged environment variables that are unset: %s (per %s)",
			kindDir, scenario.Name, strings.Join(missing, ", "), runner.ScenarioRequiredEnvAnnotation)
	}

	tc := &provider.KindTestContext{
		Kind:                   kindDir,
		Provider:               "stripe",
		Engine:                 tofuEngine,
		ModuleDir:              moduleDir,
		ManifestPath:           scenario.ManifestPath,
		RepoRoot:               repoRoot,
		RunID:                  runID,
		T:                      t,
		AssertApplyIdempotency: assertApplyIdempotency,
	}

	result := runner.RunKindTest(context.Background(), tc, testHarness)
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
		t.Fatalf("scenario %s/%s [tofu] failed (total: %s)", kindDir, scenario.Name, result.Duration)
	}
	t.Logf("scenario %s/%s [tofu] passed (total: %s)", kindDir, scenario.Name, result.Duration)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

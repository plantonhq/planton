package costprofile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	costprofilev1 "github.com/plantonhq/planton/finops/catalogkindcostprofile/v1"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/specpath"
)

// TestCostProfileConformance holds every authored cost profile to its
// contract, offline:
//
//  1. The profile parses strictly against its proto schema and names its
//     kind (metadata.name equals the kind directory).
//  2. The billing model is declared, and the profile's shape matches it:
//     always-on and hybrid models declare their baseline charges; usage-
//     driven models (usage_based, hybrid, cluster_capacity) state their
//     estimate exclusions -- an estimate that hides what it cannot know is
//     a lie with a dollar sign; free kinds declare no drivers.
//  3. Every cost driver names the spec field that moves the bill, and that
//     field path resolves against the served version's compiled descriptors
//     -- a schema rename that orphans a driver fails CI loudly.
//  4. Charge-identity fields (service_name, sku_meter, pricing_unit) are
//     present where the FOCUS-aligned vocabulary requires them, so
//     estimates stay joinable against real billing data.
func TestCostProfileConformance(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}

	discovered, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering cost profiles: %v", err)
	}
	if len(discovered) == 0 {
		t.Skip("no cost profiles authored yet")
	}

	for provider, kindDirs := range discovered {
		for _, kindDir := range kindDirs {
			kindDir := kindDir
			t.Run(provider+"/"+kindDir, func(t *testing.T) {
				profile, err := Load(root, provider, kindDir)
				if err != nil {
					t.Fatalf("cost profile: %v", err)
				}
				if profile.GetKind() != "CatalogKindCostProfile" {
					t.Fatalf("kind is %q, want CatalogKindCostProfile", profile.GetKind())
				}
				if profile.GetMetadata().GetName() != kindDir {
					t.Errorf("metadata.name is %q, want %q", profile.GetMetadata().GetName(), kindDir)
				}

				spec := profile.GetSpec()
				model := spec.GetBillingModel()
				if model == costprofilev1.BillingModel_billing_model_unspecified {
					t.Error("billing_model is unspecified -- the profile's central claim is missing")
				}

				switch model {
				case costprofilev1.BillingModel_always_on, costprofilev1.BillingModel_hybrid:
					if len(spec.GetBaselineCharges()) == 0 {
						t.Errorf("billing_model %s declares no baseline_charges -- what accrues while the resource exists?", model)
					}
				case costprofilev1.BillingModel_free, costprofilev1.BillingModel_cluster_capacity:
					if len(spec.GetBaselineCharges()) > 0 {
						t.Errorf("billing_model %s must not declare baseline_charges -- no cloud SKU bills for the resource directly", model)
					}
				}
				switch model {
				case costprofilev1.BillingModel_usage_based, costprofilev1.BillingModel_hybrid, costprofilev1.BillingModel_cluster_capacity:
					if len(spec.GetEstimateExclusions()) == 0 {
						t.Errorf("billing_model %s states no estimate_exclusions -- usage-driven estimates must state what they cannot know", model)
					}
				case costprofilev1.BillingModel_free:
					if len(spec.GetCostDrivers()) > 0 {
						t.Error("billing_model free must not declare cost_drivers")
					}
				}

				for _, charge := range spec.GetBaselineCharges() {
					if strings.TrimSpace(charge.GetServiceName()) == "" ||
						strings.TrimSpace(charge.GetSkuMeter()) == "" ||
						strings.TrimSpace(charge.GetPricingUnit()) == "" ||
						strings.TrimSpace(charge.GetDescription()) == "" {
						t.Errorf("baseline charge %+v is missing identity fields (service_name, sku_meter, pricing_unit, description are all required)", charge)
					}
				}

				specDescriptor := kindSpecDescriptor(t, kindDir)
				for _, driver := range spec.GetCostDrivers() {
					if err := specpath.Validate(specDescriptor, driver.GetFieldPath()); err != nil {
						t.Errorf("cost driver field_path %q: %v", driver.GetFieldPath(), err)
					}
					if strings.TrimSpace(driver.GetSkuMeter()) == "" {
						t.Errorf("cost driver %q has no sku_meter -- the estimator has nothing to price", driver.GetFieldPath())
					}
					if strings.TrimSpace(driver.GetImpact()) == "" {
						t.Errorf("cost driver %q has no impact sentence -- agents relay this to users verbatim", driver.GetFieldPath())
					}
				}
			})
		}
	}
}

// kindSpecDescriptor resolves a kind directory name to its kind's spec
// message descriptor via the kind registry.
func kindSpecDescriptor(t *testing.T, kindDir string) protoreflect.MessageDescriptor {
	t.Helper()
	kind := catalogkindreflect.KindFromString(kindDir)
	apiMessage, err := catalogkindreflect.NewInstance(kind)
	if err != nil {
		t.Fatalf("NewInstance(%s): %v", kindDir, err)
	}
	return specDescriptor(t, apiMessage)
}

// specDescriptor returns the descriptor of the kind's spec message (the api
// envelope's `spec` field).
func specDescriptor(t *testing.T, apiMessage proto.Message) protoreflect.MessageDescriptor {
	t.Helper()
	specField := apiMessage.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specField == nil || specField.Kind() != protoreflect.MessageKind {
		t.Fatalf("%s has no spec message field", apiMessage.ProtoReflect().Descriptor().FullName())
	}
	return specField.Message()
}

// repoRoot walks up from this test file to the directory containing go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test file")
		}
		dir = parent
	}
}

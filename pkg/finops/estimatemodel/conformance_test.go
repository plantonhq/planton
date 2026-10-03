package estimatemodel

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	estimatemodelv1 "github.com/plantonhq/planton/finops/catalogkindcostestimatemodel/v1"
	costprofilev1 "github.com/plantonhq/planton/finops/catalogkindcostprofile/v1"
	"github.com/plantonhq/planton/pkg/finops/costprofile"
)

var decimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// TestEstimateModelConformance holds every estimate model to its contract,
// offline:
//
//  1. The document parses strictly, names its kind (metadata.name
//     equals the filename), and the kind ships a cost profile -- a
//     model without a cost anatomy to stand on is unreviewable.
//  2. Every preset key resolves to an actual preset file, exactly once.
//  3. Every quantity line names a declared meter (a baseline charge or
//     cost driver sku_meter in the kind's cost.yaml), a price-book
//     entry reference, a plain-decimal quantity, and its quantity_basis
//     prose -- the audit trail back to the preset's values.
//  4. Cluster-capacity components state a capacity footprint and no
//     quantity lines; monetary presets pin region, currency, and the
//     hours-per-month convention, and state their exclusions.
//
// Whether price references resolve (and agree with the cost profile on
// units) is the estimate generator's cross-artifact check -- this gate
// keeps each model internally sound.
func TestEstimateModelConformance(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}

	kindDirs, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering estimate models: %v", err)
	}
	if len(kindDirs) == 0 {
		t.Skip("no estimate models authored yet")
	}

	for _, kindDir := range kindDirs {
		kindDir := kindDir
		t.Run(kindDir, func(t *testing.T) {
			model, err := Load(root, kindDir)
			if err != nil {
				t.Fatalf("estimate model: %v", err)
			}
			if model.GetKind() != "CatalogKindCostEstimateModel" {
				t.Fatalf("kind is %q, want CatalogKindCostEstimateModel", model.GetKind())
			}
			if model.GetMetadata().GetName() != kindDir {
				t.Errorf("metadata.name is %q, want %q (the filename is the kind's identity)",
					model.GetMetadata().GetName(), kindDir)
			}

			provider := kindProvider(t, root, kindDir)
			profile, err := costprofile.Load(root, provider, kindDir)
			if err != nil {
				t.Fatalf("the modeled kind must ship a cost profile: %v", err)
			}
			meters := declaredMeters(profile)
			clusterCapacity := profile.GetSpec().GetBillingModel() == costprofilev1.BillingModel_cluster_capacity

			if len(model.GetSpec().GetPresets()) == 0 {
				t.Fatal("estimate model declares no presets")
			}
			seen := map[string]bool{}
			for _, preset := range model.GetSpec().GetPresets() {
				preset := preset
				key := preset.GetPreset()
				if seen[key] {
					t.Errorf("preset %q modeled more than once", key)
				}
				seen[key] = true
				t.Run(key, func(t *testing.T) {
					presetPath := filepath.Join(root, "catalog", provider, kindDir, "presets", key+".yaml")
					if _, err := os.Stat(presetPath); err != nil {
						t.Fatalf("preset key %q resolves to no preset file (%s)", key, presetPath)
					}
					if len(preset.GetExclusions()) == 0 {
						t.Error("no exclusions stated -- an estimate that hides what it cannot know is a lie with a dollar sign")
					}
					if clusterCapacity {
						checkCapacityPreset(t, preset)
					} else {
						checkMonetaryPreset(t, preset, meters)
					}
				})
			}
		})
	}
}

// checkMonetaryPreset verifies a priced preset's model: pinned assumptions
// and complete, meter-bound quantity lines. A preset with no lines is
// legitimate -- it models no committed always-on spend and estimates to
// zero.
func checkMonetaryPreset(t *testing.T, preset *estimatemodelv1.PresetEstimateModel, meters map[string]bool) {
	t.Helper()
	if preset.GetCapacityFootprint() != nil {
		t.Error("capacity_footprint is for cluster-capacity kinds; priced presets carry quantity lines")
	}
	if strings.TrimSpace(preset.GetRegionAssumption()) == "" {
		t.Error("region_assumption is empty -- list prices vary by region, so an unpinned estimate is not reproducible")
	}
	if strings.TrimSpace(preset.GetCurrency()) == "" {
		t.Error("currency is empty")
	}
	if preset.GetHoursPerMonth() <= 0 {
		t.Error("hours_per_month must be declared (730 is the industry convention)")
	}
	for _, line := range preset.GetQuantityLines() {
		meter := strings.TrimSpace(line.GetSkuMeter())
		if meter == "" {
			t.Error("quantity line has no sku_meter")
		} else if !meters[meter] {
			t.Errorf("sku_meter %q is not declared by the kind's cost.yaml (baseline charges and cost drivers) -- a model cannot price an undeclared meter", meter)
		}
		if strings.TrimSpace(line.GetPrice()) == "" {
			t.Errorf("line %q has no price reference -- name the price-book entry that prices it", meter)
		}
		if !decimalPattern.MatchString(line.GetPricingQuantity()) {
			t.Errorf("line %q pricing_quantity %q is not a plain decimal string", meter, line.GetPricingQuantity())
		}
		if strings.TrimSpace(line.GetQuantityBasis()) == "" {
			t.Errorf("line %q has no quantity_basis -- the audit trail back to the preset's values", meter)
		}
	}
}

// checkCapacityPreset verifies a cluster-capacity preset's model: a
// capacity footprint with its derivation, and no priced lines or
// monetary assumptions anywhere.
func checkCapacityPreset(t *testing.T, preset *estimatemodelv1.PresetEstimateModel) {
	t.Helper()
	if len(preset.GetQuantityLines()) > 0 {
		t.Error("cluster-capacity kinds carry no quantity lines -- their price is the target cluster's economics")
	}
	if preset.GetRegionAssumption() != "" || preset.GetCurrency() != "" || preset.GetHoursPerMonth() != 0 {
		t.Error("cluster-capacity presets price nothing regional -- drop region_assumption, currency, and hours_per_month")
	}
	footprint := preset.GetCapacityFootprint()
	if footprint == nil {
		t.Fatal("cluster-capacity preset has no capacity_footprint -- capacity is the honest statement")
	}
	if strings.TrimSpace(footprint.GetBasis()) == "" {
		t.Error("capacity_footprint has no basis -- the audit trail back to the preset's values")
	}
	if strings.TrimSpace(footprint.GetCpuRequests()) == "" &&
		strings.TrimSpace(footprint.GetMemoryRequests()) == "" &&
		strings.TrimSpace(footprint.GetPersistentStorage()) == "" {
		t.Error("capacity_footprint states no capacity at all")
	}
}

// declaredMeters collects the sku_meter vocabulary a kind's cost
// profile declares across baseline charges and cost drivers.
func declaredMeters(profile *costprofilev1.CatalogKindCostProfile) map[string]bool {
	meters := map[string]bool{}
	for _, charge := range profile.GetSpec().GetBaselineCharges() {
		meters[strings.TrimSpace(charge.GetSkuMeter())] = true
	}
	for _, driver := range profile.GetSpec().GetCostDrivers() {
		meters[strings.TrimSpace(driver.GetSkuMeter())] = true
	}
	return meters
}

// kindProvider locates the provider directory a kind lives under.
func kindProvider(t *testing.T, repoRoot, kindDir string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(repoRoot, "catalog", "*", kindDir))
	if err != nil || len(matches) == 0 {
		t.Fatalf("estimate model names kind %q, which exists nowhere under catalog/", kindDir)
	}
	if len(matches) > 1 {
		t.Fatalf("kind %q exists under multiple providers: %v", kindDir, matches)
	}
	return filepath.Base(filepath.Dir(matches[0]))
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

package costestimate

import (
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	costestimatev1 "github.com/plantonhq/planton/finops/catalogkindcostestimate/v1"
	costprofilev1 "github.com/plantonhq/planton/finops/catalogkindcostprofile/v1"
	"github.com/plantonhq/planton/pkg/finops/costprofile"
)

var (
	decimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
	datePattern    = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
)

// TestCostEstimateConformance holds every authored estimate document to its
// contract, offline:
//
//  1. The document parses strictly, names its kind (metadata.name
//     equals the filename), and the kind ships a cost profile -- an
//     estimate without a cost anatomy to stand on is unreviewable.
//  2. Every preset key resolves to an actual preset file, exactly once.
//  3. Every priced meter is declared by the kind's cost.yaml (as a
//     baseline charge or cost driver sku_meter) -- an estimate cannot
//     price functionality the profile does not know about.
//  4. The arithmetic is re-computed exactly (big.Rat, no floats):
//     pricing_quantity x list_unit_price must equal list_cost on every
//     line, and the lines must sum to total_list_cost. A fudged or stale
//     figure fails CI.
//  5. Every unit price carries its source URL and retrieval date, every
//     monetary preset states its exclusions and pins region, currency,
//     and the hours-per-month convention, and lines are ordered largest
//     cost first.
//  6. Cluster-capacity components state a capacity footprint INSTEAD of
//     dollars -- their price is the target cluster's economics, and a
//     fabricated figure would be a lie with a dollar sign.
func TestCostEstimateConformance(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}

	kindDirs, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering cost estimates: %v", err)
	}
	if len(kindDirs) == 0 {
		t.Skip("no cost estimates authored yet")
	}

	for _, kindDir := range kindDirs {
		kindDir := kindDir
		t.Run(kindDir, func(t *testing.T) {
			estimate, err := Load(root, kindDir)
			if err != nil {
				t.Fatalf("cost estimate: %v", err)
			}
			if estimate.GetKind() != "CatalogKindCostEstimate" {
				t.Fatalf("kind is %q, want CatalogKindCostEstimate", estimate.GetKind())
			}
			if estimate.GetMetadata().GetName() != kindDir {
				t.Errorf("metadata.name is %q, want %q (the filename is the kind's identity)",
					estimate.GetMetadata().GetName(), kindDir)
			}

			provider := kindProvider(t, root, kindDir)
			profile, err := costprofile.Load(root, provider, kindDir)
			if err != nil {
				t.Fatalf("the estimated kind must ship a cost profile: %v", err)
			}
			meters := declaredMeters(profile)
			clusterCapacity := profile.GetSpec().GetBillingModel() == costprofilev1.BillingModel_cluster_capacity

			if len(estimate.GetSpec().GetPresets()) == 0 {
				t.Fatal("estimate document declares no presets")
			}
			seen := map[string]bool{}
			for _, preset := range estimate.GetSpec().GetPresets() {
				preset := preset
				key := preset.GetPreset()
				if seen[key] {
					t.Errorf("preset %q estimated more than once", key)
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

// checkMonetaryPreset verifies a priced preset: pinned assumptions, sourced
// and dated unit prices, meters bound to the cost profile, and exact
// arithmetic from line quantities up to the preset total.
func checkMonetaryPreset(t *testing.T, preset *costestimatev1.PresetEstimate, meters map[string]bool) {
	t.Helper()
	if preset.GetCapacityFootprint() != nil {
		t.Error("capacity_footprint is for cluster-capacity kinds; priced presets carry line items")
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

	total := parseDecimal(t, "total_list_cost", preset.GetTotalListCost())
	sum := new(big.Rat)
	previous := (*big.Rat)(nil)
	for _, line := range preset.GetLineItems() {
		if strings.TrimSpace(line.GetServiceName()) == "" {
			t.Errorf("line %q has no service_name", line.GetSkuMeter())
		}
		meter := strings.TrimSpace(line.GetSkuMeter())
		if meter == "" {
			t.Error("line has no sku_meter")
		} else if !meters[meter] {
			t.Errorf("sku_meter %q is not declared by the kind's cost.yaml (baseline charges and cost drivers) -- an estimate cannot price an undeclared meter", meter)
		}
		if strings.TrimSpace(line.GetPricingUnit()) == "" {
			t.Errorf("line %q has no pricing_unit", meter)
		}
		if strings.TrimSpace(line.GetQuantityBasis()) == "" {
			t.Errorf("line %q has no quantity_basis -- the audit trail back to the preset's values", meter)
		}
		if strings.TrimSpace(line.GetPriceSource()) == "" {
			t.Errorf("line %q has no price_source", meter)
		}
		if !datePattern.MatchString(line.GetRetrievedOn()) {
			t.Errorf("line %q retrieved_on %q is not a YYYY-MM-DD date -- a dated price is a fact, an undated price is a rumor", meter, line.GetRetrievedOn())
		}

		quantity := parseDecimal(t, meter+" pricing_quantity", line.GetPricingQuantity())
		unitPrice := parseDecimal(t, meter+" list_unit_price", line.GetListUnitPrice())
		cost := parseDecimal(t, meter+" list_cost", line.GetListCost())
		if quantity == nil || unitPrice == nil || cost == nil {
			return
		}
		product := new(big.Rat).Mul(quantity, unitPrice)
		if product.Cmp(cost) != 0 {
			t.Errorf("line %q: %s x %s = %s, but list_cost claims %s",
				meter, line.GetPricingQuantity(), line.GetListUnitPrice(),
				product.FloatString(6), line.GetListCost())
		}
		sum.Add(sum, cost)

		if previous != nil && cost.Cmp(previous) > 0 {
			t.Errorf("line %q is larger than the line above it -- order line items largest cost first", meter)
		}
		previous = cost
	}
	if total != nil && sum.Cmp(total) != 0 {
		t.Errorf("line items sum to %s, but total_list_cost claims %s", sum.FloatString(6), preset.GetTotalListCost())
	}
}

// checkCapacityPreset verifies a cluster-capacity preset: a capacity
// footprint with its derivation, and no dollar figures anywhere.
func checkCapacityPreset(t *testing.T, preset *costestimatev1.PresetEstimate) {
	t.Helper()
	if len(preset.GetLineItems()) > 0 || preset.GetTotalListCost() != "" {
		t.Error("cluster-capacity kinds carry no priced lines or totals -- their price is the target cluster's economics")
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
		t.Fatalf("estimate names kind %q, which exists nowhere under catalog/", kindDir)
	}
	if len(matches) > 1 {
		t.Fatalf("kind %q exists under multiple providers: %v", kindDir, matches)
	}
	return filepath.Base(filepath.Dir(matches[0]))
}

// parseDecimal parses a money/quantity decimal string into an exact
// rational, rejecting anything floats could have corrupted.
func parseDecimal(t *testing.T, field, value string) *big.Rat {
	t.Helper()
	if !decimalPattern.MatchString(value) {
		t.Errorf("%s %q is not a plain decimal string (money is never a YAML float)", field, value)
		return nil
	}
	rat, ok := new(big.Rat).SetString(value)
	if !ok {
		t.Errorf("%s %q does not parse as a decimal", field, value)
		return nil
	}
	return rat
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

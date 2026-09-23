package catalogbundle

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"

	controlprofilev1 "github.com/plantonhq/planton/compliance/componentcontrolprofile/v1"
	controlcatalogv1 "github.com/plantonhq/planton/compliance/controlcatalog/v1"
	frameworkcrosswalkv1 "github.com/plantonhq/planton/compliance/frameworkcrosswalk/v1"
	derivationv1 "github.com/plantonhq/planton/finops/componentcostderivation/v1"
	costestimatev1 "github.com/plantonhq/planton/finops/componentcostestimate/v1"
	costprofilev1 "github.com/plantonhq/planton/finops/componentcostprofile/v1"
	pricebookv1 "github.com/plantonhq/planton/finops/pricebook/v1"
	permissionsv1 "github.com/plantonhq/planton/iac/componentpermissions/v1"
	"github.com/plantonhq/planton/pkg/protobufyaml"
)

// Fact-sheet cargo is the bundle's verified cost/posture/permissions data:
// each covered component's cost profile, control profile, permission
// manifest, and (when priced) its generated per-preset estimate, packed
// byte-identical to their tree sources, plus the central documents that
// give them meaning (the control catalog, the framework crosswalks, and the
// pinned per-provider price books). Coverage is presence-based: a component
// without fact-sheets ships no cargo and its catalog entry carries no
// summaries -- absence is the honest state, never a fabricated zero.
const (
	costsPrefix       = "costs/"       // costs/<provider>/<kind>.yaml       <- catalog/<provider>/<kind>/cost.yaml
	controlsPrefix    = "controls/"    // controls/<provider>/<kind>.yaml    <- catalog/<provider>/<kind>/controls.yaml
	permissionsPrefix = "permissions/" // permissions/<provider>/<kind>.yaml <- catalog/<provider>/<kind>/iac/permissions.yaml
	estimatesPrefix   = "estimates/"   // estimates/<provider>/<kind>.yaml   <- catalog/_pricing/estimates/<component>.yaml
	derivationsPrefix = "derivations/" // derivations/<provider>/<kind>.yaml <- catalog/_pricing/derivations/<component>.yaml
	compliancePrefix  = "compliance/"  // the control catalog + frameworks/<framework>.yaml crosswalks
	pricebooksPrefix  = "pricebooks/"  // pricebooks/<provider>.yaml         <- catalog/_pricing/pricebook/<provider>.yaml

	controlsCatalogEntryName = compliancePrefix + "controls-catalog.yaml"
	frameworksEntryPrefix    = compliancePrefix + "frameworks/"
)

// componentCargo is one component's parsed fact-sheets. The parsed forms
// are produced from the exact bytes the bundle packs, so the entry
// summaries projected from them can never disagree with the cargo a
// consumer reads -- and the conformance gate re-proves that agreement from
// the finished zip.
type componentCargo struct {
	provider string
	kindDir  string

	cost        *costprofilev1.ComponentCostProfile
	controls    *controlprofilev1.ComponentControlProfile
	permissions *permissionsv1.ComponentPermissions
	// estimate is nil for covered components that deliberately ship no
	// estimate document (rate-delegated components whose honest price
	// exists only at composition time).
	estimate *costestimatev1.ComponentCostEstimate
	// derivation is nil for covered components whose estimates are still
	// hand-modeled -- derivations join component by component, and a
	// server-side estimator prices exactly the components whose rules are
	// aboard.
	derivation *derivationv1.ComponentCostDerivation
}

// collectCargo gathers the catalog tree's fact-sheet sidecars and central
// documents into bundle entries and returns the parsed cargo keyed by
// <provider>/<kind>. The sidecar standard is whole-or-not-at-all -- a
// component with any fact-sheet must have all three -- and an estimate must
// belong to a covered component; violations fail the build with the exact
// list. Underscore-prefixed catalog directories (central data homes and the
// _test provider) never ship component cargo.
func collectCargo(catalogDir string, entries map[string][]byte) (map[string]*componentCargo, error) {
	cargo := map[string]*componentCargo{}
	byKindDir := map[string]*componentCargo{}
	var problems []string

	costMatches, err := filepath.Glob(filepath.Join(catalogDir, "*", "*", "cost.yaml"))
	if err != nil {
		return nil, err
	}
	for _, match := range costMatches {
		componentDir := filepath.Dir(match)
		kindDir := filepath.Base(componentDir)
		provider := filepath.Base(filepath.Dir(componentDir))
		if strings.HasPrefix(provider, "_") {
			continue
		}
		key := provider + "/" + kindDir

		c := &componentCargo{
			provider:    provider,
			kindDir:     kindDir,
			cost:        &costprofilev1.ComponentCostProfile{},
			controls:    &controlprofilev1.ComponentControlProfile{},
			permissions: &permissionsv1.ComponentPermissions{},
		}
		sidecars := []struct {
			path      string
			entryName string
			doc       proto.Message
		}{
			{match, costsPrefix + key + ".yaml", c.cost},
			{filepath.Join(componentDir, "controls.yaml"), controlsPrefix + key + ".yaml", c.controls},
			{filepath.Join(componentDir, "iac", "permissions.yaml"), permissionsPrefix + key + ".yaml", c.permissions},
		}
		complete := true
		for _, sidecar := range sidecars {
			content, err := os.ReadFile(sidecar.path)
			if err != nil {
				problems = append(problems, fmt.Sprintf(
					"%s: the sidecar standard is whole-or-not-at-all, but %s is unreadable: %v", key, sidecar.path, err))
				complete = false
				continue
			}
			if err := protobufyaml.LoadYamlBytes(content, sidecar.doc); err != nil {
				problems = append(problems, fmt.Sprintf("%s does not parse against its schema: %v", sidecar.path, err))
				complete = false
				continue
			}
			entries[sidecar.entryName] = content
		}
		if !complete {
			continue
		}
		cargo[key] = c
		byKindDir[kindDir] = c
	}

	// The reverse direction of whole-or-not-at-all: a controls or
	// permissions sidecar existing without its cost profile.
	for glob, label := range map[string]string{
		filepath.Join(catalogDir, "*", "*", "controls.yaml"):           "controls.yaml",
		filepath.Join(catalogDir, "*", "*", "iac", "permissions.yaml"): "iac/permissions.yaml",
	} {
		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			rel, err := filepath.Rel(catalogDir, match)
			if err != nil {
				return nil, err
			}
			parts := strings.Split(rel, string(filepath.Separator))
			if strings.HasPrefix(parts[0], "_") {
				continue
			}
			if _, covered := cargo[parts[0]+"/"+parts[1]]; !covered {
				problems = append(problems, fmt.Sprintf(
					"%s/%s: ships %s without a cost.yaml -- the sidecar standard is whole-or-not-at-all", parts[0], parts[1], label))
			}
		}
	}

	// Estimates are generated centrally, one document per covered
	// component, re-keyed to the provider/kind convention every other
	// bundle tree uses.
	estimateMatches, err := filepath.Glob(filepath.Join(catalogDir, "_pricing", "estimates", "*.yaml"))
	if err != nil {
		return nil, err
	}
	for _, match := range estimateMatches {
		kindDir := strings.TrimSuffix(filepath.Base(match), ".yaml")
		c := byKindDir[kindDir]
		if c == nil {
			problems = append(problems, fmt.Sprintf(
				"estimate %s belongs to no component with a cost profile -- estimates price covered components only", match))
			continue
		}
		content, err := os.ReadFile(match)
		if err != nil {
			return nil, err
		}
		estimate := &costestimatev1.ComponentCostEstimate{}
		if err := protobufyaml.LoadYamlBytes(content, estimate); err != nil {
			problems = append(problems, fmt.Sprintf("%s does not parse against its schema: %v", match, err))
			continue
		}
		c.estimate = estimate
		entries[estimatesPrefix+c.provider+"/"+c.kindDir+".yaml"] = content
	}

	// Cost derivations are authored centrally, one document per derived
	// component, re-keyed to the same provider/kind convention. Presence
	// is per component: a derivation requires its component's cost cargo,
	// never the other way around.
	derivationMatches, err := filepath.Glob(filepath.Join(catalogDir, "_pricing", "derivations", "*.yaml"))
	if err != nil {
		return nil, err
	}
	for _, match := range derivationMatches {
		kindDir := strings.TrimSuffix(filepath.Base(match), ".yaml")
		c := byKindDir[kindDir]
		if c == nil {
			problems = append(problems, fmt.Sprintf(
				"derivation %s belongs to no component with a cost profile -- derivations price covered components only", match))
			continue
		}
		content, err := os.ReadFile(match)
		if err != nil {
			return nil, err
		}
		derivation := &derivationv1.ComponentCostDerivation{}
		if err := protobufyaml.LoadYamlBytes(content, derivation); err != nil {
			problems = append(problems, fmt.Sprintf("%s does not parse against its schema: %v", match, err))
			continue
		}
		c.derivation = derivation
		entries[derivationsPrefix+c.provider+"/"+c.kindDir+".yaml"] = content
	}

	// The central documents ride whenever any component cargo does: control
	// profiles are meaningless without the control catalog they cite, and
	// the crosswalks and price books are the release-owned context every
	// consuming surface reads beside the fact-sheets.
	if len(cargo) > 0 {
		centralProblems, err := collectCentralCargo(catalogDir, entries)
		if err != nil {
			return nil, err
		}
		problems = append(problems, centralProblems...)
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return nil, fmt.Errorf(
			"the catalog tree's fact-sheet cargo cannot ship (%d finding(s)):\n  %s",
			len(problems), strings.Join(problems, "\n  "))
	}
	return cargo, nil
}

// collectCentralCargo packs the control catalog, the framework crosswalks,
// and the per-provider price books, each strictly parsed before packing.
func collectCentralCargo(catalogDir string, entries map[string][]byte) ([]string, error) {
	var problems []string

	catalogPath := filepath.Join(catalogDir, "_compliance", "controls-catalog.yaml")
	content, err := os.ReadFile(catalogPath)
	if err != nil {
		problems = append(problems, fmt.Sprintf(
			"component cargo is aboard but the control catalog is unreadable at %s: %v", catalogPath, err))
	} else if err := protobufyaml.LoadYamlBytes(content, &controlcatalogv1.ControlCatalog{}); err != nil {
		problems = append(problems, fmt.Sprintf("%s does not parse against its schema: %v", catalogPath, err))
	} else {
		entries[controlsCatalogEntryName] = content
	}

	for _, central := range []struct {
		glob, entryPrefix, label string
		doc                      func() proto.Message
	}{
		{filepath.Join(catalogDir, "_compliance", "frameworks", "*.yaml"), frameworksEntryPrefix,
			"framework crosswalk", func() proto.Message { return &frameworkcrosswalkv1.FrameworkCrosswalk{} }},
		{filepath.Join(catalogDir, "_pricing", "pricebook", "*.yaml"), pricebooksPrefix,
			"price book", func() proto.Message { return &pricebookv1.PriceBook{} }},
	} {
		matches, err := filepath.Glob(central.glob)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			problems = append(problems, fmt.Sprintf(
				"component cargo is aboard but no %s exists under %s", central.label, filepath.Dir(central.glob)))
			continue
		}
		for _, match := range matches {
			content, err := os.ReadFile(match)
			if err != nil {
				return nil, err
			}
			if err := protobufyaml.LoadYamlBytes(content, central.doc()); err != nil {
				problems = append(problems, fmt.Sprintf("%s does not parse against its schema: %v", match, err))
				continue
			}
			entries[central.entryPrefix+filepath.Base(match)] = content
		}
	}
	return problems, nil
}

// applyCargoSummaries projects a covered component's fact-sheet summaries
// onto its catalog entry.
func applyCargoSummaries(entry *CatalogEntry, c *componentCargo) error {
	costSummary, err := computeCostSummary(c.cost, c.estimate)
	if err != nil {
		return err
	}
	entry.CostSummary = costSummary
	entry.ControlSummary = computeControlSummary(c.controls)
	entry.PermissionsProvenance = computePermissionsProvenance(c.permissions)
	return nil
}

// computeCostSummary derives the entry's cost summary from the cost
// profile's billing model and, when the component ships priced preset
// estimates, the min/max of their monthly totals. The bounds echo the
// estimate's own decimal strings verbatim -- the projection never
// re-renders money. Priced presets must agree on one currency; a
// disagreement is a build failure, never a published range.
func computeCostSummary(cost *costprofilev1.ComponentCostProfile, estimate *costestimatev1.ComponentCostEstimate) (*CatalogEntryCostSummary, error) {
	summary := &CatalogEntryCostSummary{BillingModel: cost.GetSpec().GetBillingModel().String()}
	if estimate == nil {
		return summary, nil
	}

	var minRat, maxRat *big.Rat
	for _, preset := range estimate.GetSpec().GetPresets() {
		total := preset.GetTotalListCost()
		if total == "" {
			// Capacity-footprint presets state no dollars -- they never
			// bound a monetary range.
			continue
		}
		rat, ok := new(big.Rat).SetString(total)
		if !ok {
			return nil, fmt.Errorf("preset %s total %q is not a decimal", preset.GetPreset(), total)
		}
		if summary.Currency == "" {
			summary.Currency = preset.GetCurrency()
		} else if summary.Currency != preset.GetCurrency() {
			return nil, fmt.Errorf("presets disagree on currency (%s vs %s) -- a range across currencies is not a number",
				summary.Currency, preset.GetCurrency())
		}
		if minRat == nil || rat.Cmp(minRat) < 0 {
			minRat, summary.MonthlyMin = rat, total
		}
		if maxRat == nil || rat.Cmp(maxRat) > 0 {
			maxRat, summary.MonthlyMax = rat, total
		}
	}
	return summary, nil
}

// computeControlSummary counts the profile's postures per status. The
// control-profile gate guarantees every catalog control is examined exactly
// once, so the counts always sum to the catalog's size.
func computeControlSummary(controls *controlprofilev1.ComponentControlProfile) *CatalogEntryControlSummary {
	summary := &CatalogEntryControlSummary{}
	for _, posture := range controls.GetSpec().GetControls() {
		switch posture.GetStatus() {
		case controlprofilev1.Status_enforced_by_default:
			summary.EnforcedByDefault++
		case controlprofilev1.Status_configurable:
			summary.Configurable++
		case controlprofilev1.Status_not_applicable:
			summary.NotApplicable++
		}
	}
	return summary
}

// computePermissionsProvenance aggregates provenance across every
// statement, group, and rule in the manifest's provider sections: one value
// when they agree, "mixed" when they do not, and empty (no claim) for a
// manifest with no entries.
func computePermissionsProvenance(permissions *permissionsv1.ComponentPermissions) string {
	seen := map[permissionsv1.Provenance]bool{}
	spec := permissions.GetSpec()
	for _, statement := range spec.GetAws().GetStatements() {
		seen[statement.GetProvenance()] = true
	}
	for _, group := range spec.GetGcp().GetGroups() {
		seen[group.GetProvenance()] = true
	}
	for _, group := range spec.GetAzure().GetGroups() {
		seen[group.GetProvenance()] = true
	}
	for _, rule := range spec.GetKubernetes().GetRules() {
		seen[rule.GetProvenance()] = true
	}
	for _, group := range spec.GetCloudflare().GetGroups() {
		seen[group.GetProvenance()] = true
	}
	for _, group := range spec.GetDigitalOcean().GetGroups() {
		seen[group.GetProvenance()] = true
	}
	for _, grant := range spec.GetDigitalOcean().GetSpacesGrants() {
		seen[grant.GetProvenance()] = true
	}
	for _, group := range spec.GetAuth0().GetGroups() {
		seen[group.GetProvenance()] = true
	}
	if len(seen) == 0 {
		return ""
	}
	if len(seen) > 1 {
		return "mixed"
	}
	for provenance := range seen {
		return provenance.String()
	}
	return ""
}

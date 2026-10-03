package costderivation

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	derivationv1 "github.com/plantonhq/planton/finops/catalogkindcostderivation/v1"
	costprofilev1 "github.com/plantonhq/planton/finops/catalogkindcostprofile/v1"
	pricebookv1 "github.com/plantonhq/planton/finops/pricebook/v1"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/finops/costprofile"
	"github.com/plantonhq/planton/pkg/finops/estimatemodel"
	"github.com/plantonhq/planton/pkg/finops/pricebook"
	"github.com/plantonhq/planton/pkg/specpath"
)

var decimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)

// TestCostDerivationConformance holds every cost derivation to its
// contract, offline:
//
//  1. The document parses strictly, names its kind (metadata.name
//     equals the filename), the kind ships a cost profile, and the
//     kind does NOT also ship a hand-authored estimate model -- a
//     kind's quantities have exactly one home.
//  2. Cluster-capacity kinds carry no COST derivation: they price
//     no meters, so their footprints derive through the capacity
//     standard (CatalogKindCapacityDerivation) -- a monetary derivation
//     for one would be dishonest.
//  3. Every field path anywhere in the document (region binding,
//     conditions, quantity factors, attribute bindings) resolves against
//     the served version's compiled descriptors -- a schema rename that
//     orphans a rule fails CI loudly -- UNDER THE ENGINE'S OWN traversal
//     contract (specpath.ResolvableTerminal): a path that walks through
//     a repeated or map field is refused here, at CI, instead of
//     erroring at replay. count_of paths are actually repeated fields;
//     expand_over and any_element_of paths are repeated MESSAGE fields,
//     and the paths inside their scopes resolve against the element.
//  4. Every line rule names a declared meter (a baseline charge or cost
//     driver sku_meter in the kind's cost.yaml), at least one
//     quantity factor, exactly one price arm (a slug resolving in the
//     provider's book, or a complete attribute lookup), and its basis
//     prose; every refusal states its reason; conditions carry an op and
//     the comparand exactly when the op compares (starts_with demands a
//     non-empty prefix); subtract_baseline baselines are positive
//     decimals.
//
// Whether the rules reproduce the hand-verified numbers is the estimate
// generator's replay check -- this gate keeps each derivation internally
// sound.
func TestCostDerivationConformance(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "catalog")); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test and the lint.catalog-data lane")
	}

	kindDirs, err := Discover(root)
	if err != nil {
		t.Fatalf("discovering cost derivations: %v", err)
	}
	if len(kindDirs) == 0 {
		t.Skip("no cost derivations authored yet")
	}

	for _, kindDir := range kindDirs {
		kindDir := kindDir
		t.Run(kindDir, func(t *testing.T) {
			derivation, err := Load(root, kindDir)
			if err != nil {
				t.Fatalf("cost derivation: %v", err)
			}
			if derivation.GetKind() != "CatalogKindCostDerivation" {
				t.Fatalf("kind is %q, want CatalogKindCostDerivation", derivation.GetKind())
			}
			if derivation.GetMetadata().GetName() != kindDir {
				t.Errorf("metadata.name is %q, want %q (the filename is the kind's identity)",
					derivation.GetMetadata().GetName(), kindDir)
			}

			if _, err := os.Stat(estimatemodel.Path(root, kindDir)); err == nil {
				t.Errorf("kind also ships an estimate model (%s) -- quantities have exactly one home; delete the model",
					estimatemodel.Path(root, kindDir))
			}

			provider := kindProvider(t, root, kindDir)
			profile, err := costprofile.Load(root, provider, kindDir)
			if err != nil {
				t.Fatalf("the derived kind must ship a cost profile: %v", err)
			}
			if profile.GetSpec().GetBillingModel() == costprofilev1.BillingModel_cluster_capacity {
				t.Fatal("cluster-capacity kinds carry no COST derivation -- they price no meters; author a CatalogKindCapacityDerivation (catalog/_pricing/capacity/) instead")
			}
			meters := declaredMeters(profile)

			entries, err := pricebook.Entries(root, provider)
			if err != nil {
				t.Fatalf("loading the provider's price book: %v", err)
			}

			specDescriptor := kindSpecDescriptor(t, kindDir)
			spec := derivation.GetSpec()

			if strings.TrimSpace(spec.GetCurrency()) == "" {
				t.Error("currency is empty")
			}
			if spec.GetHoursPerMonth() <= 0 {
				t.Error("hours_per_month must be declared (730 is the industry convention)")
			}

			if region := spec.GetRegion(); region != nil {
				if path := region.GetFromField(); path != "" {
					if _, err := specpath.ResolvableTerminal(specDescriptor, path); err != nil {
						t.Errorf("region from_field: %v", err)
					}
				}
			}

			for i, refusal := range spec.GetRefusals() {
				if len(refusal.GetWhen()) == 0 {
					t.Errorf("refusal %d has no conditions -- an unconditional refusal means the kind should ship no derivation", i)
				}
				checkConditions(t, specDescriptor, refusal.GetWhen(), false)
				if strings.TrimSpace(refusal.GetReason()) == "" {
					t.Errorf("refusal %d has no reason -- the honest sentence is the point", i)
				}
			}

			// A derivation with no lines is legitimate exactly when it
			// derives the honest zero: a zero-commit kind whose
			// exclusions name the usage meters. Nothing at all derives
			// nothing.
			if len(spec.GetLines()) == 0 && len(spec.GetRefusals()) == 0 && len(spec.GetExclusions()) == 0 {
				t.Error("derivation declares no line rules, no refusals, and no exclusions -- what does it derive?")
			}
			for _, rule := range spec.GetLines() {
				checkLineRule(t, specDescriptor, rule, meters, entries)
			}

			for i, text := range spec.GetExclusions() {
				if strings.TrimSpace(text.GetText()) == "" {
					t.Errorf("exclusion %d has no text", i)
				}
				checkConditions(t, specDescriptor, text.GetAppliesWhen(), false)
			}
			for i, note := range spec.GetNotes() {
				if strings.TrimSpace(note.GetText()) == "" {
					t.Errorf("note %d has no text", i)
				}
				checkConditions(t, specDescriptor, note.GetAppliesWhen(), false)
			}
		})
	}
}

// checkLineRule verifies one rule's internal soundness. A rule with
// expand_over validates its paths against the ELEMENT descriptor -- the
// scope the engine hands every path at replay.
func checkLineRule(
	t *testing.T,
	specDescriptor protoreflect.MessageDescriptor,
	rule *derivationv1.LineRule,
	meters map[string]bool,
	entries map[string]*pricebookv1.PriceBookEntry,
) {
	t.Helper()
	meter := strings.TrimSpace(rule.GetSkuMeter())
	if meter == "" {
		t.Error("line rule has no sku_meter")
	} else if !meters[meter] {
		t.Errorf("sku_meter %q is not declared by the kind's cost.yaml (baseline charges and cost drivers) -- a derivation cannot price an undeclared meter", meter)
	}

	// The scope the rule's VALUE-carrying paths resolve against: the
	// spec itself, or the expand_over field's element message.
	// applies_when always reads the root; element_applies_when reads
	// the element and is only legal on an expanded rule.
	scopeDescriptor := specDescriptor
	inExpansion := false
	if path := rule.GetExpandOver(); path != "" {
		terminal, err := specpath.ResolvableTerminal(specDescriptor, path)
		if err != nil {
			t.Errorf("line %q: expand_over: %v", meter, err)
			return
		}
		if !terminal.IsList() || terminal.IsMap() || terminal.Kind() != protoreflect.MessageKind {
			t.Errorf("line %q: expand_over %q is not a repeated message field", meter, path)
			return
		}
		scopeDescriptor = terminal.Message()
		inExpansion = true
	}

	checkConditions(t, specDescriptor, rule.GetAppliesWhen(), false)
	if len(rule.GetElementAppliesWhen()) > 0 && !inExpansion {
		t.Errorf("line %q carries element_applies_when without expand_over -- there are no elements to filter", meter)
	}
	checkConditions(t, scopeDescriptor, rule.GetElementAppliesWhen(), inExpansion)

	if len(rule.GetQuantity()) == 0 {
		t.Errorf("line %q has no quantity factors", meter)
	}
	for _, factor := range rule.GetQuantity() {
		switch f := factor.GetFactor().(type) {
		case *derivationv1.QuantityFactor_Constant:
			if !decimalPattern.MatchString(f.Constant) {
				t.Errorf("line %q: constant %q is not a plain decimal string", meter, f.Constant)
			}
		case *derivationv1.QuantityFactor_FieldValue:
			terminal, err := specpath.ResolvableTerminal(scopeDescriptor, f.FieldValue.GetFieldPath())
			if err != nil {
				t.Errorf("line %q: field_value: %v", meter, err)
			} else if terminal.IsList() || terminal.IsMap() {
				t.Errorf("line %q: field_value %q is not a scalar -- use count_of for repeated fields", meter, f.FieldValue.GetFieldPath())
			}
			if d := f.FieldValue.GetDefaultWhenUnset(); d != "" && !decimalPattern.MatchString(d) {
				t.Errorf("line %q: default_when_unset %q is not a plain decimal string", meter, d)
			}
		case *derivationv1.QuantityFactor_CountOf:
			terminal, err := specpath.ResolvableTerminal(scopeDescriptor, f.CountOf)
			if err != nil {
				t.Errorf("line %q: count_of: %v", meter, err)
			} else if !terminal.IsList() {
				t.Errorf("line %q: count_of %q is not a repeated field", meter, f.CountOf)
			}
		case *derivationv1.QuantityFactor_HoursInMonth:
			if !f.HoursInMonth {
				t.Errorf("line %q: hours_in_month must be true when used", meter)
			}
		case *derivationv1.QuantityFactor_SubtractBaseline:
			terminal, err := specpath.ResolvableTerminal(scopeDescriptor, f.SubtractBaseline.GetFieldPath())
			if err != nil {
				t.Errorf("line %q: subtract_baseline: %v", meter, err)
			} else if terminal.IsList() || terminal.IsMap() {
				t.Errorf("line %q: subtract_baseline %q is not a scalar", meter, f.SubtractBaseline.GetFieldPath())
			}
			baseline := f.SubtractBaseline.GetBaseline()
			if !decimalPattern.MatchString(baseline) {
				t.Errorf("line %q: subtract_baseline baseline %q is not a plain decimal string", meter, baseline)
			} else if allZeros(baseline) {
				t.Errorf("line %q: subtract_baseline baseline %q is zero -- a zero baseline is a plain field_value, not a subtraction", meter, baseline)
			}
		default:
			t.Errorf("line %q has a quantity factor with no arm set", meter)
		}
	}

	switch price := rule.GetPrice().(type) {
	case *derivationv1.LineRule_PriceSlug:
		if _, ok := entries[price.PriceSlug]; !ok {
			t.Errorf("line %q: price_slug %q resolves to no price book entry", meter, price.PriceSlug)
		}
	case *derivationv1.LineRule_PriceLookup:
		lookup := price.PriceLookup
		if strings.TrimSpace(lookup.GetServiceName()) == "" || strings.TrimSpace(lookup.GetPricingUnit()) == "" {
			t.Errorf("line %q: price_lookup must name service_name and pricing_unit", meter)
		}
		if len(lookup.GetAttributes()) == 0 {
			t.Errorf("line %q: price_lookup binds no attributes -- with nothing selecting the price, use price_slug", meter)
		}
		for _, binding := range lookup.GetAttributes() {
			if strings.TrimSpace(binding.GetKey()) == "" {
				t.Errorf("line %q: attribute binding has no key", meter)
			}
			switch value := binding.GetValue().(type) {
			case *derivationv1.AttributeBinding_Constant:
				if strings.TrimSpace(value.Constant) == "" {
					t.Errorf("line %q: attribute %q has an empty constant", meter, binding.GetKey())
				}
			case *derivationv1.AttributeBinding_FromField:
				if _, err := specpath.ResolvableTerminal(scopeDescriptor, value.FromField); err != nil {
					t.Errorf("line %q: attribute %q: %v", meter, binding.GetKey(), err)
				}
			default:
				t.Errorf("line %q: attribute %q carries neither constant nor from_field", meter, binding.GetKey())
			}
		}
	default:
		t.Errorf("line %q carries neither price_slug nor price_lookup", meter)
	}

	if strings.TrimSpace(rule.GetBasis()) == "" {
		t.Errorf("line %q has no basis -- the audit trail from manifest values to the quantity", meter)
	}
}

// checkConditions verifies each condition names a resolvable path, an op,
// and the comparand exactly when the op compares. Comparisons on
// value-or-reference wrapper fields are refused outright: the engine
// compares only the literal arm, so a manifest wiring the field by
// reference would silently compare against "" -- an unknowable value must
// never steer a comparison. Presence ops on wrappers are well-defined
// (either arm populated reads set) and stay legal.
//
// A condition with any_element_of validates its field_path against the
// named repeated message field's ELEMENT descriptor -- the scope the
// engine reads it against. inExpansion refuses any_element_of inside an
// expanded rule: there the conditions already read the current element.
func checkConditions(t *testing.T, specDescriptor protoreflect.MessageDescriptor, conditions []*derivationv1.Condition, inExpansion bool) {
	t.Helper()
	for _, condition := range conditions {
		scopeDescriptor := specDescriptor
		if listPath := condition.GetAnyElementOf(); listPath != "" {
			if inExpansion {
				t.Errorf("condition: any_element_of %q inside an expanded rule -- the conditions already read the current element", listPath)
				continue
			}
			listTerminal, err := specpath.ResolvableTerminal(specDescriptor, listPath)
			if err != nil {
				t.Errorf("condition: any_element_of %q: %v", listPath, err)
				continue
			}
			if !listTerminal.IsList() || listTerminal.IsMap() || listTerminal.Kind() != protoreflect.MessageKind {
				t.Errorf("condition: any_element_of %q is not a repeated message field", listPath)
				continue
			}
			scopeDescriptor = listTerminal.Message()
		}
		terminal, err := specpath.ResolvableTerminal(scopeDescriptor, condition.GetFieldPath())
		if err != nil {
			t.Errorf("condition: %v", err)
			continue
		}
		switch condition.GetOp() {
		case derivationv1.Condition_equals, derivationv1.Condition_not_equals:
			if condition.GetValue() == "" {
				t.Errorf("condition on %q compares against an empty value -- use is_unset for absence", condition.GetFieldPath())
			}
			if isReferenceCapable(terminal) {
				t.Errorf("condition on %q compares a value-or-reference field -- a referenced value is unknowable at estimate time; restructure the rule or use a presence op", condition.GetFieldPath())
			}
		case derivationv1.Condition_starts_with:
			if condition.GetValue() == "" {
				t.Errorf("condition on %q is a starts_with but carries no prefix -- an empty prefix would match everything", condition.GetFieldPath())
			}
			if isReferenceCapable(terminal) {
				t.Errorf("condition on %q prefix-matches a value-or-reference field -- a referenced value is unknowable at estimate time; restructure the rule or use a presence op", condition.GetFieldPath())
			}
		case derivationv1.Condition_is_set, derivationv1.Condition_is_unset:
			if condition.GetValue() != "" {
				t.Errorf("condition on %q is a presence check but carries value %q", condition.GetFieldPath(), condition.GetValue())
			}
		default:
			t.Errorf("condition on %q has no op", condition.GetFieldPath())
		}
	}
}

// allZeros reports whether a decimal string parses to zero ("0", "0.0").
func allZeros(decimal string) bool {
	for _, r := range decimal {
		if r != '0' && r != '.' {
			return false
		}
	}
	return true
}

// isReferenceCapable reports whether a field is the value-or-reference
// wrapper shape: a non-repeated message carrying a scalar `value` field
// beside at least one reference arm.
func isReferenceCapable(field protoreflect.FieldDescriptor) bool {
	if field.IsList() || field.IsMap() || field.Kind() != protoreflect.MessageKind {
		return false
	}
	value := field.Message().Fields().ByName("value")
	if value == nil || value.Kind() == protoreflect.MessageKind || value.IsList() || value.IsMap() {
		return false
	}
	return field.Message().Fields().Len() > 1
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
		t.Fatalf("cost derivation names kind %q, which exists nowhere under catalog/", kindDir)
	}
	if len(matches) > 1 {
		t.Fatalf("kind %q exists under multiple providers: %v", kindDir, matches)
	}
	return filepath.Base(filepath.Dir(matches[0]))
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

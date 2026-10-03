package costestimator

import (
	"math/big"
	"strings"
	"testing"

	albv1 "github.com/plantonhq/planton/catalog/aws/awsalb/v1alpha1"
	ebsv1 "github.com/plantonhq/planton/catalog/aws/awsebsvolume/v1alpha1"
	eksngv1 "github.com/plantonhq/planton/catalog/aws/awseksnodegroup/v1alpha1"
	eipv1 "github.com/plantonhq/planton/catalog/aws/awselasticip/v1alpha1"
	rdsclusterv1 "github.com/plantonhq/planton/catalog/aws/awsrdscluster/v1alpha1"
	rdsv1 "github.com/plantonhq/planton/catalog/aws/awsrdsinstance/v1alpha1"
	wafv1 "github.com/plantonhq/planton/catalog/aws/awswafwebacl/v1alpha1"
	derivationv1 "github.com/plantonhq/planton/finops/catalogkindcostderivation/v1"
	pricebookv1 "github.com/plantonhq/planton/finops/pricebook/v1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// entriesBySlug builds the pricebook.Entries shape from a list.
func entriesBySlug(entries ...*pricebookv1.PriceBookEntry) map[string]*pricebookv1.PriceBookEntry {
	m := map[string]*pricebookv1.PriceBookEntry{}
	for _, e := range entries {
		m[e.GetName()] = e
	}
	return m
}

func usdEntry(name, service, unit, region string) *pricebookv1.PriceBookEntry {
	return &pricebookv1.PriceBookEntry{
		Name:        name,
		ServiceName: service,
		PricingUnit: unit,
		Region:      region,
		Currency:    "USD",
	}
}

func constantFactor(value string) *derivationv1.QuantityFactor {
	return &derivationv1.QuantityFactor{Factor: &derivationv1.QuantityFactor_Constant{Constant: value}}
}

func fieldFactor(path, defaultWhenUnset string) *derivationv1.QuantityFactor {
	return &derivationv1.QuantityFactor{Factor: &derivationv1.QuantityFactor_FieldValue{
		FieldValue: &derivationv1.FieldValue{FieldPath: path, DefaultWhenUnset: defaultWhenUnset},
	}}
}

func countFactor(path string) *derivationv1.QuantityFactor {
	return &derivationv1.QuantityFactor{Factor: &derivationv1.QuantityFactor_CountOf{CountOf: path}}
}

func hoursFactor() *derivationv1.QuantityFactor {
	return &derivationv1.QuantityFactor{Factor: &derivationv1.QuantityFactor_HoursInMonth{HoursInMonth: true}}
}

func condition(path string, op derivationv1.Condition_Op, value string) *derivationv1.Condition {
	return &derivationv1.Condition{FieldPath: path, Op: op, Value: value}
}

// TestStaticSlugHoursAndCounts exercises the hours_in_month arm, count_of
// over a repeated wrapped-reference field, a conditional line, and the
// region binding falling back to its assumption when the manifest carries
// a placeholder.
func TestStaticSlugHoursAndCounts(t *testing.T) {
	manifest := &albv1.AwsAlb{Spec: &albv1.AwsAlbSpec{
		Region:   "<aws-region>",
		Subnets:  []*foreignkeyv1.StringValueOrRef{{}, {}},
		Internal: false,
	}}
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{FromField: "region", Assumption: "us-east-1"},
		Lines: []*derivationv1.LineRule{
			{
				SkuMeter: "Application Load Balancer hours",
				Quantity: []*derivationv1.QuantityFactor{constantFactor("1"), hoursFactor()},
				Price:    &derivationv1.LineRule_PriceSlug{PriceSlug: "alb-hours-us-east-1"},
				Basis:    "one load balancer, billed every hour of the month",
			},
			{
				SkuMeter:    "public IPv4 address usage",
				AppliesWhen: []*derivationv1.Condition{condition("internal", derivationv1.Condition_not_equals, "true")},
				Quantity:    []*derivationv1.QuantityFactor{countFactor("subnets"), hoursFactor()},
				Price:       &derivationv1.LineRule_PriceSlug{PriceSlug: "public-ipv4-in-use-us-east-1"},
				Basis:       "one public IPv4 node per subnet",
			},
		},
	}
	entries := entriesBySlug(
		usdEntry("alb-hours-us-east-1", "Elastic Load Balancing", "hours", "us-east-1"),
		usdEntry("public-ipv4-in-use-us-east-1", "Amazon Virtual Private Cloud", "IP-hours", "us-east-1"),
	)

	preset, refusal, err := Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate: err=%v refusal=%+v", err, refusal)
	}
	if preset.GetRegionAssumption() != "us-east-1" {
		t.Errorf("region: got %q, want the assumption us-east-1 (manifest carries a placeholder)", preset.GetRegionAssumption())
	}
	lines := preset.GetQuantityLines()
	if len(lines) != 2 {
		t.Fatalf("lines: got %d, want 2", len(lines))
	}
	if lines[0].GetPricingQuantity() != "730" {
		t.Errorf("hours line quantity: got %q, want 730", lines[0].GetPricingQuantity())
	}
	if lines[1].GetPricingQuantity() != "1460" {
		t.Errorf("IPv4 line quantity: got %q, want 1460 (2 subnets x 730)", lines[1].GetPricingQuantity())
	}

	// The internal ALB drops the IPv4 line entirely.
	manifest.Spec.Internal = true
	preset, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(internal): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 1 {
		t.Fatalf("internal ALB lines: got %d, want 1 (no public IPv4)", len(preset.GetQuantityLines()))
	}
}

// TestSlugRegionAndCurrencyAgreement pins the live boundary's agreement
// check on slug-named prices: pricing a Frankfurt manifest with a
// Virginia rate would be a wrong number that looks exactly like a right
// one -- the defining failure mode this engine exists to prevent -- so a
// slug entry whose region disagrees with the manifest's refuses (naming
// both regions), a "global" entry prices anywhere, and a currency
// disagreement refuses too. Mirrors the Java evaluator's identical pin.
func TestSlugRegionAndCurrencyAgreement(t *testing.T) {
	manifest := &albv1.AwsAlb{Spec: &albv1.AwsAlbSpec{
		Region:   "eu-central-1",
		Internal: true,
	}}
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{FromField: "region", Assumption: "us-east-1"},
		Lines: []*derivationv1.LineRule{{
			SkuMeter: "Application Load Balancer hours",
			Quantity: []*derivationv1.QuantityFactor{constantFactor("1"), hoursFactor()},
			Price:    &derivationv1.LineRule_PriceSlug{PriceSlug: "alb-hours-us-east-1"},
			Basis:    "one load balancer, billed every hour of the month",
		}},
	}

	// A slug entry pinned for another region refuses instead of silently
	// pricing this one, and the reason names both regions.
	entries := entriesBySlug(usdEntry("alb-hours-us-east-1", "Elastic Load Balancing", "hours", "us-east-1"))
	preset, refusal, err := Evaluate(manifest, spec, entries)
	if err != nil || preset != nil {
		t.Fatalf("Evaluate(wrong region): err=%v preset=%+v, want a refusal", err, preset)
	}
	if refusal == nil {
		t.Fatal("Evaluate(wrong region): priced instead of refusing")
	}
	if !strings.Contains(refusal.Reason, "eu-central-1") || !strings.Contains(refusal.Reason, "us-east-1") {
		t.Errorf("refusal reason must name both regions, got %q", refusal.Reason)
	}

	// A "global" entry prices the same manifest fine.
	entries = entriesBySlug(usdEntry("alb-hours-us-east-1", "Elastic Load Balancing", "hours", "global"))
	preset, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(global entry): err=%v refusal=%+v, want priced", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 1 {
		t.Fatalf("global entry lines: got %d, want 1", len(preset.GetQuantityLines()))
	}

	// A currency disagreement refuses too, naming both currencies.
	eurEntry := usdEntry("alb-hours-us-east-1", "Elastic Load Balancing", "hours", "eu-central-1")
	eurEntry.Currency = "EUR"
	entries = entriesBySlug(eurEntry)
	_, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil {
		t.Fatalf("Evaluate(wrong currency): %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "EUR") || !strings.Contains(refusal.Reason, "USD") {
		t.Errorf("currency refusal must name both currencies, got %+v", refusal)
	}
}

// TestAttributeLookup exercises value-keyed price selection: manifest
// values pick the entry, zero matches refuse, ambiguity refuses.
func TestAttributeLookup(t *testing.T) {
	manifest := &rdsv1.AwsRdsInstance{Spec: &rdsv1.AwsRdsInstanceSpec{
		Region:             "us-west-2",
		Engine:             "postgres",
		InstanceClass:      "db.m6g.large",
		AllocatedStorageGb: 50,
		MultiAz:            true,
	}}
	lookupRule := &derivationv1.LineRule{
		SkuMeter:    "instance hours",
		AppliesWhen: []*derivationv1.Condition{condition("multi_az", derivationv1.Condition_equals, "true")},
		Quantity:    []*derivationv1.QuantityFactor{hoursFactor()},
		Price: &derivationv1.LineRule_PriceLookup{PriceLookup: &derivationv1.PriceLookup{
			ServiceName: "Amazon Relational Database Service",
			PricingUnit: "hours",
			Attributes: []*derivationv1.AttributeBinding{
				{Key: "instance_class", Value: &derivationv1.AttributeBinding_FromField{FromField: "instance_class"}},
				{Key: "engine", Value: &derivationv1.AttributeBinding_FromField{FromField: "engine"}},
				{Key: "deployment", Value: &derivationv1.AttributeBinding_Constant{Constant: "multi-az"}},
			},
		}},
		Basis: "one instance, billed every hour",
	}
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{FromField: "region"},
		Lines:         []*derivationv1.LineRule{lookupRule},
	}
	matching := usdEntry("rds-db-m6g-large-postgres-multi-az-us-west-2", "Amazon Relational Database Service", "hours", "us-west-2")
	matching.Attributes = map[string]string{"instance_class": "db.m6g.large", "engine": "postgres", "deployment": "multi-az"}
	other := usdEntry("rds-db-r5-large-postgres-multi-az-us-west-2", "Amazon Relational Database Service", "hours", "us-west-2")
	other.Attributes = map[string]string{"instance_class": "db.r5.large", "engine": "postgres", "deployment": "multi-az"}

	preset, refusal, err := Evaluate(manifest, spec, entriesBySlug(matching, other))
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate: err=%v refusal=%+v", err, refusal)
	}
	if got := preset.GetQuantityLines()[0].GetPrice(); got != "rds-db-m6g-large-postgres-multi-az-us-west-2" {
		t.Errorf("resolved slug: got %q", got)
	}

	// An unpriced instance class refuses -- never a guess.
	manifest.Spec.InstanceClass = "db.x2iedn.24xlarge"
	_, refusal, err = Evaluate(manifest, spec, entriesBySlug(matching, other))
	if err != nil {
		t.Fatalf("Evaluate(unpriced): %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "no pinned price") {
		t.Fatalf("unpriced class: want a no-pinned-price refusal, got %+v", refusal)
	}

	// Two entries wearing the same identity refuse as ambiguous.
	manifest.Spec.InstanceClass = "db.m6g.large"
	duplicate := usdEntry("rds-duplicate", "Amazon Relational Database Service", "hours", "us-west-2")
	duplicate.Attributes = matching.GetAttributes()
	_, refusal, err = Evaluate(manifest, spec, entriesBySlug(matching, duplicate))
	if err != nil {
		t.Fatalf("Evaluate(ambiguous): %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "ambiguous") {
		t.Fatalf("ambiguous lookup: want an ambiguity refusal, got %+v", refusal)
	}
}

// TestRepeatedBindingPlurality pins the honesty rule for repeated fields:
// one element resolves, several refuse as indeterminate.
func TestRepeatedBindingPlurality(t *testing.T) {
	manifest := &eksngv1.AwsEksNodeGroup{Spec: &eksngv1.AwsEksNodeGroupSpec{
		Region:        "us-east-1",
		InstanceTypes: []string{"m6i.large"},
		Scaling:       &eksngv1.AwsEksNodeGroupScalingConfig{MinSize: 2, MaxSize: 5, DesiredSize: 2},
	}}
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{FromField: "region"},
		Lines: []*derivationv1.LineRule{{
			SkuMeter: "node instance hours",
			Quantity: []*derivationv1.QuantityFactor{fieldFactor("scaling.min_size", ""), hoursFactor()},
			Price: &derivationv1.LineRule_PriceLookup{PriceLookup: &derivationv1.PriceLookup{
				ServiceName: "Amazon Elastic Compute Cloud",
				PricingUnit: "hours",
				Attributes: []*derivationv1.AttributeBinding{
					{Key: "instance_type", Value: &derivationv1.AttributeBinding_FromField{FromField: "instance_types"}},
				},
			}},
			Basis: "the scaling floor x 730 hours",
		}},
	}
	entry := usdEntry("ec2-m6i-large-us-east-1", "Amazon Elastic Compute Cloud", "hours", "us-east-1")
	entry.Attributes = map[string]string{"instance_type": "m6i.large"}
	entries := entriesBySlug(entry)

	preset, refusal, err := Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(single type): err=%v refusal=%+v", err, refusal)
	}
	if got := preset.GetQuantityLines()[0].GetPricingQuantity(); got != "1460" {
		t.Errorf("quantity: got %q, want 1460 (floor 2 x 730)", got)
	}

	manifest.Spec.InstanceTypes = []string{"m6i.large", "m5.large", "m6a.large"}
	_, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil {
		t.Fatalf("Evaluate(several types): %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "indeterminate") {
		t.Fatalf("several instance types: want an indeterminate-rate refusal, got %+v", refusal)
	}
}

// TestRefusalRulesAndZeroQuantity pins the launch-template refusal and the
// spot scale-to-zero shape (zero-quantity lines are omitted, honesty prose
// still rides).
func TestRefusalRulesAndZeroQuantity(t *testing.T) {
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{FromField: "region", Assumption: "us-east-1"},
		Refusals: []*derivationv1.RefusalRule{{
			When:   []*derivationv1.Condition{condition("launch_template", derivationv1.Condition_is_set, "")},
			Reason: "the instance type and disk are delegated to the referenced AwsLaunchTemplate and invisible in this manifest",
		}},
		Lines: []*derivationv1.LineRule{{
			SkuMeter:    "node instance hours",
			AppliesWhen: []*derivationv1.Condition{condition("capacity_type", derivationv1.Condition_not_equals, "spot")},
			Quantity:    []*derivationv1.QuantityFactor{fieldFactor("scaling.min_size", ""), hoursFactor()},
			Price:       &derivationv1.LineRule_PriceSlug{PriceSlug: "ec2-m6i-large-us-east-1"},
			Basis:       "the scaling floor x 730 hours",
		}},
		Exclusions: []*derivationv1.ConditionalText{
			{Text: "data transfer follows pod traffic"},
			{
				AppliesWhen: []*derivationv1.Condition{condition("capacity_type", derivationv1.Condition_equals, "spot")},
				Text:        "running nodes bill the fluctuating Spot market rate",
			},
		},
		Notes: []*derivationv1.ConditionalText{{
			AppliesWhen: []*derivationv1.Condition{condition("capacity_type", derivationv1.Condition_equals, "spot")},
			Text:        "Spot has no committed list price.",
		}},
	}
	entries := entriesBySlug(usdEntry("ec2-m6i-large-us-east-1", "Amazon Elastic Compute Cloud", "hours", "us-east-1"))

	// Launch template: refusal, verbatim reason.
	withTemplate := &eksngv1.AwsEksNodeGroup{Spec: &eksngv1.AwsEksNodeGroupSpec{
		Region:         "us-east-1",
		LaunchTemplate: &eksngv1.AwsEksNodeGroupLaunchTemplate{},
		Scaling:        &eksngv1.AwsEksNodeGroupScalingConfig{MinSize: 2},
	}}
	_, refusal, err := Evaluate(withTemplate, spec, entries)
	if err != nil {
		t.Fatalf("Evaluate(launch template): %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "AwsLaunchTemplate") {
		t.Fatalf("launch template: want the delegation refusal, got %+v", refusal)
	}

	// Spot at min_size 0: no lines (the instance rule is condition-suppressed
	// AND the floor is zero), conditional prose rides.
	spot := &eksngv1.AwsEksNodeGroup{Spec: &eksngv1.AwsEksNodeGroupSpec{
		Region:        "us-east-1",
		InstanceTypes: []string{"m6i.large", "m5.large"},
		CapacityType:  eksngv1.AwsEksNodeGroupCapacityType_spot,
		Scaling:       &eksngv1.AwsEksNodeGroupScalingConfig{MinSize: 0, MaxSize: 10, DesiredSize: 3},
	}}
	preset, refusal, err := Evaluate(spot, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(spot): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 0 {
		t.Fatalf("spot lines: got %d, want 0", len(preset.GetQuantityLines()))
	}
	if len(preset.GetExclusions()) != 2 {
		t.Errorf("spot exclusions: got %d, want 2 (the always one + the spot one)", len(preset.GetExclusions()))
	}
	if preset.GetNotes() != "Spot has no committed list price." {
		t.Errorf("spot notes: got %q", preset.GetNotes())
	}

	// On-demand at the floor keeps the line and drops the spot prose.
	onDemand := &eksngv1.AwsEksNodeGroup{Spec: &eksngv1.AwsEksNodeGroupSpec{
		Region:        "us-east-1",
		InstanceTypes: []string{"m6i.large"},
		Scaling:       &eksngv1.AwsEksNodeGroupScalingConfig{MinSize: 2, MaxSize: 5, DesiredSize: 2},
	}}
	preset, refusal, err = Evaluate(onDemand, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(on-demand): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 1 || preset.GetQuantityLines()[0].GetPricingQuantity() != "1460" {
		t.Fatalf("on-demand lines: got %+v", preset.GetQuantityLines())
	}
	if len(preset.GetExclusions()) != 1 || preset.GetNotes() != "" {
		t.Errorf("on-demand prose: exclusions=%d notes=%q, want 1 and empty", len(preset.GetExclusions()), preset.GetNotes())
	}
}

// TestReferencePresence pins the presence semantics of value-or-reference
// wrapper fields: a wrapper populated on its reference arm reads SET (the
// manifest configured the field, even though the referenced value stays
// unknowable), a placeholder or absent wrapper reads unset, and a refusal
// rule keyed on the reference fires -- the silent-mispricing class where
// a knowably-understated estimate printed instead of refusing.
func TestReferencePresence(t *testing.T) {
	byReference := &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-ec2-instance", FieldPath: "status.outputs.instance_id"},
		},
	}
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{FromField: "region", Assumption: "us-east-1"},
		Lines: []*derivationv1.LineRule{
			{
				SkuMeter:    "public IPv4 address usage",
				AppliesWhen: []*derivationv1.Condition{condition("instance", derivationv1.Condition_is_set, "")},
				Quantity:    []*derivationv1.QuantityFactor{hoursFactor()},
				Price:       &derivationv1.LineRule_PriceSlug{PriceSlug: "public-ipv4-in-use-us-east-1"},
				Basis:       "attached to an instance, billed as in use",
			},
			{
				SkuMeter:    "public IPv4 address usage",
				AppliesWhen: []*derivationv1.Condition{condition("instance", derivationv1.Condition_is_unset, "")},
				Quantity:    []*derivationv1.QuantityFactor{hoursFactor()},
				Price:       &derivationv1.LineRule_PriceSlug{PriceSlug: "public-ipv4-idle-us-east-1"},
				Basis:       "unattached, billed as idle",
			},
		},
	}
	entries := entriesBySlug(
		usdEntry("public-ipv4-in-use-us-east-1", "Amazon Virtual Private Cloud", "IP-hours", "us-east-1"),
		usdEntry("public-ipv4-idle-us-east-1", "Amazon Virtual Private Cloud", "IP-hours", "us-east-1"),
	)

	// Reference arm populated: the field IS configured -- the in-use rule
	// fires and the idle rule does not.
	attached := &eipv1.AwsElasticIp{Spec: &eipv1.AwsElasticIpSpec{Region: "us-east-1", Instance: byReference}}
	preset, refusal, err := Evaluate(attached, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(by reference): err=%v refusal=%+v", err, refusal)
	}
	if lines := preset.GetQuantityLines(); len(lines) != 1 || lines[0].GetPrice() != "public-ipv4-in-use-us-east-1" {
		t.Fatalf("by reference: want exactly the in-use line, got %+v", lines)
	}

	// Placeholder literal: unset, exactly like an absent wrapper.
	for name, instance := range map[string]*foreignkeyv1.StringValueOrRef{
		"placeholder": {LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "<ec2-instance-id>"}},
		"absent":      nil,
	} {
		manifest := &eipv1.AwsElasticIp{Spec: &eipv1.AwsElasticIpSpec{Region: "us-east-1", Instance: instance}}
		preset, refusal, err := Evaluate(manifest, spec, entries)
		if err != nil || refusal != nil {
			t.Fatalf("Evaluate(%s): err=%v refusal=%+v", name, err, refusal)
		}
		if lines := preset.GetQuantityLines(); len(lines) != 1 || lines[0].GetPrice() != "public-ipv4-idle-us-east-1" {
			t.Fatalf("%s: want exactly the idle line, got %+v", name, lines)
		}
	}

	// A refusal rule keyed on the reference fires -- the honesty upgrade:
	// a fact delegated to a referenced resource refuses instead of
	// silently pricing without it.
	refusing := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{FromField: "region", Assumption: "us-east-1"},
		Refusals: []*derivationv1.RefusalRule{{
			When:   []*derivationv1.Condition{condition("instance", derivationv1.Condition_is_set, "")},
			Reason: "the committed charge rides the referenced instance and is invisible in this manifest",
		}},
		Lines: spec.Lines[1:2],
	}
	_, refusal, err = Evaluate(attached, refusing, entries)
	if err != nil {
		t.Fatalf("Evaluate(refusal by reference): %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "referenced instance") {
		t.Fatalf("refusal by reference: want the delegation refusal, got %+v", refusal)
	}
}

// TestMergeAndDefaults pins same-meter-same-price merging (quantities sum,
// bases join) and the default_when_unset contract.
func TestMergeAndDefaults(t *testing.T) {
	manifest := &rdsv1.AwsRdsInstance{Spec: &rdsv1.AwsRdsInstanceSpec{Region: "us-west-2"}}
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{FromField: "region"},
		Lines: []*derivationv1.LineRule{
			{
				SkuMeter: "provisioned storage",
				Quantity: []*derivationv1.QuantityFactor{fieldFactor("allocated_storage_gb", "100")},
				Price:    &derivationv1.LineRule_PriceSlug{PriceSlug: "rds-gp3-storage-us-west-2"},
				Basis:    "the primary volume",
			},
			{
				SkuMeter: "provisioned storage",
				Quantity: []*derivationv1.QuantityFactor{fieldFactor("allocated_storage_gb", "100"), constantFactor("2")},
				Price:    &derivationv1.LineRule_PriceSlug{PriceSlug: "rds-gp3-storage-us-west-2"},
				Basis:    "two replicas hold copies",
			},
		},
	}
	entries := entriesBySlug(usdEntry("rds-gp3-storage-us-west-2", "Amazon Relational Database Service", "GB-months", "us-west-2"))

	preset, refusal, err := Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate: err=%v refusal=%+v", err, refusal)
	}
	lines := preset.GetQuantityLines()
	if len(lines) != 1 {
		t.Fatalf("merged lines: got %d, want 1", len(lines))
	}
	if lines[0].GetPricingQuantity() != "300" {
		t.Errorf("merged quantity: got %q, want 300 (default 100 + 100x2)", lines[0].GetPricingQuantity())
	}
	if lines[0].GetQuantityBasis() != "the primary volume; two replicas hold copies" {
		t.Errorf("merged basis: got %q", lines[0].GetQuantityBasis())
	}
}

// TestExpandOverPerElementPricing pins the repeated-message expansion:
// a rule with expand_over evaluates once per element with the element
// as the scope for its value-carrying paths, so each instance's own
// class picks its own rate; identical elements merge to one line whose
// quantity is the sum and whose basis appears once, heterogeneous
// elements keep their own lines, applies_when gates the whole rule
// from the root while element_applies_when filters elements, an
// element whose class the book does not carry refuses the whole
// estimate, and an empty list emits nothing.
func TestExpandOverPerElementPricing(t *testing.T) {
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{Assumption: "us-west-2"},
		Lines: []*derivationv1.LineRule{
			{
				SkuMeter:   "instance hours",
				ExpandOver: "instances",
				AppliesWhen: []*derivationv1.Condition{
					condition("engine_mode", derivationv1.Condition_not_equals, "serverless"),
				},
				ElementAppliesWhen: []*derivationv1.Condition{
					condition("instance_class", derivationv1.Condition_not_equals, "db.serverless"),
				},
				Quantity: []*derivationv1.QuantityFactor{hoursFactor()},
				Price: &derivationv1.LineRule_PriceLookup{PriceLookup: &derivationv1.PriceLookup{
					ServiceName: "Amazon Relational Database Service",
					PricingUnit: "hours",
					Attributes: []*derivationv1.AttributeBinding{
						{Key: "instance_class", Value: &derivationv1.AttributeBinding_FromField{FromField: "instance_class"}},
					},
				}},
				Basis: "each provisioned instance's class selects its rate x 730 hours",
			},
		},
	}
	large := usdEntry("rds-large-us-west-2", "Amazon Relational Database Service", "hours", "us-west-2")
	large.Attributes = map[string]string{"instance_class": "db.r6g.large"}
	xlarge := usdEntry("rds-xlarge-us-west-2", "Amazon Relational Database Service", "hours", "us-west-2")
	xlarge.Attributes = map[string]string{"instance_class": "db.r6g.xlarge"}
	entries := entriesBySlug(large, xlarge)

	manifest := &rdsclusterv1.AwsRdsCluster{Spec: &rdsclusterv1.AwsRdsClusterSpec{
		Instances: []*rdsclusterv1.AwsRdsClusterInstance{
			{Name: "writer", InstanceClass: "db.r6g.large"},
			{Name: "reader-1", InstanceClass: "db.r6g.large"},
		},
	}}
	preset, refusal, err := Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(homogeneous): err=%v refusal=%+v", err, refusal)
	}
	lines := preset.GetQuantityLines()
	if len(lines) != 1 {
		t.Fatalf("homogeneous lines: got %d, want 1 (identical elements merge)", len(lines))
	}
	if lines[0].GetPricingQuantity() != "1460" {
		t.Errorf("merged quantity: got %q, want 1460 (2 instances x 730)", lines[0].GetPricingQuantity())
	}
	if lines[0].GetPrice() != "rds-large-us-west-2" {
		t.Errorf("merged price: got %q, want rds-large-us-west-2", lines[0].GetPrice())
	}
	if want := "each provisioned instance's class selects its rate x 730 hours"; lines[0].GetQuantityBasis() != want {
		t.Errorf("merged basis: got %q, want the rule's basis exactly once", lines[0].GetQuantityBasis())
	}

	// Heterogeneous classes keep their own lines, first-appearance order.
	manifest.Spec.Instances[1].InstanceClass = "db.r6g.xlarge"
	preset, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(heterogeneous): err=%v refusal=%+v", err, refusal)
	}
	lines = preset.GetQuantityLines()
	if len(lines) != 2 {
		t.Fatalf("heterogeneous lines: got %d, want 2", len(lines))
	}
	if lines[0].GetPrice() != "rds-large-us-west-2" || lines[1].GetPrice() != "rds-xlarge-us-west-2" {
		t.Errorf("heterogeneous prices: got %q, %q", lines[0].GetPrice(), lines[1].GetPrice())
	}
	if lines[0].GetPricingQuantity() != "730" || lines[1].GetPricingQuantity() != "730" {
		t.Errorf("heterogeneous quantities: got %q, %q, want 730 each", lines[0].GetPricingQuantity(), lines[1].GetPricingQuantity())
	}

	// element_applies_when filters elements: a serverless instance
	// contributes nothing while its provisioned sibling still prices.
	manifest.Spec.Instances[1].InstanceClass = "db.serverless"
	preset, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(filtered): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 1 {
		t.Fatalf("filtered lines: got %d, want 1 (the serverless element is filtered)", len(preset.GetQuantityLines()))
	}

	// applies_when gates the WHOLE expanded rule from the spec root: a
	// root value the rule's conditions reject drops every element.
	manifest.Spec.Instances[1].InstanceClass = "db.r6g.large"
	manifest.Spec.EngineMode = "serverless"
	preset, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(root-gated): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 0 {
		t.Fatalf("root-gated lines: got %d, want 0 (applies_when reads the root)", len(preset.GetQuantityLines()))
	}
	manifest.Spec.EngineMode = ""

	// An element whose class the book does not carry refuses the WHOLE
	// estimate -- a total missing an instance it owes is a lie.
	manifest.Spec.Instances[1].InstanceClass = "db.r6g.4xlarge"
	_, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil {
		t.Fatalf("Evaluate(unpriced): %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "no pinned price") {
		t.Fatalf("unpriced element: got refusal %+v, want a no-pinned-price refusal", refusal)
	}

	// An empty list emits nothing.
	manifest.Spec.Instances = nil
	preset, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(empty): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 0 {
		t.Fatalf("empty-list lines: got %d, want 0", len(preset.GetQuantityLines()))
	}

	// expand_over must name a repeated message field.
	spec.Lines[0].ExpandOver = "engine"
	if _, _, err := Evaluate(manifest, spec, entries); err == nil {
		t.Fatal("expand_over on a scalar field: want an error, got none")
	}
}

// TestAnyElementCondition pins the existential condition: it ranges
// over a repeated message field's elements and holds when ANY element
// matches, reading paths element-relative -- the shape of billing
// identities buried inside repeated configuration trees (a WAF rule
// referencing a subscription-bearing managed rule group).
func TestAnyElementCondition(t *testing.T) {
	managedGroup := func(name string) *wafv1.AwsWafWebAclRule {
		return &wafv1.AwsWafWebAclRule{
			Name: name,
			Statement: &wafv1.AwsWafWebAclStatement{
				Statement: &wafv1.AwsWafWebAclStatement_ManagedRuleGroup{
					ManagedRuleGroup: &wafv1.AwsWafWebAclManagedRuleGroupStatement{
						Name:       name,
						VendorName: "AWS",
					},
				},
			},
		}
	}
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{Assumption: "us-east-1"},
		Refusals: []*derivationv1.RefusalRule{
			{
				When: []*derivationv1.Condition{{
					AnyElementOf: "rules",
					FieldPath:    "statement.managed_rule_group.name",
					Op:           derivationv1.Condition_equals,
					Value:        "AWSManagedRulesBotControlRuleSet",
				}},
				Reason: "Bot Control carries a monthly subscription this price book does not carry",
			},
		},
		Lines: []*derivationv1.LineRule{
			{
				SkuMeter: "web ACL monthly fee",
				Quantity: []*derivationv1.QuantityFactor{constantFactor("1")},
				Price:    &derivationv1.LineRule_PriceSlug{PriceSlug: "waf-acl-monthly"},
				Basis:    "one web ACL",
			},
		},
	}
	entries := entriesBySlug(usdEntry("waf-acl-monthly", "AWS WAF", "months", "global"))

	// Core managed groups carry no subscription -- the refusal must not
	// fire, including for a rule whose statement is not a managed group
	// at all (the intermediate resolves absent, never errors).
	manifest := &wafv1.AwsWafWebAcl{Spec: &wafv1.AwsWafWebAclSpec{
		Rules: []*wafv1.AwsWafWebAclRule{
			managedGroup("AWSManagedRulesCommonRuleSet"),
			{Name: "rate-limit", Statement: &wafv1.AwsWafWebAclStatement{}},
		},
	}}
	preset, refusal, err := Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(core groups): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 1 {
		t.Fatalf("core-group lines: got %d, want 1", len(preset.GetQuantityLines()))
	}

	// Any element referencing the subscription group refuses.
	manifest.Spec.Rules = append(manifest.Spec.Rules, managedGroup("AWSManagedRulesBotControlRuleSet"))
	_, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil {
		t.Fatalf("Evaluate(bot control): %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "subscription") {
		t.Fatalf("bot control: got refusal %+v, want the subscription refusal", refusal)
	}

	// An empty list satisfies nothing.
	manifest.Spec.Rules = nil
	_, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(no rules): err=%v refusal=%+v", err, refusal)
	}

	// any_element_of must name a repeated message field.
	spec.Refusals[0].When[0].AnyElementOf = "name"
	if _, _, err := Evaluate(manifest, spec, entries); err == nil {
		t.Fatal("any_element_of on a scalar field: want an error, got none")
	}
}

// TestSubtractBaseline pins the excess-over-baseline factor: only the
// amount above the included allotment bills (gp3's included 3,000 IOPS
// shape), at-or-under the baseline is an honest zero that drops the
// line, and the factor multiplies with its siblings (the per-hour
// dialed-performance shape).
func TestSubtractBaseline(t *testing.T) {
	subtractFactor := func(path, baseline string) *derivationv1.QuantityFactor {
		return &derivationv1.QuantityFactor{Factor: &derivationv1.QuantityFactor_SubtractBaseline{
			SubtractBaseline: &derivationv1.SubtractBaseline{FieldPath: path, Baseline: baseline},
		}}
	}
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{Assumption: "us-east-1"},
		Lines: []*derivationv1.LineRule{
			{
				SkuMeter: "provisioned IOPS",
				Quantity: []*derivationv1.QuantityFactor{subtractFactor("iops", "3000")},
				Price:    &derivationv1.LineRule_PriceSlug{PriceSlug: "gp3-iops-us-east-1"},
				Basis:    "only IOPS above gp3's included 3,000 baseline bill monthly",
			},
		},
	}
	entries := entriesBySlug(usdEntry("gp3-iops-us-east-1", "Amazon Elastic Block Store", "IOPS-months", "us-east-1"))

	manifest := &ebsv1.AwsEbsVolume{Spec: &ebsv1.AwsEbsVolumeSpec{Iops: 6000}}
	preset, refusal, err := Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(6000): err=%v refusal=%+v", err, refusal)
	}
	if got := preset.GetQuantityLines(); len(got) != 1 || got[0].GetPricingQuantity() != "3000" {
		t.Fatalf("excess: got %+v, want one line of 3000 (6000 - 3000)", got)
	}

	// At the baseline the excess is an honest zero and the line drops.
	manifest.Spec.Iops = 3000
	preset, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(3000): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 0 {
		t.Fatalf("at-baseline lines: got %d, want 0", len(preset.GetQuantityLines()))
	}

	// Unset (zero) floors at zero, never a negative quantity.
	manifest.Spec.Iops = 0
	preset, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(unset): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 0 {
		t.Fatalf("unset lines: got %d, want 0", len(preset.GetQuantityLines()))
	}

	// The factor multiplies with its siblings: (8000 - 3000) x 730, the
	// dialed-performance per-hour shape.
	spec.Lines[0].Quantity = []*derivationv1.QuantityFactor{subtractFactor("iops", "3000"), hoursFactor()}
	manifest.Spec.Iops = 8000
	preset, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(8000 x hours): err=%v refusal=%+v", err, refusal)
	}
	if got := preset.GetQuantityLines(); len(got) != 1 || got[0].GetPricingQuantity() != "3650000" {
		t.Fatalf("product: got %+v, want one line of 3650000 (5000 x 730)", got)
	}

	// A non-positive or malformed baseline is a malformed derivation.
	for _, bad := range []string{"0", "abc", ""} {
		spec.Lines[0].Quantity = []*derivationv1.QuantityFactor{subtractFactor("iops", bad)}
		if _, _, err := Evaluate(manifest, spec, entries); err == nil {
			t.Errorf("baseline %q: want an error, got none", bad)
		}
	}
}

// TestStartsWithCondition pins the prefix op: version-family detection
// where enumerating every member is impossible ("composer-3" must catch
// composer-3.1.2 and every future patch), an unset field never matches,
// and an empty prefix is a malformed derivation.
func TestStartsWithCondition(t *testing.T) {
	spec := &derivationv1.CatalogKindCostDerivationSpec{
		Currency:      "USD",
		HoursPerMonth: 730,
		Region:        &derivationv1.RegionBinding{Assumption: "us-west-2"},
		Refusals: []*derivationv1.RefusalRule{
			{
				When:   []*derivationv1.Condition{condition("engine_version", derivationv1.Condition_starts_with, "8.")},
				Reason: "the 8.x engine family bills meters this price book does not carry",
			},
		},
		Lines: []*derivationv1.LineRule{
			{
				SkuMeter: "instance hours",
				Quantity: []*derivationv1.QuantityFactor{hoursFactor()},
				Price:    &derivationv1.LineRule_PriceSlug{PriceSlug: "rds-hours-us-west-2"},
				Basis:    "one instance x 730 hours",
			},
		},
	}
	entries := entriesBySlug(usdEntry("rds-hours-us-west-2", "Amazon Relational Database Service", "hours", "us-west-2"))

	manifest := &rdsclusterv1.AwsRdsCluster{Spec: &rdsclusterv1.AwsRdsClusterSpec{EngineVersion: "8.0.mysql_aurora.3.05.2"}}
	_, refusal, err := Evaluate(manifest, spec, entries)
	if err != nil {
		t.Fatalf("Evaluate(8.x): %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "8.x engine family") {
		t.Fatalf("8.x: got refusal %+v, want the family refusal", refusal)
	}

	// A different family prices normally.
	manifest.Spec.EngineVersion = "5.7.mysql_aurora.2.11.4"
	preset, refusal, err := Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(5.7): err=%v refusal=%+v", err, refusal)
	}
	if len(preset.GetQuantityLines()) != 1 {
		t.Fatalf("5.7 lines: got %d, want 1", len(preset.GetQuantityLines()))
	}

	// An unset field renders empty and never matches a non-empty prefix.
	manifest.Spec.EngineVersion = ""
	_, refusal, err = Evaluate(manifest, spec, entries)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(unset): err=%v refusal=%+v", err, refusal)
	}

	// An empty prefix is a malformed derivation, not a vacuous match.
	spec.Refusals[0].When[0].Value = ""
	if _, _, err := Evaluate(manifest, spec, entries); err == nil {
		t.Fatal("empty starts_with value: want an error, got none")
	}
}

// TestPlainDecimal pins the quantity rendering: numbers a person would
// write, never manufactured trailing zeros.
func TestPlainDecimal(t *testing.T) {
	cases := []struct {
		num, denom int64
		want       string
	}{
		{730, 1, "730"},
		{1460, 1, "1460"},
		{1, 2, "0.5"},
		{3, 4, "0.75"},
		{0, 1, "0"},
	}
	for _, c := range cases {
		if got := plainDecimal(big.NewRat(c.num, c.denom)); got != c.want {
			t.Errorf("plainDecimal(%d/%d): got %q, want %q", c.num, c.denom, got, c.want)
		}
	}
}

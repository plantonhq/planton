//go:build !codegen
// +build !codegen

package providerparity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// accountingFixture exercises every accounting shape in one hermetic
// catalog: exact matches, subtree and leaf mappings, exclusions, an
// iam-member secondary resource under a specRoot, an internal plumbing
// resource, machinery args, deprecated args, reverse drift, a manifest-less
// kind with renames, and every breadth disposition class.
func accountingFixture() (spec []KindCensus, modules []ModuleCensus, schemas map[string]*Schema, manifests map[string]*Manifest, ledger []LedgerEntry) {
	schemas = map[string]*Schema{"google": {
		Provider: "google",
		Source:   "hashicorp/google",
		Version:  "6.50.0",
		Resources: map[string]*Block{
			"google_widget": {
				Attributes: map[string]*Attribute{
					"name":     {Required: true},
					"location": {Optional: true},
					"id":       {Optional: true},                   // machinery: standing exclusion
					"old_knob": {Optional: true, Deprecated: true}, // deprecated: outside accounting
				},
				Blocks: map[string]*NestedBlock{
					"settings": {NestingMode: "single", Block: &Block{
						Attributes: map[string]*Attribute{"enabled": {Optional: true}},
					}},
					"rules": {NestingMode: "list", Block: &Block{
						Attributes: map[string]*Attribute{"mode": {Optional: true}},
						Blocks: map[string]*NestedBlock{
							"condition": {NestingMode: "single", Block: &Block{
								Attributes: map[string]*Attribute{
									"age":              {Optional: true},
									"send_age_if_zero": {Optional: true},
								},
							}},
						},
					}},
					// A whole embedded surface the module deliberately fixes:
					// covered by ONE subtree exclusion, never per-leaf lines.
					"inline_gadget": {NestingMode: "list", Block: &Block{
						Attributes: map[string]*Attribute{"size": {Optional: true}},
						Blocks: map[string]*NestedBlock{
							"tuning": {NestingMode: "single", Block: &Block{
								Attributes: map[string]*Attribute{"level": {Optional: true}},
							}},
						},
					}},
					"timeouts": {NestingMode: "single", Block: &Block{ // machinery
						Attributes: map[string]*Attribute{"create": {Optional: true}},
					}},
				},
			},
			"google_widget_iam_member": {
				Attributes: map[string]*Attribute{
					"widget": {Required: true}, // module-wired linkage: manifest exclusion
					"role":   {Required: true},
					"member": {Required: true},
				},
				Blocks: map[string]*NestedBlock{
					"condition": {NestingMode: "single", Block: &Block{
						Attributes: map[string]*Attribute{"title": {Optional: true}},
					}},
				},
			},
			"google_project_service": {
				Attributes: map[string]*Attribute{
					"service":            {Required: true},
					"disable_on_destroy": {Optional: true},
				},
			},
			"google_plain": {
				Attributes: map[string]*Attribute{"name": {Required: true}},
			},
			"google_widget_iam_policy": { // unconsumed iam triplet: pattern class
				Attributes: map[string]*Attribute{"policy_data": {Required: true}},
			},
			"google_dead_thing": { // schema-flagged deprecation class
				Deprecated: true,
				Attributes: map[string]*Attribute{"name": {Required: true}},
			},
			"google_composed_thing": { // ledger class
				Attributes: map[string]*Attribute{"name": {Required: true}},
			},
			"google_orphan_thing": { // no disposition anywhere: Finding
				Attributes: map[string]*Attribute{"name": {Required: true}},
			},
		},
	}}

	spec = []KindCensus{
		{Kind: "TestPlain", SpecFieldPaths: []string{"spec.plain_name"}},
		{Kind: "TestWidget", SpecFieldPaths: []string{
			"spec.widget_name",
			"spec.location",
			"spec.settings.enabled",
			"spec.rule_list.mode",
			"spec.rule_list.condition.age_days",
			"spec.iam_members.role",
			"spec.iam_members.member",
			"spec.iam_members.condition.title",
			"spec.platform_only",
			"spec.drifted_field",
		}},
	}

	modules = []ModuleCensus{
		{Kind: "TestPlain", Resources: []string{"google_plain"}, Pins: map[string]string{"google": "~> 6.0"}},
		{Kind: "TestWidget", Resources: []string{
			"google_widget", "google_widget_iam_member", "google_project_service",
		}, Pins: map[string]string{"google": "~> 6.0"}},
	}

	manifests = map[string]*Manifest{
		"TestWidget": {
			Resources: map[string]*ResourceManifest{
				"google_widget": {
					Mappings: []Mapping{
						{Spec: "spec.widget_name", Arg: "name"},
						// Subtree mapping: exact matching resumes below it...
						{Spec: "spec.rule_list", Arg: "rules"},
						// ...and a longer (leaf) mapping wins over the subtree.
						{Spec: "spec.rule_list.condition.age_days", Arg: "rules.condition.age"},
					},
					Exclusions: []ArgExclusion{
						{Arg: "rules.condition.send_age_if_zero", Reason: "proto3 optional presence covers zero-vs-unset"},
						// Subtree exclusion: one judgment covers the block
						// and every leaf under it.
						{Arg: "inline_gadget", Reason: "the module fixes the inline gadget; gadgets are the sibling kind's surface"},
					},
				},
				"google_widget_iam_member": {
					SpecRoot: "spec.iam_members",
					Exclusions: []ArgExclusion{
						{Arg: "widget", Reason: "wired by the module to the enclosing widget"},
					},
				},
				"google_project_service": {
					Internal: "API enablement plumbing; arguments are module decisions",
				},
			},
			SpecExclusions: []SpecExclusion{
				{Field: "spec.platform_only", Reason: "platform concept with no provider counterpart"},
			},
		},
	}

	ledger = []LedgerEntry{
		{Resource: "google_composed_thing", Disposition: DispositionComposed, Reason: "covered by TestWidget's composed_thing field"},
		{Resource: "google_widget", Disposition: DispositionDeferred, Reason: "stale: the module consumes it now"},
		{Resource: "google_gone_thing", Disposition: DispositionDeferred, Reason: "stale: removed at the pin"},
	}
	return
}

func kindByName(t *testing.T, acc Accounting, kind string) KindAccounting {
	t.Helper()
	for _, k := range acc.Kinds {
		if k.Kind == kind {
			return k
		}
	}
	t.Fatalf("kind %s missing from accounting", kind)
	return KindAccounting{}
}

func TestBuildAccounting_Hermetic(t *testing.T) {
	spec, modules, schemas, manifests, ledger := accountingFixture()
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, ledger, nil)

	widget := kindByName(t, acc, "TestWidget")
	// google_widget: name(mapped) location(matched) settings.enabled(matched)
	// rules.mode(mapped via subtree) rules.condition.age(mapped via leaf)
	// send_age_if_zero(excluded) inline_gadget.size + inline_gadget.tuning.level
	// (both excluded by the ONE subtree exclusion); id/timeouts.create
	// machinery and old_knob deprecated -- all outside accounting.
	// google_widget_iam_member: role/member/condition.title matched under
	// specRoot, widget excluded. google_project_service: internal.
	if widget.TotalArgs != 12 {
		t.Errorf("TotalArgs = %d, want 12 (machinery and deprecated args must not count)", widget.TotalArgs)
	}
	if widget.MatchedArgs != 5 || widget.MappedArgs != 3 || widget.ExcludedArgs != 4 {
		t.Errorf("matched/mapped/excluded = %d/%d/%d, want 5/3/4 (the subtree exclusion covers both inline_gadget leaves)",
			widget.MatchedArgs, widget.MappedArgs, widget.ExcludedArgs)
	}
	if !reflect.DeepEqual(widget.InternalResources, []string{"google_project_service"}) {
		t.Errorf("InternalResources = %v", widget.InternalResources)
	}
	if len(widget.UnaccountedArgs) != 0 {
		t.Errorf("UnaccountedArgs = %v, want none", widget.UnaccountedArgs)
	}
	// Reverse direction: the drifted field is caught; the platform field is
	// covered by its recorded exclusion.
	if !reflect.DeepEqual(widget.UncoveredSpecFields, []string{"spec.drifted_field"}) {
		t.Errorf("UncoveredSpecFields = %v, want [spec.drifted_field]", widget.UncoveredSpecFields)
	}
	if widget.Accounted() {
		t.Error("TestWidget carries reverse drift and must not count as accounted")
	}

	// A manifest-less kind is held to pure exact matching: google_plain's
	// "name" does not match spec.plain_name in either direction.
	plain := kindByName(t, acc, "TestPlain")
	if plain.HasManifest {
		t.Error("TestPlain has no manifest")
	}
	if !reflect.DeepEqual(plain.UnaccountedArgs, []string{"google_plain: name"}) {
		t.Errorf("UnaccountedArgs = %v", plain.UnaccountedArgs)
	}
	if !reflect.DeepEqual(plain.UncoveredSpecFields, []string{"spec.plain_name"}) {
		t.Errorf("UncoveredSpecFields = %v", plain.UncoveredSpecFields)
	}

	// Breadth: every fixture resource carries its class.
	byResource := map[string]ResourceDisposition{}
	for _, d := range acc.Dispositions {
		byResource[d.Resource] = d
	}
	wantDispositions := map[string]string{
		"google_widget":            DispositionModeled,
		"google_widget_iam_member": DispositionModeled,
		"google_project_service":   DispositionModeled,
		"google_plain":             DispositionModeled,
		"google_widget_iam_policy": DispositionIamCovered,
		"google_dead_thing":        DispositionExcludedDeprecated,
		"google_composed_thing":    DispositionComposed,
		"google_orphan_thing":      "",
	}
	for res, want := range wantDispositions {
		if got := byResource[res].Disposition; got != want {
			t.Errorf("disposition[%s] = %q, want %q", res, got, want)
		}
	}
	if byResource["google_widget"].Detail != "consumed by TestWidget" {
		t.Errorf("modeled detail = %q", byResource["google_widget"].Detail)
	}
	if acc.DispositionTotals[DispositionModeled] != 4 || acc.DispositionTotals[""] != 1 {
		t.Errorf("DispositionTotals = %v", acc.DispositionTotals)
	}

	// Findings: two kinds in debt, the orphan resource, the two stale
	// ledger entries (consumed + removed-at-pin).
	keys := map[string]int{}
	for _, f := range acc.Findings {
		keys[f.BaselineKey]++
	}
	want := map[string]int{
		"kind:TestWidget":              1, // reverse drift
		"kind:TestPlain":               2, // unaccounted arg + uncovered spec field
		"resource:google_orphan_thing": 1,
		"resource:google_widget":       1, // stale ledger: consumed
		"resource:google_gone_thing":   1, // stale ledger: not at the pin
	}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("finding keys = %v, want %v", keys, want)
	}
}

// TestBuildAccounting_FanInMappings proves the dual of the name/value
// idiom: several mappings recording ONE map-typed argument each cover
// their own spec field — the argument counts as mapped once, and BOTH
// spec fields are covered in the reverse direction. The shape: a
// provider packing several honest dials (cpu, memory) into one limits
// map.
func TestBuildAccounting_FanInMappings(t *testing.T) {
	schemas := map[string]*Schema{"google": {
		Provider: "google",
		Source:   "hashicorp/google",
		Version:  "7.43.0",
		Resources: map[string]*Block{
			"google_box": {
				Attributes: map[string]*Attribute{
					"name":   {Required: true},
					"limits": {Optional: true},
				},
			},
		},
	}}
	spec := []KindCensus{{Kind: "TestBox", SpecFieldPaths: []string{
		"spec.box_name",
		"spec.resources.cpu",
		"spec.resources.memory",
	}}}
	modules := []ModuleCensus{{Kind: "TestBox", Resources: []string{"google_box"},
		Pins: map[string]string{"google": "~> 7.43"}}}
	manifests := map[string]*Manifest{"TestBox": {
		Resources: map[string]*ResourceManifest{
			"google_box": {
				Mappings: []Mapping{
					{Spec: "spec.box_name", Arg: "name"},
					{Spec: "spec.resources.cpu", Arg: "limits"},
					{Spec: "spec.resources.memory", Arg: "limits"},
				},
			},
		},
	}}

	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, nil, nil)
	box := kindByName(t, acc, "TestBox")
	if box.TotalArgs != 2 || box.MappedArgs != 2 || box.MatchedArgs != 0 {
		t.Errorf("total/mapped/matched = %d/%d/%d, want 2/2/0 (the fan-in arg counts once)",
			box.TotalArgs, box.MappedArgs, box.MatchedArgs)
	}
	if len(box.UnaccountedArgs) != 0 {
		t.Errorf("UnaccountedArgs = %v, want none", box.UnaccountedArgs)
	}
	if len(box.UncoveredSpecFields) != 0 {
		t.Errorf("UncoveredSpecFields = %v, want none (both map-realized fields are covered)", box.UncoveredSpecFields)
	}
	if !box.Accounted() {
		t.Error("TestBox must be at total accounting")
	}
}

// TestBuildAccounting_LedgerShadowsComputedClasses proves the computed
// classes always win over the ledger AND that a shadowed entry is a
// staleness finding: judgment duplicating what the instrument derives on
// its own is judgment nobody re-evaluates.
// TestBuildAccounting_CollapseMapping proves the subtree-to-leaf judgment
// for recursive provider grammars: the provider expands a recursive schema
// into bounded nesting levels (thousands of leaf paths), the spec census
// records each re-entry as ONE leaf, and a collapse mapping folds every
// argument under the re-entry subtree onto that leaf. Without the flag the
// same mapping resumes exact matching below the subtree and the deep args
// surface as findings — collapse is never implicit.
func TestBuildAccounting_CollapseMapping(t *testing.T) {
	schemas := map[string]*Schema{"aws": {
		Provider: "aws",
		Source:   "hashicorp/aws",
		Version:  "6.58.0",
		Resources: map[string]*Block{
			"aws_recursive_thing": {
				Blocks: map[string]*NestedBlock{
					// The provider's bounded expansion of a recursive
					// grammar: statement.kind at the root, and the same
					// grammar again under and.statement (two levels).
					"statement": {NestingMode: "single", Block: &Block{
						Attributes: map[string]*Attribute{"kind": {Optional: true}},
						Blocks: map[string]*NestedBlock{
							"and": {NestingMode: "single", Block: &Block{
								Blocks: map[string]*NestedBlock{
									"statement": {NestingMode: "list", Block: &Block{
										Attributes: map[string]*Attribute{"kind": {Optional: true}},
										Blocks: map[string]*NestedBlock{
											"and": {NestingMode: "single", Block: &Block{
												Blocks: map[string]*NestedBlock{
													"statement": {NestingMode: "list", Block: &Block{
														Attributes: map[string]*Attribute{"kind": {Optional: true}},
													}},
												},
											}},
										},
									}},
								},
							}},
						},
					}},
				},
			},
		},
	}}
	// The spec models the grammar recursively: the census emits the root
	// arm's leaf plus the re-entry field as ONE leaf.
	spec := []KindCensus{{
		Kind: "TestRecursive",
		SpecFieldPaths: []string{
			"spec.statement.and.statements",
			"spec.statement.kind",
		},
	}}
	modules := []ModuleCensus{{
		Kind:      "TestRecursive",
		ModuleDir: "catalog/aws/testrecursive/iac/tf",
		Resources: []string{"aws_recursive_thing"},
	}}
	manifest := &Manifest{Resources: map[string]*ResourceManifest{
		"aws_recursive_thing": {
			Mappings: []Mapping{
				{Arg: "statement", Spec: "spec.statement"},
				{Arg: "statement.and.statement", Spec: "spec.statement.and.statements", Collapse: true},
			},
		},
	}}

	acc := buildAccounting("aws", spec, modules, schemas, "aws",
		map[string]*Manifest{"TestRecursive": manifest}, nil, nil)
	ka := kindByName(t, acc, "TestRecursive")
	// statement.kind (mapped via subtree), statement.and.statement.kind and
	// statement.and.statement.and.statement.kind (both collapse-mapped).
	if ka.TotalArgs != 3 || ka.MappedArgs != 3 {
		t.Errorf("total/mapped = %d/%d, want 3/3", ka.TotalArgs, ka.MappedArgs)
	}
	if len(ka.UnaccountedArgs) != 0 {
		t.Errorf("UnaccountedArgs = %v, want none", ka.UnaccountedArgs)
	}
	if len(ka.UncoveredSpecFields) != 0 {
		t.Errorf("UncoveredSpecFields = %v, want none (the collapse covers the re-entry leaf)", ka.UncoveredSpecFields)
	}
	if !ka.Accounted() {
		t.Error("TestRecursive must be at total accounting")
	}

	// The same mapping WITHOUT collapse leaves the deep args unaccounted —
	// proving the fold never happens implicitly.
	manifest.Resources["aws_recursive_thing"].Mappings[1].Collapse = false
	acc = buildAccounting("aws", spec, modules, schemas, "aws",
		map[string]*Manifest{"TestRecursive": manifest}, nil, nil)
	ka = kindByName(t, acc, "TestRecursive")
	if len(ka.UnaccountedArgs) != 2 {
		t.Errorf("UnaccountedArgs = %v, want the two deep args", ka.UnaccountedArgs)
	}

	// A collapse mapping pointing at a spec SUBTREE (not a leaf) is a
	// staleness finding: the fold target must be a single census leaf.
	manifest.Resources["aws_recursive_thing"].Mappings[1] = Mapping{
		Arg: "statement.and.statement", Spec: "spec.statement", Collapse: true,
	}
	acc = buildAccounting("aws", spec, modules, schemas, "aws",
		map[string]*Manifest{"TestRecursive": manifest}, nil, nil)
	ka = kindByName(t, acc, "TestRecursive")
	var sawLeafStale bool
	for _, s := range ka.ManifestStale {
		if strings.Contains(s, "must name a spec census leaf") {
			sawLeafStale = true
		}
	}
	if !sawLeafStale {
		t.Errorf("ManifestStale = %v, want the collapse-to-subtree finding", ka.ManifestStale)
	}
}

func TestBuildAccounting_LedgerShadowsComputedClasses(t *testing.T) {
	spec, modules, schemas, manifests, _ := accountingFixture()
	ledger := []LedgerEntry{
		{Resource: "google_widget_iam_policy", Disposition: DispositionDeferred, Reason: "stale: the iam-covered class computes this"},
		{Resource: "google_dead_thing", Disposition: DispositionExcludedDeprecated, Reason: "stale: the schema flag computes this"},
	}
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, ledger, nil)

	byResource := map[string]ResourceDisposition{}
	for _, d := range acc.Dispositions {
		byResource[d.Resource] = d
	}
	if got := byResource["google_widget_iam_policy"].Disposition; got != DispositionIamCovered {
		t.Errorf("iam triplet disposition = %q, want the computed %q", got, DispositionIamCovered)
	}
	if got := byResource["google_dead_thing"].Disposition; got != DispositionExcludedDeprecated {
		t.Errorf("deprecated disposition = %q, want the computed %q", got, DispositionExcludedDeprecated)
	}

	wantStale := map[string]string{
		"resource:google_widget_iam_policy": "iam-covered",
		"resource:google_dead_thing":        "doc-level deprecations",
	}
	for key, fragment := range wantStale {
		found := false
		for _, f := range acc.Findings {
			if f.BaselineKey == key && strings.Contains(f.Detail, "stale ledger entry") && strings.Contains(f.Detail, fragment) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing shadowing finding for %s (want detail mentioning %q); findings: %v", key, fragment, acc.Findings)
		}
	}
}

// TestBuildAccounting_MissingModule proves a kind with no Terraform module
// directory surfaces as exactly one finding -- never a run-abort, never a
// spec-field noise storm, never counted as accounted.
func TestBuildAccounting_MissingModule(t *testing.T) {
	spec, modules, schemas, manifests, _ := accountingFixture()
	modules = append(modules, ModuleCensus{
		Kind: "TestGhost", ModuleDir: "catalog/gcp/testghost/iac/tf", MissingModule: true,
	})
	spec = append(spec, KindCensus{Kind: "TestGhost", SpecFieldPaths: []string{"spec.a", "spec.b"}})
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, nil, nil)

	ghost := kindByName(t, acc, "TestGhost")
	if !ghost.MissingModule || ghost.Accounted() {
		t.Errorf("ghost = %+v, want MissingModule and not accounted", ghost)
	}
	if len(ghost.UncoveredSpecFields) != 0 {
		t.Errorf("missing-module kind must not accumulate spec-field noise, got %v", ghost.UncoveredSpecFields)
	}
	var ghostFindings []Finding
	for _, f := range acc.Findings {
		if f.BaselineKey == "kind:TestGhost" {
			ghostFindings = append(ghostFindings, f)
		}
	}
	if len(ghostFindings) != 1 || !strings.Contains(ghostFindings[0].Detail, "no Terraform module directory") {
		t.Errorf("ghost findings = %v, want exactly the missing-module finding", ghostFindings)
	}
	// The other kinds' accounting is unaffected by the ghost.
	if widget := kindByName(t, acc, "TestWidget"); widget.TotalArgs != 12 {
		t.Errorf("TestWidget accounting disturbed by the missing-module kind: TotalArgs=%d", widget.TotalArgs)
	}
}

// TestBuildAccounting_ExternalResources proves the externally-specified
// judgment end to end: an external-dispositioned resource carries no
// argument walk and no unknown-resource finding, a kind consuming ONLY
// external resources runs no reverse spec walk (its depth is accounted
// against the external contract the admission records), a MIXED kind keeps
// the reverse walk for its schema-served side, an external judgment on a
// schema-served resource is a staleness finding (the provider shipped
// native support -- the exit-to-native ratchet), and an external resource
// WITHOUT the judgment keeps the unknown-resource finding.
func TestBuildAccounting_ExternalResources(t *testing.T) {
	spec, modules, schemas, manifests, _ := accountingFixture()
	spec = append(spec,
		KindCensus{Kind: "TestArm", SpecFieldPaths: []string{"spec.auth_mode", "spec.traffic"}},
		KindCensus{Kind: "TestMixed", SpecFieldPaths: []string{"spec.plain_name", "spec.arm_only_field"}},
		KindCensus{Kind: "TestArmUnjudged", SpecFieldPaths: []string{"spec.x"}},
	)
	modules = append(modules,
		ModuleCensus{Kind: "TestArm", Resources: []string{"azapi_resource"}, Pins: map[string]string{"azapi": "2.11.0"}},
		ModuleCensus{Kind: "TestMixed", Resources: []string{"google_plain", "azapi_resource"}, Pins: map[string]string{"google": "~> 6.0", "azapi": "2.11.0"}},
		ModuleCensus{Kind: "TestArmUnjudged", Resources: []string{"azapi_resource"}, Pins: map[string]string{"azapi": "2.11.0"}},
	)
	external := &ResourceManifest{External: "raw-ARM surface at a pinned type@api-version by recorded admission"}
	manifests["TestArm"] = &Manifest{Resources: map[string]*ResourceManifest{"azapi_resource": external}}
	manifests["TestMixed"] = &Manifest{Resources: map[string]*ResourceManifest{
		"azapi_resource": external,
		"google_plain":   {Mappings: []Mapping{{Spec: "spec.plain_name", Arg: "name"}}},
	}}
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, nil, nil)

	arm := kindByName(t, acc, "TestArm")
	if !arm.Accounted() {
		t.Errorf("external-only kind must be at total accounting, got %+v", arm)
	}
	if !reflect.DeepEqual(arm.ExternalResources, []string{"azapi_resource"}) {
		t.Errorf("ExternalResources = %v, want [azapi_resource]", arm.ExternalResources)
	}
	if len(arm.UncoveredSpecFields) != 0 {
		t.Errorf("external-only kind must suspend the reverse walk, got uncovered %v", arm.UncoveredSpecFields)
	}

	// The mixed kind keeps the reverse walk: its schema-served field is
	// covered by the mapping; the external-fed field is flagged until a
	// specExclusion names the external resource.
	mixed := kindByName(t, acc, "TestMixed")
	if !reflect.DeepEqual(mixed.UncoveredSpecFields, []string{"spec.arm_only_field"}) {
		t.Errorf("mixed kind uncovered = %v, want [spec.arm_only_field]", mixed.UncoveredSpecFields)
	}

	// Without the judgment, the unknown-resource finding stands unchanged.
	unjudged := kindByName(t, acc, "TestArmUnjudged")
	if unjudged.Accounted() || len(unjudged.ManifestStale) != 1 ||
		!strings.Contains(unjudged.ManifestStale[0], "unknown to every loaded schema") {
		t.Errorf("unjudged external consumption must keep the unknown-resource finding, got %+v", unjudged)
	}
}

// TestBuildAccounting_SecondaryChannelAdmissions proves the admission gate
// end to end against a fixture with a GA schema and a secondary channel
// ("google-beta") that serves one resource the GA schema does not:
//
//   - admitted + attached: the resource is argument-walked against the
//     channel's schema and listed as admitted; the kind is at total
//     accounting.
//   - consumed with no admission: an admission gap (the silent fallback
//     this gate exists to end).
//   - admitted but not attached through the channel's provider: a gap (the
//     module would not actually deploy through the channel).
//   - attached to the channel although GA serves the resource: a gap (beta
//     reached for without need).
//   - an admission for a resource GA serves: a stale-admission finding; an
//     admission whose kind does not consume the resource: stale too.
//   - an argument set in the admitted channel's provider block joins the
//     provider-config judgment exactly like one set on the baseline block.
func TestBuildAccounting_SecondaryChannelAdmissions(t *testing.T) {
	spec, modules, schemas, manifests, _ := accountingFixture()
	schemas["google-beta"] = &Schema{
		Provider: "google-beta",
		Version:  "6.0.0",
		Resources: map[string]*Block{
			// Beta-only: what the admission list is for.
			"google_beta_only": {Attributes: map[string]*Attribute{
				"name": {Type: json.RawMessage(`"string"`), Required: true},
			}},
			// Served by both channels: GA is always the yardstick.
			"google_plain": schemas["google"].Resources["google_plain"],
		},
	}
	spec = append(spec,
		KindCensus{Kind: "TestAdmitted", SpecFieldPaths: []string{"spec.name"}},
		KindCensus{Kind: "TestUnadmitted", SpecFieldPaths: []string{"spec.name"}},
		KindCensus{Kind: "TestUnattached", SpecFieldPaths: []string{"spec.name"}},
		KindCensus{Kind: "TestNeedlessBeta", SpecFieldPaths: []string{"spec.plain_name"}},
	)
	betaPins := map[string]string{"google": "~> 6.0", "google-beta": "~> 6.0"}
	modules = append(modules,
		ModuleCensus{Kind: "TestAdmitted", Resources: []string{"google_beta_only"}, Pins: betaPins,
			ProviderAttachments: map[string]string{"google_beta_only": "google-beta"}},
		ModuleCensus{Kind: "TestUnadmitted", Resources: []string{"google_beta_only"}, Pins: betaPins,
			ProviderAttachments: map[string]string{"google_beta_only": "google-beta"}},
		ModuleCensus{Kind: "TestUnattached", Resources: []string{"google_beta_only"}, Pins: betaPins},
		ModuleCensus{Kind: "TestNeedlessBeta", Resources: []string{"google_plain"}, Pins: betaPins,
			ProviderAttachments: map[string]string{"google_plain": "google-beta"}},
	)
	admissions := []AdmissionEntry{
		{Provider: "google-beta", Resource: "google_beta_only", Kind: "TestAdmitted", Reason: "beta-only at the pin", PromotionTracking: "provider repo, services/betaonly"},
		{Provider: "google-beta", Resource: "google_beta_only", Kind: "TestUnattached", Reason: "beta-only at the pin", PromotionTracking: "provider repo, services/betaonly"},
		// Stale: GA serves google_plain.
		{Provider: "google-beta", Resource: "google_plain", Kind: "TestNeedlessBeta", Reason: "was beta once", PromotionTracking: "promoted"},
		// Stale: TestPlain's module does not consume the resource.
		{Provider: "google-beta", Resource: "google_beta_only", Kind: "TestPlain", Reason: "nobody attaches it", PromotionTracking: "n/a"},
	}
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, nil, admissions)

	admitted := kindByName(t, acc, "TestAdmitted")
	if !admitted.Accounted() {
		t.Errorf("admitted+attached kind must be at total accounting, got %+v", admitted)
	}
	if !reflect.DeepEqual(admitted.AdmittedResources, []string{"google_beta_only"}) || admitted.TotalArgs != 1 || admitted.MatchedArgs != 1 {
		t.Errorf("admitted kind must walk the channel schema: %+v", admitted)
	}

	unadmitted := kindByName(t, acc, "TestUnadmitted")
	if unadmitted.Accounted() || len(unadmitted.AdmissionGaps) != 1 || !strings.Contains(unadmitted.AdmissionGaps[0], "carries no admission") {
		t.Errorf("unadmitted secondary-channel consumption must be a gap, got %+v", unadmitted)
	}
	// The gap never hides the depth accounting behind it.
	if unadmitted.TotalArgs != 1 {
		t.Errorf("gap must still walk the arguments, TotalArgs=%d", unadmitted.TotalArgs)
	}

	unattached := kindByName(t, acc, "TestUnattached")
	if unattached.Accounted() || len(unattached.AdmissionGaps) != 1 || !strings.Contains(unattached.AdmissionGaps[0], "set `provider = google-beta`") {
		t.Errorf("admitted-but-unattached must be a gap, got %+v", unattached)
	}

	needless := kindByName(t, acc, "TestNeedlessBeta")
	if needless.Accounted() || len(needless.AdmissionGaps) != 1 || !strings.Contains(needless.AdmissionGaps[0], "baseline schema google serves it") {
		t.Errorf("beta attachment on a GA-served resource must be a gap, got %+v", needless)
	}

	var staleGA, staleUnconsumed bool
	for _, f := range acc.Findings {
		if f.BaselineKey == "admission:google_plain" && strings.Contains(f.Detail, "served by the baseline schema") {
			staleGA = true
		}
		if f.BaselineKey == "admission:google_beta_only" && strings.Contains(f.Detail, "TestPlain's module does not consume") {
			staleUnconsumed = true
		}
	}
	if !staleGA || !staleUnconsumed {
		t.Errorf("stale admissions must be findings (GA-served=%v, unconsumed=%v): %v", staleGA, staleUnconsumed, acc.Findings)
	}
	if !reflect.DeepEqual(acc.Admissions, admissions) {
		t.Errorf("the accounting must carry the admission list for the report")
	}
}

// TestBuildAccounting_ExternalJudgmentGoesStale proves the exit-to-native
// ratchet: when a loaded schema starts serving a resource the manifest
// still calls external, the judgment is a staleness finding, never a
// silent double-accounting.
func TestBuildAccounting_ExternalJudgmentGoesStale(t *testing.T) {
	spec, modules, schemas, manifests, _ := accountingFixture()
	spec = append(spec, KindCensus{Kind: "TestNative", SpecFieldPaths: []string{"spec.plain_name"}})
	modules = append(modules, ModuleCensus{Kind: "TestNative", Resources: []string{"google_plain"}})
	manifests["TestNative"] = &Manifest{Resources: map[string]*ResourceManifest{
		"google_plain": {External: "stale: the schema serves this resource now"},
	}}
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, nil, nil)

	native := kindByName(t, acc, "TestNative")
	if native.Accounted() {
		t.Error("stale external judgment must fail accounting")
	}
	found := false
	for _, s := range native.ManifestStale {
		if strings.Contains(s, "external judgment is stale") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ManifestStale = %v, want the stale-external finding", native.ManifestStale)
	}
}

// TestBuildAccounting_ManifestStaleness proves the ratchet: judgment
// referencing surface that no longer exists fails, in every reference class.
func TestBuildAccounting_ManifestStaleness(t *testing.T) {
	spec, modules, schemas, _, _ := accountingFixture()
	manifests := map[string]*Manifest{
		"TestPlain": {
			Resources: map[string]*ResourceManifest{
				"google_plain": {
					Mappings: []Mapping{
						{Spec: "spec.plain_name", Arg: "name"},     // valid
						{Spec: "spec.gone_field", Arg: "name2"},    // both sides stale
						{Spec: "spec.plain_name", Arg: "gone_arg"}, // arg side stale
					},
					Exclusions: []ArgExclusion{
						{Arg: "vanished", Reason: "stale exclusion"},
					},
				},
				"google_never_consumed": {
					Internal: "stale: the module no longer consumes this",
				},
			},
			SpecExclusions: []SpecExclusion{
				{Field: "spec.gone_platform_field", Reason: "stale spec exclusion"},
			},
		},
	}
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, nil, nil)
	plain := kindByName(t, acc, "TestPlain")

	wantStale := []string{
		"mapping arg gone_arg",
		"mapping arg name2", // reported via its arg; its spec side is also stale
		"mapping spec spec.gone_field",
		"exclusion vanished",
		"manifest judges resource google_never_consumed",
		"specExclusions: spec.gone_platform_field",
	}
	for _, want := range wantStale {
		found := false
		for _, s := range plain.ManifestStale {
			if strings.Contains(s, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ManifestStale misses %q; got %v", want, plain.ManifestStale)
		}
	}
	// The valid mapping still matched.
	if plain.MappedArgs != 1 || len(plain.UnaccountedArgs) != 0 {
		t.Errorf("mapped/unaccounted = %d/%v, want 1/none", plain.MappedArgs, plain.UnaccountedArgs)
	}
}

// A spec leaf the proto marks manifest_only reaches no provider argument by
// design, and the schema itself declares that: the reverse check excludes it
// with no specExclusions entry, and an entry that repeats the fact is stale.
func TestBuildAccounting_ManifestOnlyLeafIsExcludedBySchema(t *testing.T) {
	spec, modules, schemas, _, _ := accountingFixture()
	for i := range spec {
		if spec[i].Kind == "TestPlain" {
			spec[i].SpecFieldPaths = append(spec[i].SpecFieldPaths, "spec.select_all")
			spec[i].ManifestOnlyPaths = []string{"spec.select_all"}
		}
	}
	manifest := &Manifest{Resources: map[string]*ResourceManifest{
		"google_plain": {Mappings: []Mapping{{Spec: "spec.plain_name", Arg: "name"}}},
	}}

	acc := buildAccounting("gcp", spec, modules, schemas, "google", map[string]*Manifest{"TestPlain": manifest}, nil, nil)
	plain := kindByName(t, acc, "TestPlain")
	if len(plain.UncoveredSpecFields) != 0 {
		t.Errorf("a manifest_only leaf needs no specExclusions entry; uncovered = %v", plain.UncoveredSpecFields)
	}
	if len(plain.ManifestStale) != 0 {
		t.Errorf("nothing stale expected; got %v", plain.ManifestStale)
	}

	manifest.SpecExclusions = []SpecExclusion{{Field: "spec.select_all", Reason: "repeats the proto"}}
	acc = buildAccounting("gcp", spec, modules, schemas, "google", map[string]*Manifest{"TestPlain": manifest}, nil, nil)
	plain = kindByName(t, acc, "TestPlain")
	found := false
	for _, s := range plain.ManifestStale {
		if strings.Contains(s, "spec.select_all is manifest_only") {
			found = true
		}
	}
	if !found {
		t.Errorf("a specExclusions entry for a manifest_only leaf must be reported stale; got %v", plain.ManifestStale)
	}
}

// A consumed resource no loaded schema knows (a utility provider like
// hashicorp/time) is legal exactly when the manifest judges it internal --
// the judgment is "module plumbing, no provider surface to account", so no
// schema is needed. Without the judgment the stale-artifact guard fires.
func TestBuildAccounting_InternalUtilityResourceNeedsNoSchema(t *testing.T) {
	spec, _, schemas, _, _ := accountingFixture()

	modules := []ModuleCensus{
		{Kind: "TestPlain", Resources: []string{"google_plain", "time_sleep"}, Pins: map[string]string{"google": "~> 6.0"}},
	}

	// Judged internal: no stale finding, listed as internal.
	manifests := map[string]*Manifest{
		"TestPlain": {
			Resources: map[string]*ResourceManifest{
				"time_sleep": {Internal: "destroy-ordering utility from the hashicorp/time provider -- no provider surface"},
			},
		},
	}
	acc := buildAccounting("gcp", spec, modules, schemas, "google", manifests, nil, nil)
	plain := kindByName(t, acc, "TestPlain")
	for _, s := range plain.ManifestStale {
		if strings.Contains(s, "time_sleep") {
			t.Errorf("internal utility resource flagged stale: %v", plain.ManifestStale)
		}
	}
	if len(plain.InternalResources) != 1 || plain.InternalResources[0] != "time_sleep" {
		t.Errorf("InternalResources = %v, want [time_sleep]", plain.InternalResources)
	}

	// Unjudged: the stale-artifact guard still fires.
	acc = buildAccounting("gcp", spec, modules, schemas, "google", map[string]*Manifest{}, nil, nil)
	plain = kindByName(t, acc, "TestPlain")
	found := false
	for _, s := range plain.ManifestStale {
		if strings.Contains(s, "time_sleep is unknown to every loaded schema") {
			found = true
		}
	}
	if !found {
		t.Errorf("unjudged unknown resource not flagged; got %v", plain.ManifestStale)
	}
}

func TestGateSemantics(t *testing.T) {
	findings := []Finding{
		{BaselineKey: "kind:TestPlain", Detail: "unaccounted provider argument"},
		{BaselineKey: "resource:google_orphan_thing", Detail: "no recorded disposition"},
	}
	baseline := map[string]bool{
		"kind:TestPlain":               true,
		"resource:google_orphan_thing": true,
	}
	if res := Gate(findings, baseline); !res.OK() {
		t.Errorf("baselined findings must pass, got %+v", res)
	}

	// A new finding fails.
	novel := append(findings, Finding{BaselineKey: "kind:TestWidget", Detail: "new drift"})
	res := Gate(novel, baseline)
	if len(res.NewFindings) != 1 || res.NewFindings[0].BaselineKey != "kind:TestWidget" {
		t.Errorf("NewFindings = %+v", res.NewFindings)
	}

	// A fixed-but-still-listed entry fails too: the ledger only shrinks
	// truthfully.
	res = Gate(findings[:1], baseline)
	if len(res.StaleEntries) != 1 || res.StaleEntries[0] != "resource:google_orphan_thing" {
		t.Errorf("StaleEntries = %v", res.StaleEntries)
	}
}

func TestBaselineRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baseline.yaml")
	findings := []Finding{
		{BaselineKey: "kind:B", Detail: "x"},
		{BaselineKey: "kind:A", Detail: "y"},
		{BaselineKey: "kind:A", Detail: "z"}, // duplicate key collapses to one entry
	}
	if err := WriteBaseline(path, findings); err != nil {
		t.Fatal(err)
	}
	set, err := LoadBaseline(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 2 || !set["kind:A"] || !set["kind:B"] {
		t.Errorf("round trip = %v", set)
	}
}

func TestLoadLedger(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "google.yaml")

	if entries, err := LoadLedger(path, "google"); err != nil || entries != nil {
		t.Errorf("missing ledger must be empty, got %v / %v", entries, err)
	}

	valid := `provider: google
resources:
  - resource: google_composed_thing
    disposition: composed
    reason: covered by TestWidget
`
	if err := os.WriteFile(path, []byte(valid), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := LoadLedger(path, "google")
	if err != nil || len(entries) != 1 {
		t.Fatalf("valid ledger: %v / %v", entries, err)
	}

	for name, bad := range map[string]string{
		"wrong provider":     "provider: azurerm\nresources: []\n",
		"computed class":     "provider: google\nresources:\n  - resource: r\n    disposition: modeled\n    reason: x\n",
		"missing reason":     "provider: google\nresources:\n  - resource: r\n    disposition: deferred\n",
		"duplicate resource": "provider: google\nresources:\n  - {resource: r, disposition: deferred, reason: a}\n  - {resource: r, disposition: composed, reason: b}\n",
		"unknown field":      "provider: google\nresourcez: []\n",
	} {
		if err := os.WriteFile(path, []byte(bad), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadLedger(path, "google"); err == nil {
			t.Errorf("%s: want error, got nil", name)
		}
	}
}

func TestLoadAdmissions(t *testing.T) {
	dir := t.TempDir()

	if entries, err := LoadAdmissions(filepath.Join(dir, "absent"), "google"); err != nil || entries != nil {
		t.Errorf("missing admissions dir must be empty, got %v / %v", entries, err)
	}

	path := filepath.Join(dir, "google-beta.yaml")
	valid := `provider: google-beta
gaSchema: google
resources:
  - resource: google_firebase_project
    kind: GcpFirebaseProject
    reason: published only in the beta provider at the pin
    promotionTracking: hashicorp/terraform-provider-google, google/services/firebase/
`
	if err := os.WriteFile(path, []byte(valid), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := LoadAdmissions(dir, "google")
	if err != nil || len(entries) != 1 || entries[0].Provider != "google-beta" || entries[0].Kind != "GcpFirebaseProject" {
		t.Fatalf("valid admissions: %+v / %v", entries, err)
	}
	// A list supplementing a different baseline is skipped, not an error.
	if entries, err := LoadAdmissions(dir, "aws"); err != nil || len(entries) != 0 {
		t.Errorf("admissions for another baseline must be skipped, got %v / %v", entries, err)
	}

	for name, bad := range map[string]string{
		"no provider":             "gaSchema: google\nresources: []\n",
		"no gaSchema":             "provider: google-beta\nresources: []\n",
		"admits itself":           "provider: google\ngaSchema: google\nresources: []\n",
		"missing kind":            "provider: google-beta\ngaSchema: google\nresources:\n  - {resource: r, reason: a, promotionTracking: p}\n",
		"missing reason":          "provider: google-beta\ngaSchema: google\nresources:\n  - {resource: r, kind: K, promotionTracking: p}\n",
		"missing tracking":        "provider: google-beta\ngaSchema: google\nresources:\n  - {resource: r, kind: K, reason: a}\n",
		"duplicate resource+kind": "provider: google-beta\ngaSchema: google\nresources:\n  - {resource: r, kind: K, reason: a, promotionTracking: p}\n  - {resource: r, kind: K, reason: b, promotionTracking: p}\n",
		"unknown field":           "provider: google-beta\ngaSchema: google\nresourcez: []\n",
	} {
		if err := os.WriteFile(path, []byte(bad), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadAdmissions(dir, "google"); err == nil {
			t.Errorf("%s: want error, got nil", name)
		}
	}
}

// TestProviderParityGate is the live gate over every enrolled catalog: each
// provider's accounting runs against the committed schemas, manifests,
// ledger, and the ONE shared baseline, and any drift from the recorded state
// fails. Gating and regenerating always use the merged findings of all
// enrollments -- a single provider's findings against the shared baseline
// would misreport every other provider's entries as stale. This test IS the
// CI lane (lint.provider-parity.yaml).
func TestProviderParityGate(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, catalogRoot)); err != nil {
		t.Skip("catalog source tree not present (bazel sandbox); runs under go test")
	}
	schemas, err := LoadSchemas("schemas")
	if err != nil {
		t.Fatalf("committed schemas: %v", err)
	}
	accountings, err := EnrolledAccountings(root, schemas)
	if err != nil {
		t.Fatalf("accounting: %v", err)
	}

	for _, acc := range accountings {
		accounted := 0
		for _, k := range acc.Kinds {
			if k.Accounted() {
				accounted++
			}
		}
		t.Logf("%s depth accounting: %d/%d kinds at total accounting against %s@%s",
			acc.CloudProvider, accounted, len(acc.Kinds), acc.GASchema, acc.GASchemaVersion)
		t.Logf("%s breadth dispositions: %v", acc.CloudProvider, acc.DispositionTotals)

		if raw, err := json.MarshalIndent(acc, "", "  "); err == nil && os.Getenv("PLANTON_PROVIDERPARITY_DUMP") != "" {
			path := filepath.Join(os.TempDir(), "providerparity-"+acc.CloudProvider+"-accounting.json")
			if err := os.WriteFile(path, raw, 0o644); err == nil {
				t.Logf("full accounting written to %s", path)
			}
		}
	}

	findings := MergeFindings(accountings)
	baselinePath := filepath.Join(root, DefaultBaselinePath)
	if os.Getenv("PLANTON_REGEN_PROVIDERPARITY_BASELINE") == "1" {
		if err := WriteBaseline(baselinePath, findings); err != nil {
			t.Fatalf("regenerating baseline: %v", err)
		}
		t.Log("baseline regenerated -- review the diff before committing")
	}
	baseline, err := LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("baseline: %v (regenerate with PLANTON_REGEN_PROVIDERPARITY_BASELINE=1)", err)
	}
	res := Gate(findings, baseline)
	for _, f := range res.NewFindings {
		t.Errorf("new parity gap [%s]: %s", f.BaselineKey, f.Detail)
	}
	for _, key := range res.StaleEntries {
		t.Errorf("stale baseline entry (closed -- remove it): %s", key)
	}
}

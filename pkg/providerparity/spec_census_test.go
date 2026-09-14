//go:build !codegen
// +build !codegen

package providerparity

import (
	"reflect"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// TestCollectSpecPaths_HermeticFixture proves the descriptor walk against the
// permanent testcloudresourcegeneric fixture at its REGISTERED version
// (v1alpha2 -- the registry serves a kind's declared version, so the walk
// must be asserted against the compiled contract, not a .proto read by eye):
// scalars and optional scalars are leaves, a nested message contributes its
// own leaves (never itself), a repeated message recurses into its element
// (steps -> steps.command), a StringValueOrRef is ONE leaf (the author
// configures one value slot), and a map field is one leaf.
func TestCollectSpecPaths_HermeticFixture(t *testing.T) {
	msg, err := crkreflect.NewInstance(cloudresourcekind.CloudResourceKind_TestCloudResourceGeneric)
	if err != nil {
		t.Fatalf("new instance: %v", err)
	}
	specField := msg.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specField == nil {
		t.Fatal("fixture has no spec field")
	}

	got := CollectSpecPaths(specField.Message(), "spec")
	want := []string{
		"spec.annotated_ref",
		"spec.bool_field",
		"spec.display_name",
		"spec.double_field",
		"spec.float_field",
		"spec.int32_field",
		"spec.int64_field",
		"spec.labels",
		"spec.nested.nested_int",
		"spec.nested.nested_string",
		"spec.optional_ref",
		"spec.replicas",
		"spec.required_ref",
		"spec.sensitive_ref",
		"spec.sensitive_string",
		"spec.steps.command",
		"spec.uint32_field",
		"spec.uint64_field",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("paths mismatch:\n got %v\nwant %v", got, want)
	}
}

// TestCollectSpecCensus_ManifestOnlyLeafIsRecorded proves the walk records a
// field carrying (dev.planton.shared.options.manifest_only) in both outputs:
// it is authored surface (so it stays a spec leaf and counts) and it is a
// word no engine forwards (so the reverse check reads it as excluded by the
// schema). Its unmarked siblings appear in the first output only.
func TestCollectSpecCensus_ManifestOnlyLeafIsRecorded(t *testing.T) {
	msg, err := crkreflect.NewInstance(cloudresourcekind.CloudResourceKind_TestCloudResourceKubernetes)
	if err != nil {
		t.Fatalf("new instance: %v", err)
	}
	specField := msg.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specField == nil {
		t.Fatal("fixture has no spec field")
	}

	paths, manifestOnly := CollectSpecCensus(specField.Message(), "spec")
	if !reflect.DeepEqual(manifestOnly, []string{"spec.select_all"}) {
		t.Errorf("manifest-only leaves = %v, want [spec.select_all]", manifestOnly)
	}
	hasSelectAll, hasSibling := false, false
	for _, p := range paths {
		switch p {
		case "spec.select_all":
			hasSelectAll = true
		case "spec.create_namespace":
			hasSibling = true
		}
	}
	if !hasSelectAll || !hasSibling {
		t.Errorf("spec paths must carry both the marked leaf and its sibling; got %v", paths)
	}
}

// TestCollectSpecPaths_RecursiveReentryIsOneLeaf proves the recursive
// re-entry rule against a live recursive spec (AwsWafWebAcl's statement
// tree): a field whose message type is an ancestor on the walk path counts
// as ONE leaf (the author configures "the same grammar again" there), and
// the walk never descends into it. The previous behavior dropped such
// fields silently, hiding them from the census and the reverse-drift check.
func TestCollectSpecPaths_RecursiveReentryIsOneLeaf(t *testing.T) {
	msg, err := crkreflect.NewInstance(cloudresourcekind.CloudResourceKind_AwsWafWebAcl)
	if err != nil {
		t.Fatalf("new instance: %v", err)
	}
	specField := msg.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specField == nil {
		t.Fatal("AwsWafWebAcl has no spec field")
	}

	got := CollectSpecPaths(specField.Message(), "spec")
	leaves := map[string]bool{}
	for _, p := range got {
		leaves[p] = true
	}
	wantLeaves := []string{
		"spec.rules.statement.and_statement.statements",
		"spec.rules.statement.or_statement.statements",
		"spec.rules.statement.not_statement.statement",
		"spec.rules.statement.managed_rule_group.scope_down_statement",
		"spec.rules.statement.rate_based.scope_down_statement",
	}
	for _, want := range wantLeaves {
		if !leaves[want] {
			t.Errorf("census must record the recursive re-entry %s as a leaf", want)
		}
		for _, p := range got {
			if strings.HasPrefix(p, want+".") {
				t.Errorf("census must not descend below the re-entry leaf %s (got %s)", want, p)
			}
		}
	}
}

// TestCollectSpecPaths_StructIsOneLeaf proves the well-known-value-type
// rule against a live spec that carries a google.protobuf.Struct field
// (AwsEventBridgeRule's event_pattern): the Struct is ONE leaf the author
// writes as YAML -- the walk must never descend into its descriptor
// internals (fields.bool_value and siblings are protobuf plumbing, not
// configurable surface).
func TestCollectSpecPaths_StructIsOneLeaf(t *testing.T) {
	msg, err := crkreflect.NewInstance(cloudresourcekind.CloudResourceKind_AwsEventBridgeRule)
	if err != nil {
		t.Fatalf("new instance: %v", err)
	}
	specField := msg.ProtoReflect().Descriptor().Fields().ByName("spec")
	if specField == nil {
		t.Fatal("AwsEventBridgeRule has no spec field")
	}

	got := CollectSpecPaths(specField.Message(), "spec")
	var sawLeaf bool
	for _, p := range got {
		if p == "spec.event_pattern" {
			sawLeaf = true
			continue
		}
		if len(p) > len("spec.event_pattern") && p[:len("spec.event_pattern.")] == "spec.event_pattern." {
			t.Errorf("walk descended into the Struct's internals: %s", p)
		}
	}
	if !sawLeaf {
		t.Error("spec.event_pattern is missing from the census -- the Struct field should be one leaf")
	}
}

// TestSpecCensusGcp is the live-catalog smoke test: every implemented GCP
// kind is censused and none has an empty spec surface (a kind an author
// cannot configure at all would be a registry error, not a real component).
func TestSpecCensusGcp(t *testing.T) {
	census := SpecCensus(cloudresourcekind.CloudResourceProvider_gcp)
	if len(census) == 0 {
		t.Fatal("GCP spec census is empty -- the registry walk is broken")
	}
	for _, k := range census {
		if len(k.SpecFieldPaths) == 0 {
			t.Errorf("%s: zero spec leaf fields", k.Kind)
		}
	}
}

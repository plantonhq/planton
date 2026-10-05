package explain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func fieldByName(fields []Field, name string) *Field {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}

func mustResource(t *testing.T, name string) Resource {
	t.Helper()
	res, err := ResolveKindName(name)
	if err != nil {
		t.Fatalf("resolve %s: %v", name, err)
	}
	return res
}

func TestRootReport_AwsVpc(t *testing.T) {
	engine := DefaultEngine()
	report, err := engine.Explain(mustResource(t, "AwsVpc"), nil)
	if err != nil {
		t.Fatal(err)
	}

	if report.Kind != "AwsVpc" {
		t.Errorf("kind = %q", report.Kind)
	}
	if report.ApiVersion != "aws.planton.dev/v1alpha1" {
		t.Errorf("apiVersion = %q", report.ApiVersion)
	}
	if !strings.Contains(report.Doc, "CIDR") {
		t.Errorf("spec doc not carried onto the report: %q", report.Doc)
	}

	region := fieldByName(report.Spec, "region")
	if region == nil {
		t.Fatal("region missing from spec")
	}
	if !region.Required {
		t.Error("region should derive required from min_len")
	}
	if !strings.Contains(region.Doc, "region") {
		t.Errorf("region doc missing: %q", region.Doc)
	}
	if fieldByName(report.Spec, "cidrBlock") == nil {
		t.Error("spec fields must use protojson names (cidrBlock)")
	}

	if len(report.SpecRules) == 0 {
		t.Error("AwsVpc cross-field CEL rules missing")
	}

	vpcId := fieldByName(report.Outputs, "vpcId")
	if vpcId == nil {
		t.Fatal("vpcId missing from outputs")
	}
	if !strings.Contains(vpcId.Doc, "status.outputs.vpc_id") {
		t.Errorf("output doc missing: %q", vpcId.Doc)
	}
}

func TestForeignKeyField(t *testing.T) {
	engine := DefaultEngine()
	report, err := engine.Explain(mustResource(t, "aws-security-group"), nil)
	if err != nil {
		t.Fatal(err)
	}
	vpcId := fieldByName(report.Spec, "vpcId")
	if vpcId == nil {
		t.Fatal("vpcId missing from spec")
	}
	if vpcId.Type != "string | valueFrom" {
		t.Errorf("type = %q", vpcId.Type)
	}
	if vpcId.RefKind != "AwsVpc" || vpcId.RefFieldPath != "status.outputs.vpc_id" {
		t.Errorf("ref = %q %q", vpcId.RefKind, vpcId.RefFieldPath)
	}
	if len(vpcId.Fields) != 0 {
		t.Error("foreign-key wrapper must be terminal, not expanded")
	}

	// The YAML authoring contract must be spelled out -- a bare string does
	// not parse for these fields, and the wrapper is never expanded, so this
	// line is the only place an author learns the real shape.
	contract := false
	for _, c := range vpcId.Constraints {
		if strings.Contains(c, "write as {value:") && strings.Contains(c, "kind: AwsVpc") {
			contract = true
		}
	}
	if !contract {
		t.Errorf("valueFrom authoring contract missing: %v", vpcId.Constraints)
	}
}

func TestForeignKeyDrillDeadEndTeachesTheShape(t *testing.T) {
	engine := DefaultEngine()
	_, err := engine.Explain(mustResource(t, "aws-security-group"), []string{"spec", "vpcId", "value"})
	if err == nil || !strings.Contains(err.Error(), "write as {value:") || !strings.Contains(err.Error(), "kind: AwsVpc") {
		t.Fatalf("drill into a foreign-key wrapper should answer with the authoring shape, got: %v", err)
	}
}

func TestEnumValueDocs(t *testing.T) {
	engine := DefaultEngine()
	report, err := engine.Explain(mustResource(t, "kubernetes-namespace"), []string{"spec", "podSecurityStandard"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Field == nil {
		t.Fatal("field view expected")
	}
	if report.Path != "KubernetesNamespace.spec.podSecurityStandard" {
		t.Errorf("path = %q", report.Path)
	}
	var baseline *EnumValue
	for i := range report.Field.Enum {
		if report.Field.Enum[i].Name == "baseline" {
			baseline = &report.Field.Enum[i]
		}
	}
	if baseline == nil {
		t.Fatal("baseline enum value missing")
	}
	if !strings.Contains(baseline.Doc, "Minimally restrictive") {
		t.Errorf("enum value doc missing: %q", baseline.Doc)
	}
}

func TestPathResolution(t *testing.T) {
	engine := DefaultEngine()

	t.Run("nested field", func(t *testing.T) {
		report, err := engine.Explain(mustResource(t, "AwsVpc"), []string{"spec", "region"})
		if err != nil {
			t.Fatal(err)
		}
		if report.Field == nil || report.Field.Name != "region" {
			t.Fatalf("resolved field = %+v", report.Field)
		}
	})

	t.Run("envelope metadata resolves too", func(t *testing.T) {
		report, err := engine.Explain(mustResource(t, "AwsVpc"), []string{"metadata"})
		if err != nil {
			t.Fatal(err)
		}
		if report.Field == nil || len(report.Field.Fields) == 0 {
			t.Fatal("metadata should expand into its fields")
		}
	})

	t.Run("unknown segment names the valid drill point", func(t *testing.T) {
		_, err := engine.Explain(mustResource(t, "AwsVpc"), []string{"spec", "nope"})
		if err == nil || !strings.Contains(err.Error(), `no field "nope"`) {
			t.Fatalf("err = %v", err)
		}
		if !strings.Contains(err.Error(), "AwsVpc.spec") {
			t.Fatalf("error should point at the last resolvable path: %v", err)
		}
	})

	t.Run("scalar dead end", func(t *testing.T) {
		_, err := engine.Explain(mustResource(t, "AwsVpc"), []string{"spec", "region", "deeper"})
		if err == nil || !strings.Contains(err.Error(), "no fields to drill into") {
			t.Fatalf("err = %v", err)
		}
	})
}

// kindDispatcher is a test double for the kind-valued dispatch seam: it
// claims AwsVpc's instanceTenancy field and resolves the next segment as a
// catalog kind -- structurally identical to a platform envelope's
// kind-discriminated payload.
type kindDispatcher struct{}

func (kindDispatcher) Claims(fd protoreflect.FieldDescriptor) bool {
	return fd.JSONName() == "instanceTenancy"
}

func (kindDispatcher) Resolve(segment string) (Resource, error) {
	kind := catalogkindreflect.KindFromString(segment)
	if kind == catalogkind.CatalogKind_unspecified {
		return Resource{}, errors.Errorf("unknown kind %q", segment)
	}
	return KindResource(kind)
}

func (kindDispatcher) Hint() string { return "drill with .<kind> to see that kind's schema" }

func TestDispatcher(t *testing.T) {
	engine := DefaultEngine()
	engine.Dispatchers = []Dispatcher{kindDispatcher{}}

	t.Run("continues resolution across the boundary", func(t *testing.T) {
		report, err := engine.Explain(mustResource(t, "AwsVpc"), []string{"spec", "instanceTenancy", "aws-alb"})
		if err != nil {
			t.Fatal(err)
		}
		if report.Kind != "AwsAlb" {
			t.Errorf("dispatched kind = %q", report.Kind)
		}
		if report.Path != "AwsVpc.spec.instanceTenancy.aws-alb" {
			t.Errorf("path should stay anchored on what the user typed: %q", report.Path)
		}
		if len(report.Spec) == 0 {
			t.Error("dispatched report should carry the target's spec")
		}
	})

	t.Run("drills into the dispatched schema", func(t *testing.T) {
		report, err := engine.Explain(mustResource(t, "AwsVpc"),
			[]string{"spec", "instanceTenancy", "aws-security-group", "spec", "vpcId"})
		if err != nil {
			t.Fatal(err)
		}
		if report.Field == nil || report.Field.RefKind != "AwsVpc" {
			t.Fatalf("field across dispatch = %+v", report.Field)
		}
	})

	t.Run("bare claimed field carries the hint", func(t *testing.T) {
		report, err := engine.Explain(mustResource(t, "AwsVpc"), []string{"spec", "instanceTenancy"})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, c := range report.Field.Constraints {
			if strings.Contains(c, "drill with") {
				found = true
			}
		}
		if !found {
			t.Errorf("hint missing from constraints: %v", report.Field.Constraints)
		}
	})

	t.Run("unknown dispatch segment errors", func(t *testing.T) {
		_, err := engine.Explain(mustResource(t, "AwsVpc"), []string{"spec", "instanceTenancy", "not-a-kind"})
		if err == nil || !strings.Contains(err.Error(), "unknown kind") {
			t.Fatalf("err = %v", err)
		}
	})
}

// TestConstraintJSONDumpsAreCompact guards report determinism: the structural
// buf.validate dump comes from protojson, whose whitespace is deliberately
// randomized per compiled binary. Reports are committed and byte-compared by
// freshness gates, so every JSON constraint must survive json.Compact
// unchanged -- if this fails, applyValidateRules lost its compaction step.
func TestConstraintJSONDumpsAreCompact(t *testing.T) {
	engine := DefaultEngine()
	jsonDumps := 0
	for _, kind := range catalogkindreflect.KindsList() {
		res, err := KindResource(kind)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		report, err := engine.Explain(res, nil)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		var walk func(fields []Field)
		walk = func(fields []Field) {
			for _, f := range fields {
				for _, c := range f.Constraints {
					if !strings.HasPrefix(c, "{") {
						continue
					}
					jsonDumps++
					var compacted bytes.Buffer
					if err := json.Compact(&compacted, []byte(c)); err != nil {
						t.Errorf("%s %s: constraint is not valid JSON: %q", kind, f.Name, c)
					} else if compacted.String() != c {
						t.Errorf("%s %s: constraint dump not compact: %q", kind, f.Name, c)
					}
				}
				walk(f.Fields)
			}
		}
		walk(report.Spec)
		walk(report.Outputs)
	}
	if jsonDumps == 0 {
		t.Fatal("no structural JSON constraint dumps found across the catalog -- the guard is vacuous")
	}
}

// TestAllKindsExplain is the walker's survival gate: every compiled-in kind
// must produce a root report without panicking or erroring.
func TestAllKindsExplain(t *testing.T) {
	engine := DefaultEngine()
	for _, kind := range catalogkindreflect.KindsList() {
		res, err := KindResource(kind)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if _, err := engine.Explain(res, nil); err != nil {
			t.Errorf("%s: %v", kind, err)
		}
	}
}

func TestRenderRootView(t *testing.T) {
	engine := DefaultEngine()
	report, err := engine.Explain(mustResource(t, "AwsVpc"), nil)
	if err != nil {
		t.Fatal(err)
	}
	text := Render(report)
	for _, want := range []string{
		"KIND:", "AwsVpc", "VERSION:", "aws.planton.dev/v1alpha1",
		"MANIFEST:", "apiVersion: aws.planton.dev/v1alpha1", // the envelope skeleton
		"SPEC:", "region <string> -required-", "OUTPUTS",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("rendered text missing %q", want)
		}
	}
}

func TestRenderEnumHandling(t *testing.T) {
	big := Field{Name: "kind", Type: "enum"}
	for i := 0; i < 20; i++ {
		big.Enum = append(big.Enum, EnumValue{Name: fmt.Sprintf("Kind%d", i), Doc: "prose"})
	}
	parent := Field{Name: "relationships", Type: "Rel", Fields: []Field{big}}

	t.Run("catalog-scale enums are elided in trees", func(t *testing.T) {
		text := Render(&Report{Kind: "X", Field: &parent})
		if !strings.Contains(text, "<20 values -- drill into this field to list them>") {
			t.Errorf("big enum not elided:\n%s", text)
		}
		if strings.Contains(text, "Kind0") {
			t.Errorf("big enum values leaked into tree view")
		}
	})

	t.Run("the asked-about field expands its values with docs", func(t *testing.T) {
		text := Render(&Report{Kind: "X", Field: &big})
		if !strings.Contains(text, "values (use exactly as shown):") || !strings.Contains(text, "- Kind0") {
			t.Errorf("directly asked enum should expand:\n%s", text)
		}
	})

	t.Run("nested enums under the asked-about field stay names-only", func(t *testing.T) {
		text := Render(&Report{Kind: "X", Field: &parent})
		if strings.Contains(text, "use exactly as shown") {
			t.Errorf("nested enum docs must not expand in a parent's view:\n%s", text)
		}
	})
}

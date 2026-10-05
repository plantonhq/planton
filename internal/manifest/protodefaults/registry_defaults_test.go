// The registry-wide defaults gate: every default option authored on any
// registered kind's schema must CONVERT to its field's type, or the default
// is a lie -- an enum default written as a number (the applier resolves enum
// defaults by NAME), a non-numeric int default, a non-boolean bool default
// all make every manifest that omits the field fail to load, while nothing
// else in the pipeline notices (the reference generator deliberately absorbs
// a broken example into a report, and buf lint guards presence tracking, not
// values). This gate makes the class unshippable.
//
// It lives here, beside the applier whose ConvertStringToFieldValue it
// exercises, rather than in the optional-linter buf plugin: the plugin is a
// separate Go module and cannot import this internal package, so a lint-side
// check would have to DUPLICATE the conversion semantics -- and a duplicated
// contract drifts. One converter, one gate proving every authored default
// against it.
//
// External test package: it imports the kind registry (catalogkindreflect) and the
// manifest loader, both of which sit above this package in the import graph.
package protodefaults_test

import (
	"testing"

	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/internal/manifest/protodefaults"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
	options_pb "github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TestRegistryDefaultsConvert walks every registered kind's full resource
// descriptor (spec, status, metadata -- defaults may be authored anywhere)
// and asserts every field carrying the default option converts to the
// field's type. The walk mirrors the applier's applicability exactly:
// message-kind fields recurse, lists and maps are skipped (the applier
// never applies defaults there).
func TestRegistryDefaultsConvert(t *testing.T) {
	visited := map[protoreflect.FullName]bool{}
	for kind, msg := range catalogkindreflect.ToMessageMap {
		if kind == catalogkind.CatalogKind_unspecified || msg == nil {
			continue
		}
		walkDefaults(t, kind.String(), msg.ProtoReflect().Descriptor(), visited)
	}
}

// walkDefaults descends one message descriptor, cycle-safe: a message type's
// fields do not vary by which kind reaches it, so a type already checked is
// never re-walked (this also terminates self-referential schemas).
func walkDefaults(t *testing.T, kind string, md protoreflect.MessageDescriptor, visited map[protoreflect.FullName]bool) {
	if visited[md.FullName()] {
		return
	}
	visited[md.FullName()] = true

	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.IsList() || field.IsMap() {
			continue
		}
		if field.Kind() == protoreflect.MessageKind {
			walkDefaults(t, kind, field.Message(), visited)
			continue
		}
		opts := field.Options()
		if opts == nil || !proto.HasExtension(opts, options_pb.E_Default) {
			continue
		}
		defaultValue, ok := proto.GetExtension(opts, options_pb.E_Default).(string)
		if !ok || defaultValue == "" {
			continue
		}
		if _, err := protodefaults.ConvertStringToFieldValue(defaultValue, field); err != nil {
			t.Errorf("kind %s: field %s declares default %q that cannot convert to its type -- every manifest omitting the field fails to load: %v",
				kind, field.FullName(), defaultValue, err)
		}
	}
}

// TestEnumDefaultRejectsNumbers is the gate's deliberate red: the converter
// must refuse an enum default authored as the value's NUMBER, because the
// applier resolves enum defaults by name only. Proven against a real enum
// field descriptor so the red can never drift from production behavior.
func TestEnumDefaultRejectsNumbers(t *testing.T) {
	rulesetMsg := catalogkindreflect.ToMessageMap[catalogkind.CatalogKind_CloudflareRuleset]
	if rulesetMsg == nil {
		t.Fatal("CloudflareRuleset not in the registry")
	}
	specField := rulesetMsg.ProtoReflect().Descriptor().Fields().ByName("spec")
	enumField := specField.Message().Fields().ByName("ruleset_kind")
	if _, err := protodefaults.ConvertStringToFieldValue("1", enumField); err == nil {
		t.Error("a numeric enum default must fail conversion (enum defaults resolve by name); it converted")
	}
	if _, err := protodefaults.ConvertStringToFieldValue("zone", enumField); err != nil {
		t.Errorf("the enum value name must convert, got: %v", err)
	}
}

// TestEnumDefaultAppliesOnOmission is the gate's poster child, proven
// end to end through the real loader: a CloudflareRuleset manifest that
// omits rulesetKind must load with the authored default (zone) applied --
// exactly the round trip that failed while the default was authored as the
// enum number.
func TestEnumDefaultAppliesOnOmission(t *testing.T) {
	manifestYaml := []byte(`apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareRuleset
metadata:
  name: defaults-gate-fixture
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  phase: http_request_firewall_custom
  name: defaults gate fixture
  rules:
    - expression: 'http.user_agent eq ""'
      action: managed_challenge
`)
	loaded, err := manifest.LoadManifestBytes(manifestYaml, "defaults-gate-fixture.yaml")
	if err != nil {
		t.Fatalf("a manifest omitting rulesetKind must load with the default applied, got: %v", err)
	}
	specField := loaded.ProtoReflect().Descriptor().Fields().ByName("spec")
	spec := loaded.ProtoReflect().Get(specField).Message()
	kindField := spec.Descriptor().Fields().ByName("ruleset_kind")
	if got := spec.Get(kindField).Enum(); got != 1 {
		t.Errorf("expected the omitted rulesetKind to default to zone (1), got enum number %d", got)
	}
}

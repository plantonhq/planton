package manifestprojection_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/specprojection"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sigs.k8s.io/yaml"
)

// A projection kind promises that its spec mirrors one custom resource field
// for field. TestProjection_MirrorsItsPinnedCustomResource holds that promise
// to the custom resource's own schema: every CustomResourceDefinition pinned
// under testdata/schemas/ is compared, in both directions and type for type,
// with the spec of the projection kind that applies it.
//
//   - A property the served schema has and the spec lacks is a setting no
//     manifest can express.
//   - A spec field the served schema lacks is a key the API server prunes or
//     rejects, so the setting silently never lands.
//   - A property the schema requires must be present in every manifest that
//     validates (a required rule, a minimum count or length, or a platform
//     default), or the object is refused at apply time.
//
// A kind enrolls by committing its pinned schema (the served definition read
// from a real cluster or the upstream release, with its version and how it
// was captured in the file's header). A pinned schema no projection kind
// applies fails too: nothing is pinned ahead of the kind that uses it.

const pinnedSchemasDir = "testdata/schemas"

// crdDocument is the part of a CustomResourceDefinition the mirror check reads.
type crdDocument struct {
	Spec struct {
		Group string `json:"group"`
		Names struct {
			Kind string `json:"kind"`
		} `json:"names"`
		Versions []struct {
			Name   string `json:"name"`
			Served bool   `json:"served"`
			Schema struct {
				OpenAPIV3Schema *schemaNode `json:"openAPIV3Schema"`
			} `json:"schema"`
		} `json:"versions"`
	} `json:"spec"`
}

// schemaNode is one node of a structural OpenAPI v3 schema.
type schemaNode struct {
	Type                  string                 `json:"type"`
	Properties            map[string]*schemaNode `json:"properties"`
	AdditionalProperties  *schemaNode            `json:"additionalProperties"`
	Items                 *schemaNode            `json:"items"`
	Required              []string               `json:"required"`
	Default               json.RawMessage        `json:"default"`
	IntOrString           bool                   `json:"x-kubernetes-int-or-string"`
	PreserveUnknownFields bool                   `json:"x-kubernetes-preserve-unknown-fields"`
	Pattern               string                 `json:"pattern"`
}

// quantityPattern is the pattern controller-gen writes for a
// resource.Quantity ("500m", "2Gi"): an int-or-string the API server accepts
// as a string in every case, so the spec models it as a plain string held to
// this grammar, never with the int-or-string marker (which would turn "2"
// into a number, equally valid, but two spellings of one setting).
const quantityPattern = `^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$`

func TestProjection_MirrorsItsPinnedCustomResource(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(pinnedSchemasDir, "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var crd crdDocument
			if err := yaml.Unmarshal(raw, &crd); err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			kind, version, ok := projectionKindFor(crd.Spec.Group, crd.Spec.Names.Kind)
			if !ok {
				t.Fatalf("%s pins %s/%s, which no Kubernetes manifest projection kind applies; delete it or forge the kind that uses it", path, crd.Spec.Group, crd.Spec.Names.Kind)
			}
			var root *schemaNode
			for _, v := range crd.Spec.Versions {
				if v.Name == version && v.Served {
					root = v.Schema.OpenAPIV3Schema
				}
			}
			if root == nil || root.Properties["spec"] == nil {
				t.Fatalf("%s serves no %s schema with a spec for %s", path, version, kind)
			}
			msg, err := catalogkindreflect.NewInstance(kind)
			if err != nil {
				t.Fatal(err)
			}
			spec := msg.ProtoReflect().Descriptor().Fields().ByName("spec").Message()
			for _, finding := range mirrorFindings(root.Properties["spec"], spec, manifestprojection.EnvelopeOf(spec)) {
				t.Errorf("%s: %s", kind, finding)
			}
		})
	}
}

// projectionKindFor finds the projection kind whose api_version and kind match
// the pinned definition, and the version it applies.
func projectionKindFor(group, crdKind string) (catalogkind.CatalogKind, string, bool) {
	for _, kind := range catalogkindreflect.KindsList() {
		proj := manifestprojection.ProjectionOf(kind)
		if proj == nil || proj.GetKind() != crdKind {
			continue
		}
		g, v, found := strings.Cut(proj.GetApiVersion(), "/")
		if found && g == group {
			return kind, v, true
		}
	}
	return 0, "", false
}

// mirrorFindings compares a custom resource's spec schema with a projection
// kind's spec message and names every divergence by its upstream path.
func mirrorFindings(spec *schemaNode, md protoreflect.MessageDescriptor, envelope manifestprojection.Envelope) []string {
	var findings []string
	compareObject(spec, md, "spec", envelope, &findings)
	sort.Strings(findings)
	return findings
}

func compareObject(node *schemaNode, md protoreflect.MessageDescriptor, path string, envelope manifestprojection.Envelope, findings *[]string) {
	byKey := map[string]protoreflect.FieldDescriptor{}
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if path == "spec" && envelope.Holds(fd.JSONName()) {
			continue
		}
		byKey[fd.JSONName()] = fd
	}
	for key, child := range node.Properties {
		fd, ok := byKey[key]
		if !ok {
			*findings = append(*findings, path+"."+key+" is served by the custom resource and not modeled in the spec")
			continue
		}
		compareField(child, fd, path+"."+key, findings)
	}
	for key := range byKey {
		if _, ok := node.Properties[key]; !ok {
			*findings = append(*findings, path+"."+key+" is in the spec and not served by the custom resource")
		}
	}
	for _, key := range node.Required {
		if child := node.Properties[key]; child != nil && len(child.Default) > 0 {
			// The API server writes a schema default before it validates, so
			// a manifest that omits the property still satisfies the rule.
			continue
		}
		if fd, ok := byKey[key]; ok && !presenceGuaranteed(fd) {
			*findings = append(*findings, path+"."+key+" is required by the custom resource and may be absent from a valid manifest (add a required rule, a minimum, or a platform default)")
		}
	}
}

func compareField(node *schemaNode, fd protoreflect.FieldDescriptor, path string, findings *[]string) {
	switch {
	case fd.IsMap():
		if node.Type != "object" || node.AdditionalProperties == nil || len(node.Properties) > 0 {
			*findings = append(*findings, path+" is a map in the spec and not a free-keyed object in the custom resource")
			return
		}
		if specprojection.IsKubernetesListValuedMapField(fd) {
			wrapper := fd.MapValue().Message().Fields().Get(0)
			if node.AdditionalProperties.Type != "array" || node.AdditionalProperties.Items == nil {
				*findings = append(*findings, path+" is a list-valued map in the spec and its custom resource values are not lists")
				return
			}
			compareValue(node.AdditionalProperties.Items, wrapper, path+".*[]", findings)
			return
		}
		compareValue(node.AdditionalProperties, fd.MapValue(), path+".*", findings)
	case fd.IsList():
		if node.Type != "array" || node.Items == nil {
			*findings = append(*findings, path+" is a list in the spec and not an array in the custom resource")
			return
		}
		compareValue(node.Items, fd, path+"[]", findings)
	default:
		compareValue(node, fd, path, findings)
	}
}

// compareValue checks one value's type: fd's element type for a list, its
// value type for a map.
func compareValue(node *schemaNode, fd protoreflect.FieldDescriptor, path string, findings *[]string) {
	mismatch := func(want string) {
		*findings = append(*findings, path+" is "+want+" in the custom resource and "+describeKind(fd)+" in the spec")
	}
	if node.IntOrString && node.Pattern == quantityPattern {
		if fd.Kind() != protoreflect.StringKind || specprojection.IsKubernetesIntOrStringField(fd) {
			mismatch("a quantity")
		}
		return
	}
	if node.IntOrString {
		if fd.Kind() != protoreflect.StringKind || !specprojection.IsKubernetesIntOrStringField(fd) {
			mismatch("an int-or-string")
		}
		return
	}
	if fd.Kind() == protoreflect.MessageKind {
		switch fd.Message().FullName() {
		case "dev.planton.shared.foreignkey.v1.StringValueOrRef":
			if node.Type != "string" {
				mismatch(describeSchema(node))
			}
			return
		case "google.protobuf.Struct":
			if node.Type != "object" || !(node.PreserveUnknownFields || node.AdditionalProperties != nil) {
				mismatch(describeSchema(node))
			}
			return
		case "google.protobuf.Value", "google.protobuf.ListValue":
			if !node.PreserveUnknownFields {
				mismatch(describeSchema(node))
			}
			return
		}
		if node.Type == "object" && node.Properties == nil && node.AdditionalProperties == nil && fd.Message().Fields().Len() == 0 {
			// An object served with no properties keeps nothing set inside
			// it (the API server prunes it all); an empty message is its
			// faithful model.
			return
		}
		if node.Type != "object" || node.Properties == nil {
			mismatch(describeSchema(node))
			return
		}
		compareObject(node, fd.Message(), path, manifestprojection.Envelope{}, findings)
		return
	}
	switch node.Type {
	case "string":
		if fd.Kind() != protoreflect.StringKind {
			mismatch("a string")
		}
	case "integer":
		switch fd.Kind() {
		case protoreflect.Int32Kind, protoreflect.Int64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind,
			protoreflect.Sint32Kind, protoreflect.Sint64Kind:
		default:
			mismatch("an integer")
		}
	case "number":
		if fd.Kind() != protoreflect.DoubleKind && fd.Kind() != protoreflect.FloatKind {
			mismatch("a number")
		}
	case "boolean":
		if fd.Kind() != protoreflect.BoolKind {
			mismatch("a boolean")
		}
	default:
		mismatch(describeSchema(node))
	}
}

// presenceGuaranteed reports whether every manifest that passes validation
// carries the field, so the projection always renders it.
func presenceGuaranteed(fd protoreflect.FieldDescriptor) bool {
	opts := fd.Options()
	if opts == nil {
		return false
	}
	if proto.HasExtension(opts, options.E_Default) {
		return true
	}
	rules, _ := proto.GetExtension(opts, validate.E_Field).(*validate.FieldRules)
	if rules == nil {
		return false
	}
	if rules.GetRequired() {
		return true
	}
	if ignore := rules.GetIgnore(); ignore == validate.Ignore_IGNORE_IF_ZERO_VALUE || ignore == validate.Ignore_IGNORE_ALWAYS {
		return false
	}
	switch {
	case fd.IsList():
		return rules.GetRepeated().GetMinItems() > 0
	case fd.IsMap():
		return rules.GetMap().GetMinPairs() > 0
	case fd.HasPresence():
		// An unset optional field skips its rules, so a minimum length proves
		// nothing about presence.
		return false
	}
	str := rules.GetString()
	return str.GetMinLen() > 0 || (len(str.GetIn()) > 0 && !slices.Contains(str.GetIn(), ""))
}

func describeKind(fd protoreflect.FieldDescriptor) string {
	if fd.Kind() == protoreflect.MessageKind {
		return "message " + string(fd.Message().Name())
	}
	return fd.Kind().String()
}

func describeSchema(node *schemaNode) string {
	switch {
	case node.Type == "object" && node.Properties != nil:
		return "an object with properties"
	case node.Type == "object" && node.AdditionalProperties != nil:
		return "a free-keyed object"
	case node.PreserveUnknownFields:
		return "free-form"
	case node.Type == "":
		return "untyped"
	}
	return "a " + node.Type
}

// TestMirrorFindings_NamesEveryDivergence proves the comparison itself against
// a projection kind it is not enrolled for, with a schema built to diverge in
// each way the check exists to catch.
func TestMirrorFindings_NamesEveryDivergence(t *testing.T) {
	msg, err := catalogkindreflect.NewInstance(catalogkind.CatalogKind_KubernetesPrometheusRule)
	if err != nil {
		t.Fatal(err)
	}
	spec := msg.ProtoReflect().Descriptor().Fields().ByName("spec").Message()
	str := &schemaNode{Type: "string"}
	rule := &schemaNode{Type: "object", Required: []string{"expr", "alert"}, Properties: map[string]*schemaNode{
		"record":          str,
		"alert":           str,
		"expr":            &schemaNode{IntOrString: true},
		"for":             str,
		"keep_firing_for": str,
		"labels":          &schemaNode{Type: "object", AdditionalProperties: str},
		"annotations":     &schemaNode{Type: "object", AdditionalProperties: str},
	}}
	group := &schemaNode{Type: "object", Properties: map[string]*schemaNode{
		"name":                      str,
		"interval":                  str,
		"query_offset":              str,
		"partial_response_strategy": str,
		"limit":                     &schemaNode{Type: "boolean"},
		"labels":                    &schemaNode{Type: "object", AdditionalProperties: str},
		"rules":                     &schemaNode{Type: "array", Items: rule},
		"served_only":               str,
	}}
	schema := &schemaNode{Type: "object", Properties: map[string]*schemaNode{
		"groups": &schemaNode{Type: "array", Items: group},
	}}

	// A quantity is a plain string in the spec: interval carries no marker,
	// so a quantity-patterned int-or-string there is no finding.
	group.Properties["interval"] = &schemaNode{IntOrString: true, Pattern: quantityPattern}

	got := mirrorFindings(schema, spec, manifestprojection.EnvelopeOf(spec))
	want := []string{
		"spec.groups[].limit is a boolean in the custom resource and int32 in the spec",
		"spec.groups[].rules[].alert is required by the custom resource and may be absent from a valid manifest (add a required rule, a minimum, or a platform default)",
		"spec.groups[].rules[].expr is an int-or-string in the custom resource and string in the spec",
		"spec.groups[].served_only is served by the custom resource and not modeled in the spec",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("findings =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// TestMirrorFindings_AQuantityIsNeverMarked proves the quantity rule's other
// half: a field carrying the int-or-string marker where the served schema has
// a quantity is a finding, because the spec would then spell one setting two
// ways ("2" as a number, "2Gi" as a string).
func TestMirrorFindings_AQuantityIsNeverMarked(t *testing.T) {
	msg, err := catalogkindreflect.NewInstance(catalogkind.CatalogKind_KubernetesServiceMonitor)
	if err != nil {
		t.Fatal(err)
	}
	spec := msg.ProtoReflect().Descriptor().Fields().ByName("spec").Message()
	endpoint := spec.Fields().ByName("endpoints").Message()
	node := &schemaNode{IntOrString: true, Pattern: quantityPattern}

	var findings []string
	compareValue(node, endpoint.Fields().ByName("target_port"), "spec.endpoints[].targetPort", &findings)
	want := "spec.endpoints[].targetPort is a quantity in the custom resource and string in the spec"
	if len(findings) != 1 || findings[0] != want {
		t.Errorf("findings = %q, want [%q]", findings, want)
	}
}

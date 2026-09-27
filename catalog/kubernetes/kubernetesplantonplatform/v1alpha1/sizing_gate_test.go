package kubernetesplantonplatformv1alpha1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"

	"github.com/plantonhq/planton/catalog/kubernetes"
)

// This kind is a hand-kept mirror of the operator's PlantonPlatform schema,
// and a mirror lags silently: the day the operator sizes a component the
// catalog does not, an adopter installing through this kind simply cannot
// size it, and the day a default is re-measured, the console keeps showing
// the old number. This gate reads the operator's own generated facts -- its
// sizing registry published as api/v1/component_sizing.json and its CRD --
// and fails when a sized component is missing from either definition or when
// this kind's stated default is not the operator's.
//
// Not covered: whether the IaC modules forward the field (their render tests
// hold that) or whether the defaults are right (the operator's own tests).

const (
	operatorSizingFile = "../../../../operator/api/v1/component_sizing.json"
	operatorCRDFile    = "../../../../operator/config/crd/bases/planton.ai_plantonplatforms.yaml"
)

type sizingEntry struct {
	Path     string            `json:"path"`
	Requests map[string]string `json:"requests"`
	Limits   map[string]string `json:"limits"`
}

func readRegistry(t *testing.T) []sizingEntry {
	t.Helper()
	data, err := os.ReadFile(filepath.FromSlash(operatorSizingFile))
	if err != nil {
		t.Fatalf("the operator's sizing registry is its published fact (make -C operator manifests writes it): %v", err)
	}
	var entries []sizingEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("the operator's registry names no sized workload; the gate would prove nothing")
	}
	return entries
}

// segments turns "spec.database.postgresql.resources" into its field names
// after "spec", in this kind's snake_case.
func segments(path string) []string {
	parts := strings.Split(strings.TrimPrefix(path, "spec."), ".")
	for i, p := range parts {
		var b strings.Builder
		for j, r := range p {
			if unicode.IsUpper(r) {
				if j > 0 {
					b.WriteByte('_')
				}
				r = unicode.ToLower(r)
			}
			b.WriteRune(r)
		}
		parts[i] = b.String()
	}
	return parts
}

// catalogField walks this kind's spec to the field a registry path names,
// returning it and its parent field (nil at the top level).
func catalogField(path string) (field, parent protoreflect.FieldDescriptor) {
	msg := (&KubernetesPlantonPlatformSpec{}).ProtoReflect().Descriptor()
	for i, name := range segments(path) {
		f := msg.Fields().ByName(protoreflect.Name(name))
		if f == nil {
			return nil, nil
		}
		if i == len(segments(path))-1 {
			return f, parent
		}
		parent = f
		if f.Message() == nil {
			return nil, nil
		}
		msg = f.Message()
	}
	return nil, nil
}

// statedDefault is the default this kind states for a path: the option on the
// resources field, or -- where one message serves several services with
// different defaults (Temporal's) -- on the service's own field.
func statedDefault(field, parent protoreflect.FieldDescriptor) *kubernetes.ContainerResources {
	for _, f := range []protoreflect.FieldDescriptor{field, parent} {
		if f == nil {
			continue
		}
		if proto.HasExtension(f.Options(), kubernetes.E_DefaultContainerResources) {
			return proto.GetExtension(f.Options(), kubernetes.E_DefaultContainerResources).(*kubernetes.ContainerResources)
		}
	}
	return nil
}

func sameQuantity(t *testing.T, where, want, got string) {
	t.Helper()
	if want == "" && got == "" {
		return
	}
	if want == "" || got == "" {
		t.Errorf("%s: the operator's default is %q and this kind states %q", where, want, got)
		return
	}
	w, g := resource.MustParse(want), resource.MustParse(got)
	if w.Cmp(g) != 0 {
		t.Errorf("%s: the operator's default is %s and this kind states %s -- set the field's default_container_resources to the operator's (api/v1/component_sizing.json)", where, want, got)
	}
}

func TestEveryOperatorSizedComponentIsSizableHereAtTheOperatorsDefault(t *testing.T) {
	for _, e := range readRegistry(t) {
		field, parent := catalogField(e.Path)
		if field == nil {
			t.Errorf("%s: the operator sizes this workload and this kind has no field for it; add it to spec.proto, forward it in both IaC modules, and state the operator's default", e.Path)
			continue
		}
		if field.Message() == nil || field.Message().FullName() != (&kubernetes.ContainerResources{}).ProtoReflect().Descriptor().FullName() {
			t.Errorf("%s: this kind's field is %s, not dev.planton.kubernetes.ContainerResources", e.Path, field.Kind())
			continue
		}
		stated := statedDefault(field, parent)
		if stated == nil {
			t.Errorf("%s: this kind states no default (default_container_resources) for the field", e.Path)
			continue
		}
		sameQuantity(t, e.Path+".requests.cpu", e.Requests["cpu"], stated.GetRequests().GetCpu())
		sameQuantity(t, e.Path+".requests.memory", e.Requests["memory"], stated.GetRequests().GetMemory())
		sameQuantity(t, e.Path+".limits.cpu", e.Limits["cpu"], stated.GetLimits().GetCpu())
		sameQuantity(t, e.Path+".limits.memory", e.Limits["memory"], stated.GetLimits().GetMemory())
	}
}

func TestEveryOperatorSizedComponentIsInTheOperatorsDefinition(t *testing.T) {
	data, err := os.ReadFile(filepath.FromSlash(operatorCRDFile))
	if err != nil {
		t.Fatalf("the operator's generated CRD: %v", err)
	}
	var crd struct {
		Spec struct {
			Versions []struct {
				Schema struct {
					OpenAPIV3Schema map[string]any `json:"openAPIV3Schema"`
				} `json:"schema"`
			} `json:"versions"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(data, &crd); err != nil || len(crd.Spec.Versions) == 0 {
		t.Fatalf("reading the CRD: %v", err)
	}
	root := crd.Spec.Versions[0].Schema.OpenAPIV3Schema
	for _, e := range readRegistry(t) {
		node := root
		for _, key := range strings.Split(e.Path, ".") {
			props, _ := node["properties"].(map[string]any)
			next, ok := props[key].(map[string]any)
			if !ok {
				t.Errorf("%s: the operator's registry sizes it and its CRD has no such field -- regenerate with make -C operator manifests", e.Path)
				node = nil
				break
			}
			node = next
		}
	}
}

package generators

import (
	"encoding/json"
	"testing"

	testkubernetesv1 "github.com/plantonhq/planton/catalog/_test/testcloudresourcekubernetes/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// buildTestK8sJSON creates a TestCloudResourceKubernetes proto with
// StringValueOrRef namespace and map<string, StringValueOrRef> ref_map,
// then returns its JSON-unmarshaled map. This exercises the shapes the
// flatten logic must handle.
func buildTestK8sJSON(t *testing.T) map[string]interface{} {
	t.Helper()

	msg := &testkubernetesv1.TestCloudResourceKubernetes{
		ApiVersion: "_test.planton.dev/v1alpha1",
		Kind:       "TestCloudResourceKubernetes",
		Spec: &testkubernetesv1.TestCloudResourceKubernetesSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "e2e-test-ns",
				},
			},
			CreateNamespace: true,
			Schedule:        stringPtr("*/5 * * * *"),
			RefMap: map[string]*foreignkeyv1.StringValueOrRef{
				"DB_HOST": {
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "localhost",
					},
				},
				"DB_PORT": {
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "5432",
					},
				},
			},
		},
	}

	jsonBytes, err := protojson.MarshalOptions{EmitUnpopulated: false}.Marshal(msg)
	if err != nil {
		t.Fatalf("protojson.Marshal: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	return data
}

func stringPtr(s string) *string { return &s }

func TestFlatten_StringValueOrRef_Singular(t *testing.T) {
	data := buildTestK8sJSON(t)
	md := (&testkubernetesv1.TestCloudResourceKubernetes{}).ProtoReflect().Descriptor()

	Flatten(data, md, DefaultRules())

	spec, ok := data["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("spec should be a map after flatten")
	}

	ns, ok := spec["namespace"]
	if !ok {
		t.Fatal("spec.namespace should exist after flatten")
	}
	nsStr, ok := ns.(string)
	if !ok {
		t.Fatalf("spec.namespace should be a string, got %T", ns)
	}
	if nsStr != "e2e-test-ns" {
		t.Errorf("spec.namespace = %q, want %q", nsStr, "e2e-test-ns")
	}
}

func TestFlatten_SkipRule_RemovesField(t *testing.T) {
	data := buildTestK8sJSON(t)
	md := (&testkubernetesv1.TestCloudResourceKubernetes{}).ProtoReflect().Descriptor()

	// Populate the env field (a ContainerEnv message), then flatten with a
	// rule set that marks ContainerEnv as Skip. The field must be removed
	// from the output regardless of key casing.
	if spec, ok := data["spec"].(map[string]interface{}); ok {
		spec["env"] = map[string]interface{}{
			"variables": map[string]interface{}{"LOG_LEVEL": "debug"},
		}
	}

	rules := DefaultRules()
	rules["dev.planton.kubernetes.ContainerEnv"] = TypeRule{Skip: true}

	Flatten(data, md, rules)

	spec := data["spec"].(map[string]interface{})
	if _, exists := spec["env"]; exists {
		t.Error("spec.env should have been removed by the Skip rule")
	}
}

func TestFlatten_MapWithStringValueOrRef(t *testing.T) {
	data := buildTestK8sJSON(t)
	md := (&testkubernetesv1.TestCloudResourceKubernetes{}).ProtoReflect().Descriptor()

	Flatten(data, md, DefaultRules())

	spec := data["spec"].(map[string]interface{})
	refMap, ok := spec["ref_map"].(map[string]interface{})
	if !ok {
		t.Fatal("spec.ref_map should be a map after flatten")
	}

	dbHost, ok := refMap["DB_HOST"]
	if !ok {
		t.Fatal("DB_HOST should exist in ref_map")
	}
	if dbHost != "localhost" {
		t.Errorf("DB_HOST = %v, want %q", dbHost, "localhost")
	}

	dbPort, ok := refMap["DB_PORT"]
	if !ok {
		t.Fatal("DB_PORT should exist in ref_map")
	}
	if dbPort != "5432" {
		t.Errorf("DB_PORT = %v, want %q", dbPort, "5432")
	}
}

func TestFlatten_PreservesNonRuleFields(t *testing.T) {
	data := buildTestK8sJSON(t)
	md := (&testkubernetesv1.TestCloudResourceKubernetes{}).ProtoReflect().Descriptor()

	Flatten(data, md, DefaultRules())

	spec := data["spec"].(map[string]interface{})

	if _, ok := spec["schedule"]; !ok {
		t.Error("spec.schedule should be preserved (plain string, no rule)")
	}
	if _, ok := spec["create_namespace"]; !ok {
		t.Error("spec.create_namespace should be preserved (bool, no rule)")
	}
}

// TestFlatten_ManifestOnlyField_DroppedOnBothPaths asserts a field marked
// (dev.planton.shared.options.manifest_only) never reaches tfvars: not on the
// snake_case path a hand-written module reads, and not on the manifest
// projection a CRD-faithful module forwards verbatim (where an unknown key
// would be refused by the apiserver at apply). A sibling scalar without the
// marker rides through untouched on both.
func TestFlatten_ManifestOnlyField_DroppedOnBothPaths(t *testing.T) {
	md := (&testkubernetesv1.TestCloudResourceKubernetes{}).ProtoReflect().Descriptor()

	for _, tc := range []struct {
		name string
		opts flattenOpts
		key  string // how the marked field's sibling is spelled after the walk
	}{
		{name: "snake_case path", opts: flattenOpts{}, key: "create_namespace"},
		{name: "manifest projection", opts: flattenOpts{preserveJSONNames: true}, key: "createNamespace"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := buildTestK8sJSON(t)
			spec := data["spec"].(map[string]interface{})
			spec["selectAll"] = true // the marked field, as protojson writes it

			flattenWithOpts(data, md, DefaultRules(), tc.opts)

			spec = data["spec"].(map[string]interface{})
			if _, exists := spec["selectAll"]; exists {
				t.Error("spec.selectAll carries manifest_only and must be dropped (camelCase key)")
			}
			if _, exists := spec["select_all"]; exists {
				t.Error("spec.select_all carries manifest_only and must be dropped (snake_case key)")
			}
			if _, exists := spec[tc.key]; !exists {
				t.Errorf("spec.%s has no marker and must be preserved", tc.key)
			}
		})
	}
}

func TestFlatten_EmptyRules_NoChanges(t *testing.T) {
	data := buildTestK8sJSON(t)
	md := (&testkubernetesv1.TestCloudResourceKubernetes{}).ProtoReflect().Descriptor()

	Flatten(data, md, map[string]TypeRule{})

	spec := data["spec"].(map[string]interface{})
	ns := spec["namespace"]

	if _, ok := ns.(map[string]interface{}); !ok {
		t.Errorf("with empty rules, namespace should remain a map, got %T", ns)
	}
}

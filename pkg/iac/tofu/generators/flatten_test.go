package generators

import (
	"encoding/json"
	"testing"

	testkubernetesv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/_test/testcloudresourcekubernetes/v1"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// buildTestK8sJSON creates a TestCloudResourceKubernetes proto with
// StringValueOrRef namespace, KubernetesClusterSelector target_cluster,
// and map<string, StringValueOrRef> ref_map, then returns its
// JSON-unmarshaled map. This exercises all three shapes the flatten
// logic must handle.
func buildTestK8sJSON(t *testing.T) map[string]interface{} {
	t.Helper()

	msg := &testkubernetesv1.TestCloudResourceKubernetes{
		ApiVersion: "_test.openmcf.org/v1",
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

func TestFlatten_KubernetesClusterSelector_Skipped(t *testing.T) {
	data := buildTestK8sJSON(t)
	md := (&testkubernetesv1.TestCloudResourceKubernetes{}).ProtoReflect().Descriptor()

	// TestCloudResourceKubernetes has target_cluster but our test message
	// doesn't set it. Inject a fake target_cluster into spec using the
	// JSON key (camelCase) to verify the skip rule removes it.
	if spec, ok := data["spec"].(map[string]interface{}); ok {
		spec["targetCluster"] = map[string]interface{}{
			"clusterKind": "AzureAksCluster",
			"clusterName": "test-cluster",
		}
	}

	Flatten(data, md, DefaultRules())

	spec := data["spec"].(map[string]interface{})
	if _, exists := spec["targetCluster"]; exists {
		t.Error("spec.targetCluster (camelCase) should have been skipped by flatten")
	}
	if _, exists := spec["target_cluster"]; exists {
		t.Error("spec.target_cluster (snake_case) should have been skipped by flatten")
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

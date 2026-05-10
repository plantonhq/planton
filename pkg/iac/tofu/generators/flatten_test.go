package generators

import (
	"encoding/json"
	"testing"

	kubernetescronjobv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetescronjob/v1"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// buildCronJobJSON creates a CronJob proto with StringValueOrRef namespace,
// KubernetesClusterSelector target_cluster, and map<string, StringValueOrRef>
// env variables, then returns its JSON-unmarshaled map alongside the proto
// descriptor. This exercises all three shapes the flatten logic must handle.
func buildCronJobJSON(t *testing.T) map[string]interface{} {
	t.Helper()

	msg := &kubernetescronjobv1.KubernetesCronJob{
		ApiVersion: "kubernetes.openmcf.org/v1",
		Kind:       "KubernetesCronJob",
		Spec: &kubernetescronjobv1.KubernetesCronJobSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "e2e-cronjob-ns",
				},
			},
			CreateNamespace: true,
			Schedule:        "*/5 * * * *",
			Env: &kubernetescronjobv1.KubernetesCronJobContainerAppEnv{
				Variables: map[string]*foreignkeyv1.StringValueOrRef{
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

func TestFlatten_StringValueOrRef_Singular(t *testing.T) {
	data := buildCronJobJSON(t)
	md := (&kubernetescronjobv1.KubernetesCronJob{}).ProtoReflect().Descriptor()

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
	if nsStr != "e2e-cronjob-ns" {
		t.Errorf("spec.namespace = %q, want %q", nsStr, "e2e-cronjob-ns")
	}
}

func TestFlatten_KubernetesClusterSelector_Skipped(t *testing.T) {
	data := buildCronJobJSON(t)
	md := (&kubernetescronjobv1.KubernetesCronJob{}).ProtoReflect().Descriptor()

	// The CronJob proto has target_cluster but our test message doesn't set it.
	// Ensure that even if it were present, the flatten logic would skip it.
	// Inject a fake target_cluster into spec using the JSON key (camelCase).
	if spec, ok := data["spec"].(map[string]interface{}); ok {
		spec["targetCluster"] = map[string]interface{}{
			"clusterKind": "AzureAksCluster",
			"clusterName": "test-cluster",
		}
	}

	Flatten(data, md, DefaultRules())

	spec := data["spec"].(map[string]interface{})
	// After flatten, neither the JSON key nor the snake_case key should exist.
	if _, exists := spec["targetCluster"]; exists {
		t.Error("spec.targetCluster (camelCase) should have been skipped by flatten")
	}
	if _, exists := spec["target_cluster"]; exists {
		t.Error("spec.target_cluster (snake_case) should have been skipped by flatten")
	}
}

func TestFlatten_MapWithStringValueOrRef(t *testing.T) {
	data := buildCronJobJSON(t)
	md := (&kubernetescronjobv1.KubernetesCronJob{}).ProtoReflect().Descriptor()

	Flatten(data, md, DefaultRules())

	spec := data["spec"].(map[string]interface{})
	env, ok := spec["env"].(map[string]interface{})
	if !ok {
		t.Fatal("spec.env should be a map")
	}

	vars, ok := env["variables"].(map[string]interface{})
	if !ok {
		t.Fatal("spec.env.variables should be a map after flatten")
	}

	dbHost, ok := vars["DB_HOST"]
	if !ok {
		t.Fatal("DB_HOST should exist in variables")
	}
	if dbHost != "localhost" {
		t.Errorf("DB_HOST = %v, want %q", dbHost, "localhost")
	}

	dbPort, ok := vars["DB_PORT"]
	if !ok {
		t.Fatal("DB_PORT should exist in variables")
	}
	if dbPort != "5432" {
		t.Errorf("DB_PORT = %v, want %q", dbPort, "5432")
	}
}

func TestFlatten_PreservesNonRuleFields(t *testing.T) {
	data := buildCronJobJSON(t)
	md := (&kubernetescronjobv1.KubernetesCronJob{}).ProtoReflect().Descriptor()

	Flatten(data, md, DefaultRules())

	spec := data["spec"].(map[string]interface{})

	// After flatten, proto field names are snake_case (proto name), not
	// camelCase (JSON name).
	if _, ok := spec["schedule"]; !ok {
		t.Error("spec.schedule should be preserved (plain string, no rule)")
	}
	if _, ok := spec["create_namespace"]; !ok {
		t.Error("spec.create_namespace should be preserved (bool, no rule)")
	}
}

func TestFlatten_EmptyRules_NoChanges(t *testing.T) {
	data := buildCronJobJSON(t)
	md := (&kubernetescronjobv1.KubernetesCronJob{}).ProtoReflect().Descriptor()

	// With empty rules, nothing should be flattened or skipped.
	Flatten(data, md, map[string]TypeRule{})

	spec := data["spec"].(map[string]interface{})
	ns := spec["namespace"]

	// namespace should still be a map (not flattened).
	if _, ok := ns.(map[string]interface{}); !ok {
		t.Errorf("with empty rules, namespace should remain a map, got %T", ns)
	}
}

//go:build !codegen
// +build !codegen

package schema

import (
	"encoding/json"
	"testing"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestGenerate_ProducesValidJSON(t *testing.T) {
	data, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	defs, ok := schema["definitions"].(map[string]any)
	if !ok {
		t.Fatal("missing or invalid 'definitions' key")
	}

	if len(defs) == 0 {
		t.Fatal("expected at least one definition entry")
	}
}

func TestGenerate_KubernetesDeploymentMergeFields(t *testing.T) {
	data, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	def := extractDefinition(t, data, "kubernetes.openmcf.org.v1.KubernetesDeployment")
	assertMergeFieldExists(t, def, "spec", "container", "app", "env", "variables")
	assertMergeFieldExists(t, def, "spec", "container", "app", "env", "secrets")
	assertMergeFieldExists(t, def, "spec", "container", "app", "ports")
}

func TestGenerate_KubernetesCronJobMergeFields(t *testing.T) {
	data, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	def := extractDefinition(t, data, "kubernetes.openmcf.org.v1.KubernetesCronJob")
	assertMergeFieldExists(t, def, "spec", "env", "variables")
	assertMergeFieldExists(t, def, "spec", "env", "secrets")
}

func TestGenerate_KubernetesJobMergeFields(t *testing.T) {
	data, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	def := extractDefinition(t, data, "kubernetes.openmcf.org.v1.KubernetesJob")
	assertMergeFieldExists(t, def, "spec", "env", "variables")
	assertMergeFieldExists(t, def, "spec", "env", "secrets")
}

func TestGenerate_ExcludesKindsWithoutMergeFields(t *testing.T) {
	data, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	defs := schema["definitions"].(map[string]any)

	// AwsVpc should not be in the schema (no repeated message fields with name)
	if _, ok := defs["aws.openmcf.org.v1.AwsVpc"]; ok {
		t.Error("AwsVpc should not be in the schema (no merge fields)")
	}
}

func TestGenerate_GroupVersionKindMetadata(t *testing.T) {
	data, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	def := extractDefinition(t, data, "kubernetes.openmcf.org.v1.KubernetesDeployment")
	gvkList, ok := def["x-kubernetes-group-version-kind"].([]any)
	if !ok || len(gvkList) == 0 {
		t.Fatal("missing x-kubernetes-group-version-kind")
	}

	gvk := gvkList[0].(map[string]any)
	if gvk["group"] != "kubernetes.openmcf.org" {
		t.Errorf("expected group 'kubernetes.openmcf.org', got %v", gvk["group"])
	}
	if gvk["version"] != "v1" {
		t.Errorf("expected version 'v1', got %v", gvk["version"])
	}
	if gvk["kind"] != "KubernetesDeployment" {
		t.Errorf("expected kind 'KubernetesDeployment', got %v", gvk["kind"])
	}
}

func TestFindMergeFields_SkipsMapFields(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_KubernetesDeployment
	msg, err := crkreflect.NewInstance(kind)
	if err != nil {
		t.Fatalf("NewInstance failed: %v", err)
	}

	md := msg.ProtoReflect().Descriptor()
	specField := md.Fields().ByJSONName("spec")
	if specField == nil {
		t.Fatal("spec field not found")
	}

	visited := make(map[protoreflect.FullName]bool)
	fields := findMergeFields(specField.Message(), visited)

	for _, f := range fields {
		if f.jsonPath == "configMaps" {
			t.Error("configMaps (map field) should not appear as a merge field")
		}
	}
}

func TestFindMergeFields_ExcludesEnvFrom(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_KubernetesDeployment
	msg, err := crkreflect.NewInstance(kind)
	if err != nil {
		t.Fatalf("NewInstance failed: %v", err)
	}

	md := msg.ProtoReflect().Descriptor()
	specField := md.Fields().ByJSONName("spec")
	visited := make(map[protoreflect.FullName]bool)
	fields := findMergeFields(specField.Message(), visited)

	for _, f := range fields {
		if f.jsonPath == "container.app.env.envFrom" {
			t.Error("envFrom should be excluded (EnvFromSource has no 'name' field)")
		}
	}
}

// --- helpers ---

func extractDefinition(t *testing.T, schemaJSON []byte, defKey string) map[string]any {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	defs := schema["definitions"].(map[string]any)
	def, ok := defs[defKey].(map[string]any)
	if !ok {
		t.Fatalf("definition %q not found in schema", defKey)
	}
	return def
}

func assertMergeFieldExists(t *testing.T, def map[string]any, path ...string) {
	t.Helper()

	// Start from the definition root -- first step into "properties"
	current, ok := def["properties"].(map[string]any)
	if !ok {
		t.Fatalf("definition has no top-level 'properties' key")
	}

	for i, key := range path {
		val, ok := current[key]
		if !ok {
			t.Fatalf("key %q not found at depth %d (path: %v)", key, i, path)
		}
		node, ok := val.(map[string]any)
		if !ok {
			t.Fatalf("value at %q is not a map at depth %d (path: %v)", key, i, path)
		}

		if i == len(path)-1 {
			if node["x-kubernetes-patch-merge-key"] != "name" {
				t.Errorf("path %v: expected x-kubernetes-patch-merge-key=name, got %v",
					path, node["x-kubernetes-patch-merge-key"])
			}
			if node["x-kubernetes-patch-strategy"] != "merge" {
				t.Errorf("path %v: expected x-kubernetes-patch-strategy=merge, got %v",
					path, node["x-kubernetes-patch-strategy"])
			}
		} else {
			props, ok := node["properties"].(map[string]any)
			if !ok {
				t.Fatalf("no 'properties' at %q depth %d (path: %v)", key, i, path)
			}
			current = props
		}
	}
}

package generators

import (
	"strings"
	"testing"

	kubernetescronjobv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetescronjob/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

func TestProtoToVariablesTF_CronJob_TargetClusterSkipped(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	// target_cluster is inside the spec object. It should NOT appear as a
	// field in the spec's object type.
	if strings.Contains(got, "target_cluster") {
		t.Errorf("target_cluster should be skipped by KubernetesClusterSelector rule, got:\n%s", got)
	}
}

func TestProtoToVariablesTF_CronJob_NamespaceIsString(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	// namespace should be flattened to string, not object({value = string, ...})
	// It appears inside the spec object type.
	if strings.Contains(got, "namespace = object(") {
		t.Errorf("namespace should be 'string' (flattened from StringValueOrRef), not an object:\n%s", got)
	}
	if !strings.Contains(got, "namespace = string") {
		t.Errorf("namespace should appear as 'string' in the spec object, got:\n%s", got)
	}
}

func TestProtoToVariablesTF_CronJob_VariablesIsMapString(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	// map<string, StringValueOrRef> variables should become map(string),
	// not map(object({...})) or an object with synthetic map-entry fields.
	if !strings.Contains(got, "variables = map(string)") {
		t.Errorf("variables should be 'map(string)', got:\n%s", got)
	}
}

func TestProtoToVariablesTF_CronJob_ApiVersionKindStatusSkipped(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	if strings.Contains(got, `"api_version"`) {
		t.Error("api_version should be skipped")
	}
	if strings.Contains(got, `"kind"`) {
		t.Error("kind should be skipped")
	}
	if strings.Contains(got, `"status"`) {
		t.Error("status should be skipped")
	}
}

func TestProtoToVariablesTF_CronJob_HasMetadataAndSpec(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	if !strings.Contains(got, `variable "metadata"`) {
		t.Error("should have metadata variable")
	}
	if !strings.Contains(got, `variable "spec"`) {
		t.Error("should have spec variable")
	}
}

func TestProtoToVariablesTF_CronJob_ValidHCL(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(got), "variables.tf")
	if diags.HasErrors() {
		t.Fatalf("generated variables.tf is not valid HCL: %s\n%s", diags.Error(), got)
	}

	blockSchema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "variable", LabelNames: []string{"name"}},
		},
	}
	content, diags := file.Body.Content(blockSchema)
	if diags.HasErrors() {
		t.Fatalf("failed to get content: %s", diags.Error())
	}

	// Should have exactly 2 variable blocks: metadata and spec
	if len(content.Blocks) != 2 {
		t.Errorf("expected 2 variable blocks, got %d", len(content.Blocks))
	}
}

func TestProtoToVariablesTF_SimpleMessage_BackwardCompatible(t *testing.T) {
	// Test with CloudResourceMetadata directly to verify basic object generation.
	msg := &shared.CloudResourceMetadata{}

	got, err := ProtoToVariablesTF(msg)
	if err != nil {
		t.Fatalf("ProtoToVariablesTF: %v", err)
	}

	// Should have name and labels fields (version skipped in metadata).
	if !strings.Contains(got, `variable "name"`) {
		t.Errorf("should have name variable, got:\n%s", got)
	}
	if !strings.Contains(got, `variable "labels"`) {
		t.Errorf("should have labels variable, got:\n%s", got)
	}

	// version should be skipped (metadata convention).
	if strings.Contains(got, `"version"`) {
		t.Errorf("version should be skipped in metadata messages, got:\n%s", got)
	}
}

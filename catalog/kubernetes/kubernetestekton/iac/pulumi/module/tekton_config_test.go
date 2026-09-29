package module

import (
	"reflect"
	"testing"

	kubernetestektonv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestekton/v1alpha1"
)

// Tekton reads its CloudEvents sink from the config-events ConfigMap, whose
// `sink` supersedes config-defaults' deprecated default-cloud-events-sink; a
// sink written only to the old key stops reaching Tekton once the key is
// removed upstream. The TektonConfig carries the ConfigMap's data through
// spec.pipeline.options.configMaps, which the operator merges into the
// ConfigMap it installs. The Terraform module renders the same block
// (iac/tf/locals.tf pipeline_block_full).

func TestPipelineBody_SinkLandsInConfigEventsNotConfigDefaults(t *testing.T) {
	sink := "http://receiver.ci.svc.cluster.local/events"
	body := pipelineBody(&kubernetestektonv1alpha1.KubernetesTektonPipeline{CloudEventsSinkUrl: sink})

	if got, ok := body["default-cloud-events-sink"]; ok {
		t.Fatalf("pipeline body writes the deprecated config-defaults key default-cloud-events-sink = %v", got)
	}
	want := map[string]interface{}{
		"configMaps": map[string]interface{}{
			"config-events": map[string]interface{}{
				"data": map[string]interface{}{
					"sink":    sink,
					"formats": "tektonv1",
				},
			},
		},
	}
	if got := body["options"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("pipeline options = %#v, want %#v", got, want)
	}
}

func TestPipelineBody_NoSinkRendersNoOptions(t *testing.T) {
	beta := "beta"
	body := pipelineBody(&kubernetestektonv1alpha1.KubernetesTektonPipeline{EnableApiFields: &beta})
	if got, ok := body["options"]; ok {
		t.Fatalf("pipeline body without a sink rendered options = %#v", got)
	}
}

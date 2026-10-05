package module

import (
	"reflect"
	"testing"

	kubernetesotelcollectorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesotelcollector/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// These tests pin the self-metrics switch: on, the custom resource asks the
// operator for its monitor (spec.observability.metrics.enableMetrics), which
// the operator turns into a ServiceMonitor on the collector's monitoring
// Service; off, the body carries no observability block, so the operator
// creates none. The OpenTofu twin renders the same block from the same field.

func observabilityLocals(on bool) *Locals {
	mode := kubernetesotelcollectorv1alpha1.KubernetesOtelCollectorMode_daemonset
	return initializeLocals(nil, &kubernetesotelcollectorv1alpha1.KubernetesOtelCollectorIacInput{
		Target: &kubernetesotelcollectorv1alpha1.KubernetesOtelCollector{
			Metadata: &shared.CatalogObjectMetadata{Name: "cluster-logs"},
			Spec: &kubernetesotelcollectorv1alpha1.KubernetesOtelCollectorSpec{
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "observability"},
				},
				Mode:                  &mode,
				ConfigYaml:            "receivers: {}\n",
				ServiceMonitorEnabled: on,
			},
		},
	})
}

func TestServiceMonitorEnabledAsksTheOperatorForItsMonitor(t *testing.T) {
	body, err := collectorSpecBody(observabilityLocals(true))
	if err != nil {
		t.Fatalf("collectorSpecBody: %v", err)
	}
	want := map[string]interface{}{"metrics": map[string]interface{}{"enableMetrics": true}}
	if !reflect.DeepEqual(body["observability"], want) {
		t.Errorf("observability = %v, want %v", body["observability"], want)
	}
}

func TestServiceMonitorOffRendersNoObservabilityBlock(t *testing.T) {
	body, err := collectorSpecBody(observabilityLocals(false))
	if err != nil {
		t.Fatalf("collectorSpecBody: %v", err)
	}
	if _, present := body["observability"]; present {
		t.Errorf("observability rendered while off: %v", body["observability"])
	}
}

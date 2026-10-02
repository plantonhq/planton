package module

import (
	"testing"

	kuberneteslokiv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesloki/v1alpha1"
	"github.com/plantonhq/planton/shared"
)

// These tests pin that switching the canary off installs: the upstream
// chart refuses to render when its Helm test is on and the canary is off,
// because the test reads the canary's metrics, so the two go off together.
// Left at its default, the chart keeps both and nothing is rendered.
// The OpenTofu twin renders the same pair from the same condition.

func canaryLocals(canary *bool) *Locals {
	return initializeLocals(nil, &kuberneteslokiv1alpha1.KubernetesLokiStackInput{
		Target: &kuberneteslokiv1alpha1.KubernetesLoki{
			Metadata: &shared.CloudResourceMetadata{Name: "logs"},
			Spec: &kuberneteslokiv1alpha1.KubernetesLokiSpec{
				Namespace:     literal("observability"),
				CanaryEnabled: canary,
			},
		},
	})
}

func TestCanaryOffTurnsTheChartsHelmTestOff(t *testing.T) {
	off := false
	values, err := buildHelmValues(canaryLocals(&off))
	if err != nil {
		t.Fatalf("buildHelmValues: %v", err)
	}
	for _, key := range []string{"lokiCanary", "test"} {
		block, ok := values[key].(map[string]interface{})
		if !ok || block["enabled"] != false {
			t.Errorf("values[%q] = %v, want enabled: false", key, values[key])
		}
	}
}

func TestCanaryAtItsDefaultRendersNeither(t *testing.T) {
	on := true
	for name, canary := range map[string]*bool{"unset": nil, "true": &on} {
		values, err := buildHelmValues(canaryLocals(canary))
		if err != nil {
			t.Fatalf("%s: buildHelmValues: %v", name, err)
		}
		for _, key := range []string{"lokiCanary", "test"} {
			if _, present := values[key]; present {
				t.Errorf("%s: values[%q] rendered, want the chart's own default", name, key)
			}
		}
	}
}

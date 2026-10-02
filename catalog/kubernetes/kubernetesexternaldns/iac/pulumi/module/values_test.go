package module

import (
	"testing"

	kubernetesexternaldnsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesexternaldns/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// These tests pin the listener-set switch: on, the chart value that adds
// the controller's --gateway-listener-sets flag and its permission to read
// ListenerSets is rendered; off, nothing is, so the chart keeps its own
// default. The OpenTofu twin renders the same key from the same field.

func listenerSetLocals(on bool) *Locals {
	return initializeLocals(nil, &kubernetesexternaldnsv1alpha1.KubernetesExternalDnsStackInput{
		Target: &kubernetesexternaldnsv1alpha1.KubernetesExternalDns{
			Metadata: &shared.CloudResourceMetadata{Name: "external-dns"},
			Spec: &kubernetesexternaldnsv1alpha1.KubernetesExternalDnsSpec{
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "external-dns"},
				},
				DnsProvider: &kubernetesexternaldnsv1alpha1.KubernetesExternalDnsSpec_InMemory{
					InMemory: &kubernetesexternaldnsv1alpha1.KubernetesExternalDnsInMemory{},
				},
				Sources:             []string{"gateway-httproute"},
				GatewayListenerSets: on,
			},
		},
	})
}

func TestListenerSetsRenderTheChartsSwitch(t *testing.T) {
	values, err := buildHelmValues(listenerSetLocals(true))
	if err != nil {
		t.Fatalf("buildHelmValues: %v", err)
	}
	if values["enableGatewayListenerSets"] != true {
		t.Errorf("enableGatewayListenerSets = %v, want true", values["enableGatewayListenerSets"])
	}
}

func TestListenerSetsOffRenderNothing(t *testing.T) {
	values, err := buildHelmValues(listenerSetLocals(false))
	if err != nil {
		t.Fatalf("buildHelmValues: %v", err)
	}
	if _, present := values["enableGatewayListenerSets"]; present {
		t.Errorf("enableGatewayListenerSets rendered while off: %v", values["enableGatewayListenerSets"])
	}
}

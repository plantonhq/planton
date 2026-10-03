package manifestprojection

import (
	"encoding/json"
	"reflect"
	"testing"

	kubernetesprometheusrulev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesprometheusrule/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// These tests pin the object a projection kind renders to: the envelope found
// by annotation and routed to metadata, the custom resource's spec under its
// upstream keys (including keys protojson would spell differently by default),
// and identity labels that win over a manifest's own labels.

func ptr[T any](v T) *T { return &v }

func ruleManifest() *kubernetesprometheusrulev1alpha1.KubernetesPrometheusRule {
	return &kubernetesprometheusrulev1alpha1.KubernetesPrometheusRule{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesPrometheusRule",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "api-slo",
			Id:   "k8sprule-123",
			Org:  "acme",
			Env:  "prod",
		},
		Spec: &kubernetesprometheusrulev1alpha1.KubernetesPrometheusRuleSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "monitoring"},
			},
			Labels: map[string]string{
				"release":                  "hub",
				"planton.ai/resource-name": "spoofed",
			},
			Annotations: map[string]string{"owner": "platform"},
			Groups: []*kubernetesprometheusrulev1alpha1.KubernetesPrometheusRuleGroup{{
				Name:                    "api",
				QueryOffset:             ptr("1m"),
				PartialResponseStrategy: ptr("warn"),
				Limit:                   ptr(int32(5)),
				Rules: []*kubernetesprometheusrulev1alpha1.KubernetesPrometheusRuleRule{{
					Alert:         ptr("ApiDown"),
					Expr:          "up == 0",
					ForDuration:   ptr("5m"),
					KeepFiringFor: ptr("10m"),
					Labels:        map[string]string{"severity": "page"},
				}},
			}},
		},
	}
}

func TestEnvelopeOf_FindsEveryEnvelopeFieldByAnnotation(t *testing.T) {
	spec := (&kubernetesprometheusrulev1alpha1.KubernetesPrometheusRuleSpec{}).ProtoReflect().Descriptor()
	got := EnvelopeOf(spec)
	want := Envelope{NamespaceKey: "namespace", LabelsKey: "labels", AnnotationsKey: "annotations"}
	if got != want {
		t.Fatalf("EnvelopeOf = %+v, want %+v", got, want)
	}
	if got.Holds("groups") {
		t.Error("groups is the custom resource's spec, not the envelope")
	}
}

func TestRender_RoutesTheEnvelopeAndKeepsUpstreamKeys(t *testing.T) {
	obj, err := Render(ruleManifest())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if obj.APIVersion != "monitoring.coreos.com/v1" || obj.Kind != "PrometheusRule" {
		t.Errorf("GVK = %s %s, want monitoring.coreos.com/v1 PrometheusRule", obj.APIVersion, obj.Kind)
	}
	if obj.Name != "api-slo" || obj.Namespace != "monitoring" {
		t.Errorf("identity = %s/%s, want monitoring/api-slo", obj.Namespace, obj.Name)
	}
	if !reflect.DeepEqual(obj.Annotations, map[string]string{"owner": "platform"}) {
		t.Errorf("annotations = %v", obj.Annotations)
	}

	wantLabels := map[string]string{
		"release":         "hub",
		LabelResource:     "true",
		LabelResourceName: "api-slo",
		LabelResourceKind: "KubernetesPrometheusRule",
		LabelResourceID:   "k8sprule-123",
		LabelOrganization: "acme",
		LabelEnvironment:  "prod",
	}
	if !reflect.DeepEqual(obj.Labels, wantLabels) {
		t.Errorf("labels = %v\nwant %v (identity labels must win over the manifest's)", obj.Labels, wantLabels)
	}

	gotSpec, _ := json.Marshal(obj.Spec)
	wantSpec := `{"groups":[{"limit":5,"name":"api","partial_response_strategy":"warn","query_offset":"1m","rules":[{"alert":"ApiDown","expr":"up == 0","for":"5m","keep_firing_for":"10m","labels":{"severity":"page"}}]}]}`
	if string(gotSpec) != wantSpec {
		t.Errorf("spec =\n%s\nwant\n%s", gotSpec, wantSpec)
	}
}

func TestRender_OmitsUnsetIdentityAndAnnotations(t *testing.T) {
	m := ruleManifest()
	m.Metadata = &shared.CatalogObjectMetadata{Name: "bare"}
	m.Spec.Labels = nil
	m.Spec.Annotations = nil
	obj, err := Render(m)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	want := map[string]string{LabelResource: "true", LabelResourceName: "bare", LabelResourceKind: "KubernetesPrometheusRule"}
	if !reflect.DeepEqual(obj.Labels, want) {
		t.Errorf("labels = %v, want %v", obj.Labels, want)
	}
	if obj.Annotations != nil {
		t.Errorf("annotations = %v, want none", obj.Annotations)
	}
	if _, ok := obj.Spec["labels"]; ok {
		t.Error("the envelope's labels leaked into the custom resource's spec")
	}
}

func TestRender_RefusesAKindThatIsNotAProjection(t *testing.T) {
	m := ruleManifest()
	m.Kind = "KubernetesKubePrometheusStack"
	if _, err := Render(m); err == nil {
		t.Fatal("Render accepted a kind with no kubernetes_manifest_projection")
	}
}

package generators

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"
	kubernetesprometheusrulev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesprometheusrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/kubernetes/manifestprojection"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// TestManifestProjection_BothEnginesSeeOneSpec pins the cross-engine contract
// of a Kubernetes manifest projection kind: the spec the generated Terraform
// module hands to kubectl_manifest (the tfvars `spec` minus the envelope keys
// the module's locals exclude) equals the spec the Pulumi helper applies
// (manifestprojection.Render). Both read one projection, so this fails only
// when the HCL writer or the envelope split diverges between engines.
func TestManifestProjection_BothEnginesSeeOneSpec(t *testing.T) {
	limit := int32(5)
	forDuration, keepFiring, offset := "5m", "10m", "1m"
	manifest := &kubernetesprometheusrulev1alpha1.KubernetesPrometheusRule{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesPrometheusRule",
		Metadata:   &shared.CloudResourceMetadata{Name: "api-slo"},
		Spec: &kubernetesprometheusrulev1alpha1.KubernetesPrometheusRuleSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "monitoring"},
			},
			Labels:      map[string]string{"release": "hub"},
			Annotations: map[string]string{"owner": "platform"},
			Groups: []*kubernetesprometheusrulev1alpha1.KubernetesPrometheusRuleGroup{{
				Name:        "api",
				QueryOffset: &offset,
				Limit:       &limit,
				Labels:      map[string]string{"component": "api"},
				Rules: []*kubernetesprometheusrulev1alpha1.KubernetesPrometheusRuleRule{{
					Alert:         strPtr("ApiDown"),
					Expr:          `sum(rate(errors_total{code=~"5.."}[5m])) > 0`,
					ForDuration:   &forDuration,
					KeepFiringFor: &keepFiring,
					Annotations:   map[string]string{"summary": "{{ $labels.job }} is down"},
				}},
			}},
		},
	}

	tfvars, err := ProtoToManifestTFVars(manifest)
	if err != nil {
		t.Fatalf("ProtoToManifestTFVars: %v", err)
	}
	file, diags := hclparse.NewParser().ParseHCL([]byte(tfvars), "terraform.tfvars")
	if diags.HasErrors() {
		t.Fatalf("parse tfvars: %v\n%s", diags, tfvars)
	}
	attrs, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		t.Fatalf("tfvars attributes: %v", diags)
	}
	specValue, diags := attrs["spec"].Expr.Value(nil)
	if diags.HasErrors() {
		t.Fatalf("tfvars spec value: %v", diags)
	}
	specJSON, err := ctyjson.Marshal(specValue, specValue.Type())
	if err != nil {
		t.Fatalf("marshal tfvars spec: %v", err)
	}
	var terraformSpec map[string]interface{}
	if err := json.Unmarshal(specJSON, &terraformSpec); err != nil {
		t.Fatal(err)
	}
	envelope := manifestprojection.EnvelopeOf(manifest.Spec.ProtoReflect().Descriptor())
	for key := range terraformSpec {
		if envelope.Holds(key) {
			delete(terraformSpec, key)
		}
	}

	obj, err := manifestprojection.Render(manifest)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	gotTF, _ := json.Marshal(terraformSpec)
	gotPulumi, _ := json.Marshal(obj.Spec)
	if string(gotTF) != string(gotPulumi) {
		t.Errorf("the engines see different specs\nterraform: %s\npulumi:    %s", gotTF, gotPulumi)
	}
}

func strPtr(s string) *string { return &s }

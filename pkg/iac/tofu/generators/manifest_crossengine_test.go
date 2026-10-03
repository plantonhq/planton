package generators

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/plantonhq/planton/catalog/kubernetes"
	kubernetespodmonitorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespodmonitor/v1alpha1"
	kubernetesprometheusrulev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesprometheusrule/v1alpha1"
	kubernetesservicemonitorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesservicemonitor/v1alpha1"
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
//
// The monitor cases carry every shape the projection reshapes or must keep:
// an IntOrString written as a number, list-valued maps written as bare lists
// (one of them holding Secret references), an explicitly empty selector, and
// an integer large enough that the HCL writer prints it in exponent form.
func TestManifestProjection_BothEnginesSeeOneSpec(t *testing.T) {
	limit := int32(5)
	forDuration, keepFiring, offset := "5m", "10m", "1m"
	manifest := &kubernetesprometheusrulev1alpha1.KubernetesPrometheusRule{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesPrometheusRule",
		Metadata:   &shared.CatalogObjectMetadata{Name: "api-slo"},
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
				Labels:      map[string]string{"kind": "api"},
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

	assertBothEnginesSeeOneSpec(t, manifest)

	sampleLimit, targetPort, proxyURL := uint32(1000000), "9090", "http://proxy:3128"
	assertBothEnginesSeeOneSpec(t, &kubernetesservicemonitorv1alpha1.KubernetesServiceMonitor{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesServiceMonitor",
		Metadata:   &shared.CatalogObjectMetadata{Name: "api"},
		Spec: &kubernetesservicemonitorv1alpha1.KubernetesServiceMonitorSpec{
			Namespace:   literalRef("monitoring"),
			Labels:      map[string]string{"release": "hub"},
			Selector:    &kubernetes.KubernetesPrometheusOperatorApiLabelSelector{},
			SampleLimit: &sampleLimit,
			NamespaceSelector: &kubernetes.KubernetesPrometheusOperatorApiNamespaceSelector{
				MatchNames: []*foreignkeyv1.StringValueOrRef{literalRef("api"), literalRef("workers")},
			},
			Endpoints: []*kubernetesservicemonitorv1alpha1.KubernetesServiceMonitorEndpoint{{
				TargetPort: &targetPort,
				Params: map[string]*kubernetes.KubernetesPrometheusOperatorApiStringList{
					"module": {Values: []string{"http_2xx"}},
				},
				Authorization: &kubernetes.KubernetesPrometheusOperatorApiSafeAuthorization{
					Credentials: &kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector{Name: literalRef("api-token"), Key: "token"},
				},
				ProxyUrl: &proxyURL,
				ProxyConnectHeader: map[string]*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelectorList{
					"Proxy-Authorization": {Values: []*kubernetes.KubernetesPrometheusOperatorApiSecretKeySelector{
						{Name: literalRef("proxy"), Key: "header"},
					}},
				},
				Relabelings: []*kubernetes.KubernetesPrometheusOperatorApiRelabelConfig{
					{SourceLabels: []string{"__address__"}, TargetLabel: "shard", Modulus: &limitU, Action: "hashmod"},
				},
			}},
		},
	})

	portNumber, portName := int32(9187), "metrics"
	assertBothEnginesSeeOneSpec(t, &kubernetespodmonitorv1alpha1.KubernetesPodMonitor{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesPodMonitor",
		Metadata:   &shared.CatalogObjectMetadata{Name: "orders-db"},
		Spec: &kubernetespodmonitorv1alpha1.KubernetesPodMonitorSpec{
			Namespace: literalRef("orders"),
			Selector: &kubernetes.KubernetesPrometheusOperatorApiLabelSelector{
				MatchLabels: map[string]string{"cnpg.io/cluster": "orders-db"},
			},
			PodMetricsEndpoints: []*kubernetespodmonitorv1alpha1.KubernetesPodMonitorPodMetricsEndpoint{
				{PortNumber: &portNumber},
				{Port: &portName, TargetPort: &portName},
			},
		},
	})
}

var limitU = uint32(4)

func literalRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func assertBothEnginesSeeOneSpec(t *testing.T, manifest manifestprojection.Manifest) {
	t.Helper()
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
	specField := manifest.ProtoReflect().Descriptor().Fields().ByName("spec")
	envelope := manifestprojection.EnvelopeOf(specField.Message())
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
		t.Errorf("%s: the engines see different specs\nterraform: %s\npulumi:    %s", manifest.GetKind(), gotTF, gotPulumi)
	}
}

func strPtr(s string) *string { return &s }

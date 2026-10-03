package module

import (
	"reflect"
	"testing"

	kubernetesvalkeyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesvalkey/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The probe contract of the rendered chart values. The chart (valkey 0.11.0)
// ships readinessProbe disabled, so without it a pod that is starting, or
// still loading its dataset, joins the Service at once; every probe it does
// render execs `valkey-cli ping`, which presents no client certificate, so
// under mutual TLS each probe fails the handshake and the pod restarts in a
// loop. The Terraform module renders the same blocks (iac/tf/locals.tf).

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valuesFor(t *testing.T, spec *kubernetesvalkeyv1alpha1.KubernetesValkeySpec) map[string]interface{} {
	t.Helper()
	spec.Namespace = literal("cache")
	locals := initializeLocals(nil, &kubernetesvalkeyv1alpha1.KubernetesValkeyIacInput{
		Target: &kubernetesvalkeyv1alpha1.KubernetesValkey{
			Metadata: &shared.CatalogObjectMetadata{Name: "sessions"},
			Spec:     spec,
		},
	})
	values, err := buildHelmValues(locals)
	if err != nil {
		t.Fatalf("buildHelmValues: %v", err)
	}
	return values
}

func replicationSpec() *kubernetesvalkeyv1alpha1.KubernetesValkeySpec {
	return &kubernetesvalkeyv1alpha1.KubernetesValkeySpec{
		Replication: &kubernetesvalkeyv1alpha1.KubernetesValkeyReplication{
			Persistence: &kubernetesvalkeyv1alpha1.KubernetesValkeyPersistence{Size: "5Gi"},
		},
	}
}

func tlsSpec(requireClientCertificate bool) *kubernetesvalkeyv1alpha1.KubernetesValkeySpec {
	return &kubernetesvalkeyv1alpha1.KubernetesValkeySpec{
		Tls: &kubernetesvalkeyv1alpha1.KubernetesValkeyTls{
			Enabled:                  true,
			CertificateSecret:        literal("sessions-tls"),
			RequireClientCertificate: requireClientCertificate,
		},
	}
}

func TestValues_ReadinessProbeOn(t *testing.T) {
	want := map[string]interface{}{"enabled": true}
	for name, spec := range map[string]*kubernetesvalkeyv1alpha1.KubernetesValkeySpec{
		"standalone":  {},
		"replication": replicationSpec(),
		"tls":         tlsSpec(false),
	} {
		t.Run(name, func(t *testing.T) {
			values := valuesFor(t, spec)
			if got := values["readinessProbe"]; !reflect.DeepEqual(got, want) {
				t.Fatalf("readinessProbe = %#v, want %#v", got, want)
			}
			for _, probe := range []string{"startupProbe", "livenessProbe"} {
				if got, ok := values[probe]; ok {
					t.Fatalf("%s rendered %#v; without mutual TLS the chart's valkey-cli ping probe is kept", probe, got)
				}
			}
		})
	}
}

func TestValues_MutualTlsProbesUseTcpSocket(t *testing.T) {
	tcpSocket := map[string]interface{}{
		"tcpSocket": map[string]interface{}{"port": "tcp"},
	}
	want := map[string]interface{}{
		"startupProbe":   map[string]interface{}{"customProbe": tcpSocket},
		"livenessProbe":  map[string]interface{}{"customProbe": tcpSocket},
		"readinessProbe": map[string]interface{}{"enabled": true, "customProbe": tcpSocket},
	}
	values := valuesFor(t, tlsSpec(true))
	for probe, block := range want {
		if got := values[probe]; !reflect.DeepEqual(got, block) {
			t.Fatalf("%s = %#v, want %#v", probe, got, block)
		}
	}
}

package kubernetesoteloperatorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesOtelOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesOtelOperator Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesOtelOperator Validation Tests", func() {
	var input *KubernetesOtelOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesOtelOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesOtelOperator",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "otel-operator",
			},
			Spec: &KubernetesOtelOperatorSpec{
				Namespace: literal("otel-system"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "otel-system", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the default webhook posture (chart-owned self-signed Issuer) should be valid", func() {
			input.Spec.Webhook = &KubernetesOtelOperatorWebhook{}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a webhook certificate signed by an explicit ClusterIssuer should be valid", func() {
			input.Spec.Webhook = &KubernetesOtelOperatorWebhook{
				IssuerRef: &KubernetesOtelOperatorIssuerRef{Kind: "ClusterIssuer", Name: "internal-ca"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a webhook certificate signed by a namespaced Issuer should be valid", func() {
			input.Spec.Webhook = &KubernetesOtelOperatorWebhook{
				IssuerRef: &KubernetesOtelOperatorIssuerRef{Kind: "Issuer", Name: "team-ca"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = stringPtr("0.120.0")
			input.Spec.Crds = &KubernetesOtelOperatorCrds{Install: boolPtr(false), KeepOnUninstall: boolPtr(false)}
			input.Spec.DefaultCollectorImage = "mirror.example.com/otel/opentelemetry-collector-k8s:0.156.0"
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.ServiceMonitorEnabled = true
			input.Spec.ImageRegistry = "mirror.example.com"
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			input.Spec.Scheduling = &KubernetesOtelOperatorScheduling{
				NodeSelector: map[string]string{"workload": "system"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a namespace-less spec should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an issuer_ref with a bad kind should fail", func() {
			input.Spec.Webhook = &KubernetesOtelOperatorWebhook{
				IssuerRef: &KubernetesOtelOperatorIssuerRef{Kind: "CertificateAuthority", Name: "internal-ca"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an issuer_ref without a name should fail", func() {
			input.Spec.Webhook = &KubernetesOtelOperatorWebhook{
				IssuerRef: &KubernetesOtelOperatorIssuerRef{Kind: "Issuer"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})

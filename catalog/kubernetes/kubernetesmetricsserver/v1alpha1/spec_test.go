package kubernetesmetricsserverv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesMetricsServer(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesMetricsServer Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32p(v int32) *int32                                                 { return &v }
func strp(v string) *string                                                 { return &v }
func tlsp(v KubernetesMetricsServerTlsType) *KubernetesMetricsServerTlsType { return &v }

var _ = ginkgo.Describe("KubernetesMetricsServer Validation Tests", func() {
	var input *KubernetesMetricsServer

	ginkgo.BeforeEach(func() {
		input = &KubernetesMetricsServer{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesMetricsServer",
			Metadata:   &shared.CatalogObjectMetadata{Name: "metrics-server"},
			Spec: &KubernetesMetricsServerSpec{
				Namespace: literal("kube-system"),
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a minimal install", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts the self-signed-kubelet posture (kind/k3s)", func() {
			input.Spec.KubeletInsecureTls = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts reordered kubelet address types", func() {
			input.Spec.KubeletPreferredAddressTypes = []string{"Hostname", "InternalIP"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a tuned metric resolution", func() {
			input.Spec.MetricResolution = strp("30s")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a cert-manager issued serving certificate", func() {
			input.Spec.Tls = &KubernetesMetricsServerTls{
				Type: tlsp(KubernetesMetricsServerTlsType_cert_manager),
				CertManagerIssuer: &kubernetes.CertManagerIssuerRef{
					IssuerType: &kubernetes.CertManagerIssuerRef_Issuer{
						Issuer: &kubernetes.CertManagerNamespacedIssuerRef{Name: literal("metrics-server-issuer")},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts an existing serving-certificate secret", func() {
			input.Spec.Tls = &KubernetesMetricsServerTls{
				Type:               tlsp(KubernetesMetricsServerTlsType_existing_secret),
				ExistingSecretName: literal("metrics-server-tls"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts HA replicas with a PodDisruptionBudget", func() {
			input.Spec.Replicas = int32p(2)
			input.Spec.PodDisruptionBudget = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts prometheus telemetry with a ServiceMonitor", func() {
			input.Spec.Prometheus = &KubernetesMetricsServerPrometheus{
				Enabled:        true,
				ServiceMonitor: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields and contracts", func() {
		ginkgo.It("rejects a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unknown kubelet address type", func() {
			input.Spec.KubeletPreferredAddressTypes = []string{"NodeIP"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a malformed metric resolution", func() {
			input.Spec.MetricResolution = strp("15")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects tls existing_secret without a secret name", func() {
			input.Spec.Tls = &KubernetesMetricsServerTls{
				Type: tlsp(KubernetesMetricsServerTlsType_existing_secret),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a cert-manager issuer on a non-cert-manager tls type", func() {
			input.Spec.Tls = &KubernetesMetricsServerTls{
				Type: tlsp(KubernetesMetricsServerTlsType_helm),
				CertManagerIssuer: &kubernetes.CertManagerIssuerRef{
					IssuerType: &kubernetes.CertManagerIssuerRef_ClusterIssuer{
						ClusterIssuer: &kubernetes.CertManagerClusterIssuerRef{Name: literal("some-issuer")},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an existing-secret reference on a non-existing-secret tls type", func() {
			input.Spec.Tls = &KubernetesMetricsServerTls{
				Type:               tlsp(KubernetesMetricsServerTlsType_helm),
				ExistingSecretName: literal("metrics-server-tls"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a cert-manager issuer that names no grain", func() {
			input.Spec.Tls = &KubernetesMetricsServerTls{
				Type:              tlsp(KubernetesMetricsServerTlsType_cert_manager),
				CertManagerIssuer: &kubernetes.CertManagerIssuerRef{},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("An issuer is required"))
		})

		ginkgo.It("rejects a cert-manager issuer arm without a name", func() {
			input.Spec.Tls = &KubernetesMetricsServerTls{
				Type: tlsp(KubernetesMetricsServerTlsType_cert_manager),
				CertManagerIssuer: &kubernetes.CertManagerIssuerRef{
					IssuerType: &kubernetes.CertManagerIssuerRef_ClusterIssuer{ClusterIssuer: &kubernetes.CertManagerClusterIssuerRef{}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a ServiceMonitor without prometheus enabled", func() {
			input.Spec.Prometheus = &KubernetesMetricsServerPrometheus{
				ServiceMonitor: true,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects zero replicas", func() {
			input.Spec.Replicas = int32p(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})

package kubernetescertmanagerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesCertManager(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesCertManager Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

var _ = ginkgo.Describe("KubernetesCertManager Validation Tests", func() {
	var input *KubernetesCertManager

	ginkgo.BeforeEach(func() {
		input = &KubernetesCertManager{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesCertManager",
			Metadata:   &shared.CatalogObjectMetadata{Name: "cert-manager"},
			Spec: &KubernetesCertManagerSpec{
				Namespace:       literal("cert-manager"),
				CreateNamespace: true,
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a minimal install", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts EKS workload identity for keyless DNS-01", func() {
			input.Spec.WorkloadIdentity = &kubernetesprovider.KubernetesWorkloadIdentity{
				Provider: &kubernetesprovider.KubernetesWorkloadIdentity_Eks{
					Eks: &kubernetesprovider.KubernetesWorkloadIdentityEksIrsa{
						RoleArn: literal("arn:aws:iam::123456789012:role/cert-manager"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts DNS-01 self-check resolvers", func() {
			input.Spec.Dns01SelfCheck = &KubernetesCertManagerDns01SelfCheck{
				RecursiveNameservers:     []string{"1.1.1.1:53", "8.8.8.8:53"},
				RecursiveNameserversOnly: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts webhook host-network tuning", func() {
			replicas := int32(2)
			timeout := int32(20)
			port := int32(10260)
			input.Spec.Webhook = &KubernetesCertManagerWebhook{
				Replicas:       &replicas,
				TimeoutSeconds: &timeout,
				HostNetwork:    true,
				SecurePort:     &port,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields and contracts", func() {
		ginkgo.It("rejects a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a log level beyond the range", func() {
			level := int32(7)
			input.Spec.LogLevel = &level
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects zero controller replicas", func() {
			replicas := int32(0)
			input.Spec.Replicas = &replicas
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a nameserver without a port", func() {
			input.Spec.Dns01SelfCheck = &KubernetesCertManagerDns01SelfCheck{
				RecursiveNameservers: []string{"1.1.1.1"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects recursive_nameservers_only without resolvers", func() {
			input.Spec.Dns01SelfCheck = &KubernetesCertManagerDns01SelfCheck{
				RecursiveNameserversOnly: true,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a webhook timeout beyond the API server maximum", func() {
			timeout := int32(31)
			input.Spec.Webhook = &KubernetesCertManagerWebhook{TimeoutSeconds: &timeout}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an AKS workload identity with a malformed tenant id", func() {
			tenantId := "not-a-guid"
			input.Spec.WorkloadIdentity = &kubernetesprovider.KubernetesWorkloadIdentity{
				Provider: &kubernetesprovider.KubernetesWorkloadIdentity_Aks{
					Aks: &kubernetesprovider.KubernetesWorkloadIdentityAks{
						ClientId: literal("11111111-1111-1111-1111-111111111111"),
						TenantId: &tenantId,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})

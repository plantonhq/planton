package kubernetesexternalsecretsoperatorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesExternalSecretsOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesExternalSecretsOperator Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

var _ = ginkgo.Describe("KubernetesExternalSecretsOperator Validation Tests", func() {
	var input *KubernetesExternalSecretsOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesExternalSecretsOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesExternalSecretsOperator",
			Metadata:   &shared.CatalogObjectMetadata{Name: "external-secrets"},
			Spec: &KubernetesExternalSecretsOperatorSpec{
				Namespace:       literal("external-secrets"),
				CreateNamespace: true,
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a minimal install", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts an HA install with leader election", func() {
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.LeaderElect = true
			input.Spec.PodDisruptionBudget = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts EKS workload identity for ambient store access", func() {
			input.Spec.WorkloadIdentity = &kubernetesprovider.KubernetesWorkloadIdentity{
				Provider: &kubernetesprovider.KubernetesWorkloadIdentity_Eks{
					Eks: &kubernetesprovider.KubernetesWorkloadIdentityEksIrsa{
						RoleArn: literal("arn:aws:iam::123456789012:role/external-secrets"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a namespace-scoped install with scoped RBAC", func() {
			input.Spec.ScopedNamespace = "team-a"
			input.Spec.ScopedRbac = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts component tuning and CRD lifecycle knobs", func() {
			keep := false
			install := true
			input.Spec.Crds = &KubernetesExternalSecretsOperatorCrds{Install: &install, KeepOnUninstall: &keep}
			input.Spec.Concurrent = int32Ptr(5)
			input.Spec.Webhook = &KubernetesExternalSecretsOperatorWebhook{Replicas: int32Ptr(2)}
			input.Spec.CertController = &KubernetesExternalSecretsOperatorCertController{Replicas: int32Ptr(1)}
			input.Spec.Prometheus = &KubernetesExternalSecretsOperatorPrometheus{
				ServiceMonitor:       true,
				ServiceMonitorLabels: map[string]string{"release": "kube-prometheus-stack"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields and contracts", func() {
		ginkgo.It("rejects a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects replicas > 1 without leader election", func() {
			input.Spec.Replicas = int32Ptr(3)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects zero controller replicas", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects scoped_rbac without scoped_namespace", func() {
			input.Spec.ScopedRbac = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an unknown log level", func() {
			bad := "verbose"
			input.Spec.LogLevel = &bad
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects zero webhook replicas", func() {
			input.Spec.Webhook = &KubernetesExternalSecretsOperatorWebhook{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects zero cert-controller replicas", func() {
			input.Spec.CertController = &KubernetesExternalSecretsOperatorCertController{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a wrong kind constant", func() {
			input.Kind = "KubernetesExternalSecrets"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})

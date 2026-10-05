package kubernetesgatekeeperv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesGatekeeper(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGatekeeper Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

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

var _ = ginkgo.Describe("KubernetesGatekeeper Validation Tests", func() {
	var input *KubernetesGatekeeper

	ginkgo.BeforeEach(func() {
		input = &KubernetesGatekeeper{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesGatekeeper",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "gatekeeper",
			},
			Spec: &KubernetesGatekeeperSpec{
				Namespace: literal("gatekeeper-system"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec should be valid (chart defaults)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "gatekeeper-system", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a fail-closed validating webhook should be valid", func() {
			input.Spec.ValidatingWebhook = &KubernetesGatekeeperValidatingWebhook{
				Enabled:                  boolPtr(true),
				FailurePolicy:            strPtr("Fail"),
				TimeoutSeconds:           int32Ptr(5),
				EnableDeleteOperations:   true,
				CheckIgnoreFailurePolicy: strPtr("Ignore"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a mutating webhook with annotations should be valid", func() {
			input.Spec.MutatingWebhook = &KubernetesGatekeeperMutatingWebhook{
				Enabled:             boolPtr(true),
				FailurePolicy:       strPtr("Ignore"),
				TimeoutSeconds:      int32Ptr(3),
				MutationAnnotations: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an audit block should be valid", func() {
			input.Spec.Audit = &KubernetesGatekeeperAudit{
				IntervalSeconds:           int32Ptr(120),
				ConstraintViolationsLimit: int32Ptr(50),
				FromCache:                 true,
				MatchKindOnly:             true,
				ChunkSize:                 int32Ptr(200),
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "256Mi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "1", Memory: "1Gi"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("audit interval zero (run-once) should be valid", func() {
			input.Spec.Audit = &KubernetesGatekeeperAudit{IntervalSeconds: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("exempt namespaces and prefixes should be valid", func() {
			input.Spec.ExemptNamespaces = []string{"platform-system"}
			input.Spec.ExemptNamespacePrefixes = []string{"kube-"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an engine block should be valid", func() {
			input.Spec.Engine = &KubernetesGatekeeperEngine{
				EnableExternalData:               boolPtr(false),
				EnableK8SNativeValidation:        boolPtr(true),
				EnableGeneratorResourceExpansion: boolPtr(false),
				DisabledBuiltins:                 []string{"{http.send}"},
				LogDenies:                        true,
				LogLevel:                         strPtr("DEBUG"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an external certificate from a Certificate reference should be valid", func() {
			input.Spec.ExternalCert = &KubernetesGatekeeperExternalCert{
				SecretName: valueFrom(catalogkind.CatalogKind_KubernetesCertificate, "gatekeeper-webhook-cert", "spec.secretName"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a hooks block should be valid", func() {
			input.Spec.Hooks = &KubernetesGatekeeperHooks{
				LabelNamespace:                         boolPtr(true),
				ProbeWebhook:                           boolPtr(false),
				UpgradeCrds:                            boolPtr(true),
				DeleteWebhookConfigurationsOnUninstall: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should be invalid", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should be invalid", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a lowercase failure policy should be invalid", func() {
			input.Spec.ValidatingWebhook = &KubernetesGatekeeperValidatingWebhook{
				FailurePolicy: strPtr("fail"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a validating webhook timeout above 30s should be invalid (API-server cap)", func() {
			input.Spec.ValidatingWebhook = &KubernetesGatekeeperValidatingWebhook{
				TimeoutSeconds: int32Ptr(31),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a mutating webhook timeout of zero should be invalid", func() {
			input.Spec.MutatingWebhook = &KubernetesGatekeeperMutatingWebhook{
				TimeoutSeconds: int32Ptr(0),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a negative audit interval should be invalid", func() {
			input.Spec.Audit = &KubernetesGatekeeperAudit{IntervalSeconds: int32Ptr(-1)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero constraint-violations limit should be invalid", func() {
			input.Spec.Audit = &KubernetesGatekeeperAudit{ConstraintViolationsLimit: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase exempt namespace should be invalid", func() {
			input.Spec.ExemptNamespaces = []string{"Kube-System"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown engine log level should be invalid", func() {
			input.Spec.Engine = &KubernetesGatekeeperEngine{LogLevel: strPtr("TRACE")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an external-cert block without a secret name should be invalid", func() {
			input.Spec.ExternalCert = &KubernetesGatekeeperExternalCert{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})

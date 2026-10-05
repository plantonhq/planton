package kubernetesbackendtlspolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesBackendTlsPolicy(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesBackendTlsPolicy Suite")
}

func stringPtr(s string) *string { return &s }

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

// serviceTarget returns a Core-support target reference to a Service.
func serviceTarget(name string) *KubernetesBackendTlsPolicyTargetReference {
	return &KubernetesBackendTlsPolicyTargetReference{
		Group: stringPtr(""),
		Kind:  "Service",
		Name:  literal(name),
	}
}

// caBundleRef returns the Core-support ConfigMap CA reference.
func caBundleRef(name string) *KubernetesBackendTlsPolicyCaCertificateReference {
	return &KubernetesBackendTlsPolicyCaCertificateReference{
		Group: stringPtr(""),
		Kind:  "ConfigMap",
		Name:  literal(name),
	}
}

var _ = ginkgo.Describe("KubernetesBackendTlsPolicy Validation Tests", func() {
	var input *KubernetesBackendTlsPolicy

	ginkgo.BeforeEach(func() {
		input = &KubernetesBackendTlsPolicy{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesBackendTlsPolicy",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-backend-tls",
			},
			Spec: &KubernetesBackendTlsPolicySpec{
				Namespace:  literal("app-ns"),
				TargetRefs: []*KubernetesBackendTlsPolicyTargetReference{serviceTarget("payments")},
				Validation: &KubernetesBackendTlsPolicyValidation{
					CaCertificateRefs: []*KubernetesBackendTlsPolicyCaCertificateReference{caBundleRef("backend-ca")},
					Hostname:          "payments.internal.example.com",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal CA-bundle policy should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("system trust store arm should be valid", func() {
			input.Spec.Validation.CaCertificateRefs = nil
			input.Spec.Validation.WellKnownCaCertificates = stringPtr("System")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("implementation-specific well-known CA set should be valid", func() {
			input.Spec.Validation.CaCertificateRefs = nil
			input.Spec.Validation.WellKnownCaCertificates = stringPtr("mycompany.com/my-custom-ca-certificates")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface with FK refs, section names, SANs and options should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "app-ns", "spec.name")
			input.Spec.TargetRefs = []*KubernetesBackendTlsPolicyTargetReference{
				{
					Group:       stringPtr(""),
					Kind:        "Service",
					Name:        valueFrom(catalogkind.CatalogKind_KubernetesService, "payments", "status.outputs.service_name"),
					SectionName: stringPtr("https"),
				},
				{
					Group:       stringPtr(""),
					Kind:        "Service",
					Name:        valueFrom(catalogkind.CatalogKind_KubernetesService, "payments", "status.outputs.service_name"),
					SectionName: stringPtr("grpc"),
				},
			}
			input.Spec.Validation.CaCertificateRefs = []*KubernetesBackendTlsPolicyCaCertificateReference{
				{
					Group: stringPtr(""),
					Kind:  "ConfigMap",
					Name:  valueFrom(catalogkind.CatalogKind_KubernetesConfigMap, "backend-ca", "status.outputs.configmap_name"),
				},
			}
			input.Spec.Validation.SubjectAltNames = []*KubernetesBackendTlsPolicySubjectAltName{
				{Type: "Hostname", Hostname: stringPtr("payments.internal.example.com")},
				{Type: "URI", Uri: stringPtr("spiffe://cluster.example.com/ns/app-ns/sa/payments")},
			}
			input.Spec.Options = map[string]string{"vendor.example.com/min-tls-version": "1.3"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("two refs to the same target with distinct section names should be valid", func() {
			input.Spec.TargetRefs = []*KubernetesBackendTlsPolicyTargetReference{
				{Group: stringPtr(""), Kind: "Service", Name: literal("payments"), SectionName: stringPtr("https")},
				{Group: stringPtr(""), Kind: "Service", Name: literal("payments"), SectionName: stringPtr("grpc")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("wildcard SAN hostname should be valid", func() {
			input.Spec.Validation.SubjectAltNames = []*KubernetesBackendTlsPolicySubjectAltName{
				{Type: "Hostname", Hostname: stringPtr("*.internal.example.com")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero target refs should fail (min_items=1)", func() {
			input.Spec.TargetRefs = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("target ref without group key should fail (presence-required)", func() {
			input.Spec.TargetRefs = []*KubernetesBackendTlsPolicyTargetReference{
				{Kind: "Service", Name: literal("payments")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("target ref without kind should fail", func() {
			input.Spec.TargetRefs = []*KubernetesBackendTlsPolicyTargetReference{
				{Group: stringPtr(""), Name: literal("payments")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("missing validation block should fail", func() {
			input.Spec.Validation = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("missing hostname should fail", func() {
			input.Spec.Validation.Hostname = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("uppercase hostname should fail the DNS pattern", func() {
			input.Spec.Validation.Hostname = "Payments.Example.Com"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("both trust anchors set should fail (mutual exclusion)", func() {
			input.Spec.Validation.WellKnownCaCertificates = stringPtr("System")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("neither trust anchor set should fail (one required)", func() {
			input.Spec.Validation.CaCertificateRefs = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("well-known CA value other than System without a domain prefix should fail", func() {
			input.Spec.Validation.CaCertificateRefs = nil
			input.Spec.Validation.WellKnownCaCertificates = stringPtr("Custom")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("more than 8 CA refs should fail (max_items=8)", func() {
			refs := make([]*KubernetesBackendTlsPolicyCaCertificateReference, 0, 9)
			for i := 0; i < 9; i++ {
				refs = append(refs, caBundleRef("backend-ca"))
			}
			input.Spec.Validation.CaCertificateRefs = refs
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("CA ref without group key should fail (presence-required)", func() {
			input.Spec.Validation.CaCertificateRefs = []*KubernetesBackendTlsPolicyCaCertificateReference{
				{Kind: "ConfigMap", Name: literal("backend-ca")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate refs to the same target without section names should fail (uniqueness)", func() {
			input.Spec.TargetRefs = []*KubernetesBackendTlsPolicyTargetReference{
				serviceTarget("payments"),
				serviceTarget("payments"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("mixed section-name presence on refs to the same target should fail", func() {
			input.Spec.TargetRefs = []*KubernetesBackendTlsPolicyTargetReference{
				serviceTarget("payments"),
				{Group: stringPtr(""), Kind: "Service", Name: literal("payments"), SectionName: stringPtr("https")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate section name on refs to the same target should fail", func() {
			input.Spec.TargetRefs = []*KubernetesBackendTlsPolicyTargetReference{
				{Group: stringPtr(""), Kind: "Service", Name: literal("payments"), SectionName: stringPtr("https")},
				{Group: stringPtr(""), Kind: "Service", Name: literal("payments"), SectionName: stringPtr("https")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("SAN type Hostname without hostname should fail", func() {
			input.Spec.Validation.SubjectAltNames = []*KubernetesBackendTlsPolicySubjectAltName{
				{Type: "Hostname"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("SAN type URI carrying a hostname should fail", func() {
			input.Spec.Validation.SubjectAltNames = []*KubernetesBackendTlsPolicySubjectAltName{
				{Type: "URI", Uri: stringPtr("spiffe://c/ns/a/sa/b"), Hostname: stringPtr("a.example.com")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("SAN type URI without uri should fail", func() {
			input.Spec.Validation.SubjectAltNames = []*KubernetesBackendTlsPolicySubjectAltName{
				{Type: "URI"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("SAN with unknown type should fail (closed enum)", func() {
			input.Spec.Validation.SubjectAltNames = []*KubernetesBackendTlsPolicySubjectAltName{
				{Type: "Email", Hostname: stringPtr("a.example.com")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("more than 5 SANs should fail (max_items=5)", func() {
			sans := make([]*KubernetesBackendTlsPolicySubjectAltName, 0, 6)
			for i := 0; i < 6; i++ {
				sans = append(sans, &KubernetesBackendTlsPolicySubjectAltName{
					Type: "Hostname", Hostname: stringPtr("a.example.com"),
				})
			}
			input.Spec.Validation.SubjectAltNames = sans
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("option key without valid annotation-key shape should fail", func() {
			input.Spec.Options = map[string]string{"-bad key-": "x"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("uppercase SAN hostname should fail the DNS pattern", func() {
			input.Spec.Validation.SubjectAltNames = []*KubernetesBackendTlsPolicySubjectAltName{
				{Type: "Hostname", Hostname: stringPtr("Payments.Example.Com")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})

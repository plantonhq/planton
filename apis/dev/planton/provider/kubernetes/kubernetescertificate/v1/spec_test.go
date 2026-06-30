package kubernetescertificatev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func stringPtr(s string) *string { return &s }

func int32Ptr(i int32) *int32 { return &i }

func TestKubernetesCertificate(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesCertificate Suite")
}

var _ = ginkgo.Describe("KubernetesCertificate Validation Tests", func() {
	var input *KubernetesCertificate

	ginkgo.BeforeEach(func() {
		input = &KubernetesCertificate{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesCertificate",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-certificate",
			},
			Spec: &KubernetesCertificateSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "default",
					},
				},
				DnsNames:   []string{"example.com"},
				SecretName: "example-tls",
				IssuerRef: &CertificateIssuerRef{
					IssuerType: &CertificateIssuerRef_ClusterIssuer{
						ClusterIssuer: &ClusterIssuerRef{
							Name: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "example.com",
								},
							},
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with ClusterIssuer reference", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with namespaced Issuer reference", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.IssuerRef = &CertificateIssuerRef{
					IssuerType: &CertificateIssuerRef_Issuer{
						Issuer: &NamespacedIssuerRef{
							Name: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "my-ca-issuer",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with CA certificate", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.IsCa = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom duration", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.DurationConfig = &CertificateDuration{
					Duration:    stringPtr("8760h"),
					RenewBefore: stringPtr("720h"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom private key", func() {
			ginkgo.It("should not return a validation error", func() {
				alg := CertificatePrivateKey_ecdsa
				size := int32Ptr(256)
				enc := CertificatePrivateKey_pkcs8
				rot := CertificatePrivateKey_always
				input.Spec.PrivateKey = &CertificatePrivateKey{
					Algorithm:      &alg,
					Size:           size,
					Encoding:       &enc,
					RotationPolicy: &rot,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with multiple DNS names", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.DnsNames = []string{"example.com", "www.example.com"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing issuer_ref", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.IssuerRef = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing issuer_ref type", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.IssuerRef = &CertificateIssuerRef{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing dns_names", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.DnsNames = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("empty secret_name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.SecretName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("empty namespace value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing ClusterIssuer name in ClusterIssuerRef", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.IssuerRef = &CertificateIssuerRef{
					IssuerType: &CertificateIssuerRef_ClusterIssuer{
						ClusterIssuer: &ClusterIssuerRef{
							Name: nil,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

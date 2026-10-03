package kubernetescertificatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestKubernetesCertificate(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesCertificate Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

var _ = ginkgo.Describe("KubernetesCertificate Validation Tests", func() {
	var input *KubernetesCertificate

	ginkgo.BeforeEach(func() {
		input = &KubernetesCertificate{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesCertificate",
			Metadata:   &shared.CatalogObjectMetadata{Name: "test-certificate"},
			Spec: &KubernetesCertificateSpec{
				Namespace:  literal("team-a"),
				SecretName: "test-certificate-tls",
				IssuerRef: &KubernetesCertificateIssuerRef{
					IssuerType: &KubernetesCertificateIssuerRef_ClusterIssuer{
						ClusterIssuer: &KubernetesCertificateClusterIssuerRef{Name: literal("letsencrypt")},
					},
				},
				DnsNames: []string{"api.example.com"},
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a minimal DNS certificate from a cluster issuer", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a namespace-scoped issuer reference", func() {
			input.Spec.IssuerRef = &KubernetesCertificateIssuerRef{
				IssuerType: &KubernetesCertificateIssuerRef_Issuer{
					Issuer: &KubernetesCertificateNamespacedIssuerRef{Name: literal("team-a-ca")},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts an external issuer reference", func() {
			input.Spec.IssuerRef = &KubernetesCertificateIssuerRef{
				IssuerType: &KubernetesCertificateIssuerRef_External{
					External: &KubernetesCertificateExternalIssuerRef{
						Group: "awspca.cert-manager.io",
						Kind:  "AWSPCAClusterIssuer",
						Name:  "private-ca",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a CA certificate with name constraints", func() {
			input.Spec.IsCa = true
			input.Spec.NameConstraints = &KubernetesCertificateNameConstraints{
				Critical: true,
				Permitted: &KubernetesCertificateNameConstraintSet{
					DnsDomains: []string{"internal.example.com"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts valid usages from the x509 vocabulary", func() {
			input.Spec.Usages = []string{"digital signature", "key encipherment", "server auth"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts an ECDSA key with a matching size", func() {
			algorithm := KubernetesCertificatePrivateKey_ecdsa
			size := int32(384)
			input.Spec.PrivateKey = &KubernetesCertificatePrivateKey{Algorithm: &algorithm, Size: &size}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a URI-only (SPIFFE) certificate", func() {
			input.Spec.DnsNames = nil
			input.Spec.Uris = []string{"spiffe://cluster.local/ns/team-a/sa/api"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts keystore outputs with inline passwords", func() {
			input.Spec.Keystores = &KubernetesCertificateKeystores{
				Jks:    &KubernetesCertificateJksKeystore{Create: true, Password: "changeit"},
				Pkcs12: &KubernetesCertificatePkcs12Keystore{Create: true, Password: "changeit"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields and contracts", func() {
		ginkgo.It("rejects a missing secret_name", func() {
			input.Spec.SecretName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a missing issuer_ref", func() {
			input.Spec.IssuerRef = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an issuer_ref with no arm selected", func() {
			input.Spec.IssuerRef = &KubernetesCertificateIssuerRef{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a certificate with no requested names at all", func() {
			input.Spec.DnsNames = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects literal_subject combined with subject", func() {
			input.Spec.LiteralSubject = "CN=api,O=acme"
			input.Spec.Subject = &KubernetesCertificateSubject{Organizations: []string{"acme"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects literal_subject combined with common_name", func() {
			input.Spec.LiteralSubject = "CN=api,O=acme"
			input.Spec.CommonName = "api.example.com"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects renew_before combined with renew_before_percentage", func() {
			percentage := int32(25)
			input.Spec.RenewBefore = "360h"
			input.Spec.RenewBeforePercentage = &percentage
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an invalid duration format", func() {
			bad := "90days"
			input.Spec.Duration = &bad
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a usage outside the x509 vocabulary", func() {
			input.Spec.Usages = []string{"server auth", "quantum signing"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an invalid signature algorithm", func() {
			bad := "SHA1WithRSA"
			input.Spec.SignatureAlgorithm = &bad
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an RSA key with an ECDSA-family size", func() {
			algorithm := KubernetesCertificatePrivateKey_rsa
			size := int32(256)
			input.Spec.PrivateKey = &KubernetesCertificatePrivateKey{Algorithm: &algorithm, Size: &size}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an Ed25519 key with any size", func() {
			algorithm := KubernetesCertificatePrivateKey_ed25519
			size := int32(256)
			input.Spec.PrivateKey = &KubernetesCertificatePrivateKey{Algorithm: &algorithm, Size: &size}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a JKS keystore without a password", func() {
			input.Spec.Keystores = &KubernetesCertificateKeystores{
				Jks: &KubernetesCertificateJksKeystore{Create: true},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an invalid pkcs12 profile", func() {
			bad := "modern2024"
			input.Spec.Keystores = &KubernetesCertificateKeystores{
				Pkcs12: &KubernetesCertificatePkcs12Keystore{Create: true, Password: "changeit", Profile: &bad},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an invalid additional output format", func() {
			input.Spec.AdditionalOutputFormats = []*KubernetesCertificateAdditionalOutputFormat{{Type: "pem"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an over-long common name", func() {
			input.Spec.CommonName = string(make([]byte, 65))
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a renew_before_percentage out of range", func() {
			percentage := int32(100)
			clone := proto.Clone(input).(*KubernetesCertificate)
			clone.Spec.RenewBeforePercentage = &percentage
			gomega.Expect(protovalidate.Validate(clone)).NotTo(gomega.Succeed())
		})
	})
})

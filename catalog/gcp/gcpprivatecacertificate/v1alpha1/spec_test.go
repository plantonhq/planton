package gcpprivatecacertificatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpPrivateCaCertificateSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpPrivateCaCertificateSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	fromConfig := func() *GcpPrivateCaCertificate {
		return &GcpPrivateCaCertificate{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpPrivateCaCertificate",
			Metadata:   &shared.CloudResourceMetadata{Name: "api-server"},
			Spec: &GcpPrivateCaCertificateSpec{
				Location: "us-central1",
				Pool:     litRef("internal-tls"),
				Config: &GcpPrivateCaCertificateConfig{
					SubjectConfig: &GcpPrivateCaCertificateSubjectConfig{
						Subject:        &GcpPrivateCaCertificateSubject{CommonName: "api.internal.example.com"},
						SubjectAltName: &GcpPrivateCaCertificateSubjectAltName{DnsNames: []string{"api.internal.example.com"}},
					},
					X509Config: &GcpPrivateCaCertificateX509Parameters{
						KeyUsage: &GcpPrivateCaCertificateKeyUsage{
							BaseKeyUsage:     &GcpPrivateCaCertificateBaseKeyUsage{DigitalSignature: true, KeyEncipherment: true},
							ExtendedKeyUsage: &GcpPrivateCaCertificateExtendedKeyUsage{ServerAuth: true},
						},
					},
					PublicKey: &GcpPrivateCaCertificatePublicKey{Key: "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0="},
				},
			},
		}
	}

	ginkgo.It("should accept a config-described certificate", func() {
		gomega.Expect(validator.Validate(fromConfig())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a CSR with every reference and lever", func() {
		r := fromConfig()
		r.Spec.Config = nil
		r.Spec.PemCsr = "-----BEGIN CERTIFICATE REQUEST-----"
		r.Spec.ProjectId = litRef("p")
		r.Spec.CertificateId = "api-server-2026-09"
		r.Spec.CertificateAuthority = litRef("root-ca")
		r.Spec.CertificateTemplate = litRef("projects/p/locations/us-central1/certificateTemplates/tls-server")
		r.Spec.Lifetime = "2592000s"
		r.Spec.Labels = map[string]string{"app": "api"}
		r.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	})

	ginkgo.It("should take exactly one of a CSR or config", func() {
		r := fromConfig()
		r.Spec.PemCsr = "-----BEGIN CERTIFICATE REQUEST-----"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "both")
		r = fromConfig()
		r.Spec.Config = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "neither")
	})

	ginkgo.It("should require the pool, the subject, the X.509 fields, and the public key", func() {
		mutations := map[string]func(*GcpPrivateCaCertificateSpec){
			"pool":        func(s *GcpPrivateCaCertificateSpec) { s.Pool = nil },
			"subject":     func(s *GcpPrivateCaCertificateSpec) { s.Config.SubjectConfig = nil },
			"common name": func(s *GcpPrivateCaCertificateSpec) { s.Config.SubjectConfig.Subject.CommonName = "" },
			"x509":        func(s *GcpPrivateCaCertificateSpec) { s.Config.X509Config = nil },
			"public key":  func(s *GcpPrivateCaCertificateSpec) { s.Config.PublicKey = nil },
			"key bytes":   func(s *GcpPrivateCaCertificateSpec) { s.Config.PublicKey.Key = "" },
		}
		for name, mutate := range mutations {
			r := fromConfig()
			mutate(r.Spec)
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), name)
		}
	})

	ginkgo.It("should reject an unknown key format, lifetime, id, or deletion policy", func() {
		mutations := map[string]func(*GcpPrivateCaCertificateSpec){
			"format":          func(s *GcpPrivateCaCertificateSpec) { s.Config.PublicKey.Format = "DER" },
			"lifetime":        func(s *GcpPrivateCaCertificateSpec) { s.Lifetime = "1y" },
			"id":              func(s *GcpPrivateCaCertificateSpec) { s.CertificateId = "api.server" },
			"deletion_policy": func(s *GcpPrivateCaCertificateSpec) { s.DeletionPolicy = "REVOKE" },
		}
		for name, mutate := range mutations {
			r := fromConfig()
			mutate(r.Spec)
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), name)
		}
	})
})

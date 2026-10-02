package gcpprivatecacertificateauthorityv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpPrivateCaCertificateAuthoritySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpPrivateCaCertificateAuthoritySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	root := func() *GcpPrivateCaCertificateAuthority {
		return &GcpPrivateCaCertificateAuthority{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpPrivateCaCertificateAuthority",
			Metadata:   &shared.CloudResourceMetadata{Name: "root-ca"},
			Spec: &GcpPrivateCaCertificateAuthoritySpec{
				Location: "us-central1",
				Pool:     litRef("internal-tls"),
				Config: &GcpPrivateCaCertificateAuthorityConfig{
					SubjectConfig: &GcpPrivateCaCertificateAuthoritySubjectConfig{
						Subject: &GcpPrivateCaCertificateAuthoritySubject{CommonName: "Example Root CA", Organization: "Example"},
					},
					X509Config: &GcpPrivateCaCertificateAuthorityX509Parameters{
						CaOptions: &GcpPrivateCaCertificateAuthorityCaOptions{IsCa: proto.Bool(true)},
						KeyUsage: &GcpPrivateCaCertificateAuthorityKeyUsage{
							BaseKeyUsage: &GcpPrivateCaCertificateAuthorityBaseKeyUsage{CertSign: true, CrlSign: true},
						},
					},
				},
				KeySpec: &GcpPrivateCaCertificateAuthorityKeySpec{Algorithm: "EC_P384_SHA384"},
			},
		}
	}

	subordinate := func() *GcpPrivateCaCertificateAuthority {
		r := root()
		r.Spec.Type = "SUBORDINATE"
		r.Spec.SubordinateConfig = &GcpPrivateCaCertificateAuthoritySubordinateConfig{
			CertificateAuthority: litRef("projects/p/locations/us-central1/caPools/root-pool/certificateAuthorities/root-ca"),
		}
		r.Spec.Config.X509Config.CaOptions.MaxIssuerPathLength = proto.Int32(0)
		return r
	}

	ginkgo.It("should accept a self-signed root", func() {
		gomega.Expect(validator.Validate(root())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a root with every lever", func() {
		r := root()
		r.Spec.ProjectId = litRef("p")
		r.Spec.CertificateAuthorityId = "root-ca-2026"
		r.Spec.Type = "SELF_SIGNED"
		r.Spec.Lifetime = "630720000s"
		r.Spec.GcsBucket = litRef("ca-publish")
		r.Spec.UserDefinedAccessUrls = &GcpPrivateCaCertificateAuthorityUserDefinedAccessUrls{CrlAccessUrls: []string{"http://crl.example.com/root.crl"}}
		r.Spec.DesiredState = "STAGED"
		r.Spec.DeletionProtection = proto.Bool(false)
		r.Spec.SkipGracePeriod = true
		r.Spec.IgnoreActiveCertificatesOnDeletion = true
		r.Spec.Labels = map[string]string{"tier": "root"}
		r.Spec.DeletionPolicy = "ABANDON"
		r.Spec.Config.SubjectKeyId = "4cf35b0e"
		r.Spec.Config.SubjectConfig.SubjectAltName = &GcpPrivateCaCertificateAuthoritySubjectAltName{Uris: []string{"spiffe://example.org"}}
		r.Spec.KeySpec = &GcpPrivateCaCertificateAuthorityKeySpec{CloudKmsKeyVersion: litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k/cryptoKeyVersions/1")}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a subordinate signed by reference or by an outside CA", func() {
		gomega.Expect(validator.Validate(subordinate())).To(gomega.Succeed())
		r := subordinate()
		r.Spec.SubordinateConfig = &GcpPrivateCaCertificateAuthoritySubordinateConfig{PemIssuerChain: []string{"-----BEGIN CERTIFICATE-----"}}
		r.Spec.PemCaCertificate = "-----BEGIN CERTIFICATE-----"
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		r = subordinate()
		r.Spec.SubordinateConfig = nil
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "awaiting activation")
	})

	ginkgo.It("should require is_ca on the authority's own certificate", func() {
		r := root()
		r.Spec.Config.X509Config.CaOptions = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "no ca_options")
		r = root()
		r.Spec.Config.X509Config.CaOptions = &GcpPrivateCaCertificateAuthorityCaOptions{MaxIssuerPathLength: proto.Int32(1)}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "no is_ca")
	})

	ginkgo.It("should require the pool, the config, the subject, and the key", func() {
		r := root()
		r.Spec.Pool = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "pool")
		r = root()
		r.Spec.Config.SubjectConfig.Subject.CommonName = ""
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "common name")
		r = root()
		r.Spec.Config.SubjectConfig = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "subject config")
		r = root()
		r.Spec.KeySpec = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "key spec")
	})

	ginkgo.It("should reject an empty subject alternative name block", func() {
		r := root()
		r.Spec.Config.SubjectConfig.SubjectAltName = &GcpPrivateCaCertificateAuthoritySubjectAltName{}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should take exactly one signing key", func() {
		r := root()
		r.Spec.KeySpec = &GcpPrivateCaCertificateAuthorityKeySpec{}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "neither")
		r.Spec.KeySpec = &GcpPrivateCaCertificateAuthorityKeySpec{Algorithm: "EC_P256_SHA256", CloudKmsKeyVersion: litRef("v")}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "both")
		r.Spec.KeySpec = &GcpPrivateCaCertificateAuthorityKeySpec{Algorithm: "EC_P521_SHA512"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "unknown algorithm")
	})

	ginkgo.It("should take exactly one issuer for a subordinate", func() {
		r := subordinate()
		r.Spec.SubordinateConfig.PemIssuerChain = []string{"pem"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "both")
		r.Spec.SubordinateConfig = &GcpPrivateCaCertificateAuthoritySubordinateConfig{}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "neither")
	})

	ginkgo.It("should keep activation fields to subordinates", func() {
		r := subordinate()
		r.Spec.Type = ""
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "subordinate_config on a root")
		r = root()
		r.Spec.Type = "SELF_SIGNED"
		r.Spec.PemCaCertificate = "pem"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "pem_ca_certificate on a root")
	})

	ginkgo.It("should need the issuer chain for an outside CA's certificate", func() {
		r := subordinate()
		r.Spec.PemCaCertificate = "pem"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "signed by reference")
		r.Spec.SubordinateConfig = nil
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "no subordinate_config")
	})

	ginkgo.It("should reject unknown type, state, lifetime, id, or deletion policy", func() {
		mutations := map[string]func(*GcpPrivateCaCertificateAuthoritySpec){
			"type":            func(s *GcpPrivateCaCertificateAuthoritySpec) { s.Type = "ROOT" },
			"desired_state":   func(s *GcpPrivateCaCertificateAuthoritySpec) { s.DesiredState = "PAUSED" },
			"lifetime":        func(s *GcpPrivateCaCertificateAuthoritySpec) { s.Lifetime = "10y" },
			"id":              func(s *GcpPrivateCaCertificateAuthoritySpec) { s.CertificateAuthorityId = "root/ca" },
			"deletion_policy": func(s *GcpPrivateCaCertificateAuthoritySpec) { s.DeletionPolicy = "FORCE" },
		}
		for name, mutate := range mutations {
			r := root()
			mutate(r.Spec)
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), name)
		}
	})
})

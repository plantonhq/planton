package gcpprivatecacertificatetemplatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpPrivateCaCertificateTemplateSpec Suite")
}

var _ = ginkgo.Describe("GcpPrivateCaCertificateTemplateSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	base := func() *GcpPrivateCaCertificateTemplate {
		return &GcpPrivateCaCertificateTemplate{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpPrivateCaCertificateTemplate",
			Metadata:   &shared.CloudResourceMetadata{Name: "tls-server"},
			Spec:       &GcpPrivateCaCertificateTemplateSpec{Location: "us-central1"},
		}
	}

	ginkgo.It("should accept a minimal template", func() {
		gomega.Expect(validator.Validate(base())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a TLS server template with every lever", func() {
		r := base()
		r.Spec.TemplateId = "tls_server"
		r.Spec.Description = "Internal TLS server leaf"
		r.Spec.MaximumLifetime = "2592000s"
		r.Spec.Labels = map[string]string{"use": "tls"}
		r.Spec.DeletionPolicy = "DELETE"
		r.Spec.PredefinedValues = &GcpPrivateCaCertificateTemplateX509Parameters{
			KeyUsage: &GcpPrivateCaCertificateTemplateKeyUsage{
				BaseKeyUsage:     &GcpPrivateCaCertificateTemplateBaseKeyUsage{DigitalSignature: true, KeyEncipherment: true},
				ExtendedKeyUsage: &GcpPrivateCaCertificateTemplateExtendedKeyUsage{ServerAuth: true},
			},
			CaOptions: &GcpPrivateCaCertificateTemplateCaOptions{IsCa: proto.Bool(false)},
		}
		r.Spec.IdentityConstraints = &GcpPrivateCaCertificateTemplateIdentityConstraints{
			AllowSubjectAltNamesPassthrough: true,
			CelExpression:                   &GcpPrivateCaCertificateTemplateCelExpression{Title: "no expression yet"},
		}
		r.Spec.PassthroughExtensions = &GcpPrivateCaCertificateTemplatePassthroughExtensions{
			KnownExtensions:      []string{"BASE_KEY_USAGE", "EXTENDED_KEY_USAGE"},
			AdditionalExtensions: []*GcpPrivateCaCertificateTemplateObjectId{{ObjectIdPath: []int32{1, 2, 3}}},
		}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown passthrough extension", func() {
		r := base()
		r.Spec.PassthroughExtensions = &GcpPrivateCaCertificateTemplatePassthroughExtensions{KnownExtensions: []string{"SUBJECT_ALT_NAME"}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed id, lifetime, location, or deletion policy", func() {
		mutations := map[string]func(*GcpPrivateCaCertificateTemplateSpec){
			"id":              func(s *GcpPrivateCaCertificateTemplateSpec) { s.TemplateId = "tls server" },
			"lifetime":        func(s *GcpPrivateCaCertificateTemplateSpec) { s.MaximumLifetime = "30d" },
			"location":        func(s *GcpPrivateCaCertificateTemplateSpec) { s.Location = "global" },
			"deletion_policy": func(s *GcpPrivateCaCertificateTemplateSpec) { s.DeletionPolicy = "KEEP" },
		}
		for name, mutate := range mutations {
			r := base()
			mutate(r.Spec)
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), name)
		}
	})

	ginkgo.It("should reject malformed predefined values", func() {
		r := base()
		r.Spec.PredefinedValues = &GcpPrivateCaCertificateTemplateX509Parameters{PolicyIds: []*GcpPrivateCaCertificateTemplateObjectId{{}}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})
})

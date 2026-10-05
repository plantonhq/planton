package gcpcloudrundomainmappingv1alpha1

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
	ginkgo.RunSpecs(t, "GcpCloudRunDomainMappingSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpCloudRunDomainMappingSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCloudRunDomainMapping {
		return &GcpCloudRunDomainMapping{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudRunDomainMapping",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-domain-mapping",
			},
			Spec: &GcpCloudRunDomainMappingSpec{
				Region: "us-central1",
				Domain: "app.example.com",
				Route:  litRef("my-service"),
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal domain mapping", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a fully configured domain mapping", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		target.Spec.CertificateMode = "AUTOMATIC"
		target.Spec.ForceOverride = true
		target.Spec.Namespace = "1234567890"
		target.Spec.Labels = map[string]string{"team": "payments"}
		target.Spec.Annotations = map[string]string{"note": "primary domain"}
		target.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each certificate_mode value", func() {
		for _, v := range []string{"AUTOMATIC", "NONE"} {
			target := minimal()
			target.Spec.CertificateMode = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept a root domain and a deep subdomain", func() {
		for _, d := range []string{"example.com", "api.eu.app.example.co.uk"} {
			target := minimal()
			target.Spec.Domain = d
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing region", func() {
		target := minimal()
		target.Spec.Region = ""
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed region", func() {
		target := minimal()
		target.Spec.Region = "US-Central-1"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject the global region (mappings are regional)", func() {
		target := minimal()
		target.Spec.Region = "global"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a missing domain", func() {
		target := minimal()
		target.Spec.Domain = ""
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a single-label domain (no dot)", func() {
		target := minimal()
		target.Spec.Domain = "localhost"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an uppercase domain", func() {
		target := minimal()
		target.Spec.Domain = "App.Example.com"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a missing route", func() {
		target := minimal()
		target.Spec.Route = nil
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown certificate_mode", func() {
		target := minimal()
		target.Spec.CertificateMode = "MANAGED"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})
})

package gcpdocumentaiprocessorv1alpha1

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
	ginkgo.RunSpecs(t, "GcpDocumentAiProcessorSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpDocumentAiProcessorSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpDocumentAiProcessor {
		return &GcpDocumentAiProcessor{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDocumentAiProcessor",
			Metadata:   &shared.CloudResourceMetadata{Name: "invoice-ocr"},
			Spec: &GcpDocumentAiProcessorSpec{
				Location: "us",
				Type:     "OCR_PROCESSOR",
			},
		}
	}

	ginkgo.It("should accept the smallest processor", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("docs-project")
		msg.Spec.Location = "eu"
		msg.Spec.Type = "INVOICE_PROCESSOR"
		msg.Spec.DisplayName = "Invoices"
		msg.Spec.KmsKeyName = litRef("projects/p/locations/eu/keyRings/r/cryptoKeys/k")
		msg.Spec.DefaultVersion = "pretrained-invoice-v2.0-2023-12-06"
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a location and an upper-case type", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Type = "ocr_processor"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse the stable and rc channel aliases as a default version", func() {
		for _, alias := range []string{"stable", "rc"} {
			msg := minimal()
			msg.Spec.DefaultVersion = alias
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		}
	})

	ginkgo.It("should refuse a full path as a default version", func() {
		msg := minimal()
		msg.Spec.DefaultVersion = "projects/p/locations/us/processors/1/processorVersions/v1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})

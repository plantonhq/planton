package gcpvertexairagengineconfigv1alpha1

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
	ginkgo.RunSpecs(t, "GcpVertexAiRagEngineConfigSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpVertexAiRagEngineConfigSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpVertexAiRagEngineConfig {
		return &GcpVertexAiRagEngineConfig{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiRagEngineConfig",
			Metadata:   &shared.CloudResourceMetadata{Name: "rag-engine-us-central1"},
			Spec: &GcpVertexAiRagEngineConfigSpec{
				Location: "us-central1",
				Tier:     "BASIC",
			},
		}
	}

	ginkgo.It("should accept the basic tier", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every tier with an explicit project and deletion policy", func() {
		for _, tier := range []string{"BASIC", "SCALED", "UNPROVISIONED"} {
			msg := minimal()
			msg.Spec.ProjectId = litRef("ai-project")
			msg.Spec.Tier = tier
			msg.Spec.DeletionPolicy = "ABANDON"
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should require location and tier", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Location = "us"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Tier = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a tier Google does not offer", func() {
		msg := minimal()
		msg.Spec.Tier = "ENTERPRISE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})

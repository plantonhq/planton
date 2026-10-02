package gcpvertexaidatasetv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiDatasetSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpVertexAiDatasetSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpVertexAiDataset {
		return &GcpVertexAiDataset{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiDataset",
			Metadata:   &shared.CloudResourceMetadata{Name: "product-images"},
			Spec: &GcpVertexAiDatasetSpec{
				Location:          "us-central1",
				MetadataSchemaUri: "gs://google-cloud-aiplatform/schema/dataset/metadata/image_1.0.0.yaml",
			},
		}
	}

	ginkgo.It("should accept the smallest dataset", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.DisplayName = "Product images"
		msg.Spec.Labels = map[string]string{"team": "vision"}
		msg.Spec.KmsKeyName = litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k")
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a regional location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Location = "global"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a Cloud Storage YAML schema", func() {
		msg := minimal()
		msg.Spec.MetadataSchemaUri = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.MetadataSchemaUri = "https://example.com/image_1.0.0.yaml"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a display name over 128 characters", func() {
		msg := minimal()
		msg.Spec.DisplayName = strings.Repeat("d", 129)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})

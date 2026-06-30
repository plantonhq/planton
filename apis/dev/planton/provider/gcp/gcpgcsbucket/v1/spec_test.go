package gcpgcsbucketv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestGcpGcsBucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGcsBucketSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpGcsBucketSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_gcs_bucket", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &GcpGcsBucket{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpGcsBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gcs-bucket",
					},
					Spec: &GcpGcsBucketSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Location:                        "us-central1",
						BucketName:                      "test-bucket-123",
						UniformBucketLevelAccessEnabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference", func() {
				input := &GcpGcsBucket{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpGcsBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gcs-bucket-ref",
					},
					Spec: &GcpGcsBucketSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name:      "main-project",
									FieldPath: "status.outputs.project_id",
								},
							},
						},
						Location:                        "us-east1",
						BucketName:                      "test-bucket-with-ref",
						UniformBucketLevelAccessEnabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("gcp_gcs_bucket", func() {

			ginkgo.It("should return a validation error when gcp_project_id is missing", func() {
				input := &GcpGcsBucket{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpGcsBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gcs-bucket",
					},
					Spec: &GcpGcsBucketSpec{
						Location:                        "us-central1",
						BucketName:                      "test-bucket-123",
						UniformBucketLevelAccessEnabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when bucket_name has invalid format", func() {
				input := &GcpGcsBucket{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpGcsBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gcs-bucket",
					},
					Spec: &GcpGcsBucketSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Location:                        "us-central1",
						BucketName:                      "INVALID-BUCKET-NAME", // uppercase not allowed
						UniformBucketLevelAccessEnabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

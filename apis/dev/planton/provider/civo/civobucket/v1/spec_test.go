package civobucketv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestCivoBucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoBucketSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoBucketSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("civo_bucket with minimal valid fields", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &CivoBucketSpec{
						BucketName: "test-bucket",
						Region:     civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with versioning enabled", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket-versioned",
					},
					Spec: &CivoBucketSpec{
						BucketName:        "test-bucket-versioned",
						Region:            civo.CivoRegion_fra1,
						VersioningEnabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with tags", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket-tagged",
					},
					Spec: &CivoBucketSpec{
						BucketName: "test-bucket-tagged",
						Region:     civo.CivoRegion_nyc1,
						Tags:       []string{"env:prod", "team:backend", "criticality:high"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all fields", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-backups",
					},
					Spec: &CivoBucketSpec{
						BucketName:        "prod-backups",
						Region:            civo.CivoRegion_lon1,
						VersioningEnabled: true,
						Tags:              []string{"env:prod", "retention:180-days"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("bucket_name validation", func() {

			ginkgo.It("should return a validation error when bucket_name is empty", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &CivoBucketSpec{
						BucketName: "",
						Region:     civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when bucket_name is too short", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ab",
					},
					Spec: &CivoBucketSpec{
						BucketName: "ab", // less than 3 chars
						Region:     civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when bucket_name is too long", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &CivoBucketSpec{
						BucketName: "this-is-a-very-long-bucket-name-that-exceeds-the-maximum-allowed-length-of-sixty-three-characters", // more than 63 chars
						Region:     civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when bucket_name contains uppercase", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &CivoBucketSpec{
						BucketName: "Test-Bucket", // uppercase not allowed
						Region:     civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when bucket_name contains underscores", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &CivoBucketSpec{
						BucketName: "test_bucket", // underscores not allowed
						Region:     civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when bucket_name starts with hyphen", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &CivoBucketSpec{
						BucketName: "-test-bucket", // can't start with hyphen
						Region:     civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when bucket_name ends with hyphen", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &CivoBucketSpec{
						BucketName: "test-bucket-", // can't end with hyphen
						Region:     civo.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("region validation", func() {

			ginkgo.It("should return a validation error when region is not set", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &CivoBucketSpec{
						BucketName: "test-bucket",
						// Region not set
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("tags validation", func() {

			ginkgo.It("should return a validation error when tags are not unique", func() {
				input := &CivoBucket{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &CivoBucketSpec{
						BucketName: "test-bucket",
						Region:     civo.CivoRegion_lon1,
						Tags:       []string{"env:prod", "env:prod"}, // duplicate tags
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

package cloudflarer2bucketv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestCloudflareR2BucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareR2BucketSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareR2BucketSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("cloudflare_r2_bucket", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-r2-bucket",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "test-bucket",
						AccountId:  "00000000000000000000000000000000",
						Location:   CloudflareR2Location_WEUR,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with public access enabled", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-r2-bucket-public",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName:   "test-public-bucket",
						AccountId:    "00000000000000000000000000000000",
						Location:     CloudflareR2Location_WEUR,
						PublicAccess: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

		ginkgo.It("should not return a validation error with custom domain enabled and valid config", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-r2-bucket-custom-domain",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "test-custom-domain-bucket",
						AccountId:  "00000000000000000000000000000000",
						Location:   CloudflareR2Location_WEUR,
						CustomDomain: &CloudflareR2BucketCustomDomainConfig{
							Enabled: true,
							ZoneId: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "00000000000000000000000000000000",
								},
							},
							Domain: "media.example.com",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with custom domain disabled and no zone_id/domain", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-r2-bucket-custom-domain-disabled",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "test-disabled-custom-domain",
						AccountId:  "00000000000000000000000000000000",
						Location:   CloudflareR2Location_WEUR,
						CustomDomain: &CloudflareR2BucketCustomDomainConfig{
							Enabled: false,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("account_id validation", func() {

			ginkgo.It("should return error if account_id is missing", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-no-account",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "test-bucket",
						Location:   CloudflareR2Location_WEUR,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if account_id is not 32 characters", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-short-account",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "test-bucket",
						AccountId:  "123",
						Location:   CloudflareR2Location_WEUR,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if account_id contains non-hex characters", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-invalid-hex",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "test-bucket",
						AccountId:  "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ",
						Location:   CloudflareR2Location_WEUR,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("bucket_name validation", func() {

			ginkgo.It("should return error if bucket_name is missing", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-no-bucket-name",
					},
					Spec: &CloudflareR2BucketSpec{
						AccountId: "00000000000000000000000000000000",
						Location:  CloudflareR2Location_WEUR,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if bucket_name is too short", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-short-bucket",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "ab",
						AccountId:  "00000000000000000000000000000000",
						Location:   CloudflareR2Location_WEUR,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if bucket_name contains invalid characters", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-invalid-bucket",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "Test_Bucket",
						AccountId:  "00000000000000000000000000000000",
						Location:   CloudflareR2Location_WEUR,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("custom_domain validation", func() {

			ginkgo.It("should return error if custom domain enabled but zone_id is missing", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-custom-domain-no-zone",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "test-bucket",
						AccountId:  "00000000000000000000000000000000",
						Location:   CloudflareR2Location_WEUR,
						CustomDomain: &CloudflareR2BucketCustomDomainConfig{
							Enabled: true,
							Domain:  "media.example.com",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if custom domain enabled but domain is missing", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-custom-domain-no-domain",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "test-bucket",
						AccountId:  "00000000000000000000000000000000",
						Location:   CloudflareR2Location_WEUR,
						CustomDomain: &CloudflareR2BucketCustomDomainConfig{
							Enabled: true,
							ZoneId: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "00000000000000000000000000000000",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if custom domain enabled but domain is empty string", func() {
				input := &CloudflareR2Bucket{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareR2Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-custom-domain-empty-domain",
					},
					Spec: &CloudflareR2BucketSpec{
						BucketName: "test-bucket",
						AccountId:  "00000000000000000000000000000000",
						Location:   CloudflareR2Location_WEUR,
						CustomDomain: &CloudflareR2BucketCustomDomainConfig{
							Enabled: true,
							ZoneId: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "00000000000000000000000000000000",
								},
							},
							Domain: "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

package awss3objectsetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func stringPtr(s string) *string {
	return &s
}

func TestAwsS3ObjectSetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsS3ObjectSetSpec Validation Tests")
}

// minimalValidSpec returns a minimal valid AwsS3ObjectSetSpec.
func minimalValidSpec() *AwsS3ObjectSetSpec {
	return &AwsS3ObjectSetSpec{
		Bucket: &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "my-test-bucket",
			},
		},
		AwsRegion: "us-east-1",
		Objects: []*AwsS3Object{
			{
				Key: "config/app.json",
				Source: &AwsS3Object_Content{
					Content: "{\"key\": \"value\"}",
				},
				ContentType: stringPtr("application/json"),
			},
		},
	}
}

var _ = ginkgo.Describe("AwsS3ObjectSetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_s3_object_set_spec", func() {

			ginkgo.It("should not return a validation error for minimal valid spec", func() {
				spec := minimalValidSpec()
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple objects", func() {
				spec := minimalValidSpec()
				spec.Objects = append(spec.Objects, &AwsS3Object{
					Key: "assets/logo.png",
					Source: &AwsS3Object_ContentBase64{
						ContentBase64: "iVBORw0KGgoAAAANSUhEUg==",
					},
					ContentType: stringPtr("image/png"),
				})
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with set-level tags", func() {
				spec := minimalValidSpec()
				spec.Tags = map[string]string{
					"environment": "production",
					"team":        "platform",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with object-level tags", func() {
				spec := minimalValidSpec()
				spec.Objects[0].Tags = map[string]string{
					"purpose": "config",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with cache_control and content_encoding", func() {
				spec := minimalValidSpec()
				spec.Objects[0].CacheControl = "max-age=86400"
				spec.Objects[0].ContentEncoding = "gzip"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with acl set", func() {
				spec := minimalValidSpec()
				spec.Objects[0].Acl = "private"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with bucket as value_from reference", func() {
				spec := minimalValidSpec()
				spec.Bucket = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-s3-bucket",
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_s3_object_set_spec", func() {

			ginkgo.It("should return a validation error when bucket is missing", func() {
				spec := minimalValidSpec()
				spec.Bucket = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when aws_region is empty", func() {
				spec := minimalValidSpec()
				spec.AwsRegion = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when objects list is empty", func() {
				spec := minimalValidSpec()
				spec.Objects = []*AwsS3Object{}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when objects is nil", func() {
				spec := minimalValidSpec()
				spec.Objects = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when object key is empty", func() {
				spec := minimalValidSpec()
				spec.Objects[0].Key = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when object has no content source", func() {
				spec := minimalValidSpec()
				spec.Objects[0].Source = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

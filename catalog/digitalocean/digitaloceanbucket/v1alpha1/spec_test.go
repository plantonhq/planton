package digitaloceanbucketv1alpha1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestDigitalOceanBucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanBucketSpec Custom Validation Tests")
}

// minimalBucket returns a valid bucket resource the cases below mutate.
func minimalBucket() *DigitalOceanBucket {
	return &DigitalOceanBucket{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanBucket",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-bucket",
		},
		Spec: &DigitalOceanBucketSpec{
			BucketName: "test-bucket",
		},
	}
}

var _ = ginkgo.Describe("DigitalOceanBucketSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("digitalocean_spaces_bucket", func() {

			ginkgo.It("should not return a validation error for the minimal bucket (no region: provider default applies)", func() {
				err := protovalidate.Validate(minimalBucket())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with an explicit Spaces region and versioning", func() {
				input := minimalBucket()
				input.Spec.Region = digitalocean.DigitalOceanRegion_nyc3
				input.Spec.VersioningEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with force_destroy and public-read access", func() {
				input := minimalBucket()
				input.Spec.AccessControl = DigitalOceanBucketAccessControl_PUBLIC_READ
				input.Spec.ForceDestroy = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a days-based lifecycle rule", func() {
				input := minimalBucket()
				enabled := true
				input.Spec.LifecycleRules = []*DigitalOceanBucketLifecycleRule{
					{
						Id:      "expire-logs",
						Prefix:  "logs/",
						Enabled: &enabled,
						Expiration: &DigitalOceanBucketLifecycleExpiration{
							Days: 30,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a date-based rule with noncurrent expiration and multipart abort", func() {
				input := minimalBucket()
				enabled := true
				input.Spec.LifecycleRules = []*DigitalOceanBucketLifecycleRule{
					{
						Enabled:                            &enabled,
						AbortIncompleteMultipartUploadDays: 7,
						Expiration: &DigitalOceanBucketLifecycleExpiration{
							Date: "2027-01-01",
						},
						NoncurrentVersionExpiration: &DigitalOceanBucketLifecycleNoncurrentVersionExpiration{
							Days: 90,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a staged-but-disabled lifecycle rule", func() {
				input := minimalBucket()
				disabled := false
				input.Spec.LifecycleRules = []*DigitalOceanBucketLifecycleRule{
					{
						Enabled: &disabled,
						Expiration: &DigitalOceanBucketLifecycleExpiration{
							ExpiredObjectDeleteMarker: true,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CORS rules with an explicit region", func() {
				input := minimalBucket()
				input.Spec.Region = digitalocean.DigitalOceanRegion_fra1
				input.Spec.CorsRules = []*DigitalOceanBucketCorsRule{
					{
						AllowedMethods: []string{"GET", "HEAD"},
						AllowedOrigins: []string{"https://example.com"},
						AllowedHeaders: []string{"*"},
						ExposeHeaders:  []string{"ETag"},
						Id:             "web-read",
						MaxAgeSeconds:  3000,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a policy with an explicit region", func() {
				input := minimalBucket()
				input.Spec.Region = digitalocean.DigitalOceanRegion_ams3
				input.Spec.Policy = `{"Version":"2012-10-17","Statement":[]}`
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for logging with an explicit region", func() {
				input := minimalBucket()
				input.Spec.Region = digitalocean.DigitalOceanRegion_sgp1
				input.Spec.Logging = &DigitalOceanBucketLogging{
					TargetBucket: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "log-sink-bucket"},
					},
					TargetPrefix: "logs/",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("bucket identity", func() {

			ginkgo.It("should return a validation error when bucket_name is missing", func() {
				input := minimalBucket()
				input.Spec.BucketName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an uppercase bucket name", func() {
				input := minimalBucket()
				input.Spec.BucketName = "Test-Bucket"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a too-short bucket name", func() {
				input := minimalBucket()
				input.Spec.BucketName = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a region without Spaces (nyc1)", func() {
				input := minimalBucket()
				input.Spec.Region = digitalocean.DigitalOceanRegion_nyc1
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("satellites require an explicit region", func() {

			ginkgo.It("should return a validation error for CORS rules without a region", func() {
				input := minimalBucket()
				input.Spec.CorsRules = []*DigitalOceanBucketCorsRule{
					{
						AllowedMethods: []string{"GET"},
						AllowedOrigins: []string{"*"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a policy without a region", func() {
				input := minimalBucket()
				input.Spec.Policy = `{"Version":"2012-10-17","Statement":[]}`
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for logging without a region", func() {
				input := minimalBucket()
				input.Spec.Logging = &DigitalOceanBucketLogging{
					TargetBucket: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "log-sink-bucket"},
					},
					TargetPrefix: "logs/",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("lifecycle rules", func() {

			ginkgo.It("should return a validation error when a rule omits enabled", func() {
				input := minimalBucket()
				input.Spec.LifecycleRules = []*DigitalOceanBucketLifecycleRule{
					{
						Expiration: &DigitalOceanBucketLifecycleExpiration{Days: 30},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an expiration with no trigger", func() {
				input := minimalBucket()
				enabled := true
				input.Spec.LifecycleRules = []*DigitalOceanBucketLifecycleRule{
					{
						Enabled:    &enabled,
						Expiration: &DigitalOceanBucketLifecycleExpiration{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an expiration with two triggers", func() {
				input := minimalBucket()
				enabled := true
				input.Spec.LifecycleRules = []*DigitalOceanBucketLifecycleRule{
					{
						Enabled: &enabled,
						Expiration: &DigitalOceanBucketLifecycleExpiration{
							Date: "2027-01-01",
							Days: 30,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed expiration date", func() {
				input := minimalBucket()
				enabled := true
				input.Spec.LifecycleRules = []*DigitalOceanBucketLifecycleRule{
					{
						Enabled: &enabled,
						Expiration: &DigitalOceanBucketLifecycleExpiration{
							Date: "01/01/2027",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a zero-day noncurrent expiration", func() {
				input := minimalBucket()
				enabled := true
				input.Spec.LifecycleRules = []*DigitalOceanBucketLifecycleRule{
					{
						Enabled:                     &enabled,
						NoncurrentVersionExpiration: &DigitalOceanBucketLifecycleNoncurrentVersionExpiration{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("CORS and logging shapes", func() {

			ginkgo.It("should return a validation error for a CORS rule without methods", func() {
				input := minimalBucket()
				input.Spec.Region = digitalocean.DigitalOceanRegion_nyc3
				input.Spec.CorsRules = []*DigitalOceanBucketCorsRule{
					{
						AllowedOrigins: []string{"*"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a CORS rule without origins", func() {
				input := minimalBucket()
				input.Spec.Region = digitalocean.DigitalOceanRegion_nyc3
				input.Spec.CorsRules = []*DigitalOceanBucketCorsRule{
					{
						AllowedMethods: []string{"GET"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for logging without target_bucket", func() {
				input := minimalBucket()
				input.Spec.Region = digitalocean.DigitalOceanRegion_nyc3
				input.Spec.Logging = &DigitalOceanBucketLogging{
					TargetPrefix: "logs/",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for logging without target_prefix", func() {
				input := minimalBucket()
				input.Spec.Region = digitalocean.DigitalOceanRegion_nyc3
				input.Spec.Logging = &DigitalOceanBucketLogging{
					TargetBucket: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "log-sink-bucket"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

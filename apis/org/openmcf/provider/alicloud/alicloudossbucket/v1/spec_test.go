package alicloudossbucketv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAlicloudOssBucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudOssBucketSpec Validation Tests")
}

var _ = ginkgo.Describe("AlicloudOssBucketSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-bucket",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-test-bucket",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config-bucket",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:            "cn-shanghai",
					BucketName:        "prod-assets-bucket",
					Acl:               strPtr("public-read"),
					StorageClass:      strPtr("IA"),
					RedundancyType:    strPtr("ZRS"),
					VersioningEnabled: true,
					ServerSideEncryption: &AlicloudOssBucketEncryption{
						SseAlgorithm:   "KMS",
						KmsMasterKeyId: "kms-key-123",
					},
					LifecycleRules: []*AlicloudOssBucketLifecycleRule{
						{
							Prefix:         "logs/",
							Enabled:        true,
							ExpirationDays: 90,
							Transitions: []*AlicloudOssBucketLifecycleTransition{
								{Days: 30, StorageClass: "IA"},
								{Days: 60, StorageClass: "Archive"},
							},
							AbortMultipartUploadDays:       7,
							NoncurrentVersionExpirationDays: 30,
						},
					},
					CorsRules: []*AlicloudOssBucketCorsRule{
						{
							AllowedOrigins: []string{"https://example.com"},
							AllowedMethods: []string{"GET", "PUT"},
							AllowedHeaders: []string{"*"},
							ExposeHeaders:  []string{"ETag"},
							MaxAgeSeconds:  3600,
						},
					},
					Logging: &AlicloudOssBucketLogging{
						TargetBucket: "log-bucket",
						TargetPrefix: "access-logs/",
					},
					ForceDestroy:    true,
					ResourceGroupId: "rg-abc123",
					Tags:            map[string]string{"team": "platform", "cost-center": "eng"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with AES256 encryption", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "encrypted-bucket",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "aes-encrypted-bucket",
					ServerSideEncryption: &AlicloudOssBucketEncryption{
						SseAlgorithm: "AES256",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with lifecycle rules and versioning", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "lifecycle-bucket",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:            "us-west-1",
					BucketName:        "lifecycle-versioned-bucket",
					VersioningEnabled: true,
					LifecycleRules: []*AlicloudOssBucketLifecycleRule{
						{
							Prefix:  "",
							Enabled: true,
							Transitions: []*AlicloudOssBucketLifecycleTransition{
								{Days: 30, StorageClass: "IA"},
								{Days: 90, StorageClass: "Archive"},
							},
							AbortMultipartUploadDays:       7,
							NoncurrentVersionExpirationDays: 60,
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					BucketName: "my-bucket",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when bucket_name is missing", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when bucket_name is too short", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "ab",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when bucket_name exceeds max length", func() {
			longName := ""
			for i := 0; i < 64; i++ {
				longName += "a"
			}
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: longName,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when acl has invalid value", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-bucket",
					Acl:        strPtr("invalid-acl"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when storage_class has invalid value", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:       "cn-hangzhou",
					BucketName:   "my-bucket",
					StorageClass: strPtr("InvalidClass"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when redundancy_type has invalid value", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:         "cn-hangzhou",
					BucketName:     "my-bucket",
					RedundancyType: strPtr("GRS"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when sse_algorithm has invalid value", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-bucket",
					ServerSideEncryption: &AlicloudOssBucketEncryption{
						SseAlgorithm: "DES",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when transition storage_class has invalid value", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-bucket",
					LifecycleRules: []*AlicloudOssBucketLifecycleRule{
						{
							Enabled: true,
							Transitions: []*AlicloudOssBucketLifecycleTransition{
								{Days: 30, StorageClass: "Standard"},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cors_rule has no allowed_origins", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-bucket",
					CorsRules: []*AlicloudOssBucketCorsRule{
						{
							AllowedMethods: []string{"GET"},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cors_rule has no allowed_methods", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-bucket",
					CorsRules: []*AlicloudOssBucketCorsRule{
						{
							AllowedOrigins: []string{"*"},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when logging target_bucket is missing", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-bucket",
					Logging: &AlicloudOssBucketLogging{
						TargetPrefix: "logs/",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudOssBucket",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-bucket",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-bucket",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudOssBucket{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudOssBucket",
				Spec: &AlicloudOssBucketSpec{
					Region:     "cn-hangzhou",
					BucketName: "my-bucket",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})

func strPtr(s string) *string {
	return &s
}

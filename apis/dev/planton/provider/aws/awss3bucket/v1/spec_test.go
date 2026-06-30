package awss3bucketv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsS3BucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsS3BucketSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsS3BucketSpec Custom Validation Tests", func() {

	ginkgo.Describe("Valid inputs", func() {
		ginkgo.Context("minimal configuration", func() {
			ginkgo.It("should accept minimal valid fields", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-s3-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("private bucket with versioning", func() {
			ginkgo.It("should accept private bucket with versioning enabled", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-versioned-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:            "us-west-2",
						IsPublic:          false,
						VersioningEnabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("public bucket", func() {
			ginkgo.It("should accept public bucket configuration", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-public-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:   "eu-west-1",
						IsPublic: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("SSE-S3 encryption", func() {
			ginkgo.It("should accept SSE-S3 encryption type", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-encrypted-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:         "us-east-1",
						EncryptionType: AwsS3BucketSpec_ENCRYPTION_TYPE_SSE_S3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("SSE-KMS encryption", func() {
			ginkgo.It("should accept SSE-KMS encryption with KMS key", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-kms-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:         "us-east-1",
						EncryptionType: AwsS3BucketSpec_ENCRYPTION_TYPE_SSE_KMS,
						KmsKeyId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("tags", func() {
			ginkgo.It("should accept bucket with tags", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-tagged-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
						Tags: map[string]string{
							"Environment": "production",
							"Project":     "myproject",
							"Owner":       "team@example.com",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("lifecycle rules", func() {
			ginkgo.It("should accept lifecycle rules for storage transitions", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-lifecycle-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
						LifecycleRules: []*AwsS3BucketSpec_LifecycleRule{
							{
								Id:                                 "transition-to-ia",
								Enabled:                            true,
								Prefix:                             "logs/",
								TransitionDays:                     30,
								TransitionStorageClass:             AwsS3BucketSpec_STORAGE_CLASS_STANDARD_IA,
								ExpirationDays:                     90,
								AbortIncompleteMultipartUploadDays: 7,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("replication", func() {
			ginkgo.It("should accept replication configuration with versioning", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-replicated-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:            "us-east-1",
						VersioningEnabled: true,
						Replication: &AwsS3BucketSpec_ReplicationConfiguration{
							Enabled: true,
							RoleArn: &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/replication-role"}},
							Destination: &AwsS3BucketSpec_ReplicationConfiguration_Destination{
								BucketArn: "arn:aws:s3:::destination-bucket",
							},
							Priority: 1,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("logging", func() {
			ginkgo.It("should accept logging configuration", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-logged-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
						Logging: &AwsS3BucketSpec_LoggingConfiguration{
							Enabled:      true,
							TargetBucket: "logging-bucket",
							TargetPrefix: "logs/mybucket/",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("CORS", func() {
			ginkgo.It("should accept CORS configuration", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cors-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
						Cors: &AwsS3BucketSpec_CorsConfiguration{
							CorsRules: []*AwsS3BucketSpec_CorsConfiguration_CorsRule{
								{
									AllowedMethods: []string{"GET", "PUT", "POST"},
									AllowedOrigins: []string{"https://example.com"},
									AllowedHeaders: []string{"*"},
									MaxAgeSeconds:  3600,
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Invalid inputs", func() {
		ginkgo.Context("aws_region validation", func() {
			ginkgo.It("should fail when aws_region is empty", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "", // Empty region should fail
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("encryption type validation", func() {
			ginkgo.It("should fail when encryption_type uses undefined enum value", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:         "us-east-1",
						EncryptionType: AwsS3BucketSpec_EncryptionType(999), // Invalid enum value
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("lifecycle rule validation", func() {
			ginkgo.It("should fail when lifecycle rule id is empty", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
						LifecycleRules: []*AwsS3BucketSpec_LifecycleRule{
							{
								Id:             "", // Empty ID should fail
								Enabled:        true,
								ExpirationDays: 30,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should fail when storage class is undefined", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
						LifecycleRules: []*AwsS3BucketSpec_LifecycleRule{
							{
								Id:                     "test-rule",
								Enabled:                true,
								TransitionDays:         30,
								TransitionStorageClass: AwsS3BucketSpec_StorageClass(999), // Invalid enum
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("replication validation", func() {
			ginkgo.It("should fail when replication role_arn is empty", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:            "us-east-1",
						VersioningEnabled: true,
						Replication: &AwsS3BucketSpec_ReplicationConfiguration{
							Enabled: true,
							RoleArn: nil, // nil role ARN should fail
							Destination: &AwsS3BucketSpec_ReplicationConfiguration_Destination{
								BucketArn: "arn:aws:s3:::destination-bucket",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should fail when replication destination is not provided", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:            "us-east-1",
						VersioningEnabled: true,
						Replication: &AwsS3BucketSpec_ReplicationConfiguration{
							Enabled:     true,
							RoleArn:     &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/replication-role"}},
							Destination: nil, // Missing destination should fail
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should fail when replication destination bucket_arn is empty", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:            "us-east-1",
						VersioningEnabled: true,
						Replication: &AwsS3BucketSpec_ReplicationConfiguration{
							Enabled: true,
							RoleArn: &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/replication-role"}},
							Destination: &AwsS3BucketSpec_ReplicationConfiguration_Destination{
								BucketArn: "", // Empty bucket ARN should fail
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should fail when replication destination storage class is undefined", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region:            "us-east-1",
						VersioningEnabled: true,
						Replication: &AwsS3BucketSpec_ReplicationConfiguration{
							Enabled: true,
							RoleArn: &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/replication-role"}},
							Destination: &AwsS3BucketSpec_ReplicationConfiguration_Destination{
								BucketArn:    "arn:aws:s3:::destination-bucket",
								StorageClass: AwsS3BucketSpec_StorageClass(999), // Invalid enum
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("logging validation", func() {
			ginkgo.It("should fail when logging target_bucket is empty", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
						Logging: &AwsS3BucketSpec_LoggingConfiguration{
							Enabled:      true,
							TargetBucket: "", // Empty target bucket should fail
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("CORS validation", func() {
			ginkgo.It("should fail when CORS allowed_methods is empty", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
						Cors: &AwsS3BucketSpec_CorsConfiguration{
							CorsRules: []*AwsS3BucketSpec_CorsConfiguration_CorsRule{
								{
									AllowedMethods: []string{}, // Empty methods should fail
									AllowedOrigins: []string{"https://example.com"},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should fail when CORS allowed_origins is empty", func() {
				input := &AwsS3Bucket{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsS3Bucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &AwsS3BucketSpec{
						Region: "us-east-1",
						Cors: &AwsS3BucketSpec_CorsConfiguration{
							CorsRules: []*AwsS3BucketSpec_CorsConfiguration_CorsRule{
								{
									AllowedMethods: []string{"GET"},
									AllowedOrigins: []string{}, // Empty origins should fail
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

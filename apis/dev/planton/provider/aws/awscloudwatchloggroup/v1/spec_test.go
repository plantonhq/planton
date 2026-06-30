package awscloudwatchloggroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	fkv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsCloudwatchLogGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCloudwatchLogGroupSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsCloudwatchLogGroupSpec validations", func() {

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal log group with empty spec (never-expire retention)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "my-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{Region: "us-west-2"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with 1-day retention", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "short-lived-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 1,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with 30-day retention", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "standard-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 30,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with 365-day retention", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "annual-retention-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 365,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts the maximum retention of 3653 days (~10 years)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "decade-retention-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 3653,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with KMS encryption", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "encrypted-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 90,
				KmsKeyId: &fkv1.StringValueOrRef{
					LiteralOrRef: &fkv1.StringValueOrRef_Value{
						Value: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with KMS encryption via valueFrom", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "encrypted-logs-ref",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 90,
				KmsKeyId: &fkv1.StringValueOrRef{
					LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
						ValueFrom: &fkv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AwsKmsKey,
							Name:      "log-encryption-key",
							FieldPath: "status.outputs.key_arn",
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with STANDARD class", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "standard-class-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 30,
				LogGroupClass:   "STANDARD",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with INFREQUENT_ACCESS class", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "ia-class-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 365,
				LogGroupClass:   "INFREQUENT_ACCESS",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with DELIVERY class (no retention)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "delivery-class-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:        "us-west-2",
				LogGroupClass: "DELIVERY",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with deletion protection enabled", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "protected-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:                    "us-west-2",
				RetentionInDays:           90,
				DeletionProtectionEnabled: true,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready log group with all features", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "prod-app-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 90,
				KmsKeyId: &fkv1.StringValueOrRef{
					LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
						ValueFrom: &fkv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AwsKmsKey,
							Name:      "log-key",
							FieldPath: "status.outputs.key_arn",
						},
					},
				},
				LogGroupClass:             "STANDARD",
				DeletionProtectionEnabled: true,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: retention_in_days_valid_values
	// -------------------------------------------------------------------------

	ginkgo.It("fails when retention_in_days is 2 (not a valid AWS value)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-retention-2",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				RetentionInDays: 2,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when retention_in_days is 10 (not a valid AWS value)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-retention-10",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				RetentionInDays: 10,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when retention_in_days is 45 (not a valid AWS value)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-retention-45",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				RetentionInDays: 45,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when retention_in_days is 100 (not a valid AWS value)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-retention-100",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				RetentionInDays: 100,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when retention_in_days is negative", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-retention-negative",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				RetentionInDays: -1,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: log_group_class_valid_values
	// -------------------------------------------------------------------------

	ginkgo.It("fails when log_group_class is an invalid value", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-class-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				LogGroupClass: "PREMIUM",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when log_group_class is lowercase (must be uppercase)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-class-lowercase",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				LogGroupClass: "standard",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: delivery_class_no_retention
	// -------------------------------------------------------------------------

	ginkgo.It("fails when DELIVERY class has retention set", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-delivery-retention",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				LogGroupClass:   "DELIVERY",
				RetentionInDays: 30,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// api.proto: api_version and kind constants
	// -------------------------------------------------------------------------

	ginkgo.It("fails when api_version is wrong", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "wrong.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{Region: "us-west-2"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "WrongKind",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{Region: "us-west-2"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Spec:       &AwsCloudwatchLogGroupSpec{Region: "us-west-2"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-logs",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})

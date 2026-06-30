package awskinesisstreamconsumerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsKinesisStreamConsumerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsKinesisStreamConsumerSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsKinesisStreamConsumerSpec validations", func() {

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal consumer with stream_arn as a literal value", func() {
		input := &AwsKinesisStreamConsumer{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsKinesisStreamConsumer",
			Metadata: &shared.CloudResourceMetadata{
				Name: "my-consumer",
			},
			Spec: &AwsKinesisStreamConsumerSpec{
				Region: "us-west-2",
				StreamArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "arn:aws:kinesis:us-east-1:123456789012:stream/my-stream",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a consumer with stream_arn using valueFrom reference", func() {
		input := &AwsKinesisStreamConsumer{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsKinesisStreamConsumer",
			Metadata: &shared.CloudResourceMetadata{
				Name: "analytics-consumer",
			},
			Spec: &AwsKinesisStreamConsumerSpec{
				Region: "us-west-2",
				StreamArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AwsKinesisStream,
							Name:      "my-stream",
							FieldPath: "status.outputs.stream_arn",
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a consumer with metadata labels and org/env/id", func() {
		input := &AwsKinesisStreamConsumer{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsKinesisStreamConsumer",
			Metadata: &shared.CloudResourceMetadata{
				Name: "audit-consumer",
				Org:  "acme",
				Env:  "production",
				Id:   "audit-consumer-prod",
				Labels: map[string]string{
					"team":        "data-engineering",
					"cost-center": "analytics",
				},
			},
			Spec: &AwsKinesisStreamConsumerSpec{
				Region: "us-west-2",
				StreamArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "arn:aws:kinesis:us-west-2:123456789012:stream/audit-stream",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: stream_arn is required
	// -------------------------------------------------------------------------

	ginkgo.It("fails when stream_arn is missing (nil)", func() {
		input := &AwsKinesisStreamConsumer{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsKinesisStreamConsumer",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-consumer",
			},
			Spec: &AwsKinesisStreamConsumerSpec{
				Region: "us-west-2",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// api.proto: api_version and kind constants
	// -------------------------------------------------------------------------

	ginkgo.It("fails when api_version is wrong", func() {
		input := &AwsKinesisStreamConsumer{
			ApiVersion: "wrong.planton.dev/v1",
			Kind:       "AwsKinesisStreamConsumer",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-consumer",
			},
			Spec: &AwsKinesisStreamConsumerSpec{
				Region: "us-west-2",
				StreamArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "arn:aws:kinesis:us-east-1:123456789012:stream/test",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		input := &AwsKinesisStreamConsumer{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "WrongKind",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-consumer",
			},
			Spec: &AwsKinesisStreamConsumerSpec{
				Region: "us-west-2",
				StreamArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "arn:aws:kinesis:us-east-1:123456789012:stream/test",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		input := &AwsKinesisStreamConsumer{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsKinesisStreamConsumer",
			Spec: &AwsKinesisStreamConsumerSpec{
				Region: "us-west-2",
				StreamArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "arn:aws:kinesis:us-east-1:123456789012:stream/test",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		input := &AwsKinesisStreamConsumer{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsKinesisStreamConsumer",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-consumer",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})

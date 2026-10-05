package awscloudwatchloggroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	fkv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "my-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{Region: "us-west-2"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a log group with 1-day retention", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "encrypted-logs-ref",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 90,
				KmsKeyId: &fkv1.StringValueOrRef{
					LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
						ValueFrom: &fkv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AwsKmsKey,
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "protected-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:                    "us-west-2",
				RetentionInDays:           90,
				DeletionProtectionEnabled: proto.Bool(true),
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts explicit-false deletion protection (the only way to disable an enabled protection)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "unprotected-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:                    "us-west-2",
				DeletionProtectionEnabled: proto.Bool(false),
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready log group with all features", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "prod-app-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:          "us-west-2",
				RetentionInDays: 90,
				KmsKeyId: &fkv1.StringValueOrRef{
					LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
						ValueFrom: &fkv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AwsKmsKey,
							Name:      "log-key",
							FieldPath: "status.outputs.key_arn",
						},
					},
				},
				LogGroupClass:             "STANDARD",
				DeletionProtectionEnabled: proto.Bool(true),
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
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
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{Region: "us-west-2"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "WrongKind",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{Region: "us-west-2"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Spec:       &AwsCloudwatchLogGroupSpec{Region: "us-west-2"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-logs",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Metric filters
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a log group with a counting metric filter", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "app-logs-with-filter",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{
					{
						Name:    "error-count",
						Pattern: "ERROR",
						Transformation: &AwsCloudwatchLogGroupMetricTransformation{
							MetricName:      "ErrorCount",
							MetricNamespace: "MyApp/Errors",
							MetricValue:     "1",
							DefaultValue:    float64Ptr(0),
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a metric filter extracting a value with dimensions", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "latency-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{
					{
						Name:    "latency-by-route",
						Pattern: "{ $.latencyMs = * }",
						Transformation: &AwsCloudwatchLogGroupMetricTransformation{
							MetricName:      "RequestLatency",
							MetricNamespace: "MyApp/Performance",
							MetricValue:     "$.latencyMs",
							Dimensions: map[string]string{
								"Route": "$.route",
							},
							Unit: "Milliseconds",
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when a metric filter has no transformation", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "filter-no-transformation",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{
					{
						Name:    "broken",
						Pattern: "ERROR",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metric filter names are duplicated", func() {
		filter := func() *AwsCloudwatchLogGroupMetricFilter {
			return &AwsCloudwatchLogGroupMetricFilter{
				Name:    "dup",
				Pattern: "ERROR",
				Transformation: &AwsCloudwatchLogGroupMetricTransformation{
					MetricName:      "ErrorCount",
					MetricNamespace: "MyApp",
					MetricValue:     "1",
				},
			}
		}
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "dup-filter-names",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:        "us-west-2",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{filter(), filter()},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when default_value is combined with dimensions", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "default-with-dimensions",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{
					{
						Name:    "conflicted",
						Pattern: "{ $.latencyMs = * }",
						Transformation: &AwsCloudwatchLogGroupMetricTransformation{
							MetricName:      "RequestLatency",
							MetricNamespace: "MyApp",
							MetricValue:     "$.latencyMs",
							DefaultValue:    float64Ptr(0),
							Dimensions: map[string]string{
								"Route": "$.route",
							},
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a metric filter has more than 3 dimensions", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-dimensions",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{
					{
						Name:    "over-dimensional",
						Pattern: "{ $.a = * }",
						Transformation: &AwsCloudwatchLogGroupMetricTransformation{
							MetricName:      "M",
							MetricNamespace: "MyApp",
							MetricValue:     "$.a",
							Dimensions: map[string]string{
								"A": "$.a", "B": "$.b", "C": "$.c", "D": "$.d",
							},
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when the transformation unit is not a valid StandardUnit", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-unit",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{
					{
						Name:    "bad-unit-filter",
						Pattern: "ERROR",
						Transformation: &AwsCloudwatchLogGroupMetricTransformation{
							MetricName:      "ErrorCount",
							MetricNamespace: "MyApp",
							MetricValue:     "1",
							Unit:            "Percentage",
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Subscription filters
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a subscription filter to a Kinesis stream with a delivery role", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "streamed-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				SubscriptionFilters: []*AwsCloudwatchLogGroupSubscriptionFilter{
					{
						Name: "to-analytics",
						DestinationArn: &fkv1.StringValueOrRef{
							LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
								ValueFrom: &fkv1.ValueFromRef{
									Kind:      catalogkind.CatalogKind_AwsKinesisStream,
									Name:      "analytics-stream",
									FieldPath: "status.outputs.stream_arn",
								},
							},
						},
						RoleArn: &fkv1.StringValueOrRef{
							LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
								ValueFrom: &fkv1.ValueFromRef{
									Kind: catalogkind.CatalogKind_AwsIamRole,
									Name: "cwl-to-kinesis",
								},
							},
						},
						FilterPattern:    "",
						Distribution:     "ByLogStream",
						EmitSystemFields: []string{"@aws.account", "@aws.region"},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when more than 2 subscription filters are configured", func() {
		filter := func(name string) *AwsCloudwatchLogGroupSubscriptionFilter {
			return &AwsCloudwatchLogGroupSubscriptionFilter{
				Name: name,
				DestinationArn: &fkv1.StringValueOrRef{
					LiteralOrRef: &fkv1.StringValueOrRef_Value{
						Value: "arn:aws:kinesis:us-west-2:123456789012:stream/" + name,
					},
				},
			}
		}
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-subscriptions",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				SubscriptionFilters: []*AwsCloudwatchLogGroupSubscriptionFilter{
					filter("a"), filter("b"), filter("c"),
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a subscription filter has no destination", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "subscription-no-destination",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				SubscriptionFilters: []*AwsCloudwatchLogGroupSubscriptionFilter{
					{Name: "dangling"},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when distribution is an invalid value", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-distribution",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				SubscriptionFilters: []*AwsCloudwatchLogGroupSubscriptionFilter{
					{
						Name: "bad",
						DestinationArn: &fkv1.StringValueOrRef{
							LiteralOrRef: &fkv1.StringValueOrRef_Value{
								Value: "arn:aws:kinesis:us-west-2:123456789012:stream/x",
							},
						},
						Distribution: "RoundRobin",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when emit_system_fields carries an unsupported field", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-system-field",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				SubscriptionFilters: []*AwsCloudwatchLogGroupSubscriptionFilter{
					{
						Name: "bad",
						DestinationArn: &fkv1.StringValueOrRef{
							LiteralOrRef: &fkv1.StringValueOrRef_Value{
								Value: "arn:aws:kinesis:us-west-2:123456789012:stream/x",
							},
						},
						EmitSystemFields: []string{"@aws.availability_zone"},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when subscription filter names are duplicated", func() {
		filter := func() *AwsCloudwatchLogGroupSubscriptionFilter {
			return &AwsCloudwatchLogGroupSubscriptionFilter{
				Name: "dup",
				DestinationArn: &fkv1.StringValueOrRef{
					LiteralOrRef: &fkv1.StringValueOrRef_Value{
						Value: "arn:aws:kinesis:us-west-2:123456789012:stream/x",
					},
				},
			}
		}
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "dup-subscription-names",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:              "us-west-2",
				SubscriptionFilters: []*AwsCloudwatchLogGroupSubscriptionFilter{filter(), filter()},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Class / filter coupling
	// -------------------------------------------------------------------------

	ginkgo.It("fails when an INFREQUENT_ACCESS group configures a metric filter", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "ia-with-filter",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:        "us-west-2",
				LogGroupClass: "INFREQUENT_ACCESS",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{
					{
						Name:    "not-allowed",
						Pattern: "ERROR",
						Transformation: &AwsCloudwatchLogGroupMetricTransformation{
							MetricName:      "ErrorCount",
							MetricNamespace: "MyApp",
							MetricValue:     "1",
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Policies (Struct folds)
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a data protection policy and a field index policy", func() {
		dataProtection, err := structpb.NewStruct(map[string]any{
			"Name":    "protect-pii",
			"Version": "2021-06-01",
			"Statement": []any{
				map[string]any{
					"Sid":            "audit",
					"DataIdentifier": []any{"arn:aws:dataprotection::aws:data-identifier/EmailAddress"},
					"Operation":      map[string]any{"Audit": map[string]any{"FindingsDestination": map[string]any{}}},
				},
				map[string]any{
					"Sid":            "redact",
					"DataIdentifier": []any{"arn:aws:dataprotection::aws:data-identifier/EmailAddress"},
					"Operation":      map[string]any{"Deidentify": map[string]any{"MaskConfig": map[string]any{}}},
				},
			},
		})
		gomega.Expect(err).To(gomega.BeNil())

		fieldIndex, err := structpb.NewStruct(map[string]any{
			"Fields": []any{"requestId", "userId"},
		})
		gomega.Expect(err).To(gomega.BeNil())

		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "governed-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:               "us-west-2",
				DataProtectionPolicy: dataProtection,
				FieldIndexPolicy:     fieldIndex,
			},
		}
		err = protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Filter name formats (PutMetricFilter / PutSubscriptionFilter contracts)
	// -------------------------------------------------------------------------

	ginkgo.It("fails when a metric filter name contains a colon", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-filter-name",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{
					{
						Name:    "errors:count",
						Pattern: "ERROR",
						Transformation: &AwsCloudwatchLogGroupMetricTransformation{
							MetricName:      "ErrorCount",
							MetricNamespace: "MyApp",
							MetricValue:     "1",
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a metric transformation name contains a dollar sign", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-metric-name",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				MetricFilters: []*AwsCloudwatchLogGroupMetricFilter{
					{
						Name:    "errors",
						Pattern: "ERROR",
						Transformation: &AwsCloudwatchLogGroupMetricTransformation{
							MetricName:      "Error$Count",
							MetricNamespace: "MyApp",
							MetricValue:     "1",
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a subscription filter name contains an asterisk", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-subscription-name",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				SubscriptionFilters: []*AwsCloudwatchLogGroupSubscriptionFilter{
					{
						Name: "fan*out",
						DestinationArn: &fkv1.StringValueOrRef{
							LiteralOrRef: &fkv1.StringValueOrRef_Value{
								Value: "arn:aws:kinesis:us-west-2:123456789012:stream/logs",
							},
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// emit_system_fields (incl. @source.log, provider 6.56.0)
	// -------------------------------------------------------------------------

	ginkgo.It("accepts @source.log in emit_system_fields", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "source-log-enriched",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				SubscriptionFilters: []*AwsCloudwatchLogGroupSubscriptionFilter{
					{
						Name: "central-delivery",
						DestinationArn: &fkv1.StringValueOrRef{
							LiteralOrRef: &fkv1.StringValueOrRef_Value{
								Value: "arn:aws:kinesis:us-west-2:123456789012:stream/logs",
							},
						},
						EmitSystemFields: []string{"@aws.account", "@aws.region", "@source.log"},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Log streams fold
	// -------------------------------------------------------------------------

	ginkgo.It("accepts pre-created log stream names", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "streamed-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:     "us-west-2",
				LogStreams: []string{"agent-primary", "agent-secondary"},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when log stream names are duplicated", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "dup-streams",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:     "us-west-2",
				LogStreams: []string{"agent", "agent"},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a log stream name contains a colon", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-stream-name",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:     "us-west-2",
				LogStreams: []string{"agent:primary"},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Transformer fold
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a transformer whose pipeline starts with a parser", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "transformed-logs",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{}},
						{RenameKeys: &AwsCloudwatchLogGroupTransformerRenameKeys{
							Entries: []*AwsCloudwatchLogGroupTransformerRenameKeysEntry{
								{Key: "msg", RenameTo: "message"},
							},
						}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when the first processor is not a parser", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "mutate-first",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{RenameKeys: &AwsCloudwatchLogGroupTransformerRenameKeys{
							Entries: []*AwsCloudwatchLogGroupTransformerRenameKeysEntry{
								{Key: "msg", RenameTo: "message"},
							},
						}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a pipeline entry sets no processor", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "empty-entry",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{}},
						{},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a pipeline entry sets two processors", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "double-entry",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{
							ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{},
							TrimString: &AwsCloudwatchLogGroupTransformerWithKeys{
								WithKeys: []string{"message"},
							},
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when grok appears twice in one pipeline", func() {
		grokEntry := func() *AwsCloudwatchLogGroupTransformerProcessor {
			return &AwsCloudwatchLogGroupTransformerProcessor{
				Grok: &AwsCloudwatchLogGroupTransformerGrok{Match: "%{COMMONAPACHELOG}"},
			}
		}
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "double-grok",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{grokEntry(), grokEntry()},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a vended parser is not the first processor", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "vended-not-first",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{}},
						{ParseVpc: &AwsCloudwatchLogGroupTransformerVendedLogParser{}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a vended parser as the first processor", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "vpc-flow-parsed",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseVpc: &AwsCloudwatchLogGroupTransformerVendedLogParser{Source: "@message"}},
						{AddKeys: &AwsCloudwatchLogGroupTransformerAddKeys{
							Entries: []*AwsCloudwatchLogGroupTransformerAddKeysEntry{
								{Key: "environment", Value: "production"},
							},
						}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when a vended parser source is not @message", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-vended-source",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseWaf: &AwsCloudwatchLogGroupTransformerVendedLogParser{Source: "detail.message"}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when parse_to_ocsf uses an unsupported event_source", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-ocsf-source",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseToOcsf: &AwsCloudwatchLogGroupTransformerParseToOcsf{
							EventSource: "GuardDuty",
							OcsfVersion: "V1.1",
						}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts parse_to_ocsf with OCSF V1.5", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "ocsf-converted",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseToOcsf: &AwsCloudwatchLogGroupTransformerParseToOcsf{
							EventSource: "VPCFlow",
							OcsfVersion: "V1.5",
						}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when list_to_map flattens without naming the surviving element", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "flatten-without-element",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{}},
						{ListToMap: &AwsCloudwatchLogGroupTransformerListToMap{
							Source:  "items",
							Key:     "name",
							Flatten: true,
						}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a transformer is configured on an INFREQUENT_ACCESS group", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "ia-with-transformer",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region:        "us-west-2",
				LogGroupClass: "INFREQUENT_ACCESS",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when the pipeline exceeds 20 processors", func() {
		processors := []*AwsCloudwatchLogGroupTransformerProcessor{
			{ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{}},
		}
		for i := 0; i < 20; i++ {
			processors = append(processors, &AwsCloudwatchLogGroupTransformerProcessor{
				TrimString: &AwsCloudwatchLogGroupTransformerWithKeys{WithKeys: []string{"message"}},
			})
		}
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "oversized-pipeline",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: processors,
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when parse_json appears six times", func() {
		processors := make([]*AwsCloudwatchLogGroupTransformerProcessor, 0, 6)
		for i := 0; i < 6; i++ {
			processors = append(processors, &AwsCloudwatchLogGroupTransformerProcessor{
				ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{},
			})
		}
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-parse-json",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: processors,
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when add_keys has more than five entries", func() {
		entries := make([]*AwsCloudwatchLogGroupTransformerAddKeysEntry, 0, 6)
		for _, key := range []string{"a", "b", "c", "d", "e", "f"} {
			entries = append(entries, &AwsCloudwatchLogGroupTransformerAddKeysEntry{Key: key, Value: "v"})
		}
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-add-keys",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{}},
						{AddKeys: &AwsCloudwatchLogGroupTransformerAddKeys{Entries: entries}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when type_converter uses an unsupported type", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-type-conversion",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{ParseJson: &AwsCloudwatchLogGroupTransformerParseJson{}},
						{TypeConverter: &AwsCloudwatchLogGroupTransformerTypeConverter{
							Entries: []*AwsCloudwatchLogGroupTransformerTypeConverterEntry{
								{Key: "count", Type: "float"},
							},
						}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a full observability pipeline (grok + datetime + substitute + type conversion)", func() {
		input := &AwsCloudwatchLogGroup{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchLogGroup",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "full-pipeline",
			},
			Spec: &AwsCloudwatchLogGroupSpec{
				Region: "us-west-2",
				Transformer: &AwsCloudwatchLogGroupTransformer{
					Processors: []*AwsCloudwatchLogGroupTransformerProcessor{
						{Grok: &AwsCloudwatchLogGroupTransformerGrok{Match: "%{COMMONAPACHELOG}"}},
						{DateTimeConverter: &AwsCloudwatchLogGroupTransformerDateTimeConverter{
							Source:        "timestamp",
							Target:        "ingestedAt",
							MatchPatterns: []string{"dd/MMM/yyyy:HH:mm:ss Z"},
						}},
						{SubstituteString: &AwsCloudwatchLogGroupTransformerSubstituteString{
							Entries: []*AwsCloudwatchLogGroupTransformerSubstituteStringEntry{
								{Source: "request", From: "/api/v[0-9]+", To: "/api"},
							},
						}},
						{TypeConverter: &AwsCloudwatchLogGroupTransformerTypeConverter{
							Entries: []*AwsCloudwatchLogGroupTransformerTypeConverterEntry{
								{Key: "bytes", Type: "integer"},
							},
						}},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})
})

func float64Ptr(f float64) *float64 { return &f }

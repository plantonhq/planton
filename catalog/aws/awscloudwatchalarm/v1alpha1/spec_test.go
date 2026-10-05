package awscloudwatchalarmv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	fkv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsCloudwatchAlarmSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCloudwatchAlarmSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsCloudwatchAlarmSpec validations", func() {

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal simple metric alarm (CPUUtilization)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "cpu-high",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanOrEqualToThreshold",
				EvaluationPeriods:  1,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts simple metric with GreaterThanThreshold operator", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "cpu-gt",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          90.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Maximum",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts simple metric with LessThanThreshold operator", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "free-mem-low",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "LessThanThreshold",
				EvaluationPeriods:  3,
				Threshold:          100000000,
				MetricName:         "FreeableMemory",
				Namespace:          "AWS/RDS",
				Period:             60,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts simple metric with LessThanOrEqualToThreshold operator", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "disk-space-low",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "LessThanOrEqualToThreshold",
				EvaluationPeriods:  2,
				Threshold:          10.0,
				MetricName:         "DiskSpaceAvailable",
				Namespace:          "System/Linux",
				Period:             120,
				Statistic:          "Minimum",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts simple metric with extended_statistic (p95)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "latency-p95",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  3,
				Threshold:          500.0,
				MetricName:         "TargetResponseTime",
				Namespace:          "AWS/ApplicationELB",
				Period:             60,
				ExtendedStatistic:  "p95",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts simple metric with dimensions", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "ec2-cpu-instance",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanOrEqualToThreshold",
				EvaluationPeriods:  2,
				Threshold:          85.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				Dimensions: map[string]string{
					"InstanceId": "i-1234567890abcdef0",
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a metric math alarm (error rate = m1/m2*100)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "error-rate-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  3,
				Threshold:          5.0,
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "5XXError",
							Namespace:  "AWS/ApplicationELB",
							Period:     300,
							Stat:       "Sum",
						},
						ReturnData: false,
					},
					{
						Id: "m2",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "RequestCount",
							Namespace:  "AWS/ApplicationELB",
							Period:     300,
							Stat:       "Sum",
						},
						ReturnData: false,
					},
					{
						Id:         "e1",
						Expression: "m1/m2*100",
						Label:      "Error Rate (%)",
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an anomaly detection alarm (threshold_metric_id)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "cpu-anomaly",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "LessThanLowerOrGreaterThanUpperThreshold",
				EvaluationPeriods:  3,
				ThresholdMetricId:  "ad1",
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CPUUtilization",
							Namespace:  "AWS/EC2",
							Period:     300,
							Stat:       "Average",
						},
						ReturnData: true,
					},
					{
						Id:         "ad1",
						Expression: "ANOMALY_DETECTION_BAND(m1, 2)",
						Label:      "Anomaly Detection Band",
						ReturnData: false,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts alarm with all 3 action types (alarm, ok, insufficient_data)", func() {
		snsArn := &fkv1.StringValueOrRef{
			LiteralOrRef: &fkv1.StringValueOrRef_Value{
				Value: "arn:aws:sns:us-east-1:123456789012:my-topic",
			},
		}
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "full-actions-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:                  "us-west-2",
				ComparisonOperator:      "GreaterThanOrEqualToThreshold",
				EvaluationPeriods:       1,
				Threshold:               80.0,
				MetricName:              "CPUUtilization",
				Namespace:               "AWS/EC2",
				Period:                  300,
				Statistic:               "Average",
				AlarmActions:            []*fkv1.StringValueOrRef{snsArn},
				OkActions:               []*fkv1.StringValueOrRef{snsArn},
				InsufficientDataActions: []*fkv1.StringValueOrRef{snsArn},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts alarm with actions via valueFrom (SNS topic reference)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "ref-actions-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanOrEqualToThreshold",
				EvaluationPeriods:  1,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				AlarmActions: []*fkv1.StringValueOrRef{
					{
						LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
							ValueFrom: &fkv1.ValueFromRef{
								Kind:      catalogkind.CatalogKind_AwsSnsTopic,
								Name:      "alerts-topic",
								FieldPath: "status.outputs.topic_arn",
							},
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts M-of-N evaluation (datapoints_to_alarm < evaluation_periods)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "m-of-n-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanOrEqualToThreshold",
				EvaluationPeriods:  5,
				DatapointsToAlarm:  3,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             60,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts treat_missing_data = breaching", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "heartbeat-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "LessThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          1.0,
				MetricName:         "Heartbeat",
				Namespace:          "Custom/App",
				Period:             60,
				Statistic:          "SampleCount",
				TreatMissingData:   "breaching",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts treat_missing_data = notBreaching", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "intermittent-errors",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  3,
				Threshold:          10.0,
				MetricName:         "Errors",
				Namespace:          "AWS/Lambda",
				Period:             300,
				Statistic:          "Sum",
				TreatMissingData:   "notBreaching",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts actions_enabled = false (suppressed alarm)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "maintenance-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanOrEqualToThreshold",
				EvaluationPeriods:  1,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				ActionsEnabled:     boolPtr(false),
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts high-resolution alarm (period=10)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "high-res-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  6,
				Threshold:          95.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             10,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts evaluate_low_sample_count_percentiles = ignore", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "percentile-ignore",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:                            "us-west-2",
				ComparisonOperator:                "GreaterThanThreshold",
				EvaluationPeriods:                 3,
				Threshold:                         500.0,
				MetricName:                        "TargetResponseTime",
				Namespace:                         "AWS/ApplicationELB",
				Period:                            60,
				ExtendedStatistic:                 "p99",
				EvaluateLowSampleCountPercentiles: "ignore",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready alarm with everything set", func() {
		snsArn := &fkv1.StringValueOrRef{
			LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
				ValueFrom: &fkv1.ValueFromRef{
					Kind:      catalogkind.CatalogKind_AwsSnsTopic,
					Name:      "prod-alerts",
					FieldPath: "status.outputs.topic_arn",
				},
			},
		}
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "prod-cpu-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanOrEqualToThreshold",
				EvaluationPeriods:  5,
				DatapointsToAlarm:  3,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/ECS",
				Period:             60,
				Statistic:          "Average",
				Dimensions: map[string]string{
					"ClusterName": "production",
					"ServiceName": "api",
				},
				TreatMissingData:        "notBreaching",
				ActionsEnabled:          boolPtr(true),
				AlarmDescription:        "ECS API service CPU utilization is too high. Consider scaling out or investigating hot code paths.",
				AlarmActions:            []*fkv1.StringValueOrRef{snsArn},
				OkActions:               []*fkv1.StringValueOrRef{snsArn},
				InsufficientDataActions: []*fkv1.StringValueOrRef{snsArn},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: comparison_operator_valid
	// -------------------------------------------------------------------------

	ginkgo.It("fails when comparison_operator is invalid", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-operator",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "EqualTo",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: evaluation_periods (int32.gte = 1)
	// -------------------------------------------------------------------------

	ginkgo.It("fails when evaluation_periods is 0", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-eval-periods",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: statistic_valid
	// -------------------------------------------------------------------------

	ginkgo.It("fails when statistic is an invalid value", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-stat",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Median",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: treat_missing_data_valid
	// -------------------------------------------------------------------------

	ginkgo.It("fails when treat_missing_data is an invalid value", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-missing-data",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				TreatMissingData:   "skip",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: evaluate_low_sample_count_percentiles_valid
	// -------------------------------------------------------------------------

	ginkgo.It("fails when evaluate_low_sample_count_percentiles is an invalid value", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-low-sample",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:                            "us-west-2",
				ComparisonOperator:                "GreaterThanThreshold",
				EvaluationPeriods:                 1,
				MetricName:                        "TargetResponseTime",
				Namespace:                         "AWS/ApplicationELB",
				Period:                            60,
				ExtendedStatistic:                 "p95",
				EvaluateLowSampleCountPercentiles: "skip",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: statistic_extended_statistic_exclusive
	// -------------------------------------------------------------------------

	ginkgo.It("fails when both statistic and extended_statistic are set", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "both-stats",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				ExtendedStatistic:  "p95",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: simple_metric_or_metric_queries
	// -------------------------------------------------------------------------

	ginkgo.It("fails when both metric_name and metric_queries are set", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "both-modes",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CPUUtilization",
							Namespace:  "AWS/EC2",
							Period:     300,
							Stat:       "Average",
						},
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: metric_source_required
	// -------------------------------------------------------------------------

	ginkgo.It("fails when neither metric_name nor metric_queries is set", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-metric-source",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: namespace_required_with_metric_name
	// -------------------------------------------------------------------------

	ginkgo.It("fails when metric_name is set but namespace is missing", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-namespace",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Period:             300,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: period_required_with_metric_name
	// -------------------------------------------------------------------------

	ginkgo.It("fails when metric_name is set but period is missing", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-period",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: statistic_required_with_metric_name
	// -------------------------------------------------------------------------

	ginkgo.It("fails when metric_name is set but no statistic or extended_statistic", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-statistic",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: period_valid_values
	// -------------------------------------------------------------------------

	ginkgo.It("fails when period is 45 (not 10, 20, 30, or multiple of 60)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-period-45",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             45,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when period is 100 (not a multiple of 60)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-period-100",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             100,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: datapoints_to_alarm_lte_evaluation_periods
	// -------------------------------------------------------------------------

	ginkgo.It("fails when datapoints_to_alarm > evaluation_periods", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-datapoints",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  3,
				DatapointsToAlarm:  5,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: metric_queries_max_20
	// -------------------------------------------------------------------------

	ginkgo.It("fails when more than 20 metric_queries are provided", func() {
		queries := make([]*AwsCloudwatchAlarmMetricQuery, 21)
		for i := 0; i < 21; i++ {
			queries[i] = &AwsCloudwatchAlarmMetricQuery{
				Id: "m" + string(rune('a'+i)),
				Metric: &AwsCloudwatchAlarmMetricQueryMetric{
					MetricName: "CPUUtilization",
					Namespace:  "AWS/EC2",
					Period:     300,
					Stat:       "Average",
				},
				ReturnData: i == 0,
			}
		}
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-queries",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricQueries:      queries,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: alarm_actions_max_5
	// -------------------------------------------------------------------------

	ginkgo.It("fails when more than 5 alarm_actions are provided", func() {
		actions := make([]*fkv1.StringValueOrRef, 6)
		for i := range actions {
			actions[i] = &fkv1.StringValueOrRef{
				LiteralOrRef: &fkv1.StringValueOrRef_Value{
					Value: "arn:aws:sns:us-east-1:123456789012:topic-" + string(rune('a'+i)),
				},
			}
		}
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-alarm-actions",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				AlarmActions:       actions,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: ok_actions_max_5
	// -------------------------------------------------------------------------

	ginkgo.It("fails when more than 5 ok_actions are provided", func() {
		actions := make([]*fkv1.StringValueOrRef, 6)
		for i := range actions {
			actions[i] = &fkv1.StringValueOrRef{
				LiteralOrRef: &fkv1.StringValueOrRef_Value{
					Value: "arn:aws:sns:us-east-1:123456789012:topic-" + string(rune('a'+i)),
				},
			}
		}
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-ok-actions",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				OkActions:          actions,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: insufficient_data_actions_max_5
	// -------------------------------------------------------------------------

	ginkgo.It("fails when more than 5 insufficient_data_actions are provided", func() {
		actions := make([]*fkv1.StringValueOrRef, 6)
		for i := range actions {
			actions[i] = &fkv1.StringValueOrRef{
				LiteralOrRef: &fkv1.StringValueOrRef_Value{
					Value: "arn:aws:sns:us-east-1:123456789012:topic-" + string(rune('a'+i)),
				},
			}
		}
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-insuf-actions",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:                  "us-west-2",
				ComparisonOperator:      "GreaterThanThreshold",
				EvaluationPeriods:       1,
				MetricName:              "CPUUtilization",
				Namespace:               "AWS/EC2",
				Period:                  300,
				Statistic:               "Average",
				InsufficientDataActions: actions,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// api.proto: api_version and kind constants
	// -------------------------------------------------------------------------

	ginkgo.It("fails when api_version is wrong", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "wrong.planton.dev/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-version",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "WrongKind",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-kind",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-spec",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// PromQL mode (evaluation_criteria)
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a PromQL alarm (evaluation_criteria only, no threshold fields)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "promql-cpu-high",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region: "us-west-2",
				EvaluationCriteria: &AwsCloudwatchAlarmEvaluationCriteria{
					PromqlCriteria: &AwsCloudwatchAlarmPromqlCriteria{
						Query:          `avg(rate(node_cpu_seconds_total{mode!="idle"}[5m])) > 0.8`,
						PendingPeriod:  int32Ptr(300),
						RecoveryPeriod: int32Ptr(120),
					},
				},
				EvaluationInterval: 60,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when evaluation_criteria is combined with metric_name", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "promql-and-metric",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				EvaluationCriteria: &AwsCloudwatchAlarmEvaluationCriteria{
					PromqlCriteria: &AwsCloudwatchAlarmPromqlCriteria{
						Query: "up == 0",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a PromQL alarm sets comparison_operator", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "promql-with-operator",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationCriteria: &AwsCloudwatchAlarmEvaluationCriteria{
					PromqlCriteria: &AwsCloudwatchAlarmPromqlCriteria{
						Query: "up == 0",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a PromQL alarm sets evaluation_periods", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "promql-with-periods",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:            "us-west-2",
				EvaluationPeriods: 3,
				EvaluationCriteria: &AwsCloudwatchAlarmEvaluationCriteria{
					PromqlCriteria: &AwsCloudwatchAlarmPromqlCriteria{
						Query: "up == 0",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when evaluation_interval is set without evaluation_criteria", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "interval-without-promql",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				EvaluationInterval: 60,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when evaluation_interval is 45 (not 10/20/30 or a multiple of 60)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "promql-bad-interval",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region: "us-west-2",
				EvaluationCriteria: &AwsCloudwatchAlarmEvaluationCriteria{
					PromqlCriteria: &AwsCloudwatchAlarmPromqlCriteria{
						Query: "up == 0",
					},
				},
				EvaluationInterval: 45,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when evaluation_criteria is present without promql_criteria", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "empty-criteria",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				EvaluationCriteria: &AwsCloudwatchAlarmEvaluationCriteria{},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when pending_period exceeds 86400 seconds", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "promql-pending-too-long",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region: "us-west-2",
				EvaluationCriteria: &AwsCloudwatchAlarmEvaluationCriteria{
					PromqlCriteria: &AwsCloudwatchAlarmPromqlCriteria{
						Query:         "up == 0",
						PendingPeriod: int32Ptr(90000),
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Per-query contract (expression XOR metric; exactly one return_data)
	// -------------------------------------------------------------------------

	ginkgo.It("fails when a metric query sets both expression and metric", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "query-both-arms",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id:         "m1",
						Expression: "m2*2",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CPUUtilization",
							Namespace:  "AWS/EC2",
							Period:     300,
							Stat:       "Average",
						},
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a metric query sets neither expression nor metric", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "query-neither-arm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id:         "m1",
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when no metric query sets return_data", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-return-data",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CPUUtilization",
							Namespace:  "AWS/EC2",
							Period:     300,
							Stat:       "Average",
						},
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when two metric queries set return_data", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "double-return-data",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "HTTPCode_Target_5XX_Count",
							Namespace:  "AWS/ApplicationELB",
							Period:     60,
							Stat:       "Sum",
						},
						ReturnData: true,
					},
					{
						Id: "m2",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "RequestCount",
							Namespace:  "AWS/ApplicationELB",
							Period:     60,
							Stat:       "Sum",
						},
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a query metric without namespace (provider-optional)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-namespace-metric",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CustomMetric",
							Period:     60,
							Stat:       "Sum",
						},
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when a query metric period is 45", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-query-period",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CPUUtilization",
							Namespace:  "AWS/EC2",
							Period:     45,
							Stat:       "Average",
						},
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when unit is not a valid CloudWatch StandardUnit", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-unit",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				Unit:               "Percentage",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: threshold_conflicts_with_threshold_metric_id
	// -------------------------------------------------------------------------

	ginkgo.It("fails when both threshold and threshold_metric_id are set", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "double-threshold",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanUpperThreshold",
				EvaluationPeriods:  3,
				Threshold:          80.0,
				ThresholdMetricId:  "ad1",
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CPUUtilization",
							Namespace:  "AWS/EC2",
							Period:     300,
							Stat:       "Average",
						},
						ReturnData: true,
					},
					{
						Id:         "ad1",
						Expression: "ANOMALY_DETECTION_BAND(m1, 2)",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: metric_queries_forbid_simple_metric_fields
	// -------------------------------------------------------------------------

	ginkgo.It("fails when metric_queries is combined with a top-level unit", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "query-with-stray-unit",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				Unit:               "Count",
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "5XXError",
							Namespace:  "AWS/ApplicationELB",
							Period:     300,
							Stat:       "Sum",
						},
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metric_queries is combined with top-level dimensions", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "query-with-stray-dimensions",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				Dimensions:         map[string]string{"InstanceId": "i-1234567890abcdef0"},
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CPUUtilization",
							Namespace:  "AWS/EC2",
							Period:     300,
							Stat:       "Average",
						},
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: extended_statistic_format / stat_valid_format
	// -------------------------------------------------------------------------

	ginkgo.It("accepts trimmed-mean and percentile extended statistics", func() {
		build := func(name, extendedStat string) *AwsCloudwatchAlarm {
			return &AwsCloudwatchAlarm{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "AwsCloudwatchAlarm",
				Metadata:   &shared.CatalogObjectMetadata{Name: name},
				Spec: &AwsCloudwatchAlarmSpec{
					Region:             "us-west-2",
					ComparisonOperator: "GreaterThanThreshold",
					EvaluationPeriods:  3,
					Threshold:          500.0,
					MetricName:         "TargetResponseTime",
					Namespace:          "AWS/ApplicationELB",
					Period:             60,
					ExtendedStatistic:  extendedStat,
				},
			}
		}
		for _, stat := range []string{"p95", "p99.9", "tm99", "IQM", "TM(10%:90%)", "PR(:300)"} {
			err := protovalidate.Validate(build("ext-stat-ok", stat))
			gomega.Expect(err).To(gomega.BeNil(), "expected %q to be accepted", stat)
		}
	})

	ginkgo.It("fails when extended_statistic is not a valid percentile form", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-ext-stat",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  3,
				Threshold:          500.0,
				MetricName:         "TargetResponseTime",
				Namespace:          "AWS/ApplicationELB",
				Period:             60,
				ExtendedStatistic:  "percentile95",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a metric query stat is neither standard nor percentile", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-query-stat",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CPUUtilization",
							Namespace:  "AWS/EC2",
							Period:     300,
							Stat:       "Mean",
						},
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: namespace_no_leading_colon (top-level and metric query)
	// -------------------------------------------------------------------------

	ginkgo.It("fails when the namespace starts with a colon", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-namespace",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          ":AWS/EC2",
				Period:             300,
				Statistic:          "Average",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a metric query namespace starts with a colon", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "bad-query-namespace",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				Region:             "us-west-2",
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				Threshold:          5.0,
				MetricQueries: []*AwsCloudwatchAlarmMetricQuery{
					{
						Id: "m1",
						Metric: &AwsCloudwatchAlarmMetricQueryMetric{
							MetricName: "CPUUtilization",
							Namespace:  ":AWS/EC2",
							Period:     300,
							Stat:       "Average",
						},
						ReturnData: true,
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})

func boolPtr(b bool) *bool { return &b }

func int32Ptr(i int32) *int32 { return &i }

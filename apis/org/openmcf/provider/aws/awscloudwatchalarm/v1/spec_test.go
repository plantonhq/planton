package awscloudwatchalarmv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "cpu-high",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "cpu-gt",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "free-mem-low",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "disk-space-low",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "latency-p95",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "ec2-cpu-instance",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "error-rate-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "cpu-anomaly",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "full-actions-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				ComparisonOperator: "GreaterThanOrEqualToThreshold",
				EvaluationPeriods:  1,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				AlarmActions:       []*fkv1.StringValueOrRef{snsArn},
				OkActions:          []*fkv1.StringValueOrRef{snsArn},
				InsufficientDataActions: []*fkv1.StringValueOrRef{snsArn},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts alarm with actions via valueFrom (SNS topic reference)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "ref-actions-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
								Kind:      cloudresourcekind.CloudResourceKind_AwsSnsTopic,
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "m-of-n-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "heartbeat-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "intermittent-errors",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "maintenance-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				ComparisonOperator: "GreaterThanOrEqualToThreshold",
				EvaluationPeriods:  1,
				Threshold:          80.0,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
				ActionsEnabled:     false,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts high-resolution alarm (period=10)", func() {
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "high-res-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "percentile-ignore",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
					Kind:      cloudresourcekind.CloudResourceKind_AwsSnsTopic,
					Name:      "prod-alerts",
					FieldPath: "status.outputs.topic_arn",
				},
			},
		}
		input := &AwsCloudwatchAlarm{
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "prod-cpu-alarm",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
				TreatMissingData:   "notBreaching",
				ActionsEnabled:     true,
				AlarmDescription:   "ECS API service CPU utilization is too high. Consider scaling out or investigating hot code paths.",
				AlarmActions:       []*fkv1.StringValueOrRef{snsArn},
				OkActions:          []*fkv1.StringValueOrRef{snsArn},
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-operator",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-eval-periods",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-stat",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-missing-data",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-low-sample",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "both-stats",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "both-modes",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "no-metric-source",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "no-namespace",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "no-period",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "no-statistic",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-period-45",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-period-100",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-datapoints",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "too-many-queries",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "too-many-alarm-actions",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "too-many-ok-actions",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "too-many-insuf-actions",
			},
			Spec: &AwsCloudwatchAlarmSpec{
				ComparisonOperator: "GreaterThanThreshold",
				EvaluationPeriods:  1,
				MetricName:         "CPUUtilization",
				Namespace:          "AWS/EC2",
				Period:             300,
				Statistic:          "Average",
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
			ApiVersion: "wrong.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-version",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "WrongKind",
			Metadata: &shared.CloudResourceMetadata{
				Name: "bad-kind",
			},
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Spec: &AwsCloudwatchAlarmSpec{
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
			ApiVersion: "aws.openmcf.org/v1",
			Kind:       "AwsCloudwatchAlarm",
			Metadata: &shared.CloudResourceMetadata{
				Name: "no-spec",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})

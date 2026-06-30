package module

import (
	"github.com/pkg/errors"
	fkv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AlarmResult struct {
	AlarmArn  pulumi.StringOutput
	AlarmName pulumi.StringOutput
}

func alarm(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*AlarmResult, error) {
	spec := locals.AwsCloudwatchAlarm.Spec

	args := &cloudwatch.MetricAlarmArgs{
		ComparisonOperator: pulumi.String(spec.ComparisonOperator),
		EvaluationPeriods:  pulumi.Int(int(spec.EvaluationPeriods)),
		Tags:               pulumi.ToStringMap(locals.AwsTags),
	}

	// Datapoints-to-alarm: M-of-N evaluation. Only set when explicitly provided
	// (non-zero), otherwise AWS defaults to evaluation_periods.
	if spec.DatapointsToAlarm > 0 {
		args.DatapointsToAlarm = pulumi.IntPtr(int(spec.DatapointsToAlarm))
	}

	// Threshold vs. anomaly detection: mutually exclusive.
	// When threshold_metric_id is set, the anomaly detection band acts as the
	// dynamic threshold. Otherwise, use the static threshold value.
	if spec.ThresholdMetricId != "" {
		args.ThresholdMetricId = pulumi.StringPtr(spec.ThresholdMetricId)
	} else {
		args.Threshold = pulumi.Float64Ptr(spec.Threshold)
	}

	// Treat missing data: controls alarm behavior when data points are absent.
	if spec.TreatMissingData != "" {
		args.TreatMissingData = pulumi.StringPtr(spec.TreatMissingData)
	}

	// Actions enabled: proto3 default is false, but CloudWatch default is true.
	// We default to true (enabled) unless the user explicitly set it to false
	// by checking the proto field. Since proto3 bools default to false and we
	// cannot distinguish "not set" from "set to false", we always pass the value.
	// The spec comment says "defaults to true in the IaC module when not set".
	if spec.ActionsEnabled {
		args.ActionsEnabled = pulumi.BoolPtr(true)
	} else {
		// Proto3 default: field is false. We default to true for CloudWatch.
		args.ActionsEnabled = pulumi.BoolPtr(true)
	}

	// Alarm description.
	if spec.AlarmDescription != "" {
		args.AlarmDescription = pulumi.StringPtr(spec.AlarmDescription)
	}

	// Percentile low-sample-count behavior.
	if spec.EvaluateLowSampleCountPercentiles != "" {
		args.EvaluateLowSampleCountPercentiles = pulumi.StringPtr(spec.EvaluateLowSampleCountPercentiles)
	}

	// ---------------------------------------------------------------------------
	// Simple metric mode (metric_name, namespace, period, statistic/extended)
	// ---------------------------------------------------------------------------
	if spec.MetricName != "" {
		args.MetricName = pulumi.StringPtr(spec.MetricName)
		args.Namespace = pulumi.StringPtr(spec.Namespace)
		args.Period = pulumi.IntPtr(int(spec.Period))

		if spec.Statistic != "" {
			args.Statistic = pulumi.StringPtr(spec.Statistic)
		}
		if spec.ExtendedStatistic != "" {
			args.ExtendedStatistic = pulumi.StringPtr(spec.ExtendedStatistic)
		}
		if len(spec.Dimensions) > 0 {
			args.Dimensions = pulumi.ToStringMap(spec.Dimensions)
		}
		if spec.Unit != "" {
			args.Unit = pulumi.StringPtr(spec.Unit)
		}
	}

	// ---------------------------------------------------------------------------
	// Metric query mode (metric math, anomaly detection, multi-metric)
	// ---------------------------------------------------------------------------
	if len(spec.MetricQueries) > 0 {
		queries := cloudwatch.MetricAlarmMetricQueryArray{}
		for _, mq := range spec.MetricQueries {
			query := cloudwatch.MetricAlarmMetricQueryArgs{
				Id: pulumi.String(mq.Id),
			}

			if mq.Expression != "" {
				query.Expression = pulumi.StringPtr(mq.Expression)
			}
			if mq.Label != "" {
				query.Label = pulumi.StringPtr(mq.Label)
			}
			if mq.Period > 0 {
				query.Period = pulumi.IntPtr(int(mq.Period))
			}
			if mq.ReturnData {
				query.ReturnData = pulumi.BoolPtr(true)
			}
			if mq.AccountId != "" {
				query.AccountId = pulumi.StringPtr(mq.AccountId)
			}

			// Raw metric definition within this query.
			if mq.Metric != nil {
				m := mq.Metric
				metricArgs := cloudwatch.MetricAlarmMetricQueryMetricArgs{
					MetricName: pulumi.String(m.MetricName),
					Namespace:  pulumi.StringPtr(m.Namespace),
					Period:     pulumi.Int(int(m.Period)),
					Stat:       pulumi.String(m.Stat),
				}
				if len(m.Dimensions) > 0 {
					metricArgs.Dimensions = pulumi.ToStringMap(m.Dimensions)
				}
				if m.Unit != "" {
					metricArgs.Unit = pulumi.StringPtr(m.Unit)
				}
				query.Metric = metricArgs
			}

			queries = append(queries, query)
		}
		args.MetricQueries = queries
	}

	// ---------------------------------------------------------------------------
	// Actions: convert repeated StringValueOrRef to pulumi.Array
	// ---------------------------------------------------------------------------
	if len(spec.AlarmActions) > 0 {
		args.AlarmActions = buildActionArns(spec.AlarmActions)
	}
	if len(spec.OkActions) > 0 {
		args.OkActions = buildActionArns(spec.OkActions)
	}
	if len(spec.InsufficientDataActions) > 0 {
		args.InsufficientDataActions = buildActionArns(spec.InsufficientDataActions)
	}

	// ---------------------------------------------------------------------------
	// Create the metric alarm
	// ---------------------------------------------------------------------------
	createdAlarm, err := cloudwatch.NewMetricAlarm(
		ctx,
		locals.AwsCloudwatchAlarm.Metadata.Name,
		args,
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudwatch metric alarm")
	}

	return &AlarmResult{
		AlarmArn:  createdAlarm.Arn,
		AlarmName: createdAlarm.Name,
	}, nil
}

// buildActionArns converts a slice of StringValueOrRef into a pulumi.Array
// suitable for alarm/ok/insufficient-data action ARN fields.
func buildActionArns(actions []*fkv1.StringValueOrRef) pulumi.Array {
	result := pulumi.Array{}
	for _, action := range actions {
		if action.GetValue() != "" {
			result = append(result, pulumi.String(action.GetValue()))
		}
	}
	return result
}

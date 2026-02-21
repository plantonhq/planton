package module

import (
	"strings"

	"github.com/pkg/errors"
	ocialarmv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocialarm/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/monitoring"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func alarmResource(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciAlarm.Spec

	args := &monitoring.AlarmArgs{
		CompartmentId:       pulumi.String(spec.CompartmentId.GetValue()),
		MetricCompartmentId: pulumi.String(spec.MetricCompartmentId.GetValue()),
		Namespace:           pulumi.String(spec.Namespace),
		Query:               pulumi.String(spec.Query),
		Severity:            pulumi.String(strings.ToUpper(spec.Severity.String())),
		Destinations:        pulumi.ToStringArray(spec.Destinations),
		DisplayName:         pulumi.String(locals.AlarmName),
		IsEnabled:           pulumi.Bool(spec.IsEnabled),
		FreeformTags:        pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.Body != "" {
		args.Body = pulumi.String(spec.Body)
	}

	if spec.AlarmSummary != "" {
		args.AlarmSummary = pulumi.String(spec.AlarmSummary)
	}

	if spec.NotificationTitle != "" {
		args.NotificationTitle = pulumi.String(spec.NotificationTitle)
	}

	if spec.PendingDuration != "" {
		args.PendingDuration = pulumi.String(spec.PendingDuration)
	}

	if spec.EvaluationSlackDuration != "" {
		args.EvaluationSlackDuration = pulumi.String(spec.EvaluationSlackDuration)
	}

	if spec.RepeatNotificationDuration != "" {
		args.RepeatNotificationDuration = pulumi.String(spec.RepeatNotificationDuration)
	}

	if spec.MessageFormat != ocialarmv1.OciAlarmSpec_raw {
		args.MessageFormat = pulumi.String(strings.ToUpper(spec.MessageFormat.String()))
	}

	if spec.MetricCompartmentIdInSubtree != nil {
		args.MetricCompartmentIdInSubtree = pulumi.Bool(*spec.MetricCompartmentIdInSubtree)
	}

	if spec.IsNotificationsPerMetricDimensionEnabled != nil {
		args.IsNotificationsPerMetricDimensionEnabled = pulumi.Bool(*spec.IsNotificationsPerMetricDimensionEnabled)
	}

	if spec.ResourceGroup != "" {
		args.ResourceGroup = pulumi.String(spec.ResourceGroup)
	}

	if spec.NotificationVersion != "" {
		args.NotificationVersion = pulumi.String(spec.NotificationVersion)
	}

	if spec.RuleName != "" {
		args.RuleName = pulumi.String(spec.RuleName)
	}

	if len(spec.Overrides) > 0 {
		args.Overrides = buildOverrides(spec.Overrides)
	}

	alarm, err := monitoring.NewAlarm(ctx, locals.AlarmName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create alarm")
	}

	ctx.Export(OpAlarmId, alarm.ID())

	return nil
}

func buildOverrides(overrides []*ocialarmv1.OciAlarmSpec_AlarmOverride) monitoring.AlarmOverrideArray {
	var result monitoring.AlarmOverrideArray

	for _, o := range overrides {
		override := &monitoring.AlarmOverrideArgs{}

		if o.RuleName != "" {
			override.RuleName = pulumi.String(o.RuleName)
		}

		if o.Query != "" {
			override.Query = pulumi.String(o.Query)
		}

		if o.Severity != ocialarmv1.OciAlarmSpec_unspecified {
			override.Severity = pulumi.String(strings.ToUpper(o.Severity.String()))
		}

		if o.Body != "" {
			override.Body = pulumi.String(o.Body)
		}

		if o.PendingDuration != "" {
			override.PendingDuration = pulumi.String(o.PendingDuration)
		}

		result = append(result, override)
	}

	return result
}

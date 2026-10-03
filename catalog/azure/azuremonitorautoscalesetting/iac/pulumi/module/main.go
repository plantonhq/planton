package module

import (
	"github.com/pkg/errors"
	azuremonitorautoscalesettingv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitorautoscalesetting/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/monitoring"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremonitorautoscalesettingv1alpha1.AzureMonitorAutoscaleSettingIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMonitorAutoscaleSetting.Spec

	// The provider requires an explicit value; the platform default is
	// true (a setting exists to scale).
	enabled := true
	if spec.Enabled != nil {
		enabled = *spec.Enabled
	}

	profiles := monitoring.AutoscaleSettingProfileArray{}
	for _, profile := range spec.Profiles {
		profileArgs := &monitoring.AutoscaleSettingProfileArgs{
			Name: pulumi.String(profile.Name),
			Capacity: &monitoring.AutoscaleSettingProfileCapacityArgs{
				Minimum: pulumi.Int(int(*profile.Capacity.Minimum)),
				Maximum: pulumi.Int(int(*profile.Capacity.Maximum)),
				Default: pulumi.Int(int(*profile.Capacity.Default)),
			},
		}

		if len(profile.Rules) > 0 {
			rules := monitoring.AutoscaleSettingProfileRuleArray{}
			for _, rule := range profile.Rules {
				trigger := rule.MetricTrigger
				triggerArgs := &monitoring.AutoscaleSettingProfileRuleMetricTriggerArgs{
					MetricName:            pulumi.String(trigger.MetricName),
					MetricResourceId:      pulumi.String(trigger.MetricResourceId.GetValue()),
					TimeGrain:             pulumi.String(trigger.TimeGrain),
					Statistic:             pulumi.String(trigger.Statistic),
					TimeWindow:            pulumi.String(trigger.TimeWindow),
					TimeAggregation:       pulumi.String(trigger.TimeAggregation),
					Operator:              pulumi.String(trigger.Operator),
					Threshold:             pulumi.Float64(trigger.Threshold),
					DivideByInstanceCount: pulumi.Bool(trigger.DivideByInstanceCount),
				}

				// The provider validates a non-empty namespace -- send the
				// argument only when set (platform metrics imply their own).
				if trigger.MetricNamespace != "" {
					triggerArgs.MetricNamespace = pulumi.String(trigger.MetricNamespace)
				}

				if len(trigger.Dimensions) > 0 {
					dimensions := monitoring.AutoscaleSettingProfileRuleMetricTriggerDimensionArray{}
					for _, dimension := range trigger.Dimensions {
						dimensions = append(dimensions, &monitoring.AutoscaleSettingProfileRuleMetricTriggerDimensionArgs{
							Name:     pulumi.String(dimension.Name),
							Operator: pulumi.String(dimension.Operator),
							Values:   pulumi.ToStringArray(dimension.Values),
						})
					}
					triggerArgs.Dimensions = dimensions
				}

				rules = append(rules, &monitoring.AutoscaleSettingProfileRuleArgs{
					MetricTrigger: triggerArgs,
					ScaleAction: &monitoring.AutoscaleSettingProfileRuleScaleActionArgs{
						Direction: pulumi.String(rule.ScaleAction.Direction),
						Type:      pulumi.String(rule.ScaleAction.Type),
						Value:     pulumi.Int(int(*rule.ScaleAction.Value)),
						Cooldown:  pulumi.String(rule.ScaleAction.Cooldown),
					},
				})
			}
			profileArgs.Rules = rules
		}

		if profile.FixedDate != nil {
			// The platform default is UTC, always sent explicitly.
			timezone := "UTC"
			if profile.FixedDate.Timezone != nil && *profile.FixedDate.Timezone != "" {
				timezone = *profile.FixedDate.Timezone
			}
			profileArgs.FixedDate = &monitoring.AutoscaleSettingProfileFixedDateArgs{
				Timezone: pulumi.String(timezone),
				Start:    pulumi.String(profile.FixedDate.Start),
				End:      pulumi.String(profile.FixedDate.End),
			}
		}

		if profile.Recurrence != nil {
			timezone := "UTC"
			if profile.Recurrence.Timezone != nil && *profile.Recurrence.Timezone != "" {
				timezone = *profile.Recurrence.Timezone
			}
			// ENGINE-SHAPE: the classic SDK flattens the provider's one-item
			// hours/minutes lists to scalar ints; both engines write the same
			// single-element ARM arrays.
			profileArgs.Recurrence = &monitoring.AutoscaleSettingProfileRecurrenceArgs{
				Timezone: pulumi.String(timezone),
				Days:     pulumi.ToStringArray(profile.Recurrence.Days),
				Hours:    pulumi.Int(int(*profile.Recurrence.Hour)),
				Minutes:  pulumi.Int(int(*profile.Recurrence.Minute)),
			}
		}

		profiles = append(profiles, profileArgs)
	}

	autoscaleSettingArgs := &monitoring.AutoscaleSettingArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		TargetResourceId:  pulumi.String(spec.TargetResourceId.GetValue()),
		Enabled:           pulumi.Bool(enabled),
		Profiles:          profiles,
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Predictive autoscale (VM Scale Set targets only). Omitting the block
	// IS the "disabled" state -- the API exposes no Disabled mode.
	if spec.Predictive != nil {
		predictiveArgs := &monitoring.AutoscaleSettingPredictiveArgs{
			ScaleMode: pulumi.String(spec.Predictive.ScaleMode),
		}
		// The provider validates a non-empty ISO 8601 duration -- send the
		// argument only when set.
		if spec.Predictive.LookAheadTime != "" {
			predictiveArgs.LookAheadTime = pulumi.String(spec.Predictive.LookAheadTime)
		}
		autoscaleSettingArgs.Predictive = predictiveArgs
	}

	if spec.Notification != nil {
		notificationArgs := &monitoring.AutoscaleSettingNotificationArgs{}
		if spec.Notification.Email != nil {
			// The provider still ships the retired subscription-administrator /
			// co-administrator flags (ARM rejects them since April 2024); left
			// unset they default to false on the wire, which ARM accepts.
			notificationArgs.Email = &monitoring.AutoscaleSettingNotificationEmailArgs{
				CustomEmails: pulumi.ToStringArray(spec.Notification.Email.CustomEmails),
			}
		}
		if len(spec.Notification.Webhooks) > 0 {
			webhooks := monitoring.AutoscaleSettingNotificationWebhookArray{}
			for _, webhook := range spec.Notification.Webhooks {
				webhooks = append(webhooks, &monitoring.AutoscaleSettingNotificationWebhookArgs{
					ServiceUri: pulumi.String(webhook.ServiceUri),
					Properties: pulumi.ToStringMap(webhook.Properties),
				})
			}
			notificationArgs.Webhooks = webhooks
		}
		autoscaleSettingArgs.Notification = notificationArgs
	}

	createdAutoscaleSetting, err := monitoring.NewAutoscaleSetting(ctx,
		spec.Name,
		autoscaleSettingArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create autoscale setting %s", spec.Name)
	}

	ctx.Export(OpAutoscaleSettingId, createdAutoscaleSetting.ID())
	ctx.Export(OpAutoscaleSettingName, createdAutoscaleSetting.Name)

	return nil
}

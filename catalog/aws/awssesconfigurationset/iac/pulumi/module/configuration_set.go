package module

import (
	"fmt"

	"github.com/pkg/errors"
	awssesconfigurationsetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssesconfigurationset/v1alpha1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sesv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ConfigurationSetResult holds the created configuration set outputs.
type ConfigurationSetResult struct {
	Arn                  pulumi.StringOutput
	ConfigurationSetName pulumi.StringOutput
}

// configurationSet creates the SESv2 configuration set and its per-name event
// destination satellites — mirroring the Terraform module's resource split.
func configurationSet(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
) (*ConfigurationSetResult, error) {
	spec := locals.AwsSesConfigurationSet.Spec

	args := &sesv2.ConfigurationSetArgs{
		// The cloud name comes from metadata.name (the catalog naming basis) --
		// set explicitly so both engines create the same configuration set.
		ConfigurationSetName: pulumi.String(locals.AwsSesConfigurationSet.Metadata.Name),
		ReputationOptions: &sesv2.ConfigurationSetReputationOptionsArgs{
			ReputationMetricsEnabled: pulumi.Bool(spec.ReputationMetricsEnabled),
		},
		SendingOptions: &sesv2.ConfigurationSetSendingOptionsArgs{
			SendingEnabled: pulumi.Bool(locals.SendingEnabled),
		},
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Transport controls: emitted only when the manifest sets delivery_options.
	if spec.DeliveryOptions != nil {
		delivery := &sesv2.ConfigurationSetDeliveryOptionsArgs{}
		tlsPolicy := spec.DeliveryOptions.GetTlsPolicy()
		if tlsPolicy != "" {
			delivery.TlsPolicy = pulumi.StringPtr(tlsPolicy)
		} else {
			delivery.TlsPolicy = pulumi.StringPtr("OPTIONAL")
		}
		if spec.DeliveryOptions.GetMaxDeliverySeconds() > 0 {
			delivery.MaxDeliverySeconds = pulumi.IntPtr(int(spec.DeliveryOptions.GetMaxDeliverySeconds()))
		}
		if spec.DeliveryOptions.GetSendingPoolName() != "" {
			delivery.SendingPoolName = pulumi.StringPtr(spec.DeliveryOptions.GetSendingPoolName())
		}
		args.DeliveryOptions = delivery
	}

	// Suppression override: emitted only when the manifest lists reasons.
	if len(spec.GetSuppressedReasons()) > 0 {
		args.SuppressionOptions = &sesv2.ConfigurationSetSuppressionOptionsArgs{
			SuppressedReasons: pulumi.ToStringArray(spec.GetSuppressedReasons()),
		}
	}

	// Custom open/click tracking domain.
	if spec.TrackingOptions != nil {
		tracking := &sesv2.ConfigurationSetTrackingOptionsArgs{
			CustomRedirectDomain: pulumi.String(spec.TrackingOptions.CustomRedirectDomain),
		}
		if spec.TrackingOptions.GetHttpsPolicy() != "" {
			tracking.HttpsPolicy = pulumi.StringPtr(spec.TrackingOptions.GetHttpsPolicy())
		} else {
			tracking.HttpsPolicy = pulumi.StringPtr("OPTIONAL")
		}
		args.TrackingOptions = tracking
	}

	// Virtual Deliverability Manager overrides.
	if spec.VdmOptions != nil {
		args.VdmOptions = &sesv2.ConfigurationSetVdmOptionsArgs{
			DashboardOptions: &sesv2.ConfigurationSetVdmOptionsDashboardOptionsArgs{
				EngagementMetrics: engagementMetricsString(spec.VdmOptions.GetEngagementMetricsEnabled()),
			},
			GuardianOptions: &sesv2.ConfigurationSetVdmOptionsGuardianOptionsArgs{
				OptimizedSharedDelivery: optimizedSharedDeliveryString(spec.VdmOptions.GetOptimizedSharedDeliveryEnabled()),
			},
		}
	}

	createdSet, err := sesv2.NewConfigurationSet(
		ctx,
		locals.AwsSesConfigurationSet.Metadata.Name,
		args,
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create ses configuration set")
	}

	// Event destinations -- one AWS sub-resource per named entry.
	for name, dest := range locals.EventDestinations {
		if err := createEventDestination(ctx, createdSet, name, dest, provider); err != nil {
			return nil, err
		}
	}

	return &ConfigurationSetResult{
		Arn:                  createdSet.Arn,
		ConfigurationSetName: createdSet.ConfigurationSetName,
	}, nil
}

func createEventDestination(
	ctx *pulumi.Context,
	createdSet *sesv2.ConfigurationSet,
	name string,
	dest *awssesconfigurationsetv1alpha1.AwsSesConfigurationSetEventDestination,
	provider *aws.Provider,
) error {
	// AWS defaults enabled to FALSE; the catalog defaults it to true and always
	// sends the value explicitly so a created destination actually delivers events.
	enabled := true
	if dest.Enabled != nil {
		enabled = dest.GetEnabled()
	}

	eventDestArgs := &sesv2.ConfigurationSetEventDestinationEventDestinationArgs{
		Enabled:            pulumi.Bool(enabled),
		MatchingEventTypes: pulumi.ToStringArray(dest.GetMatchingEventTypes()),
	}

	if dest.CloudWatch != nil {
		var dimensions sesv2.ConfigurationSetEventDestinationEventDestinationCloudWatchDestinationDimensionConfigurationArray
		for _, dim := range dest.CloudWatch.Dimensions {
			dimensions = append(dimensions, &sesv2.ConfigurationSetEventDestinationEventDestinationCloudWatchDestinationDimensionConfigurationArgs{
				DimensionName:         pulumi.String(dim.Name),
				DimensionValueSource:  pulumi.String(dim.ValueSource),
				DefaultDimensionValue: pulumi.String(dim.DefaultValue),
			})
		}
		eventDestArgs.CloudWatchDestination = &sesv2.ConfigurationSetEventDestinationEventDestinationCloudWatchDestinationArgs{
			DimensionConfigurations: dimensions,
		}
	}

	if dest.EventBus != nil && dest.EventBus.GetValue() != "" {
		eventDestArgs.EventBridgeDestination = &sesv2.ConfigurationSetEventDestinationEventDestinationEventBridgeDestinationArgs{
			EventBusArn: pulumi.String(dest.EventBus.GetValue()),
		}
	}

	if dest.Firehose != nil {
		eventDestArgs.KinesisFirehoseDestination = &sesv2.ConfigurationSetEventDestinationEventDestinationKinesisFirehoseDestinationArgs{
			DeliveryStreamArn: pulumi.String(dest.Firehose.DeliveryStream.GetValue()),
			IamRoleArn:        pulumi.String(dest.Firehose.IamRole.GetValue()),
		}
	}

	if dest.SnsTopic != nil && dest.SnsTopic.GetValue() != "" {
		eventDestArgs.SnsDestination = &sesv2.ConfigurationSetEventDestinationEventDestinationSnsDestinationArgs{
			TopicArn: pulumi.String(dest.SnsTopic.GetValue()),
		}
	}

	if dest.PinpointApplicationArn != "" {
		eventDestArgs.PinpointDestination = &sesv2.ConfigurationSetEventDestinationEventDestinationPinpointDestinationArgs{
			ApplicationArn: pulumi.String(dest.PinpointApplicationArn),
		}
	}

	_, err := sesv2.NewConfigurationSetEventDestination(
		ctx,
		fmt.Sprintf("event-destination-%s", name),
		&sesv2.ConfigurationSetEventDestinationArgs{
			ConfigurationSetName: createdSet.ConfigurationSetName,
			EventDestinationName: pulumi.String(name),
			EventDestination:     eventDestArgs,
		},
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{createdSet}),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create event destination %s", name)
	}
	return nil
}

func engagementMetricsString(enabled bool) pulumi.StringPtrInput {
	if enabled {
		return pulumi.StringPtr("ENABLED")
	}
	return pulumi.StringPtr("DISABLED")
}

func optimizedSharedDeliveryString(enabled bool) pulumi.StringPtrInput {
	if enabled {
		return pulumi.StringPtr("ENABLED")
	}
	return pulumi.StringPtr("DISABLED")
}

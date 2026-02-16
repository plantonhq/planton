package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureeventhubnamespacev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureeventhubnamespace/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventhub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureeventhubnamespacev1.AzureEventHubNamespaceStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input.
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventHubNamespace.Spec

	// Build the Event Hub namespace arguments.
	namespaceArgs := &eventhub.EventHubNamespaceArgs{
		Name:                       pulumi.String(spec.Name),
		Location:                   pulumi.String(spec.Region),
		ResourceGroupName:          pulumi.String(locals.ResourceGroupName),
		Sku:                        pulumi.String(spec.GetSku()),
		MinimumTlsVersion:          pulumi.StringPtr(spec.GetMinimumTlsVersion()),
		PublicNetworkAccessEnabled: pulumi.BoolPtr(spec.GetPublicNetworkAccessEnabled()),
		Tags:                       pulumi.ToStringMap(locals.AzureTags),
	}

	// Optional fields.
	if spec.Capacity != nil {
		namespaceArgs.Capacity = pulumi.IntPtr(int(spec.GetCapacity()))
	}
	if spec.AutoInflateEnabled != nil {
		namespaceArgs.AutoInflateEnabled = pulumi.BoolPtr(spec.GetAutoInflateEnabled())
	}
	if spec.MaximumThroughputUnits != nil {
		namespaceArgs.MaximumThroughputUnits = pulumi.IntPtr(int(spec.GetMaximumThroughputUnits()))
	}
	// Note: ZoneRedundant is not exposed in the Pulumi Azure classic SDK v6.
	// It is available in the Terraform module (azurerm_eventhub_namespace.zone_redundant).
	// When Pulumi SDK adds support, uncomment:
	// if spec.ZoneRedundant != nil {
	//     namespaceArgs.ZoneRedundant = pulumi.BoolPtr(spec.GetZoneRedundant())
	// }

	// Create the Event Hub namespace.
	namespace, err := eventhub.NewEventHubNamespace(ctx,
		spec.Name,
		namespaceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Event Hub namespace %s", spec.Name)
	}

	// Create event hubs and their consumer groups.
	eventHubIDs := pulumi.StringMap{}
	for _, eh := range spec.EventHubs {
		eventHubArgs := &eventhub.EventHubArgs{
			Name:             pulumi.String(eh.Name),
			NamespaceId:      namespace.ID(),
			PartitionCount:   pulumi.Int(int(eh.PartitionCount)),
			MessageRetention: pulumi.Int(int(eh.GetMessageRetention())),
		}

		createdEventHub, err := eventhub.NewEventHub(ctx,
			fmt.Sprintf("%s-%s", spec.Name, eh.Name),
			eventHubArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return errors.Wrapf(err, "failed to create Event Hub %s", eh.Name)
		}

		eventHubIDs[eh.Name] = createdEventHub.ID().ToStringOutput()

		// Create consumer groups for this event hub.
		// Note: The Pulumi Azure classic SDK uses the legacy pattern with
		// NamespaceName, EventhubName, and ResourceGroupName instead of NamespaceId.
		for _, cg := range eh.ConsumerGroups {
			consumerGroupArgs := &eventhub.ConsumerGroupArgs{
				Name:              pulumi.String(cg.Name),
				NamespaceName:     pulumi.String(spec.Name),
				EventhubName:      pulumi.String(eh.Name),
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			}

			if cg.UserMetadata != nil {
				consumerGroupArgs.UserMetadata = pulumi.StringPtr(cg.GetUserMetadata())
			}

			_, err := eventhub.NewConsumerGroup(ctx,
				fmt.Sprintf("%s-%s-%s", spec.Name, eh.Name, cg.Name),
				consumerGroupArgs,
				pulumi.Provider(azureProvider),
				pulumi.DependsOn([]pulumi.Resource{createdEventHub}))
			if err != nil {
				return errors.Wrapf(err, "failed to create consumer group %s for Event Hub %s", cg.Name, eh.Name)
			}
		}
	}

	// Export stack outputs.
	ctx.Export(OpNamespaceId, namespace.ID())
	ctx.Export(OpNamespaceName, namespace.Name)
	ctx.Export(OpPrimaryConnectionString, namespace.DefaultPrimaryConnectionString)
	ctx.Export(OpPrimaryKey, namespace.DefaultPrimaryKey)
	ctx.Export(OpEventHubIds, eventHubIDs)

	return nil
}

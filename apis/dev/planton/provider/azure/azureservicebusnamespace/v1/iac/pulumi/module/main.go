package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureservicebusnamespacev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azureservicebusnamespace/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/servicebus"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureservicebusnamespacev1.AzureServiceBusNamespaceStackInput) error {
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

	spec := locals.AzureServiceBusNamespace.Spec

	// Build the Service Bus namespace arguments.
	namespaceArgs := &servicebus.NamespaceArgs{
		Name:                       pulumi.String(spec.Name),
		Location:                   pulumi.String(spec.Region),
		ResourceGroupName:          pulumi.String(locals.ResourceGroupName),
		Sku:                        pulumi.String(spec.GetSku()),
		MinimumTlsVersion:          pulumi.StringPtr(spec.GetMinimumTlsVersion()),
		PublicNetworkAccessEnabled: pulumi.BoolPtr(spec.GetPublicNetworkAccessEnabled()),
		Tags:                       pulumi.ToStringMap(locals.AzureTags),
	}

	// Premium-only fields: capacity, premium_messaging_partitions, zone_redundant.
	if spec.Capacity != nil {
		namespaceArgs.Capacity = pulumi.IntPtr(int(spec.GetCapacity()))
	}
	if spec.PremiumMessagingPartitions != nil {
		namespaceArgs.PremiumMessagingPartitions = pulumi.IntPtr(int(spec.GetPremiumMessagingPartitions()))
	}
	// Note: ZoneRedundant is not exposed in the Pulumi Azure classic SDK v6.
	// It is available in the Terraform module (azurerm_servicebus_namespace.zone_redundant).
	// When Pulumi SDK adds support, uncomment:
	// if spec.ZoneRedundant != nil {
	//     namespaceArgs.ZoneRedundant = pulumi.BoolPtr(spec.GetZoneRedundant())
	// }

	// Create the Service Bus namespace.
	namespace, err := servicebus.NewNamespace(ctx,
		spec.Name,
		namespaceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Service Bus namespace %s", spec.Name)
	}

	// Create queues.
	queueIDs := pulumi.StringMap{}
	for _, q := range spec.Queues {
		queueArgs := &servicebus.QueueArgs{
			Name:        pulumi.String(q.Name),
			NamespaceId: namespace.ID(),
		}

		if q.MaxSizeInMegabytes != nil {
			queueArgs.MaxSizeInMegabytes = pulumi.IntPtr(int(q.GetMaxSizeInMegabytes()))
		}
		if q.PartitioningEnabled != nil {
			queueArgs.PartitioningEnabled = pulumi.BoolPtr(q.GetPartitioningEnabled())
		}
		if q.DefaultMessageTtl != nil {
			queueArgs.DefaultMessageTtl = pulumi.StringPtr(q.GetDefaultMessageTtl())
		}
		if q.LockDuration != nil {
			queueArgs.LockDuration = pulumi.StringPtr(q.GetLockDuration())
		}
		if q.MaxDeliveryCount != nil {
			queueArgs.MaxDeliveryCount = pulumi.IntPtr(int(q.GetMaxDeliveryCount()))
		}
		if q.RequiresDuplicateDetection != nil {
			queueArgs.RequiresDuplicateDetection = pulumi.BoolPtr(q.GetRequiresDuplicateDetection())
		}
		if q.RequiresSession != nil {
			queueArgs.RequiresSession = pulumi.BoolPtr(q.GetRequiresSession())
		}
		if q.DeadLetteringOnMessageExpiration != nil {
			queueArgs.DeadLetteringOnMessageExpiration = pulumi.BoolPtr(q.GetDeadLetteringOnMessageExpiration())
		}
		if q.ForwardTo != nil {
			queueArgs.ForwardTo = pulumi.StringPtr(q.GetForwardTo())
		}
		if q.ForwardDeadLetteredMessagesTo != nil {
			queueArgs.ForwardDeadLetteredMessagesTo = pulumi.StringPtr(q.GetForwardDeadLetteredMessagesTo())
		}

		queue, err := servicebus.NewQueue(ctx,
			fmt.Sprintf("%s-%s", spec.Name, q.Name),
			queueArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return errors.Wrapf(err, "failed to create Service Bus queue %s", q.Name)
		}

		queueIDs[q.Name] = queue.ID().ToStringOutput()
	}

	// Create topics.
	topicIDs := pulumi.StringMap{}
	for _, t := range spec.Topics {
		topicArgs := &servicebus.TopicArgs{
			Name:        pulumi.String(t.Name),
			NamespaceId: namespace.ID(),
		}

		if t.MaxSizeInMegabytes != nil {
			topicArgs.MaxSizeInMegabytes = pulumi.IntPtr(int(t.GetMaxSizeInMegabytes()))
		}
		if t.PartitioningEnabled != nil {
			topicArgs.PartitioningEnabled = pulumi.BoolPtr(t.GetPartitioningEnabled())
		}
		if t.DefaultMessageTtl != nil {
			topicArgs.DefaultMessageTtl = pulumi.StringPtr(t.GetDefaultMessageTtl())
		}
		if t.RequiresDuplicateDetection != nil {
			topicArgs.RequiresDuplicateDetection = pulumi.BoolPtr(t.GetRequiresDuplicateDetection())
		}
		if t.SupportOrdering != nil {
			topicArgs.SupportOrdering = pulumi.BoolPtr(t.GetSupportOrdering())
		}

		topic, err := servicebus.NewTopic(ctx,
			fmt.Sprintf("%s-%s", spec.Name, t.Name),
			topicArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{namespace}))
		if err != nil {
			return errors.Wrapf(err, "failed to create Service Bus topic %s", t.Name)
		}

		topicIDs[t.Name] = topic.ID().ToStringOutput()
	}

	// Export stack outputs.
	ctx.Export(OpNamespaceId, namespace.ID())
	ctx.Export(OpNamespaceName, namespace.Name)
	ctx.Export(OpEndpoint, namespace.Endpoint)
	ctx.Export(OpPrimaryConnectionString, namespace.DefaultPrimaryConnectionString)
	ctx.Export(OpPrimaryKey, namespace.DefaultPrimaryKey)
	ctx.Export(OpQueueIds, queueIDs)
	ctx.Export(OpTopicIds, topicIDs)

	return nil
}

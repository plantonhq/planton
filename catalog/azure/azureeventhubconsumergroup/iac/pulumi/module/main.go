package module

import (
	"github.com/pkg/errors"
	azureeventhubconsumergroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubconsumergroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventhub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureeventhubconsumergroupv1alpha1.AzureEventHubConsumerGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventHubConsumerGroup.Spec

	// azurerm still addresses consumer groups by discrete names (resource
	// group, namespace, event hub) rather than the parent's ARM id --
	// derive them from the spec's single parent reference so the spec
	// stays on the ARM-id grain.
	resourceGroupName, namespaceName, eventHubName, err := parseEventHubId(locals.EventHubId)
	if err != nil {
		return err
	}
	locals.ResourceGroupName = resourceGroupName
	locals.NamespaceName = namespaceName
	locals.EventHubName = eventHubName

	// The consumer group: one application's independent read cursor over
	// the hub's partitions. Tier limits are enforced by Azure at apply
	// time (BASIC hubs allow no additional groups beyond the
	// service-created $Default; STANDARD allows 20 per hub), so quota
	// errors surface verbatim from Azure.
	consumerGroupArgs := &eventhub.EventHubConsumerGroupArgs{
		// ForceNew: the group is its name -- renaming replaces it and
		// resets its consumers' stored offsets.
		Name:              pulumi.String(spec.ConsumerGroupName),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		NamespaceName:     pulumi.String(locals.NamespaceName),
		EventhubName:      pulumi.String(locals.EventHubName),
	}

	if spec.UserMetadata != nil {
		consumerGroupArgs.UserMetadata = pulumi.StringPtr(spec.GetUserMetadata())
	}

	createdConsumerGroup, err := eventhub.NewEventHubConsumerGroup(ctx,
		spec.ConsumerGroupName,
		consumerGroupArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Event Hub consumer group %s", spec.ConsumerGroupName)
	}

	// Export outputs: what consumer applications pass to their SDK
	// client alongside the hub name.
	ctx.Export(OpConsumerGroupId, createdConsumerGroup.ID())
	ctx.Export(OpConsumerGroupName, createdConsumerGroup.Name)

	return nil
}

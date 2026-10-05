package module

import (
	"github.com/pkg/errors"
	azureservicebustopicv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebustopic/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/servicebus"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureservicebustopicv1alpha1.AzureServiceBusTopicIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureServiceBusTopic.Spec

	// The namespace name, parsed from the resolved namespace ARM id for the
	// output -- consumers frequently need the namespace/topic name
	// pair, and this saves them a second reference.
	namespaceName, err := parseNamespaceName(locals.NamespaceId)
	if err != nil {
		return err
	}
	locals.NamespaceName = namespaceName

	// The topic, addressed by the parent namespace's ARM id (azurerm's v4
	// child-addressing grain). Topics require STANDARD or PREMIUM (BASIC is
	// queue-only -- Azure rejects the create). Premium-tier contracts the
	// provider enforces at apply time are documented on the spec fields;
	// the module passes the dials through unchanged so those contracts
	// surface verbatim from Azure.
	topicArgs := &servicebus.TopicArgs{
		// ForceNew: renaming replaces the topic and every subscription
		// under it.
		Name:        pulumi.String(spec.TopicName),
		NamespaceId: pulumi.String(locals.NamespaceId),
		Status:      pulumi.String(statusStrings[spec.Status]),
	}

	// Capacity dials. Unset sizes let Azure default for the namespace's
	// tier.
	if spec.MaxSizeInMegabytes != nil {
		topicArgs.MaxSizeInMegabytes = pulumi.IntPtr(int(spec.GetMaxSizeInMegabytes()))
	}
	if spec.MaxMessageSizeInKilobytes != nil {
		topicArgs.MaxMessageSizeInKilobytes = pulumi.IntPtr(int(spec.GetMaxMessageSizeInKilobytes()))
	}

	// ForceNew pair: the storage layout and dedup model are fixed at
	// creation.
	if spec.PartitioningEnabled != nil {
		topicArgs.PartitioningEnabled = pulumi.BoolPtr(spec.GetPartitioningEnabled())
	}
	if spec.RequiresDuplicateDetection != nil {
		topicArgs.RequiresDuplicateDetection = pulumi.BoolPtr(spec.GetRequiresDuplicateDetection())
	}

	// Lifecycle dials -- unset leaves Azure's defaults in place (TTL
	// unbounded, dedup window PT10M).
	if spec.DefaultMessageTtl != nil {
		topicArgs.DefaultMessageTtl = pulumi.StringPtr(spec.GetDefaultMessageTtl())
	}
	if spec.DuplicateDetectionHistoryTimeWindow != nil {
		topicArgs.DuplicateDetectionHistoryTimeWindow = pulumi.StringPtr(spec.GetDuplicateDetectionHistoryTimeWindow())
	}
	if spec.AutoDeleteOnIdle != nil {
		topicArgs.AutoDeleteOnIdle = pulumi.StringPtr(spec.GetAutoDeleteOnIdle())
	}

	// Presence-guarded to Azure's default (true): IaC inputs built from a
	// manifest materialize proto defaults, but direct paths do not.
	topicArgs.BatchedOperationsEnabled = pulumi.Bool(presenceGuardedBool(spec.BatchedOperationsEnabled, true))

	if spec.ExpressEnabled != nil {
		topicArgs.ExpressEnabled = pulumi.BoolPtr(spec.GetExpressEnabled())
	}

	// Publish-order preservation -- pair with session-aware subscriptions
	// for strictly-ordered publish-subscribe.
	if spec.SupportOrdering != nil {
		topicArgs.SupportOrdering = pulumi.BoolPtr(spec.GetSupportOrdering())
	}

	createdTopic, err := servicebus.NewTopic(ctx,
		spec.TopicName,
		topicArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Service Bus topic %s", spec.TopicName)
	}

	// Export outputs. No connection string on purpose: credentials
	// are minted by AzureServiceBusAuthorizationRule or granted keyless via
	// Entra data-plane roles on topic_id.
	ctx.Export(OpTopicId, createdTopic.ID())
	ctx.Export(OpTopicName, createdTopic.Name)
	ctx.Export(OpNamespaceName, pulumi.String(locals.NamespaceName))

	return nil
}

// presenceGuardedBool returns the field's value when set and the proto
// default otherwise -- default materialization is middleware behavior, not a
// wire guarantee.
func presenceGuardedBool(field *bool, defaultValue bool) bool {
	if field == nil {
		return defaultValue
	}
	return *field
}

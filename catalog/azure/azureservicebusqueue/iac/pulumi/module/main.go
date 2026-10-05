package module

import (
	"github.com/pkg/errors"
	azureservicebusqueuev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureservicebusqueue/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/servicebus"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureservicebusqueuev1alpha1.AzureServiceBusQueueIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureServiceBusQueue.Spec

	// The namespace name, parsed from the resolved namespace ARM id for the
	// output -- consumers frequently need the namespace/queue name
	// pair, and this saves them a second reference.
	namespaceName, err := parseNamespaceName(locals.NamespaceId)
	if err != nil {
		return err
	}
	locals.NamespaceName = namespaceName

	// The queue, addressed by the parent namespace's ARM id (azurerm's v4
	// child-addressing grain). Premium-tier contracts the provider enforces
	// at apply time -- express rejected on Premium, partitioning must match
	// the namespace's partition layout, large messages Premium-only -- are
	// documented on the spec fields; the module passes the dials through
	// unchanged so those contracts surface verbatim from Azure.
	queueArgs := &servicebus.QueueArgs{
		// ForceNew: renaming replaces the queue and drops any messages in it.
		Name:        pulumi.String(spec.QueueName),
		NamespaceId: pulumi.String(locals.NamespaceId),
		Status:      pulumi.String(statusStrings[spec.Status]),
	}

	// Capacity dials. Unset sizes let Azure default for the namespace's
	// tier (1024 MB multi-tenant, 81920 MB premium).
	if spec.MaxSizeInMegabytes != nil {
		queueArgs.MaxSizeInMegabytes = pulumi.IntPtr(int(spec.GetMaxSizeInMegabytes()))
	}
	if spec.MaxMessageSizeInKilobytes != nil {
		queueArgs.MaxMessageSizeInKilobytes = pulumi.IntPtr(int(spec.GetMaxMessageSizeInKilobytes()))
	}

	// ForceNew trio: the storage layout and dedup/session models are fixed
	// at creation.
	if spec.PartitioningEnabled != nil {
		queueArgs.PartitioningEnabled = pulumi.BoolPtr(spec.GetPartitioningEnabled())
	}
	if spec.RequiresDuplicateDetection != nil {
		queueArgs.RequiresDuplicateDetection = pulumi.BoolPtr(spec.GetRequiresDuplicateDetection())
	}
	if spec.RequiresSession != nil {
		queueArgs.RequiresSession = pulumi.BoolPtr(spec.GetRequiresSession())
	}

	// Lifecycle dials -- unset leaves Azure's defaults in place (TTL
	// unbounded, dedup window PT10M, lock PT1M, 10 deliveries).
	if spec.DefaultMessageTtl != nil {
		queueArgs.DefaultMessageTtl = pulumi.StringPtr(spec.GetDefaultMessageTtl())
	}
	if spec.DuplicateDetectionHistoryTimeWindow != nil {
		queueArgs.DuplicateDetectionHistoryTimeWindow = pulumi.StringPtr(spec.GetDuplicateDetectionHistoryTimeWindow())
	}
	if spec.LockDuration != nil {
		queueArgs.LockDuration = pulumi.StringPtr(spec.GetLockDuration())
	}
	if spec.MaxDeliveryCount != nil {
		queueArgs.MaxDeliveryCount = pulumi.IntPtr(int(spec.GetMaxDeliveryCount()))
	}
	if spec.DeadLetteringOnMessageExpiration != nil {
		queueArgs.DeadLetteringOnMessageExpiration = pulumi.BoolPtr(spec.GetDeadLetteringOnMessageExpiration())
	}
	if spec.AutoDeleteOnIdle != nil {
		queueArgs.AutoDeleteOnIdle = pulumi.StringPtr(spec.GetAutoDeleteOnIdle())
	}

	// Presence-guarded to Azure's default (true): IaC inputs built from a
	// manifest materialize proto defaults, but direct paths do not.
	queueArgs.BatchedOperationsEnabled = pulumi.Bool(presenceGuardedBool(spec.BatchedOperationsEnabled, true))

	if spec.ExpressEnabled != nil {
		queueArgs.ExpressEnabled = pulumi.BoolPtr(spec.GetExpressEnabled())
	}

	// Routing chains: targets are entity NAMES in the same namespace (not
	// ARM ids) -- Azure's own addressing for auto-forwarding. References
	// resolve to the target's queue_name/topic_name output before the
	// module runs. The target must exist first; compose with depends-on
	// ordering in charts.
	if spec.ForwardTo.GetValue() != "" {
		queueArgs.ForwardTo = pulumi.StringPtr(spec.ForwardTo.GetValue())
	}
	if spec.ForwardDeadLetteredMessagesTo.GetValue() != "" {
		queueArgs.ForwardDeadLetteredMessagesTo = pulumi.StringPtr(spec.ForwardDeadLetteredMessagesTo.GetValue())
	}

	createdQueue, err := servicebus.NewQueue(ctx,
		spec.QueueName,
		queueArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Service Bus queue %s", spec.QueueName)
	}

	// Export outputs. No connection string on purpose: credentials
	// are minted by AzureServiceBusAuthorizationRule or granted keyless via
	// Entra data-plane roles on queue_id.
	ctx.Export(OpQueueId, createdQueue.ID())
	ctx.Export(OpQueueName, createdQueue.Name)
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

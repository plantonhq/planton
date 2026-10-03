package module

import (
	"github.com/pkg/errors"
	azureeventhubv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhub/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/eventhub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureeventhubv1alpha1.AzureEventHubIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureEventHub.Spec

	// One partitioned, replayable event stream inside an Event Hubs
	// namespace. Consumer groups, SAS rules, and data-plane role
	// assignments are first-class kinds that reference this hub's ARM id
	// -- nothing is bundled here.
	hubArgs := &eventhub.EventHubArgs{
		// ForceNew: renaming replaces the hub and its retained events. The
		// name is also the Kafka topic name on the namespace's Kafka
		// endpoint.
		Name: pulumi.String(spec.EventHubName),
		// ForceNew: the hub cannot move between namespaces.
		NamespaceId: pulumi.String(spec.NamespaceId.GetValue()),
		// The unit of parallelism and ordering. Azure enforces the tier
		// caps (32 shared, 1024 PREMIUM/dedicated) and the never-decrease /
		// increase-only-on-PREMIUM-or-dedicated contracts at apply time --
		// they depend on the parent namespace's tier, which this module
		// cannot see.
		PartitionCount: pulumi.Int(int(spec.PartitionCount)),
		// The administrative gate: Active, Disabled, or SendDisabled
		// (receive-only drain).
		Status: pulumi.String(statusStrings[spec.Status]),
	}

	// Simple retention in days. Exactly one of message_retention and
	// retention_description is set (spec-enforced XOR); Azure caps the
	// window by tier (1/7/90 days) at apply time.
	if spec.MessageRetention != nil {
		hubArgs.MessageRetention = pulumi.IntPtr(int(spec.GetMessageRetention()))
	}

	// Rich retention: hour-granular windows and Kafka-style compaction.
	// The spec pairs the hour field to the policy (DELETE takes
	// retention_time_in_hours, COMPACT takes the tombstone window), so
	// each variant sends only its own field -- Azure silently ignores the
	// mismatched one, which the spec rejects up front instead.
	if spec.RetentionDescription != nil {
		retentionArgs := &eventhub.EventHubRetentionDescriptionArgs{
			// ForceNew: the cleanup policy is fixed at creation.
			CleanupPolicy: pulumi.String(cleanupPolicyStrings[spec.RetentionDescription.CleanupPolicy]),
		}
		if spec.RetentionDescription.RetentionTimeInHours != nil {
			retentionArgs.RetentionTimeInHours = pulumi.IntPtr(int(spec.RetentionDescription.GetRetentionTimeInHours()))
		}
		if spec.RetentionDescription.TombstoneRetentionTimeInHours != nil {
			retentionArgs.TombstoneRetentionTimeInHours = pulumi.IntPtr(int(spec.RetentionDescription.GetTombstoneRetentionTimeInHours()))
		}
		hubArgs.RetentionDescription = retentionArgs
	}

	// Capture: the built-in streaming-to-batch bridge -- every event is
	// archived to Blob Storage in Avro on a size-or-interval cadence, with
	// no consumer application to run.
	if spec.CaptureDescription != nil {
		capture := spec.CaptureDescription
		destinationArgs := &eventhub.EventHubCaptureDescriptionDestinationArgs{
			// Azure's destination name accepts exactly one value (Blob
			// Storage; the Data Lake variant retired with Gen1) -- a
			// constant, not configuration, so the module sends it
			// unconditionally.
			Name:              pulumi.String("EventHubArchive.AzureBlockBlob"),
			ArchiveNameFormat: pulumi.String(capture.Destination.ArchiveNameFormat),
			BlobContainerName: pulumi.String(capture.Destination.BlobContainerName.GetValue()),
			StorageAccountId:  pulumi.String(capture.Destination.StorageAccountId.GetValue()),
			// "StorageSAS" (Azure's default) means service-managed SAS --
			// the provider sends no identity for it. The identity paths are
			// keyless: grant the chosen identity Storage Blob Data
			// Contributor on the account and attach it via the namespace's
			// identity block.
			StorageAuthenticationType: pulumi.String(captureAuthStrings[capture.Destination.StorageAuthenticationType]),
		}
		if capture.Destination.StorageAuthenticationId.GetValue() != "" {
			destinationArgs.StorageAuthenticationId = pulumi.String(capture.Destination.StorageAuthenticationId.GetValue())
		}

		captureArgs := &eventhub.EventHubCaptureDescriptionArgs{
			Enabled:     pulumi.Bool(capture.GetEnabled()),
			Encoding:    pulumi.String(captureEncodingStrings[capture.Encoding]),
			Destination: destinationArgs,
		}
		if capture.IntervalInSeconds != nil {
			captureArgs.IntervalInSeconds = pulumi.IntPtr(int(capture.GetIntervalInSeconds()))
		}
		if capture.SizeLimitInBytes != nil {
			captureArgs.SizeLimitInBytes = pulumi.IntPtr(int(capture.GetSizeLimitInBytes()))
		}
		if capture.SkipEmptyArchives != nil {
			captureArgs.SkipEmptyArchives = pulumi.BoolPtr(capture.GetSkipEmptyArchives())
		}
		hubArgs.CaptureDescription = captureArgs
	}

	createdHub, err := eventhub.NewEventHub(ctx,
		spec.EventHubName,
		hubArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create event hub %s", spec.EventHubName)
	}

	// Export outputs.
	ctx.Export(OpEventHubId, createdHub.ID())
	ctx.Export(OpEventHubName, createdHub.Name)
	ctx.Export(OpPartitionIds, createdHub.PartitionIds)

	return nil
}

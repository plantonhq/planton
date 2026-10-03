package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	azurestoragequeuev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestoragequeue/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurestoragequeuev1alpha1.AzureStorageQueueIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureStorageQueue.Spec

	// The account name, parsed from the resolved account ARM ID for the
	// output -- consumers frequently need the account/queue name
	// pair, and this saves them a second reference. The id must END
	// with /storageAccounts/{name} (matching the Terraform module's
	// anchored regex), so a malformed or over-long id fails loudly here
	// instead of computing a wrong name.
	accountIdParts := strings.Split(locals.StorageAccountId, "/storageAccounts/")
	if len(accountIdParts) != 2 || accountIdParts[1] == "" || strings.Contains(accountIdParts[1], "/") {
		return errors.Errorf("storage_account_id %q is not a storage-account ARM id", locals.StorageAccountId)
	}
	storageAccountName := accountIdParts[1]

	// The queue is addressed by the parent account's ARM ID (the
	// control-plane path -- the account-name form is the provider's
	// legacy data-plane path, removed in azurerm v5). Queues carry no
	// Azure tags: ARM does not support tags on queueServices/queues.
	queueArgs := &storage.QueueArgs{
		Name:             pulumi.String(spec.QueueName),
		StorageAccountId: pulumi.String(locals.StorageAccountId),
	}

	// Queue metadata is NOT Azure tags -- free-form key/value pairs on
	// the queue itself, visible to anyone who can read queue properties.
	if len(spec.Metadata) > 0 {
		queueArgs.Metadata = pulumi.ToStringMap(spec.Metadata)
	}

	createdQueue, err := storage.NewQueue(ctx,
		fmt.Sprintf("%s-%s", storageAccountName, spec.QueueName),
		queueArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create storage queue %s", spec.QueueName)
	}

	// Export outputs. The queue's data-plane URL is deliberately
	// NOT exported -- compose client URLs from the ACCOUNT's
	// primary_queue_endpoint output + queue_name (only the account knows
	// its real endpoint; partitioned-DNS accounts differ).
	ctx.Export(OpQueueId, createdQueue.ID())
	ctx.Export(OpQueueName, createdQueue.Name)
	ctx.Export(OpStorageAccountName, pulumi.String(storageAccountName))

	return nil
}

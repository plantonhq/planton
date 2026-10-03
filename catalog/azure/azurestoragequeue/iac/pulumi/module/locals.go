package module

import (
	azurestoragequeuev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestoragequeue/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureStorageQueue *azurestoragequeuev1alpha1.AzureStorageQueue
	StorageAccountId  string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurestoragequeuev1alpha1.AzureStorageQueueIacInput) *Locals {
	locals := &Locals{}

	locals.AzureStorageQueue = iacInput.Target
	locals.StorageAccountId = iacInput.Target.Spec.StorageAccountId.GetValue()

	// No Azure tags: ARM does not support tags on queueServices/queues,
	// so the platform's identity tags live on the parent account.

	return locals
}

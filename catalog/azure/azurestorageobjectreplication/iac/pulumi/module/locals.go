package module

import (
	azurestorageobjectreplicationv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestorageobjectreplication/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureStorageObjectReplication *azurestorageobjectreplicationv1alpha1.AzureStorageObjectReplication
	SourceStorageAccountId        string
	DestinationStorageAccountId   string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurestorageobjectreplicationv1alpha1.AzureStorageObjectReplicationIacInput) *Locals {
	locals := &Locals{}

	locals.AzureStorageObjectReplication = iacInput.Target
	locals.SourceStorageAccountId = iacInput.Target.Spec.SourceStorageAccountId.GetValue()
	locals.DestinationStorageAccountId = iacInput.Target.Spec.DestinationStorageAccountId.GetValue()

	// No Azure tags: ARM does not support tags on
	// objectReplicationPolicies, so the platform's identity tags live on
	// the two accounts.

	return locals
}

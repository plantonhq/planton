package module

import (
	azurecontainerappenvironmentstoragev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironmentstorage/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureContainerAppEnvironmentStorage *azurecontainerappenvironmentstoragev1alpha1.AzureContainerAppEnvironmentStorage
}

// accessModeStrings maps the access-mode enum to ARM's values.
var accessModeStrings = map[azurecontainerappenvironmentstoragev1alpha1.AzureContainerAppEnvironmentStorageAccessMode]string{
	azurecontainerappenvironmentstoragev1alpha1.AzureContainerAppEnvironmentStorageAccessMode_READ_ONLY:  "ReadOnly",
	azurecontainerappenvironmentstoragev1alpha1.AzureContainerAppEnvironmentStorageAccessMode_READ_WRITE: "ReadWrite",
}

// The storage registration carries no tags (ARM does not support them on
// managedEnvironments/storages), so locals stay minimal.
func initializeLocals(ctx *pulumi.Context, iacInput *azurecontainerappenvironmentstoragev1alpha1.AzureContainerAppEnvironmentStorageIacInput) *Locals {
	return &Locals{
		AzureContainerAppEnvironmentStorage: iacInput.Target,
	}
}

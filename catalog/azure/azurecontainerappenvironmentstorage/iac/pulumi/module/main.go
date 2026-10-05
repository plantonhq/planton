package module

import (
	"github.com/pkg/errors"
	azurecontainerappenvironmentstoragev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironmentstorage/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecontainerappenvironmentstoragev1alpha1.AzureContainerAppEnvironmentStorageIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerAppEnvironmentStorage.Spec

	// The storage registration makes an Azure Files share mountable by
	// apps and jobs in the environment. Everything except the SMB
	// access_key is ForceNew -- key rotation is the one in-place update.
	storageArgs := &containerapp.EnvironmentStorageArgs{
		Name:                      pulumi.String(spec.StorageName),
		ContainerAppEnvironmentId: pulumi.String(spec.ContainerAppEnvironmentId.GetValue()),
		ShareName:                 pulumi.String(spec.ShareName.GetValue()),
		AccessMode:                pulumi.String(accessModeStrings[spec.AccessMode]),
	}

	// Exactly one protocol (spec-enforced): SMB addresses the share by
	// account name + access key; NFS addresses the account's file
	// endpoint and requires a VNet-injected environment.
	if spec.NfsServerUrl != "" {
		storageArgs.NfsServerUrl = pulumi.String(spec.NfsServerUrl)
	} else {
		storageArgs.AccountName = pulumi.String(spec.AccountName.GetValue())
		storageArgs.AccessKey = pulumi.String(spec.AccessKey.GetValue())
	}

	createdStorage, err := containerapp.NewEnvironmentStorage(ctx,
		spec.StorageName,
		storageArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Container App Environment storage %s", spec.StorageName)
	}

	// Export outputs. storage_name is the composition seam app and
	// job volumes reference.
	ctx.Export(OpStorageId, createdStorage.ID())
	ctx.Export(OpStorageName, createdStorage.Name)

	return nil
}

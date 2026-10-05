package module

import (
	"github.com/pkg/errors"
	azuremachinelearningdatastorev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningdatastore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/machinelearning"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremachinelearningdatastorev1alpha1.AzureMachineLearningDatastoreIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMachineLearningDatastore.Spec

	// Create the Machine Learning datastore -- the saved connection
	// that tells the workspace where data lives. Exactly one variant
	// block is set (spec CEL) and selects which provider resource is
	// created; all three write the same ARM child collection
	// (.../workspaces/{ws}/dataStores/{name}).
	//
	// Credentials are WRITE-ONLY: ARM never returns account keys, SAS
	// tokens or client secrets -- the provider echoes them from
	// configuration (recorded in the import catalog as
	// write-normalized).
	switch {
	case spec.BlobStorage != nil:
		// The blob-container variant. The only variant where
		// is_default is settable; auth is account key or SAS unless a
		// workspace-identity mode covers service-side access (spec
		// CEL).
		blobArgs := &machinelearning.DatastoreBlobstorageArgs{
			Name:               pulumi.String(spec.Name),
			WorkspaceId:        pulumi.String(locals.WorkspaceId),
			StorageContainerId: pulumi.String(spec.BlobStorage.StorageContainerId.GetValue()),
			IsDefault:          pulumi.Bool(spec.BlobStorage.IsDefault),
			Tags:               pulumi.ToStringMap(locals.AzureTags),
		}
		if spec.Description != "" {
			blobArgs.Description = pulumi.String(spec.Description)
		}
		// Enum name -> wire value; unspecified omits the property so
		// the provider applies its default, "None".
		if identityMode, ok := serviceDataIdentityWire[spec.ServiceDataIdentity]; ok {
			blobArgs.ServiceDataAuthIdentity = pulumi.String(identityMode)
		}
		// Sensitive -- resolved from secret references, masked in
		// state/preview by the provider schema. When both are set the
		// provider sends the SAS token (its own precedence).
		if spec.BlobStorage.AccountKey.GetValue() != "" {
			blobArgs.AccountKey = pulumi.String(spec.BlobStorage.AccountKey.GetValue())
		}
		if spec.BlobStorage.SharedAccessSignature.GetValue() != "" {
			blobArgs.SharedAccessSignature = pulumi.String(spec.BlobStorage.SharedAccessSignature.GetValue())
		}

		createdDatastore, err := machinelearning.NewDatastoreBlobstorage(ctx,
			spec.Name,
			blobArgs,
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create blob storage datastore %s", spec.Name)
		}

		ctx.Export(OpDatastoreId, createdDatastore.ID())
		ctx.Export(OpDatastoreName, createdDatastore.Name)
		ctx.Export(OpIsDefault, createdDatastore.IsDefault)

	case spec.DataLakeGen2 != nil:
		// The Data Lake Gen2 variant. Auth is the service-principal
		// triad (all-or-none, spec CEL) or workspace identity / none;
		// no account key or SAS on this variant.
		dataLakeArgs := &machinelearning.DatastoreDatalakeGen2Args{
			Name:               pulumi.String(spec.Name),
			WorkspaceId:        pulumi.String(locals.WorkspaceId),
			StorageContainerId: pulumi.String(spec.DataLakeGen2.StorageContainerId.GetValue()),
			Tags:               pulumi.ToStringMap(locals.AzureTags),
		}
		if spec.Description != "" {
			dataLakeArgs.Description = pulumi.String(spec.Description)
		}
		if identityMode, ok := serviceDataIdentityWire[spec.ServiceDataIdentity]; ok {
			dataLakeArgs.ServiceDataIdentity = pulumi.String(identityMode)
		}
		if spec.DataLakeGen2.TenantId != "" {
			dataLakeArgs.TenantId = pulumi.String(spec.DataLakeGen2.TenantId)
		}
		if spec.DataLakeGen2.ClientId != "" {
			dataLakeArgs.ClientId = pulumi.String(spec.DataLakeGen2.ClientId)
		}
		if spec.DataLakeGen2.ClientSecret.GetValue() != "" {
			dataLakeArgs.ClientSecret = pulumi.String(spec.DataLakeGen2.ClientSecret.GetValue())
		}
		if spec.DataLakeGen2.AuthorityUrl != "" {
			dataLakeArgs.AuthorityUrl = pulumi.String(spec.DataLakeGen2.AuthorityUrl)
		}

		createdDatastore, err := machinelearning.NewDatastoreDatalakeGen2(ctx,
			spec.Name,
			dataLakeArgs,
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create data lake gen2 datastore %s", spec.Name)
		}

		ctx.Export(OpDatastoreId, createdDatastore.ID())
		ctx.Export(OpDatastoreName, createdDatastore.Name)
		ctx.Export(OpIsDefault, createdDatastore.IsDefault)

	case spec.FileShare != nil:
		// The Azure Files variant. The provider's schema requires
		// exactly one of account key / SAS here regardless of identity
		// mode (spec CEL mirrors it). The share id uses the v5 format
		// (.../fileServices/default/shares/{name}).
		fileShareArgs := &machinelearning.DatastoreFileshareArgs{
			Name:               pulumi.String(spec.Name),
			WorkspaceId:        pulumi.String(locals.WorkspaceId),
			StorageFileshareId: pulumi.String(spec.FileShare.StorageFileshareId.GetValue()),
			Tags:               pulumi.ToStringMap(locals.AzureTags),
		}
		if spec.Description != "" {
			fileShareArgs.Description = pulumi.String(spec.Description)
		}
		if identityMode, ok := serviceDataIdentityWire[spec.ServiceDataIdentity]; ok {
			fileShareArgs.ServiceDataIdentity = pulumi.String(identityMode)
		}
		if spec.FileShare.AccountKey.GetValue() != "" {
			fileShareArgs.AccountKey = pulumi.String(spec.FileShare.AccountKey.GetValue())
		}
		if spec.FileShare.SharedAccessSignature.GetValue() != "" {
			fileShareArgs.SharedAccessSignature = pulumi.String(spec.FileShare.SharedAccessSignature.GetValue())
		}

		createdDatastore, err := machinelearning.NewDatastoreFileshare(ctx,
			spec.Name,
			fileShareArgs,
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create file share datastore %s", spec.Name)
		}

		ctx.Export(OpDatastoreId, createdDatastore.ID())
		ctx.Export(OpDatastoreName, createdDatastore.Name)
		ctx.Export(OpIsDefault, createdDatastore.IsDefault)

	default:
		// Unreachable: the spec's exactly_one_variant CEL rejects the
		// manifest before any module runs.
		return errors.New("no datastore variant block is set")
	}

	return nil
}

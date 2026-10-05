package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	azurestoragecontainerv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestoragecontainer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurestoragecontainerv1alpha1.AzureStorageContainerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureStorageContainer.Spec

	// The account name, parsed from the resolved account ARM ID for the
	// output -- consumers frequently need the account/container
	// name pair, and this saves them a second reference. The id must END
	// with /storageAccounts/{name} (matching the Terraform module's
	// anchored regex), so a malformed or over-long id fails loudly here
	// instead of computing a wrong name.
	accountIdParts := strings.Split(locals.StorageAccountId, "/storageAccounts/")
	if len(accountIdParts) != 2 || accountIdParts[1] == "" || strings.Contains(accountIdParts[1], "/") {
		return errors.Errorf("storage_account_id %q is not a storage-account ARM id", locals.StorageAccountId)
	}
	storageAccountName := accountIdParts[1]

	// Unspecified access type materializes "private" -- the container is
	// born locked down unless the spec says otherwise. Anonymous access
	// also requires the ACCOUNT's allow_nested_items_to_be_public; when
	// the account forbids it, Azure forces private regardless.
	containerAccessType := "private"
	if spec.ContainerAccessType != azurestoragecontainerv1alpha1.AzureStorageContainerAccessType_azure_storage_container_access_type_unspecified {
		containerAccessType = containerAccessTypeStrings[spec.ContainerAccessType]
	}

	// The container is addressed by the parent account's ARM ID (the
	// control-plane path -- the account-name form is the provider's
	// legacy data-plane path, removed in azurerm v5). Containers carry no
	// Azure tags: ARM does not support tags on blobServices/containers.
	containerArgs := &storage.ContainerArgs{
		Name:                pulumi.String(spec.ContainerName),
		StorageAccountId:    pulumi.String(locals.StorageAccountId),
		ContainerAccessType: pulumi.String(containerAccessType),
	}

	// Sub-account key isolation: blobs without their own scope encrypt
	// under this one (a reference to an AzureStorageEncryptionScope's
	// name, resolved to a literal before the module runs). Both fields
	// are fixed at creation; the override flag rides with the scope
	// (Azure's default is true when a scope is set, so it is
	// presence-guarded rather than defaulted).
	if spec.DefaultEncryptionScope.GetValue() != "" {
		containerArgs.DefaultEncryptionScope = pulumi.String(spec.DefaultEncryptionScope.GetValue())
	}
	if spec.EncryptionScopeOverrideEnabled != nil {
		containerArgs.EncryptionScopeOverrideEnabled = pulumi.Bool(spec.GetEncryptionScopeOverrideEnabled())
	}

	if len(spec.Metadata) > 0 {
		containerArgs.Metadata = pulumi.ToStringMap(spec.Metadata)
	}

	createdContainer, err := storage.NewContainer(ctx,
		fmt.Sprintf("%s-%s", storageAccountName, spec.ContainerName),
		containerArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create storage container %s", spec.ContainerName)
	}

	// Export outputs. The container's data-plane URL is deliberately
	// NOT exported -- compose it from the ACCOUNT's primary_blob_endpoint
	// output + container_name (only the account knows its real endpoint;
	// partitioned-DNS accounts use a different hostname).
	ctx.Export(OpContainerId, createdContainer.ID())
	ctx.Export(OpContainerName, createdContainer.Name)
	ctx.Export(OpStorageAccountName, pulumi.String(storageAccountName))

	return nil
}

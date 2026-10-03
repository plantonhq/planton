package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	azurestoragedatalakegen2filesystemv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestoragedatalakegen2filesystem/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurestoragedatalakegen2filesystemv1alpha1.AzureStorageDataLakeGen2FilesystemIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureStorageDataLakeGen2Filesystem.Spec

	// The account name, parsed from the resolved account ARM ID -- used
	// for the storage_account_name output and to construct the
	// filesystem's ARM container-proxy ID. The id must END with
	// /storageAccounts/{name} (matching the Terraform module's anchored
	// regex), so a malformed or over-long id fails loudly here instead
	// of computing a wrong name.
	accountIdParts := strings.Split(locals.StorageAccountId, "/storageAccounts/")
	if len(accountIdParts) != 2 || accountIdParts[1] == "" || strings.Contains(accountIdParts[1], "/") {
		return errors.Errorf("storage_account_id %q is not a storage-account ARM id", locals.StorageAccountId)
	}
	storageAccountName := accountIdParts[1]

	// The filesystem is a DATA-PLANE resource: the provider talks to the
	// account's dfs endpoint (with the account's shared key by default),
	// so the account must be reachable from where the deploy runs -- a
	// data-plane firewall that blocks the runner blocks the create even
	// though ARM would allow it. POSIX access control (owner/group/aces)
	// additionally requires hierarchical namespace on the ACCOUNT --
	// Azure rejects it on flat-namespace accounts at apply time,
	// deliberately not mirrored as spec validation because the account
	// arrives as a reference.
	filesystemArgs := &storage.DataLakeGen2FilesystemArgs{
		Name:             pulumi.String(spec.FilesystemName),
		StorageAccountId: pulumi.String(locals.StorageAccountId),
		// Azure requires the VALUES to be base64-encoded; keys stay plain.
		Properties: pulumi.ToStringMap(spec.Properties),
	}

	// Optional inputs are sent only when set so unset stays unset on the
	// wire (owner/group and the scope are Computed on the provider --
	// always sending an empty string would fight Azure's server-side
	// defaults).
	if spec.DefaultEncryptionScope.GetValue() != "" {
		// Sub-account key isolation: data that doesn't name its own scope
		// encrypts under this one. Fixed at creation.
		filesystemArgs.DefaultEncryptionScope = pulumi.String(spec.DefaultEncryptionScope.GetValue())
	}
	if spec.Owner.GetValue() != "" {
		filesystemArgs.Owner = pulumi.String(spec.Owner.GetValue())
	}
	if spec.Group.GetValue() != "" {
		filesystemArgs.Group = pulumi.String(spec.Group.GetValue())
	}

	// The root path's POSIX ACL. Access entries gate the root itself;
	// default entries are the template newly created children inherit --
	// how a zone's permission posture propagates to files landing in it.
	if len(spec.Aces) > 0 {
		aces := make(storage.DataLakeGen2FilesystemAceArray, 0, len(spec.Aces))
		for _, ace := range spec.Aces {
			aceArgs := storage.DataLakeGen2FilesystemAceArgs{
				Type:        pulumi.String(aceTypeStrings[ace.Type]),
				Permissions: pulumi.String(ace.Permissions),
			}
			// Unset leaves the provider's own default (access), matching
			// the Terraform module's null.
			if scope, ok := aceScopeStrings[ace.Scope]; ok {
				aceArgs.Scope = pulumi.String(scope)
			}
			// Only USER/GROUP entries name a principal (enforced in the
			// spec); an unqualified entry addresses the owning user/group.
			if ace.ObjectId.GetValue() != "" {
				aceArgs.Id = pulumi.String(ace.ObjectId.GetValue())
			}
			aces = append(aces, aceArgs)
		}
		filesystemArgs.Aces = aces
	}

	createdFilesystem, err := storage.NewDataLakeGen2Filesystem(ctx,
		fmt.Sprintf("%s-%s", storageAccountName, spec.FilesystemName),
		filesystemArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create data lake gen2 filesystem %s", spec.FilesystemName)
	}

	// ADLS filesystems surface in ARM as blob containers, so the
	// filesystem's management/RBAC identity is the container-proxy ID --
	// constructed from the account ID + name (identically on both
	// engines) because the provider's own resource id is a data-plane
	// dfs URL nothing management-grain can consume.
	ctx.Export(OpFilesystemId, pulumi.Sprintf("%s/blobServices/default/containers/%s",
		locals.StorageAccountId, spec.FilesystemName))
	ctx.Export(OpFilesystemName, createdFilesystem.Name)
	ctx.Export(OpStorageAccountName, pulumi.String(storageAccountName))

	return nil
}

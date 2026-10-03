package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	azurestoragetablev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestoragetable/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurestoragetablev1alpha1.AzureStorageTableIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureStorageTable.Spec

	// The account name, parsed from the resolved account ARM ID. The id
	// must END with /storageAccounts/{name} (matching the Terraform
	// module's anchored regex), so a malformed or over-long id fails
	// loudly here instead of computing a wrong name.
	accountIdParts := strings.Split(locals.StorageAccountId, "/storageAccounts/")
	if len(accountIdParts) != 2 || accountIdParts[1] == "" || strings.Contains(accountIdParts[1], "/") {
		return errors.Errorf("storage_account_id %q is not a storage-account ARM id", locals.StorageAccountId)
	}
	storageAccountName := accountIdParts[1]

	// PARITY-EXCEPTION: this module addresses the table by the account
	// NAME parsed from the spec's ARM id, while the Terraform module
	// passes storage_account_id (the resource-manager path) --
	// pulumi-azure v6 has not yet bridged the table's storage_account_id
	// input (verified at v6.38, the latest v6). The created table is
	// identical and all outputs match byte-for-byte (both engines
	// export the same resource_manager_id); only the provider's internal
	// addressing differs. Re-align to StorageAccountId when a bridge
	// release carries it on storage.Table.
	//
	// Operational contract: the provider drives table creation and ACLs
	// through the table DATA PLANE with shared-key authorization, so the
	// parent account must keep shared_access_key_enabled true (Azure's
	// default) for deploys to work. Tables carry no Azure tags: ARM does
	// not support tags on tableServices/tables.
	tableArgs := &storage.TableArgs{
		Name:               pulumi.String(spec.TableName),
		StorageAccountName: pulumi.String(storageAccountName),
	}

	// Stored access policies (signed identifiers): revoking or
	// shortening a policy immediately revokes every SAS token anchored
	// to it. Table policies require the full validity window (start +
	// expiry -- enforced in the spec). At most five per table (Azure's
	// limit, enforced in the spec).
	if len(spec.Acls) > 0 {
		aclArray := storage.TableAclArray{}
		for _, acl := range spec.Acls {
			policyArray := storage.TableAclAccessPolicyArray{}
			for _, policy := range acl.AccessPolicies {
				policyArray = append(policyArray, storage.TableAclAccessPolicyArgs{
					Permissions: pulumi.String(policy.Permissions),
					Start:       pulumi.String(policy.Start),
					Expiry:      pulumi.String(policy.Expiry),
				})
			}
			aclArray = append(aclArray, storage.TableAclArgs{
				Id:             pulumi.String(acl.Id),
				AccessPolicies: policyArray,
			})
		}
		tableArgs.Acls = aclArray
	}

	createdTable, err := storage.NewTable(ctx,
		fmt.Sprintf("%s-%s", storageAccountName, spec.TableName),
		tableArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create storage table %s", spec.TableName)
	}

	// Export outputs. The resource_manager_id attribute (rather
	// than the resource id) keeps table_id byte-identical across engines
	// regardless of which addressing path the provider used. The table's
	// data-plane URL is deliberately NOT exported -- compose client URLs
	// from the ACCOUNT's primary_table_endpoint output + table_name
	// (only the account knows its real endpoint; partitioned-DNS
	// accounts differ).
	ctx.Export(OpTableId, createdTable.ResourceManagerId)
	ctx.Export(OpTableName, createdTable.Name)
	ctx.Export(OpStorageAccountName, pulumi.String(storageAccountName))

	return nil
}

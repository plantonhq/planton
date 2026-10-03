package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	azurestoragesharev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestorageshare/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurestoragesharev1alpha1.AzureStorageShareIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureStorageShare.Spec

	// The account name, parsed from the resolved account ARM ID for the
	// output -- consumers frequently need the account/share name
	// pair, and this saves them a second reference. The id must END
	// with /storageAccounts/{name} (matching the Terraform module's
	// anchored regex), so a malformed or over-long id fails loudly here
	// instead of computing a wrong name.
	accountIdParts := strings.Split(locals.StorageAccountId, "/storageAccounts/")
	if len(accountIdParts) != 2 || accountIdParts[1] == "" || strings.Contains(accountIdParts[1], "/") {
		return errors.Errorf("storage_account_id %q is not a storage-account ARM id", locals.StorageAccountId)
	}
	storageAccountName := accountIdParts[1]

	// Unspecified protocol materializes SMB -- azurerm's own default
	// (and what Windows plus most Linux mounts use). NFS requires a
	// premium FileStorage account; Azure enforces the pairing at apply.
	enabledProtocol := "SMB"
	if spec.EnabledProtocol != azurestoragesharev1alpha1.AzureStorageShareProtocol_azure_storage_share_protocol_unspecified {
		enabledProtocol = enabledProtocolStrings[spec.EnabledProtocol]
	}

	// The share is addressed by the parent account's ARM ID (the
	// control-plane path -- the account-name form is the provider's
	// legacy data-plane path, removed in azurerm v5). Shares carry no
	// Azure tags: ARM does not support tags on fileServices/shares.
	shareArgs := &storage.ShareArgs{
		Name:             pulumi.String(spec.ShareName),
		StorageAccountId: pulumi.String(locals.StorageAccountId),
		// The provisioned quota in GB -- what SMB clients see as the
		// drive size and Azure enforces on writes. Grows in place;
		// shrinking below used capacity fails.
		Quota:           pulumi.Int(int(spec.QuotaGb)),
		EnabledProtocol: pulumi.String(enabledProtocol),
	}

	// The tier is sent only when the spec chooses one, so Azure's
	// per-account-kind default (TransactionOptimized on standard,
	// Premium on FileStorage) applies when unset.
	if spec.AccessTier != azurestoragesharev1alpha1.AzureStorageShareAccessTier_azure_storage_share_access_tier_unspecified {
		shareArgs.AccessTier = pulumi.String(accessTierStrings[spec.AccessTier])
	}

	// Stored access policies (signed identifiers): revoking or
	// shortening a policy immediately revokes every SAS token anchored
	// to it. At most five per share (Azure's limit, enforced in the
	// spec).
	if len(spec.Acls) > 0 {
		aclArray := storage.ShareAclArray{}
		for _, acl := range spec.Acls {
			policyArray := storage.ShareAclAccessPolicyArray{}
			for _, policy := range acl.AccessPolicies {
				policyArgs := storage.ShareAclAccessPolicyArgs{
					Permissions: pulumi.String(policy.Permissions),
				}
				// The validity window is optional on share policies --
				// the SAS token may carry it instead. Sent only when
				// set so an omitted end stays open on the policy.
				if policy.Start != "" {
					policyArgs.Start = pulumi.StringPtr(policy.Start)
				}
				if policy.Expiry != "" {
					policyArgs.Expiry = pulumi.StringPtr(policy.Expiry)
				}
				policyArray = append(policyArray, policyArgs)
			}
			aclArray = append(aclArray, storage.ShareAclArgs{
				Id:             pulumi.String(acl.Id),
				AccessPolicies: policyArray,
			})
		}
		shareArgs.Acls = aclArray
	}

	if len(spec.Metadata) > 0 {
		shareArgs.Metadata = pulumi.ToStringMap(spec.Metadata)
	}

	createdShare, err := storage.NewShare(ctx,
		fmt.Sprintf("%s-%s", storageAccountName, spec.ShareName),
		shareArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create storage share %s", spec.ShareName)
	}

	// Export outputs. The share's data-plane URL is deliberately
	// NOT exported -- compose mount paths from the ACCOUNT's
	// primary_file_endpoint output + share_name (only the account knows
	// its real endpoint; partitioned-DNS accounts differ). The RBAC
	// scope rides the provider's own attribute: Azure Files role
	// assignments scope to .../fileServices/default/fileshares/{name},
	// a DIFFERENT segment than the management ID.
	ctx.Export(OpShareId, createdShare.ID())
	ctx.Export(OpRbacScopeId, createdShare.RbacScopeId)
	ctx.Export(OpShareName, createdShare.Name)
	ctx.Export(OpStorageAccountName, pulumi.String(storageAccountName))

	return nil
}

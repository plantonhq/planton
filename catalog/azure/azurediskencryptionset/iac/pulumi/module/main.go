package module

import (
	"github.com/pkg/errors"
	azurediskencryptionsetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurediskencryptionset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurediskencryptionsetv1alpha1.AzureDiskEncryptionSetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDiskEncryptionSet.Spec

	// The identity block is required: a set cannot unwrap its key without a
	// managed identity. identity_ids carry the user-assigned identities (a
	// spec CEL guarantees they are present iff the type is user-assigned).
	identityArgs := &compute.DiskEncryptionSetIdentityArgs{
		Type: pulumi.String(locals.IdentityType),
	}
	if len(locals.IdentityIds) > 0 {
		identityArgs.IdentityIds = pulumi.ToStringArray(locals.IdentityIds)
	}

	desArgs := &compute.DiskEncryptionSetArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// The key URL must be versionless when auto-rotation is on and
		// versioned when it is off; the referenced key output supplies the
		// right form (versionless_id vs key_id) and the provider validates
		// the pairing at apply.
		KeyVaultKeyId: pulumi.String(locals.KeyVaultKeyId),
		Identity:      identityArgs,
		Tags:          pulumi.ToStringMap(locals.AzureTags),
	}

	// Azure defaults auto-rotation to false and encryption_type to
	// EncryptionAtRestWithCustomerKey; only an explicit choice is sent so an
	// unspecified spec deploys identically on both engines.
	if spec.AutoKeyRotationEnabled != nil {
		desArgs.AutoKeyRotationEnabled = pulumi.Bool(spec.GetAutoKeyRotationEnabled())
	}
	if locals.EncryptionType != "" {
		desArgs.EncryptionType = pulumi.String(locals.EncryptionType)
	}
	if spec.FederatedClientId != "" {
		desArgs.FederatedClientId = pulumi.String(spec.FederatedClientId)
	}

	createdDes, err := compute.NewDiskEncryptionSet(ctx,
		spec.Name,
		desArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create disk encryption set %s", spec.Name)
	}

	ctx.Export(OpDiskEncryptionSetId, createdDes.ID())
	ctx.Export(OpDiskEncryptionSetName, createdDes.Name)

	// The system-assigned identity's principal/tenant are known only after
	// creation; export them as the grant target (empty for user-assigned-only
	// sets, which grant their own identities out of band).
	principalId := createdDes.Identity.ApplyT(func(id compute.DiskEncryptionSetIdentity) string {
		if id.PrincipalId != nil {
			return *id.PrincipalId
		}
		return ""
	}).(pulumi.StringOutput)
	ctx.Export(OpIdentityPrincipalId, principalId)

	tenantId := createdDes.Identity.ApplyT(func(id compute.DiskEncryptionSetIdentity) string {
		if id.TenantId != nil {
			return *id.TenantId
		}
		return ""
	}).(pulumi.StringOutput)
	ctx.Export(OpIdentityTenantId, tenantId)

	return nil
}

package module

import (
	"github.com/pkg/errors"
	azuredataprotectionresourceguardv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredataprotectionresourceguard/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/dataprotection"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuredataprotectionresourceguardv1alpha1.AzureDataProtectionResourceGuardIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDataProtectionResourceGuard.Spec

	// Create the Data Protection Resource Guard -- the approval gate
	// behind Multi-User Authorization. The guard's protection comes
	// from SCOPE SEPARATION: deploy it in a resource group a different
	// administrator controls than the vaults it guards; a guard in the
	// same scope as its vaults is a speed bump, not a control.
	//
	// The guard is a free configuration object; vaults opt in by
	// referencing its ARM ID.
	guardArgs := &dataprotection.ResourceGuardArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Location:          pulumi.String(spec.Region),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Operations EXCLUDED from the approval requirement. Empty guards
	// everything (the strongest posture). Updates in place.
	if len(spec.VaultCriticalOperationExclusionList) > 0 {
		guardArgs.VaultCriticalOperationExclusionLists = pulumi.ToStringArray(spec.VaultCriticalOperationExclusionList)
	}

	createdGuard, err := dataprotection.NewResourceGuard(ctx,
		spec.Name,
		guardArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create data protection resource guard %s", spec.Name)
	}

	ctx.Export(OpResourceGuardId, createdGuard.ID())
	ctx.Export(OpResourceGuardName, createdGuard.Name)

	return nil
}

package module

import (
	"github.com/pkg/errors"
	azurevirtualmachinescalesetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualmachinescaleset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// scaleSetOutputs carries the mode-independent export surface each of
// the three resource shapes produces.
type scaleSetOutputs struct {
	id          pulumi.StringOutput
	name        pulumi.StringOutput
	uniqueId    pulumi.StringOutput
	principalId pulumi.StringOutput
}

func Resources(ctx *pulumi.Context, iacInput *azurevirtualmachinescalesetv1alpha1.AzureVirtualMachineScaleSetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// ONE spec surface realizes onto azurerm's three scale-set resources:
	// linux/windows (UNIFORM orchestration) and orchestrated (FLEXIBLE).
	// ARM has a single scale-set resource type with an orchestration-mode
	// property; the three-resource split is the provider's ergonomics, so
	// the dispatch lives here rather than in the user's model.
	var outputs *scaleSetOutputs
	switch {
	case locals.IsUniform && locals.IsLinux:
		outputs, err = createUniformLinux(ctx, locals, azureProvider)
	case locals.IsUniform:
		outputs, err = createUniformWindows(ctx, locals, azureProvider)
	default:
		outputs, err = createFlexible(ctx, locals, azureProvider)
	}
	if err != nil {
		return err
	}

	// Export outputs. The scale set's ARM id is the seam a
	// standalone VM's scale-set attachment consumes; the system-assigned
	// principal is the AzureRoleAssignment seam (UNIFORM sets only --
	// FLEXIBLE sets carry user-assigned identities whose principals live
	// on the identity resources themselves).
	ctx.Export(OpScaleSetId, outputs.id)
	ctx.Export(OpScaleSetName, outputs.name)
	ctx.Export(OpUniqueId, outputs.uniqueId)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, outputs.principalId)

	return nil
}

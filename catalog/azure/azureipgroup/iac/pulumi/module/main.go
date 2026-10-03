package module

import (
	"github.com/pkg/errors"
	azureipgroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureipgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureipgroupv1alpha1.AzureIpGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureIpGroup.Spec

	// The IP Group is a passive, named address set -- it carries no rules
	// of its own. Consumption is declared from the rule's side (a firewall
	// policy rule lists source/destination IP Groups), which is exactly
	// what makes the group a stable composition target: its address list
	// updates in place and every referencing rule follows immediately,
	// without touching the rules themselves.
	createdIpGroup, err := network.NewIPGroup(ctx,
		spec.Name,
		&network.IPGroupArgs{
			Name:              pulumi.String(spec.Name),
			Location:          pulumi.String(spec.Region),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Cidrs:             pulumi.ToStringArray(spec.Cidrs),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create ip group %s", spec.Name)
	}

	ctx.Export(OpIpGroupId, createdIpGroup.ID())
	ctx.Export(OpIpGroupName, createdIpGroup.Name)

	return nil
}

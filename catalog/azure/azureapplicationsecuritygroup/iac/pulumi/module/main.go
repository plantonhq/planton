package module

import (
	"github.com/pkg/errors"
	azureapplicationsecuritygroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureapplicationsecuritygroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureapplicationsecuritygroupv1alpha1.AzureApplicationSecurityGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureApplicationSecurityGroup.Spec

	// The application security group carries no members of its own -- it is
	// an empty, named anchor. Membership is declared from the member side
	// (a network interface lists the ASGs it joins; an NSG rule references
	// source/destination ASGs), which is exactly what makes the ASG a
	// stable composition target. Everything except tags is fixed at
	// creation; changing name or region replaces the group.
	createdAsg, err := network.NewApplicationSecurityGroup(ctx,
		spec.Name,
		&network.ApplicationSecurityGroupArgs{
			Name:              pulumi.String(spec.Name),
			Location:          pulumi.String(spec.Region),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create application security group %s", spec.Name)
	}

	ctx.Export(OpApplicationSecurityGroupId, createdAsg.ID())
	ctx.Export(OpApplicationSecurityGroupName, createdAsg.Name)

	return nil
}

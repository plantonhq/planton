package module

import (
	"github.com/pkg/errors"
	azureavailabilitysetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureavailabilityset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureavailabilitysetv1alpha1.AzureAvailabilitySetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureAvailabilitySet.Spec

	// Create the availability set. The whole configuration is fixed at
	// creation (only tags update in place). Optional fields are sent
	// only when set so the provider's own defaults apply (5 update
	// domains, 3 fault domains, managed=true -- managed aligns fault
	// domains with the VMs' managed-disk storage).
	args := &compute.AvailabilitySetArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.PlatformUpdateDomainCount != nil {
		args.PlatformUpdateDomainCount = pulumi.Int(int(spec.GetPlatformUpdateDomainCount()))
	}
	if spec.PlatformFaultDomainCount != nil {
		args.PlatformFaultDomainCount = pulumi.Int(int(spec.GetPlatformFaultDomainCount()))
	}
	if spec.Managed != nil {
		args.Managed = pulumi.Bool(spec.GetManaged())
	}
	if spec.ProximityPlacementGroupId.GetValue() != "" {
		args.ProximityPlacementGroupId = pulumi.String(spec.ProximityPlacementGroupId.GetValue())
	}

	createdAvailabilitySet, err := compute.NewAvailabilitySet(ctx,
		locals.AzureAvailabilitySet.Metadata.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create availability set %s",
			locals.AzureAvailabilitySet.Metadata.Name)
	}

	ctx.Export(OpAvailabilitySetId, createdAvailabilitySet.ID())
	ctx.Export(OpAvailabilitySetName, createdAvailabilitySet.Name)

	return nil
}

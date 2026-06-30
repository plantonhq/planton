package module

import (
	"github.com/pkg/errors"
	azurepublicipv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurepublicip/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurepublicipv1.AzurePublicIpStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePublicIp.Spec

	// Build Public IP arguments.
	// SKU is always Standard (Basic was retired Sept 2025).
	// Allocation is always Static (Standard SKU requires it).
	publicIpArgs := &network.PublicIpArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		AllocationMethod:  pulumi.String("Static"),
		Sku:               pulumi.String("Standard"),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Set domain name label if specified
	if spec.DomainNameLabel != "" {
		publicIpArgs.DomainNameLabel = pulumi.String(spec.DomainNameLabel)
	}

	// Set availability zones if specified
	if len(spec.Zones) > 0 {
		zones := pulumi.StringArray{}
		for _, zone := range spec.Zones {
			zones = append(zones, pulumi.String(zone))
		}
		publicIpArgs.Zones = zones
	}

	// Set idle timeout (uses default from proto if not explicitly set)
	publicIpArgs.IdleTimeoutInMinutes = pulumi.Int(int(spec.GetIdleTimeoutInMinutes()))

	// Create the Public IP
	publicIp, err := network.NewPublicIp(ctx,
		spec.Name,
		publicIpArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Public IP %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpPublicIpId, publicIp.ID())
	ctx.Export(OpIpAddress, publicIp.IpAddress)
	ctx.Export(OpFqdn, publicIp.Fqdn)
	ctx.Export(OpPublicIpName, publicIp.Name)

	return nil
}

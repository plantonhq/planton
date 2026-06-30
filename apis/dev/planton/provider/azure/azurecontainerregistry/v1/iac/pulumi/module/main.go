package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurecontainerregistryv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurecontainerregistry/v1"
	"github.com/pulumi/pulumi-azure-native-sdk/containerregistry/v3"
	azurenative "github.com/pulumi/pulumi-azure-native-sdk/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurecontainerregistryv1.AzureContainerRegistryStackInput) error {
	azureProviderConfig := stackInput.ProviderConfig

	// Create Azure provider using the credentials from the input
	provider, err := azurenative.NewProvider(ctx,
		"azure",
		&azurenative.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// Get inputs
	target := stackInput.Target
	spec := target.Spec

	// Resource group is provided externally
	resourceGroupName := spec.ResourceGroup.GetValue()

	// Determine SKU name
	skuName := "Standard" // Default
	switch spec.Sku {
	case azurecontainerregistryv1.AzureContainerRegistrySku_BASIC:
		skuName = "Basic"
	case azurecontainerregistryv1.AzureContainerRegistrySku_PREMIUM:
		skuName = "Premium"
	}

	// Build registry arguments
	registryArgs := &containerregistry.RegistryArgs{
		ResourceGroupName: pulumi.String(resourceGroupName),
		RegistryName:      pulumi.String(spec.RegistryName),
		Location:          pulumi.String(spec.Region),

		// SKU
		Sku: &containerregistry.SkuArgs{
			Name: pulumi.String(skuName),
		},

		// Admin user
		AdminUserEnabled: pulumi.Bool(spec.AdminUserEnabled),

		// Network rule bypass
		NetworkRuleBypassOptions: pulumi.String("AzureServices"),
	}

	// Create the registry
	registry, err := containerregistry.NewRegistry(ctx, spec.RegistryName, registryArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create container registry")
	}

	// Create geo-replications if Premium SKU
	if spec.Sku == azurecontainerregistryv1.AzureContainerRegistrySku_PREMIUM {
		for _, replicaRegion := range spec.GeoReplicationRegions {
			replicationName := fmt.Sprintf("%s-%s", spec.RegistryName, replicaRegion)
			_, err := containerregistry.NewReplication(ctx, replicationName, &containerregistry.ReplicationArgs{
				ResourceGroupName: pulumi.String(resourceGroupName),
				RegistryName:      registry.Name,
				ReplicationName:   pulumi.String(replicaRegion),
				Location:          pulumi.String(replicaRegion),
			}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{registry}))
			if err != nil {
				return errors.Wrapf(err, "failed to create geo-replication for region: %s", replicaRegion)
			}
		}
	}

	// Export outputs
	ctx.Export(OpRegistryLoginServer, registry.LoginServer)
	ctx.Export(OpRegistryResourceId, registry.ID())

	return nil
}

package module

import (
	"github.com/pkg/errors"
	azureaksnodepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azureaksnodepool/v1"
	"github.com/pulumi/pulumi-azure-native-sdk/containerservice/v3"
	azurenative "github.com/pulumi/pulumi-azure-native-sdk/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureaksnodepoolv1.AzureAksNodePoolStackInput) error {
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

	// Get cluster name from foreign key reference
	clusterName := spec.ClusterName.GetValue()

	// Determine OS type
	osType := "Linux"
	if spec.OsType != nil && *spec.OsType == azureaksnodepoolv1.AzureAksNodePoolOsType_WINDOWS {
		osType = "Windows"
	}

	// Determine mode
	mode := "User" // Default to User
	if spec.Mode != nil && *spec.Mode == azureaksnodepoolv1.AzureAksNodePoolMode_SYSTEM {
		mode = "System"
	}

	// Build availability zones array
	var availabilityZones pulumi.StringArray
	for _, az := range spec.AvailabilityZones {
		availabilityZones = append(availabilityZones, pulumi.String(az))
	}

	// Build node pool arguments
	nodePoolArgs := &containerservice.AgentPoolArgs{
		// The resource_group field is a StringValueOrRef. The platform middleware resolves
		// valueFrom references before IaC modules run, so .GetValue() always returns the
		// resolved literal string.
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),

		// The parent cluster name
		ResourceName: pulumi.String(clusterName),

		// Node pool name
		AgentPoolName: pulumi.String(target.Metadata.Name),

		// VM size
		VmSize: pulumi.String(spec.VmSize),

		// Node count
		Count: pulumi.Int(int(spec.InitialNodeCount)),

		// Mode (System or User)
		Mode: pulumi.String(mode),

		// OS type
		OsType: pulumi.String(osType),

		// Availability zones
		AvailabilityZones: availabilityZones,
	}

	// Configure autoscaling if specified
	if spec.Autoscaling != nil {
		nodePoolArgs.EnableAutoScaling = pulumi.Bool(true)
		nodePoolArgs.MinCount = pulumi.Int(int(spec.Autoscaling.MinNodes))
		nodePoolArgs.MaxCount = pulumi.Int(int(spec.Autoscaling.MaxNodes))
	}

	// Configure Spot instances if enabled
	if spec.SpotEnabled {
		nodePoolArgs.ScaleSetPriority = pulumi.String("Spot")
		nodePoolArgs.ScaleSetEvictionPolicy = pulumi.String("Delete")
		nodePoolArgs.SpotMaxPrice = pulumi.Float64(-1) // Pay up to regular price
	}

	// Create the node pool
	nodePool, err := containerservice.NewAgentPool(ctx, target.Metadata.Name, nodePoolArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create AKS node pool")
	}

	// Export outputs
	ctx.Export(OpNodePoolName, nodePool.Name)
	ctx.Export(OpAgentPoolResourceId, nodePool.ID())
	ctx.Export(OpMaxPodsPerNode, nodePool.MaxPods)

	return nil
}

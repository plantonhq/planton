package module

import (
	"github.com/pkg/errors"
	azurecontainerappenvironmentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurecontainerappenvironment/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurecontainerappenvironmentv1.AzureContainerAppEnvironmentStackInput) error {
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

	spec := locals.AzureContainerAppEnvironment.Spec

	// Build Container App Environment arguments
	envArgs := &containerapp.EnvironmentArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Configure VNet injection if subnet is provided
	if spec.InfrastructureSubnetId != nil {
		envArgs.InfrastructureSubnetId = pulumi.String(spec.InfrastructureSubnetId.GetValue())
	}

	// Configure Log Analytics if workspace is provided
	// Auto-derive logs_destination: "log-analytics" when workspace provided
	if spec.LogAnalyticsWorkspaceId != nil {
		envArgs.LogAnalyticsWorkspaceId = pulumi.String(spec.LogAnalyticsWorkspaceId.GetValue())
		envArgs.LogsDestination = pulumi.String("log-analytics")
	}

	// Set internal load balancer mode if specified
	if spec.InternalLoadBalancerEnabled != nil {
		envArgs.InternalLoadBalancerEnabled = pulumi.Bool(spec.GetInternalLoadBalancerEnabled())
	}

	// Set zone redundancy if specified
	if spec.ZoneRedundancyEnabled != nil {
		envArgs.ZoneRedundancyEnabled = pulumi.Bool(spec.GetZoneRedundancyEnabled())
	}

	// Configure workload profiles (dedicated compute)
	// Note: Azure auto-adds the "Consumption" profile -- we only pass user-defined profiles
	if len(spec.WorkloadProfiles) > 0 {
		profiles := make(containerapp.EnvironmentWorkloadProfileArray, 0, len(spec.WorkloadProfiles))
		for _, wp := range spec.WorkloadProfiles {
			profile := &containerapp.EnvironmentWorkloadProfileArgs{
				Name:                pulumi.String(wp.Name),
				WorkloadProfileType: pulumi.String(wp.WorkloadProfileType),
			}
			if wp.MinimumCount != nil {
				profile.MinimumCount = pulumi.Int(int(wp.GetMinimumCount()))
			}
			if wp.MaximumCount != nil {
				profile.MaximumCount = pulumi.Int(int(wp.GetMaximumCount()))
			}
			profiles = append(profiles, profile)
		}
		envArgs.WorkloadProfiles = profiles
	}

	// Create the Container App Environment
	env, err := containerapp.NewEnvironment(ctx,
		spec.Name,
		envArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Container App Environment %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpEnvironmentId, env.ID())
	ctx.Export(OpDefaultDomain, env.DefaultDomain)
	ctx.Export(OpStaticIpAddress, env.StaticIpAddress)
	ctx.Export(OpPlatformReservedCidr, env.PlatformReservedCidr)
	ctx.Export(OpPlatformReservedDnsIpAddress, env.PlatformReservedDnsIpAddress)

	return nil
}

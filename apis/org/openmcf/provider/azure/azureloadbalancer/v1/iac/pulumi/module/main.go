package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureloadbalancer/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureloadbalancerv1.AzureLoadBalancerStackInput) error {
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

	spec := locals.AzureLoadBalancer.Spec

	// Build the frontend IP configuration.
	// The LB mode (public vs internal) is determined by which field is set:
	// - public_ip_id set --> public frontend
	// - subnet_id set --> internal frontend with optional static private IP
	frontendConfig := lb.LoadBalancerFrontendIpConfigurationArgs{
		Name: pulumi.String(locals.FrontendConfigName),
	}

	if spec.PublicIpId != nil && spec.PublicIpId.GetValue() != "" {
		// Public LB: frontend uses a public IP address
		frontendConfig.PublicIpAddressId = pulumi.StringPtr(spec.PublicIpId.GetValue())
	} else if spec.SubnetId != nil && spec.SubnetId.GetValue() != "" {
		// Internal LB: frontend uses a subnet with optional static private IP
		frontendConfig.SubnetId = pulumi.StringPtr(spec.SubnetId.GetValue())
		if spec.PrivateIpAddress != "" {
			frontendConfig.PrivateIpAddress = pulumi.StringPtr(spec.PrivateIpAddress)
			frontendConfig.PrivateIpAddressAllocation = pulumi.StringPtr("Static")
		} else {
			frontendConfig.PrivateIpAddressAllocation = pulumi.StringPtr("Dynamic")
		}
	}

	// Create the Load Balancer with Standard SKU (hardcoded).
	// Basic SKU was retired Sept 2025 and lacks zone redundancy, SLA, and outbound rules.
	loadBalancer, err := lb.NewLoadBalancer(ctx,
		spec.Name,
		&lb.LoadBalancerArgs{
			Name:              pulumi.String(spec.Name),
			Location:          pulumi.String(spec.Region),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Sku:               pulumi.String("Standard"),
			FrontendIpConfigurations: lb.LoadBalancerFrontendIpConfigurationArray{
				&frontendConfig,
			},
			Tags: pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Load Balancer %s", spec.Name)
	}

	// Create backend address pools.
	// Pool membership (VMs, VMSS, NICs) is managed externally.
	// We track pools by name for rule references.
	backendPools := make(map[string]*lb.BackendAddressPool)
	for _, pool := range spec.BackendPools {
		bp, err := lb.NewBackendAddressPool(ctx,
			fmt.Sprintf("%s-%s", spec.Name, pool.Name),
			&lb.BackendAddressPoolArgs{
				Name:           pulumi.String(pool.Name),
				LoadbalancerId: loadBalancer.ID(),
			},
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{loadBalancer}))
		if err != nil {
			return errors.Wrapf(err, "failed to create backend pool %s", pool.Name)
		}
		backendPools[pool.Name] = bp
	}

	// Create health probes.
	// We track probes by name for rule references.
	probes := make(map[string]*lb.Probe)
	for _, probe := range spec.HealthProbes {
		probeArgs := &lb.ProbeArgs{
			Name:              pulumi.String(probe.Name),
			LoadbalancerId:    loadBalancer.ID(),
			Protocol:          pulumi.String(probe.Protocol),
			Port:              pulumi.Int(int(probe.Port)),
			IntervalInSeconds: pulumi.Int(int(probe.GetIntervalInSeconds())),
			NumberOfProbes:    pulumi.Int(int(probe.GetNumberOfProbes())),
		}

		// request_path is required for Http/Https probes, ignored for Tcp
		if probe.RequestPath != "" {
			probeArgs.RequestPath = pulumi.String(probe.RequestPath)
		}

		p, err := lb.NewProbe(ctx,
			fmt.Sprintf("%s-%s", spec.Name, probe.Name),
			probeArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{loadBalancer}))
		if err != nil {
			return errors.Wrapf(err, "failed to create health probe %s", probe.Name)
		}
		probes[probe.Name] = p
	}

	// Create load balancing rules.
	// Each rule references a backend pool and probe by name.
	for _, rule := range spec.Rules {
		pool, ok := backendPools[rule.BackendPoolName]
		if !ok {
			return errors.Errorf("rule %s references unknown backend pool %s", rule.Name, rule.BackendPoolName)
		}

		probe, ok := probes[rule.ProbeName]
		if !ok {
			return errors.Errorf("rule %s references unknown health probe %s", rule.Name, rule.ProbeName)
		}

		_, err := lb.NewRule(ctx,
			fmt.Sprintf("%s-%s", spec.Name, rule.Name),
			&lb.RuleArgs{
				Name:                     pulumi.String(rule.Name),
				LoadbalancerId:           loadBalancer.ID(),
				FrontendIpConfigurationName: pulumi.String(locals.FrontendConfigName),
				Protocol:                 pulumi.String(rule.Protocol),
				FrontendPort:             pulumi.Int(int(rule.FrontendPort)),
				BackendPort:              pulumi.Int(int(rule.BackendPort)),
				BackendAddressPoolIds:    pulumi.StringArray{pool.ID()},
				ProbeId:                  probe.ID(),
				IdleTimeoutInMinutes:     pulumi.Int(int(rule.GetIdleTimeoutInMinutes())),
				EnableFloatingIp:         pulumi.Bool(rule.GetEnableFloatingIp()),
				DisableOutboundSnat:      pulumi.Bool(rule.GetDisableOutboundSnat()),
			},
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{loadBalancer, pool, probe}))
		if err != nil {
			return errors.Wrapf(err, "failed to create load balancing rule %s", rule.Name)
		}
	}

	// Export stack outputs.
	ctx.Export(OpLbId, loadBalancer.ID())
	ctx.Export(OpLbName, loadBalancer.Name)

	// Export the frontend IP address.
	// For public LB, this comes from the referenced public IP.
	// For internal LB, this is the private IP allocated to the frontend config.
	ctx.Export(OpFrontendIpAddress, loadBalancer.FrontendIpConfigurations.Index(pulumi.Int(0)).PrivateIpAddress())

	// Export the frontend IP configuration ID.
	ctx.Export(OpFrontendIpConfigurationId, loadBalancer.FrontendIpConfigurations.Index(pulumi.Int(0)).Id())

	// Export the first (default) backend pool ID.
	if len(spec.BackendPools) > 0 {
		firstPool := backendPools[spec.BackendPools[0].Name]
		ctx.Export(OpBackendPoolId, firstPool.ID())
	}

	return nil
}

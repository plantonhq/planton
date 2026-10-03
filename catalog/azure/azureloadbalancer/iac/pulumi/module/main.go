package module

import (
	"fmt"

	"github.com/pkg/errors"
	azureloadbalancerv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureloadbalancer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Map the spec enums to ARM's exact API values. Unset enums apply Azure's
// defaults (Standard SKU, Regional tier, Tcp probe, Default distribution),
// so an unspecified spec and Azure's default deploy identically on both
// engines.
func transportProtocolToArm(p azureloadbalancerv1alpha1.AzureLoadBalancerTransportProtocol) string {
	switch p {
	case azureloadbalancerv1alpha1.AzureLoadBalancerTransportProtocol_UDP:
		return "Udp"
	case azureloadbalancerv1alpha1.AzureLoadBalancerTransportProtocol_ALL:
		return "All"
	default:
		return "Tcp"
	}
}

func probeProtocolToArm(p azureloadbalancerv1alpha1.AzureLoadBalancerProbeProtocol) string {
	switch p {
	case azureloadbalancerv1alpha1.AzureLoadBalancerProbeProtocol_PROBE_HTTP:
		return "Http"
	case azureloadbalancerv1alpha1.AzureLoadBalancerProbeProtocol_PROBE_HTTPS:
		return "Https"
	default:
		return "Tcp"
	}
}

func tunnelProtocolToArm(p azureloadbalancerv1alpha1.AzureLoadBalancerTunnelProtocol) string {
	switch p {
	case azureloadbalancerv1alpha1.AzureLoadBalancerTunnelProtocol_NATIVE:
		return "Native"
	case azureloadbalancerv1alpha1.AzureLoadBalancerTunnelProtocol_VXLAN:
		return "VXLAN"
	default:
		return "None"
	}
}

func tunnelTypeToArm(t azureloadbalancerv1alpha1.AzureLoadBalancerTunnelType) string {
	switch t {
	case azureloadbalancerv1alpha1.AzureLoadBalancerTunnelType_INTERNAL:
		return "Internal"
	case azureloadbalancerv1alpha1.AzureLoadBalancerTunnelType_EXTERNAL:
		return "External"
	default:
		return "None"
	}
}

// optionalInt32 resolves an optional int32 field to its value or, when
// unset, the proto-declared default. IaC-input paths that bypass the
// manifest loader deliver unset optionals as nil, and sending a bare
// getter's zero would fail the provider's range validations (e.g. a
// probe interval must be >= 5) -- the fallback keeps unset meaning
// "Azure's default" on both engines, matching the Terraform module's
// optional(number, N) encodings.
func optionalInt32(v *int32, def int32) int32 {
	if v != nil {
		return *v
	}
	return def
}

func Resources(ctx *pulumi.Context, iacInput *azureloadbalancerv1alpha1.AzureLoadBalancerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureLoadBalancer.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - SKU, SKU tier, and edge zone are fixed at creation -- changing
	//   any of them replaces the load balancer. Frontends, pools, probes,
	//   and rules all update in place.
	// - Azure does not allow removing ALL frontends from an existing load
	//   balancer; going from some to none replaces the resource.
	// - Changing a frontend's zones replaces that frontend (and briefly
	//   its address) -- pick the zone posture up front.
	sku := "Standard"
	if spec.Sku == azureloadbalancerv1alpha1.AzureLoadBalancerSku_GATEWAY {
		sku = "Gateway"
	}
	skuTier := "Regional"
	if spec.SkuTier == azureloadbalancerv1alpha1.AzureLoadBalancerSkuTier_GLOBAL {
		skuTier = "Global"
	}

	// Each frontend is public (public IP / prefix) or internal (subnet).
	// The allocation is derived: a pinned private address means Static, an
	// internal frontend without a pin stays Dynamic, and public frontends
	// carry no allocation at all.
	frontends := lb.LoadBalancerFrontendIpConfigurationArray{}
	for _, f := range spec.FrontendIpConfigurations {
		frontendArgs := lb.LoadBalancerFrontendIpConfigurationArgs{
			Name: pulumi.String(f.Name),
		}
		isInternal := f.SubnetId.GetValue() != ""
		if isInternal {
			frontendArgs.SubnetId = pulumi.StringPtr(f.SubnetId.GetValue())
			if f.PrivateIpAddress != "" {
				frontendArgs.PrivateIpAddress = pulumi.StringPtr(f.PrivateIpAddress)
				frontendArgs.PrivateIpAddressAllocation = pulumi.StringPtr("Static")
			} else {
				frontendArgs.PrivateIpAddressAllocation = pulumi.StringPtr("Dynamic")
			}
			version := "IPv4"
			if f.PrivateIpAddressVersion == azureloadbalancerv1alpha1.AzureLoadBalancerPrivateIpVersion_IPV6 {
				version = "IPv6"
			}
			frontendArgs.PrivateIpAddressVersion = pulumi.StringPtr(version)
		}
		if f.PublicIpAddressId.GetValue() != "" {
			frontendArgs.PublicIpAddressId = pulumi.StringPtr(f.PublicIpAddressId.GetValue())
		}
		if f.PublicIpPrefixId.GetValue() != "" {
			frontendArgs.PublicIpPrefixId = pulumi.StringPtr(f.PublicIpPrefixId.GetValue())
		}
		if len(f.Zones) > 0 {
			frontendArgs.Zones = pulumi.ToStringArray(f.Zones)
		}
		if f.GatewayLoadBalancerFrontendIpConfigurationId != "" {
			frontendArgs.GatewayLoadBalancerFrontendIpConfigurationId = pulumi.StringPtr(f.GatewayLoadBalancerFrontendIpConfigurationId)
		}
		frontends = append(frontends, frontendArgs)
	}

	loadBalancerArgs := &lb.LoadBalancerArgs{
		Name:                     pulumi.String(spec.Name),
		Location:                 pulumi.String(spec.Region),
		ResourceGroupName:        pulumi.String(locals.ResourceGroupName),
		Sku:                      pulumi.String(sku),
		SkuTier:                  pulumi.String(skuTier),
		FrontendIpConfigurations: frontends,
		Tags:                     pulumi.ToStringMap(locals.AzureTags),
	}
	if spec.EdgeZone != "" {
		loadBalancerArgs.EdgeZone = pulumi.String(spec.EdgeZone)
	}

	loadBalancer, err := lb.NewLoadBalancer(ctx,
		spec.Name,
		loadBalancerArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create load balancer %s", spec.Name)
	}

	// The default frontend name: rules may omit the frontend when exactly
	// one is declared (spec-level validation guarantees the omission only
	// happens then).
	defaultFrontendName := spec.FrontendIpConfigurations[0].Name
	frontendNameFor := func(explicit string) string {
		if explicit != "" {
			return explicit
		}
		return defaultFrontendName
	}

	// Backend address pools. NIC-based membership joins from the member
	// side (via the exported pool IDs); only vnet-scoped IP-based
	// membership is declared here. Tunnel interfaces exist only on
	// GATEWAY-SKU pools (spec-level validation pairs them with the SKU).
	backendPools := make(map[string]*lb.BackendAddressPool)
	poolIds := pulumi.StringMap{}
	for _, pool := range spec.BackendPools {
		poolArgs := &lb.BackendAddressPoolArgs{
			Name:           pulumi.String(pool.Name),
			LoadbalancerId: loadBalancer.ID(),
		}
		if pool.VirtualNetworkId.GetValue() != "" {
			poolArgs.VirtualNetworkId = pulumi.StringPtr(pool.VirtualNetworkId.GetValue())
		}
		if pool.SynchronousMode == azureloadbalancerv1alpha1.AzureLoadBalancerBackendPoolSyncMode_AUTOMATIC {
			poolArgs.SynchronousMode = pulumi.StringPtr("Automatic")
		} else if pool.SynchronousMode == azureloadbalancerv1alpha1.AzureLoadBalancerBackendPoolSyncMode_MANUAL {
			poolArgs.SynchronousMode = pulumi.StringPtr("Manual")
		}
		if len(pool.TunnelInterfaces) > 0 {
			tunnels := lb.BackendAddressPoolTunnelInterfaceArray{}
			for _, t := range pool.TunnelInterfaces {
				tunnels = append(tunnels, lb.BackendAddressPoolTunnelInterfaceArgs{
					Identifier: pulumi.Int(int(t.Identifier)),
					Port:       pulumi.Int(int(t.Port)),
					Protocol:   pulumi.String(tunnelProtocolToArm(t.Protocol)),
					Type:       pulumi.String(tunnelTypeToArm(t.Type)),
				})
			}
			poolArgs.TunnelInterfaces = tunnels
		}

		createdPool, err := lb.NewBackendAddressPool(ctx,
			fmt.Sprintf("%s-%s", spec.Name, pool.Name),
			poolArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{loadBalancer}))
		if err != nil {
			return errors.Wrapf(err, "failed to create backend pool %s", pool.Name)
		}
		backendPools[pool.Name] = createdPool
		poolIds[pool.Name] = createdPool.ID().ToStringOutput()

		// IP-based backend members: appliances/servers addressed by IP
		// (REGIONAL tier) or regional load balancer frontends (GLOBAL
		// tier). NIC-based members never appear here -- they associate
		// from the NIC side.
		for _, addr := range pool.Addresses {
			addressArgs := &lb.BackendAddressPoolAddressArgs{
				Name:                 pulumi.String(addr.Name),
				BackendAddressPoolId: createdPool.ID(),
			}
			if addr.IpAddress != "" {
				addressArgs.IpAddress = pulumi.StringPtr(addr.IpAddress)
				addressArgs.VirtualNetworkId = pulumi.StringPtr(pool.VirtualNetworkId.GetValue())
			}
			if addr.LoadBalancerFrontendIpConfigurationId != "" {
				addressArgs.BackendAddressIpConfigurationId = pulumi.StringPtr(addr.LoadBalancerFrontendIpConfigurationId)
			}
			if _, err := lb.NewBackendAddressPoolAddress(ctx,
				fmt.Sprintf("%s-%s-%s", spec.Name, pool.Name, addr.Name),
				addressArgs,
				pulumi.Provider(azureProvider),
				pulumi.DependsOn([]pulumi.Resource{createdPool})); err != nil {
				return errors.Wrapf(err, "failed to create backend pool address %s/%s", pool.Name, addr.Name)
			}
		}
	}

	// Health probes. probe_threshold is the flap dampener: consecutive
	// successes required before a recovered instance is re-admitted.
	// The optional dials are presence-guarded: on iac-input paths that
	// bypass the manifest loader an unset field arrives as nil, and a
	// bare getter's zero would fail the provider's range validation --
	// the fallbacks are the proto defaults, matching the Terraform
	// module's optional(number, N) encodings.
	probes := make(map[string]*lb.Probe)
	probeIds := pulumi.StringMap{}
	for _, probe := range spec.HealthProbes {
		probeArgs := &lb.ProbeArgs{
			Name:              pulumi.String(probe.Name),
			LoadbalancerId:    loadBalancer.ID(),
			Protocol:          pulumi.String(probeProtocolToArm(probe.Protocol)),
			Port:              pulumi.Int(int(probe.Port)),
			IntervalInSeconds: pulumi.Int(int(optionalInt32(probe.IntervalInSeconds, 15))),
			NumberOfProbes:    pulumi.Int(int(optionalInt32(probe.NumberOfProbes, 2))),
			ProbeThreshold:    pulumi.Int(int(optionalInt32(probe.ProbeThreshold, 1))),
		}
		if probe.RequestPath != "" {
			probeArgs.RequestPath = pulumi.String(probe.RequestPath)
		}

		createdProbe, err := lb.NewProbe(ctx,
			fmt.Sprintf("%s-%s", spec.Name, probe.Name),
			probeArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{loadBalancer}))
		if err != nil {
			return errors.Wrapf(err, "failed to create health probe %s", probe.Name)
		}
		probes[probe.Name] = createdProbe
		probeIds[probe.Name] = createdProbe.ID().ToStringOutput()
	}

	// Load-balancing rules. Pools and probes are resolved by name (a
	// stale name is caught by spec-level validation before it reaches
	// here); the frontend defaults to the sole declared frontend when
	// the rule omits it.
	for _, rule := range spec.Rules {
		poolIdList := pulumi.StringArray{}
		dependencies := []pulumi.Resource{loadBalancer}
		for _, poolName := range rule.BackendPoolNames {
			pool, ok := backendPools[poolName]
			if !ok {
				return errors.Errorf("rule %s references unknown backend pool %s", rule.Name, poolName)
			}
			poolIdList = append(poolIdList, pool.ID().ToStringOutput())
			dependencies = append(dependencies, pool)
		}

		ruleArgs := &lb.RuleArgs{
			Name:                        pulumi.String(rule.Name),
			LoadbalancerId:              loadBalancer.ID(),
			FrontendIpConfigurationName: pulumi.String(frontendNameFor(rule.FrontendIpConfigurationName)),
			Protocol:                    pulumi.String(transportProtocolToArm(rule.Protocol)),
			FrontendPort:                pulumi.Int(int(rule.FrontendPort)),
			BackendPort:                 pulumi.Int(int(rule.BackendPort)),
			BackendAddressPoolIds:       poolIdList,
			IdleTimeoutInMinutes:        pulumi.Int(int(optionalInt32(rule.IdleTimeoutInMinutes, 4))),
			FloatingIpEnabled:           pulumi.Bool(rule.GetFloatingIpEnabled()),
			TcpResetEnabled:             pulumi.Bool(rule.GetTcpResetEnabled()),
			DisableOutboundSnat:         pulumi.Bool(rule.GetDisableOutboundSnat()),
		}
		if rule.ProbeName != "" {
			probe, ok := probes[rule.ProbeName]
			if !ok {
				return errors.Errorf("rule %s references unknown health probe %s", rule.Name, rule.ProbeName)
			}
			ruleArgs.ProbeId = probe.ID()
			dependencies = append(dependencies, probe)
		}
		if rule.LoadDistribution == azureloadbalancerv1alpha1.AzureLoadBalancerLoadDistribution_SOURCE_IP {
			ruleArgs.LoadDistribution = pulumi.StringPtr("SourceIP")
		} else if rule.LoadDistribution == azureloadbalancerv1alpha1.AzureLoadBalancerLoadDistribution_SOURCE_IP_PROTOCOL {
			ruleArgs.LoadDistribution = pulumi.StringPtr("SourceIPProtocol")
		} else if rule.LoadDistribution == azureloadbalancerv1alpha1.AzureLoadBalancerLoadDistribution_DEFAULT {
			ruleArgs.LoadDistribution = pulumi.StringPtr("Default")
		}

		if _, err := lb.NewRule(ctx,
			fmt.Sprintf("%s-%s", spec.Name, rule.Name),
			ruleArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn(dependencies)); err != nil {
			return errors.Wrapf(err, "failed to create load balancing rule %s", rule.Name)
		}
	}

	// Inbound NAT rules. Single-target mode (frontend_port) leaves
	// attachment to the member side (a NIC's NAT-rule association
	// referencing the exported rule ID); pool-style mode (port range +
	// pool) gives every pool member its own frontend port.
	natRuleIds := pulumi.StringMap{}
	for _, natRule := range spec.NatRules {
		natArgs := &lb.NatRuleArgs{
			Name:                        pulumi.String(natRule.Name),
			ResourceGroupName:           pulumi.String(locals.ResourceGroupName),
			LoadbalancerId:              loadBalancer.ID(),
			FrontendIpConfigurationName: pulumi.String(frontendNameFor(natRule.FrontendIpConfigurationName)),
			Protocol:                    pulumi.String(transportProtocolToArm(natRule.Protocol)),
			BackendPort:                 pulumi.Int(int(natRule.BackendPort)),
			FloatingIpEnabled:           pulumi.Bool(natRule.GetFloatingIpEnabled()),
			TcpResetEnabled:             pulumi.Bool(natRule.GetTcpResetEnabled()),
			IdleTimeoutInMinutes:        pulumi.Int(int(optionalInt32(natRule.IdleTimeoutInMinutes, 4))),
		}
		dependencies := []pulumi.Resource{loadBalancer}
		// Exactly one mode is set (spec-level validation): send only that
		// mode's fields so the provider sees a clean single-target or
		// pool-style rule.
		if natRule.FrontendPort > 0 {
			natArgs.FrontendPort = pulumi.IntPtr(int(natRule.FrontendPort))
		}
		if natRule.BackendPoolName != "" {
			pool, ok := backendPools[natRule.BackendPoolName]
			if !ok {
				return errors.Errorf("NAT rule %s references unknown backend pool %s", natRule.Name, natRule.BackendPoolName)
			}
			natArgs.BackendAddressPoolId = pool.ID()
			natArgs.FrontendPortStart = pulumi.IntPtr(int(natRule.FrontendPortStart))
			natArgs.FrontendPortEnd = pulumi.IntPtr(int(natRule.FrontendPortEnd))
			dependencies = append(dependencies, pool)
		}

		createdNatRule, err := lb.NewNatRule(ctx,
			fmt.Sprintf("%s-%s", spec.Name, natRule.Name),
			natArgs,
			pulumi.Provider(azureProvider),
			pulumi.DependsOn(dependencies))
		if err != nil {
			return errors.Wrapf(err, "failed to create inbound NAT rule %s", natRule.Name)
		}
		natRuleIds[natRule.Name] = createdNatRule.ID().ToStringOutput()
	}

	// Outbound rules: explicit SNAT through public frontends. Combine
	// with disable_outbound_snat on the load-balancing rules that share
	// the pool.
	for _, outboundRule := range spec.OutboundRules {
		pool, ok := backendPools[outboundRule.BackendPoolName]
		if !ok {
			return errors.Errorf("outbound rule %s references unknown backend pool %s", outboundRule.Name, outboundRule.BackendPoolName)
		}

		outboundFrontends := lb.OutboundRuleFrontendIpConfigurationArray{}
		for _, frontendName := range outboundRule.FrontendIpConfigurationNames {
			outboundFrontends = append(outboundFrontends, lb.OutboundRuleFrontendIpConfigurationArgs{
				Name: pulumi.String(frontendName),
			})
		}

		if _, err := lb.NewOutboundRule(ctx,
			fmt.Sprintf("%s-%s", spec.Name, outboundRule.Name),
			&lb.OutboundRuleArgs{
				Name:                     pulumi.String(outboundRule.Name),
				LoadbalancerId:           loadBalancer.ID(),
				BackendAddressPoolId:     pool.ID(),
				Protocol:                 pulumi.String(transportProtocolToArm(outboundRule.Protocol)),
				AllocatedOutboundPorts:   pulumi.IntPtr(int(optionalInt32(outboundRule.AllocatedOutboundPorts, 1024))),
				TcpResetEnabled:          pulumi.Bool(outboundRule.GetTcpResetEnabled()),
				IdleTimeoutInMinutes:     pulumi.IntPtr(int(optionalInt32(outboundRule.IdleTimeoutInMinutes, 4))),
				FrontendIpConfigurations: outboundFrontends,
			},
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{loadBalancer, pool})); err != nil {
			return errors.Wrapf(err, "failed to create outbound rule %s", outboundRule.Name)
		}
	}

	// Export outputs. The maps keyed by sub-resource name are the
	// composition seams members reference (backend_pool_ids for pool
	// membership, nat_rule_ids for NIC NAT-rule associations, probe_ids
	// for a scale set's rolling-upgrade health probe).
	ctx.Export(OpLoadBalancerId, loadBalancer.ID())
	ctx.Export(OpLoadBalancerName, loadBalancer.Name)

	// The first internal frontend's private address (empty when every
	// frontend is public -- a public frontend's address lives on its
	// referenced AzurePublicIp resource). The SDK types this as a plain
	// StringOutput that carries "" when absent, so it exports directly.
	ctx.Export(OpPrivateIpAddress, loadBalancer.PrivateIpAddress)
	ctx.Export(OpPrivateIpAddresses, loadBalancer.PrivateIpAddresses)

	ctx.Export(OpFrontendIpConfigurationIds, loadBalancer.FrontendIpConfigurations.ApplyT(func(frontends []lb.LoadBalancerFrontendIpConfiguration) map[string]string {
		ids := make(map[string]string, len(frontends))
		for _, f := range frontends {
			if f.Id != nil {
				ids[f.Name] = *f.Id
			}
		}
		return ids
	}))
	ctx.Export(OpBackendPoolIds, poolIds)
	ctx.Export(OpProbeIds, probeIds)
	ctx.Export(OpNatRuleIds, natRuleIds)

	return nil
}

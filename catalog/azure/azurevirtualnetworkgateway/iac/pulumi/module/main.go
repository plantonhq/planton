package module

import (
	"github.com/pkg/errors"
	azurevirtualnetworkgatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualnetworkgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVirtualNetworkGateway.Spec

	// PARITY-EXCEPTION: the classic Pulumi SDK does not expose the
	// ER_GW_SCALE autoscale bounds (minimum/maximum scale units), so a
	// manifest that sets them deploys via the Terraform engine only.
	// Failing loudly here beats silently dropping an autoscale contract
	// the user wrote down.
	if spec.MinimumScaleUnit != nil || spec.MaximumScaleUnit != nil {
		return errors.New("the Pulumi engine cannot express ER_GW_SCALE autoscale bounds " +
			"(minimum_scale_unit/maximum_scale_unit) -- deploy this gateway with the Terraform engine")
	}

	gatewayArgs := &network.VirtualNetworkGatewayArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// Type, routing model, and SKU are always sent explicitly
		// (Vpn/RouteBased when unspecified) -- deterministic payloads on
		// both engines. The SKU cross-checks (type/generation pairing)
		// are spec-validated.
		Type:    pulumi.String(typeWireValue(spec.Type)),
		VpnType: pulumi.String(vpnTypeWireValue(spec.VpnType)),
		Sku:     pulumi.String(skuWireValue(spec.Sku)),
		Tags:    pulumi.ToStringMap(locals.AzureTags),
	}

	// Sent only when specified -- the provider treats generation as
	// Computed, so omission lets Azure pick the SKU's default.
	if generation := generationWireValue(spec.Generation); generation != "" {
		gatewayArgs.Generation = pulumi.String(generation)
	}

	// The IP configurations: each binds the GatewaySubnet (ARM name
	// contract validated by the provider at plan time) and, on VPN
	// gateways, a public IP (spec-validated pairing). FIXED AT CREATION.
	ipConfigurations := network.VirtualNetworkGatewayIpConfigurationArray{}
	for _, ipConfiguration := range spec.IpConfigurations {
		configurationArgs := &network.VirtualNetworkGatewayIpConfigurationArgs{
			PrivateIpAddressAllocation: pulumi.String(allocationWireValue(ipConfiguration.PrivateIpAddressAllocation)),
			SubnetId:                   pulumi.String(ipConfiguration.SubnetId.GetValue()),
		}
		// The provider defaults an empty name to "vnetGatewayConfig" (the
		// portal's name); sent explicitly so both engines produce an
		// identical payload.
		configurationName := ipConfiguration.Name
		if configurationName == "" {
			configurationName = "vnetGatewayConfig"
		}
		configurationArgs.Name = pulumi.String(configurationName)
		if ipConfiguration.PublicIpAddressId.GetValue() != "" {
			configurationArgs.PublicIpAddressId = pulumi.String(ipConfiguration.PublicIpAddressId.GetValue())
		}
		ipConfigurations = append(ipConfigurations, configurationArgs)
	}
	gatewayArgs.IpConfigurations = ipConfigurations

	if spec.ActiveActive {
		gatewayArgs.ActiveActive = pulumi.Bool(true)
	}
	if spec.PrivateIpAddressEnabled {
		gatewayArgs.PrivateIpAddressEnabled = pulumi.Bool(true)
	}
	if spec.EdgeZone != "" {
		gatewayArgs.EdgeZone = pulumi.String(spec.EdgeZone)
	}

	// BGP: BgpEnabled is the v5-era argument name; the classic SDK
	// carries both it and the deprecated EnableBgp mapped to the same ARM
	// property -- the modern name is used.
	if spec.BgpEnabled {
		gatewayArgs.BgpEnabled = pulumi.Bool(true)
	}
	if spec.BgpSettings != nil {
		bgpSettingsArgs := &network.VirtualNetworkGatewayBgpSettingsArgs{}
		if spec.BgpSettings.Asn > 0 {
			bgpSettingsArgs.Asn = pulumi.Int(int(spec.BgpSettings.Asn))
		}
		if spec.BgpSettings.PeerWeight > 0 {
			bgpSettingsArgs.PeerWeight = pulumi.Int(int(spec.BgpSettings.PeerWeight))
		}
		if len(spec.BgpSettings.PeeringAddresses) > 0 {
			peeringAddresses := network.VirtualNetworkGatewayBgpSettingsPeeringAddressArray{}
			for _, peeringAddress := range spec.BgpSettings.PeeringAddresses {
				peeringAddressArgs := &network.VirtualNetworkGatewayBgpSettingsPeeringAddressArgs{
					ApipaAddresses: pulumi.ToStringArray(peeringAddress.ApipaAddresses),
				}
				if peeringAddress.IpConfigurationName != "" {
					peeringAddressArgs.IpConfigurationName = pulumi.String(peeringAddress.IpConfigurationName)
				}
				peeringAddresses = append(peeringAddresses, peeringAddressArgs)
			}
			bgpSettingsArgs.PeeringAddresses = peeringAddresses
		}
		gatewayArgs.BgpSettings = bgpSettingsArgs
	}

	// Custom routes: prefixes the gateway advertises to every connected
	// client and tunnel beyond the VNet's own space.
	if len(spec.CustomRouteAddressPrefixes) > 0 {
		gatewayArgs.CustomRoute = &network.VirtualNetworkGatewayCustomRouteArgs{
			AddressPrefixes: pulumi.ToStringArray(spec.CustomRouteAddressPrefixes),
		}
	}

	// Forced tunneling: the default-route site. References resolve to the
	// local network gateway's ARM id before the module runs.
	if spec.DefaultLocalNetworkGatewayId.GetValue() != "" {
		gatewayArgs.DefaultLocalNetworkGatewayId = pulumi.String(spec.DefaultLocalNetworkGatewayId.GetValue())
	}

	// Point-to-site: address pool, authentication, protocols, and
	// per-group routing (VPN gateways only -- spec-validated).
	if spec.VpnClientConfiguration != nil {
		gatewayArgs.VpnClientConfiguration = expandVpnClientConfiguration(spec.VpnClientConfiguration)
	}

	// Policy groups for point-to-site segmentation.
	if len(spec.PolicyGroups) > 0 {
		policyGroups := network.VirtualNetworkGatewayPolicyGroupArray{}
		for _, policyGroup := range spec.PolicyGroups {
			policyMembers := network.VirtualNetworkGatewayPolicyGroupPolicyMemberArray{}
			for _, policyMember := range policyGroup.PolicyMembers {
				policyMembers = append(policyMembers, &network.VirtualNetworkGatewayPolicyGroupPolicyMemberArgs{
					Name:  pulumi.String(policyMember.Name),
					Type:  pulumi.String(policyMember.Type),
					Value: pulumi.String(policyMember.Value),
				})
			}
			policyGroups = append(policyGroups, &network.VirtualNetworkGatewayPolicyGroupArgs{
				Name:          pulumi.String(policyGroup.Name),
				PolicyMembers: policyMembers,
				IsDefault:     pulumi.Bool(policyGroup.IsDefault),
				Priority:      pulumi.Int(int(policyGroup.Priority)),
			})
		}
		gatewayArgs.PolicyGroups = policyGroups
	}

	if spec.BgpRouteTranslationForNatEnabled {
		gatewayArgs.BgpRouteTranslationForNatEnabled = pulumi.Bool(true)
	}
	// Sent only when enabled -- ARM rejects the parameter on SKUs/types
	// without DNS forwarding support, so omission is the safe default.
	if spec.DnsForwardingEnabled {
		gatewayArgs.DnsForwardingEnabled = pulumi.Bool(true)
	}
	// Omitted when unset: the provider's default (true) matches the
	// spec's documented default, so both engines send nothing unless the
	// user takes an explicit position.
	if spec.IpSecReplayProtectionEnabled != nil {
		gatewayArgs.IpSecReplayProtectionEnabled = pulumi.Bool(spec.GetIpSecReplayProtectionEnabled())
	}
	if spec.RemoteVnetTrafficEnabled {
		gatewayArgs.RemoteVnetTrafficEnabled = pulumi.Bool(true)
	}
	if spec.VirtualWanTrafficEnabled {
		gatewayArgs.VirtualWanTrafficEnabled = pulumi.Bool(true)
	}

	createdGateway, err := network.NewVirtualNetworkGateway(ctx,
		spec.Name,
		gatewayArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create virtual network gateway %s", spec.Name)
	}

	// The composed NAT rules: standalone ARM children of the gateway,
	// deployed after it. Their ids surface in the nat_rule_ids output so
	// connections can opt into them.
	natRuleIds := pulumi.Map{}
	for _, natRule := range spec.NatRules {
		natRuleArgs := &network.VirtualNetworkGatewayNatRuleArgs{
			Name:                    pulumi.String(natRule.Name),
			ResourceGroupName:       pulumi.String(locals.ResourceGroupName),
			VirtualNetworkGatewayId: createdGateway.ID(),
			Mode:                    pulumi.String(natRuleModeWireValue(natRule.Mode)),
			Type:                    pulumi.String(natRuleTypeWireValue(natRule.Type)),
			ExternalMappings:        expandNatRuleExternalMappings(natRule.ExternalMappings),
			InternalMappings:        expandNatRuleInternalMappings(natRule.InternalMappings),
		}
		if natRule.IpConfigurationId != "" {
			natRuleArgs.IpConfigurationId = pulumi.String(natRule.IpConfigurationId)
		}

		createdNatRule, err := network.NewVirtualNetworkGatewayNatRule(ctx,
			spec.Name+"-"+natRule.Name,
			natRuleArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdGateway))
		if err != nil {
			return errors.Wrapf(err, "failed to create nat rule %s on gateway %s", natRule.Name, spec.Name)
		}
		natRuleIds[natRule.Name] = createdNatRule.ID()
	}

	ctx.Export(OpVirtualNetworkGatewayId, createdGateway.ID())
	ctx.Export(OpVirtualNetworkGatewayName, createdGateway.Name)
	ctx.Export(OpNatRuleIds, natRuleIds)

	return nil
}

// expandVpnClientConfiguration maps the point-to-site block. All three
// authentication families (Entra ID, certificate, RADIUS) map field for
// field; the spec's CEL contracts guarantee coherent combinations.
func expandVpnClientConfiguration(clientConfiguration *azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayVpnClientConfiguration) *network.VirtualNetworkGatewayVpnClientConfigurationArgs {
	configurationArgs := &network.VirtualNetworkGatewayVpnClientConfigurationArgs{
		AddressSpaces: pulumi.ToStringArray(clientConfiguration.AddressSpaces),
	}

	if clientConfiguration.AadTenant != "" {
		configurationArgs.AadTenant = pulumi.String(clientConfiguration.AadTenant)
		configurationArgs.AadAudience = pulumi.String(clientConfiguration.AadAudience)
		configurationArgs.AadIssuer = pulumi.String(clientConfiguration.AadIssuer)
	}

	if len(clientConfiguration.RootCertificates) > 0 {
		rootCertificates := network.VirtualNetworkGatewayVpnClientConfigurationRootCertificateArray{}
		for _, rootCertificate := range clientConfiguration.RootCertificates {
			rootCertificates = append(rootCertificates, &network.VirtualNetworkGatewayVpnClientConfigurationRootCertificateArgs{
				Name:           pulumi.String(rootCertificate.Name),
				PublicCertData: pulumi.String(rootCertificate.PublicCertData),
			})
		}
		configurationArgs.RootCertificates = rootCertificates
	}

	if len(clientConfiguration.RevokedCertificates) > 0 {
		revokedCertificates := network.VirtualNetworkGatewayVpnClientConfigurationRevokedCertificateArray{}
		for _, revokedCertificate := range clientConfiguration.RevokedCertificates {
			revokedCertificates = append(revokedCertificates, &network.VirtualNetworkGatewayVpnClientConfigurationRevokedCertificateArgs{
				Name:       pulumi.String(revokedCertificate.Name),
				Thumbprint: pulumi.String(revokedCertificate.Thumbprint),
			})
		}
		configurationArgs.RevokedCertificates = revokedCertificates
	}

	if clientConfiguration.RadiusServerAddress != "" {
		configurationArgs.RadiusServerAddress = pulumi.String(clientConfiguration.RadiusServerAddress)
		configurationArgs.RadiusServerSecret = pulumi.String(clientConfiguration.RadiusServerSecret.GetValue())
	}

	if len(clientConfiguration.RadiusServers) > 0 {
		radiusServers := network.VirtualNetworkGatewayVpnClientConfigurationRadiusServerArray{}
		for _, radiusServer := range clientConfiguration.RadiusServers {
			radiusServers = append(radiusServers, &network.VirtualNetworkGatewayVpnClientConfigurationRadiusServerArgs{
				Address: pulumi.String(radiusServer.Address),
				Secret:  pulumi.String(radiusServer.Secret.GetValue()),
				Score:   pulumi.Int(int(radiusServer.Score)),
			})
		}
		configurationArgs.RadiusServers = radiusServers
	}

	if clientConfiguration.IpsecPolicy != nil {
		ipsecPolicy := clientConfiguration.IpsecPolicy
		configurationArgs.IpsecPolicy = &network.VirtualNetworkGatewayVpnClientConfigurationIpsecPolicyArgs{
			DhGroup:               pulumi.String(ipsecPolicy.DhGroup),
			IkeEncryption:         pulumi.String(ipsecPolicy.IkeEncryption),
			IkeIntegrity:          pulumi.String(ipsecPolicy.IkeIntegrity),
			IpsecEncryption:       pulumi.String(ipsecPolicy.IpsecEncryption),
			IpsecIntegrity:        pulumi.String(ipsecPolicy.IpsecIntegrity),
			PfsGroup:              pulumi.String(ipsecPolicy.PfsGroup),
			SaLifetimeInSeconds:   pulumi.Int(int(ipsecPolicy.SaLifetimeSeconds)),
			SaDataSizeInKilobytes: pulumi.Int(int(ipsecPolicy.SaDataSizeKilobytes)),
		}
	}

	if len(clientConfiguration.VpnClientProtocols) > 0 {
		configurationArgs.VpnClientProtocols = pulumi.ToStringArray(clientConfiguration.VpnClientProtocols)
	}
	if len(clientConfiguration.VpnAuthTypes) > 0 {
		configurationArgs.VpnAuthTypes = pulumi.ToStringArray(clientConfiguration.VpnAuthTypes)
	}

	if len(clientConfiguration.ClientConnections) > 0 {
		clientConnections := network.VirtualNetworkGatewayVpnClientConfigurationVirtualNetworkGatewayClientConnectionArray{}
		for _, clientConnection := range clientConfiguration.ClientConnections {
			clientConnections = append(clientConnections, &network.VirtualNetworkGatewayVpnClientConfigurationVirtualNetworkGatewayClientConnectionArgs{
				Name:             pulumi.String(clientConnection.Name),
				PolicyGroupNames: pulumi.ToStringArray(clientConnection.PolicyGroupNames),
				AddressPrefixes:  pulumi.ToStringArray(clientConnection.AddressPrefixes),
			})
		}
		configurationArgs.VirtualNetworkGatewayClientConnections = clientConnections
	}

	return configurationArgs
}

// expandNatRuleExternalMappings maps NAT rule external mappings.
func expandNatRuleExternalMappings(mappings []*azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayNatRuleMapping) network.VirtualNetworkGatewayNatRuleExternalMappingArray {
	result := network.VirtualNetworkGatewayNatRuleExternalMappingArray{}
	for _, mapping := range mappings {
		mappingArgs := &network.VirtualNetworkGatewayNatRuleExternalMappingArgs{
			AddressSpace: pulumi.String(mapping.AddressSpace),
		}
		if mapping.PortRange != "" {
			mappingArgs.PortRange = pulumi.String(mapping.PortRange)
		}
		result = append(result, mappingArgs)
	}
	return result
}

// expandNatRuleInternalMappings maps NAT rule internal mappings.
func expandNatRuleInternalMappings(mappings []*azurevirtualnetworkgatewayv1alpha1.AzureVirtualNetworkGatewayNatRuleMapping) network.VirtualNetworkGatewayNatRuleInternalMappingArray {
	result := network.VirtualNetworkGatewayNatRuleInternalMappingArray{}
	for _, mapping := range mappings {
		mappingArgs := &network.VirtualNetworkGatewayNatRuleInternalMappingArgs{
			AddressSpace: pulumi.String(mapping.AddressSpace),
		}
		if mapping.PortRange != "" {
			mappingArgs.PortRange = pulumi.String(mapping.PortRange)
		}
		result = append(result, mappingArgs)
	}
	return result
}

package module

import (
	"github.com/pkg/errors"
	azurevpngatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevpngateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevpngatewayv1alpha1.AzureVpnGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVpnGateway.Spec

	// Create the Virtual WAN VPN gateway -- the managed site-to-site VPN
	// terminator inside a virtual hub (ARM allows one per hub). The
	// gateway bills from creation (~$0.36/hr per scale unit class) and
	// is a SLOW resource: creates run 30-45 minutes, deletes 10-20.
	// Deleting it requires its connections to be gone first.
	gatewayArgs := &network.VpnGatewayArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		VirtualHubId:      pulumi.String(spec.VirtualHubId.GetValue()),
		// ARM's default ("Microsoft Network") rendered explicitly --
		// ForceNew: changing it replaces the gateway.
		RoutingPreference: pulumi.String(routingPreferenceWireValue(spec.RoutingPreference)),
		// 500 Mbps per unit across the managed active-active pair. The
		// provider's default of 1 rendered explicitly.
		ScaleUnit: pulumi.Int(int(optionalInt32(spec.ScaleUnit, 1))),
		// Only meaningful when NAT rules are configured on BGP-enabled
		// tunnels; off is ARM's default.
		BgpRouteTranslationForNatEnabled: pulumi.Bool(spec.BgpRouteTranslationForNatEnabled),
		Tags:                             pulumi.ToStringMap(locals.AzureTags),
	}

	// asn/peer_weight go in the create; the custom APIPA addresses are
	// the part ARM only accepts AFTER the gateway exists -- the provider
	// applies them in a second call, which is why they update in place
	// while asn/peer_weight are ForceNew.
	if spec.BgpSettings != nil {
		bgpSettingsArgs := &network.VpnGatewayBgpSettingsArgs{
			Asn:        pulumi.Int(int(spec.BgpSettings.Asn)),
			PeerWeight: pulumi.Int(int(spec.BgpSettings.PeerWeight)),
		}
		if spec.BgpSettings.Instance_0BgpPeeringAddress != nil {
			bgpSettingsArgs.Instance0BgpPeeringAddress = &network.VpnGatewayBgpSettingsInstance0BgpPeeringAddressArgs{
				CustomIps: pulumi.ToStringArray(spec.BgpSettings.Instance_0BgpPeeringAddress.CustomIps),
			}
		}
		if spec.BgpSettings.Instance_1BgpPeeringAddress != nil {
			bgpSettingsArgs.Instance1BgpPeeringAddress = &network.VpnGatewayBgpSettingsInstance1BgpPeeringAddressArgs{
				CustomIps: pulumi.ToStringArray(spec.BgpSettings.Instance_1BgpPeeringAddress.CustomIps),
			}
		}
		gatewayArgs.BgpSettings = bgpSettingsArgs
	}

	createdGateway, err := network.NewVpnGateway(ctx,
		spec.Name,
		gatewayArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create vpn gateway %s", spec.Name)
	}

	// The composed NAT rules: standalone ARM children of the gateway,
	// one per spec entry, keyed by name (how overlapping branch address
	// spaces are translated). Tunnels opt in via a connection link's
	// egress/ingress NAT rule id lists. The SDK's constructor is
	// NewVnpGatewayNatRule -- the resource lives at a legacy, typo'd
	// "VnpGatewayNatRule" token but creates the SAME ARM object as
	// azurerm_vpn_gateway_nat_rule; do not "fix" the name.
	natRuleIds := pulumi.Map{}
	for _, natRule := range spec.NatRules {
		externalMappings := network.VnpGatewayNatRuleExternalMappingArray{}
		for _, mapping := range natRule.ExternalMappings {
			mappingArgs := &network.VnpGatewayNatRuleExternalMappingArgs{
				AddressSpace: pulumi.String(mapping.AddressSpace),
			}
			if mapping.PortRange != "" {
				mappingArgs.PortRange = pulumi.String(mapping.PortRange)
			}
			externalMappings = append(externalMappings, mappingArgs)
		}
		internalMappings := network.VnpGatewayNatRuleInternalMappingArray{}
		for _, mapping := range natRule.InternalMappings {
			mappingArgs := &network.VnpGatewayNatRuleInternalMappingArgs{
				AddressSpace: pulumi.String(mapping.AddressSpace),
			}
			if mapping.PortRange != "" {
				mappingArgs.PortRange = pulumi.String(mapping.PortRange)
			}
			internalMappings = append(internalMappings, mappingArgs)
		}

		natRuleArgs := &network.VnpGatewayNatRuleArgs{
			Name:             pulumi.String(natRule.Name),
			VpnGatewayId:     createdGateway.ID(),
			Mode:             pulumi.String(natRuleModeWireValue(natRule.Mode)),
			Type:             pulumi.String(natRuleTypeWireValue(natRule.Type)),
			ExternalMappings: externalMappings,
			InternalMappings: internalMappings,
		}
		// Unspecified applies the rule on both gateway instances (ARM's
		// default) -- omit, not a value.
		switch natRule.IpConfiguration {
		case azurevpngatewayv1alpha1.AzureVpnGatewayNatRuleIpConfiguration_INSTANCE_0:
			natRuleArgs.IpConfigurationId = pulumi.String("Instance0")
		case azurevpngatewayv1alpha1.AzureVpnGatewayNatRuleIpConfiguration_INSTANCE_1:
			natRuleArgs.IpConfigurationId = pulumi.String("Instance1")
		}

		createdNatRule, err := network.NewVnpGatewayNatRule(ctx,
			spec.Name+"-"+natRule.Name,
			natRuleArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdGateway))
		if err != nil {
			return errors.Wrapf(err, "failed to create nat rule %s", natRule.Name)
		}
		natRuleIds[natRule.Name] = createdNatRule.ID()
	}

	// ARM assigns the BGP defaults (ASN 65515) and the instance
	// addresses at creation; republish the facts a branch device's
	// configuration needs.
	bgpAsn := createdGateway.BgpSettings.Asn()
	publicIpAddresses := createdGateway.IpConfigurations.ApplyT(func(ipConfigurations []network.VpnGatewayIpConfiguration) []string {
		addresses := make([]string, 0, len(ipConfigurations))
		for _, ipConfiguration := range ipConfigurations {
			if ipConfiguration.PublicIpAddress != nil {
				addresses = append(addresses, *ipConfiguration.PublicIpAddress)
			}
		}
		return addresses
	}).(pulumi.StringArrayOutput)
	privateIpAddresses := createdGateway.IpConfigurations.ApplyT(func(ipConfigurations []network.VpnGatewayIpConfiguration) []string {
		addresses := make([]string, 0, len(ipConfigurations))
		for _, ipConfiguration := range ipConfigurations {
			if ipConfiguration.PrivateIpAddress != nil {
				addresses = append(addresses, *ipConfiguration.PrivateIpAddress)
			}
		}
		return addresses
	}).(pulumi.StringArrayOutput)

	ctx.Export(OpVpnGatewayId, createdGateway.ID())
	ctx.Export(OpVpnGatewayName, createdGateway.Name)
	ctx.Export(OpBgpAsn, bgpAsn)
	ctx.Export(OpPublicIpAddresses, publicIpAddresses)
	ctx.Export(OpPrivateIpAddresses, privateIpAddresses)
	ctx.Export(OpNatRuleIds, natRuleIds)

	return nil
}

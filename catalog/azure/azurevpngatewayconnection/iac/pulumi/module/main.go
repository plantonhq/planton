package module

import (
	"github.com/pkg/errors"
	azurevpngatewayconnectionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevpngatewayconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevpngatewayconnectionv1alpha1.AzureVpnGatewayConnectionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVpnGatewayConnection.Spec

	// Create the VPN gateway connection -- the tunnel bundle joining one
	// branch (a VPN site) to a hub's VPN gateway. ARM addresses it as a
	// child of the gateway. The object is free and provisions in
	// minutes; each tunnel reaches Connected only when the branch device
	// negotiates (provisioned-is-not-connected).
	connectionArgs := &network.VpnGatewayConnectionArgs{
		Name:            pulumi.String(spec.Name),
		VpnGatewayId:    pulumi.String(spec.VpnGatewayId.GetValue()),
		RemoteVpnSiteId: pulumi.String(spec.RemoteVpnSiteId.GetValue()),
		// Off is ARM's default: the branch keeps its own internet
		// egress. On, the hub advertises 0.0.0.0/0 to this branch.
		InternetSecurityEnabled: pulumi.Bool(spec.InternetSecurityEnabled),
	}

	// Unset routing applies ARM's default behavior: associate with and
	// propagate to the hub's built-in default route table. A configured
	// block carries its association (the spec requires it -- the
	// provider's own contract).
	if spec.Routing != nil {
		routingArgs := &network.VpnGatewayConnectionRoutingArgs{
			AssociatedRouteTable: pulumi.String(spec.Routing.AssociatedRouteTableId.GetValue()),
		}
		if spec.Routing.InboundRouteMapId.GetValue() != "" {
			routingArgs.InboundRouteMapId = pulumi.String(spec.Routing.InboundRouteMapId.GetValue())
		}
		if spec.Routing.OutboundRouteMapId.GetValue() != "" {
			routingArgs.OutboundRouteMapId = pulumi.String(spec.Routing.OutboundRouteMapId.GetValue())
		}
		if spec.Routing.PropagatedRouteTable != nil {
			routingArgs.PropagatedRouteTable = &network.VpnGatewayConnectionRoutingPropagatedRouteTableArgs{
				RouteTableIds: pulumi.ToStringArray(resolveReferences(spec.Routing.PropagatedRouteTable.RouteTableIds)),
				Labels:        pulumi.ToStringArray(spec.Routing.PropagatedRouteTable.Labels),
			}
		}
		connectionArgs.Routing = routingArgs
	}

	// One tunnel per site link being connected. vpn_site_link_id and
	// bgp_enabled are ForceNew on each tunnel; everything else updates
	// in place.
	vpnLinks := network.VpnGatewayConnectionVpnLinkArray{}
	for _, vpnLink := range spec.VpnLinks {
		linkArgs := &network.VpnGatewayConnectionVpnLinkArgs{
			Name:          pulumi.String(vpnLink.Name),
			VpnSiteLinkId: pulumi.String(vpnLink.VpnSiteLinkId.GetValue()),
			// ARM's defaults rendered explicitly -- mirroring the
			// Terraform module's null handling.
			BandwidthMbps:  pulumi.Int(int(optionalInt32(vpnLink.BandwidthMbps, 10))),
			Protocol:       pulumi.String(protocolWireValue(vpnLink.Protocol)),
			ConnectionMode: pulumi.String(connectionModeWireValue(vpnLink.ConnectionMode)),
			RouteWeight:    pulumi.Int(int(vpnLink.RouteWeight)),

			BgpEnabled:                        pulumi.Bool(vpnLink.BgpEnabled),
			RatelimitEnabled:                  pulumi.Bool(vpnLink.RatelimitEnabled),
			LocalAzureIpAddressEnabled:        pulumi.Bool(vpnLink.LocalAzureIpAddressEnabled),
			PolicyBasedTrafficSelectorEnabled: pulumi.Bool(vpnLink.PolicyBasedTrafficSelectorEnabled),
		}

		// Omitted (not 0) when unset -- ARM's default is 45 seconds.
		if vpnLink.DpdTimeoutSeconds != nil {
			linkArgs.DpdTimeoutSeconds = pulumi.Int(int(*vpnLink.DpdTimeoutSeconds))
		}

		// Omit to let Azure generate a key. Sensitive: never logged.
		if vpnLink.SharedKey.GetValue() != "" {
			linkArgs.SharedKey = pulumi.String(vpnLink.SharedKey.GetValue())
		}

		if len(vpnLink.EgressNatRuleIds) > 0 {
			linkArgs.EgressNatRuleIds = pulumi.ToStringArray(resolveReferences(vpnLink.EgressNatRuleIds))
		}
		if len(vpnLink.IngressNatRuleIds) > 0 {
			linkArgs.IngressNatRuleIds = pulumi.ToStringArray(resolveReferences(vpnLink.IngressNatRuleIds))
		}

		// The spec requires every field of a configured proposal (the
		// provider's contract -- no partial pinning).
		if len(vpnLink.IpsecPolicies) > 0 {
			ipsecPolicies := network.VpnGatewayConnectionVpnLinkIpsecPolicyArray{}
			for _, ipsecPolicy := range vpnLink.IpsecPolicies {
				ipsecPolicies = append(ipsecPolicies, &network.VpnGatewayConnectionVpnLinkIpsecPolicyArgs{
					SaLifetimeSec:          pulumi.Int(int(ipsecPolicy.SaLifetimeSec)),
					SaDataSizeKb:           pulumi.Int(int(ipsecPolicy.SaDataSizeKb)),
					EncryptionAlgorithm:    pulumi.String(ipsecPolicy.EncryptionAlgorithm),
					IntegrityAlgorithm:     pulumi.String(ipsecPolicy.IntegrityAlgorithm),
					IkeEncryptionAlgorithm: pulumi.String(ipsecPolicy.IkeEncryptionAlgorithm),
					IkeIntegrityAlgorithm:  pulumi.String(ipsecPolicy.IkeIntegrityAlgorithm),
					DhGroup:                pulumi.String(ipsecPolicy.DhGroup),
					PfsGroup:               pulumi.String(ipsecPolicy.PfsGroup),
				})
			}
			linkArgs.IpsecPolicies = ipsecPolicies
		}

		// Which of the gateway's custom APIPA addresses each instance
		// peers from on this tunnel ("Instance0"/"Instance1" are ARM's
		// identifiers -- the spec validates the vocabulary).
		if len(vpnLink.CustomBgpAddresses) > 0 {
			customBgpAddresses := network.VpnGatewayConnectionVpnLinkCustomBgpAddressArray{}
			for _, customBgpAddress := range vpnLink.CustomBgpAddresses {
				customBgpAddresses = append(customBgpAddresses, &network.VpnGatewayConnectionVpnLinkCustomBgpAddressArgs{
					IpAddress:         pulumi.String(customBgpAddress.IpAddress),
					IpConfigurationId: pulumi.String(customBgpAddress.IpConfigurationId),
				})
			}
			linkArgs.CustomBgpAddresses = customBgpAddresses
		}

		vpnLinks = append(vpnLinks, linkArgs)
	}
	connectionArgs.VpnLinks = vpnLinks

	// Most connections leave this empty (routing comes from the site's
	// prefixes or BGP).
	if len(spec.TrafficSelectorPolicies) > 0 {
		trafficSelectorPolicies := network.VpnGatewayConnectionTrafficSelectorPolicyArray{}
		for _, trafficSelectorPolicy := range spec.TrafficSelectorPolicies {
			trafficSelectorPolicies = append(trafficSelectorPolicies, &network.VpnGatewayConnectionTrafficSelectorPolicyArgs{
				LocalAddressRanges:  pulumi.ToStringArray(trafficSelectorPolicy.LocalAddressCidrs),
				RemoteAddressRanges: pulumi.ToStringArray(trafficSelectorPolicy.RemoteAddressCidrs),
			})
		}
		connectionArgs.TrafficSelectorPolicies = trafficSelectorPolicies
	}

	createdConnection, err := network.NewVpnGatewayConnection(ctx,
		spec.Name,
		connectionArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create vpn gateway connection %s", spec.Name)
	}

	ctx.Export(OpConnectionId, createdConnection.ID())
	ctx.Export(OpConnectionName, createdConnection.Name)

	return nil
}

// resolveReferences unwraps a list of StringValueOrRef fields into their
// middleware-resolved literal values.
func resolveReferences(references []*foreignkeyv1.StringValueOrRef) []string {
	values := make([]string, 0, len(references))
	for _, reference := range references {
		values = append(values, reference.GetValue())
	}
	return values
}

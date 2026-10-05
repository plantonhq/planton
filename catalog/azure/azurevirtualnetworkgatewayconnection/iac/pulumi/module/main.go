package module

import (
	"github.com/pkg/errors"
	azurevirtualnetworkgatewayconnectionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualnetworkgatewayconnection/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevirtualnetworkgatewayconnectionv1alpha1.AzureVirtualNetworkGatewayConnectionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVirtualNetworkGatewayConnection.Spec

	// PARITY-EXCEPTION: the classic Pulumi SDK models exactly ONE traffic
	// selector policy where the provider accepts a list. Failing loudly
	// beats silently dropping selectors the user wrote down -- manifests
	// needing several deploy via the Terraform engine.
	if len(spec.TrafficSelectorPolicies) > 1 {
		return errors.New("the Pulumi engine supports at most one traffic_selector_policy on a " +
			"gateway connection -- deploy this connection with the Terraform engine")
	}

	connectionArgs := &network.VirtualNetworkGatewayConnectionArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// The type decides the required far side (spec-validated); sent
		// explicitly. References resolve to ARM ids before the module
		// runs.
		Type:                    pulumi.String(typeWireValue(spec.Type)),
		VirtualNetworkGatewayId: pulumi.String(spec.VirtualNetworkGatewayId.GetValue()),
		// Sent explicitly (Default when unspecified) -- deterministic
		// payloads on both engines.
		ConnectionMode: pulumi.String(modeWireValue(spec.ConnectionMode)),
		Tags:           pulumi.ToStringMap(locals.AzureTags),
	}

	// The far side, by connection type (spec-validated pairings).
	if spec.LocalNetworkGatewayId.GetValue() != "" {
		connectionArgs.LocalNetworkGatewayId = pulumi.String(spec.LocalNetworkGatewayId.GetValue())
	}
	if spec.PeerVirtualNetworkGatewayId.GetValue() != "" {
		connectionArgs.PeerVirtualNetworkGatewayId = pulumi.String(spec.PeerVirtualNetworkGatewayId.GetValue())
	}
	if spec.ExpressRouteCircuitId.GetValue() != "" {
		connectionArgs.ExpressRouteCircuitId = pulumi.String(spec.ExpressRouteCircuitId.GetValue())
	}

	// The pre-shared key: omitted when unset so Azure generates one
	// (readable back from the connection's shared-key API). Secrets are
	// reference-resolved at deploy time -- never stored in manifests.
	if spec.SharedKey.GetValue() != "" {
		connectionArgs.SharedKey = pulumi.String(spec.SharedKey.GetValue())
	}
	if spec.AuthorizationKey.GetValue() != "" {
		connectionArgs.AuthorizationKey = pulumi.String(spec.AuthorizationKey.GetValue())
	}

	// BGP: BgpEnabled is the v5-era argument name; the classic SDK
	// carries both it and the deprecated EnableBgp mapped to the same ARM
	// property -- the modern name is used.
	if spec.BgpEnabled {
		connectionArgs.BgpEnabled = pulumi.Bool(true)
	}
	if spec.CustomBgpAddresses != nil {
		customBgpAddressesArgs := &network.VirtualNetworkGatewayConnectionCustomBgpAddressesArgs{
			Primary: pulumi.String(spec.CustomBgpAddresses.Primary),
		}
		if spec.CustomBgpAddresses.Secondary != "" {
			customBgpAddressesArgs.Secondary = pulumi.String(spec.CustomBgpAddresses.Secondary)
		}
		connectionArgs.CustomBgpAddresses = customBgpAddressesArgs
	}

	if spec.DpdTimeoutSeconds != nil {
		connectionArgs.DpdTimeoutSeconds = pulumi.Int(int(spec.GetDpdTimeoutSeconds()))
	}
	// Sent only when specified -- the provider treats the protocol as
	// Computed, so omission lets Azure apply its default (IKEv2).
	if protocol := protocolWireValue(spec.ConnectionProtocol); protocol != "" {
		connectionArgs.ConnectionProtocol = pulumi.String(protocol)
	}
	if spec.RoutingWeight != nil {
		connectionArgs.RoutingWeight = pulumi.Int(int(spec.GetRoutingWeight()))
	}

	// Gateway NAT rules this connection opts into, by ARM id (the owning
	// gateway publishes them in its nat_rule_ids output).
	if len(spec.EgressNatRuleIds) > 0 {
		egressNatRuleIds := pulumi.StringArray{}
		for _, natRuleId := range spec.EgressNatRuleIds {
			egressNatRuleIds = append(egressNatRuleIds, pulumi.String(natRuleId.GetValue()))
		}
		connectionArgs.EgressNatRuleIds = egressNatRuleIds
	}
	if len(spec.IngressNatRuleIds) > 0 {
		ingressNatRuleIds := pulumi.StringArray{}
		for _, natRuleId := range spec.IngressNatRuleIds {
			ingressNatRuleIds = append(ingressNatRuleIds, pulumi.String(natRuleId.GetValue()))
		}
		connectionArgs.IngressNatRuleIds = ingressNatRuleIds
	}

	if spec.UsePolicyBasedTrafficSelectors {
		connectionArgs.UsePolicyBasedTrafficSelectors = pulumi.Bool(true)
	}
	if spec.ExpressRouteGatewayBypass {
		connectionArgs.ExpressRouteGatewayBypass = pulumi.Bool(true)
	}
	if spec.PrivateLinkFastPathEnabled {
		connectionArgs.PrivateLinkFastPathEnabled = pulumi.Bool(true)
	}
	if spec.LocalAzureIpAddressEnabled {
		connectionArgs.LocalAzureIpAddressEnabled = pulumi.Bool(true)
	}

	if len(spec.TrafficSelectorPolicies) == 1 {
		trafficSelectorPolicy := spec.TrafficSelectorPolicies[0]
		connectionArgs.TrafficSelectorPolicy = &network.VirtualNetworkGatewayConnectionTrafficSelectorPolicyArgs{
			LocalAddressCidrs:  pulumi.ToStringArray(trafficSelectorPolicy.LocalAddressCidrs),
			RemoteAddressCidrs: pulumi.ToStringArray(trafficSelectorPolicy.RemoteAddressCidrs),
		}
	}

	// The custom IPsec/IKE proposal: mapped field for field; the SA
	// bounds are omitted when unset so Azure fills its defaults.
	if spec.IpsecPolicy != nil {
		ipsecPolicyArgs := &network.VirtualNetworkGatewayConnectionIpsecPolicyArgs{
			DhGroup:         pulumi.String(spec.IpsecPolicy.DhGroup),
			IkeEncryption:   pulumi.String(spec.IpsecPolicy.IkeEncryption),
			IkeIntegrity:    pulumi.String(spec.IpsecPolicy.IkeIntegrity),
			IpsecEncryption: pulumi.String(spec.IpsecPolicy.IpsecEncryption),
			IpsecIntegrity:  pulumi.String(spec.IpsecPolicy.IpsecIntegrity),
			PfsGroup:        pulumi.String(spec.IpsecPolicy.PfsGroup),
		}
		if spec.IpsecPolicy.SaDatasize != nil {
			ipsecPolicyArgs.SaDatasize = pulumi.Int(int(spec.IpsecPolicy.GetSaDatasize()))
		}
		if spec.IpsecPolicy.SaLifetime != nil {
			ipsecPolicyArgs.SaLifetime = pulumi.Int(int(spec.IpsecPolicy.GetSaLifetime()))
		}
		connectionArgs.IpsecPolicy = ipsecPolicyArgs
	}

	createdConnection, err := network.NewVirtualNetworkGatewayConnection(ctx,
		spec.Name,
		connectionArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create virtual network gateway connection %s", spec.Name)
	}

	ctx.Export(OpConnectionId, createdConnection.ID())
	ctx.Export(OpConnectionName, createdConnection.Name)

	return nil
}

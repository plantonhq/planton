package module

import (
	"github.com/pkg/errors"
	azurefirewallv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefirewall/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefirewallv1alpha1.AzureFirewallIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFirewall.Spec

	firewallArgs := &network.FirewallArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// The deployment model and tier are always sent explicitly
		// (AZFW_VNet/Standard when unspecified) -- deterministic payloads
		// on both engines. Both are ARM-required; the tier must match the
		// attached policy's tier (ARM validates the pairing).
		SkuName: pulumi.String(skuNameWireValue(spec.SkuName)),
		SkuTier: pulumi.String(skuTierWireValue(spec.SkuTier)),
		Tags:    pulumi.ToStringMap(locals.AzureTags),
	}

	// The data-path IP configurations: exactly one carries the
	// AzureFirewallSubnet subnet (spec-validated); additional blocks add
	// public IPs (each adds SNAT ports and a DNAT frontend). The subnet
	// name/size contract (exactly "AzureFirewallSubnet", /26+) is ARM's;
	// the referenced AzureSubnet must be created to it.
	if len(spec.IpConfigurations) > 0 {
		ipConfigurations := network.FirewallIpConfigurationArray{}
		for _, ipConfiguration := range spec.IpConfigurations {
			configurationArgs := &network.FirewallIpConfigurationArgs{
				Name: pulumi.String(ipConfiguration.Name),
			}
			if ipConfiguration.SubnetId.GetValue() != "" {
				configurationArgs.SubnetId = pulumi.String(ipConfiguration.SubnetId.GetValue())
			}
			if ipConfiguration.PublicIpAddressId.GetValue() != "" {
				configurationArgs.PublicIpAddressId = pulumi.String(ipConfiguration.PublicIpAddressId.GetValue())
			}
			ipConfigurations = append(ipConfigurations, configurationArgs)
		}
		firewallArgs.IpConfigurations = ipConfigurations
	}

	// The management path (forced tunneling / BASIC tier): its own
	// dedicated subnet ("AzureFirewallManagementSubnet", /26+) and a
	// REQUIRED public IP. ForceNew -- adding or removing it replaces the
	// firewall, which is why it is modeled as a distinct block rather
	// than folded into ip_configurations.
	if spec.ManagementIpConfiguration != nil {
		firewallArgs.ManagementIpConfiguration = &network.FirewallManagementIpConfigurationArgs{
			Name:              pulumi.String(spec.ManagementIpConfiguration.Name),
			SubnetId:          pulumi.String(spec.ManagementIpConfiguration.SubnetId.GetValue()),
			PublicIpAddressId: pulumi.String(spec.ManagementIpConfiguration.PublicIpAddressId.GetValue()),
		}
	}

	// The policy attachment: rules, threat intelligence, TLS inspection,
	// and IDPS all live on the policy; the firewall is the enforcement
	// point. Classic inline rule collections are deliberately not modeled
	// -- policy-based management is Azure's direction and ARM rejects
	// mixing the two.
	if spec.FirewallPolicyId.GetValue() != "" {
		firewallArgs.FirewallPolicyId = pulumi.String(spec.FirewallPolicyId.GetValue())
	}

	// Sent only when specified: the ARM field is server-defaulted (Alert)
	// and Computed in the provider, so omission lets Azure own the
	// default instead of the module guessing it.
	if mode := threatIntelModeWireValue(spec.ThreatIntelMode); mode != "" {
		firewallArgs.ThreatIntelMode = pulumi.String(mode)
	}

	// DNS: setting servers implicitly forces the DNS proxy ON in Azure's
	// wire encoding (the provider couples them); dns_proxy_enabled alone
	// enables proxying without custom upstreams. Both are passed through
	// verbatim so the coupling stays Azure's, not the module's.
	if len(spec.DnsServers) > 0 {
		firewallArgs.DnsServers = pulumi.ToStringArray(spec.DnsServers)
	}
	if spec.DnsProxyEnabled {
		firewallArgs.DnsProxyEnabled = pulumi.Bool(true)
	}

	// SNAT ranges: CIDRs or the literal "IANAPrivateRanges" token. Sent
	// only when the user overrides Azure's IANA-private default.
	if len(spec.PrivateIpRanges) > 0 {
		firewallArgs.PrivateIpRanges = pulumi.ToStringArray(spec.PrivateIpRanges)
	}

	// The Virtual WAN hub deployment target (AZFW_HUB model only --
	// spec-validated pairing). Azure manages the hub firewall's
	// addressing and surfaces it through outputs.
	if spec.VirtualHub != nil {
		virtualHubArgs := &network.FirewallVirtualHubArgs{
			VirtualHubId: pulumi.String(spec.VirtualHub.VirtualHubId.GetValue()),
		}
		if spec.VirtualHub.PublicIpCount != nil {
			virtualHubArgs.PublicIpCount = pulumi.Int(int(spec.VirtualHub.GetPublicIpCount()))
		}
		firewallArgs.VirtualHub = virtualHubArgs
	}

	if len(spec.Zones) > 0 {
		firewallArgs.Zones = pulumi.ToStringArray(spec.Zones)
	}

	createdFirewall, err := network.NewFirewall(ctx,
		spec.Name,
		firewallArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create firewall %s", spec.Name)
	}

	ctx.Export(OpFirewallId, createdFirewall.ID())
	ctx.Export(OpFirewallName, createdFirewall.Name)

	// The data-path private IP -- THE hub-spoke seam: spoke route tables
	// send egress here via a VIRTUAL_APPLIANCE next hop. Azure computes it
	// on the subnet-bearing ip configuration; empty for hub firewalls.
	ctx.Export(OpPrivateIpAddress, createdFirewall.IpConfigurations.ApplyT(func(configurations []network.FirewallIpConfiguration) string {
		for _, configuration := range configurations {
			if configuration.PrivateIpAddress != nil && *configuration.PrivateIpAddress != "" {
				return *configuration.PrivateIpAddress
			}
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpManagementPrivateIpAddress, createdFirewall.ManagementIpConfiguration.ApplyT(func(configuration *network.FirewallManagementIpConfiguration) string {
		if configuration == nil || configuration.PrivateIpAddress == nil {
			return ""
		}
		return *configuration.PrivateIpAddress
	}).(pulumi.StringOutput))

	// Hub-firewall addressing is Azure-assigned and only known after
	// deployment -- exported for DNS records and route configuration.
	ctx.Export(OpVirtualHubPublicIpAddresses, createdFirewall.VirtualHub.ApplyT(func(virtualHub *network.FirewallVirtualHub) []string {
		if virtualHub == nil {
			return []string{}
		}
		return virtualHub.PublicIpAddresses
	}).(pulumi.StringArrayOutput))

	ctx.Export(OpVirtualHubPrivateIpAddress, createdFirewall.VirtualHub.ApplyT(func(virtualHub *network.FirewallVirtualHub) string {
		if virtualHub == nil || virtualHub.PrivateIpAddress == nil {
			return ""
		}
		return *virtualHub.PrivateIpAddress
	}).(pulumi.StringOutput))

	return nil
}

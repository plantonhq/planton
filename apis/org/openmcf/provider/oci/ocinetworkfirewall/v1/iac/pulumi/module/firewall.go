package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/networkfirewall"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func firewallResource(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	policy *networkfirewall.NetworkFirewallPolicy,
	securityRuleDeps []pulumi.Resource,
) error {
	spec := locals.OciNetworkFirewall.Spec

	args := &networkfirewall.NetworkFirewallArgs{
		CompartmentId:           pulumi.String(spec.CompartmentId.GetValue()),
		NetworkFirewallPolicyId: policy.ID(),
		SubnetId:                pulumi.String(spec.SubnetId.GetValue()),
		DisplayName:             pulumi.String(locals.DisplayName),
		FreeformTags:            pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.Ipv4Address != "" {
		args.Ipv4address = pulumi.String(spec.Ipv4Address)
	}

	if spec.Ipv6Address != "" {
		args.Ipv6address = pulumi.String(spec.Ipv6Address)
	}

	if spec.AvailabilityDomain != "" {
		args.AvailabilityDomain = pulumi.String(spec.AvailabilityDomain)
	}

	if len(spec.NetworkSecurityGroupIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(spec.NetworkSecurityGroupIds))
		for i, n := range spec.NetworkSecurityGroupIds {
			nsgIds[i] = pulumi.String(n.GetValue())
		}
		args.NetworkSecurityGroupIds = nsgIds
	}

	if spec.NatConfiguration != nil {
		args.NatConfiguration = &networkfirewall.NetworkFirewallNatConfigurationArgs{
			MustEnablePrivateNat: pulumi.Bool(spec.NatConfiguration.MustEnablePrivateNat),
		}
	}

	if spec.Shape != "" {
		args.Shape = pulumi.String(spec.Shape)
	}

	deps := []pulumi.Resource{policy}
	deps = append(deps, securityRuleDeps...)

	createdFirewall, err := networkfirewall.NewNetworkFirewall(
		ctx,
		locals.DisplayName,
		args,
		pulumiOciOpt(provider),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create network firewall")
	}

	ctx.Export(OpFirewallId, createdFirewall.ID())
	ctx.Export(OpIpv4Address, createdFirewall.Ipv4address)

	return nil
}

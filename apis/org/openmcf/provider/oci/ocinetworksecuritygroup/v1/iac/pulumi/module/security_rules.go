package module

import (
	"fmt"

	"github.com/pkg/errors"
	ocinetworksecuritygroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocinetworksecuritygroup/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func securityRules(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, createdNsg *core.NetworkSecurityGroup) error {
	spec := locals.OciNetworkSecurityGroup.Spec

	for i, rule := range spec.IngressRules {
		name := fmt.Sprintf("%s-ingress-%d", locals.DisplayName, i)

		args := &core.NetworkSecurityGroupSecurityRuleArgs{
			NetworkSecurityGroupId: createdNsg.ID(),
			Direction:              pulumi.String("INGRESS"),
			Protocol:               pulumi.String(protocolString(rule.Protocol)),
			Source:                 pulumi.StringPtr(rule.Source),
			SourceType:             pulumi.StringPtr(targetTypeString(rule.SourceType)),
			Stateless:              pulumi.BoolPtr(rule.Stateless),
		}

		if rule.Description != "" {
			args.Description = pulumi.StringPtr(rule.Description)
		}

		if rule.TcpOptions != nil {
			args.TcpOptions = buildTcpOptions(rule.TcpOptions)
		}

		if rule.UdpOptions != nil {
			args.UdpOptions = buildUdpOptions(rule.UdpOptions)
		}

		if rule.IcmpOptions != nil {
			args.IcmpOptions = buildIcmpOptions(rule.IcmpOptions)
		}

		if _, err := core.NewNetworkSecurityGroupSecurityRule(ctx, name, args, pulumiOciOpt(provider)); err != nil {
			return errors.Wrapf(err, "failed to create ingress rule %d", i)
		}
	}

	for i, rule := range spec.EgressRules {
		name := fmt.Sprintf("%s-egress-%d", locals.DisplayName, i)

		args := &core.NetworkSecurityGroupSecurityRuleArgs{
			NetworkSecurityGroupId: createdNsg.ID(),
			Direction:              pulumi.String("EGRESS"),
			Protocol:               pulumi.String(protocolString(rule.Protocol)),
			Destination:            pulumi.StringPtr(rule.Destination),
			DestinationType:        pulumi.StringPtr(targetTypeString(rule.DestinationType)),
			Stateless:              pulumi.BoolPtr(rule.Stateless),
		}

		if rule.Description != "" {
			args.Description = pulumi.StringPtr(rule.Description)
		}

		if rule.TcpOptions != nil {
			args.TcpOptions = buildTcpOptions(rule.TcpOptions)
		}

		if rule.UdpOptions != nil {
			args.UdpOptions = buildUdpOptions(rule.UdpOptions)
		}

		if rule.IcmpOptions != nil {
			args.IcmpOptions = buildIcmpOptions(rule.IcmpOptions)
		}

		if _, err := core.NewNetworkSecurityGroupSecurityRule(ctx, name, args, pulumiOciOpt(provider)); err != nil {
			return errors.Wrapf(err, "failed to create egress rule %d", i)
		}
	}

	return nil
}

func protocolString(p ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_Protocol) string {
	switch p {
	case ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_all:
		return "all"
	case ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_icmp:
		return "1"
	case ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_tcp:
		return "6"
	case ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_udp:
		return "17"
	case ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_icmpv6:
		return "58"
	default:
		return "all"
	}
}

func targetTypeString(t ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_TargetType) string {
	switch t {
	case ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_cidr_block:
		return "CIDR_BLOCK"
	case ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_service_cidr_block:
		return "SERVICE_CIDR_BLOCK"
	case ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_network_security_group:
		return "NETWORK_SECURITY_GROUP"
	default:
		return "CIDR_BLOCK"
	}
}

func buildTcpOptions(opts *ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_TcpOptions) core.NetworkSecurityGroupSecurityRuleTcpOptionsPtrInput {
	tcpArgs := &core.NetworkSecurityGroupSecurityRuleTcpOptionsArgs{}

	if opts.DestinationPortRange != nil {
		tcpArgs.DestinationPortRange = &core.NetworkSecurityGroupSecurityRuleTcpOptionsDestinationPortRangeArgs{
			Min: pulumi.Int(int(opts.DestinationPortRange.Min)),
			Max: pulumi.Int(int(opts.DestinationPortRange.Max)),
		}
	}

	if opts.SourcePortRange != nil {
		tcpArgs.SourcePortRange = &core.NetworkSecurityGroupSecurityRuleTcpOptionsSourcePortRangeArgs{
			Min: pulumi.Int(int(opts.SourcePortRange.Min)),
			Max: pulumi.Int(int(opts.SourcePortRange.Max)),
		}
	}

	return tcpArgs
}

func buildUdpOptions(opts *ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_UdpOptions) core.NetworkSecurityGroupSecurityRuleUdpOptionsPtrInput {
	udpArgs := &core.NetworkSecurityGroupSecurityRuleUdpOptionsArgs{}

	if opts.DestinationPortRange != nil {
		udpArgs.DestinationPortRange = &core.NetworkSecurityGroupSecurityRuleUdpOptionsDestinationPortRangeArgs{
			Min: pulumi.Int(int(opts.DestinationPortRange.Min)),
			Max: pulumi.Int(int(opts.DestinationPortRange.Max)),
		}
	}

	if opts.SourcePortRange != nil {
		udpArgs.SourcePortRange = &core.NetworkSecurityGroupSecurityRuleUdpOptionsSourcePortRangeArgs{
			Min: pulumi.Int(int(opts.SourcePortRange.Min)),
			Max: pulumi.Int(int(opts.SourcePortRange.Max)),
		}
	}

	return udpArgs
}

func buildIcmpOptions(opts *ocinetworksecuritygroupv1.OciNetworkSecurityGroupSpec_IcmpOptions) core.NetworkSecurityGroupSecurityRuleIcmpOptionsPtrInput {
	icmpArgs := &core.NetworkSecurityGroupSecurityRuleIcmpOptionsArgs{
		Type: pulumi.Int(int(opts.Type)),
	}

	if opts.Code != nil {
		icmpArgs.Code = pulumi.IntPtr(int(*opts.Code))
	}

	return icmpArgs
}

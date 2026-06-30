package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func subnet(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, customRouteTableId pulumi.StringOutput) error {
	spec := locals.OciSubnet.Spec

	subnetArgs := &core.SubnetArgs{
		CompartmentId:           pulumi.String(spec.CompartmentId.GetValue()),
		VcnId:                   pulumi.String(spec.VcnId.GetValue()),
		CidrBlock:               pulumi.StringPtr(spec.CidrBlock),
		DisplayName:             pulumi.StringPtr(locals.DisplayName),
		FreeformTags:            pulumi.ToStringMap(locals.FreeformTags),
		ProhibitPublicIpOnVnic:  pulumi.BoolPtr(spec.ProhibitPublicIpOnVnic),
		ProhibitInternetIngress: pulumi.BoolPtr(spec.ProhibitInternetIngress),
	}

	if spec.DnsLabel != "" {
		subnetArgs.DnsLabel = pulumi.StringPtr(spec.DnsLabel)
	}

	if spec.AvailabilityDomain != "" {
		subnetArgs.AvailabilityDomain = pulumi.StringPtr(spec.AvailabilityDomain)
	}

	if spec.DhcpOptionsId != nil {
		subnetArgs.DhcpOptionsId = pulumi.StringPtr(spec.DhcpOptionsId.GetValue())
	}

	if spec.Ipv6CidrBlock != "" {
		subnetArgs.Ipv6cidrBlock = pulumi.StringPtr(spec.Ipv6CidrBlock)
	}

	if len(spec.SecurityListIds) > 0 {
		secListIds := make([]string, len(spec.SecurityListIds))
		for i, ref := range spec.SecurityListIds {
			secListIds[i] = ref.GetValue()
		}
		subnetArgs.SecurityListIds = pulumi.ToStringArray(secListIds)
	}

	if spec.RouteTableId != nil {
		subnetArgs.RouteTableId = pulumi.StringPtr(spec.RouteTableId.GetValue())
	} else if len(spec.RouteRules) > 0 {
		subnetArgs.RouteTableId = customRouteTableId
	}

	createdSubnet, err := core.NewSubnet(ctx, locals.DisplayName, subnetArgs, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci subnet")
	}

	ctx.Export(OpSubnetId, createdSubnet.ID())
	ctx.Export(OpSubnetDomainName, createdSubnet.SubnetDomainName)
	ctx.Export(OpVirtualRouterIp, createdSubnet.VirtualRouterIp)
	ctx.Export(OpVirtualRouterMac, createdSubnet.VirtualRouterMac)
	ctx.Export(OpRouteTableId, createdSubnet.RouteTableId)

	return nil
}

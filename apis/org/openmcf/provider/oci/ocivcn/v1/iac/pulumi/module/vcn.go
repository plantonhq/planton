package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func vcn(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*core.Vcn, error) {
	spec := locals.OciVcn.Spec

	createdVcn, err := core.NewVcn(ctx, locals.DisplayName, &core.VcnArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		CidrBlocks:    pulumi.ToStringArray(spec.CidrBlocks),
		DisplayName:   pulumi.StringPtr(locals.DisplayName),
		DnsLabel:      pulumi.StringPtr(spec.DnsLabel),
		IsIpv6enabled: pulumi.BoolPtr(spec.IsIpv6Enabled),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}, pulumiOciOpt(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create oci vcn")
	}

	ctx.Export(OpVcnId, createdVcn.ID())
	ctx.Export(OpDefaultRouteTableId, createdVcn.DefaultRouteTableId)
	ctx.Export(OpDefaultSecurityListId, createdVcn.DefaultSecurityListId)
	ctx.Export(OpDefaultDhcpOptionsId, createdVcn.DefaultDhcpOptionsId)

	return createdVcn, nil
}

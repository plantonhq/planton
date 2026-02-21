package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func publicIp(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciPublicIp.Spec

	args := &core.PublicIpArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		Lifetime:      pulumi.String(spec.Lifetime),
		DisplayName:   pulumi.StringPtr(locals.DisplayName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.PrivateIpId != nil {
		args.PrivateIpId = pulumi.StringPtr(spec.PrivateIpId.GetValue())
	}

	if spec.PublicIpPoolId != nil {
		args.PublicIpPoolId = pulumi.StringPtr(spec.PublicIpPoolId.GetValue())
	}

	createdPublicIp, err := core.NewPublicIp(ctx, locals.DisplayName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci public ip")
	}

	ctx.Export(OpPublicIpId, createdPublicIp.ID())
	ctx.Export(OpIpAddress, createdPublicIp.IpAddress)

	return nil
}

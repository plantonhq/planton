package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/bastion"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func bastionResource(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciBastion.Spec

	bastionArgs := &bastion.BastionArgs{
		CompartmentId:  pulumi.String(spec.CompartmentId.GetValue()),
		TargetSubnetId: pulumi.String(spec.TargetSubnetId.GetValue()),
		BastionType:    pulumi.String("STANDARD"),
		Name:           pulumi.String(locals.DisplayName),
		FreeformTags:   pulumi.ToStringMap(locals.FreeformTags),
	}

	if len(spec.ClientCidrBlockAllowList) > 0 {
		bastionArgs.ClientCidrBlockAllowLists = pulumi.ToStringArray(spec.ClientCidrBlockAllowList)
	}

	if spec.MaxSessionTtlInSeconds != nil {
		bastionArgs.MaxSessionTtlInSeconds = pulumi.Int(int(*spec.MaxSessionTtlInSeconds))
	}

	if spec.IsDnsProxyEnabled != nil {
		if *spec.IsDnsProxyEnabled {
			bastionArgs.DnsProxyStatus = pulumi.String("ENABLED")
		} else {
			bastionArgs.DnsProxyStatus = pulumi.String("DISABLED")
		}
	}

	createdBastion, err := bastion.NewBastion(ctx, locals.DisplayName, bastionArgs, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create bastion")
	}

	ctx.Export(OpBastionId, createdBastion.ID())
	ctx.Export(OpPrivateEndpointIpAddress, createdBastion.PrivateEndpointIpAddress)

	return nil
}

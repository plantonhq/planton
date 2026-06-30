package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func nsg(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*core.NetworkSecurityGroup, error) {
	spec := locals.OciSecurityGroup.Spec

	createdNsg, err := core.NewNetworkSecurityGroup(ctx, locals.DisplayName, &core.NetworkSecurityGroupArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		VcnId:         pulumi.String(spec.VcnId.GetValue()),
		DisplayName:   pulumi.StringPtr(locals.DisplayName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}, pulumiOciOpt(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create oci network security group")
	}

	ctx.Export(OpNetworkSecurityGroupId, createdNsg.ID())

	return createdNsg, nil
}

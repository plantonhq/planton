package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func natGateway(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	createdVcn *core.Vcn,
) error {
	spec := locals.OciVcn.Spec

	createdNgw, err := core.NewNatGateway(ctx,
		fmt.Sprintf("%s-ngw", locals.DisplayName),
		&core.NatGatewayArgs{
			CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
			VcnId:         createdVcn.ID(),
			DisplayName:   pulumi.StringPtr(fmt.Sprintf("%s-ngw", locals.DisplayName)),
			BlockTraffic:  pulumi.BoolPtr(false),
			FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
		},
		pulumiOciOpt(provider),
		pulumi.Parent(createdVcn),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create nat gateway")
	}

	ctx.Export(OpNatGatewayId, createdNgw.ID())

	return nil
}

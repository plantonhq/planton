package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func internetGateway(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	createdVcn *core.Vcn,
) error {
	spec := locals.OciVcn.Spec

	createdIgw, err := core.NewInternetGateway(ctx,
		fmt.Sprintf("%s-igw", locals.DisplayName),
		&core.InternetGatewayArgs{
			CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
			VcnId:         createdVcn.ID(),
			DisplayName:   pulumi.StringPtr(fmt.Sprintf("%s-igw", locals.DisplayName)),
			Enabled:       pulumi.BoolPtr(true),
			FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
		},
		pulumiOciOpt(provider),
		pulumi.Parent(createdVcn),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create internet gateway")
	}

	ctx.Export(OpInternetGatewayId, createdIgw.ID())

	return nil
}

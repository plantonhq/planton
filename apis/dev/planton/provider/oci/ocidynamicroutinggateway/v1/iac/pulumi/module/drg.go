package module

import (
	"fmt"

	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createDrg(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*core.Drg, error) {
	spec := locals.OciDynamicRoutingGateway.Spec

	createdDrg, err := core.NewDrg(ctx, locals.DisplayName, &core.DrgArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:   pulumi.StringPtr(locals.DisplayName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}, pulumiOciOpt(provider))
	if err != nil {
		return nil, fmt.Errorf("failed to create drg: %w", err)
	}

	ctx.Export(OpDrgId, createdDrg.ID())
	ctx.Export(OpDefaultExportDrgRouteDistributionId, createdDrg.DefaultExportDrgRouteDistributionId)

	return createdDrg, nil
}

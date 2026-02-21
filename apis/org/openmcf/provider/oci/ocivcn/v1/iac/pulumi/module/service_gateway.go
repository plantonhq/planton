package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func serviceGateway(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	createdVcn *core.Vcn,
) error {
	spec := locals.OciVcn.Spec

	// Look up all available OCI services so we can wire the Service Gateway
	// to "All Services in Oracle Services Network" automatically. This saves
	// users from having to know the service OCID, which varies by region.
	allServices, err := core.GetServices(ctx, nil, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to look up oci services for service gateway")
	}

	var serviceEntries core.ServiceGatewayServiceArray
	for _, svc := range allServices.Services {
		serviceEntries = append(serviceEntries, &core.ServiceGatewayServiceArgs{
			ServiceId: pulumi.String(svc.Id),
		})
	}

	createdSgw, err := core.NewServiceGateway(ctx,
		fmt.Sprintf("%s-sgw", locals.DisplayName),
		&core.ServiceGatewayArgs{
			CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
			VcnId:         createdVcn.ID(),
			DisplayName:   pulumi.StringPtr(fmt.Sprintf("%s-sgw", locals.DisplayName)),
			Services:      serviceEntries,
			FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
		},
		pulumiOciOpt(provider),
		pulumi.Parent(createdVcn),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create service gateway")
	}

	ctx.Export(OpServiceGatewayId, createdSgw.ID())

	return nil
}

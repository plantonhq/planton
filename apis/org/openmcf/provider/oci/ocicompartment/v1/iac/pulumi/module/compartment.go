package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/identity"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func compartment(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciCompartment.Spec

	createdCompartment, err := identity.NewCompartment(ctx, locals.Name, &identity.CompartmentArgs{
		CompartmentId: pulumi.StringPtr(spec.CompartmentId.GetValue()),
		Name:          pulumi.StringPtr(locals.Name),
		Description:   pulumi.String(spec.Description),
		EnableDelete:  pulumi.BoolPtr(spec.EnableDelete),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci compartment")
	}

	ctx.Export(OpCompartmentId, createdCompartment.ID())

	return nil
}

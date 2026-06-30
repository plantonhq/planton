package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/identity"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func dynamicGroup(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciDynamicGroup.Spec

	createdDynamicGroup, err := identity.NewDynamicGroup(ctx, locals.Name, &identity.DynamicGroupArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		Name:          pulumi.StringPtr(locals.Name),
		Description:   pulumi.String(spec.Description),
		MatchingRule:  pulumi.String(spec.MatchingRule),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci identity dynamic group")
	}

	ctx.Export(OpDynamicGroupId, createdDynamicGroup.ID())

	return nil
}

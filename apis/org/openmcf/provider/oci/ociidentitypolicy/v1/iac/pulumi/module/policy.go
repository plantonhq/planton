package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/identity"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func policy(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciIdentityPolicy.Spec

	args := &identity.PolicyArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		Name:          pulumi.StringPtr(locals.Name),
		Description:   pulumi.String(spec.Description),
		Statements:    pulumi.ToStringArray(spec.Statements),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.VersionDate != "" {
		args.VersionDate = pulumi.StringPtr(spec.VersionDate)
	}

	createdPolicy, err := identity.NewPolicy(ctx, locals.Name, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci identity policy")
	}

	ctx.Export(OpPolicyId, createdPolicy.ID())

	return nil
}

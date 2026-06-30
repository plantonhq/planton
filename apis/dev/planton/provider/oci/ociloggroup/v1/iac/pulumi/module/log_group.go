package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func logGroupResource(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*logging.LogGroup, error) {
	spec := locals.OciLogGroup.Spec

	args := &logging.LogGroupArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:   pulumi.String(locals.GroupName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	logGroup, err := logging.NewLogGroup(ctx, locals.GroupName, args, pulumiOciOpt(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create log group")
	}

	ctx.Export(OpLogGroupId, logGroup.ID())

	return logGroup, nil
}

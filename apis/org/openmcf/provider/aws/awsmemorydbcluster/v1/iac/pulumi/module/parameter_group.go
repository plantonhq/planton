package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/memorydb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// parameterGroup creates a custom MemoryDB parameter group when inline
// parameters are provided and a parameter_group_family is specified.
func parameterGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*memorydb.ParameterGroup, error) {
	spec := locals.Spec
	if spec == nil || len(spec.Parameters) == 0 || spec.ParameterGroupFamily == "" {
		return nil, nil
	}

	var params memorydb.ParameterGroupParameterArray
	for _, p := range spec.Parameters {
		params = append(params, &memorydb.ParameterGroupParameterArgs{
			Name:  pulumi.String(p.Name),
			Value: pulumi.String(p.Value),
		})
	}

	pg, err := memorydb.NewParameterGroup(ctx, "parameter-group", &memorydb.ParameterGroupArgs{
		Name:        pulumi.Sprintf("%s-custom", locals.Target.Metadata.Id),
		Family:      pulumi.String(spec.ParameterGroupFamily),
		Description: pulumi.Sprintf("Custom parameter group for %s", locals.Target.Metadata.Id),
		Parameters:  params,
		Tags:        pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create parameter group")
	}

	ctx.Export(OpParameterGroupName, pg.Name)
	return pg, nil
}

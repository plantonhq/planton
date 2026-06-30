package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/redshift"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// parameterGroup creates a Redshift Parameter Group when inline parameters are provided.
func parameterGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*redshift.ParameterGroup, error) {
	spec := locals.AwsRedshiftCluster.Spec
	if spec == nil {
		return nil, nil
	}

	if spec.ClusterParameterGroupName != "" && len(spec.Parameters) == 0 {
		return nil, nil
	}

	if len(spec.Parameters) == 0 {
		return nil, nil
	}

	var params redshift.ParameterGroupParameterArrayInput
	var paramsArr redshift.ParameterGroupParameterArray
	for _, p := range spec.Parameters {
		paramsArr = append(paramsArr, &redshift.ParameterGroupParameterArgs{
			Name:  pulumi.String(p.Name),
			Value: pulumi.String(p.Value),
		})
	}
	params = paramsArr

	args := &redshift.ParameterGroupArgs{
		Name:       pulumi.Sprintf("%s-params", locals.AwsRedshiftCluster.Metadata.Id),
		Family:     pulumi.String("redshift-1.0"),
		Tags:       pulumi.ToStringMap(locals.Labels),
		Parameters: params,
	}

	pg, err := redshift.NewParameterGroup(ctx, "cluster-parameter-group", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create parameter group")
	}
	return pg, nil
}

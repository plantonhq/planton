package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/neptune"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// clusterParameterGroup creates a Neptune cluster parameter group when inline
// parameters are provided and an existing group name is not set.
func clusterParameterGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*neptune.ClusterParameterGroup, error) {
	spec := locals.AwsNeptuneCluster.Spec
	if spec == nil {
		return nil, nil
	}

	if spec.ClusterParameterGroupName != "" && len(spec.ClusterParameters) == 0 {
		return nil, nil
	}

	if len(spec.ClusterParameters) == 0 {
		return nil, nil
	}

	family := getParameterGroupFamily(spec.GetEngineVersion())

	var params neptune.ClusterParameterGroupParameterArray
	for _, p := range spec.ClusterParameters {
		param := &neptune.ClusterParameterGroupParameterArgs{
			Name:  pulumi.String(p.Name),
			Value: pulumi.String(p.Value),
		}
		if p.ApplyMethod != "" {
			param.ApplyMethod = pulumi.String(p.ApplyMethod)
		}
		params = append(params, param)
	}

	args := &neptune.ClusterParameterGroupArgs{
		NamePrefix:  pulumi.Sprintf("%s-", locals.AwsNeptuneCluster.Metadata.Id),
		Family:      pulumi.String(family),
		Description: pulumi.String("Neptune cluster parameter group for " + locals.AwsNeptuneCluster.Metadata.Id),
		Tags:        pulumi.ToStringMap(locals.Labels),
		Parameters:  params,
	}

	pg, err := neptune.NewClusterParameterGroup(ctx, "neptune-cluster-parameter-group", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create neptune cluster parameter group")
	}
	return pg, nil
}

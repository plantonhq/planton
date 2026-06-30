package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// clusterParameterGroup creates a DB Cluster Parameter Group when inline parameters are provided or when
// an explicit db_cluster_parameter_group_name is not provided but we want a managed group.
func clusterParameterGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*rds.ClusterParameterGroup, error) {
	spec := locals.AwsRdsCluster.Spec
	if spec == nil {
		return nil, nil
	}

	if spec.DbClusterParameterGroupName != "" && len(spec.Parameters) == 0 {
		// Using an existing parameter group, and no inline params to manage
		return nil, nil
	}

	// Derive family from engine/engine_version conservatively.
	// For aurora-mysql: e.g., aurora-mysql8.0
	// For aurora-postgresql: e.g., aurora-postgresql14
	family := ""
	engine := spec.Engine
	version := spec.EngineVersion
	if strings.HasPrefix(engine, "aurora-mysql") {
		parts := strings.Split(version, ".")
		if len(parts) > 0 {
			major := parts[0]
			if major != "" {
				family = "aurora-mysql" + major
			}
		}
	} else if strings.HasPrefix(engine, "aurora-postgresql") {
		parts := strings.Split(version, ".")
		if len(parts) > 0 {
			family = "aurora-postgresql" + parts[0]
		}
	}

	var params rds.ClusterParameterGroupParameterArray
	for _, p := range spec.Parameters {
		params = append(params, &rds.ClusterParameterGroupParameterArgs{
			ApplyMethod: pulumi.String(p.ApplyMethod),
			Name:        pulumi.String(p.Name),
			Value:       pulumi.String(p.Value),
		})
	}

	args := &rds.ClusterParameterGroupArgs{
		NamePrefix: pulumi.Sprintf("%s-", locals.AwsRdsCluster.Metadata.Id),
		Family:     pulumi.String(family),
		Tags:       pulumi.ToStringMap(locals.Labels),
		Parameters: params,
	}

	pg, err := rds.NewClusterParameterGroup(ctx, "cluster-parameter-group", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create cluster parameter group")
	}
	return pg, nil
}

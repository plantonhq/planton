package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/docdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// clusterParameterGroup creates a DocumentDB cluster parameter group when custom parameters are provided.
func clusterParameterGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*docdb.ClusterParameterGroup, error) {
	spec := locals.AwsDocumentDb.Spec
	if spec == nil {
		return nil, nil
	}

	// Skip if no custom parameters are provided
	if len(spec.ClusterParameters) == 0 {
		return nil, nil
	}

	// Determine engine family based on engine version
	engineFamily := getEngineFamily(spec.GetEngineVersion())

	// Build parameter array
	var params docdb.ClusterParameterGroupParameterArray
	for _, p := range spec.ClusterParameters {
		param := &docdb.ClusterParameterGroupParameterArgs{
			Name:  pulumi.String(p.Name),
			Value: pulumi.String(p.Value),
		}
		if p.ApplyMethod != "" {
			param.ApplyMethod = pulumi.String(p.ApplyMethod)
		}
		params = append(params, param)
	}

	pg, err := docdb.NewClusterParameterGroup(ctx, "cluster-param-group", &docdb.ClusterParameterGroupArgs{
		Name:        pulumi.String(locals.AwsDocumentDb.Metadata.Id),
		Family:      pulumi.String(engineFamily),
		Description: pulumi.String("DocumentDB cluster parameter group for " + locals.AwsDocumentDb.Metadata.Id),
		Parameters:  params,
		Tags:        pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create cluster parameter group")
	}
	return pg, nil
}

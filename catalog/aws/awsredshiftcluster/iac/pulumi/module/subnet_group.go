package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/redshift"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// subnetGroup manages the cluster's Redshift subnet group when the spec
// brings raw subnets; an existing group name short-circuits it. The group
// itself is pure glue (a named list of subnets), which is why it stays
// inside this module instead of being its own node -- the referenced
// subnets are first-class AwsSubnet nodes this module never modifies.
func subnetGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*redshift.SubnetGroup, error) {
	spec := locals.AwsRedshiftCluster.Spec
	if spec.ClusterSubnetGroupName != "" || len(spec.SubnetIds) == 0 {
		return nil, nil
	}

	subnetIds := pulumi.StringArray{}
	for _, subnetId := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(subnetId.GetValue()))
	}

	createdSubnetGroup, err := redshift.NewSubnetGroup(ctx, "subnet-group",
		&redshift.SubnetGroupArgs{
			Name:      pulumi.String(locals.ClusterIdentifier),
			SubnetIds: subnetIds,
			Tags:      pulumi.ToStringMap(locals.AwsTags),
		}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Redshift subnet group")
	}
	return createdSubnetGroup, nil
}

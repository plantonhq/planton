package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/docdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// subnetGroup manages the cluster's DB subnet group when the spec brings
// raw subnets; an existing group name short-circuits it. The group itself
// is pure glue (a named list of subnets), which is why it stays inside
// this module instead of being its own node -- the referenced subnets are
// first-class AwsSubnet nodes this module never modifies.
func subnetGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*docdb.SubnetGroup, error) {
	spec := locals.AwsDocumentDb.Spec
	if spec.DbSubnetGroupName != "" || len(spec.SubnetIds) == 0 {
		return nil, nil
	}

	subnetIds := pulumi.StringArray{}
	for _, subnetId := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(subnetId.GetValue()))
	}

	createdSubnetGroup, err := docdb.NewSubnetGroup(ctx, "subnet-group",
		&docdb.SubnetGroupArgs{
			Name:      pulumi.String(locals.ClusterIdentifier),
			SubnetIds: subnetIds,
			Tags:      pulumi.ToStringMap(locals.AwsTags),
		}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create DB subnet group")
	}
	return createdSubnetGroup, nil
}

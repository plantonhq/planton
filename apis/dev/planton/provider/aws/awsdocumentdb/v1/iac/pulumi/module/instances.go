package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/docdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func docdbClusterInstances(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	cluster *docdb.Cluster,
) error {
	spec := locals.AwsDocumentDb.Spec

	instanceCount := spec.GetInstanceCount()
	if instanceCount == 0 {
		instanceCount = 1
	}

	instanceClass := spec.GetInstanceClass()
	if instanceClass == "" {
		instanceClass = "db.r6g.large"
	}

	for i := int32(0); i < instanceCount; i++ {
		instanceIdentifier := fmt.Sprintf("%s-%d", locals.AwsDocumentDb.Metadata.Id, i+1)

		args := &docdb.ClusterInstanceArgs{
			Identifier:              pulumi.String(instanceIdentifier),
			ClusterIdentifier:       cluster.ID(),
			InstanceClass:           pulumi.String(instanceClass),
			AutoMinorVersionUpgrade: pulumi.Bool(spec.GetAutoMinorVersionUpgrade()),
			Tags:                    pulumi.ToStringMap(locals.Labels),
		}

		// Apply immediately setting
		if spec.ApplyImmediately {
			args.ApplyImmediately = pulumi.Bool(true)
		}

		_, err := docdb.NewClusterInstance(ctx, fmt.Sprintf("docdb-instance-%d", i+1), args,
			pulumi.Provider(provider),
			pulumi.DependsOn([]pulumi.Resource{cluster}),
		)
		if err != nil {
			return errors.Wrapf(err, "create documentdb cluster instance %d", i+1)
		}
	}

	return nil
}

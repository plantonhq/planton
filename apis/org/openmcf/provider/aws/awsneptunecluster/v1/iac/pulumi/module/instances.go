package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/neptune"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func neptuneClusterInstances(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	cluster *neptune.Cluster,
) error {
	spec := locals.AwsNeptuneCluster.Spec

	instanceCount := spec.GetInstanceCount()
	if instanceCount == 0 {
		instanceCount = 1
	}

	instanceClass := spec.GetInstanceClass()
	if instanceClass == "" {
		instanceClass = "db.r6g.large"
	}

	for i := int32(0); i < instanceCount; i++ {
		instanceIdentifier := fmt.Sprintf("%s-%d", locals.AwsNeptuneCluster.Metadata.Id, i+1)

		args := &neptune.ClusterInstanceArgs{
			Identifier:        pulumi.String(instanceIdentifier),
			ClusterIdentifier: cluster.ID(),
			InstanceClass:     pulumi.String(instanceClass),
			Engine:            pulumi.String("neptune"),
			Tags:              pulumi.ToStringMap(locals.Labels),
		}

		if spec.ApplyImmediately {
			args.ApplyImmediately = pulumi.Bool(true)
		}

		// Promotion tier: first instance gets tier 0 (primary), replicas get tier 1
		if i > 0 {
			args.PromotionTier = pulumi.Int(1)
		}

		_, err := neptune.NewClusterInstance(ctx, fmt.Sprintf("neptune-instance-%d", i+1), args,
			pulumi.Provider(provider),
			pulumi.DependsOn([]pulumi.Resource{cluster}),
		)
		if err != nil {
			return errors.Wrapf(err, "create neptune cluster instance %d", i+1)
		}
	}

	return nil
}

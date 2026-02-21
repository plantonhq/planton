package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/objectstorage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func replicationPolicies(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, bucket *objectstorage.Bucket) error {
	spec := locals.OciObjectStorageBucket.Spec

	for _, policy := range spec.ReplicationPolicies {
		resourceName := fmt.Sprintf("%s-%s", locals.BucketName, policy.Name)
		_, err := objectstorage.NewReplicationPolicy(ctx, resourceName, &objectstorage.ReplicationPolicyArgs{
			Bucket:                pulumi.String(spec.Name),
			Namespace:             pulumi.String(spec.Namespace),
			Name:                  pulumi.String(policy.Name),
			DestinationBucketName: pulumi.String(policy.DestinationBucketName),
			DestinationRegionName: pulumi.String(policy.DestinationRegionName),
		}, pulumiOciOpt(provider), pulumi.DependsOn([]pulumi.Resource{bucket}))
		if err != nil {
			return errors.Wrapf(err, "failed to create replication policy %s", policy.Name)
		}
	}

	return nil
}

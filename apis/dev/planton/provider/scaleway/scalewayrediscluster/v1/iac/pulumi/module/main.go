package module

import (
	"github.com/pkg/errors"
	scalewayredisclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayrediscluster/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions a Scaleway Redis
// cluster with optional ACL rules or Private Network attachment.
//
// This is a standalone resource (not composite): it wraps a single
// scaleway_redis_cluster Terraform resource. ACL rules and Private
// Network configuration are inline properties of the cluster itself.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayredisclusterv1.ScalewayRedisClusterStackInput,
) error {
	// 1. Prepare locals (metadata, labels, resolved references).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the Redis cluster and export outputs.
	if err := redisCluster(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create redis cluster")
	}

	return nil
}

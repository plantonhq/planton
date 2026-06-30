package module

import (
	"github.com/pkg/errors"
	scalewaykapsuleclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewaykapsulecluster/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions the complete
// ScalewayKapsuleCluster composite:
//
//  1. The Kapsule cluster (managed Kubernetes control plane).
//  2. The default node pool (so the cluster is immediately usable).
//
// Resources are created in dependency order: cluster first (the pool
// references the cluster's ID), then the default node pool.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewaykapsuleclusterv1.ScalewayKapsuleClusterStackInput,
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

	// 3. Create the Kapsule cluster + default node pool.
	if err := cluster(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create kapsule cluster")
	}

	return nil
}

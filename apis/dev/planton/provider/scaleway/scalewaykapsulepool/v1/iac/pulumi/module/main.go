package module

import (
	"github.com/pkg/errors"
	scalewaykapsulepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewaykapsulepool/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions a ScalewayKapsulePool.
//
// This is a standalone (non-composite) resource that creates a single
// Scaleway Kubernetes node pool in an existing Kapsule cluster.
//
// Kubernetes labels and taints are applied via Scaleway's CCM tag convention
// (merged into the pool's tags during locals initialization).
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewaykapsulepoolv1.ScalewayKapsulePoolStackInput,
) error {
	// 1. Prepare locals (metadata, resolved references, merged tags).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the node pool.
	if err := nodePool(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create kapsule node pool")
	}

	return nil
}

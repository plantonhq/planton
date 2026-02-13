package module

import (
	"github.com/pkg/errors"
	scalewayblockvolumev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewayblockvolume/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions a Scaleway Block
// Storage volume with the specified performance tier and size.
//
// This is a standalone resource (not composite): it wraps a single
// scaleway_block_volume resource. The volume is a raw block device
// that must be formatted and mounted at the OS level after attachment
// to an Instance.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayblockvolumev1.ScalewayBlockVolumeStackInput,
) error {
	// 1. Prepare locals (metadata, tags, resolved references).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the block volume and export outputs.
	if err := volume(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create block volume")
	}

	return nil
}

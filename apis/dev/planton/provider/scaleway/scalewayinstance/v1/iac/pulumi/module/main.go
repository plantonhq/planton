package module

import (
	"github.com/pkg/errors"
	scalewayinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayinstance/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions the complete
// ScalewayInstance composite:
//
//  1. Optional Flexible IP (dedicated public IPv4).
//  2. Optional additional local volumes (l_ssd, scratch).
//  3. The instance server (with root volume, optional private network
//     attachment, and optional security group).
//
// Resources are created in dependency order: IP and volumes first (the
// server references their IDs), then the server itself.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayinstancev1.ScalewayInstanceStackInput,
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

	// 3. Create the instance composite (IP + volumes + server).
	if err := createInstance(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create instance")
	}

	return nil
}

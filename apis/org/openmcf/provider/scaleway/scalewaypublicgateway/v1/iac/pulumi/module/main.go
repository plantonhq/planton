package module

import (
	"github.com/pkg/errors"
	scalewaypublicgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewaypublicgateway/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewaypublicgatewayv1.ScalewayPublicGatewayStackInput,
) error {
	// 1. Prepare locals (metadata, labels, resolved references, etc.).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the Public Gateway composite (IP + Gateway + GatewayNetwork).
	if err := gateway(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create public gateway")
	}

	return nil
}

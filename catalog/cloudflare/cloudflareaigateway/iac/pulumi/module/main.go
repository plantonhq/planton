package module

import (
	"github.com/pkg/errors"
	cloudflareaigatewayv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareaigateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/cloudflare/pulumicloudflareprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *cloudflareaigatewayv1alpha1.CloudflareAiGatewayIacInput,
) error {
	// 1. Prepare locals (metadata, credentials).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a Cloudflare provider from the supplied credential.
	cloudflareProvider, err := pulumicloudflareprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup cloudflare provider")
	}

	// 3. Create the gateway, then its dynamic routes (each route is its own
	// provider object attached to the gateway).
	createdGateway, err := aiGateway(ctx, locals, cloudflareProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create ai gateway")
	}

	if err := dynamicRoutes(ctx, locals, cloudflareProvider, createdGateway); err != nil {
		return errors.Wrap(err, "failed to create dynamic routes")
	}

	return nil
}

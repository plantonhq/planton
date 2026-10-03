package module

import (
	"github.com/pkg/errors"
	awshttpapigatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awshttpapigateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates HTTP API Gateway creation and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awshttpapigatewayv1alpha1.AwsHttpApiGatewayIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// -----------------------------------------------------------------------
	// AWS provider
	// -----------------------------------------------------------------------

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// -----------------------------------------------------------------------
	// 1. Create the HTTP API
	// -----------------------------------------------------------------------

	createdApi, err := httpApi(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "http api")
	}

	// -----------------------------------------------------------------------
	// 2. Create integrations (deduplicated)
	// -----------------------------------------------------------------------

	createdIntegrations, err := integrations(ctx, locals, createdApi, provider)
	if err != nil {
		return errors.Wrap(err, "api integrations")
	}

	// -----------------------------------------------------------------------
	// 3. Create authorizers (if any)
	// -----------------------------------------------------------------------

	authorizerMap, err := authorizers(ctx, locals, createdApi, provider)
	if err != nil {
		return errors.Wrap(err, "api authorizers")
	}

	// -----------------------------------------------------------------------
	// 4. Create routes
	// -----------------------------------------------------------------------

	createdRoutes, err := routes(ctx, locals, createdApi, createdIntegrations, authorizerMap, provider)
	if err != nil {
		return errors.Wrap(err, "api routes")
	}

	// -----------------------------------------------------------------------
	// 5. Create the stage (after routes: per-route settings reference them)
	// -----------------------------------------------------------------------

	if err := stage(ctx, locals, createdApi, createdRoutes, provider); err != nil {
		return errors.Wrap(err, "api stage")
	}

	return nil
}

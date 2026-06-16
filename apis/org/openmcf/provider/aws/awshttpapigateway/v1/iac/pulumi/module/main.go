package module

import (
	"github.com/pkg/errors"
	awshttpapigatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awshttpapigateway/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates HTTP API Gateway creation and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awshttpapigatewayv1.AwsHttpApiGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// -----------------------------------------------------------------------
	// AWS provider
	// -----------------------------------------------------------------------

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Target.Spec.Region)
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
	// 2. Create the stage
	// -----------------------------------------------------------------------

	if err := stage(ctx, locals, createdApi, provider); err != nil {
		return errors.Wrap(err, "api stage")
	}

	// -----------------------------------------------------------------------
	// 3. Create integrations (deduplicated)
	// -----------------------------------------------------------------------

	integrationMap, err := integrations(ctx, locals, createdApi, provider)
	if err != nil {
		return errors.Wrap(err, "api integrations")
	}

	// -----------------------------------------------------------------------
	// 4. Create authorizers (if any)
	// -----------------------------------------------------------------------

	authorizerMap, err := authorizers(ctx, locals, createdApi, provider)
	if err != nil {
		return errors.Wrap(err, "api authorizers")
	}

	// -----------------------------------------------------------------------
	// 5. Create routes
	// -----------------------------------------------------------------------

	if err := routes(ctx, locals, createdApi, integrationMap, authorizerMap, provider); err != nil {
		return errors.Wrap(err, "api routes")
	}

	return nil
}

package module

import (
	"github.com/pkg/errors"
	awsrestapigatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsrestapigateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the REST API: the API itself with
// its resource/method tree (or imported OpenAPI document), the named
// satellites, the hash-triggered deployment, and the stage.
func Resources(ctx *pulumi.Context, iacInput *awsrestapigatewayv1alpha1.AwsRestApiGatewayIacInput) error {
	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return errors.Wrap(err, "initialize locals")
	}

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	api, apiExtras, err := restApi(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "rest api")
	}

	satellites, err := apiSatellites(ctx, locals, provider, api)
	if err != nil {
		return errors.Wrap(err, "api satellites")
	}

	tree, err := routeTree(ctx, locals, provider, api, satellites)
	if err != nil {
		return errors.Wrap(err, "route tree")
	}

	if err := stage(ctx, locals, provider, api, tree, satellites, apiExtras); err != nil {
		return errors.Wrap(err, "stage")
	}

	return nil
}

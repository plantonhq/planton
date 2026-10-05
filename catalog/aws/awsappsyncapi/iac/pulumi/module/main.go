package module

import (
	"github.com/pkg/errors"
	awsappsyncapiv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsappsyncapi/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the API (the graphql XOR events
// arm), its data sources, the arm's satellites (types, functions,
// resolvers, cache, channel namespaces), API keys, the custom domain,
// and MERGED source-API associations, and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsappsyncapiv1alpha1.AwsAppSyncApiIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Exactly one arm creates the API (the spec's CEL wall); everything
	// downstream hangs off its id.
	var api *createdApi
	if locals.Spec.GetGraphql() != nil {
		api, err = graphqlApi(ctx, locals, provider)
		if err != nil {
			return errors.Wrap(err, "graphql api")
		}
	} else {
		api, err = eventsApi(ctx, locals, provider)
		if err != nil {
			return errors.Wrap(err, "events api")
		}
	}

	createdDatasources, err := datasources(ctx, locals, provider, api)
	if err != nil {
		return errors.Wrap(err, "datasources")
	}

	if err := satellites(ctx, locals, provider, api, createdDatasources); err != nil {
		return errors.Wrap(err, "satellites")
	}

	return nil
}

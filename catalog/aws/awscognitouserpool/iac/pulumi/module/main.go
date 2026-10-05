package module

import (
	"github.com/pkg/errors"
	awscognitouserpoolv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscognitouserpool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of a Cognito User Pool with its folded
// pool-scoped satellites (hosted-UI domain, log delivery), then exports
// outputs for downstream references. App clients, identity providers, and
// resource servers are separate kinds that compose onto the pool by
// reference.
func Resources(ctx *pulumi.Context, iacInput *awscognitouserpoolv1alpha1.AwsCognitoUserPoolIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// User pool (always created)
	createdPool, err := userPool(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "cognito user pool")
	}

	// Hosted-UI domain (optional; one per pool)
	if err := domain(ctx, locals, createdPool, provider); err != nil {
		return errors.Wrap(err, "cognito user pool domain")
	}

	// Log delivery (optional; one pool-scoped configuration)
	if err := logDelivery(ctx, locals, createdPool, provider); err != nil {
		return errors.Wrap(err, "cognito user pool log delivery")
	}

	// User groups (pool-scoped; membership is data-plane, managed at runtime)
	if err := userGroups(ctx, locals, createdPool, provider); err != nil {
		return errors.Wrap(err, "cognito user pool groups")
	}

	// Pool-wide risk configuration (threat protection's automated responses)
	if err := riskConfiguration(ctx, locals, createdPool, provider); err != nil {
		return errors.Wrap(err, "cognito user pool risk configuration")
	}

	return nil
}

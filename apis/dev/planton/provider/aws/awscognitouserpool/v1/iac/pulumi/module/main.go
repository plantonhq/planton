package module

import (
	"github.com/pkg/errors"
	awscognitouserpoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awscognitouserpool/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of a Cognito User Pool with app clients and
// an optional domain, then exports outputs for downstream references.
func Resources(ctx *pulumi.Context, stackInput *awscognitouserpoolv1.AwsCognitoUserPoolStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// User pool (always created)
	createdPool, err := userPool(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "cognito user pool")
	}

	// App clients (always created — at least one required by spec)
	if err := clients(ctx, locals, createdPool, provider); err != nil {
		return errors.Wrap(err, "cognito user pool clients")
	}

	// Domain (optional)
	if err := domain(ctx, locals, createdPool, provider); err != nil {
		return errors.Wrap(err, "cognito user pool domain")
	}

	return nil
}

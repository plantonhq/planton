package module

import (
	"github.com/pkg/errors"
	awscognitouserpoolclientv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscognitouserpoolclient/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates app-client creation and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awscognitouserpoolclientv1alpha1.AwsCognitoUserPoolClientIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdClient, err := client(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "cognito user pool client")
	}

	// Client-scoped risk configuration (threat protection's automated
	// responses for this client only)
	if err := riskConfiguration(ctx, locals, createdClient, provider); err != nil {
		return errors.Wrap(err, "cognito user pool client risk configuration")
	}

	return nil
}

package module

import (
	"github.com/pkg/errors"
	awssnssubscriptionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssnssubscription/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the SNS subscription and exports its outputs.
func Resources(ctx *pulumi.Context, iacInput *awssnssubscriptionv1alpha1.AwsSnsSubscriptionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if _, err := subscription(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "sns subscription")
	}

	return nil
}

package module

import (
	"github.com/pkg/errors"
	awscloudwatchlogdeliveryv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudwatchlogdelivery/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the vended pipeline and/or the
// cross-account destination, and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awscloudwatchlogdeliveryv1alpha1.AwsCloudwatchLogDeliveryIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := vended(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "vended")
	}

	if err := crossAccountDestination(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "cross-account destination")
	}

	return nil
}

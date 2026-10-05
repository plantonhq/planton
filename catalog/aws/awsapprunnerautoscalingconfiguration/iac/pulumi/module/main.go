package module

import (
	"github.com/pkg/errors"
	awsapprunnerautoscalingconfigurationv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsapprunnerautoscalingconfiguration/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates auto scaling configuration creation and exports
// outputs.
func Resources(ctx *pulumi.Context, iacInput *awsapprunnerautoscalingconfigurationv1alpha1.AwsAppRunnerAutoScalingConfigurationIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsAppRunnerAutoScalingConfiguration.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := autoScalingConfiguration(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "auto scaling configuration")
	}

	return nil
}

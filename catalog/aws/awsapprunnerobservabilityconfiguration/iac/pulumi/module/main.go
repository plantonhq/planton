package module

import (
	"github.com/pkg/errors"
	awsapprunnerobservabilityconfigurationv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsapprunnerobservabilityconfiguration/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates observability configuration creation and exports
// outputs.
func Resources(ctx *pulumi.Context, iacInput *awsapprunnerobservabilityconfigurationv1alpha1.AwsAppRunnerObservabilityConfigurationIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsAppRunnerObservabilityConfiguration.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := observabilityConfiguration(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "observability configuration")
	}

	return nil
}

package module

import (
	"github.com/pkg/errors"
	awsapprunnerservicev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsapprunnerservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates App Runner service creation (with its custom domain
// associations and optional WAF association) and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsapprunnerservicev1alpha1.AwsAppRunnerServiceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsAppRunnerService.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := service(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "app runner service")
	}

	return nil
}

package module

import (
	"github.com/pkg/errors"
	awsecsservicev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecsservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the ECS service and, when configured, its folded
// Application Auto Scaling target and policies.
func Resources(ctx *pulumi.Context, iacInput *awsecsservicev1alpha1.AwsEcsServiceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEcsService.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdService, err := ecsService(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create aws ecs service resource")
	}

	if locals.AwsEcsService.Spec.Autoscaling != nil {
		if err := autoscaling(ctx, locals, provider, createdService); err != nil {
			return errors.Wrap(err, "failed to configure service autoscaling")
		}
	}

	return nil
}

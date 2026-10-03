package module

import (
	"github.com/pkg/errors"
	awsautoscalinggroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsautoscalinggroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsautoscalinggroupv1alpha1.AwsAutoScalingGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsAutoScalingGroup.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdGroup, err := group(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create auto-scaling group")
	}

	// The folded sub-resources (scaling policies, scheduled actions,
	// lifecycle hooks, notifications) each materialize as their own provider
	// resource keyed on the group's name, so adding or removing one is an
	// in-place update that never replaces the group.
	if err := scaling(ctx, locals, provider, createdGroup); err != nil {
		return errors.Wrap(err, "failed to create scaling resources")
	}

	return nil
}

package module

import (
	"github.com/pkg/errors"
	awsbudgetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbudget/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the budget and its folded
// actions, and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsbudgetv1alpha1.AwsBudgetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := budget(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "budget")
	}

	return nil
}

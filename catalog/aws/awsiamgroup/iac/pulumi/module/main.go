package module

import (
	"github.com/pkg/errors"
	awsiamgroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiamgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of the IAM group, its declarative
// membership, and its policies, and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awsiamgroupv1alpha1.AwsIamGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := group(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "group")
	}

	return nil
}

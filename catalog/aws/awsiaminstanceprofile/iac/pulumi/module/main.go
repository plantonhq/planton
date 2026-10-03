package module

import (
	"github.com/pkg/errors"
	awsiaminstanceprofilev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiaminstanceprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awsiaminstanceprofilev1alpha1.AwsIamInstanceProfileIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsIamInstanceProfile.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := instanceProfile(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create iam instance profile")
	}

	return nil
}

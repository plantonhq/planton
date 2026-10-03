package module

import (
	"github.com/pkg/errors"
	iamrolev1 "github.com/plantonhq/planton/catalog/aws/awsiamrole/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *iamrolev1.AwsIamRoleIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsIamRole.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// create the IAM role
	if err := iamRole(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "failed to create iam role resource")
	}

	return nil
}

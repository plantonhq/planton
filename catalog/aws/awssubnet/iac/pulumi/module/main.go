package module

import (
	"github.com/pkg/errors"
	awssubnetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssubnet/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awssubnetv1alpha1.AwsSubnetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsSubnet.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdSubnet, err := subnet(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create subnet")
	}

	if err := configureRouting(ctx, locals, provider, createdSubnet); err != nil {
		return errors.Wrap(err, "failed to configure subnet routing")
	}

	return nil
}

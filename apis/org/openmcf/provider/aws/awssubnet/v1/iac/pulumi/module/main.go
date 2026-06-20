package module

import (
	"github.com/pkg/errors"
	awssubnetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awssubnet/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awssubnetv1.AwsSubnetStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsSubnet.Spec.Region)
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

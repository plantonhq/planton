package module

import (
	"github.com/pkg/errors"
	awskinesisstream "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awskinesisstream/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates Kinesis Data Stream creation and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awskinesisstream.AwsKinesisStreamStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := stream(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "kinesis stream")
	}

	return nil
}

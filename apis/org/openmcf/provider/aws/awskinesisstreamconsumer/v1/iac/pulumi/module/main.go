package module

import (
	"github.com/pkg/errors"
	awskinesisstreamconsumer "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awskinesisstreamconsumer/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates Kinesis Stream Consumer creation and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awskinesisstreamconsumer.AwsKinesisStreamConsumerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	if err := consumer(ctx, locals, provider); err != nil {
		return errors.Wrap(err, "kinesis stream consumer")
	}

	return nil
}

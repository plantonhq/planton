package module

import (
	"github.com/pkg/errors"
	awskinesisfirehose "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awskinesisfirehose/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates Kinesis Data Firehose delivery stream creation and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awskinesisfirehose.AwsKinesisFirehoseStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Target.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	stream, err := deliveryStream(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "firehose delivery stream")
	}

	if err := outputs(ctx, stream); err != nil {
		return errors.Wrap(err, "outputs")
	}

	return nil
}

package module

import (
	"github.com/pkg/errors"
	awskinesisfirehose "github.com/plantonhq/planton/catalog/aws/awskinesisfirehose/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates Kinesis Data Firehose delivery stream creation and exports outputs.
func Resources(ctx *pulumi.Context, iacInput *awskinesisfirehose.AwsKinesisFirehoseIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.Target.Spec.Region)
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

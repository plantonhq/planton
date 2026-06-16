package module

import (
	"github.com/pkg/errors"
	awscloudwatchloggroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscloudwatchloggroup/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awscloudwatchloggroupv1.AwsCloudwatchLogGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsCloudwatchLogGroup.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	result, err := logGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudwatch log group")
	}

	ctx.Export(OpLogGroupArn, result.LogGroupArn)
	ctx.Export(OpLogGroupName, result.LogGroupName)

	return nil
}

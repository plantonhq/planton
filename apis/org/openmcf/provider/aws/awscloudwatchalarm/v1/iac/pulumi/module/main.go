package module

import (
	"github.com/pkg/errors"
	awscloudwatchalarmv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscloudwatchalarm/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awscloudwatchalarmv1.AwsCloudwatchAlarmStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsCloudwatchAlarm.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	result, err := alarm(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudwatch metric alarm")
	}

	ctx.Export(OpAlarmArn, result.AlarmArn)
	ctx.Export(OpAlarmName, result.AlarmName)

	return nil
}

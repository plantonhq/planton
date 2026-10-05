package module

import (
	"github.com/pkg/errors"
	awscloudwatchalarmv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudwatchalarm/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awscloudwatchalarmv1alpha1.AwsCloudwatchAlarmIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsCloudwatchAlarm.Spec.Region)
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

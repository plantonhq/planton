package module

import (
	"github.com/pkg/errors"
	awscloudwatchloggroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudwatchloggroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *awscloudwatchloggroupv1alpha1.AwsCloudwatchLogGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsCloudwatchLogGroup.Spec.Region)
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

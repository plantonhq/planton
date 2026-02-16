package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apprunner"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// autoScalingConfig creates an App Runner Auto Scaling Configuration Version.
// It controls how App Runner scales the number of instances based on incoming
// request concurrency. Only called when the auto_scaling block is present in the spec.
func autoScalingConfig(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*apprunner.AutoScalingConfigurationVersion, error) {
	spec := locals.AwsAppRunnerService.Spec
	scaling := spec.GetAutoScaling()
	resourceName := locals.AwsAppRunnerService.Metadata.Name

	args := &apprunner.AutoScalingConfigurationVersionArgs{
		AutoScalingConfigurationName: pulumi.String(resourceName),
		Tags:                         pulumi.ToStringMap(locals.AwsTags),
	}

	// Use getter methods which safely return zero-values when the optional field is nil.
	if scaling.GetMinSize() > 0 {
		args.MinSize = pulumi.Int(int(scaling.GetMinSize()))
	}
	if scaling.GetMaxSize() > 0 {
		args.MaxSize = pulumi.Int(int(scaling.GetMaxSize()))
	}
	if scaling.GetMaxConcurrency() > 0 {
		args.MaxConcurrency = pulumi.Int(int(scaling.GetMaxConcurrency()))
	}

	config, err := apprunner.NewAutoScalingConfigurationVersion(ctx, resourceName+"-auto-scaling", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create App Runner auto scaling configuration")
	}

	// Export the Auto Scaling Configuration ARN.
	ctx.Export(OpAutoScalingConfigurationArn, config.Arn)

	return config, nil
}

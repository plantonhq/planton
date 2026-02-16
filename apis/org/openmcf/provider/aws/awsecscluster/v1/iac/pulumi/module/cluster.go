package module

import (
	"fmt"

	"github.com/pkg/errors"
	awsecsclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsecscluster/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func cluster(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	clusterName := locals.AwsEcsCluster.Metadata.Name

	args := &ecs.ClusterArgs{
		Name: pulumi.String(clusterName),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Optional: enable CloudWatch Container Insights
	if locals.AwsEcsCluster.Spec.EnableContainerInsights {
		args.Settings = ecs.ClusterSettingArray{
			&ecs.ClusterSettingArgs{
				Name:  pulumi.String("containerInsights"),
				Value: pulumi.String("enabled"),
			},
		}
	}

	// Optional: configure ECS Exec with auditing
	if locals.AwsEcsCluster.Spec.ExecuteCommandConfiguration != nil {
		execConfig := locals.AwsEcsCluster.Spec.ExecuteCommandConfiguration
		if execConfig.Logging != awsecsclusterv1.ExecConfiguration_LOGGING_UNSPECIFIED {
			execCmdConfig := &ecs.ClusterConfigurationExecuteCommandConfigurationArgs{}

			switch execConfig.Logging {
			case awsecsclusterv1.ExecConfiguration_DEFAULT:
				execCmdConfig.Logging = pulumi.String("DEFAULT")
			case awsecsclusterv1.ExecConfiguration_NONE:
				execCmdConfig.Logging = pulumi.String("NONE")
			case awsecsclusterv1.ExecConfiguration_OVERRIDE:
				execCmdConfig.Logging = pulumi.String("OVERRIDE")
				if execConfig.LogConfiguration != nil {
					logConfig := &ecs.ClusterConfigurationExecuteCommandConfigurationLogConfigurationArgs{}

					if execConfig.LogConfiguration.CloudWatchLogGroupName != "" {
						logConfig.CloudWatchLogGroupName = pulumi.String(execConfig.LogConfiguration.CloudWatchLogGroupName)
					}
					if execConfig.LogConfiguration.CloudWatchEncryptionEnabled && execConfig.KmsKeyId != nil && execConfig.KmsKeyId.GetValue() != "" {
						logConfig.CloudWatchEncryptionEnabled = pulumi.Bool(true)
					}
					if execConfig.LogConfiguration.S3BucketName != "" {
						logConfig.S3BucketName = pulumi.String(execConfig.LogConfiguration.S3BucketName)
					}
					if execConfig.LogConfiguration.S3KeyPrefix != "" {
						logConfig.S3KeyPrefix = pulumi.String(execConfig.LogConfiguration.S3KeyPrefix)
					}
					execCmdConfig.LogConfiguration = logConfig
				}
			}

			if execConfig.KmsKeyId != nil && execConfig.KmsKeyId.GetValue() != "" {
				execCmdConfig.KmsKeyId = pulumi.String(execConfig.KmsKeyId.GetValue())
			}

			args.Configuration = &ecs.ClusterConfigurationArgs{
				ExecuteCommandConfiguration: execCmdConfig,
			}
		}
	}

	awsEcsCluster, err := ecs.NewCluster(ctx, locals.AwsEcsCluster.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to create ECS cluster: %s", clusterName))
	}

	// Optional: capacity providers with default strategy
	if len(locals.AwsEcsCluster.Spec.CapacityProviders) > 0 {
		var cps pulumi.StringArray
		for _, cp := range locals.AwsEcsCluster.Spec.CapacityProviders {
			cps = append(cps, pulumi.String(cp))
		}

		cpArgs := &ecs.ClusterCapacityProvidersArgs{
			ClusterName:       awsEcsCluster.Name,
			CapacityProviders: cps,
		}

		// Configure default capacity provider strategy if specified
		if len(locals.AwsEcsCluster.Spec.DefaultCapacityProviderStrategy) > 0 {
			var strategies ecs.ClusterCapacityProvidersDefaultCapacityProviderStrategyArray
			for _, strategy := range locals.AwsEcsCluster.Spec.DefaultCapacityProviderStrategy {
				strategies = append(strategies, &ecs.ClusterCapacityProvidersDefaultCapacityProviderStrategyArgs{
					CapacityProvider: pulumi.String(strategy.CapacityProvider),
					Base:             pulumi.Int(int(strategy.Base)),
					Weight:           pulumi.Int(int(strategy.Weight)),
				})
			}
			cpArgs.DefaultCapacityProviderStrategies = strategies
		}

		_, err := ecs.NewClusterCapacityProviders(ctx,
			fmt.Sprintf("%s-capacity-providers", locals.AwsEcsCluster.Metadata.Name),
			cpArgs, pulumi.Provider(provider), pulumi.Parent(awsEcsCluster))
		if err != nil {
			return errors.Wrap(err, "unable to attach capacity providers")
		}
	}

	ctx.Export(OpClusterName, awsEcsCluster.Name)
	ctx.Export(OpClusterArn, awsEcsCluster.Arn)
	// Export capacity providers if configured
	if len(locals.AwsEcsCluster.Spec.CapacityProviders) > 0 {
		for i, cp := range locals.AwsEcsCluster.Spec.CapacityProviders {
			ctx.Export(fmt.Sprintf("%s.%d", OpClusterCapacityProviders, i), pulumi.String(cp))
		}
	}

	return nil
}

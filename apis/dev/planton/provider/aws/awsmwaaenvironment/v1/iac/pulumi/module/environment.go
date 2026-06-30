package module

import (
	"github.com/pkg/errors"
	awsmwaaenvironmentv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsmwaaenvironment/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/mwaa"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func environment(ctx *pulumi.Context, locals *Locals, provider *aws.Provider, createdSg *ec2.SecurityGroup) (*mwaa.Environment, error) {
	spec := locals.AwsMwaaEnvironment.Spec

	// Build the security group ID list: managed SG + associate_security_group_ids
	sgIds := pulumi.StringArray{}
	if createdSg != nil {
		sgIds = append(sgIds, createdSg.ID())
	}
	for _, sgOrRef := range spec.AssociateSecurityGroupIds {
		sgIds = append(sgIds, pulumi.String(sgOrRef.GetValue()))
	}

	// Build subnet ID list
	subnetIds := pulumi.StringArray{}
	for _, s := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
	}

	args := &mwaa.EnvironmentArgs{
		Name:                 pulumi.String(locals.AwsMwaaEnvironment.Metadata.Id),
		DagS3Path:            pulumi.String(spec.DagS3Path),
		ExecutionRoleArn:     pulumi.String(spec.ExecutionRoleArn.GetValue()),
		SourceBucketArn:      pulumi.String(spec.SourceBucketArn.GetValue()),
		NetworkConfiguration: buildNetworkConfiguration(subnetIds, sgIds),
		Tags:                 pulumi.ToStringMap(locals.Labels),
	}

	// Airflow version
	if spec.AirflowVersion != "" {
		args.AirflowVersion = pulumi.StringPtr(spec.AirflowVersion)
	}

	// Airflow configuration options
	if len(spec.AirflowConfigurationOptions) > 0 {
		args.AirflowConfigurationOptions = pulumi.ToStringMap(spec.AirflowConfigurationOptions)
	}

	// S3 artifact paths
	if spec.PluginsS3Path != "" {
		args.PluginsS3Path = pulumi.StringPtr(spec.PluginsS3Path)
	}
	if spec.PluginsS3ObjectVersion != "" {
		args.PluginsS3ObjectVersion = pulumi.StringPtr(spec.PluginsS3ObjectVersion)
	}
	if spec.RequirementsS3Path != "" {
		args.RequirementsS3Path = pulumi.StringPtr(spec.RequirementsS3Path)
	}
	if spec.RequirementsS3ObjectVersion != "" {
		args.RequirementsS3ObjectVersion = pulumi.StringPtr(spec.RequirementsS3ObjectVersion)
	}
	if spec.StartupScriptS3Path != "" {
		args.StartupScriptS3Path = pulumi.StringPtr(spec.StartupScriptS3Path)
	}
	if spec.StartupScriptS3ObjectVersion != "" {
		args.StartupScriptS3ObjectVersion = pulumi.StringPtr(spec.StartupScriptS3ObjectVersion)
	}

	// KMS encryption
	if spec.KmsKeyArn != nil {
		args.KmsKey = pulumi.StringPtr(spec.KmsKeyArn.GetValue())
	}

	// Environment sizing
	if spec.EnvironmentClass != "" {
		args.EnvironmentClass = pulumi.StringPtr(spec.EnvironmentClass)
	}
	if spec.MinWorkers > 0 {
		args.MinWorkers = pulumi.IntPtr(int(spec.MinWorkers))
	}
	if spec.MaxWorkers > 0 {
		args.MaxWorkers = pulumi.IntPtr(int(spec.MaxWorkers))
	}
	if spec.MinWebservers > 0 {
		args.MinWebservers = pulumi.IntPtr(int(spec.MinWebservers))
	}
	if spec.MaxWebservers > 0 {
		args.MaxWebservers = pulumi.IntPtr(int(spec.MaxWebservers))
	}
	if spec.Schedulers > 0 {
		args.Schedulers = pulumi.IntPtr(int(spec.Schedulers))
	}

	// Access and networking
	if spec.WebserverAccessMode != nil {
		args.WebserverAccessMode = pulumi.StringPtr(*spec.WebserverAccessMode)
	}
	if spec.EndpointManagement != "" {
		args.EndpointManagement = pulumi.StringPtr(spec.EndpointManagement)
	}

	// Logging configuration
	if spec.LoggingConfiguration != nil {
		args.LoggingConfiguration = buildLoggingConfiguration(spec.LoggingConfiguration)
	}

	// Maintenance window
	if spec.WeeklyMaintenanceWindowStart != "" {
		args.WeeklyMaintenanceWindowStart = pulumi.StringPtr(spec.WeeklyMaintenanceWindowStart)
	}

	// NOTE: WorkerReplacementStrategy is included in the spec but not available in
	// pulumi-aws SDK v7.3.0. The Terraform module supports it. When the SDK is upgraded,
	// uncomment this block:
	// if spec.WorkerReplacementStrategy != "" {
	//     args.WorkerReplacementStrategy = pulumi.StringPtr(spec.WorkerReplacementStrategy)
	// }

	env, err := mwaa.NewEnvironment(ctx, "mwaa-environment", args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create MWAA environment")
	}

	return env, nil
}

func buildNetworkConfiguration(subnetIds, sgIds pulumi.StringArray) mwaa.EnvironmentNetworkConfigurationArgs {
	return mwaa.EnvironmentNetworkConfigurationArgs{
		SubnetIds:        subnetIds,
		SecurityGroupIds: sgIds,
	}
}

func buildLoggingConfiguration(config *awsmwaaenvironmentv1.AwsMwaaEnvironmentLoggingConfiguration) mwaa.EnvironmentLoggingConfigurationPtrInput {
	if config == nil {
		return nil
	}

	args := &mwaa.EnvironmentLoggingConfigurationArgs{}

	if config.DagProcessingLogs != nil {
		args.DagProcessingLogs = buildLoggingModuleConfig(config.DagProcessingLogs)
	}
	if config.SchedulerLogs != nil {
		args.SchedulerLogs = buildSchedulerLoggingModuleConfig(config.SchedulerLogs)
	}
	if config.TaskLogs != nil {
		args.TaskLogs = buildTaskLoggingModuleConfig(config.TaskLogs)
	}
	if config.WebserverLogs != nil {
		args.WebserverLogs = buildWebserverLoggingModuleConfig(config.WebserverLogs)
	}
	if config.WorkerLogs != nil {
		args.WorkerLogs = buildWorkerLoggingModuleConfig(config.WorkerLogs)
	}

	return args
}

// Each Pulumi log module type is a distinct Go type despite having identical fields.
// Builder functions for each type are required for type safety.

func buildLoggingModuleConfig(config *awsmwaaenvironmentv1.AwsMwaaEnvironmentLoggingModuleConfig) mwaa.EnvironmentLoggingConfigurationDagProcessingLogsPtrInput {
	args := &mwaa.EnvironmentLoggingConfigurationDagProcessingLogsArgs{
		Enabled: pulumi.BoolPtr(config.Enabled),
	}
	if config.LogLevel != "" {
		args.LogLevel = pulumi.StringPtr(config.LogLevel)
	}
	return args
}

func buildSchedulerLoggingModuleConfig(config *awsmwaaenvironmentv1.AwsMwaaEnvironmentLoggingModuleConfig) mwaa.EnvironmentLoggingConfigurationSchedulerLogsPtrInput {
	args := &mwaa.EnvironmentLoggingConfigurationSchedulerLogsArgs{
		Enabled: pulumi.BoolPtr(config.Enabled),
	}
	if config.LogLevel != "" {
		args.LogLevel = pulumi.StringPtr(config.LogLevel)
	}
	return args
}

func buildTaskLoggingModuleConfig(config *awsmwaaenvironmentv1.AwsMwaaEnvironmentLoggingModuleConfig) mwaa.EnvironmentLoggingConfigurationTaskLogsPtrInput {
	args := &mwaa.EnvironmentLoggingConfigurationTaskLogsArgs{
		Enabled: pulumi.BoolPtr(config.Enabled),
	}
	if config.LogLevel != "" {
		args.LogLevel = pulumi.StringPtr(config.LogLevel)
	}
	return args
}

func buildWebserverLoggingModuleConfig(config *awsmwaaenvironmentv1.AwsMwaaEnvironmentLoggingModuleConfig) mwaa.EnvironmentLoggingConfigurationWebserverLogsPtrInput {
	args := &mwaa.EnvironmentLoggingConfigurationWebserverLogsArgs{
		Enabled: pulumi.BoolPtr(config.Enabled),
	}
	if config.LogLevel != "" {
		args.LogLevel = pulumi.StringPtr(config.LogLevel)
	}
	return args
}

func buildWorkerLoggingModuleConfig(config *awsmwaaenvironmentv1.AwsMwaaEnvironmentLoggingModuleConfig) mwaa.EnvironmentLoggingConfigurationWorkerLogsPtrInput {
	args := &mwaa.EnvironmentLoggingConfigurationWorkerLogsArgs{
		Enabled: pulumi.BoolPtr(config.Enabled),
	}
	if config.LogLevel != "" {
		args.LogLevel = pulumi.StringPtr(config.LogLevel)
	}
	return args
}

package module

import (
	"github.com/pkg/errors"
	awssagemakerendpointv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakerendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/decimalfloat"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sagemaker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// endpoint creates the endpoint configuration and the endpoint and
// exports outputs.
//
// Lifecycle facts the renders below depend on:
//   - the endpoint CONFIGURATION is immutable upstream (every argument
//     ForceNew), while the ENDPOINT's pointer to it updates in place
//     (UpdateEndpoint, optionally shaped by deployment_config). The
//     declarative fold therefore rolls configurations: NamePrefix +
//     Pulumi's default create-before-delete replacement mint a NEW
//     suffixed configuration on any capacity change, UpdateEndpoint
//     repoints, and only then is the old configuration deleted -- the
//     endpoint never references a deleted configuration (AWS's own
//     documented pattern);
//   - variant names default deterministically per position (locals) so
//     re-previews never regenerate them;
//   - the capacity-reservation preference has ONE legal value
//     ("capacity-reservations-only") -- the module owns the constant and
//     sends it exactly when an ML reservation ARN is configured.
func endpoint(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	configArgs := &sagemaker.EndpointConfigurationArgs{
		// Suffixed name so a changed configuration can coexist with its
		// predecessor during the endpoint repoint.
		NamePrefix: pulumi.String(locals.EndpointName + "-cfg-"),
		Tags:       pulumi.ToStringMap(locals.AwsTags),
	}

	// Required only for inference-component endpoints (variants without
	// a model - spec-validated).
	if spec.ExecutionRoleArn.GetValue() != "" {
		configArgs.ExecutionRoleArn = pulumi.String(spec.ExecutionRoleArn.GetValue())
	}
	if spec.KmsKeyArn.GetValue() != "" {
		configArgs.KmsKeyArn = pulumi.String(spec.KmsKeyArn.GetValue())
	}

	var productionVariants sagemaker.EndpointConfigurationProductionVariantArray
	for i, v := range spec.ProductionVariants {
		productionVariants = append(productionVariants, productionVariantArgs(v, i))
	}
	configArgs.ProductionVariants = productionVariants

	var shadowVariants sagemaker.EndpointConfigurationShadowProductionVariantArray
	for i, v := range spec.ShadowVariants {
		shadowVariants = append(shadowVariants, shadowVariantArgs(v, i))
	}
	if len(shadowVariants) > 0 {
		configArgs.ShadowProductionVariants = shadowVariants
	}

	// Queue requests and deliver responses to S3.
	if spec.AsyncInference != nil {
		configArgs.AsyncInferenceConfig = asyncInferenceArgs(spec.AsyncInference)
	}

	// Capture request/response payloads to S3 (the Model Monitor feed).
	if spec.DataCapture != nil {
		configArgs.DataCaptureConfig = dataCaptureArgs(spec.DataCapture)
	}

	createdConfig, err := sagemaker.NewEndpointConfiguration(ctx, locals.EndpointName+"-config", configArgs,
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create endpoint configuration")
	}

	endpointArgs := &sagemaker.EndpointArgs{
		// The component's name IS the endpoint name.
		Name: pulumi.String(locals.EndpointName),
		// Updates in place - the config roll's repoint edge.
		EndpointConfigName: createdConfig.Name,
		Tags:               pulumi.ToStringMap(locals.AwsTags),
	}

	// How UpdateEndpoint rolls new capacity (exactly-one strategy,
	// spec-validated).
	if spec.Deployment != nil {
		endpointArgs.DeploymentConfig = deploymentArgs(spec.Deployment)
	}

	createdEndpoint, err := sagemaker.NewEndpoint(ctx, locals.EndpointName, endpointArgs,
		pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{createdConfig}))
	if err != nil {
		return errors.Wrap(err, "create endpoint")
	}

	ctx.Export(OpEndpointName, createdEndpoint.Name)
	ctx.Export(OpEndpointArn, createdEndpoint.Arn)
	ctx.Export(OpEndpointConfigName, createdConfig.Name)
	ctx.Export(OpEndpointConfigArn, createdConfig.Arn)

	return nil
}

func productionVariantArgs(v *awssagemakerendpointv1alpha1.AwsSagemakerEndpointVariant, index int) *sagemaker.EndpointConfigurationProductionVariantArgs {
	args := &sagemaker.EndpointConfigurationProductionVariantArgs{
		VariantName: pulumi.String(resolvedVariantName(v, index, false)),
	}
	if v.Model.GetValue() != "" {
		args.ModelName = pulumi.String(v.Model.GetValue())
	}
	if v.InstanceType != "" {
		args.InstanceType = pulumi.String(v.InstanceType)
	}
	if v.InitialInstanceCount != nil {
		args.InitialInstanceCount = pulumi.Int(int(*v.InitialInstanceCount))
	}
	if v.InitialVariantWeight != nil {
		args.InitialVariantWeight = pulumi.Float64(decimalfloat.Float64(*v.InitialVariantWeight))
	}
	if v.AcceleratorType != "" {
		args.AcceleratorType = pulumi.String(v.AcceleratorType)
	}
	if v.InferenceAmiVersion != "" {
		args.InferenceAmiVersion = pulumi.String(v.InferenceAmiVersion)
	}
	if v.EnableSsmAccess {
		args.EnableSsmAccess = pulumi.Bool(true)
	}
	if v.VolumeSizeGb != nil {
		args.VolumeSizeInGb = pulumi.Int(int(*v.VolumeSizeGb))
	}
	if v.ContainerStartupHealthCheckTimeoutSeconds != nil {
		args.ContainerStartupHealthCheckTimeoutInSeconds = pulumi.Int(int(*v.ContainerStartupHealthCheckTimeoutSeconds))
	}
	if v.ModelDataDownloadTimeoutSeconds != nil {
		args.ModelDataDownloadTimeoutInSeconds = pulumi.Int(int(*v.ModelDataDownloadTimeoutSeconds))
	}
	if v.Serverless != nil {
		serverless := &sagemaker.EndpointConfigurationProductionVariantServerlessConfigArgs{
			MaxConcurrency: pulumi.Int(int(v.Serverless.MaxConcurrency)),
			MemorySizeInMb: pulumi.Int(int(v.Serverless.MemorySizeMb)),
		}
		if v.Serverless.ProvisionedConcurrency != nil {
			serverless.ProvisionedConcurrency = pulumi.Int(int(*v.Serverless.ProvisionedConcurrency))
		}
		args.ServerlessConfig = serverless
	}
	if v.ManagedInstanceScaling != nil {
		scaling := &sagemaker.EndpointConfigurationProductionVariantManagedInstanceScalingArgs{}
		if v.ManagedInstanceScaling.Status != "" {
			scaling.Status = pulumi.String(v.ManagedInstanceScaling.Status)
		}
		if v.ManagedInstanceScaling.MinInstanceCount != nil {
			scaling.MinInstanceCount = pulumi.Int(int(*v.ManagedInstanceScaling.MinInstanceCount))
		}
		if v.ManagedInstanceScaling.MaxInstanceCount != nil {
			scaling.MaxInstanceCount = pulumi.Int(int(*v.ManagedInstanceScaling.MaxInstanceCount))
		}
		args.ManagedInstanceScaling = scaling
	}
	if v.RoutingStrategy != "" {
		args.RoutingConfigs = sagemaker.EndpointConfigurationProductionVariantRoutingConfigArray{
			&sagemaker.EndpointConfigurationProductionVariantRoutingConfigArgs{
				RoutingStrategy: pulumi.String(v.RoutingStrategy),
			},
		}
	}
	if v.CoreDump != nil {
		coreDump := &sagemaker.EndpointConfigurationProductionVariantCoreDumpConfigArgs{
			DestinationS3Uri: pulumi.String(v.CoreDump.DestinationS3Uri),
		}
		if v.CoreDump.KmsKeyArn.GetValue() != "" {
			coreDump.KmsKeyId = pulumi.String(v.CoreDump.KmsKeyArn.GetValue())
		}
		args.CoreDumpConfig = coreDump
	}
	if v.MlCapacityReservationArn != "" {
		args.CapacityReservationConfig = &sagemaker.EndpointConfigurationProductionVariantCapacityReservationConfigArgs{
			// The single legal preference value - the module's constant.
			CapacityReservationPreference: pulumi.String("capacity-reservations-only"),
			MlReservationArn:              pulumi.String(v.MlCapacityReservationArn),
		}
	}
	return args
}

func shadowVariantArgs(v *awssagemakerendpointv1alpha1.AwsSagemakerEndpointVariant, index int) *sagemaker.EndpointConfigurationShadowProductionVariantArgs {
	args := &sagemaker.EndpointConfigurationShadowProductionVariantArgs{
		VariantName: pulumi.String(resolvedVariantName(v, index, true)),
	}
	if v.Model.GetValue() != "" {
		args.ModelName = pulumi.String(v.Model.GetValue())
	}
	if v.InstanceType != "" {
		args.InstanceType = pulumi.String(v.InstanceType)
	}
	if v.InitialInstanceCount != nil {
		args.InitialInstanceCount = pulumi.Int(int(*v.InitialInstanceCount))
	}
	if v.InitialVariantWeight != nil {
		args.InitialVariantWeight = pulumi.Float64(decimalfloat.Float64(*v.InitialVariantWeight))
	}
	if v.AcceleratorType != "" {
		args.AcceleratorType = pulumi.String(v.AcceleratorType)
	}
	if v.InferenceAmiVersion != "" {
		args.InferenceAmiVersion = pulumi.String(v.InferenceAmiVersion)
	}
	if v.EnableSsmAccess {
		args.EnableSsmAccess = pulumi.Bool(true)
	}
	if v.VolumeSizeGb != nil {
		args.VolumeSizeInGb = pulumi.Int(int(*v.VolumeSizeGb))
	}
	if v.ContainerStartupHealthCheckTimeoutSeconds != nil {
		args.ContainerStartupHealthCheckTimeoutInSeconds = pulumi.Int(int(*v.ContainerStartupHealthCheckTimeoutSeconds))
	}
	if v.ModelDataDownloadTimeoutSeconds != nil {
		args.ModelDataDownloadTimeoutInSeconds = pulumi.Int(int(*v.ModelDataDownloadTimeoutSeconds))
	}
	if v.Serverless != nil {
		serverless := &sagemaker.EndpointConfigurationShadowProductionVariantServerlessConfigArgs{
			MaxConcurrency: pulumi.Int(int(v.Serverless.MaxConcurrency)),
			MemorySizeInMb: pulumi.Int(int(v.Serverless.MemorySizeMb)),
		}
		if v.Serverless.ProvisionedConcurrency != nil {
			serverless.ProvisionedConcurrency = pulumi.Int(int(*v.Serverless.ProvisionedConcurrency))
		}
		args.ServerlessConfig = serverless
	}
	if v.ManagedInstanceScaling != nil {
		scaling := &sagemaker.EndpointConfigurationShadowProductionVariantManagedInstanceScalingArgs{}
		if v.ManagedInstanceScaling.Status != "" {
			scaling.Status = pulumi.String(v.ManagedInstanceScaling.Status)
		}
		if v.ManagedInstanceScaling.MinInstanceCount != nil {
			scaling.MinInstanceCount = pulumi.Int(int(*v.ManagedInstanceScaling.MinInstanceCount))
		}
		if v.ManagedInstanceScaling.MaxInstanceCount != nil {
			scaling.MaxInstanceCount = pulumi.Int(int(*v.ManagedInstanceScaling.MaxInstanceCount))
		}
		args.ManagedInstanceScaling = scaling
	}
	if v.RoutingStrategy != "" {
		args.RoutingConfigs = sagemaker.EndpointConfigurationShadowProductionVariantRoutingConfigArray{
			&sagemaker.EndpointConfigurationShadowProductionVariantRoutingConfigArgs{
				RoutingStrategy: pulumi.String(v.RoutingStrategy),
			},
		}
	}
	if v.CoreDump != nil {
		coreDump := &sagemaker.EndpointConfigurationShadowProductionVariantCoreDumpConfigArgs{
			DestinationS3Uri: pulumi.String(v.CoreDump.DestinationS3Uri),
		}
		if v.CoreDump.KmsKeyArn.GetValue() != "" {
			coreDump.KmsKeyId = pulumi.String(v.CoreDump.KmsKeyArn.GetValue())
		}
		args.CoreDumpConfig = coreDump
	}
	if v.MlCapacityReservationArn != "" {
		args.CapacityReservationConfig = &sagemaker.EndpointConfigurationShadowProductionVariantCapacityReservationConfigArgs{
			CapacityReservationPreference: pulumi.String("capacity-reservations-only"),
			MlReservationArn:              pulumi.String(v.MlCapacityReservationArn),
		}
	}
	return args
}

func asyncInferenceArgs(a *awssagemakerendpointv1alpha1.AwsSagemakerEndpointAsyncInference) *sagemaker.EndpointConfigurationAsyncInferenceConfigArgs {
	outputConfig := &sagemaker.EndpointConfigurationAsyncInferenceConfigOutputConfigArgs{
		S3OutputPath: pulumi.String(a.OutputS3Path),
	}
	if a.FailureS3Path != "" {
		outputConfig.S3FailurePath = pulumi.String(a.FailureS3Path)
	}
	if a.KmsKeyArn.GetValue() != "" {
		outputConfig.KmsKeyId = pulumi.String(a.KmsKeyArn.GetValue())
	}
	if a.SuccessTopicArn.GetValue() != "" || a.ErrorTopicArn.GetValue() != "" || len(a.IncludeInferenceResponseIn) > 0 {
		notification := &sagemaker.EndpointConfigurationAsyncInferenceConfigOutputConfigNotificationConfigArgs{}
		if a.SuccessTopicArn.GetValue() != "" {
			notification.SuccessTopic = pulumi.String(a.SuccessTopicArn.GetValue())
		}
		if a.ErrorTopicArn.GetValue() != "" {
			notification.ErrorTopic = pulumi.String(a.ErrorTopicArn.GetValue())
		}
		if len(a.IncludeInferenceResponseIn) > 0 {
			notification.IncludeInferenceResponseIns = pulumi.ToStringArray(a.IncludeInferenceResponseIn)
		}
		outputConfig.NotificationConfig = notification
	}

	args := &sagemaker.EndpointConfigurationAsyncInferenceConfigArgs{
		OutputConfig: outputConfig,
	}
	if a.MaxConcurrentInvocationsPerInstance != nil {
		args.ClientConfig = &sagemaker.EndpointConfigurationAsyncInferenceConfigClientConfigArgs{
			MaxConcurrentInvocationsPerInstance: pulumi.Int(int(*a.MaxConcurrentInvocationsPerInstance)),
		}
	}
	return args
}

func dataCaptureArgs(d *awssagemakerendpointv1alpha1.AwsSagemakerEndpointDataCapture) *sagemaker.EndpointConfigurationDataCaptureConfigArgs {
	var captureOptions sagemaker.EndpointConfigurationDataCaptureConfigCaptureOptionArray
	for _, mode := range d.CaptureModes {
		captureOptions = append(captureOptions, &sagemaker.EndpointConfigurationDataCaptureConfigCaptureOptionArgs{
			CaptureMode: pulumi.String(mode),
		})
	}

	args := &sagemaker.EndpointConfigurationDataCaptureConfigArgs{
		DestinationS3Uri:          pulumi.String(d.DestinationS3Uri),
		InitialSamplingPercentage: pulumi.Int(int(d.InitialSamplingPercentage)),
		CaptureOptions:            captureOptions,
	}
	if d.EnableCapture {
		args.EnableCapture = pulumi.Bool(true)
	}
	if d.KmsKeyArn.GetValue() != "" {
		args.KmsKeyId = pulumi.String(d.KmsKeyArn.GetValue())
	}
	if len(d.CsvContentTypes) > 0 || len(d.JsonContentTypes) > 0 {
		header := &sagemaker.EndpointConfigurationDataCaptureConfigCaptureContentTypeHeaderArgs{}
		if len(d.CsvContentTypes) > 0 {
			header.CsvContentTypes = pulumi.ToStringArray(d.CsvContentTypes)
		}
		if len(d.JsonContentTypes) > 0 {
			header.JsonContentTypes = pulumi.ToStringArray(d.JsonContentTypes)
		}
		args.CaptureContentTypeHeader = header
	}
	return args
}

func deploymentArgs(d *awssagemakerendpointv1alpha1.AwsSagemakerEndpointDeployment) *sagemaker.EndpointDeploymentConfigArgs {
	args := &sagemaker.EndpointDeploymentConfigArgs{}

	if d.BlueGreen != nil {
		routing := &sagemaker.EndpointDeploymentConfigBlueGreenUpdatePolicyTrafficRoutingConfigurationArgs{
			Type:                  pulumi.String(d.BlueGreen.TrafficRoutingType),
			WaitIntervalInSeconds: pulumi.Int(int(d.BlueGreen.WaitIntervalSeconds)),
		}
		if d.BlueGreen.CanarySize != nil {
			routing.CanarySize = &sagemaker.EndpointDeploymentConfigBlueGreenUpdatePolicyTrafficRoutingConfigurationCanarySizeArgs{
				Type:  pulumi.String(d.BlueGreen.CanarySize.Type),
				Value: pulumi.Int(int(d.BlueGreen.CanarySize.Value)),
			}
		}
		if d.BlueGreen.LinearStepSize != nil {
			routing.LinearStepSize = &sagemaker.EndpointDeploymentConfigBlueGreenUpdatePolicyTrafficRoutingConfigurationLinearStepSizeArgs{
				Type:  pulumi.String(d.BlueGreen.LinearStepSize.Type),
				Value: pulumi.Int(int(d.BlueGreen.LinearStepSize.Value)),
			}
		}
		blueGreen := &sagemaker.EndpointDeploymentConfigBlueGreenUpdatePolicyArgs{
			TrafficRoutingConfiguration: routing,
		}
		if d.BlueGreen.TerminationWaitSeconds != nil {
			blueGreen.TerminationWaitInSeconds = pulumi.Int(int(*d.BlueGreen.TerminationWaitSeconds))
		}
		if d.BlueGreen.MaximumExecutionTimeoutSeconds != nil {
			blueGreen.MaximumExecutionTimeoutInSeconds = pulumi.Int(int(*d.BlueGreen.MaximumExecutionTimeoutSeconds))
		}
		args.BlueGreenUpdatePolicy = blueGreen
	}

	if d.Rolling != nil {
		rolling := &sagemaker.EndpointDeploymentConfigRollingUpdatePolicyArgs{
			MaximumBatchSize: &sagemaker.EndpointDeploymentConfigRollingUpdatePolicyMaximumBatchSizeArgs{
				Type:  pulumi.String(d.Rolling.MaximumBatchSize.Type),
				Value: pulumi.Int(int(d.Rolling.MaximumBatchSize.Value)),
			},
			WaitIntervalInSeconds: pulumi.Int(int(d.Rolling.WaitIntervalSeconds)),
		}
		if d.Rolling.RollbackMaximumBatchSize != nil {
			rolling.RollbackMaximumBatchSize = &sagemaker.EndpointDeploymentConfigRollingUpdatePolicyRollbackMaximumBatchSizeArgs{
				Type:  pulumi.String(d.Rolling.RollbackMaximumBatchSize.Type),
				Value: pulumi.Int(int(d.Rolling.RollbackMaximumBatchSize.Value)),
			}
		}
		args.RollingUpdatePolicy = rolling
	}

	if len(d.AutoRollbackAlarmNames) > 0 {
		var alarms sagemaker.EndpointDeploymentConfigAutoRollbackConfigurationAlarmArray
		for _, name := range d.AutoRollbackAlarmNames {
			alarms = append(alarms, &sagemaker.EndpointDeploymentConfigAutoRollbackConfigurationAlarmArgs{
				AlarmName: pulumi.String(name),
			})
		}
		args.AutoRollbackConfiguration = &sagemaker.EndpointDeploymentConfigAutoRollbackConfigurationArgs{
			Alarms: alarms,
		}
	}

	return args
}

package module

import (
	"github.com/pkg/errors"
	azuremachinelearningbatchdeploymentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningbatchdeployment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazurenativeprovider"
	"github.com/pulumi/pulumi-azure-native-sdk/machinelearningservices/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremachinelearningbatchdeploymentv1alpha1.AzureMachineLearningBatchDeploymentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the azure-native provider from the IaC input via the shared
	// builder, which resolves the right credential mechanism (static client
	// secret, keyless web identity, or ambient chain). This kind rides
	// azure-native, not the classic provider: azurerm/classic carry NO ML
	// deployment resources, and azure-native's typed resources are
	// generated from the same ARM specification the Terraform module's
	// raw-API shape pins (tracked at
	// hashicorp/terraform-provider-azurerm#32011 with a mandatory move to
	// native resources when azurerm ships them).
	azureNativeProvider, err := pulumiazurenativeprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure-native provider")
	}

	spec := locals.AzureMachineLearningBatchDeployment.Spec

	resourceGroupName, workspaceName, endpointName, err := parseEndpointId(locals.EndpointId)
	if err != nil {
		return errors.Wrap(err, "failed to parse the batch endpoint ARM id")
	}

	// The ARM properties object, assembled field-by-field so unset
	// optionals are OMITTED (ARM applies its own defaults: miniBatchSize
	// 10, maxConcurrencyPerInstance 1, loggingLevel Info, outputAction
	// AppendRow, outputFileName predictions.csv) -- the same wire shape
	// the Terraform module's body builds. One SDK nuance: the typed args'
	// own Defaults() fills errorThreshold with -1 when unset, which IS
	// ARM's default -- wire-equivalent to the Terraform module's omission.
	deploymentProperties := &machinelearningservices.BatchDeploymentTypeArgs{}

	if spec.ComputeId.GetValue() != "" {
		deploymentProperties.Compute = pulumi.String(spec.ComputeId.GetValue())
	}

	if spec.EnvironmentId != "" {
		deploymentProperties.EnvironmentId = pulumi.String(spec.EnvironmentId)
	}

	if len(spec.EnvironmentVariables) > 0 {
		deploymentProperties.EnvironmentVariables = pulumi.ToStringMap(spec.EnvironmentVariables)
	}

	if spec.MiniBatchSize != nil {
		// ARM types miniBatchSize as an int64; the SDK's generated arg is
		// float64 -- the same JSON number on the wire.
		deploymentProperties.MiniBatchSize = pulumi.Float64(float64(*spec.MiniBatchSize))
	}

	if spec.MaxConcurrencyPerInstance != nil {
		deploymentProperties.MaxConcurrencyPerInstance = pulumi.Int(int(*spec.MaxConcurrencyPerInstance))
	}

	if spec.ErrorThreshold != nil {
		deploymentProperties.ErrorThreshold = pulumi.Int(int(*spec.ErrorThreshold))
	}

	if spec.OutputAction != "" {
		deploymentProperties.OutputAction = pulumi.String(spec.OutputAction)
	}

	if spec.OutputFileName != "" {
		deploymentProperties.OutputFileName = pulumi.String(spec.OutputFileName)
	}

	if spec.LoggingLevel != "" {
		deploymentProperties.LoggingLevel = pulumi.String(spec.LoggingLevel)
	}

	if spec.Description != "" {
		deploymentProperties.Description = pulumi.String(spec.Description)
	}

	if len(spec.Properties) > 0 {
		deploymentProperties.Properties = pulumi.ToStringMap(spec.Properties)
	}

	// The model reference: the variant block set in the spec IS the ARM
	// referenceType discriminator -- the spec's exactly-one rule
	// guarantees a single arm.
	if spec.Model != nil {
		switch {
		case spec.Model.Id != nil:
			deploymentProperties.Model = machinelearningservices.IdAssetReferenceArgs{
				ReferenceType: pulumi.String("Id"),
				AssetId:       pulumi.String(spec.Model.Id.AssetId),
			}
		case spec.Model.DataPath != nil:
			dataPath := machinelearningservices.DataPathAssetReferenceArgs{
				ReferenceType: pulumi.String("DataPath"),
			}
			if spec.Model.DataPath.DatastoreId != "" {
				dataPath.DatastoreId = pulumi.String(spec.Model.DataPath.DatastoreId)
			}
			if spec.Model.DataPath.Path != "" {
				dataPath.Path = pulumi.String(spec.Model.DataPath.Path)
			}
			deploymentProperties.Model = dataPath
		case spec.Model.OutputPath != nil:
			outputPath := machinelearningservices.OutputPathAssetReferenceArgs{
				ReferenceType: pulumi.String("OutputPath"),
			}
			if spec.Model.OutputPath.JobId != "" {
				outputPath.JobId = pulumi.String(spec.Model.OutputPath.JobId)
			}
			if spec.Model.OutputPath.Path != "" {
				outputPath.Path = pulumi.String(spec.Model.OutputPath.Path)
			}
			deploymentProperties.Model = outputPath
		}
	}

	// Per-job compute sizing. ARM's untyped resources.properties bag is
	// deliberately not modeled (recorded exclusion on the spec message).
	if spec.Resources != nil {
		resources := &machinelearningservices.DeploymentResourceConfigurationArgs{}
		if spec.Resources.InstanceCount != nil {
			resources.InstanceCount = pulumi.Int(int(*spec.Resources.InstanceCount))
		}
		if spec.Resources.InstanceType != "" {
			resources.InstanceType = pulumi.String(spec.Resources.InstanceType)
		}
		deploymentProperties.Resources = resources
	}

	if spec.RetrySettings != nil {
		retrySettings := &machinelearningservices.BatchRetrySettingsArgs{}
		if spec.RetrySettings.MaxRetries != nil {
			retrySettings.MaxRetries = pulumi.Int(int(*spec.RetrySettings.MaxRetries))
		}
		if spec.RetrySettings.Timeout != "" {
			retrySettings.Timeout = pulumi.String(spec.RetrySettings.Timeout)
		}
		deploymentProperties.RetrySettings = retrySettings
	}

	if spec.CodeConfiguration != nil {
		codeConfiguration := &machinelearningservices.CodeConfigurationArgs{
			ScoringScript: pulumi.String(spec.CodeConfiguration.ScoringScript),
		}
		if spec.CodeConfiguration.CodeId != "" {
			codeConfiguration.CodeId = pulumi.String(spec.CodeConfiguration.CodeId)
		}
		deploymentProperties.CodeConfiguration = codeConfiguration
	}

	// The PipelineComponent deployment type: present only when the spec's
	// block is -- absent means ARM's default Model type.
	if spec.PipelineComponent != nil {
		pipelineComponent := &machinelearningservices.BatchPipelineComponentDeploymentConfigurationArgs{
			DeploymentConfigurationType: pulumi.String("PipelineComponent"),
			ComponentId: &machinelearningservices.IdAssetReferenceArgs{
				ReferenceType: pulumi.String("Id"),
				AssetId:       pulumi.String(spec.PipelineComponent.ComponentId),
			},
		}
		if len(spec.PipelineComponent.Settings) > 0 {
			pipelineComponent.Settings = pulumi.ToStringMap(spec.PipelineComponent.Settings)
		}
		if len(spec.PipelineComponent.JobTags) > 0 {
			pipelineComponent.Tags = pulumi.ToStringMap(spec.PipelineComponent.JobTags)
		}
		if spec.PipelineComponent.JobDescription != "" {
			pipelineComponent.Description = pulumi.String(spec.PipelineComponent.JobDescription)
		}
		deploymentProperties.DeploymentConfiguration = pipelineComponent
	}

	// Create the batch deployment -- the job recipe behind its endpoint's
	// address, as an ARM child of the endpoint. Everything in the recipe
	// updates in place via full PUT (ARM flags nothing immutable on this
	// surface); name, region and endpoint replace the deployment. Nothing
	// runs or bills at create time -- each endpoint invocation
	// materializes a job from this recipe. The envelope's sku is
	// deliberately absent: batch scale lives on resources.instanceCount
	// per job, not on an autoscaling SKU (the online deployment's dial).
	createdDeployment, err := machinelearningservices.NewBatchDeployment(ctx,
		spec.Name,
		&machinelearningservices.BatchDeploymentArgs{
			DeploymentName:            pulumi.String(spec.Name),
			EndpointName:              pulumi.String(endpointName),
			ResourceGroupName:         pulumi.String(resourceGroupName),
			WorkspaceName:             pulumi.String(workspaceName),
			Location:                  pulumi.String(spec.Region),
			BatchDeploymentProperties: deploymentProperties,
			Tags:                      pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureNativeProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create batch deployment %s", spec.Name)
	}

	ctx.Export(OpBatchDeploymentId, createdDeployment.ID())
	ctx.Export(OpBatchDeploymentName, createdDeployment.Name)

	return nil
}

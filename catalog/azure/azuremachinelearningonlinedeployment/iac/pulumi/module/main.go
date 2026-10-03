package module

import (
	"github.com/pkg/errors"
	azuremachinelearningonlinedeploymentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningonlinedeployment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazurenativeprovider"
	"github.com/pulumi/pulumi-azure-native-sdk/machinelearningservices/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremachinelearningonlinedeploymentv1alpha1.AzureMachineLearningOnlineDeploymentIacInput) error {
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

	spec := locals.AzureMachineLearningOnlineDeployment.Spec

	resourceGroupName, workspaceName, endpointName, err := parseEndpointId(locals.EndpointId)
	if err != nil {
		return errors.Wrap(err, "failed to parse the endpoint ARM id")
	}

	// The ARM properties object, assembled field-by-field so unset
	// optionals are OMITTED (ARM applies its own defaults) -- the same
	// wire shape the Terraform module's body builds. endpointComputeType
	// is PINNED to Managed: this kind models the managed compute type
	// only (the Kubernetes and AzureMLCompute variants are a recorded
	// deferral). Scale settings are deliberately absent: the managed
	// variant's only legal mode is Default (fixed instance count via the
	// SKU capacity below); TargetUtilization is Kubernetes-only.
	deploymentProperties := &machinelearningservices.ManagedOnlineDeploymentArgs{
		EndpointComputeType: pulumi.String("Managed"),
		AppInsightsEnabled:  pulumi.Bool(spec.AppInsightsEnabled),
	}

	if spec.InstanceType != "" {
		deploymentProperties.InstanceType = pulumi.String(spec.InstanceType)
	}

	if spec.Model != "" {
		deploymentProperties.Model = pulumi.String(spec.Model)
	}

	if spec.ModelMountPath != "" {
		deploymentProperties.ModelMountPath = pulumi.String(spec.ModelMountPath)
	}

	if spec.EnvironmentId != "" {
		deploymentProperties.EnvironmentId = pulumi.String(spec.EnvironmentId)
	}

	if len(spec.EnvironmentVariables) > 0 {
		deploymentProperties.EnvironmentVariables = pulumi.ToStringMap(spec.EnvironmentVariables)
	}

	if spec.EgressPublicNetworkAccessEnabled != nil {
		wire := "Disabled"
		if *spec.EgressPublicNetworkAccessEnabled {
			wire = "Enabled"
		}
		deploymentProperties.EgressPublicNetworkAccess = pulumi.String(wire)
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

	deploymentProperties.LivenessProbe = buildProbeArgs(spec.LivenessProbe)
	deploymentProperties.ReadinessProbe = buildProbeArgs(spec.ReadinessProbe)
	deploymentProperties.StartupProbe = buildProbeArgs(spec.StartupProbe)

	if spec.RequestSettings != nil {
		requestSettings := &machinelearningservices.OnlineRequestSettingsArgs{}
		if spec.RequestSettings.MaxConcurrentRequestsPerInstance != nil {
			requestSettings.MaxConcurrentRequestsPerInstance = pulumi.Int(int(*spec.RequestSettings.MaxConcurrentRequestsPerInstance))
		}
		if spec.RequestSettings.RequestTimeout != "" {
			requestSettings.RequestTimeout = pulumi.String(spec.RequestSettings.RequestTimeout)
		}
		deploymentProperties.RequestSettings = requestSettings
	}

	if spec.DataCollector != nil {
		collections := machinelearningservices.CollectionMap{}
		for name, collection := range spec.DataCollector.Collections {
			// Each collection's two-value Enabled/Disabled surface rides
			// the spec's bool.
			mode := "Disabled"
			if collection.Enabled {
				mode = "Enabled"
			}
			collectionArgs := &machinelearningservices.CollectionArgs{
				DataCollectionMode: pulumi.String(mode),
			}
			if collection.DataId != "" {
				collectionArgs.DataId = pulumi.String(collection.DataId)
			}
			if collection.ClientId != "" {
				collectionArgs.ClientId = pulumi.String(collection.ClientId)
			}
			if collection.SamplingRate != nil {
				collectionArgs.SamplingRate = pulumi.Float64(*collection.SamplingRate)
			}
			collections[name] = collectionArgs
		}
		dataCollector := &machinelearningservices.DataCollectorArgs{
			Collections: collections,
		}
		if spec.DataCollector.RollingRate != "" {
			dataCollector.RollingRate = pulumi.String(spec.DataCollector.RollingRate)
		}
		if spec.DataCollector.RequestLogging != nil && len(spec.DataCollector.RequestLogging.CaptureHeaders) > 0 {
			dataCollector.RequestLogging = &machinelearningservices.RequestLoggingArgs{
				CaptureHeaders: pulumi.ToStringArray(spec.DataCollector.RequestLogging.CaptureHeaders),
			}
		}
		deploymentProperties.DataCollector = dataCollector
	}

	// The service's ARM contract for autoscaling: managed deployments
	// carry SKU name "Default" with capacity as the instance count -- the
	// one dial the service scales without touching the deployment's
	// containers.
	instanceCount := 1
	if spec.InstanceCount != nil {
		instanceCount = int(*spec.InstanceCount)
	}

	// Create the managed online deployment -- a running copy of a model
	// behind its endpoint's address. Updates go through full PUT (the
	// service rolls the deployment's containers); name, region and
	// endpoint replace the deployment.
	createdDeployment, err := machinelearningservices.NewOnlineDeployment(ctx,
		spec.Name,
		&machinelearningservices.OnlineDeploymentArgs{
			DeploymentName:    pulumi.String(spec.Name),
			EndpointName:      pulumi.String(endpointName),
			ResourceGroupName: pulumi.String(resourceGroupName),
			WorkspaceName:     pulumi.String(workspaceName),
			Location:          pulumi.String(spec.Region),
			Sku: &machinelearningservices.SkuArgs{
				Name:     pulumi.String("Default"),
				Capacity: pulumi.Int(instanceCount),
			},
			OnlineDeploymentProperties: deploymentProperties,
			Tags:                       pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureNativeProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create online deployment %s", spec.Name)
	}

	ctx.Export(OpOnlineDeploymentId, createdDeployment.ID())
	ctx.Export(OpOnlineDeploymentName, createdDeployment.Name)

	return nil
}

// buildProbeArgs assembles one probe's ARM shape, omitting unset
// optionals so the service's own defaults apply. All three probes share
// this shape and these service defaults.
func buildProbeArgs(probe *azuremachinelearningonlinedeploymentv1alpha1.AzureMachineLearningOnlineDeploymentProbeSettings) *machinelearningservices.ProbeSettingsArgs {
	if probe == nil {
		return nil
	}
	probeArgs := &machinelearningservices.ProbeSettingsArgs{}
	if probe.FailureThreshold != nil {
		probeArgs.FailureThreshold = pulumi.Int(int(*probe.FailureThreshold))
	}
	if probe.SuccessThreshold != nil {
		probeArgs.SuccessThreshold = pulumi.Int(int(*probe.SuccessThreshold))
	}
	if probe.InitialDelay != "" {
		probeArgs.InitialDelay = pulumi.String(probe.InitialDelay)
	}
	if probe.Period != "" {
		probeArgs.Period = pulumi.String(probe.Period)
	}
	if probe.Timeout != "" {
		probeArgs.Timeout = pulumi.String(probe.Timeout)
	}
	return probeArgs
}

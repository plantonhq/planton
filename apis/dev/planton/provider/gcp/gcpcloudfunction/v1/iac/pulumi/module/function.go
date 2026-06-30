package module

import (
	"fmt"

	gcpcloudfunctionv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudfunction/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudfunctionsv2"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createCloudFunction creates a Google Cloud Function (Gen 2) resource.
func createCloudFunction(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*cloudfunctionsv2.Function, error) {

	spec := locals.GcpCloudFunction.Spec

	// Build config - storage source
	storageSourceArgs := &cloudfunctionsv2.FunctionBuildConfigSourceStorageSourceArgs{
		Bucket: pulumi.String(spec.BuildConfig.Source.Bucket),
		Object: pulumi.String(spec.BuildConfig.Source.Object),
	}

	// Add optional generation if specified
	if spec.BuildConfig.Source.Generation != nil && *spec.BuildConfig.Source.Generation > 0 {
		storageSourceArgs.Generation = pulumi.Int(*spec.BuildConfig.Source.Generation)
	}

	buildConfig := &cloudfunctionsv2.FunctionBuildConfigArgs{
		Runtime:    pulumi.String(spec.BuildConfig.Runtime),
		EntryPoint: pulumi.String(spec.BuildConfig.EntryPoint),
		Source: &cloudfunctionsv2.FunctionBuildConfigSourceArgs{
			StorageSource: storageSourceArgs,
		},
	}

	// Add build environment variables if specified
	if len(spec.BuildConfig.BuildEnvironmentVariables) > 0 {
		buildConfig.EnvironmentVariables = pulumi.ToStringMap(spec.BuildConfig.BuildEnvironmentVariables)
	}

	// Service config with defaults
	serviceConfig := &cloudfunctionsv2.FunctionServiceConfigArgs{
		AllTrafficOnLatestRevision: pulumi.Bool(true),
	}

	// Apply service configuration if provided
	if spec.ServiceConfig != nil {
		sc := spec.ServiceConfig

		// Memory (default 256MB)
		memoryMb := int32(256)
		if sc.AvailableMemoryMb > 0 {
			memoryMb = sc.AvailableMemoryMb
		}
		serviceConfig.AvailableMemory = pulumi.String(fmt.Sprintf("%dM", memoryMb))

		// Timeout (default 60 seconds)
		timeoutSeconds := int32(60)
		if sc.TimeoutSeconds > 0 {
			timeoutSeconds = sc.TimeoutSeconds
		}
		serviceConfig.TimeoutSeconds = pulumi.Int(int(timeoutSeconds))

		// Max concurrency (default 80)
		maxConcurrency := int32(80)
		if sc.MaxInstanceRequestConcurrency > 0 {
			maxConcurrency = sc.MaxInstanceRequestConcurrency
		}
		serviceConfig.MaxInstanceRequestConcurrency = pulumi.Int(int(maxConcurrency))

		// Service account
		if sc.ServiceAccountEmail != "" {
			serviceConfig.ServiceAccountEmail = pulumi.String(sc.ServiceAccountEmail)
		}

		// Environment variables
		if len(sc.EnvironmentVariables) > 0 {
			serviceConfig.EnvironmentVariables = pulumi.ToStringMap(sc.EnvironmentVariables)
		}

		// Secret environment variables
		if len(sc.SecretEnvironmentVariables) > 0 {
			secretEnvVars := cloudfunctionsv2.FunctionServiceConfigSecretEnvironmentVariableArray{}
			for key, secretName := range sc.SecretEnvironmentVariables {
				secretEnvVars = append(secretEnvVars, &cloudfunctionsv2.FunctionServiceConfigSecretEnvironmentVariableArgs{
					Key:       pulumi.String(key),
					ProjectId: pulumi.String(spec.ProjectId.GetValue()),
					Secret:    pulumi.String(secretName),
					Version:   pulumi.String("latest"),
				})
			}
			serviceConfig.SecretEnvironmentVariables = secretEnvVars
		}

		// VPC connector
		if sc.VpcConnector != "" {
			serviceConfig.VpcConnector = pulumi.String(sc.VpcConnector)

			// VPC egress settings (default PRIVATE_RANGES_ONLY)
			vpcEgress := "PRIVATE_RANGES_ONLY"
			if sc.VpcConnectorEgressSettings == gcpcloudfunctionv1.GcpCloudFunctionVpcEgressSetting_ALL_TRAFFIC {
				vpcEgress = "ALL_TRAFFIC"
			}
			serviceConfig.VpcConnectorEgressSettings = pulumi.String(vpcEgress)
		}

		// Ingress settings (default ALLOW_ALL)
		ingressSettings := "ALLOW_ALL"
		switch sc.IngressSettings {
		case gcpcloudfunctionv1.GcpCloudFunctionIngressSetting_ALLOW_INTERNAL_ONLY:
			ingressSettings = "ALLOW_INTERNAL_ONLY"
		case gcpcloudfunctionv1.GcpCloudFunctionIngressSetting_ALLOW_INTERNAL_AND_GCLB:
			ingressSettings = "ALLOW_INTERNAL_AND_GCLB"
		}
		serviceConfig.IngressSettings = pulumi.String(ingressSettings)

		// Scaling configuration
		if sc.Scaling != nil {
			minInstances := int32(0)
			if sc.Scaling.MinInstanceCount > 0 {
				minInstances = sc.Scaling.MinInstanceCount
			}
			serviceConfig.MinInstanceCount = pulumi.Int(int(minInstances))

			maxInstances := int32(100)
			if sc.Scaling.MaxInstanceCount > 0 {
				maxInstances = sc.Scaling.MaxInstanceCount
			}
			serviceConfig.MaxInstanceCount = pulumi.Int(int(maxInstances))
		} else {
			serviceConfig.MinInstanceCount = pulumi.Int(0)
			serviceConfig.MaxInstanceCount = pulumi.Int(100)
		}
	} else {
		// Apply defaults if service_config is nil
		serviceConfig.AvailableMemory = pulumi.String("256M")
		serviceConfig.TimeoutSeconds = pulumi.Int(60)
		serviceConfig.MaxInstanceRequestConcurrency = pulumi.Int(80)
		serviceConfig.IngressSettings = pulumi.String("ALLOW_ALL")
		serviceConfig.MinInstanceCount = pulumi.Int(0)
		serviceConfig.MaxInstanceCount = pulumi.Int(100)
	}

	// Prepare function args
	functionArgs := &cloudfunctionsv2.FunctionArgs{
		Name:          pulumi.String(locals.FunctionName),
		Project:       pulumi.String(spec.ProjectId.GetValue()),
		Location:      pulumi.String(spec.Region),
		BuildConfig:   buildConfig,
		ServiceConfig: serviceConfig,
		Labels:        pulumi.ToStringMap(locals.GcpLabels),
	}

	// Add event trigger if this is an event-driven function
	if !locals.IsHttpTrigger && spec.Trigger != nil && spec.Trigger.EventTrigger != nil {
		et := spec.Trigger.EventTrigger

		eventTrigger := &cloudfunctionsv2.FunctionEventTriggerArgs{
			EventType: pulumi.String(et.EventType),
		}

		// Trigger region (defaults to function region)
		if et.TriggerRegion != "" {
			eventTrigger.TriggerRegion = pulumi.String(et.TriggerRegion)
		} else {
			eventTrigger.TriggerRegion = pulumi.String(spec.Region)
		}

		// Pub/Sub topic
		if et.PubsubTopic != "" {
			eventTrigger.PubsubTopic = pulumi.String(et.PubsubTopic)
		}

		// Retry policy (default DO_NOT_RETRY)
		retryPolicy := "RETRY_POLICY_DO_NOT_RETRY"
		if et.RetryPolicy == gcpcloudfunctionv1.GcpCloudFunctionRetryPolicy_RETRY_POLICY_RETRY {
			retryPolicy = "RETRY_POLICY_RETRY"
		}
		eventTrigger.RetryPolicy = pulumi.String(retryPolicy)

		// Service account for trigger
		if et.ServiceAccountEmail != "" {
			eventTrigger.ServiceAccountEmail = pulumi.String(et.ServiceAccountEmail)
		}

		// Event filters
		if len(et.EventFilters) > 0 {
			eventFilters := cloudfunctionsv2.FunctionEventTriggerEventFilterArray{}
			for _, filter := range et.EventFilters {
				eventFilter := &cloudfunctionsv2.FunctionEventTriggerEventFilterArgs{
					Attribute: pulumi.String(filter.Attribute),
					Value:     pulumi.String(filter.Value),
				}
				if filter.Operator != nil && *filter.Operator != "" {
					eventFilter.Operator = pulumi.String(*filter.Operator)
				}
				eventFilters = append(eventFilters, eventFilter)
			}
			eventTrigger.EventFilters = eventFilters
		}

		functionArgs.EventTrigger = eventTrigger
	}

	// Create the function
	function, err := cloudfunctionsv2.NewFunction(
		ctx,
		locals.FunctionName,
		functionArgs,
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return nil, err
	}

	// Add IAM policy binding for public access if required (HTTP only)
	if locals.IsHttpTrigger &&
		spec.ServiceConfig != nil &&
		spec.ServiceConfig.AllowUnauthenticated {

		_, err = cloudrunv2.NewServiceIamMember(
			ctx,
			fmt.Sprintf("%s-public-invoker", locals.FunctionName),
			&cloudrunv2.ServiceIamMemberArgs{
				Project:  pulumi.String(spec.ProjectId.GetValue()),
				Location: pulumi.String(spec.Region),
				Name:     function.Name,
				Role:     pulumi.String("roles/run.invoker"),
				Member:   pulumi.String("allUsers"),
			},
			pulumi.Provider(gcpProvider),
		)
		if err != nil {
			return nil, err
		}
	}

	return function, nil
}

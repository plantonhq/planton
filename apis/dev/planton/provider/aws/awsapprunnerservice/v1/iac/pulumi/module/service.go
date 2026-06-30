package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apprunner"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// service creates the main AWS App Runner Service resource. It builds the source
// configuration (image or code), instance configuration, health check, networking,
// encryption, and observability settings from the spec, then exports all stack outputs.
func service(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	createdVpcConnector *apprunner.VpcConnector,
	createdAutoScaling *apprunner.AutoScalingConfigurationVersion,
) error {
	spec := locals.AwsAppRunnerService.Spec
	resourceName := locals.AwsAppRunnerService.Metadata.Name

	// --- Source configuration ---
	sourceConfig := buildSourceConfiguration(locals)

	// --- Service args ---
	args := &apprunner.ServiceArgs{
		ServiceName:         pulumi.String(resourceName),
		SourceConfiguration: sourceConfig,
		Tags:                pulumi.ToStringMap(locals.AwsTags),
	}

	// --- Instance configuration ---
	instanceConfig := &apprunner.ServiceInstanceConfigurationArgs{}
	hasInstanceConfig := false

	if spec.GetCpu() != "" {
		instanceConfig.Cpu = pulumi.String(spec.GetCpu())
		hasInstanceConfig = true
	}
	if spec.GetMemory() != "" {
		instanceConfig.Memory = pulumi.String(spec.GetMemory())
		hasInstanceConfig = true
	}
	if spec.GetInstanceRoleArn().GetValue() != "" {
		instanceConfig.InstanceRoleArn = pulumi.String(spec.GetInstanceRoleArn().GetValue())
		hasInstanceConfig = true
	}
	if hasInstanceConfig {
		args.InstanceConfiguration = instanceConfig
	}

	// --- Health check configuration ---
	if hc := spec.GetHealthCheck(); hc != nil {
		healthCheckArgs := &apprunner.ServiceHealthCheckConfigurationArgs{}

		if hc.GetProtocol() != "" {
			healthCheckArgs.Protocol = pulumi.String(hc.GetProtocol())
		}
		if hc.GetPath() != "" {
			healthCheckArgs.Path = pulumi.String(hc.GetPath())
		}
		if hc.GetIntervalSeconds() > 0 {
			healthCheckArgs.Interval = pulumi.Int(int(hc.GetIntervalSeconds()))
		}
		if hc.GetTimeoutSeconds() > 0 {
			healthCheckArgs.Timeout = pulumi.Int(int(hc.GetTimeoutSeconds()))
		}
		if hc.GetHealthyThreshold() > 0 {
			healthCheckArgs.HealthyThreshold = pulumi.Int(int(hc.GetHealthyThreshold()))
		}
		if hc.GetUnhealthyThreshold() > 0 {
			healthCheckArgs.UnhealthyThreshold = pulumi.Int(int(hc.GetUnhealthyThreshold()))
		}

		args.HealthCheckConfiguration = healthCheckArgs
	}

	// --- Network configuration ---
	// Attach a VPC Connector (inline-created or externally referenced) for egress,
	// and configure ingress access (public/private, IP address type).
	networkConfig := buildNetworkConfiguration(locals, createdVpcConnector)
	if networkConfig != nil {
		args.NetworkConfiguration = networkConfig
	}

	// --- Auto Scaling Configuration ---
	if createdAutoScaling != nil {
		args.AutoScalingConfigurationArn = createdAutoScaling.Arn
	}

	// --- Encryption configuration ---
	if spec.GetKmsKeyArn().GetValue() != "" {
		args.EncryptionConfiguration = &apprunner.ServiceEncryptionConfigurationArgs{
			KmsKey: pulumi.String(spec.GetKmsKeyArn().GetValue()),
		}
	}

	// --- Observability configuration ---
	if spec.GetObservabilityEnabled() {
		obsArgs := &apprunner.ServiceObservabilityConfigurationArgs{
			ObservabilityEnabled: pulumi.Bool(true),
		}
		if spec.GetObservabilityConfigurationArn().GetValue() != "" {
			obsArgs.ObservabilityConfigurationArn = pulumi.String(spec.GetObservabilityConfigurationArn().GetValue())
		}
		args.ObservabilityConfiguration = obsArgs
	}

	// --- Create the service ---
	svc, err := apprunner.NewService(ctx, resourceName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create App Runner service")
	}

	// --- Export stack outputs ---
	ctx.Export(OpServiceArn, svc.Arn)
	ctx.Export(OpServiceId, svc.ServiceId)
	ctx.Export(OpServiceUrl, svc.ServiceUrl)
	ctx.Export(OpServiceName, pulumi.String(resourceName))
	ctx.Export(OpServiceStatus, svc.Status)

	return nil
}

// buildSourceConfiguration constructs the ServiceSourceConfigurationArgs from the spec.
// Exactly one of image_source or code_source must be set (enforced by proto validation).
func buildSourceConfiguration(locals *Locals) *apprunner.ServiceSourceConfigurationArgs {
	spec := locals.AwsAppRunnerService.Spec
	sourceConfig := &apprunner.ServiceSourceConfigurationArgs{}

	// Auto-deployments enabled (defaults to true via proto).
	sourceConfig.AutoDeploymentsEnabled = pulumi.Bool(spec.GetAutoDeploymentsEnabled())

	// --- Image source ---
	if img := spec.GetImageSource(); img != nil && img.GetImageIdentifier() != "" {
		imageRepoArgs := &apprunner.ServiceSourceConfigurationImageRepositoryArgs{
			ImageIdentifier:     pulumi.String(img.GetImageIdentifier()),
			ImageRepositoryType: pulumi.String(img.GetImageRepositoryType()),
		}

		// Build ImageConfiguration for port, start_command, env vars, and env secrets.
		imageConfigArgs := buildImageConfiguration(spec)
		if imageConfigArgs != nil {
			imageRepoArgs.ImageConfiguration = imageConfigArgs
		}

		sourceConfig.ImageRepository = imageRepoArgs

		// Authentication: access_role_arn is required for private ECR registries.
		if img.GetAccessRoleArn().GetValue() != "" {
			sourceConfig.AuthenticationConfiguration = &apprunner.ServiceSourceConfigurationAuthenticationConfigurationArgs{
				AccessRoleArn: pulumi.String(img.GetAccessRoleArn().GetValue()),
			}
		}
	}

	// --- Code source ---
	if code := spec.GetCodeSource(); code != nil && code.GetRepositoryUrl() != "" {
		codeRepoArgs := &apprunner.ServiceSourceConfigurationCodeRepositoryArgs{
			RepositoryUrl: pulumi.String(code.GetRepositoryUrl()),
			SourceCodeVersion: &apprunner.ServiceSourceConfigurationCodeRepositorySourceCodeVersionArgs{
				Type:  pulumi.String("BRANCH"),
				Value: pulumi.String(code.GetBranch()),
			},
		}

		// Code configuration: runtime, build command, port, start command, env vars.
		codeConfigArgs := &apprunner.ServiceSourceConfigurationCodeRepositoryCodeConfigurationArgs{
			ConfigurationSource: pulumi.String(code.GetConfigurationSource()),
		}

		// When configuration_source is "API", provide code configuration values.
		if code.GetConfigurationSource() == "API" {
			codeValues := &apprunner.ServiceSourceConfigurationCodeRepositoryCodeConfigurationCodeConfigurationValuesArgs{
				Runtime: pulumi.String(code.GetRuntime()),
			}
			if code.GetBuildCommand() != "" {
				codeValues.BuildCommand = pulumi.String(code.GetBuildCommand())
			}
			if spec.GetPort() != "" {
				codeValues.Port = pulumi.String(spec.GetPort())
			}
			if spec.GetStartCommand() != "" {
				codeValues.StartCommand = pulumi.String(spec.GetStartCommand())
			}
			if len(spec.GetEnvironmentVariables()) > 0 {
				codeValues.RuntimeEnvironmentVariables = pulumi.ToStringMap(spec.GetEnvironmentVariables())
			}
			if len(spec.GetEnvironmentSecrets()) > 0 {
				codeValues.RuntimeEnvironmentSecrets = pulumi.ToStringMap(spec.GetEnvironmentSecrets())
			}
			codeConfigArgs.CodeConfigurationValues = codeValues
		}

		codeRepoArgs.CodeConfiguration = codeConfigArgs

		// Set source_directory if provided.
		if code.GetSourceDirectory() != "" {
			codeRepoArgs.SourceDirectory = pulumi.String(code.GetSourceDirectory())
		}

		sourceConfig.CodeRepository = codeRepoArgs

		// Authentication: connection_arn is required for GitHub code source access.
		if code.GetConnectionArn().GetValue() != "" {
			sourceConfig.AuthenticationConfiguration = &apprunner.ServiceSourceConfigurationAuthenticationConfigurationArgs{
				ConnectionArn: pulumi.String(code.GetConnectionArn().GetValue()),
			}
		}
	}

	return sourceConfig
}

// buildImageConfiguration constructs the ImageConfigurationArgs for image-based
// deployments. It maps port, start_command, environment variables, and secrets.
func buildImageConfiguration(spec interface {
	GetPort() string
	GetStartCommand() string
	GetEnvironmentVariables() map[string]string
	GetEnvironmentSecrets() map[string]string
}) *apprunner.ServiceSourceConfigurationImageRepositoryImageConfigurationArgs {
	args := &apprunner.ServiceSourceConfigurationImageRepositoryImageConfigurationArgs{}
	hasConfig := false

	if spec.GetPort() != "" {
		args.Port = pulumi.String(spec.GetPort())
		hasConfig = true
	}
	if spec.GetStartCommand() != "" {
		args.StartCommand = pulumi.String(spec.GetStartCommand())
		hasConfig = true
	}
	if len(spec.GetEnvironmentVariables()) > 0 {
		args.RuntimeEnvironmentVariables = pulumi.ToStringMap(spec.GetEnvironmentVariables())
		hasConfig = true
	}
	if len(spec.GetEnvironmentSecrets()) > 0 {
		args.RuntimeEnvironmentSecrets = pulumi.ToStringMap(spec.GetEnvironmentSecrets())
		hasConfig = true
	}

	if !hasConfig {
		return nil
	}
	return args
}

// buildNetworkConfiguration constructs the ServiceNetworkConfigurationArgs.
// It wires up egress (VPC Connector) and ingress (public accessibility, IP type).
func buildNetworkConfiguration(
	locals *Locals,
	createdVpcConnector *apprunner.VpcConnector,
) *apprunner.ServiceNetworkConfigurationArgs {
	spec := locals.AwsAppRunnerService.Spec
	hasNetworkConfig := false

	netConfig := &apprunner.ServiceNetworkConfigurationArgs{}

	// --- Egress: VPC Connector ---
	// Prefer inline-created connector; fall back to externally referenced ARN.
	if createdVpcConnector != nil {
		netConfig.EgressConfiguration = &apprunner.ServiceNetworkConfigurationEgressConfigurationArgs{
			EgressType:      pulumi.String("VPC"),
			VpcConnectorArn: createdVpcConnector.Arn,
		}
		hasNetworkConfig = true
	} else if spec.GetVpcConnectorArn().GetValue() != "" {
		netConfig.EgressConfiguration = &apprunner.ServiceNetworkConfigurationEgressConfigurationArgs{
			EgressType:      pulumi.String("VPC"),
			VpcConnectorArn: pulumi.String(spec.GetVpcConnectorArn().GetValue()),
		}
		hasNetworkConfig = true
	}

	// --- Ingress: public accessibility and IP address type ---
	ingressConfig := &apprunner.ServiceNetworkConfigurationIngressConfigurationArgs{}
	hasIngress := false

	// is_publicly_accessible: the proto optional field returns false when nil,
	// but we want to set it explicitly when the user provides it.
	if spec.IsPubliclyAccessible != nil {
		ingressConfig.IsPubliclyAccessible = pulumi.Bool(spec.GetIsPubliclyAccessible())
		hasIngress = true
	}

	if hasIngress {
		netConfig.IngressConfiguration = ingressConfig
		hasNetworkConfig = true
	}

	if spec.GetIpAddressType() != "" {
		netConfig.IpAddressType = pulumi.String(spec.GetIpAddressType())
		hasNetworkConfig = true
	}

	if !hasNetworkConfig {
		return nil
	}
	return netConfig
}

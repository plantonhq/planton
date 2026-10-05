package module

import (
	"github.com/pkg/errors"
	azurecontainerappenvironmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerAppEnvironment.Spec

	// The environment is the secure boundary every app, job, storage
	// registration, and Dapr component in it shares. Name, region,
	// resource group, subnet, ILB, and zone redundancy are all ForceNew --
	// recreating the environment takes every workload in it down.
	envArgs := &containerapp.EnvironmentArgs{
		Name:              pulumi.String(spec.EnvironmentName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Logging destination. An explicit choice is honored as-is; when the
	// destination is unspecified but a workspace is referenced, the
	// modules deploy log-analytics (the destination the workspace
	// implies -- azurerm's own legacy inference). Without either, the
	// property is omitted and logs are streaming-only.
	if spec.LogAnalyticsWorkspaceId != nil {
		envArgs.LogAnalyticsWorkspaceId = pulumi.String(spec.LogAnalyticsWorkspaceId.GetValue())
	}
	if spec.LogsDestination != azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentLogsDestination_azure_container_app_environment_logs_destination_unspecified {
		envArgs.LogsDestination = pulumi.String(logsDestinationStrings[spec.LogsDestination])
	} else if spec.LogAnalyticsWorkspaceId != nil {
		envArgs.LogsDestination = pulumi.String("log-analytics")
	}

	// The Dapr telemetry connection string is write-only in ARM (never
	// returned on read), which is why it is ForceNew.
	if spec.DaprApplicationInsightsConnectionString != "" {
		envArgs.DaprApplicationInsightsConnectionString = pulumi.String(spec.DaprApplicationInsightsConnectionString)
	}

	// VNet injection: the subnet must be /21 or larger; ILB and zone
	// redundancy only exist for VNet-injected environments (the spec's
	// CELs mirror that pairing).
	if spec.InfrastructureSubnetId != nil {
		envArgs.InfrastructureSubnetId = pulumi.String(spec.InfrastructureSubnetId.GetValue())
	}

	// Only valid alongside workload profiles (spec-enforced); when
	// omitted Azure generates the platform resource-group name itself.
	if spec.InfrastructureResourceGroupName != "" {
		envArgs.InfrastructureResourceGroupName = pulumi.String(spec.InfrastructureResourceGroupName)
	}

	// ILB and zone redundancy only exist for VNet-injected environments:
	// the provider requires the subnet whenever either is SPECIFIED (even
	// as false), so both are sent only alongside the subnet.
	if spec.InfrastructureSubnetId != nil {
		if spec.InternalLoadBalancerEnabled != nil {
			envArgs.InternalLoadBalancerEnabled = pulumi.Bool(spec.GetInternalLoadBalancerEnabled())
		}
		if spec.ZoneRedundancyEnabled != nil {
			envArgs.ZoneRedundancyEnabled = pulumi.Bool(spec.GetZoneRedundancyEnabled())
		}
	}

	// Sent only when chosen: unset lets Azure derive the value from the
	// network configuration (Enabled externally, Disabled behind an ILB).
	if spec.PublicNetworkAccess != azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentPublicNetworkAccess_azure_container_app_environment_public_network_access_unspecified {
		envArgs.PublicNetworkAccess = pulumi.String(publicNetworkAccessStrings[spec.PublicNetworkAccess])
	}

	// mTLS encrypts and authenticates all app-to-app traffic in the
	// environment; costs latency and peak throughput under load.
	if spec.MutualTlsEnabled != nil {
		envArgs.MutualTlsEnabled = pulumi.Bool(spec.GetMutualTlsEnabled())
	}

	// The environment-level managed identity (used by platform
	// operations, e.g. Key Vault certificate reads). The spec's CEL
	// guarantees identity ids are present exactly when the type includes
	// UserAssigned.
	if spec.Identity != nil {
		identityArgs := &containerapp.EnvironmentIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.UserAssignedIdentityIds) > 0 {
			identityIds := make(pulumi.StringArray, 0, len(spec.Identity.UserAssignedIdentityIds))
			for _, identityId := range spec.Identity.UserAssignedIdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		envArgs.Identity = identityArgs
	}

	// Dedicated / GPU compute pools. Azure always includes the standard
	// Consumption profile itself, so only the declared profiles are sent.
	// The Consumption-family profiles are serverless: Azure rejects
	// instance counts on them, which the spec's CEL front-loads.
	if len(spec.WorkloadProfiles) > 0 {
		profiles := make(containerapp.EnvironmentWorkloadProfileArray, 0, len(spec.WorkloadProfiles))
		for _, workloadProfile := range spec.WorkloadProfiles {
			profileArgs := &containerapp.EnvironmentWorkloadProfileArgs{
				Name:                pulumi.String(workloadProfile.Name),
				WorkloadProfileType: pulumi.String(workloadProfileTypeStrings[workloadProfile.WorkloadProfileType]),
			}
			if workloadProfile.MinimumCount != nil {
				profileArgs.MinimumCount = pulumi.Int(int(workloadProfile.GetMinimumCount()))
			}
			if workloadProfile.MaximumCount != nil {
				profileArgs.MaximumCount = pulumi.Int(int(workloadProfile.GetMaximumCount()))
			}
			profiles = append(profiles, profileArgs)
		}
		envArgs.WorkloadProfiles = profiles
	}

	createdEnvironment, err := containerapp.NewEnvironment(ctx,
		spec.EnvironmentName,
		envArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Container App Environment %s", spec.EnvironmentName)
	}

	// The custom DNS suffix is a property of the environment itself (ARM
	// patches the environment's CustomDomainConfiguration; the resource's
	// id IS the environment id), so the spec folds it in and the module
	// realizes it through the association resource.
	if spec.CustomDomain != nil {
		_, err := containerapp.NewEnvironmentCustomDomain(ctx,
			"custom-domain",
			&containerapp.EnvironmentCustomDomainArgs{
				ContainerAppEnvironmentId: createdEnvironment.ID(),
				DnsSuffix:                 pulumi.String(spec.CustomDomain.DnsSuffix),
				CertificateBlobBase64:     pulumi.String(spec.CustomDomain.CertificateBlobBase64),
				CertificatePassword:       pulumi.String(spec.CustomDomain.CertificatePassword),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdEnvironment))
		if err != nil {
			return errors.Wrap(err, "failed to configure environment custom domain")
		}
	}

	// Export outputs. The platform-reserved values only populate
	// for VNet-injected environments; exported as-is so the output shape
	// stays constant across configurations.
	ctx.Export(OpEnvironmentId, createdEnvironment.ID())
	ctx.Export(OpEnvironmentName, createdEnvironment.Name)
	ctx.Export(OpDefaultDomain, createdEnvironment.DefaultDomain)
	ctx.Export(OpStaticIpAddress, createdEnvironment.StaticIpAddress)
	ctx.Export(OpPlatformReservedCidr, createdEnvironment.PlatformReservedCidr)
	ctx.Export(OpPlatformReservedDnsIpAddress, createdEnvironment.PlatformReservedDnsIpAddress)
	ctx.Export(OpDockerBridgeCidr, createdEnvironment.DockerBridgeCidr)
	ctx.Export(OpCustomDomainVerificationId, createdEnvironment.CustomDomainVerificationId)
	// The principal id exists only when the identity block carries a
	// system-assigned identity; exported empty otherwise so the output
	// shape stays constant across configurations.
	ctx.Export(OpIdentityPrincipalId, createdEnvironment.Identity.ApplyT(func(identity *containerapp.EnvironmentIdentity) string {
		if identity == nil || identity.PrincipalId == nil {
			return ""
		}
		return *identity.PrincipalId
	}).(pulumi.StringOutput))

	return nil
}

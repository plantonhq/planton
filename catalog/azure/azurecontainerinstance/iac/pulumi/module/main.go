package module

import (
	"github.com/pkg/errors"
	azurecontainerinstancev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// identityTypeStrings maps the spec enum's values to the provider's
// identity tokens. Like the Data Factory, a container group supports
// the combined mode -- the third token carries both flavors.
var identityTypeStrings = map[azurecontainerinstancev1alpha1.AzureContainerInstanceIdentityType]string{
	azurecontainerinstancev1alpha1.AzureContainerInstanceIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurecontainerinstancev1alpha1.AzureContainerInstanceIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurecontainerinstancev1alpha1.AzureContainerInstanceIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func Resources(ctx *pulumi.Context, iacInput *azurecontainerinstancev1alpha1.AzureContainerInstanceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerInstance.Spec

	// Create the container group. Almost the whole resource is
	// create-only: Azure applies only identity and tag changes in
	// place -- anything else replaces the group. Unset optional scalars
	// ride the provider defaults (ip_address_type "Public",
	// restart_policy "Always", sku "Standard",
	// dns_name_label_reuse_policy "Unsecure").
	args := &containerservice.GroupArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		OsType:            pulumi.String(spec.OsType),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.RestartPolicy != "" {
		args.RestartPolicy = pulumi.String(spec.RestartPolicy)
	}
	if spec.Sku != "" {
		args.Sku = pulumi.String(spec.Sku)
	}
	if spec.Priority != "" {
		args.Priority = pulumi.String(spec.Priority)
	}
	if spec.IpAddressType != "" {
		args.IpAddressType = pulumi.String(spec.IpAddressType)
	}
	if spec.DnsNameLabel != "" {
		args.DnsNameLabel = pulumi.String(spec.DnsNameLabel)
	}
	if spec.DnsNameLabelReusePolicy != "" {
		args.DnsNameLabelReusePolicy = pulumi.String(spec.DnsNameLabelReusePolicy)
	}
	// ENGINE SHAPE: the classic SDK flattens the provider's
	// one-element subnet_ids set to a scalar. Azure serializes
	// container-group operations per subnet.
	if spec.SubnetId.GetValue() != "" {
		args.SubnetIds = pulumi.String(spec.SubnetId.GetValue())
	}
	if len(spec.Zones) > 0 {
		args.Zones = pulumi.ToStringArray(spec.Zones)
	}

	// Customer-managed-key encryption: a VERSIONED Key Vault key
	// identifier (ACI pins the exact version -- rotation does not
	// follow). BEHAVIOR: the provider applies
	// key_vault_user_assigned_identity_id at CREATE only; a later
	// change to it alone is silently never applied (the provider's
	// update path covers only identity and tags) -- treat it as
	// create-only.
	if spec.KeyVaultKeyId.GetValue() != "" {
		args.KeyVaultKeyId = pulumi.String(spec.KeyVaultKeyId.GetValue())
	}
	if spec.KeyVaultUserAssignedIdentityId.GetValue() != "" {
		args.KeyVaultUserAssignedIdentityId = pulumi.String(spec.KeyVaultUserAssignedIdentityId.GetValue())
	}

	// Group-level exposed ports. Omit to expose every container port
	// (the provider derives the group ports). Protocol normalizes to
	// the provider default "TCP" so the rendered set is deterministic.
	if len(spec.ExposedPorts) > 0 {
		exposedPorts := containerservice.GroupExposedPortArray{}
		for _, exposed := range spec.ExposedPorts {
			exposedPorts = append(exposedPorts, &containerservice.GroupExposedPortArgs{
				Port:     pulumi.Int(int(exposed.Port)),
				Protocol: pulumi.String(normalizedProtocol(exposed.Protocol)),
			})
		}
		args.ExposedPorts = exposedPorts
	}

	if spec.Identity != nil {
		identityArgs := &containerservice.GroupIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		args.Identity = identityArgs
	}

	if len(spec.ImageRegistryCredentials) > 0 {
		credentials := containerservice.GroupImageRegistryCredentialArray{}
		for _, credential := range spec.ImageRegistryCredentials {
			credentialArgs := &containerservice.GroupImageRegistryCredentialArgs{
				Server: pulumi.String(credential.Server.GetValue()),
			}
			if credential.Username != "" {
				credentialArgs.Username = pulumi.String(credential.Username)
			}
			// Azure never returns the password on reads; the provider
			// echoes it from state.
			if credential.Password != "" {
				credentialArgs.Password = pulumi.String(credential.Password)
			}
			if credential.UserAssignedIdentityId.GetValue() != "" {
				credentialArgs.UserAssignedIdentityId = pulumi.String(credential.UserAssignedIdentityId.GetValue())
			}
			credentials = append(credentials, credentialArgs)
		}
		args.ImageRegistryCredentials = credentials
	}

	if len(spec.InitContainers) > 0 {
		initContainers := containerservice.GroupInitContainerArray{}
		for _, initContainer := range spec.InitContainers {
			initArgs := &containerservice.GroupInitContainerArgs{
				Name:  pulumi.String(initContainer.Name),
				Image: pulumi.String(initContainer.Image),
			}
			if len(initContainer.EnvironmentVariables) > 0 {
				initArgs.EnvironmentVariables = pulumi.ToStringMap(initContainer.EnvironmentVariables)
			}
			// Azure never returns secure values on reads; the provider
			// echoes them from state.
			if len(initContainer.SecureEnvironmentVariables) > 0 {
				initArgs.SecureEnvironmentVariables = pulumi.ToStringMap(initContainer.SecureEnvironmentVariables)
			}
			if len(initContainer.Commands) > 0 {
				initArgs.Commands = pulumi.ToStringArray(initContainer.Commands)
			}
			if len(initContainer.Volumes) > 0 {
				volumes := containerservice.GroupInitContainerVolumeArray{}
				for _, volume := range initContainer.Volumes {
					volumes = append(volumes, initContainerVolumeArgs(volume))
				}
				initArgs.Volumes = volumes
			}
			if initContainer.Security != nil {
				// ENGINE SHAPE: the classic SDK pluralizes the provider's
				// security block to Securities.
				initArgs.Securities = containerservice.GroupInitContainerSecurityArray{
					&containerservice.GroupInitContainerSecurityArgs{
						PrivilegeEnabled: pulumi.Bool(initContainer.Security.PrivilegeEnabled),
					},
				}
			}
			initContainers = append(initContainers, initArgs)
		}
		args.InitContainers = initContainers
	}

	containers := containerservice.GroupContainerArray{}
	for _, container := range spec.Containers {
		containerArgs := &containerservice.GroupContainerArgs{
			Name:   pulumi.String(container.Name),
			Image:  pulumi.String(container.Image),
			Cpu:    pulumi.Float64(container.Cpu),
			Memory: pulumi.Float64(container.Memory),
		}
		// BEHAVIOR: the provider applies the limits at CREATE only -- a
		// later change to either alone is silently never applied (the
		// provider's update path covers only identity and tags).
		if container.CpuLimit != nil {
			containerArgs.CpuLimit = pulumi.Float64(container.GetCpuLimit())
		}
		if container.MemoryLimit != nil {
			containerArgs.MemoryLimit = pulumi.Float64(container.GetMemoryLimit())
		}
		if len(container.Ports) > 0 {
			ports := containerservice.GroupContainerPortArray{}
			for _, port := range container.Ports {
				portArgs := &containerservice.GroupContainerPortArgs{
					Port: pulumi.Int(int(port.Port)),
				}
				if port.Protocol != "" {
					portArgs.Protocol = pulumi.String(port.Protocol)
				}
				ports = append(ports, portArgs)
			}
			containerArgs.Ports = ports
		}
		if len(container.EnvironmentVariables) > 0 {
			containerArgs.EnvironmentVariables = pulumi.ToStringMap(container.EnvironmentVariables)
		}
		// Azure never returns secure values on reads; the provider
		// echoes them from state.
		if len(container.SecureEnvironmentVariables) > 0 {
			containerArgs.SecureEnvironmentVariables = pulumi.ToStringMap(container.SecureEnvironmentVariables)
		}
		// Unset lets the image's own entrypoint run -- reads echo it
		// back (expected, not drift).
		if len(container.Commands) > 0 {
			containerArgs.Commands = pulumi.ToStringArray(container.Commands)
		}
		if len(container.Volumes) > 0 {
			volumes := containerservice.GroupContainerVolumeArray{}
			for _, volume := range container.Volumes {
				volumes = append(volumes, containerVolumeArgs(volume))
			}
			containerArgs.Volumes = volumes
		}
		if container.Security != nil {
			// ENGINE SHAPE: the classic SDK pluralizes the provider's
			// security block to Securities.
			containerArgs.Securities = containerservice.GroupContainerSecurityArray{
				&containerservice.GroupContainerSecurityArgs{
					PrivilegeEnabled: pulumi.Bool(container.Security.PrivilegeEnabled),
				},
			}
		}
		if probe := container.LivenessProbe; probe != nil {
			probeArgs := &containerservice.GroupContainerLivenessProbeArgs{}
			if len(probe.Exec) > 0 {
				probeArgs.Execs = pulumi.ToStringArray(probe.Exec)
			}
			// The spec's singular http_get as the SDK's one-element list
			// (the provider keeps only one on the wire).
			if probe.HttpGet != nil {
				httpGetArgs := &containerservice.GroupContainerLivenessProbeHttpGetArgs{}
				if probe.HttpGet.Path != "" {
					httpGetArgs.Path = pulumi.String(probe.HttpGet.Path)
				}
				if probe.HttpGet.Port != 0 {
					httpGetArgs.Port = pulumi.Int(int(probe.HttpGet.Port))
				}
				// Explicit-send "http" when unset: ARM materializes the
				// scheme on reads and the provider treats it as
				// replace-forcing, so an omitted scheme re-plans as a
				// destroy+create (live-proven by the idempotency gate).
				scheme := "http"
				if probe.HttpGet.Scheme != "" {
					scheme = probe.HttpGet.Scheme
				}
				httpGetArgs.Scheme = pulumi.String(scheme)
				if len(probe.HttpGet.HttpHeaders) > 0 {
					httpGetArgs.HttpHeaders = pulumi.ToStringMap(probe.HttpGet.HttpHeaders)
				}
				probeArgs.HttpGets = containerservice.GroupContainerLivenessProbeHttpGetArray{httpGetArgs}
			}
			// Zero means unset -- the provider sends timings only when > 0.
			if probe.InitialDelaySeconds != 0 {
				probeArgs.InitialDelaySeconds = pulumi.Int(int(probe.InitialDelaySeconds))
			}
			if probe.PeriodSeconds != 0 {
				probeArgs.PeriodSeconds = pulumi.Int(int(probe.PeriodSeconds))
			}
			if probe.FailureThreshold != 0 {
				probeArgs.FailureThreshold = pulumi.Int(int(probe.FailureThreshold))
			}
			if probe.SuccessThreshold != 0 {
				probeArgs.SuccessThreshold = pulumi.Int(int(probe.SuccessThreshold))
			}
			if probe.TimeoutSeconds != 0 {
				probeArgs.TimeoutSeconds = pulumi.Int(int(probe.TimeoutSeconds))
			}
			containerArgs.LivenessProbe = probeArgs
		}
		if probe := container.ReadinessProbe; probe != nil {
			probeArgs := &containerservice.GroupContainerReadinessProbeArgs{}
			if len(probe.Exec) > 0 {
				probeArgs.Execs = pulumi.ToStringArray(probe.Exec)
			}
			if probe.HttpGet != nil {
				httpGetArgs := &containerservice.GroupContainerReadinessProbeHttpGetArgs{}
				if probe.HttpGet.Path != "" {
					httpGetArgs.Path = pulumi.String(probe.HttpGet.Path)
				}
				if probe.HttpGet.Port != 0 {
					httpGetArgs.Port = pulumi.Int(int(probe.HttpGet.Port))
				}
				// Explicit-send "http" when unset -- see the liveness
				// probe's scheme note (replace-forcing echo).
				scheme := "http"
				if probe.HttpGet.Scheme != "" {
					scheme = probe.HttpGet.Scheme
				}
				httpGetArgs.Scheme = pulumi.String(scheme)
				if len(probe.HttpGet.HttpHeaders) > 0 {
					httpGetArgs.HttpHeaders = pulumi.ToStringMap(probe.HttpGet.HttpHeaders)
				}
				probeArgs.HttpGets = containerservice.GroupContainerReadinessProbeHttpGetArray{httpGetArgs}
			}
			if probe.InitialDelaySeconds != 0 {
				probeArgs.InitialDelaySeconds = pulumi.Int(int(probe.InitialDelaySeconds))
			}
			if probe.PeriodSeconds != 0 {
				probeArgs.PeriodSeconds = pulumi.Int(int(probe.PeriodSeconds))
			}
			if probe.FailureThreshold != 0 {
				probeArgs.FailureThreshold = pulumi.Int(int(probe.FailureThreshold))
			}
			if probe.SuccessThreshold != 0 {
				probeArgs.SuccessThreshold = pulumi.Int(int(probe.SuccessThreshold))
			}
			if probe.TimeoutSeconds != 0 {
				probeArgs.TimeoutSeconds = pulumi.Int(int(probe.TimeoutSeconds))
			}
			containerArgs.ReadinessProbe = probeArgs
		}
		containers = append(containers, containerArgs)
	}
	args.Containers = containers

	// The provider's diagnostics block has exactly one member -- this
	// Log Analytics form (the spec collapses the wrapper level). The
	// workspace_id is the CUSTOMER ID (GUID), not the ARM resource ID;
	// Azure never returns the key on reads (the provider echoes it from
	// configuration).
	if diagnostics := spec.DiagnosticsLogAnalytics; diagnostics != nil {
		logAnalyticsArgs := &containerservice.GroupDiagnosticsLogAnalyticsArgs{
			WorkspaceId:  pulumi.String(diagnostics.WorkspaceId.GetValue()),
			WorkspaceKey: pulumi.String(diagnostics.WorkspaceKey.GetValue()),
		}
		if diagnostics.LogType != "" {
			logAnalyticsArgs.LogType = pulumi.String(diagnostics.LogType)
		}
		// The provider only sends metadata alongside a log type
		// (validated upstream). Note the provider ALSO attaches an
		// empty metadata object whenever a log type is set, which ARM
		// rejects for ContainerInstanceLogs
		// (LogAnalyticsMetadataNotAllowed, live-proven) -- a spec CEL
		// blocks that log type until the provider stops sending the
		// empty map.
		if len(diagnostics.Metadata) > 0 {
			logAnalyticsArgs.Metadata = pulumi.ToStringMap(diagnostics.Metadata)
		}
		args.Diagnostics = &containerservice.GroupDiagnosticsArgs{
			LogAnalytics: logAnalyticsArgs,
		}
	}

	if dnsConfig := spec.DnsConfig; dnsConfig != nil {
		dnsConfigArgs := &containerservice.GroupDnsConfigArgs{
			Nameservers: pulumi.ToStringArray(dnsConfig.Nameservers),
		}
		if len(dnsConfig.SearchDomains) > 0 {
			dnsConfigArgs.SearchDomains = pulumi.ToStringArray(dnsConfig.SearchDomains)
		}
		if len(dnsConfig.Options) > 0 {
			dnsConfigArgs.Options = pulumi.ToStringArray(dnsConfig.Options)
		}
		args.DnsConfig = dnsConfigArgs
	}

	createdGroup, err := containerservice.NewGroup(ctx,
		locals.AzureContainerInstance.Metadata.Name,
		args,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create container group %s",
			locals.AzureContainerInstance.Metadata.Name)
	}

	ctx.Export(OpContainerGroupId, createdGroup.ID())
	ctx.Export(OpContainerGroupName, createdGroup.Name)
	ctx.Export(OpIpAddress, createdGroup.IpAddress)
	ctx.Export(OpFqdn, createdGroup.Fqdn)
	// Empty unless SYSTEM_ASSIGNED is enabled -- mirrors the TF module's
	// try(identity[0].principal_id, "").
	ctx.Export(OpIdentityPrincipalId, createdGroup.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))
	ctx.Export(OpIdentityTenantId, createdGroup.Identity.TenantId().ApplyT(func(tenantId *string) string {
		if tenantId == nil {
			return ""
		}
		return *tenantId
	}).(pulumi.StringOutput))

	return nil
}

// normalizedProtocol returns the provider default "TCP" for an unset
// protocol so the rendered exposed-port set is deterministic.
func normalizedProtocol(protocol string) string {
	if protocol == "" {
		return "TCP"
	}
	return protocol
}

// containerVolumeArgs flattens the spec's volume UNION (azure_file XOR
// empty_dir XOR git_repo XOR secret, validated before apply) onto the
// provider's flat volume shape.
func containerVolumeArgs(volume *azurecontainerinstancev1alpha1.AzureContainerInstanceVolume) *containerservice.GroupContainerVolumeArgs {
	volumeArgs := &containerservice.GroupContainerVolumeArgs{
		Name:      pulumi.String(volume.Name),
		MountPath: pulumi.String(volume.MountPath),
	}
	if volume.ReadOnly {
		volumeArgs.ReadOnly = pulumi.Bool(true)
	}
	if azureFile := volume.AzureFile; azureFile != nil {
		volumeArgs.ShareName = pulumi.String(azureFile.ShareName.GetValue())
		volumeArgs.StorageAccountName = pulumi.String(azureFile.StorageAccountName.GetValue())
		// Azure never returns the storage key on reads; the provider
		// echoes it from configuration.
		volumeArgs.StorageAccountKey = pulumi.String(azureFile.StorageAccountKey.GetValue())
	}
	if volume.EmptyDir {
		volumeArgs.EmptyDir = pulumi.Bool(true)
	}
	if gitRepo := volume.GitRepo; gitRepo != nil {
		gitRepoArgs := &containerservice.GroupContainerVolumeGitRepoArgs{
			Url: pulumi.String(gitRepo.Url),
		}
		if gitRepo.Directory != "" {
			gitRepoArgs.Directory = pulumi.String(gitRepo.Directory)
		}
		if gitRepo.Revision != "" {
			gitRepoArgs.Revision = pulumi.String(gitRepo.Revision)
		}
		volumeArgs.GitRepo = gitRepoArgs
	}
	if len(volume.Secret) > 0 {
		volumeArgs.Secret = pulumi.ToStringMap(volume.Secret)
	}
	return volumeArgs
}

// initContainerVolumeArgs is containerVolumeArgs for the init-container
// volume type (the SDK generates a distinct type per parent).
func initContainerVolumeArgs(volume *azurecontainerinstancev1alpha1.AzureContainerInstanceVolume) *containerservice.GroupInitContainerVolumeArgs {
	volumeArgs := &containerservice.GroupInitContainerVolumeArgs{
		Name:      pulumi.String(volume.Name),
		MountPath: pulumi.String(volume.MountPath),
	}
	if volume.ReadOnly {
		volumeArgs.ReadOnly = pulumi.Bool(true)
	}
	if azureFile := volume.AzureFile; azureFile != nil {
		volumeArgs.ShareName = pulumi.String(azureFile.ShareName.GetValue())
		volumeArgs.StorageAccountName = pulumi.String(azureFile.StorageAccountName.GetValue())
		volumeArgs.StorageAccountKey = pulumi.String(azureFile.StorageAccountKey.GetValue())
	}
	if volume.EmptyDir {
		volumeArgs.EmptyDir = pulumi.Bool(true)
	}
	if gitRepo := volume.GitRepo; gitRepo != nil {
		gitRepoArgs := &containerservice.GroupInitContainerVolumeGitRepoArgs{
			Url: pulumi.String(gitRepo.Url),
		}
		if gitRepo.Directory != "" {
			gitRepoArgs.Directory = pulumi.String(gitRepo.Directory)
		}
		if gitRepo.Revision != "" {
			gitRepoArgs.Revision = pulumi.String(gitRepo.Revision)
		}
		volumeArgs.GitRepo = gitRepoArgs
	}
	if len(volume.Secret) > 0 {
		volumeArgs.Secret = pulumi.ToStringMap(volume.Secret)
	}
	return volumeArgs
}

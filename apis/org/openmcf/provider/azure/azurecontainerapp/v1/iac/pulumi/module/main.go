package module

import (
	"github.com/pkg/errors"
	azurecontainerappv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurecontainerapp/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurecontainerappv1.AzureContainerAppStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx, "azure", &azure.ProviderArgs{
		ClientId:       pulumi.String(azureProviderConfig.ClientId),
		ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
		SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
		TenantId:       pulumi.String(azureProviderConfig.TenantId),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerApp.Spec

	// Build template
	templateArgs := &containerapp.AppTemplateArgs{
		Containers:  buildContainers(spec.Containers),
		MinReplicas: pulumi.IntPtr(int(spec.GetMinReplicas())),
		MaxReplicas: pulumi.IntPtr(int(spec.GetMaxReplicas())),
	}

	// Set init containers if provided
	if len(spec.InitContainers) > 0 {
		templateArgs.InitContainers = buildInitContainers(spec.InitContainers)
	}

	// Set volumes if provided
	if len(spec.Volumes) > 0 {
		templateArgs.Volumes = buildVolumes(spec.Volumes)
	}

	// Set revision suffix if provided
	if spec.RevisionSuffix != "" {
		templateArgs.RevisionSuffix = pulumi.StringPtr(spec.RevisionSuffix)
	}

	// Set HTTP scale rules
	if len(spec.HttpScaleRules) > 0 {
		templateArgs.HttpScaleRules = buildHttpScaleRules(spec.HttpScaleRules)
	}

	// Set TCP scale rules
	if len(spec.TcpScaleRules) > 0 {
		templateArgs.TcpScaleRules = buildTcpScaleRules(spec.TcpScaleRules)
	}

	// Set Azure Queue scale rules
	if len(spec.AzureQueueScaleRules) > 0 {
		templateArgs.AzureQueueScaleRules = buildAzureQueueScaleRules(spec.AzureQueueScaleRules)
	}

	// Set Custom scale rules
	if len(spec.CustomScaleRules) > 0 {
		templateArgs.CustomScaleRules = buildCustomScaleRules(spec.CustomScaleRules)
	}

	// Build Container App arguments
	revisionMode := spec.GetRevisionMode()
	if revisionMode == "" {
		revisionMode = "Single"
	}

	appArgs := &containerapp.AppArgs{
		Name:                      pulumi.String(spec.Name),
		ResourceGroupName:         pulumi.String(locals.ResourceGroupName),
		ContainerAppEnvironmentId: pulumi.String(spec.ContainerAppEnvironmentId.GetValue()),
		RevisionMode:              pulumi.String(revisionMode),
		Template:                  templateArgs,
		Tags:                      pulumi.ToStringMap(locals.AzureTags),
	}

	// Set workload profile name if provided
	if spec.WorkloadProfileName != "" {
		appArgs.WorkloadProfileName = pulumi.StringPtr(spec.WorkloadProfileName)
	}

	// Set max inactive revisions if provided
	if spec.MaxInactiveRevisions != nil {
		appArgs.MaxInactiveRevisions = pulumi.IntPtr(int(spec.GetMaxInactiveRevisions()))
	}

	// Set secrets if provided
	if len(spec.Secrets) > 0 {
		appArgs.Secrets = buildSecrets(spec.Secrets)
	}

	// Set registries if provided
	if len(spec.Registries) > 0 {
		appArgs.Registries = buildRegistries(spec.Registries)
	}

	// Set ingress if provided
	if spec.Ingress != nil {
		appArgs.Ingress = buildIngress(spec.Ingress)
	}

	// Set Dapr if provided
	if spec.Dapr != nil {
		appArgs.Dapr = buildDapr(spec.Dapr)
	}

	// Set identity if provided
	if spec.Identity != nil {
		appArgs.Identity = buildIdentity(spec.Identity)
	}

	// Create the Container App
	app, err := containerapp.NewApp(ctx, spec.Name, appArgs, pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Container App %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpContainerAppId, app.ID())
	ctx.Export(OpLatestRevisionName, app.LatestRevisionName)
	ctx.Export(OpLatestRevisionFqdn, app.LatestRevisionFqdn)
	ctx.Export(OpOutboundIpAddresses, app.OutboundIpAddresses)

	// Export ingress FQDN conditionally -- only present when ingress is configured
	ctx.Export(OpIngressFqdn, app.Ingress.ApplyT(func(ingress *containerapp.AppIngress) string {
		if ingress != nil && ingress.Fqdn != nil {
			return *ingress.Fqdn
		}
		return ""
	}).(pulumi.StringOutput))

	return nil
}

// ---------------------------------------------------------------------------
// Containers
// ---------------------------------------------------------------------------

func buildContainers(specs []*azurecontainerappv1.AzureContainerAppContainer) containerapp.AppTemplateContainerArray {
	containers := make(containerapp.AppTemplateContainerArray, 0, len(specs))
	for _, c := range specs {
		container := containerapp.AppTemplateContainerArgs{
			Name:   pulumi.String(c.Name),
			Image:  pulumi.String(c.Image),
			Cpu:    pulumi.Float64(c.Cpu),
			Memory: pulumi.String(c.Memory),
		}

		// Environment variables
		if len(c.Env) > 0 {
			container.Envs = buildEnvVars(c.Env)
		}

		// Command (entrypoint override)
		if len(c.Command) > 0 {
			container.Commands = pulumi.ToStringArray(c.Command)
		}

		// Args (CMD override)
		if len(c.Args) > 0 {
			container.Args = pulumi.ToStringArray(c.Args)
		}

		// Health probes
		if c.LivenessProbe != nil {
			container.LivenessProbes = containerapp.AppTemplateContainerLivenessProbeArray{
				buildLivenessProbe(c.LivenessProbe),
			}
		}
		if c.ReadinessProbe != nil {
			container.ReadinessProbes = containerapp.AppTemplateContainerReadinessProbeArray{
				buildReadinessProbe(c.ReadinessProbe),
			}
		}
		if c.StartupProbe != nil {
			container.StartupProbes = containerapp.AppTemplateContainerStartupProbeArray{
				buildStartupProbe(c.StartupProbe),
			}
		}

		// Volume mounts
		if len(c.VolumeMounts) > 0 {
			container.VolumeMounts = buildVolumeMounts(c.VolumeMounts)
		}

		containers = append(containers, container)
	}
	return containers
}

// ---------------------------------------------------------------------------
// Init Containers
// ---------------------------------------------------------------------------

func buildInitContainers(specs []*azurecontainerappv1.AzureContainerAppInitContainer) containerapp.AppTemplateInitContainerArray {
	initContainers := make(containerapp.AppTemplateInitContainerArray, 0, len(specs))
	for _, ic := range specs {
		initContainer := containerapp.AppTemplateInitContainerArgs{
			Name:  pulumi.String(ic.Name),
			Image: pulumi.String(ic.Image),
		}

		// CPU is optional for init containers
		if ic.Cpu != nil {
			initContainer.Cpu = pulumi.Float64(ic.GetCpu())
		}

		// Memory is optional for init containers
		if ic.Memory != nil {
			initContainer.Memory = pulumi.StringPtr(ic.GetMemory())
		}

		// Environment variables
		if len(ic.Env) > 0 {
			initContainer.Envs = buildInitContainerEnvVars(ic.Env)
		}

		// Command
		if len(ic.Command) > 0 {
			initContainer.Commands = pulumi.ToStringArray(ic.Command)
		}

		// Args
		if len(ic.Args) > 0 {
			initContainer.Args = pulumi.ToStringArray(ic.Args)
		}

		// Volume mounts
		if len(ic.VolumeMounts) > 0 {
			initContainer.VolumeMounts = buildInitContainerVolumeMounts(ic.VolumeMounts)
		}

		initContainers = append(initContainers, initContainer)
	}
	return initContainers
}

// ---------------------------------------------------------------------------
// Environment Variables
// ---------------------------------------------------------------------------

func buildEnvVars(specs []*azurecontainerappv1.AzureContainerAppEnvVar) containerapp.AppTemplateContainerEnvArray {
	envVars := make(containerapp.AppTemplateContainerEnvArray, 0, len(specs))
	for _, e := range specs {
		envVar := containerapp.AppTemplateContainerEnvArgs{
			Name: pulumi.String(e.Name),
		}

		// Secret-backed env var takes precedence over literal value
		if e.SecretName != "" {
			envVar.SecretName = pulumi.StringPtr(e.SecretName)
		} else if e.Value != "" {
			envVar.Value = pulumi.StringPtr(e.Value)
		}

		envVars = append(envVars, envVar)
	}
	return envVars
}

// ---------------------------------------------------------------------------
// Probes
// ---------------------------------------------------------------------------

func buildLivenessProbe(spec *azurecontainerappv1.AzureContainerAppProbe) containerapp.AppTemplateContainerLivenessProbeArgs {
	probe := containerapp.AppTemplateContainerLivenessProbeArgs{
		Transport: pulumi.String(spec.Transport),
		Port:      pulumi.Int(int(spec.Port)),
	}

	if spec.Path != "" {
		probe.Path = pulumi.StringPtr(spec.Path)
	}
	if spec.Host != "" {
		probe.Host = pulumi.StringPtr(spec.Host)
	}
	if len(spec.Headers) > 0 {
		probe.Headers = buildLivenessProbeHeaders(spec.Headers)
	}
	if spec.InitialDelayInSeconds != nil {
		probe.InitialDelay = pulumi.IntPtr(int(spec.GetInitialDelayInSeconds()))
	}
	if spec.IntervalSeconds != nil {
		probe.IntervalSeconds = pulumi.IntPtr(int(spec.GetIntervalSeconds()))
	}
	if spec.TimeoutSeconds != nil {
		probe.Timeout = pulumi.IntPtr(int(spec.GetTimeoutSeconds()))
	}
	if spec.FailureCountThreshold != nil {
		probe.FailureCountThreshold = pulumi.IntPtr(int(spec.GetFailureCountThreshold()))
	}

	return probe
}

func buildReadinessProbe(spec *azurecontainerappv1.AzureContainerAppProbe) containerapp.AppTemplateContainerReadinessProbeArgs {
	probe := containerapp.AppTemplateContainerReadinessProbeArgs{
		Transport: pulumi.String(spec.Transport),
		Port:      pulumi.Int(int(spec.Port)),
	}

	if spec.Path != "" {
		probe.Path = pulumi.StringPtr(spec.Path)
	}
	if spec.Host != "" {
		probe.Host = pulumi.StringPtr(spec.Host)
	}
	if len(spec.Headers) > 0 {
		probe.Headers = buildReadinessProbeHeaders(spec.Headers)
	}
	if spec.InitialDelayInSeconds != nil {
		probe.InitialDelay = pulumi.IntPtr(int(spec.GetInitialDelayInSeconds()))
	}
	if spec.IntervalSeconds != nil {
		probe.IntervalSeconds = pulumi.IntPtr(int(spec.GetIntervalSeconds()))
	}
	if spec.TimeoutSeconds != nil {
		probe.Timeout = pulumi.IntPtr(int(spec.GetTimeoutSeconds()))
	}
	if spec.FailureCountThreshold != nil {
		probe.FailureCountThreshold = pulumi.IntPtr(int(spec.GetFailureCountThreshold()))
	}
	if spec.SuccessCountThreshold != nil {
		probe.SuccessCountThreshold = pulumi.IntPtr(int(spec.GetSuccessCountThreshold()))
	}

	return probe
}

func buildStartupProbe(spec *azurecontainerappv1.AzureContainerAppProbe) containerapp.AppTemplateContainerStartupProbeArgs {
	probe := containerapp.AppTemplateContainerStartupProbeArgs{
		Transport: pulumi.String(spec.Transport),
		Port:      pulumi.Int(int(spec.Port)),
	}

	if spec.Path != "" {
		probe.Path = pulumi.StringPtr(spec.Path)
	}
	if spec.Host != "" {
		probe.Host = pulumi.StringPtr(spec.Host)
	}
	if len(spec.Headers) > 0 {
		probe.Headers = buildStartupProbeHeaders(spec.Headers)
	}
	if spec.InitialDelayInSeconds != nil {
		probe.InitialDelay = pulumi.IntPtr(int(spec.GetInitialDelayInSeconds()))
	}
	if spec.IntervalSeconds != nil {
		probe.IntervalSeconds = pulumi.IntPtr(int(spec.GetIntervalSeconds()))
	}
	if spec.TimeoutSeconds != nil {
		probe.Timeout = pulumi.IntPtr(int(spec.GetTimeoutSeconds()))
	}
	if spec.FailureCountThreshold != nil {
		probe.FailureCountThreshold = pulumi.IntPtr(int(spec.GetFailureCountThreshold()))
	}

	return probe
}

// Probe header builders for each probe type (each has its own Pulumi type)

func buildLivenessProbeHeaders(specs []*azurecontainerappv1.AzureContainerAppProbeHeader) containerapp.AppTemplateContainerLivenessProbeHeaderArray {
	headers := make(containerapp.AppTemplateContainerLivenessProbeHeaderArray, 0, len(specs))
	for _, h := range specs {
		headers = append(headers, containerapp.AppTemplateContainerLivenessProbeHeaderArgs{
			Name:  pulumi.String(h.Name),
			Value: pulumi.String(h.Value),
		})
	}
	return headers
}

func buildReadinessProbeHeaders(specs []*azurecontainerappv1.AzureContainerAppProbeHeader) containerapp.AppTemplateContainerReadinessProbeHeaderArray {
	headers := make(containerapp.AppTemplateContainerReadinessProbeHeaderArray, 0, len(specs))
	for _, h := range specs {
		headers = append(headers, containerapp.AppTemplateContainerReadinessProbeHeaderArgs{
			Name:  pulumi.String(h.Name),
			Value: pulumi.String(h.Value),
		})
	}
	return headers
}

func buildStartupProbeHeaders(specs []*azurecontainerappv1.AzureContainerAppProbeHeader) containerapp.AppTemplateContainerStartupProbeHeaderArray {
	headers := make(containerapp.AppTemplateContainerStartupProbeHeaderArray, 0, len(specs))
	for _, h := range specs {
		headers = append(headers, containerapp.AppTemplateContainerStartupProbeHeaderArgs{
			Name:  pulumi.String(h.Name),
			Value: pulumi.String(h.Value),
		})
	}
	return headers
}

// ---------------------------------------------------------------------------
// Init Container Environment Variables
// ---------------------------------------------------------------------------

func buildInitContainerEnvVars(specs []*azurecontainerappv1.AzureContainerAppEnvVar) containerapp.AppTemplateInitContainerEnvArray {
	envVars := make(containerapp.AppTemplateInitContainerEnvArray, 0, len(specs))
	for _, e := range specs {
		envVar := containerapp.AppTemplateInitContainerEnvArgs{
			Name: pulumi.String(e.Name),
		}

		if e.SecretName != "" {
			envVar.SecretName = pulumi.StringPtr(e.SecretName)
		} else if e.Value != "" {
			envVar.Value = pulumi.StringPtr(e.Value)
		}

		envVars = append(envVars, envVar)
	}
	return envVars
}

// ---------------------------------------------------------------------------
// Init Container Volume Mounts
// ---------------------------------------------------------------------------

func buildInitContainerVolumeMounts(specs []*azurecontainerappv1.AzureContainerAppVolumeMount) containerapp.AppTemplateInitContainerVolumeMountArray {
	mounts := make(containerapp.AppTemplateInitContainerVolumeMountArray, 0, len(specs))
	for _, vm := range specs {
		mount := containerapp.AppTemplateInitContainerVolumeMountArgs{
			Name: pulumi.String(vm.Name),
			Path: pulumi.String(vm.Path),
		}
		if vm.SubPath != "" {
			mount.SubPath = pulumi.StringPtr(vm.SubPath)
		}
		mounts = append(mounts, mount)
	}
	return mounts
}

// ---------------------------------------------------------------------------
// Volume Mounts
// ---------------------------------------------------------------------------

func buildVolumeMounts(specs []*azurecontainerappv1.AzureContainerAppVolumeMount) containerapp.AppTemplateContainerVolumeMountArray {
	mounts := make(containerapp.AppTemplateContainerVolumeMountArray, 0, len(specs))
	for _, vm := range specs {
		mount := containerapp.AppTemplateContainerVolumeMountArgs{
			Name: pulumi.String(vm.Name),
			Path: pulumi.String(vm.Path),
		}
		if vm.SubPath != "" {
			mount.SubPath = pulumi.StringPtr(vm.SubPath)
		}
		mounts = append(mounts, mount)
	}
	return mounts
}

// ---------------------------------------------------------------------------
// Volumes
// ---------------------------------------------------------------------------

func buildVolumes(specs []*azurecontainerappv1.AzureContainerAppVolume) containerapp.AppTemplateVolumeArray {
	volumes := make(containerapp.AppTemplateVolumeArray, 0, len(specs))
	for _, v := range specs {
		volume := containerapp.AppTemplateVolumeArgs{
			Name: pulumi.String(v.Name),
		}

		storageType := v.GetStorageType()
		if storageType == "" {
			storageType = "EmptyDir"
		}
		volume.StorageType = pulumi.StringPtr(storageType)

		if v.StorageName != "" {
			volume.StorageName = pulumi.StringPtr(v.StorageName)
		}

		volumes = append(volumes, volume)
	}
	return volumes
}

// ---------------------------------------------------------------------------
// Scale Rules
// ---------------------------------------------------------------------------

func buildHttpScaleRules(specs []*azurecontainerappv1.AzureContainerAppHttpScaleRule) containerapp.AppTemplateHttpScaleRuleArray {
	rules := make(containerapp.AppTemplateHttpScaleRuleArray, 0, len(specs))
	for _, r := range specs {
		rule := containerapp.AppTemplateHttpScaleRuleArgs{
			Name:               pulumi.String(r.Name),
			ConcurrentRequests: pulumi.String(r.ConcurrentRequests),
		}
		if len(r.Authentication) > 0 {
			rule.Authentications = buildHttpScaleRuleAuth(r.Authentication)
		}
		rules = append(rules, rule)
	}
	return rules
}

func buildTcpScaleRules(specs []*azurecontainerappv1.AzureContainerAppTcpScaleRule) containerapp.AppTemplateTcpScaleRuleArray {
	rules := make(containerapp.AppTemplateTcpScaleRuleArray, 0, len(specs))
	for _, r := range specs {
		rule := containerapp.AppTemplateTcpScaleRuleArgs{
			Name:               pulumi.String(r.Name),
			ConcurrentRequests: pulumi.String(r.ConcurrentRequests),
		}
		if len(r.Authentication) > 0 {
			rule.Authentications = buildTcpScaleRuleAuth(r.Authentication)
		}
		rules = append(rules, rule)
	}
	return rules
}

func buildAzureQueueScaleRules(specs []*azurecontainerappv1.AzureContainerAppAzureQueueScaleRule) containerapp.AppTemplateAzureQueueScaleRuleArray {
	rules := make(containerapp.AppTemplateAzureQueueScaleRuleArray, 0, len(specs))
	for _, r := range specs {
		rule := containerapp.AppTemplateAzureQueueScaleRuleArgs{
			Name:        pulumi.String(r.Name),
			QueueName:   pulumi.String(r.QueueName),
			QueueLength: pulumi.Int(int(r.QueueLength)),
		}
		if len(r.Authentication) > 0 {
			rule.Authentications = buildAzureQueueScaleRuleAuth(r.Authentication)
		}
		rules = append(rules, rule)
	}
	return rules
}

func buildCustomScaleRules(specs []*azurecontainerappv1.AzureContainerAppCustomScaleRule) containerapp.AppTemplateCustomScaleRuleArray {
	rules := make(containerapp.AppTemplateCustomScaleRuleArray, 0, len(specs))
	for _, r := range specs {
		rule := containerapp.AppTemplateCustomScaleRuleArgs{
			Name:           pulumi.String(r.Name),
			CustomRuleType: pulumi.String(r.CustomRuleType),
		}
		if len(r.Metadata) > 0 {
			rule.Metadata = pulumi.ToStringMap(r.Metadata)
		}
		if len(r.Authentication) > 0 {
			rule.Authentications = buildCustomScaleRuleAuth(r.Authentication)
		}
		rules = append(rules, rule)
	}
	return rules
}

// Scale rule authentication builders (each scale rule type has its own auth type)

func buildHttpScaleRuleAuth(specs []*azurecontainerappv1.AzureContainerAppScaleRuleAuth) containerapp.AppTemplateHttpScaleRuleAuthenticationArray {
	auths := make(containerapp.AppTemplateHttpScaleRuleAuthenticationArray, 0, len(specs))
	for _, a := range specs {
		auths = append(auths, containerapp.AppTemplateHttpScaleRuleAuthenticationArgs{
			SecretName:       pulumi.String(a.SecretName),
			TriggerParameter: pulumi.String(a.TriggerParameter),
		})
	}
	return auths
}

func buildTcpScaleRuleAuth(specs []*azurecontainerappv1.AzureContainerAppScaleRuleAuth) containerapp.AppTemplateTcpScaleRuleAuthenticationArray {
	auths := make(containerapp.AppTemplateTcpScaleRuleAuthenticationArray, 0, len(specs))
	for _, a := range specs {
		auths = append(auths, containerapp.AppTemplateTcpScaleRuleAuthenticationArgs{
			SecretName:       pulumi.String(a.SecretName),
			TriggerParameter: pulumi.String(a.TriggerParameter),
		})
	}
	return auths
}

func buildAzureQueueScaleRuleAuth(specs []*azurecontainerappv1.AzureContainerAppScaleRuleAuth) containerapp.AppTemplateAzureQueueScaleRuleAuthenticationArray {
	auths := make(containerapp.AppTemplateAzureQueueScaleRuleAuthenticationArray, 0, len(specs))
	for _, a := range specs {
		auths = append(auths, containerapp.AppTemplateAzureQueueScaleRuleAuthenticationArgs{
			SecretName:       pulumi.String(a.SecretName),
			TriggerParameter: pulumi.String(a.TriggerParameter),
		})
	}
	return auths
}

func buildCustomScaleRuleAuth(specs []*azurecontainerappv1.AzureContainerAppScaleRuleAuth) containerapp.AppTemplateCustomScaleRuleAuthenticationArray {
	auths := make(containerapp.AppTemplateCustomScaleRuleAuthenticationArray, 0, len(specs))
	for _, a := range specs {
		auths = append(auths, containerapp.AppTemplateCustomScaleRuleAuthenticationArgs{
			SecretName:       pulumi.String(a.SecretName),
			TriggerParameter: pulumi.String(a.TriggerParameter),
		})
	}
	return auths
}

// ---------------------------------------------------------------------------
// Secrets
// ---------------------------------------------------------------------------

func buildSecrets(specs []*azurecontainerappv1.AzureContainerAppSecret) containerapp.AppSecretArray {
	secrets := make(containerapp.AppSecretArray, 0, len(specs))
	for _, s := range specs {
		secret := containerapp.AppSecretArgs{
			Name: pulumi.String(s.Name),
		}

		// Plain-text value takes precedence when key_vault_secret_id is not set
		if s.KeyVaultSecretId != "" {
			secret.KeyVaultSecretId = pulumi.StringPtr(s.KeyVaultSecretId)
			if s.Identity != "" {
				secret.Identity = pulumi.StringPtr(s.Identity)
			}
		} else if s.Value != "" {
			secret.Value = pulumi.StringPtr(s.Value)
		}

		secrets = append(secrets, secret)
	}
	return secrets
}

// ---------------------------------------------------------------------------
// Registries
// ---------------------------------------------------------------------------

func buildRegistries(specs []*azurecontainerappv1.AzureContainerAppRegistry) containerapp.AppRegistryArray {
	registries := make(containerapp.AppRegistryArray, 0, len(specs))
	for _, r := range specs {
		registry := containerapp.AppRegistryArgs{
			Server: pulumi.String(r.Server),
		}

		// Username/password authentication
		if r.Username != "" {
			registry.Username = pulumi.StringPtr(r.Username)
		}
		if r.PasswordSecretName != "" {
			registry.PasswordSecretName = pulumi.StringPtr(r.PasswordSecretName)
		}

		// Managed identity authentication
		if r.Identity != "" {
			registry.Identity = pulumi.StringPtr(r.Identity)
		}

		registries = append(registries, registry)
	}
	return registries
}

// ---------------------------------------------------------------------------
// Ingress
// ---------------------------------------------------------------------------

func buildIngress(spec *azurecontainerappv1.AzureContainerAppIngress) *containerapp.AppIngressArgs {
	ingress := &containerapp.AppIngressArgs{
		TargetPort: pulumi.Int(int(spec.TargetPort)),
	}

	// External access (default: false)
	if spec.ExternalEnabled != nil {
		ingress.ExternalEnabled = pulumi.BoolPtr(spec.GetExternalEnabled())
	}

	// Exposed port (TCP transport only)
	if spec.ExposedPort != nil {
		ingress.ExposedPort = pulumi.IntPtr(int(spec.GetExposedPort()))
	}

	// Transport protocol
	if spec.Transport != nil {
		ingress.Transport = pulumi.StringPtr(spec.GetTransport())
	}

	// Allow insecure (HTTP) connections
	if spec.AllowInsecureConnections != nil {
		ingress.AllowInsecureConnections = pulumi.BoolPtr(spec.GetAllowInsecureConnections())
	}

	// Client certificate mode
	if spec.ClientCertificateMode != "" {
		ingress.ClientCertificateMode = pulumi.StringPtr(spec.ClientCertificateMode)
	}

	// Traffic weight distribution
	if len(spec.TrafficWeight) > 0 {
		ingress.TrafficWeights = buildTrafficWeights(spec.TrafficWeight)
	}

	// IP security restrictions
	if len(spec.IpSecurityRestrictions) > 0 {
		ingress.IpSecurityRestrictions = buildIpSecurityRestrictions(spec.IpSecurityRestrictions)
	}

	// CORS policy
	if spec.CorsPolicy != nil {
		ingress.CustomDomains = nil // placeholder -- custom_domains is separate
		// Build the CORS policy block if the provider supports it
		// The azurerm provider does not expose cors_policy on the ingress block as of v4.x
		// CORS configuration may need to be handled at a different layer
	}

	return ingress
}

func buildTrafficWeights(specs []*azurecontainerappv1.AzureContainerAppTrafficWeight) containerapp.AppIngressTrafficWeightArray {
	weights := make(containerapp.AppIngressTrafficWeightArray, 0, len(specs))
	for _, tw := range specs {
		weight := containerapp.AppIngressTrafficWeightArgs{
			Percentage: pulumi.Int(int(tw.Percentage)),
		}

		if tw.LatestRevision != nil {
			weight.LatestRevision = pulumi.BoolPtr(tw.GetLatestRevision())
		}
		if tw.RevisionSuffix != "" {
			weight.RevisionSuffix = pulumi.StringPtr(tw.RevisionSuffix)
		}
		if tw.Label != "" {
			weight.Label = pulumi.StringPtr(tw.Label)
		}

		weights = append(weights, weight)
	}
	return weights
}

func buildIpSecurityRestrictions(specs []*azurecontainerappv1.AzureContainerAppIpSecurityRestriction) containerapp.AppIngressIpSecurityRestrictionArray {
	restrictions := make(containerapp.AppIngressIpSecurityRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := containerapp.AppIngressIpSecurityRestrictionArgs{
			Name:           pulumi.String(r.Name),
			Action:         pulumi.String(r.Action),
			IpAddressRange: pulumi.String(r.IpAddressRange),
		}
		if r.Description != "" {
			restriction.Description = pulumi.StringPtr(r.Description)
		}
		restrictions = append(restrictions, restriction)
	}
	return restrictions
}

// ---------------------------------------------------------------------------
// Dapr
// ---------------------------------------------------------------------------

func buildDapr(spec *azurecontainerappv1.AzureContainerAppDapr) *containerapp.AppDaprArgs {
	dapr := &containerapp.AppDaprArgs{
		AppId: pulumi.String(spec.AppId),
	}

	if spec.AppPort != nil {
		dapr.AppPort = pulumi.IntPtr(int(spec.GetAppPort()))
	}

	appProtocol := spec.GetAppProtocol()
	if appProtocol == "" {
		appProtocol = "http"
	}
	dapr.AppProtocol = pulumi.StringPtr(appProtocol)

	return dapr
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func buildIdentity(spec *azurecontainerappv1.AzureContainerAppIdentity) *containerapp.AppIdentityArgs {
	identity := &containerapp.AppIdentityArgs{
		Type: pulumi.String(spec.Type),
	}

	// Resolve identity IDs from StringValueOrRef
	if len(spec.IdentityIds) > 0 {
		ids := make(pulumi.StringArray, 0, len(spec.IdentityIds))
		for _, ref := range spec.IdentityIds {
			ids = append(ids, pulumi.String(ref.GetValue()))
		}
		identity.IdentityIds = ids
	}

	return identity
}

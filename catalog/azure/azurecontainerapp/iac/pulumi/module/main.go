package module

import (
	"github.com/pkg/errors"
	azurecontainerappv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerapp/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurecontainerappv1alpha1.AzureContainerAppStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerApp.Spec

	// The template is the revision unit: every change to it creates a new
	// revision. Replica bounds and the scaler dials carry documented
	// defaults the platform does not materialize into the stack input, so
	// each is sent value-or-default (presence-guarded).
	templateArgs := &containerapp.AppTemplateArgs{
		Containers:  buildContainers(spec.Containers),
		MinReplicas: pulumi.IntPtr(intOrDefault(spec.MinReplicas, 0)),
		MaxReplicas: pulumi.IntPtr(intOrDefault(spec.MaxReplicas, 10)),
	}

	// The scaler dials: how long to wait after a scale event and how
	// often KEDA evaluates the rules.
	templateArgs.CooldownPeriodInSeconds = pulumi.IntPtr(intOrDefault(spec.CooldownPeriodInSeconds, 300))
	templateArgs.PollingIntervalInSeconds = pulumi.IntPtr(intOrDefault(spec.PollingIntervalInSeconds, 30))
	templateArgs.TerminationGracePeriodSeconds = pulumi.IntPtr(intOrDefault(spec.TerminationGracePeriodSeconds, 0))

	if len(spec.InitContainers) > 0 {
		templateArgs.InitContainers = buildInitContainers(spec.InitContainers)
	}

	if len(spec.Volumes) > 0 {
		templateArgs.Volumes = buildVolumes(spec.Volumes)
	}

	// Sent only when pinned: Azure generates a revision suffix itself
	// when the property is omitted.
	if spec.RevisionSuffix != "" {
		templateArgs.RevisionSuffix = pulumi.StringPtr(spec.RevisionSuffix)
	}

	if len(spec.HttpScaleRules) > 0 {
		templateArgs.HttpScaleRules = buildHttpScaleRules(spec.HttpScaleRules)
	}
	if len(spec.TcpScaleRules) > 0 {
		templateArgs.TcpScaleRules = buildTcpScaleRules(spec.TcpScaleRules)
	}
	if len(spec.AzureQueueScaleRules) > 0 {
		templateArgs.AzureQueueScaleRules = buildAzureQueueScaleRules(spec.AzureQueueScaleRules)
	}
	if len(spec.CustomScaleRules) > 0 {
		templateArgs.CustomScaleRules = buildCustomScaleRules(spec.CustomScaleRules)
	}

	// Unspecified revision mode deploys Single -- the right choice for
	// most workloads; Multiple exists for blue-green/canary splitting.
	revisionMode := "Single"
	if spec.RevisionMode != azurecontainerappv1alpha1.AzureContainerAppRevisionMode_azure_container_app_revision_mode_unspecified {
		revisionMode = revisionModeStrings[spec.RevisionMode]
	}

	appArgs := &containerapp.AppArgs{
		Name:                      pulumi.String(spec.ContainerAppName),
		ResourceGroupName:         pulumi.String(locals.ResourceGroupName),
		ContainerAppEnvironmentId: pulumi.String(spec.ContainerAppEnvironmentId.GetValue()),
		RevisionMode:              pulumi.String(revisionMode),
		Template:                  templateArgs,
		Tags:                      pulumi.ToStringMap(locals.AzureTags),
	}

	// Omitted runs on the environment's serverless Consumption profile.
	if spec.WorkloadProfileName != "" {
		appArgs.WorkloadProfileName = pulumi.StringPtr(spec.WorkloadProfileName)
	}

	if spec.MaxInactiveRevisions != nil {
		appArgs.MaxInactiveRevisions = pulumi.IntPtr(int(spec.GetMaxInactiveRevisions()))
	}

	if len(spec.Secrets) > 0 {
		appArgs.Secrets = buildSecrets(spec.Secrets)
	}

	if len(spec.Registries) > 0 {
		appArgs.Registries = buildRegistries(spec.Registries)
	}

	if spec.Ingress != nil {
		appArgs.Ingress = buildIngress(spec.Ingress)
	}

	if spec.Dapr != nil {
		appArgs.Dapr = buildDapr(spec.Dapr)
	}

	if spec.Identity != nil {
		appArgs.Identity = buildIdentity(spec.Identity)
	}

	createdApp, err := containerapp.NewApp(ctx,
		spec.ContainerAppName,
		appArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Container App %s", spec.ContainerAppName)
	}

	// Export stack outputs.
	ctx.Export(OpContainerAppId, createdApp.ID())
	ctx.Export(OpContainerAppName, createdApp.Name)
	ctx.Export(OpLatestRevisionName, createdApp.LatestRevisionName)
	ctx.Export(OpLatestRevisionFqdn, createdApp.LatestRevisionFqdn)
	ctx.Export(OpOutboundIpAddresses, createdApp.OutboundIpAddresses)
	ctx.Export(OpCustomDomainVerificationId, createdApp.CustomDomainVerificationId)

	// The app FQDN only exists when ingress is configured; exported empty
	// otherwise so the output shape stays constant across configurations.
	ctx.Export(OpIngressFqdn, createdApp.Ingress.ApplyT(func(ingress *containerapp.AppIngress) string {
		if ingress == nil || ingress.Fqdn == nil {
			return ""
		}
		return *ingress.Fqdn
	}).(pulumi.StringOutput))

	// The principal id exists only when the identity block carries a
	// system-assigned identity; exported empty otherwise.
	ctx.Export(OpIdentityPrincipalId, createdApp.Identity.ApplyT(func(identity *containerapp.AppIdentity) string {
		if identity == nil || identity.PrincipalId == nil {
			return ""
		}
		return *identity.PrincipalId
	}).(pulumi.StringOutput))

	return nil
}

// intOrDefault presence-guards an optional int32 field: stack inputs never
// materialize proto defaults, so an unset field must deploy the spec's
// documented default, not the Go zero value.
func intOrDefault(value *int32, defaultValue int) int {
	if value != nil {
		return int(*value)
	}
	return defaultValue
}

// ---------------------------------------------------------------------------
// Containers
// ---------------------------------------------------------------------------

func buildContainers(specs []*azurecontainerappv1alpha1.AzureContainerAppContainer) containerapp.AppTemplateContainerArray {
	containers := make(containerapp.AppTemplateContainerArray, 0, len(specs))
	for _, c := range specs {
		container := containerapp.AppTemplateContainerArgs{
			Name:   pulumi.String(c.Name),
			Image:  pulumi.String(c.Image),
			Cpu:    pulumi.Float64(c.Cpu),
			Memory: pulumi.String(c.Memory),
		}

		if len(c.Env) > 0 {
			container.Envs = buildEnvVars(c.Env)
		}

		if len(c.Command) > 0 {
			container.Commands = pulumi.ToStringArray(c.Command)
		}

		if len(c.Args) > 0 {
			container.Args = pulumi.ToStringArray(c.Args)
		}

		// Health probes. The per-type contracts (success threshold is
		// readiness-only, per-type failure ceilings, per-type initial
		// delay defaults) are front-loaded by the spec's CELs and the
		// per-type builders below.
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

func buildInitContainers(specs []*azurecontainerappv1alpha1.AzureContainerAppInitContainer) containerapp.AppTemplateInitContainerArray {
	initContainers := make(containerapp.AppTemplateInitContainerArray, 0, len(specs))
	for _, ic := range specs {
		initContainer := containerapp.AppTemplateInitContainerArgs{
			Name:  pulumi.String(ic.Name),
			Image: pulumi.String(ic.Image),
		}

		// CPU/memory are optional on init containers: omitted, they
		// inherit the app's overall allocation.
		if ic.Cpu != nil {
			initContainer.Cpu = pulumi.Float64(ic.GetCpu())
		}
		if ic.Memory != nil {
			initContainer.Memory = pulumi.StringPtr(ic.GetMemory())
		}

		if len(ic.Env) > 0 {
			initContainer.Envs = buildInitContainerEnvVars(ic.Env)
		}

		if len(ic.Command) > 0 {
			initContainer.Commands = pulumi.ToStringArray(ic.Command)
		}

		if len(ic.Args) > 0 {
			initContainer.Args = pulumi.ToStringArray(ic.Args)
		}

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

func buildEnvVars(specs []*azurecontainerappv1alpha1.AzureContainerAppEnvVar) containerapp.AppTemplateContainerEnvArray {
	envVars := make(containerapp.AppTemplateContainerEnvArray, 0, len(specs))
	for _, e := range specs {
		envVar := containerapp.AppTemplateContainerEnvArgs{
			Name: pulumi.String(e.Name),
		}

		// The spec's CEL guarantees value and secret_name never coexist.
		if e.SecretName != "" {
			envVar.SecretName = pulumi.StringPtr(e.SecretName)
		} else if e.Value != "" {
			envVar.Value = pulumi.StringPtr(e.Value)
		}

		envVars = append(envVars, envVar)
	}
	return envVars
}

func buildInitContainerEnvVars(specs []*azurecontainerappv1alpha1.AzureContainerAppEnvVar) containerapp.AppTemplateInitContainerEnvArray {
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
// Probes
// ---------------------------------------------------------------------------
// Each probe type gets Azure's own per-type default for initial_delay when
// the field is unset (1 for liveness, 0 for readiness/startup) -- matching
// what azurerm's schema defaults would send.

func buildLivenessProbe(spec *azurecontainerappv1alpha1.AzureContainerAppProbe) containerapp.AppTemplateContainerLivenessProbeArgs {
	probe := containerapp.AppTemplateContainerLivenessProbeArgs{
		Transport:             pulumi.String(probeTransportStrings[spec.Transport]),
		Port:                  pulumi.Int(int(spec.Port)),
		InitialDelay:          pulumi.IntPtr(intOrDefault(spec.InitialDelayInSeconds, 1)),
		IntervalSeconds:       pulumi.IntPtr(intOrDefault(spec.IntervalSeconds, 10)),
		Timeout:               pulumi.IntPtr(intOrDefault(spec.TimeoutSeconds, 1)),
		FailureCountThreshold: pulumi.IntPtr(intOrDefault(spec.FailureCountThreshold, 3)),
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

	return probe
}

func buildReadinessProbe(spec *azurecontainerappv1alpha1.AzureContainerAppProbe) containerapp.AppTemplateContainerReadinessProbeArgs {
	probe := containerapp.AppTemplateContainerReadinessProbeArgs{
		Transport:             pulumi.String(probeTransportStrings[spec.Transport]),
		Port:                  pulumi.Int(int(spec.Port)),
		InitialDelay:          pulumi.IntPtr(intOrDefault(spec.InitialDelayInSeconds, 0)),
		IntervalSeconds:       pulumi.IntPtr(intOrDefault(spec.IntervalSeconds, 10)),
		Timeout:               pulumi.IntPtr(intOrDefault(spec.TimeoutSeconds, 1)),
		FailureCountThreshold: pulumi.IntPtr(intOrDefault(spec.FailureCountThreshold, 3)),
		// Readiness is the only probe type Azure gives a success
		// threshold (the spec's CEL rejects it elsewhere).
		SuccessCountThreshold: pulumi.IntPtr(intOrDefault(spec.SuccessCountThreshold, 3)),
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

	return probe
}

func buildStartupProbe(spec *azurecontainerappv1alpha1.AzureContainerAppProbe) containerapp.AppTemplateContainerStartupProbeArgs {
	probe := containerapp.AppTemplateContainerStartupProbeArgs{
		Transport:             pulumi.String(probeTransportStrings[spec.Transport]),
		Port:                  pulumi.Int(int(spec.Port)),
		InitialDelay:          pulumi.IntPtr(intOrDefault(spec.InitialDelayInSeconds, 0)),
		IntervalSeconds:       pulumi.IntPtr(intOrDefault(spec.IntervalSeconds, 10)),
		Timeout:               pulumi.IntPtr(intOrDefault(spec.TimeoutSeconds, 1)),
		FailureCountThreshold: pulumi.IntPtr(intOrDefault(spec.FailureCountThreshold, 3)),
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

	return probe
}

// Probe header builders for each probe type (each has its own Pulumi type).

func buildLivenessProbeHeaders(specs []*azurecontainerappv1alpha1.AzureContainerAppProbeHeader) containerapp.AppTemplateContainerLivenessProbeHeaderArray {
	headers := make(containerapp.AppTemplateContainerLivenessProbeHeaderArray, 0, len(specs))
	for _, h := range specs {
		headers = append(headers, containerapp.AppTemplateContainerLivenessProbeHeaderArgs{
			Name:  pulumi.String(h.Name),
			Value: pulumi.String(h.Value),
		})
	}
	return headers
}

func buildReadinessProbeHeaders(specs []*azurecontainerappv1alpha1.AzureContainerAppProbeHeader) containerapp.AppTemplateContainerReadinessProbeHeaderArray {
	headers := make(containerapp.AppTemplateContainerReadinessProbeHeaderArray, 0, len(specs))
	for _, h := range specs {
		headers = append(headers, containerapp.AppTemplateContainerReadinessProbeHeaderArgs{
			Name:  pulumi.String(h.Name),
			Value: pulumi.String(h.Value),
		})
	}
	return headers
}

func buildStartupProbeHeaders(specs []*azurecontainerappv1alpha1.AzureContainerAppProbeHeader) containerapp.AppTemplateContainerStartupProbeHeaderArray {
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
// Volumes and Volume Mounts
// ---------------------------------------------------------------------------

func buildVolumeMounts(specs []*azurecontainerappv1alpha1.AzureContainerAppVolumeMount) containerapp.AppTemplateContainerVolumeMountArray {
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

func buildInitContainerVolumeMounts(specs []*azurecontainerappv1alpha1.AzureContainerAppVolumeMount) containerapp.AppTemplateInitContainerVolumeMountArray {
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

func buildVolumes(specs []*azurecontainerappv1alpha1.AzureContainerAppVolume) containerapp.AppTemplateVolumeArray {
	volumes := make(containerapp.AppTemplateVolumeArray, 0, len(specs))
	for _, v := range specs {
		volume := containerapp.AppTemplateVolumeArgs{
			Name: pulumi.String(v.Name),
		}

		// Unspecified deploys EmptyDir -- ephemeral scratch space.
		storageType := "EmptyDir"
		if v.StorageType != azurecontainerappv1alpha1.AzureContainerAppVolumeStorageType_azure_container_app_volume_storage_type_unspecified {
			storageType = volumeStorageTypeStrings[v.StorageType]
		}
		volume.StorageType = pulumi.StringPtr(storageType)

		// The environment storage registration backing file-share
		// volumes; the spec's CEL pairs it with the share-backed types.
		if v.StorageName != nil {
			volume.StorageName = pulumi.StringPtr(v.StorageName.GetValue())
		}

		if v.MountOptions != "" {
			volume.MountOptions = pulumi.StringPtr(v.MountOptions)
		}

		volumes = append(volumes, volume)
	}
	return volumes
}

// ---------------------------------------------------------------------------
// Scale Rules
// ---------------------------------------------------------------------------

func buildHttpScaleRules(specs []*azurecontainerappv1alpha1.AzureContainerAppHttpScaleRule) containerapp.AppTemplateHttpScaleRuleArray {
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

func buildTcpScaleRules(specs []*azurecontainerappv1alpha1.AzureContainerAppTcpScaleRule) containerapp.AppTemplateTcpScaleRuleArray {
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

func buildAzureQueueScaleRules(specs []*azurecontainerappv1alpha1.AzureContainerAppAzureQueueScaleRule) containerapp.AppTemplateAzureQueueScaleRuleArray {
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

func buildCustomScaleRules(specs []*azurecontainerappv1alpha1.AzureContainerAppCustomScaleRule) containerapp.AppTemplateCustomScaleRuleArray {
	rules := make(containerapp.AppTemplateCustomScaleRuleArray, 0, len(specs))
	for _, r := range specs {
		rule := containerapp.AppTemplateCustomScaleRuleArgs{
			Name:           pulumi.String(r.Name),
			CustomRuleType: pulumi.String(r.CustomRuleType),
			Metadata:       pulumi.ToStringMap(r.Metadata),
		}
		if len(r.Authentication) > 0 {
			rule.Authentications = buildCustomScaleRuleAuth(r.Authentication)
		}
		// Workload identity for the scaler instead of connection-string
		// secrets ("System" or a user-assigned identity ARM id; foreign-key
		// references arrive pre-resolved to the literal id).
		if r.IdentityId.GetValue() != "" {
			rule.IdentityId = pulumi.StringPtr(r.IdentityId.GetValue())
		}
		rules = append(rules, rule)
	}
	return rules
}

// Scale rule authentication builders (each scale rule type has its own
// Pulumi type). trigger_parameter is optional on HTTP/TCP rules (their
// scaler has a single implicit parameter) and spec-required elsewhere.

func buildHttpScaleRuleAuth(specs []*azurecontainerappv1alpha1.AzureContainerAppScaleRuleAuth) containerapp.AppTemplateHttpScaleRuleAuthenticationArray {
	auths := make(containerapp.AppTemplateHttpScaleRuleAuthenticationArray, 0, len(specs))
	for _, a := range specs {
		auth := containerapp.AppTemplateHttpScaleRuleAuthenticationArgs{
			SecretName: pulumi.String(a.SecretName),
		}
		if a.TriggerParameter != "" {
			auth.TriggerParameter = pulumi.StringPtr(a.TriggerParameter)
		}
		auths = append(auths, auth)
	}
	return auths
}

func buildTcpScaleRuleAuth(specs []*azurecontainerappv1alpha1.AzureContainerAppScaleRuleAuth) containerapp.AppTemplateTcpScaleRuleAuthenticationArray {
	auths := make(containerapp.AppTemplateTcpScaleRuleAuthenticationArray, 0, len(specs))
	for _, a := range specs {
		auth := containerapp.AppTemplateTcpScaleRuleAuthenticationArgs{
			SecretName: pulumi.String(a.SecretName),
		}
		if a.TriggerParameter != "" {
			auth.TriggerParameter = pulumi.StringPtr(a.TriggerParameter)
		}
		auths = append(auths, auth)
	}
	return auths
}

func buildAzureQueueScaleRuleAuth(specs []*azurecontainerappv1alpha1.AzureContainerAppScaleRuleAuth) containerapp.AppTemplateAzureQueueScaleRuleAuthenticationArray {
	auths := make(containerapp.AppTemplateAzureQueueScaleRuleAuthenticationArray, 0, len(specs))
	for _, a := range specs {
		auths = append(auths, containerapp.AppTemplateAzureQueueScaleRuleAuthenticationArgs{
			SecretName:       pulumi.String(a.SecretName),
			TriggerParameter: pulumi.String(a.TriggerParameter),
		})
	}
	return auths
}

func buildCustomScaleRuleAuth(specs []*azurecontainerappv1alpha1.AzureContainerAppScaleRuleAuth) containerapp.AppTemplateCustomScaleRuleAuthenticationArray {
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

func buildSecrets(specs []*azurecontainerappv1alpha1.AzureContainerAppSecret) containerapp.AppSecretArray {
	secrets := make(containerapp.AppSecretArray, 0, len(specs))
	for _, s := range specs {
		secret := containerapp.AppSecretArgs{
			Name: pulumi.String(s.Name),
		}

		// The spec's CELs guarantee value XOR key_vault_secret_id and
		// that identity travels with the Key Vault reference.
		if s.KeyVaultSecretId != "" {
			secret.KeyVaultSecretId = pulumi.StringPtr(s.KeyVaultSecretId)
			secret.Identity = pulumi.StringPtr(s.Identity)
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

func buildRegistries(specs []*azurecontainerappv1alpha1.AzureContainerAppRegistry) containerapp.AppRegistryArray {
	registries := make(containerapp.AppRegistryArray, 0, len(specs))
	for _, r := range specs {
		registry := containerapp.AppRegistryArgs{
			Server: pulumi.String(r.GetServer().GetValue()),
		}

		// Exactly one auth mode (spec-enforced): managed identity, or
		// username + password-secret together.
		if r.Identity != "" {
			registry.Identity = pulumi.StringPtr(r.Identity)
		} else {
			registry.Username = pulumi.StringPtr(r.Username)
			registry.PasswordSecretName = pulumi.StringPtr(r.PasswordSecretName)
		}

		registries = append(registries, registry)
	}
	return registries
}

// ---------------------------------------------------------------------------
// Ingress
// ---------------------------------------------------------------------------

func buildIngress(spec *azurecontainerappv1alpha1.AzureContainerAppIngress) *containerapp.AppIngressArgs {
	ingress := &containerapp.AppIngressArgs{
		TargetPort: pulumi.Int(int(spec.TargetPort)),
	}

	if spec.ExternalEnabled != nil {
		ingress.ExternalEnabled = pulumi.BoolPtr(spec.GetExternalEnabled())
	}

	// Only meaningful for TCP transport (spec-enforced).
	if spec.ExposedPort != nil {
		ingress.ExposedPort = pulumi.IntPtr(int(spec.GetExposedPort()))
	}

	// Unspecified deploys auto (Azure detects HTTP/1.1 vs HTTP/2).
	transport := "auto"
	if spec.Transport != azurecontainerappv1alpha1.AzureContainerAppIngressTransport_azure_container_app_ingress_transport_unspecified {
		transport = ingressTransportStrings[spec.Transport]
	}
	ingress.Transport = pulumi.StringPtr(transport)

	if spec.AllowInsecureConnections != nil {
		ingress.AllowInsecureConnections = pulumi.BoolPtr(spec.GetAllowInsecureConnections())
	}

	// Sent only when chosen: unset leaves Azure's default behavior (no
	// client certificate requirement).
	if spec.ClientCertificateMode != azurecontainerappv1alpha1.AzureContainerAppIngressClientCertificateMode_azure_container_app_ingress_client_certificate_mode_unspecified {
		ingress.ClientCertificateMode = pulumi.StringPtr(clientCertificateModeStrings[spec.ClientCertificateMode])
	}

	ingress.TrafficWeights = buildTrafficWeights(spec.TrafficWeight)

	if len(spec.IpSecurityRestrictions) > 0 {
		ingress.IpSecurityRestrictions = buildIpSecurityRestrictions(spec.IpSecurityRestrictions)
	}

	// CORS for browser-based clients.
	if spec.Cors != nil {
		corsArgs := &containerapp.AppIngressCorsArgs{
			AllowedOrigins: pulumi.ToStringArray(spec.Cors.AllowedOrigins),
		}
		if len(spec.Cors.AllowedHeaders) > 0 {
			corsArgs.AllowedHeaders = pulumi.ToStringArray(spec.Cors.AllowedHeaders)
		}
		if len(spec.Cors.AllowedMethods) > 0 {
			corsArgs.AllowedMethods = pulumi.ToStringArray(spec.Cors.AllowedMethods)
		}
		if len(spec.Cors.ExposedHeaders) > 0 {
			corsArgs.ExposedHeaders = pulumi.ToStringArray(spec.Cors.ExposedHeaders)
		}
		if spec.Cors.MaxAgeInSeconds != nil {
			corsArgs.MaxAgeInSeconds = pulumi.IntPtr(int(spec.Cors.GetMaxAgeInSeconds()))
		}
		if spec.Cors.AllowCredentialsEnabled != nil {
			corsArgs.AllowCredentialsEnabled = pulumi.BoolPtr(spec.Cors.GetAllowCredentialsEnabled())
		}
		ingress.Cors = corsArgs
	}

	return ingress
}

func buildTrafficWeights(specs []*azurecontainerappv1alpha1.AzureContainerAppTrafficWeight) containerapp.AppIngressTrafficWeightArray {
	weights := make(containerapp.AppIngressTrafficWeightArray, 0, len(specs))
	for _, tw := range specs {
		weight := containerapp.AppIngressTrafficWeightArgs{
			Percentage: pulumi.Int(int(tw.Percentage)),
		}

		// The spec's CEL guarantees exactly one target per weight.
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

func buildIpSecurityRestrictions(specs []*azurecontainerappv1alpha1.AzureContainerAppIpSecurityRestriction) containerapp.AppIngressIpSecurityRestrictionArray {
	restrictions := make(containerapp.AppIngressIpSecurityRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := containerapp.AppIngressIpSecurityRestrictionArgs{
			Name:           pulumi.String(r.Name),
			Action:         pulumi.String(ipRestrictionActionStrings[r.Action]),
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

func buildDapr(spec *azurecontainerappv1alpha1.AzureContainerAppDapr) *containerapp.AppDaprArgs {
	dapr := &containerapp.AppDaprArgs{
		AppId: pulumi.String(spec.AppId),
	}

	if spec.AppPort != nil {
		dapr.AppPort = pulumi.IntPtr(int(spec.GetAppPort()))
	}

	// Unspecified deploys http.
	appProtocol := "http"
	if spec.AppProtocol != azurecontainerappv1alpha1.AzureContainerAppDaprProtocol_azure_container_app_dapr_protocol_unspecified {
		appProtocol = daprProtocolStrings[spec.AppProtocol]
	}
	dapr.AppProtocol = pulumi.StringPtr(appProtocol)

	return dapr
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func buildIdentity(spec *azurecontainerappv1alpha1.AzureContainerAppIdentity) *containerapp.AppIdentityArgs {
	identity := &containerapp.AppIdentityArgs{
		Type: pulumi.String(identityTypeStrings[spec.Type]),
	}

	// The spec's CEL guarantees identity ids are present exactly when the
	// type includes UserAssigned.
	if len(spec.UserAssignedIdentityIds) > 0 {
		identityIds := make(pulumi.StringArray, 0, len(spec.UserAssignedIdentityIds))
		for _, identityId := range spec.UserAssignedIdentityIds {
			identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
		}
		identity.IdentityIds = identityIds
	}

	return identity
}

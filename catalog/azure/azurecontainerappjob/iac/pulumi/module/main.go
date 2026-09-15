package module

import (
	"github.com/pkg/errors"
	azurecontainerappjobv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappjob/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurecontainerappjobv1alpha1.AzureContainerAppJobStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerAppJob.Spec

	// The job: a run-to-completion workload inside the environment. Name,
	// region, resource group, environment, and the trigger choice are all
	// ForceNew.
	jobArgs := &containerapp.JobArgs{
		Name:                      pulumi.String(spec.JobName),
		Location:                  pulumi.String(spec.Region),
		ResourceGroupName:         pulumi.String(locals.ResourceGroupName),
		ContainerAppEnvironmentId: pulumi.String(spec.ContainerAppEnvironmentId.GetValue()),
		ReplicaTimeoutInSeconds:   pulumi.Int(int(spec.ReplicaTimeoutInSeconds)),
		Template:                  buildTemplate(spec),
		Tags:                      pulumi.ToStringMap(locals.AzureTags),
	}

	// 0 means no retries; omitted lets Azure apply its default (0).
	if spec.ReplicaRetryLimit != nil {
		jobArgs.ReplicaRetryLimit = pulumi.IntPtr(int(spec.GetReplicaRetryLimit()))
	}

	// Omitted runs on the environment's serverless Consumption profile.
	if spec.WorkloadProfileName != "" {
		jobArgs.WorkloadProfileName = pulumi.StringPtr(spec.WorkloadProfileName)
	}

	// Exactly one trigger (spec-enforced). Each carries parallelism and
	// the completion count -- documented defaults sent value-or-default
	// because the platform never materializes proto defaults.
	if spec.ManualTrigger != nil {
		jobArgs.ManualTriggerConfig = &containerapp.JobManualTriggerConfigArgs{
			Parallelism:            pulumi.IntPtr(intOrDefault(spec.ManualTrigger.Parallelism, 1)),
			ReplicaCompletionCount: pulumi.IntPtr(intOrDefault(spec.ManualTrigger.ReplicaCompletionCount, 1)),
		}
	}

	if spec.ScheduleTrigger != nil {
		jobArgs.ScheduleTriggerConfig = &containerapp.JobScheduleTriggerConfigArgs{
			CronExpression:         pulumi.String(spec.ScheduleTrigger.CronExpression),
			Parallelism:            pulumi.IntPtr(intOrDefault(spec.ScheduleTrigger.Parallelism, 1)),
			ReplicaCompletionCount: pulumi.IntPtr(intOrDefault(spec.ScheduleTrigger.ReplicaCompletionCount, 1)),
		}
	}

	if spec.EventTrigger != nil {
		eventTriggerArgs := &containerapp.JobEventTriggerConfigArgs{
			Parallelism:            pulumi.IntPtr(intOrDefault(spec.EventTrigger.Parallelism, 1)),
			ReplicaCompletionCount: pulumi.IntPtr(intOrDefault(spec.EventTrigger.ReplicaCompletionCount, 1)),
		}
		if spec.EventTrigger.Scale != nil {
			eventTriggerArgs.Scales = containerapp.JobEventTriggerConfigScaleArray{
				buildEventScale(spec.EventTrigger.Scale),
			}
		}
		jobArgs.EventTriggerConfig = eventTriggerArgs
	}

	if len(spec.Secrets) > 0 {
		jobArgs.Secrets = buildSecrets(spec.Secrets)
	}

	if len(spec.Registries) > 0 {
		jobArgs.Registries = buildRegistries(spec.Registries)
	}

	if spec.Identity != nil {
		identityArgs := &containerapp.JobIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.UserAssignedIdentityIds) > 0 {
			identityIds := make(pulumi.StringArray, 0, len(spec.Identity.UserAssignedIdentityIds))
			for _, identityId := range spec.Identity.UserAssignedIdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		jobArgs.Identity = identityArgs
	}

	createdJob, err := containerapp.NewJob(ctx,
		spec.JobName,
		jobArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Container App Job %s", spec.JobName)
	}

	// Export stack outputs.
	ctx.Export(OpJobId, createdJob.ID())
	ctx.Export(OpJobName, createdJob.Name)
	ctx.Export(OpEventStreamEndpoint, createdJob.EventStreamEndpoint)
	ctx.Export(OpOutboundIpAddresses, createdJob.OutboundIpAddresses)
	// The principal id exists only when the identity block carries a
	// system-assigned identity; exported empty otherwise.
	ctx.Export(OpIdentityPrincipalId, createdJob.Identity.ApplyT(func(identity *containerapp.JobIdentity) string {
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

// buildTemplate assembles the job's revision template: containers, init
// containers, and volumes (there is no scale block -- executions are the
// scaling unit).
func buildTemplate(spec *azurecontainerappjobv1alpha1.AzureContainerAppJobSpec) *containerapp.JobTemplateArgs {
	templateArgs := &containerapp.JobTemplateArgs{
		Containers: buildContainers(spec.Containers),
	}

	if len(spec.InitContainers) > 0 {
		templateArgs.InitContainers = buildInitContainers(spec.InitContainers)
	}

	if len(spec.Volumes) > 0 {
		templateArgs.Volumes = buildVolumes(spec.Volumes)
	}

	return templateArgs
}

// buildEventScale assembles the event trigger's scale contract.
func buildEventScale(scale *azurecontainerappjobv1alpha1.AzureContainerAppJobEventScale) *containerapp.JobEventTriggerConfigScaleArgs {
	scaleArgs := &containerapp.JobEventTriggerConfigScaleArgs{
		MaxExecutions:            pulumi.IntPtr(intOrDefault(scale.MaxExecutions, 100)),
		MinExecutions:            pulumi.IntPtr(intOrDefault(scale.MinExecutions, 0)),
		PollingIntervalInSeconds: pulumi.IntPtr(intOrDefault(scale.PollingIntervalInSeconds, 30)),
	}

	if len(scale.Rules) > 0 {
		rules := make(containerapp.JobEventTriggerConfigScaleRuleArray, 0, len(scale.Rules))
		for _, rule := range scale.Rules {
			ruleArgs := &containerapp.JobEventTriggerConfigScaleRuleArgs{
				Name:           pulumi.String(rule.Name),
				CustomRuleType: pulumi.String(rule.CustomRuleType),
				Metadata:       pulumi.ToStringMap(rule.Metadata),
			}
			if len(rule.Authentication) > 0 {
				auths := make(containerapp.JobEventTriggerConfigScaleRuleAuthenticationArray, 0, len(rule.Authentication))
				for _, auth := range rule.Authentication {
					auths = append(auths, &containerapp.JobEventTriggerConfigScaleRuleAuthenticationArgs{
						SecretName:       pulumi.String(auth.SecretName),
						TriggerParameter: pulumi.String(auth.TriggerParameter),
					})
				}
				ruleArgs.Authentications = auths
			}
			// Workload identity for the scaler instead of
			// connection-string secrets (foreign-key references arrive
			// pre-resolved to the literal id).
			if rule.IdentityId.GetValue() != "" {
				ruleArgs.IdentityId = pulumi.StringPtr(rule.IdentityId.GetValue())
			}
			rules = append(rules, ruleArgs)
		}
		scaleArgs.Rules = rules
	}

	return scaleArgs
}

// ---------------------------------------------------------------------------
// Containers
// ---------------------------------------------------------------------------

func buildContainers(specs []*azurecontainerappjobv1alpha1.AzureContainerAppJobContainer) containerapp.JobTemplateContainerArray {
	containers := make(containerapp.JobTemplateContainerArray, 0, len(specs))
	for _, c := range specs {
		container := containerapp.JobTemplateContainerArgs{
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
			container.LivenessProbes = containerapp.JobTemplateContainerLivenessProbeArray{
				buildLivenessProbe(c.LivenessProbe),
			}
		}
		if c.ReadinessProbe != nil {
			container.ReadinessProbes = containerapp.JobTemplateContainerReadinessProbeArray{
				buildReadinessProbe(c.ReadinessProbe),
			}
		}
		if c.StartupProbe != nil {
			container.StartupProbes = containerapp.JobTemplateContainerStartupProbeArray{
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

func buildInitContainers(specs []*azurecontainerappjobv1alpha1.AzureContainerAppJobInitContainer) containerapp.JobTemplateInitContainerArray {
	initContainers := make(containerapp.JobTemplateInitContainerArray, 0, len(specs))
	for _, ic := range specs {
		initContainer := containerapp.JobTemplateInitContainerArgs{
			Name:  pulumi.String(ic.Name),
			Image: pulumi.String(ic.Image),
		}

		// CPU/memory are optional on init containers: omitted, they
		// inherit the job's overall allocation.
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

func buildEnvVars(specs []*azurecontainerappjobv1alpha1.AzureContainerAppJobEnvVar) containerapp.JobTemplateContainerEnvArray {
	envVars := make(containerapp.JobTemplateContainerEnvArray, 0, len(specs))
	for _, e := range specs {
		envVar := containerapp.JobTemplateContainerEnvArgs{
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

func buildInitContainerEnvVars(specs []*azurecontainerappjobv1alpha1.AzureContainerAppJobEnvVar) containerapp.JobTemplateInitContainerEnvArray {
	envVars := make(containerapp.JobTemplateInitContainerEnvArray, 0, len(specs))
	for _, e := range specs {
		envVar := containerapp.JobTemplateInitContainerEnvArgs{
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

func buildLivenessProbe(spec *azurecontainerappjobv1alpha1.AzureContainerAppJobProbe) containerapp.JobTemplateContainerLivenessProbeArgs {
	probe := containerapp.JobTemplateContainerLivenessProbeArgs{
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
		headers := make(containerapp.JobTemplateContainerLivenessProbeHeaderArray, 0, len(spec.Headers))
		for _, h := range spec.Headers {
			headers = append(headers, containerapp.JobTemplateContainerLivenessProbeHeaderArgs{
				Name:  pulumi.String(h.Name),
				Value: pulumi.String(h.Value),
			})
		}
		probe.Headers = headers
	}

	return probe
}

func buildReadinessProbe(spec *azurecontainerappjobv1alpha1.AzureContainerAppJobProbe) containerapp.JobTemplateContainerReadinessProbeArgs {
	probe := containerapp.JobTemplateContainerReadinessProbeArgs{
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
		headers := make(containerapp.JobTemplateContainerReadinessProbeHeaderArray, 0, len(spec.Headers))
		for _, h := range spec.Headers {
			headers = append(headers, containerapp.JobTemplateContainerReadinessProbeHeaderArgs{
				Name:  pulumi.String(h.Name),
				Value: pulumi.String(h.Value),
			})
		}
		probe.Headers = headers
	}

	return probe
}

func buildStartupProbe(spec *azurecontainerappjobv1alpha1.AzureContainerAppJobProbe) containerapp.JobTemplateContainerStartupProbeArgs {
	probe := containerapp.JobTemplateContainerStartupProbeArgs{
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
		headers := make(containerapp.JobTemplateContainerStartupProbeHeaderArray, 0, len(spec.Headers))
		for _, h := range spec.Headers {
			headers = append(headers, containerapp.JobTemplateContainerStartupProbeHeaderArgs{
				Name:  pulumi.String(h.Name),
				Value: pulumi.String(h.Value),
			})
		}
		probe.Headers = headers
	}

	return probe
}

// ---------------------------------------------------------------------------
// Volumes and Volume Mounts
// ---------------------------------------------------------------------------

func buildVolumeMounts(specs []*azurecontainerappjobv1alpha1.AzureContainerAppJobVolumeMount) containerapp.JobTemplateContainerVolumeMountArray {
	mounts := make(containerapp.JobTemplateContainerVolumeMountArray, 0, len(specs))
	for _, vm := range specs {
		mount := containerapp.JobTemplateContainerVolumeMountArgs{
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

func buildInitContainerVolumeMounts(specs []*azurecontainerappjobv1alpha1.AzureContainerAppJobVolumeMount) containerapp.JobTemplateInitContainerVolumeMountArray {
	mounts := make(containerapp.JobTemplateInitContainerVolumeMountArray, 0, len(specs))
	for _, vm := range specs {
		mount := containerapp.JobTemplateInitContainerVolumeMountArgs{
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

func buildVolumes(specs []*azurecontainerappjobv1alpha1.AzureContainerAppJobVolume) containerapp.JobTemplateVolumeArray {
	volumes := make(containerapp.JobTemplateVolumeArray, 0, len(specs))
	for _, v := range specs {
		volume := containerapp.JobTemplateVolumeArgs{
			Name: pulumi.String(v.Name),
		}

		// Unspecified deploys EmptyDir -- ephemeral scratch space.
		storageType := "EmptyDir"
		if v.StorageType != azurecontainerappjobv1alpha1.AzureContainerAppJobVolumeStorageType_azure_container_app_job_volume_storage_type_unspecified {
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
// Secrets
// ---------------------------------------------------------------------------

func buildSecrets(specs []*azurecontainerappjobv1alpha1.AzureContainerAppJobSecret) containerapp.JobSecretArray {
	secrets := make(containerapp.JobSecretArray, 0, len(specs))
	for _, s := range specs {
		secret := containerapp.JobSecretArgs{
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

func buildRegistries(specs []*azurecontainerappjobv1alpha1.AzureContainerAppJobRegistry) containerapp.JobRegistryArray {
	registries := make(containerapp.JobRegistryArray, 0, len(specs))
	for _, r := range specs {
		registry := containerapp.JobRegistryArgs{
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

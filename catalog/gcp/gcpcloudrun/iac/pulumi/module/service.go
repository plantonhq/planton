package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpcloudrunv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrun/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// service provisions the Cloud Run v2 service — a stable serving endpoint
// plus a revision template. Every update that touches the template stamps
// out a new immutable revision; the traffic block controls how requests are
// split across revisions, which is what makes blue/green and canary rollouts
// declarative.
func service(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*cloudrunv2.Service, error) {
	spec := locals.GcpCloudRun.Spec

	// Enable the Cloud Run Admin API first so a fresh project works on the
	// first deploy. disable_on_destroy stays false: tearing down one service
	// must never disable the API for everything else in the project.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("run.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"cloudrun-run.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable run.googleapis.com api")
	}

	// Secret values the env carries are stored in Secret Manager before the
	// service exists, and the service reads each one by reference.
	storedSecrets, err := cloudrunenv.Store(ctx, secretPlacement(locals), secretVariables(spec), gcpProvider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store the environment's secret values")
	}

	// Deletion guard, honest by default: an unset spec field means true, so
	// a destroy fails until the manifest explicitly opts out.
	deletionProtection := true
	if spec.DeletionProtection != nil {
		deletionProtection = spec.GetDeletionProtection()
	}

	args := &cloudrunv2.ServiceArgs{
		Name:               pulumi.String(locals.ServiceName),
		Location:           pulumi.String(spec.Region),
		Template:           buildTemplate(spec, storedSecrets.Refs),
		Labels:             pulumi.ToStringMap(locals.GcpLabels),
		DeletionProtection: pulumi.Bool(deletionProtection),
	}

	// An empty project falls back to the provider's default project — the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}

	// The proto enum NAMES match the API values exactly; the UNSPECIFIED
	// zero value means "let GCP default" and must not reach the API.
	if spec.Ingress != gcpcloudrunv1alpha1.GcpCloudRunIngress_INGRESS_TRAFFIC_UNSPECIFIED {
		args.Ingress = pulumi.String(spec.Ingress.String())
	}
	if spec.LaunchStage != "" {
		args.LaunchStage = pulumi.String(spec.LaunchStage)
	}

	// Service-object annotations for external tools (system namespaces are
	// API-rejected). Revision-level annotations ride the template instead.
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringMap(spec.Annotations)
	}

	// Identity-Aware Proxy in front of the service — Google's managed
	// login wall. Omitted when false so the provider default applies.
	if spec.IapEnabled {
		args.IapEnabled = pulumi.Bool(true)
	}

	// Turns off the default *.run.app URL, leaving custom-domain / LB
	// paths as the only front doors.
	if spec.DefaultUriDisabled {
		args.DefaultUriDisabled = pulumi.Bool(true)
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// the service from management without deleting it in GCP.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	// Resource Manager tags, bound at creation only (ForceNew): omitted
	// when empty so a service without tags carries no tag surface.
	if len(spec.ResourceManagerTags) > 0 {
		args.Tags = pulumi.ToStringMap(spec.ResourceManagerTags)
	}

	// Deploy-from-source: Cloud Build produces the serving image (the
	// Cloud Run functions build path) instead of a prebuilt image.
	if spec.BuildConfig != nil {
		buildConfig := &cloudrunv2.ServiceBuildConfigArgs{}
		if spec.BuildConfig.SourceLocation != "" {
			buildConfig.SourceLocation = pulumi.String(spec.BuildConfig.SourceLocation)
		}
		if spec.BuildConfig.FunctionTarget != "" {
			buildConfig.FunctionTarget = pulumi.String(spec.BuildConfig.FunctionTarget)
		}
		if spec.BuildConfig.ImageUri != "" {
			buildConfig.ImageUri = pulumi.String(spec.BuildConfig.ImageUri)
		}
		if spec.BuildConfig.BaseImage != "" {
			buildConfig.BaseImage = pulumi.String(spec.BuildConfig.BaseImage)
		}
		if spec.BuildConfig.EnableAutomaticUpdates {
			buildConfig.EnableAutomaticUpdates = pulumi.Bool(true)
		}
		if len(spec.BuildConfig.EnvironmentVariables) > 0 {
			buildConfig.EnvironmentVariables = pulumi.ToStringMap(spec.BuildConfig.EnvironmentVariables)
		}
		if spec.BuildConfig.WorkerPool != "" {
			buildConfig.WorkerPool = pulumi.String(spec.BuildConfig.WorkerPool)
		}
		if spec.BuildConfig.ServiceAccount != "" {
			buildConfig.ServiceAccount = pulumi.String(spec.BuildConfig.ServiceAccount)
		}
		args.BuildConfig = buildConfig
	}

	// Multi-region service: one identity serving from several regions. The
	// proto guarantees region is "global" whenever this is present.
	if spec.MultiRegionSettings != nil {
		args.MultiRegionSettings = &cloudrunv2.ServiceMultiRegionSettingsArgs{
			Regions: pulumi.ToStringArray(spec.MultiRegionSettings.Regions),
		}
	}

	// Turns the IAM run.routes.invoke check off entirely — the org-policy
	// alternative to granting allUsers (the IAM member below). The proto
	// rejects setting both.
	if spec.InvokerIamDisabled {
		args.InvokerIamDisabled = pulumi.Bool(true)
	}

	if len(spec.CustomAudiences) > 0 {
		args.CustomAudiences = pulumi.ToStringArray(spec.CustomAudiences)
	}

	// Binary Authorization deploy gate: the project default policy XOR a
	// named platform policy (the proto rejects both).
	if spec.BinaryAuthorization != nil {
		binaryAuthorization := &cloudrunv2.ServiceBinaryAuthorizationArgs{}
		if spec.BinaryAuthorization.UseDefault {
			binaryAuthorization.UseDefault = pulumi.Bool(true)
		}
		if spec.BinaryAuthorization.Policy != "" {
			binaryAuthorization.Policy = pulumi.String(spec.BinaryAuthorization.Policy)
		}
		if spec.BinaryAuthorization.BreakglassJustification != "" {
			binaryAuthorization.BreakglassJustification = pulumi.String(spec.BinaryAuthorization.BreakglassJustification)
		}
		args.BinaryAuthorization = binaryAuthorization
	}

	// Service-wide scaling posture (distinct from the per-revision bounds
	// in the template): MANUAL pins the total instance count.
	if spec.ServiceScaling != nil {
		serviceScaling := &cloudrunv2.ServiceScalingArgs{}
		if spec.ServiceScaling.ScalingMode != "" {
			serviceScaling.ScalingMode = pulumi.String(spec.ServiceScaling.ScalingMode)
		}
		if spec.ServiceScaling.ManualInstanceCount != nil {
			serviceScaling.ManualInstanceCount = pulumi.Int(int(spec.ServiceScaling.GetManualInstanceCount()))
		}
		if spec.ServiceScaling.MinInstanceCount != nil {
			serviceScaling.MinInstanceCount = pulumi.Int(int(spec.ServiceScaling.GetMinInstanceCount()))
		}
		if spec.ServiceScaling.MaxInstanceCount != nil {
			serviceScaling.MaxInstanceCount = pulumi.Int(int(spec.ServiceScaling.GetMaxInstanceCount()))
		}
		args.Scaling = serviceScaling
	}

	// Traffic split across revisions. An empty spec list means "100% to the
	// latest ready revision" — achieved by omitting the field entirely so
	// the provider applies GCP's default without recording a diff-prone
	// split.
	if len(spec.Traffic) > 0 {
		traffics := cloudrunv2.ServiceTrafficArray{}
		for _, target := range spec.Traffic {
			traffic := &cloudrunv2.ServiceTrafficArgs{
				Type: pulumi.String(target.Type),
			}
			if target.Revision != "" {
				traffic.Revision = pulumi.String(target.Revision)
			}
			if target.Percent != nil {
				traffic.Percent = pulumi.Int(int(target.GetPercent()))
			}
			if target.Tag != "" {
				traffic.Tag = pulumi.String(target.Tag)
			}
			traffics = append(traffics, traffic)
		}
		args.Traffics = traffics
	}

	createdService, err := cloudrunv2.NewService(ctx,
		locals.GcpCloudRun.Metadata.Name,
		args,
		pulumi.Provider(gcpProvider),
		// Cloud Run checks the runtime identity's access to every secret a
		// revision reads when it creates the revision, so the grants land first.
		pulumi.DependsOn(append([]pulumi.Resource{createdProjectService}, storedSecrets.Grants...)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Cloud Run v2 service")
	}

	// Public access: grants roles/run.invoker to allUsers when the spec says
	// so. This is the additive-IAM path (access THROUGH the invoker check);
	// invoker_iam_disabled above is the org-policy alternative that turns
	// the check off instead. Destroying the grant restores
	// authenticated-only.
	if spec.AllowUnauthenticated {
		_, err := cloudrunv2.NewServiceIamMember(ctx,
			"public-invoker",
			&cloudrunv2.ServiceIamMemberArgs{
				Project:  createdService.Project,
				Location: createdService.Location,
				Name:     createdService.Name,
				Role:     pulumi.String("roles/run.invoker"),
				Member:   pulumi.String("allUsers"),
			},
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdService),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to grant public invoker access")
		}
	}

	return createdService, nil
}

// buildTemplate maps the spec's revision-level surface onto the v2 revision
// template: containers, volumes, scaling, networking, and hardware.
func buildTemplate(spec *gcpcloudrunv1alpha1.GcpCloudRunSpec, secretRefs map[cloudrunenv.Key]cloudrunenv.Ref) *cloudrunv2.ServiceTemplateArgs {
	template := &cloudrunv2.ServiceTemplateArgs{
		Containers: buildContainers(spec, secretRefs),
	}

	// Explicit revision naming makes declarative blue/green possible; unset
	// (the norm) lets Cloud Run generate names.
	if spec.Revision != "" {
		template.Revision = pulumi.String(spec.Revision)
	}

	// Revision-level metadata, stamped on every revision the template
	// creates (distinct from the service-object labels/annotations).
	if len(spec.RevisionLabels) > 0 {
		template.Labels = pulumi.ToStringMap(spec.RevisionLabels)
	}
	if len(spec.RevisionAnnotations) > 0 {
		template.Annotations = pulumi.ToStringMap(spec.RevisionAnnotations)
	}

	// Escape hatch disabling ALL probes (startup and liveness) for
	// workloads whose serving model breaks the probe contract.
	if spec.HealthCheckDisabled {
		template.HealthCheckDisabled = pulumi.Bool(true)
	}

	// The runtime identity whose permissions the code exercises. Unset uses
	// the project's Compute Engine default service account.
	if spec.ServiceAccount.GetValue() != "" {
		template.ServiceAccount = pulumi.String(spec.ServiceAccount.GetValue())
	}

	if spec.ExecutionEnvironment != gcpcloudrunv1alpha1.GcpCloudRunExecutionEnvironment_EXECUTION_ENVIRONMENT_UNSPECIFIED {
		template.ExecutionEnvironment = pulumi.String(spec.ExecutionEnvironment.String())
	}

	if spec.MaxInstanceRequestConcurrency != nil {
		template.MaxInstanceRequestConcurrency = pulumi.Int(int(spec.GetMaxInstanceRequestConcurrency()))
	}

	// The API takes the timeout as a duration string ("300s"); the spec
	// keeps the honest integer-seconds shape.
	if spec.TimeoutSeconds != nil {
		template.Timeout = pulumi.String(fmt.Sprintf("%ds", spec.GetTimeoutSeconds()))
	}

	if spec.SessionAffinity {
		template.SessionAffinity = pulumi.Bool(true)
	}

	if spec.EncryptionKey.GetValue() != "" {
		template.EncryptionKey = pulumi.String(spec.EncryptionKey.GetValue())
	}

	// Single-zone GPU serving opt-in (cheaper GPU capacity for zonal risk).
	if spec.GpuZonalRedundancyDisabled {
		template.GpuZonalRedundancyDisabled = pulumi.Bool(true)
	}

	// Per-revision instance bounds; an omitted block scales 0..default cap.
	if spec.Scaling != nil {
		scaling := &cloudrunv2.ServiceTemplateScalingArgs{}
		if spec.Scaling.MinInstanceCount != nil {
			scaling.MinInstanceCount = pulumi.Int(int(spec.Scaling.GetMinInstanceCount()))
		}
		if spec.Scaling.MaxInstanceCount != nil {
			scaling.MaxInstanceCount = pulumi.Int(int(spec.Scaling.GetMaxInstanceCount()))
		}
		template.Scaling = scaling
	}

	// GPU hardware requirement (e.g. "nvidia-l4").
	if spec.NodeSelector != nil {
		template.NodeSelector = &cloudrunv2.ServiceTemplateNodeSelectorArgs{
			Accelerator: pulumi.String(spec.NodeSelector.Accelerator),
		}
	}

	// Outbound VPC networking: a Serverless VPC Access connector XOR direct
	// VPC egress network_interfaces — the proto guarantees exactly one.
	if spec.VpcAccess != nil {
		vpcAccess := &cloudrunv2.ServiceTemplateVpcAccessArgs{}
		if spec.VpcAccess.Connector.GetValue() != "" {
			vpcAccess.Connector = pulumi.String(spec.VpcAccess.Connector.GetValue())
		}
		if spec.VpcAccess.Egress != "" {
			vpcAccess.Egress = pulumi.String(spec.VpcAccess.Egress)
		}
		if len(spec.VpcAccess.NetworkInterfaces) > 0 {
			interfaces := cloudrunv2.ServiceTemplateVpcAccessNetworkInterfaceArray{}
			for _, networkInterface := range spec.VpcAccess.NetworkInterfaces {
				interfaceArgs := &cloudrunv2.ServiceTemplateVpcAccessNetworkInterfaceArgs{}
				if networkInterface.Network.GetValue() != "" {
					interfaceArgs.Network = pulumi.String(networkInterface.Network.GetValue())
				}
				if networkInterface.Subnetwork.GetValue() != "" {
					interfaceArgs.Subnetwork = pulumi.String(networkInterface.Subnetwork.GetValue())
				}
				if len(networkInterface.Tags) > 0 {
					interfaceArgs.Tags = pulumi.ToStringArray(networkInterface.Tags)
				}
				interfaces = append(interfaces, interfaceArgs)
			}
			vpcAccess.NetworkInterfaces = interfaces
		}
		template.VpcAccess = vpcAccess
	}

	// Named volumes; each spec entry carries exactly one source arm
	// (proto-oneof-enforced).
	if len(spec.Volumes) > 0 {
		volumes := cloudrunv2.ServiceTemplateVolumeArray{}
		for _, volume := range spec.Volumes {
			volumeArgs := &cloudrunv2.ServiceTemplateVolumeArgs{
				Name: pulumi.String(volume.Name),
			}
			switch source := volume.Source.(type) {
			case *gcpcloudrunv1alpha1.GcpCloudRunVolume_CloudSqlInstance:
				// Cloud SQL Unix sockets — GCP manages the proxying;
				// connect via /cloudsql/<project:region:instance>.
				instances := pulumi.StringArray{}
				for _, instance := range source.CloudSqlInstance.Instances {
					instances = append(instances, pulumi.String(instance.GetValue()))
				}
				volumeArgs.CloudSqlInstance = &cloudrunv2.ServiceTemplateVolumeCloudSqlInstanceArgs{
					Instances: instances,
				}
			case *gcpcloudrunv1alpha1.GcpCloudRunVolume_Secret:
				secret := &cloudrunv2.ServiceTemplateVolumeSecretArgs{
					Secret: pulumi.String(source.Secret.Secret),
				}
				if source.Secret.DefaultMode != nil {
					secret.DefaultMode = pulumi.Int(int(source.Secret.GetDefaultMode()))
				}
				if len(source.Secret.Items) > 0 {
					items := cloudrunv2.ServiceTemplateVolumeSecretItemArray{}
					for _, item := range source.Secret.Items {
						itemArgs := &cloudrunv2.ServiceTemplateVolumeSecretItemArgs{
							Path: pulumi.String(item.Path),
						}
						if item.Version != "" {
							itemArgs.Version = pulumi.String(item.Version)
						}
						if item.Mode != nil {
							itemArgs.Mode = pulumi.Int(int(item.GetMode()))
						}
						items = append(items, itemArgs)
					}
					secret.Items = items
				}
				volumeArgs.Secret = secret
			case *gcpcloudrunv1alpha1.GcpCloudRunVolume_EmptyDir:
				emptyDir := &cloudrunv2.ServiceTemplateVolumeEmptyDirArgs{}
				if source.EmptyDir.Medium != "" {
					emptyDir.Medium = pulumi.String(source.EmptyDir.Medium)
				}
				if source.EmptyDir.SizeLimit != "" {
					emptyDir.SizeLimit = pulumi.String(source.EmptyDir.SizeLimit)
				}
				volumeArgs.EmptyDir = emptyDir
			case *gcpcloudrunv1alpha1.GcpCloudRunVolume_Gcs:
				// GCS FUSE mounts require the GEN2 execution environment.
				gcs := &cloudrunv2.ServiceTemplateVolumeGcsArgs{
					Bucket:   pulumi.String(source.Gcs.Bucket.GetValue()),
					ReadOnly: pulumi.Bool(source.Gcs.ReadOnly),
				}
				if len(source.Gcs.MountOptions) > 0 {
					gcs.MountOptions = pulumi.ToStringArray(source.Gcs.MountOptions)
				}
				volumeArgs.Gcs = gcs
			case *gcpcloudrunv1alpha1.GcpCloudRunVolume_Nfs:
				volumeArgs.Nfs = &cloudrunv2.ServiceTemplateVolumeNfsArgs{
					Server:   pulumi.String(source.Nfs.Server),
					Path:     pulumi.String(source.Nfs.Path),
					ReadOnly: pulumi.Bool(source.Nfs.ReadOnly),
				}
			}
			volumes = append(volumes, volumeArgs)
		}
		template.Volumes = volumes
	}

	// Sandbox templates the supervisor container launches on demand.
	// Emitted only when declared: the provider's wrapper block is a
	// single-item list around the template list.
	if len(spec.SandboxTemplates) > 0 {
		templates := cloudrunv2.ServiceTemplateSandboxesTemplateArray{}
		for _, sandbox := range spec.SandboxTemplates {
			sandboxArgs := &cloudrunv2.ServiceTemplateSandboxesTemplateArgs{
				Name:  pulumi.String(sandbox.Name),
				Image: pulumi.String(sandbox.Image),
			}
			if len(sandbox.Command) > 0 {
				sandboxArgs.Commands = pulumi.ToStringArray(sandbox.Command)
			}
			if len(sandbox.Args) > 0 {
				sandboxArgs.Args = pulumi.ToStringArray(sandbox.Args)
			}
			if sandbox.WorkingDir != "" {
				sandboxArgs.WorkingDir = pulumi.StringPtr(sandbox.WorkingDir)
			}
			if len(sandbox.Env) > 0 {
				envs := cloudrunv2.ServiceTemplateSandboxesTemplateEnvArray{}
				for _, env := range sandbox.Env {
					envs = append(envs, &cloudrunv2.ServiceTemplateSandboxesTemplateEnvArgs{
						Name:  pulumi.String(env.Name),
						Value: pulumi.StringPtr(env.Value),
					})
				}
				sandboxArgs.Envs = envs
			}
			if len(sandbox.VolumeMounts) > 0 {
				mounts := cloudrunv2.ServiceTemplateSandboxesTemplateVolumeMountArray{}
				for _, mount := range sandbox.VolumeMounts {
					mountArgs := &cloudrunv2.ServiceTemplateSandboxesTemplateVolumeMountArgs{
						Name:      pulumi.String(mount.Name),
						MountPath: pulumi.String(mount.MountPath),
					}
					if mount.SubPath != "" {
						mountArgs.SubPath = pulumi.StringPtr(mount.SubPath)
					}
					mounts = append(mounts, mountArgs)
				}
				sandboxArgs.VolumeMounts = mounts
			}
			templates = append(templates, sandboxArgs)
		}
		template.Sandboxes = &cloudrunv2.ServiceTemplateSandboxesArgs{Templates: templates}
	}

	return template
}

// buildContainers maps the spec's containers — the serving container plus
// any sidecars sharing localhost and volumes, ordered by depends_on.
func buildContainers(spec *gcpcloudrunv1alpha1.GcpCloudRunSpec, secretRefs map[cloudrunenv.Key]cloudrunenv.Ref) cloudrunv2.ServiceTemplateContainerArray {
	containers := cloudrunv2.ServiceTemplateContainerArray{}

	for containerIndex, container := range spec.Containers {
		containerArgs := &cloudrunv2.ServiceTemplateContainerArgs{
			Image: pulumi.String(container.Image),
		}

		if container.Name != "" {
			containerArgs.Name = pulumi.StringPtr(container.Name)
		}
		if len(container.Command) > 0 {
			containerArgs.Commands = pulumi.ToStringArray(container.Command)
		}
		if len(container.Args) > 0 {
			containerArgs.Args = pulumi.ToStringArray(container.Args)
		}
		if container.WorkingDir != "" {
			containerArgs.WorkingDir = pulumi.String(container.WorkingDir)
		}
		if len(container.DependsOn) > 0 {
			containerArgs.DependsOns = pulumi.ToStringArray(container.DependsOn)
		}

		// Sandbox supervisor flag; omitted when false so the provider
		// default applies cleanly.
		if container.SandboxLauncher {
			containerArgs.SandboxLauncher = pulumi.BoolPtr(true)
		}

		// Base image for automatic base-image updates on source deploys
		// (pairs with build_config.enable_automatic_updates).
		if container.BaseImageUri != "" {
			containerArgs.BaseImageUri = pulumi.String(container.BaseImageUri)
		}

		// Environment: a literal, a Secret Manager secret the author owns, or
		// a secret value this module stored (one of the three -- proto-enforced).
		if len(container.Env) > 0 {
			envs := cloudrunv2.ServiceTemplateContainerEnvArray{}
			for _, envVar := range container.Env {
				envArgs := &cloudrunv2.ServiceTemplateContainerEnvArgs{
					Name: pulumi.String(envVar.Name),
				}
				if ref, stored := secretRefs[cloudrunenv.Key{ContainerIndex: containerIndex, Name: envVar.Name}]; stored {
					envArgs.ValueSource = &cloudrunv2.ServiceTemplateContainerEnvValueSourceArgs{
						SecretKeyRef: &cloudrunv2.ServiceTemplateContainerEnvValueSourceSecretKeyRefArgs{
							Secret:  ref.Secret,
							Version: ref.Version,
						},
					}
				} else if envVar.ValueFromSecret != nil {
					secretKeyRef := &cloudrunv2.ServiceTemplateContainerEnvValueSourceSecretKeyRefArgs{
						Secret: pulumi.String(envVar.ValueFromSecret.Secret),
					}
					if envVar.ValueFromSecret.Version != "" {
						secretKeyRef.Version = pulumi.String(envVar.ValueFromSecret.Version)
					}
					envArgs.ValueSource = &cloudrunv2.ServiceTemplateContainerEnvValueSourceArgs{
						SecretKeyRef: secretKeyRef,
					}
				} else {
					envArgs.Value = pulumi.String(envVar.Value)
				}
				envs = append(envs, envArgs)
			}
			containerArgs.Envs = envs
		}

		// The single traffic-serving port; name "h2c" enables end-to-end
		// HTTP/2 (required for gRPC streaming).
		if container.Ports != nil {
			ports := &cloudrunv2.ServiceTemplateContainerPortsArgs{}
			if container.Ports.ContainerPort != nil {
				ports.ContainerPort = pulumi.Int(int(container.Ports.GetContainerPort()))
			}
			if container.Ports.Name != "" {
				ports.Name = pulumi.String(container.Ports.Name)
			}
			containerArgs.Ports = ports
		}

		// CPU/memory land in the API's limits map; the allocation levers
		// (cpu_idle, startup_cpu_boost) ride alongside.
		if container.Resources != nil {
			resources := &cloudrunv2.ServiceTemplateContainerResourcesArgs{}
			limits := pulumi.StringMap{}
			if container.Resources.Cpu != "" {
				limits["cpu"] = pulumi.String(container.Resources.Cpu)
			}
			if container.Resources.Memory != "" {
				limits["memory"] = pulumi.String(container.Resources.Memory)
			}
			if len(limits) > 0 {
				resources.Limits = limits
			}
			if container.Resources.CpuIdle != nil {
				resources.CpuIdle = pulumi.Bool(container.Resources.GetCpuIdle())
			}
			if container.Resources.StartupCpuBoost {
				resources.StartupCpuBoost = pulumi.Bool(true)
			}
			containerArgs.Resources = resources
		}

		if len(container.VolumeMounts) > 0 {
			mounts := cloudrunv2.ServiceTemplateContainerVolumeMountArray{}
			for _, mount := range container.VolumeMounts {
				mountArgs := &cloudrunv2.ServiceTemplateContainerVolumeMountArgs{
					Name:      pulumi.String(mount.Name),
					MountPath: pulumi.String(mount.MountPath),
				}
				if mount.SubPath != "" {
					mountArgs.SubPath = pulumi.String(mount.SubPath)
				}
				mounts = append(mounts, mountArgs)
			}
			containerArgs.VolumeMounts = mounts
		}

		// Startup probe: gates traffic and depends_on waiters until the
		// container is ready.
		if container.StartupProbe != nil {
			probe := container.StartupProbe
			startupProbe := &cloudrunv2.ServiceTemplateContainerStartupProbeArgs{}
			if probe.InitialDelaySeconds != nil {
				startupProbe.InitialDelaySeconds = pulumi.Int(int(probe.GetInitialDelaySeconds()))
			}
			if probe.TimeoutSeconds != nil {
				startupProbe.TimeoutSeconds = pulumi.Int(int(probe.GetTimeoutSeconds()))
			}
			if probe.PeriodSeconds != nil {
				startupProbe.PeriodSeconds = pulumi.Int(int(probe.GetPeriodSeconds()))
			}
			if probe.FailureThreshold != nil {
				startupProbe.FailureThreshold = pulumi.Int(int(probe.GetFailureThreshold()))
			}
			switch handler := probe.Handler.(type) {
			case *gcpcloudrunv1alpha1.GcpCloudRunStartupProbe_HttpGet:
				httpGet := &cloudrunv2.ServiceTemplateContainerStartupProbeHttpGetArgs{}
				if handler.HttpGet.Path != "" {
					httpGet.Path = pulumi.String(handler.HttpGet.Path)
				}
				if handler.HttpGet.Port != nil {
					httpGet.Port = pulumi.Int(int(handler.HttpGet.GetPort()))
				}
				if len(handler.HttpGet.HttpHeaders) > 0 {
					headers := cloudrunv2.ServiceTemplateContainerStartupProbeHttpGetHttpHeaderArray{}
					for _, header := range handler.HttpGet.HttpHeaders {
						headers = append(headers, &cloudrunv2.ServiceTemplateContainerStartupProbeHttpGetHttpHeaderArgs{
							Name:  pulumi.String(header.Name),
							Value: pulumi.String(header.Value),
						})
					}
					httpGet.HttpHeaders = headers
				}
				startupProbe.HttpGet = httpGet
			case *gcpcloudrunv1alpha1.GcpCloudRunStartupProbe_TcpSocket:
				tcpSocket := &cloudrunv2.ServiceTemplateContainerStartupProbeTcpSocketArgs{}
				if handler.TcpSocket.Port != nil {
					tcpSocket.Port = pulumi.Int(int(handler.TcpSocket.GetPort()))
				}
				startupProbe.TcpSocket = tcpSocket
			case *gcpcloudrunv1alpha1.GcpCloudRunStartupProbe_Grpc:
				grpc := &cloudrunv2.ServiceTemplateContainerStartupProbeGrpcArgs{}
				if handler.Grpc.Port != nil {
					grpc.Port = pulumi.Int(int(handler.Grpc.GetPort()))
				}
				if handler.Grpc.Service != "" {
					grpc.Service = pulumi.String(handler.Grpc.Service)
				}
				startupProbe.Grpc = grpc
			}
			containerArgs.StartupProbe = startupProbe
		}

		// Liveness probe: restarts an unhealthy container. HTTP/gRPC only —
		// the proto rejects TCP liveness (Cloud Run does not support it).
		if container.LivenessProbe != nil {
			probe := container.LivenessProbe
			livenessProbe := &cloudrunv2.ServiceTemplateContainerLivenessProbeArgs{}
			if probe.InitialDelaySeconds != nil {
				livenessProbe.InitialDelaySeconds = pulumi.Int(int(probe.GetInitialDelaySeconds()))
			}
			if probe.TimeoutSeconds != nil {
				livenessProbe.TimeoutSeconds = pulumi.Int(int(probe.GetTimeoutSeconds()))
			}
			if probe.PeriodSeconds != nil {
				livenessProbe.PeriodSeconds = pulumi.Int(int(probe.GetPeriodSeconds()))
			}
			if probe.FailureThreshold != nil {
				livenessProbe.FailureThreshold = pulumi.Int(int(probe.GetFailureThreshold()))
			}
			switch handler := probe.Handler.(type) {
			case *gcpcloudrunv1alpha1.GcpCloudRunLivenessProbe_HttpGet:
				httpGet := &cloudrunv2.ServiceTemplateContainerLivenessProbeHttpGetArgs{}
				if handler.HttpGet.Path != "" {
					httpGet.Path = pulumi.String(handler.HttpGet.Path)
				}
				if handler.HttpGet.Port != nil {
					httpGet.Port = pulumi.Int(int(handler.HttpGet.GetPort()))
				}
				if len(handler.HttpGet.HttpHeaders) > 0 {
					headers := cloudrunv2.ServiceTemplateContainerLivenessProbeHttpGetHttpHeaderArray{}
					for _, header := range handler.HttpGet.HttpHeaders {
						headers = append(headers, &cloudrunv2.ServiceTemplateContainerLivenessProbeHttpGetHttpHeaderArgs{
							Name:  pulumi.String(header.Name),
							Value: pulumi.String(header.Value),
						})
					}
					httpGet.HttpHeaders = headers
				}
				livenessProbe.HttpGet = httpGet
			case *gcpcloudrunv1alpha1.GcpCloudRunLivenessProbe_Grpc:
				grpc := &cloudrunv2.ServiceTemplateContainerLivenessProbeGrpcArgs{}
				if handler.Grpc.Port != nil {
					grpc.Port = pulumi.Int(int(handler.Grpc.GetPort()))
				}
				if handler.Grpc.Service != "" {
					grpc.Service = pulumi.String(handler.Grpc.Service)
				}
				livenessProbe.Grpc = grpc
			}
			containerArgs.LivenessProbe = livenessProbe
		}

		// Readiness probe: pulls a failing instance from serving (and
		// re-admits it on recovery) without restarting it. HTTP/gRPC only,
		// no initial delay — the API's readiness shape differs from the
		// other probes deliberately. The SDK's SuccessThreshold is never
		// sent: the GA API's probe message carries no such field and
		// silently drops it, leaving a perpetual re-plan diff
		// (live-verified; the spec has no such field).
		if container.ReadinessProbe != nil {
			probe := container.ReadinessProbe
			readinessProbe := &cloudrunv2.ServiceTemplateContainerReadinessProbeArgs{}
			if probe.TimeoutSeconds != nil {
				readinessProbe.TimeoutSeconds = pulumi.Int(int(probe.GetTimeoutSeconds()))
			}
			if probe.PeriodSeconds != nil {
				readinessProbe.PeriodSeconds = pulumi.Int(int(probe.GetPeriodSeconds()))
			}
			if probe.FailureThreshold != nil {
				readinessProbe.FailureThreshold = pulumi.Int(int(probe.GetFailureThreshold()))
			}
			switch handler := probe.Handler.(type) {
			case *gcpcloudrunv1alpha1.GcpCloudRunReadinessProbe_HttpGet:
				httpGet := &cloudrunv2.ServiceTemplateContainerReadinessProbeHttpGetArgs{}
				if handler.HttpGet.Path != "" {
					httpGet.Path = pulumi.String(handler.HttpGet.Path)
				}
				if handler.HttpGet.Port != nil {
					httpGet.Port = pulumi.Int(int(handler.HttpGet.GetPort()))
				}
				readinessProbe.HttpGet = httpGet
			case *gcpcloudrunv1alpha1.GcpCloudRunReadinessProbe_Grpc:
				grpc := &cloudrunv2.ServiceTemplateContainerReadinessProbeGrpcArgs{}
				if handler.Grpc.Port != nil {
					grpc.Port = pulumi.Int(int(handler.Grpc.GetPort()))
				}
				if handler.Grpc.Service != "" {
					grpc.Service = pulumi.String(handler.Grpc.Service)
				}
				readinessProbe.Grpc = grpc
			}
			containerArgs.ReadinessProbe = readinessProbe
		}

		containers = append(containers, containerArgs)
	}

	return containers
}

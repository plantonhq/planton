package module

import (
	"github.com/pkg/errors"
	gcpcloudrunworkerpoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrunworkerpool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// workerPool provisions the Cloud Run v2 worker pool -- a Cloud Run
// service with the request path removed: the same revision template, but
// no port, no ingress, no invoker IAM, and instances scaled to a count (or
// between bounds by the owner's signal) instead of by traffic. Every update
// that touches the template stamps out a new immutable revision;
// instance_splits decides how many instances each revision runs.
//
// Immutable: location, name, and a container's depends_on. Everything else
// rolls a new revision in place.
func workerPool(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpCloudRunWorkerPool.Spec

	// Enable the Cloud Run Admin API first so a fresh project works on the
	// first deploy. disable_on_destroy stays false: tearing down one pool
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
		"cloudrunwp-run.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable run.googleapis.com api")
	}

	// Secret values the env carries are stored in Secret Manager before the
	// pool exists, and each instance reads them by reference.
	storedSecrets, err := cloudrunenv.Store(ctx, secretPlacement(locals), secretVariables(spec), gcpProvider)
	if err != nil {
		return errors.Wrap(err, "failed to store the environment's secret values")
	}

	// Deletion guard, honest by default: an unset spec field means true, so
	// a destroy fails until the manifest explicitly opts out (identical to
	// the Terraform module).
	deletionProtection := true
	if spec.DeletionProtection != nil {
		deletionProtection = spec.GetDeletionProtection()
	}

	args := &cloudrunv2.WorkerPoolArgs{
		Name:               pulumi.String(locals.WorkerPoolName),
		Location:           pulumi.String(spec.Region),
		Template:           buildTemplate(spec, storedSecrets.Refs),
		Labels:             pulumi.ToStringMap(locals.GcpLabels),
		DeletionProtection: pulumi.Bool(deletionProtection),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringMap(spec.Annotations)
	}
	if spec.LaunchStage != "" {
		args.LaunchStage = pulumi.String(spec.LaunchStage)
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// the pool from management without deleting it in GCP.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	// Binary Authorization deploy gate: the project default policy XOR a
	// named platform policy (the proto rejects both).
	if spec.BinaryAuthorization != nil {
		binaryAuthorization := &cloudrunv2.WorkerPoolBinaryAuthorizationArgs{}
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

	// Instance count: MANUAL pins the total (Google's default mode);
	// AUTOMATIC moves between bounds. Each lever is sent only when set so
	// Google's defaults stay in charge otherwise.
	if spec.Scaling != nil {
		scaling := &cloudrunv2.WorkerPoolScalingArgs{}
		if spec.Scaling.ScalingMode != "" {
			scaling.ScalingMode = pulumi.String(spec.Scaling.ScalingMode)
		}
		if spec.Scaling.ManualInstanceCount != nil {
			scaling.ManualInstanceCount = pulumi.Int(int(spec.Scaling.GetManualInstanceCount()))
		}
		if spec.Scaling.MinInstanceCount != nil {
			scaling.MinInstanceCount = pulumi.Int(int(spec.Scaling.GetMinInstanceCount()))
		}
		if spec.Scaling.MaxInstanceCount != nil {
			scaling.MaxInstanceCount = pulumi.Int(int(spec.Scaling.GetMaxInstanceCount()))
		}
		args.Scaling = scaling
	}

	// Instance split across revisions. An empty spec list means "every
	// instance on the latest ready revision" -- achieved by omitting the
	// field so the provider applies GCP's default without a diff-prone
	// split.
	if len(spec.InstanceSplits) > 0 {
		splits := cloudrunv2.WorkerPoolInstanceSplitArray{}
		for _, split := range spec.InstanceSplits {
			splitArgs := &cloudrunv2.WorkerPoolInstanceSplitArgs{
				Type: pulumi.String(split.Type),
			}
			if split.Revision != "" {
				splitArgs.Revision = pulumi.String(split.Revision)
			}
			if split.Percent != nil {
				splitArgs.Percent = pulumi.Int(int(split.GetPercent()))
			}
			splits = append(splits, splitArgs)
		}
		args.InstanceSplits = splits
	}

	createdWorkerPool, err := cloudrunv2.NewWorkerPool(ctx,
		locals.GcpCloudRunWorkerPool.Metadata.Name,
		args,
		pulumi.Provider(gcpProvider),
		// The runtime identity must already read every secret the revision
		// references, so the grants land first.
		pulumi.DependsOn(append([]pulumi.Resource{createdProjectService}, storedSecrets.Grants...)),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create Cloud Run v2 worker pool")
	}

	ctx.Export(OpName, createdWorkerPool.ID().ToStringOutput())
	ctx.Export(OpWorkerPoolName, createdWorkerPool.Name)
	ctx.Export(OpUid, createdWorkerPool.Uid)
	ctx.Export(OpLocation, createdWorkerPool.Location)
	ctx.Export(OpProjectId, createdWorkerPool.Project)
	ctx.Export(OpLatestCreatedRevision, createdWorkerPool.LatestCreatedRevision)
	ctx.Export(OpLatestReadyRevision, createdWorkerPool.LatestReadyRevision)
	ctx.Export(OpObservedGeneration, createdWorkerPool.ObservedGeneration)
	ctx.Export(OpEtag, createdWorkerPool.Etag)

	return nil
}

// buildTemplate maps the spec's revision-level surface onto the v2 worker
// pool template: containers, volumes, networking, hardware, encryption.
func buildTemplate(
	spec *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolSpec,
	secretRefs map[cloudrunenv.Key]cloudrunenv.Ref,
) *cloudrunv2.WorkerPoolTemplateArgs {
	template := &cloudrunv2.WorkerPoolTemplateArgs{
		Containers: buildContainers(spec, secretRefs),
	}

	// Explicit revision naming makes declarative rollouts by revision
	// possible; unset (the norm) lets Cloud Run generate names.
	if spec.Revision != "" {
		template.Revision = pulumi.String(spec.Revision)
	}

	// Revision-level metadata, stamped on every revision the template
	// creates (distinct from the pool-object labels/annotations).
	if len(spec.RevisionLabels) > 0 {
		template.Labels = pulumi.ToStringMap(spec.RevisionLabels)
	}
	if len(spec.RevisionAnnotations) > 0 {
		template.Annotations = pulumi.ToStringMap(spec.RevisionAnnotations)
	}

	// The runtime identity whose permissions the code exercises. Unset uses
	// the project's Compute Engine default service account.
	if spec.ServiceAccount.GetValue() != "" {
		template.ServiceAccount = pulumi.String(spec.ServiceAccount.GetValue())
	}

	// CMEK for the deployed images, and what happens to running instances
	// if the key is revoked (the proto ties the two levers to the key).
	if spec.EncryptionKey.GetValue() != "" {
		template.EncryptionKey = pulumi.String(spec.EncryptionKey.GetValue())
	}
	if spec.EncryptionKeyRevocationAction != "" {
		template.EncryptionKeyRevocationAction = pulumi.String(spec.EncryptionKeyRevocationAction)
	}
	if spec.EncryptionKeyShutdownDuration != "" {
		template.EncryptionKeyShutdownDuration = pulumi.String(spec.EncryptionKeyShutdownDuration)
	}

	// Single-zone GPU serving opt-in (cheaper GPU capacity for zonal risk).
	if spec.GpuZonalRedundancyDisabled {
		template.GpuZonalRedundancyDisabled = pulumi.Bool(true)
	}

	// GPU hardware requirement (e.g. "nvidia-l4").
	if spec.NodeSelector != nil {
		template.NodeSelector = &cloudrunv2.WorkerPoolTemplateNodeSelectorArgs{
			Accelerator: pulumi.String(spec.NodeSelector.Accelerator),
		}
	}

	// Outbound VPC networking: a Serverless VPC Access connector XOR direct
	// VPC egress network_interfaces -- the proto guarantees exactly one.
	if spec.VpcAccess != nil {
		vpcAccess := &cloudrunv2.WorkerPoolTemplateVpcAccessArgs{}
		if spec.VpcAccess.Connector.GetValue() != "" {
			vpcAccess.Connector = pulumi.String(spec.VpcAccess.Connector.GetValue())
		}
		if spec.VpcAccess.Egress != "" {
			vpcAccess.Egress = pulumi.String(spec.VpcAccess.Egress)
		}
		if len(spec.VpcAccess.NetworkInterfaces) > 0 {
			interfaces := cloudrunv2.WorkerPoolTemplateVpcAccessNetworkInterfaceArray{}
			for _, networkInterface := range spec.VpcAccess.NetworkInterfaces {
				interfaceArgs := &cloudrunv2.WorkerPoolTemplateVpcAccessNetworkInterfaceArgs{}
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
		volumes := cloudrunv2.WorkerPoolTemplateVolumeArray{}
		for _, volume := range spec.Volumes {
			volumeArgs := &cloudrunv2.WorkerPoolTemplateVolumeArgs{
				Name: pulumi.String(volume.Name),
			}
			switch source := volume.Source.(type) {
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolVolume_CloudSqlInstance:
				// Cloud SQL Unix sockets -- GCP manages the proxying;
				// connect via /cloudsql/<project:region:instance>.
				instances := pulumi.StringArray{}
				for _, instance := range source.CloudSqlInstance.Instances {
					instances = append(instances, pulumi.String(instance.GetValue()))
				}
				volumeArgs.CloudSqlInstance = &cloudrunv2.WorkerPoolTemplateVolumeCloudSqlInstanceArgs{
					Instances: instances,
				}
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolVolume_Secret:
				secret := &cloudrunv2.WorkerPoolTemplateVolumeSecretArgs{
					Secret: pulumi.String(source.Secret.Secret),
				}
				if source.Secret.DefaultMode != nil {
					secret.DefaultMode = pulumi.Int(int(source.Secret.GetDefaultMode()))
				}
				if len(source.Secret.Items) > 0 {
					items := cloudrunv2.WorkerPoolTemplateVolumeSecretItemArray{}
					for _, item := range source.Secret.Items {
						itemArgs := &cloudrunv2.WorkerPoolTemplateVolumeSecretItemArgs{
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
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolVolume_EmptyDir:
				emptyDir := &cloudrunv2.WorkerPoolTemplateVolumeEmptyDirArgs{}
				if source.EmptyDir.Medium != "" {
					emptyDir.Medium = pulumi.String(source.EmptyDir.Medium)
				}
				if source.EmptyDir.SizeLimit != "" {
					emptyDir.SizeLimit = pulumi.String(source.EmptyDir.SizeLimit)
				}
				volumeArgs.EmptyDir = emptyDir
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolVolume_Gcs:
				gcs := &cloudrunv2.WorkerPoolTemplateVolumeGcsArgs{
					Bucket:   pulumi.String(source.Gcs.Bucket.GetValue()),
					ReadOnly: pulumi.Bool(source.Gcs.ReadOnly),
				}
				if len(source.Gcs.MountOptions) > 0 {
					gcs.MountOptions = pulumi.ToStringArray(source.Gcs.MountOptions)
				}
				volumeArgs.Gcs = gcs
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolVolume_Nfs:
				volumeArgs.Nfs = &cloudrunv2.WorkerPoolTemplateVolumeNfsArgs{
					Server:   pulumi.String(source.Nfs.Server),
					Path:     pulumi.String(source.Nfs.Path),
					ReadOnly: pulumi.Bool(source.Nfs.ReadOnly),
				}
			}
			volumes = append(volumes, volumeArgs)
		}
		template.Volumes = volumes
	}

	return template
}

// buildContainers maps the spec's containers -- the worker plus any
// sidecars sharing localhost and volumes, ordered by depends_on.
func buildContainers(
	spec *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolSpec,
	secretRefs map[cloudrunenv.Key]cloudrunenv.Ref,
) cloudrunv2.WorkerPoolTemplateContainerArray {
	containers := cloudrunv2.WorkerPoolTemplateContainerArray{}

	for containerIndex, container := range spec.Containers {
		containerArgs := &cloudrunv2.WorkerPoolTemplateContainerArgs{
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

		if len(container.Env) > 0 {
			envs := cloudrunv2.WorkerPoolTemplateContainerEnvArray{}
			for _, envVar := range container.Env {
				envArgs := &cloudrunv2.WorkerPoolTemplateContainerEnvArgs{
					Name: pulumi.String(envVar.Name),
				}
				// A literal, a Secret Manager secret the author owns, or a
				// secret value this module stored (one of the three).
				if ref, stored := secretRefs[cloudrunenv.Key{ContainerIndex: containerIndex, Name: envVar.Name}]; stored {
					envArgs.ValueSource = &cloudrunv2.WorkerPoolTemplateContainerEnvValueSourceArgs{
						SecretKeyRef: &cloudrunv2.WorkerPoolTemplateContainerEnvValueSourceSecretKeyRefArgs{
							Secret:  ref.Secret,
							Version: ref.Version,
						},
					}
				} else if envVar.ValueFromSecret != nil {
					secretKeyRef := &cloudrunv2.WorkerPoolTemplateContainerEnvValueSourceSecretKeyRefArgs{
						Secret: pulumi.String(envVar.ValueFromSecret.Secret),
					}
					if envVar.ValueFromSecret.Version != "" {
						secretKeyRef.Version = pulumi.String(envVar.ValueFromSecret.Version)
					}
					envArgs.ValueSource = &cloudrunv2.WorkerPoolTemplateContainerEnvValueSourceArgs{
						SecretKeyRef: secretKeyRef,
					}
				} else {
					envArgs.Value = pulumi.String(envVar.Value)
				}
				envs = append(envs, envArgs)
			}
			containerArgs.Envs = envs
		}

		// CPU/memory land in the API's limits map. A worker pool has no
		// idle-CPU or startup-boost lever: CPU is always allocated.
		if container.Resources != nil {
			limits := pulumi.StringMap{}
			if container.Resources.Cpu != "" {
				limits["cpu"] = pulumi.String(container.Resources.Cpu)
			}
			if container.Resources.Memory != "" {
				limits["memory"] = pulumi.String(container.Resources.Memory)
			}
			if len(limits) > 0 {
				containerArgs.Resources = &cloudrunv2.WorkerPoolTemplateContainerResourcesArgs{Limits: limits}
			}
		}

		if len(container.VolumeMounts) > 0 {
			mounts := cloudrunv2.WorkerPoolTemplateContainerVolumeMountArray{}
			for _, mount := range container.VolumeMounts {
				mountArgs := &cloudrunv2.WorkerPoolTemplateContainerVolumeMountArgs{
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

		// Startup probe: gates depends_on waiters until the container is
		// ready; HTTP, TCP, or gRPC.
		if container.StartupProbe != nil {
			probe := container.StartupProbe
			startupProbe := &cloudrunv2.WorkerPoolTemplateContainerStartupProbeArgs{}
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
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolStartupProbe_HttpGet:
				httpGet := &cloudrunv2.WorkerPoolTemplateContainerStartupProbeHttpGetArgs{}
				if handler.HttpGet.Path != "" {
					httpGet.Path = pulumi.String(handler.HttpGet.Path)
				}
				if handler.HttpGet.Port != nil {
					httpGet.Port = pulumi.Int(int(handler.HttpGet.GetPort()))
				}
				// The pinned SDK models a worker-pool probe's headers as one
				// header (the spec caps the list at one for that reason).
				if len(handler.HttpGet.HttpHeaders) > 0 {
					header := handler.HttpGet.HttpHeaders[0]
					httpGet.HttpHeaders = &cloudrunv2.WorkerPoolTemplateContainerStartupProbeHttpGetHttpHeadersArgs{
						Name:  pulumi.String(header.Name),
						Value: pulumi.String(header.Value),
					}
				}
				startupProbe.HttpGet = httpGet
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolStartupProbe_TcpSocket:
				tcpSocket := &cloudrunv2.WorkerPoolTemplateContainerStartupProbeTcpSocketArgs{}
				if handler.TcpSocket.Port != nil {
					tcpSocket.Port = pulumi.Int(int(handler.TcpSocket.GetPort()))
				}
				startupProbe.TcpSocket = tcpSocket
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolStartupProbe_Grpc:
				grpc := &cloudrunv2.WorkerPoolTemplateContainerStartupProbeGrpcArgs{}
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

		// Liveness probe: restarts an unhealthy container. HTTP/gRPC only --
		// the proto rejects TCP liveness (Cloud Run does not support it).
		if container.LivenessProbe != nil {
			probe := container.LivenessProbe
			livenessProbe := &cloudrunv2.WorkerPoolTemplateContainerLivenessProbeArgs{}
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
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolLivenessProbe_HttpGet:
				httpGet := &cloudrunv2.WorkerPoolTemplateContainerLivenessProbeHttpGetArgs{}
				if handler.HttpGet.Path != "" {
					httpGet.Path = pulumi.String(handler.HttpGet.Path)
				}
				if handler.HttpGet.Port != nil {
					httpGet.Port = pulumi.Int(int(handler.HttpGet.GetPort()))
				}
				if len(handler.HttpGet.HttpHeaders) > 0 {
					header := handler.HttpGet.HttpHeaders[0]
					httpGet.HttpHeaders = &cloudrunv2.WorkerPoolTemplateContainerLivenessProbeHttpGetHttpHeadersArgs{
						Name:  pulumi.String(header.Name),
						Value: pulumi.String(header.Value),
					}
				}
				livenessProbe.HttpGet = httpGet
			case *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolLivenessProbe_Grpc:
				grpc := &cloudrunv2.WorkerPoolTemplateContainerLivenessProbeGrpcArgs{}
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

		containers = append(containers, containerArgs)
	}

	return containers
}

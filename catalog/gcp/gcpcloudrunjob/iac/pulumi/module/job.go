package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpcloudrunjobv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrunjob/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/cloudrunv2"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// job provisions the Cloud Run v2 job — a run-to-completion workload
// definition. Each trigger (manual, Scheduler, Eventarc) creates an
// execution that stamps out task_count tasks with up to parallelism
// running concurrently.
func job(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*cloudrunv2.Job, error) {
	spec := locals.GcpCloudRunJob.Spec
	tmpl := spec.Template

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
	// job exists, and each task reads them by reference.
	storedSecrets, err := cloudrunenv.Store(ctx, secretPlacement(locals), secretVariables(tmpl), gcpProvider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store the environment's secret values")
	}

	deletionProtection := true
	if spec.DeletionProtection != nil {
		deletionProtection = spec.GetDeletionProtection()
	}

	taskTemplate := buildTaskTemplate(spec, tmpl, storedSecrets.Refs)

	executionTemplate := &cloudrunv2.JobTemplateArgs{
		Template: taskTemplate,
	}
	if spec.TaskCount != nil {
		executionTemplate.TaskCount = pulumi.Int(int(spec.GetTaskCount()))
	}
	if spec.Parallelism != nil {
		executionTemplate.Parallelism = pulumi.Int(int(spec.GetParallelism()))
	}

	// Execution-level metadata, stamped on every execution the job creates
	// (distinct from the job-object labels/annotations below).
	if len(spec.ExecutionLabels) > 0 {
		executionTemplate.Labels = pulumi.ToStringMap(spec.ExecutionLabels)
	}
	if len(spec.ExecutionAnnotations) > 0 {
		executionTemplate.Annotations = pulumi.ToStringMap(spec.ExecutionAnnotations)
	}

	args := &cloudrunv2.JobArgs{
		Name:               pulumi.String(locals.JobName),
		Location:           pulumi.String(spec.Region),
		Template:           executionTemplate,
		Labels:             pulumi.ToStringMap(locals.GcpLabels),
		DeletionProtection: pulumi.Bool(deletionProtection),
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.LaunchStage != "" {
		args.LaunchStage = pulumi.String(spec.LaunchStage)
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringMap(spec.Annotations)
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// the job from management without deleting it in GCP.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	// Resource Manager tags, bound at creation only (ForceNew): omitted
	// when empty so a job without tags carries no tag surface.
	if len(spec.ResourceManagerTags) > 0 {
		args.Tags = pulumi.ToStringMap(spec.ResourceManagerTags)
	}

	// Declarative run-on-deploy tokens (mutually exclusive —
	// proto-enforced): start_* counts the job ready when the triggered
	// execution STARTS; run_* when it COMPLETES.
	if spec.StartExecutionToken != "" {
		args.StartExecutionToken = pulumi.String(spec.StartExecutionToken)
	}
	if spec.RunExecutionToken != "" {
		args.RunExecutionToken = pulumi.String(spec.RunExecutionToken)
	}

	if spec.BinaryAuthorization != nil {
		binaryAuthorization := &cloudrunv2.JobBinaryAuthorizationArgs{}
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

	createdJob, err := cloudrunv2.NewJob(ctx,
		locals.GcpCloudRunJob.Metadata.Name,
		args,
		pulumi.Provider(gcpProvider),
		// The runtime identity must already read every secret the task
		// template references, so the grants land first.
		pulumi.DependsOn(append([]pulumi.Resource{createdProjectService}, storedSecrets.Grants...)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Cloud Run v2 job")
	}

	return createdJob, nil
}

func buildTaskTemplate(
	spec *gcpcloudrunjobv1alpha1.GcpCloudRunJobSpec,
	tmpl *gcpcloudrunjobv1alpha1.GcpCloudRunJobTemplate,
	secretRefs map[cloudrunenv.Key]cloudrunenv.Ref,
) *cloudrunv2.JobTemplateTemplateArgs {
	taskTemplate := &cloudrunv2.JobTemplateTemplateArgs{
		Containers: buildContainers(tmpl, secretRefs),
	}

	if tmpl.ServiceAccount.GetValue() != "" {
		taskTemplate.ServiceAccount = pulumi.String(tmpl.ServiceAccount.GetValue())
	}

	if tmpl.ExecutionEnvironment != gcpcloudrunjobv1alpha1.GcpCloudRunJobExecutionEnvironment_EXECUTION_ENVIRONMENT_UNSPECIFIED {
		taskTemplate.ExecutionEnvironment = pulumi.String(tmpl.ExecutionEnvironment.String())
	}

	if tmpl.EncryptionKey.GetValue() != "" {
		taskTemplate.EncryptionKey = pulumi.String(tmpl.EncryptionKey.GetValue())
	}

	if tmpl.TimeoutSeconds != nil {
		taskTemplate.Timeout = pulumi.String(fmt.Sprintf("%ds", tmpl.GetTimeoutSeconds()))
	}

	if tmpl.MaxRetries != nil {
		taskTemplate.MaxRetries = pulumi.Int(int(tmpl.GetMaxRetries()))
	}

	if spec.GpuZonalRedundancyDisabled {
		taskTemplate.GpuZonalRedundancyDisabled = pulumi.Bool(true)
	}

	if tmpl.NodeSelector != nil {
		taskTemplate.NodeSelector = &cloudrunv2.JobTemplateTemplateNodeSelectorArgs{
			Accelerator: pulumi.String(tmpl.NodeSelector.Accelerator),
		}
	}

	if tmpl.VpcAccess != nil {
		vpcAccess := &cloudrunv2.JobTemplateTemplateVpcAccessArgs{}
		if tmpl.VpcAccess.Connector.GetValue() != "" {
			vpcAccess.Connector = pulumi.String(tmpl.VpcAccess.Connector.GetValue())
		}
		if tmpl.VpcAccess.Egress != "" {
			vpcAccess.Egress = pulumi.String(tmpl.VpcAccess.Egress)
		}
		if len(tmpl.VpcAccess.NetworkInterfaces) > 0 {
			interfaces := cloudrunv2.JobTemplateTemplateVpcAccessNetworkInterfaceArray{}
			for _, networkInterface := range tmpl.VpcAccess.NetworkInterfaces {
				interfaceArgs := &cloudrunv2.JobTemplateTemplateVpcAccessNetworkInterfaceArgs{}
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
		taskTemplate.VpcAccess = vpcAccess
	}

	if len(tmpl.Volumes) > 0 {
		volumes := cloudrunv2.JobTemplateTemplateVolumeArray{}
		for _, volume := range tmpl.Volumes {
			volumeArgs := &cloudrunv2.JobTemplateTemplateVolumeArgs{
				Name: pulumi.String(volume.Name),
			}
			switch source := volume.Source.(type) {
			case *gcpcloudrunjobv1alpha1.GcpCloudRunJobVolume_CloudSqlInstance:
				instances := pulumi.StringArray{}
				for _, instance := range source.CloudSqlInstance.Instances {
					instances = append(instances, pulumi.String(instance.GetValue()))
				}
				volumeArgs.CloudSqlInstance = &cloudrunv2.JobTemplateTemplateVolumeCloudSqlInstanceArgs{
					Instances: instances,
				}
			case *gcpcloudrunjobv1alpha1.GcpCloudRunJobVolume_Secret:
				secret := &cloudrunv2.JobTemplateTemplateVolumeSecretArgs{
					Secret: pulumi.String(source.Secret.Secret),
				}
				if source.Secret.DefaultMode != nil {
					secret.DefaultMode = pulumi.Int(int(source.Secret.GetDefaultMode()))
				}
				if len(source.Secret.Items) > 0 {
					items := cloudrunv2.JobTemplateTemplateVolumeSecretItemArray{}
					for _, item := range source.Secret.Items {
						itemArgs := &cloudrunv2.JobTemplateTemplateVolumeSecretItemArgs{
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
			case *gcpcloudrunjobv1alpha1.GcpCloudRunJobVolume_EmptyDir:
				emptyDir := &cloudrunv2.JobTemplateTemplateVolumeEmptyDirArgs{}
				if source.EmptyDir.Medium != "" {
					emptyDir.Medium = pulumi.String(source.EmptyDir.Medium)
				}
				if source.EmptyDir.SizeLimit != "" {
					emptyDir.SizeLimit = pulumi.String(source.EmptyDir.SizeLimit)
				}
				volumeArgs.EmptyDir = emptyDir
			case *gcpcloudrunjobv1alpha1.GcpCloudRunJobVolume_Gcs:
				gcs := &cloudrunv2.JobTemplateTemplateVolumeGcsArgs{
					Bucket:   pulumi.String(source.Gcs.Bucket.GetValue()),
					ReadOnly: pulumi.Bool(source.Gcs.ReadOnly),
				}
				if len(source.Gcs.MountOptions) > 0 {
					gcs.MountOptions = pulumi.ToStringArray(source.Gcs.MountOptions)
				}
				volumeArgs.Gcs = gcs
			case *gcpcloudrunjobv1alpha1.GcpCloudRunJobVolume_Nfs:
				volumeArgs.Nfs = &cloudrunv2.JobTemplateTemplateVolumeNfsArgs{
					Server:   pulumi.String(source.Nfs.Server),
					Path:     pulumi.String(source.Nfs.Path),
					ReadOnly: pulumi.Bool(source.Nfs.ReadOnly),
				}
			}
			volumes = append(volumes, volumeArgs)
		}
		taskTemplate.Volumes = volumes
	}

	return taskTemplate
}

func buildContainers(tmpl *gcpcloudrunjobv1alpha1.GcpCloudRunJobTemplate, secretRefs map[cloudrunenv.Key]cloudrunenv.Ref) cloudrunv2.JobTemplateTemplateContainerArray {
	containers := cloudrunv2.JobTemplateTemplateContainerArray{}

	for containerIndex, container := range tmpl.Containers {
		containerArgs := &cloudrunv2.JobTemplateTemplateContainerArgs{
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
			envs := cloudrunv2.JobTemplateTemplateContainerEnvArray{}
			for _, envVar := range container.Env {
				envArgs := &cloudrunv2.JobTemplateTemplateContainerEnvArgs{
					Name: pulumi.String(envVar.Name),
				}
				// A literal, a Secret Manager secret the author owns, or a
				// secret value this module stored (one of the three).
				if ref, stored := secretRefs[cloudrunenv.Key{ContainerIndex: containerIndex, Name: envVar.Name}]; stored {
					envArgs.ValueSource = &cloudrunv2.JobTemplateTemplateContainerEnvValueSourceArgs{
						SecretKeyRef: &cloudrunv2.JobTemplateTemplateContainerEnvValueSourceSecretKeyRefArgs{
							Secret:  ref.Secret,
							Version: ref.Version,
						},
					}
				} else if envVar.ValueFromSecret != nil {
					secretKeyRef := &cloudrunv2.JobTemplateTemplateContainerEnvValueSourceSecretKeyRefArgs{
						Secret: pulumi.String(envVar.ValueFromSecret.Secret),
					}
					if envVar.ValueFromSecret.Version != "" {
						secretKeyRef.Version = pulumi.String(envVar.ValueFromSecret.Version)
					}
					envArgs.ValueSource = &cloudrunv2.JobTemplateTemplateContainerEnvValueSourceArgs{
						SecretKeyRef: secretKeyRef,
					}
				} else {
					envArgs.Value = pulumi.String(envVar.Value)
				}
				envs = append(envs, envArgs)
			}
			containerArgs.Envs = envs
		}

		if container.Resources != nil {
			limits := pulumi.StringMap{}
			if container.Resources.Cpu != "" {
				limits["cpu"] = pulumi.String(container.Resources.Cpu)
			}
			if container.Resources.Memory != "" {
				limits["memory"] = pulumi.String(container.Resources.Memory)
			}
			if len(limits) > 0 {
				containerArgs.Resources = &cloudrunv2.JobTemplateTemplateContainerResourcesArgs{
					Limits: limits,
				}
			}
		}

		if len(container.VolumeMounts) > 0 {
			mounts := cloudrunv2.JobTemplateTemplateContainerVolumeMountArray{}
			for _, mount := range container.VolumeMounts {
				mountArgs := &cloudrunv2.JobTemplateTemplateContainerVolumeMountArgs{
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

		// The port a probe targets by default — jobs serve no traffic, so
		// this exists purely for the startup probe. The provider models
		// job ports as a list; the spec's single message lands as its one
		// element (the API accepts at most one).
		if container.Ports != nil {
			port := &cloudrunv2.JobTemplateTemplateContainerPortArgs{}
			if container.Ports.ContainerPort != nil {
				port.ContainerPort = pulumi.Int(int(container.Ports.GetContainerPort()))
			}
			if container.Ports.Name != "" {
				port.Name = pulumi.String(container.Ports.Name)
			}
			containerArgs.Ports = cloudrunv2.JobTemplateTemplateContainerPortArray{port}
		}

		// Startup probe: gates task start; a container that never passes
		// is shut down and the task retried per max_retries. The only
		// probe type jobs have.
		if container.StartupProbe != nil {
			probe := container.StartupProbe
			startupProbe := &cloudrunv2.JobTemplateTemplateContainerStartupProbeArgs{}
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
			case *gcpcloudrunjobv1alpha1.GcpCloudRunJobProbe_HttpGet:
				httpGet := &cloudrunv2.JobTemplateTemplateContainerStartupProbeHttpGetArgs{}
				if handler.HttpGet.Path != "" {
					httpGet.Path = pulumi.String(handler.HttpGet.Path)
				}
				if handler.HttpGet.Port != nil {
					httpGet.Port = pulumi.Int(int(handler.HttpGet.GetPort()))
				}
				if len(handler.HttpGet.HttpHeaders) > 0 {
					headers := cloudrunv2.JobTemplateTemplateContainerStartupProbeHttpGetHttpHeaderArray{}
					for _, header := range handler.HttpGet.HttpHeaders {
						headers = append(headers, &cloudrunv2.JobTemplateTemplateContainerStartupProbeHttpGetHttpHeaderArgs{
							Name:  pulumi.String(header.Name),
							Value: pulumi.String(header.Value),
						})
					}
					httpGet.HttpHeaders = headers
				}
				startupProbe.HttpGet = httpGet
			case *gcpcloudrunjobv1alpha1.GcpCloudRunJobProbe_TcpSocket:
				tcpSocket := &cloudrunv2.JobTemplateTemplateContainerStartupProbeTcpSocketArgs{}
				if handler.TcpSocket.Port != nil {
					tcpSocket.Port = pulumi.Int(int(handler.TcpSocket.GetPort()))
				}
				startupProbe.TcpSocket = tcpSocket
			case *gcpcloudrunjobv1alpha1.GcpCloudRunJobProbe_Grpc:
				grpc := &cloudrunv2.JobTemplateTemplateContainerStartupProbeGrpcArgs{}
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

		containers = append(containers, containerArgs)
	}

	return containers
}

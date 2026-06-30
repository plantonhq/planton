package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	kubernetesdeploymentv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerenv"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func deployment(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) (*appsv1.Deployment, error) {

	// create service account
	serviceAccountArgs := &kubernetescorev1.ServiceAccountArgs{
		Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.KubernetesDeployment.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
		}),
	}

	saOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdServiceAccount, err := kubernetescorev1.NewServiceAccount(ctx,
		locals.KubernetesDeployment.Metadata.Name,
		serviceAccountArgs,
		saOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add service account")
	}

	envVarInputs := containerenv.BuildEnvVars(
		locals.KubernetesDeployment.Spec.Container.App.Env,
		locals.EnvSecretName,
	)

	envFromInputs := containerenv.BuildEnvFrom(locals.KubernetesDeployment.Spec.Container.App.Env)

	portsArray := make(kubernetescorev1.ContainerPortArray, 0)
	for _, p := range locals.KubernetesDeployment.Spec.Container.App.Ports {
		portsArray = append(portsArray, &kubernetescorev1.ContainerPortArgs{
			Name:          pulumi.String(p.Name),
			ContainerPort: pulumi.Int(p.ContainerPort),
		})
	}

	// Build volume mounts and volumes from spec
	volumeMounts, volumes := buildVolumeMountsAndVolumes(locals.KubernetesDeployment.Spec)

	// Build container args
	containerArgs := &kubernetescorev1.ContainerArgs{
		Name: pulumi.String("microservice"),
		Image: pulumi.String(fmt.Sprintf("%s:%s",
			locals.KubernetesDeployment.Spec.Container.App.Image.Repo,
			locals.KubernetesDeployment.Spec.Container.App.Image.Tag)),
		Env:          kubernetescorev1.EnvVarArray(envVarInputs),
		EnvFrom:      envFromInputs,
		Ports:        portsArray,
		VolumeMounts: volumeMounts,
		Resources: kubernetescorev1.ResourceRequirementsArgs{
			Limits: pulumi.ToStringMap(map[string]string{
				"cpu":    locals.KubernetesDeployment.Spec.Container.App.Resources.Limits.Cpu,
				"memory": locals.KubernetesDeployment.Spec.Container.App.Resources.Limits.Memory,
			}),
			Requests: pulumi.ToStringMap(map[string]string{
				"cpu":    locals.KubernetesDeployment.Spec.Container.App.Resources.Requests.Cpu,
				"memory": locals.KubernetesDeployment.Spec.Container.App.Resources.Requests.Memory,
			}),
		},
		LivenessProbe:  buildProbe(locals.KubernetesDeployment.Spec.Container.App.LivenessProbe),
		ReadinessProbe: buildProbe(locals.KubernetesDeployment.Spec.Container.App.ReadinessProbe),
		StartupProbe:   buildProbe(locals.KubernetesDeployment.Spec.Container.App.StartupProbe),
		Lifecycle: kubernetescorev1.LifecycleArgs{
			PreStop: kubernetescorev1.LifecycleHandlerArgs{
				Exec: kubernetescorev1.ExecActionArgs{
					//wait for 60 seconds before killing the main process
					Command: pulumi.ToStringArray([]string{"/bin/sleep", "60"}),
				},
			},
		},
	}

	// Add command if specified
	if len(locals.KubernetesDeployment.Spec.Container.App.Command) > 0 {
		containerArgs.Command = pulumi.ToStringArray(locals.KubernetesDeployment.Spec.Container.App.Command)
	}

	// Add args if specified
	if len(locals.KubernetesDeployment.Spec.Container.App.Args) > 0 {
		containerArgs.Args = pulumi.ToStringArray(locals.KubernetesDeployment.Spec.Container.App.Args)
	}

	containerInputs := make([]kubernetescorev1.ContainerInput, 0)
	//add main container
	containerInputs = append(containerInputs, kubernetescorev1.ContainerInput(containerArgs))

	podSpecArgs := &kubernetescorev1.PodSpecArgs{
		ServiceAccountName: createdServiceAccount.Metadata.Name(),
		Containers:         kubernetescorev1.ContainerArray(containerInputs),
		Volumes:            volumes,
		//wait for 60 seconds before sending the termination signal to the processes in the pod
		TerminationGracePeriodSeconds: pulumi.IntPtr(60),
	}

	// Create image pull secret if configured
	createdImagePullSecret, err := imagePullSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return nil, err
	}
	if createdImagePullSecret != nil {
		podSpecArgs.ImagePullSecrets = kubernetescorev1.LocalObjectReferenceArray{
			kubernetescorev1.LocalObjectReferenceArgs{
				Name: createdImagePullSecret.Metadata.Name(),
			}}
	}

	//create deployment
	deploymentArgs := &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.KubernetesDeployment.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
			Annotations: pulumi.StringMap{
				"pulumi.com/patchForce": pulumi.String("true"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(locals.KubernetesDeployment.Spec.Availability.MinReplicas),
			Strategy: buildDeploymentStrategy(locals.KubernetesDeployment.Spec.Availability.DeploymentStrategy),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.ToStringMap(locals.SelectorLabels),
			},
			Template: &kubernetescorev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.ToStringMap(locals.Labels),
				},
				Spec: podSpecArgs,
			},
		},
	}

	ignoreChangesOpt := pulumi.IgnoreChanges([]string{
		//WARNING: adding metdata.managedFields to ignoreChanges is rejected from kubernetes api-server for some reason
		//although the issue must have been resolved by now,per, https://github.com/pulumi/pulumi-kubernetes/issues/1075,
		//apparently it is not.
		//error from the api-server is "metadata.managedFields must be nil"
		//"metadata.managedFields", "status",
	})

	deployOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider), ignoreChangesOpt}, namespaceDeps...)
	createdDeployment, err := appsv1.NewDeployment(ctx,
		locals.KubernetesDeployment.Metadata.Name,
		deploymentArgs,
		deployOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add deployment")
	}

	return createdDeployment, nil
}

// buildProbe converts a proto Probe definition into a Pulumi Kubernetes ProbeArgs.
// Returns nil if the probe is not configured.
func buildProbe(protoProbe *kubernetesv1.Probe) *kubernetescorev1.ProbeArgs {
	if protoProbe == nil {
		return nil
	}

	probe := &kubernetescorev1.ProbeArgs{}

	// Set timing configuration
	if protoProbe.InitialDelaySeconds > 0 {
		probe.InitialDelaySeconds = pulumi.Int(protoProbe.InitialDelaySeconds)
	}
	if protoProbe.PeriodSeconds > 0 {
		probe.PeriodSeconds = pulumi.Int(protoProbe.PeriodSeconds)
	}
	if protoProbe.TimeoutSeconds > 0 {
		probe.TimeoutSeconds = pulumi.Int(protoProbe.TimeoutSeconds)
	}
	if protoProbe.SuccessThreshold > 0 {
		probe.SuccessThreshold = pulumi.Int(protoProbe.SuccessThreshold)
	}
	if protoProbe.FailureThreshold > 0 {
		probe.FailureThreshold = pulumi.Int(protoProbe.FailureThreshold)
	}

	// Set handler based on the configured type
	switch handler := protoProbe.Handler.(type) {
	case *kubernetesv1.Probe_Grpc:
		grpcAction := &kubernetescorev1.GRPCActionArgs{
			Port: pulumi.Int(handler.Grpc.Port),
		}
		if handler.Grpc.Service != "" {
			grpcAction.Service = pulumi.StringPtr(handler.Grpc.Service)
		}
		probe.Grpc = grpcAction

	case *kubernetesv1.Probe_HttpGet:
		httpGet := &kubernetescorev1.HTTPGetActionArgs{}

		if handler.HttpGet.Path != "" {
			httpGet.Path = pulumi.String(handler.HttpGet.Path)
		}

		// Handle port (either port_number or port_name)
		switch port := handler.HttpGet.Port.(type) {
		case *kubernetesv1.HTTPGetAction_PortNumber:
			httpGet.Port = pulumi.Int(port.PortNumber)
		case *kubernetesv1.HTTPGetAction_PortName:
			httpGet.Port = pulumi.String(port.PortName)
		}

		if handler.HttpGet.Host != "" {
			httpGet.Host = pulumi.String(handler.HttpGet.Host)
		}

		if handler.HttpGet.Scheme != "" {
			httpGet.Scheme = pulumi.String(handler.HttpGet.Scheme)
		}

		// Add custom headers if any
		if len(handler.HttpGet.HttpHeaders) > 0 {
			headers := make(kubernetescorev1.HTTPHeaderArray, 0, len(handler.HttpGet.HttpHeaders))
			for _, h := range handler.HttpGet.HttpHeaders {
				headers = append(headers, &kubernetescorev1.HTTPHeaderArgs{
					Name:  pulumi.String(h.Name),
					Value: pulumi.String(h.Value),
				})
			}
			httpGet.HttpHeaders = headers
		}

		probe.HttpGet = httpGet

	case *kubernetesv1.Probe_TcpSocket:
		tcpSocket := &kubernetescorev1.TCPSocketActionArgs{}

		// Handle port (either port_number or port_name)
		switch port := handler.TcpSocket.Port.(type) {
		case *kubernetesv1.TCPSocketAction_PortNumber:
			tcpSocket.Port = pulumi.Int(port.PortNumber)
		case *kubernetesv1.TCPSocketAction_PortName:
			tcpSocket.Port = pulumi.String(port.PortName)
		}

		if handler.TcpSocket.Host != "" {
			tcpSocket.Host = pulumi.String(handler.TcpSocket.Host)
		}

		probe.TcpSocket = tcpSocket

	case *kubernetesv1.Probe_Exec:
		if len(handler.Exec.Command) > 0 {
			probe.Exec = &kubernetescorev1.ExecActionArgs{
				Command: pulumi.ToStringArray(handler.Exec.Command),
			}
		}
	}

	return probe
}

// buildDeploymentStrategy converts a proto DeploymentStrategy into Pulumi DeploymentStrategyArgs.
// Returns nil if no strategy is configured (uses Kubernetes defaults).
func buildDeploymentStrategy(protoStrategy *kubernetesdeploymentv1.KubernetesDeploymentDeploymentStrategy) *appsv1.DeploymentStrategyArgs {
	if protoStrategy == nil {
		return nil
	}

	rollingUpdate := &appsv1.RollingUpdateDeploymentArgs{}

	// Set maxUnavailable if configured
	if protoStrategy.MaxUnavailable != "" {
		rollingUpdate.MaxUnavailable = parseIntOrString(protoStrategy.MaxUnavailable)
	}

	// Set maxSurge if configured
	if protoStrategy.MaxSurge != "" {
		rollingUpdate.MaxSurge = parseIntOrString(protoStrategy.MaxSurge)
	}

	strategy := &appsv1.DeploymentStrategyArgs{
		Type:          pulumi.String("RollingUpdate"),
		RollingUpdate: rollingUpdate,
	}

	return strategy
}

// buildVolumeMountsAndVolumes processes the volume_mounts spec and returns
// Pulumi volume mounts for the container and volumes for the pod spec.
func buildVolumeMountsAndVolumes(
	spec *kubernetesdeploymentv1.KubernetesDeploymentSpec,
) (kubernetescorev1.VolumeMountArray, kubernetescorev1.VolumeArray) {
	volumeMounts := make(kubernetescorev1.VolumeMountArray, 0)
	volumes := make(kubernetescorev1.VolumeArray, 0)

	if spec.Container == nil || spec.Container.App == nil || spec.Container.App.VolumeMounts == nil {
		return volumeMounts, volumes
	}

	for _, vm := range spec.Container.App.VolumeMounts {
		// Add volume mount to container
		volumeMountArgs := &kubernetescorev1.VolumeMountArgs{
			Name:      pulumi.String(vm.Name),
			MountPath: pulumi.String(vm.MountPath),
			ReadOnly:  pulumi.Bool(vm.ReadOnly),
		}
		if vm.SubPath != "" {
			volumeMountArgs.SubPath = pulumi.String(vm.SubPath)
		}
		volumeMounts = append(volumeMounts, volumeMountArgs)

		// Add corresponding volume to pod spec
		volumeArgs := &kubernetescorev1.VolumeArgs{
			Name: pulumi.String(vm.Name),
		}

		// Determine volume type based on which source is set
		switch {
		case vm.ConfigMap != nil:
			configMapVolumeSource := &kubernetescorev1.ConfigMapVolumeSourceArgs{
				Name: pulumi.String(vm.ConfigMap.Name),
			}
			if vm.ConfigMap.Key != "" {
				path := vm.ConfigMap.Path
				if path == "" {
					path = vm.ConfigMap.Key
				}
				configMapVolumeSource.Items = kubernetescorev1.KeyToPathArray{
					&kubernetescorev1.KeyToPathArgs{
						Key:  pulumi.String(vm.ConfigMap.Key),
						Path: pulumi.String(path),
					},
				}
			}
			if vm.ConfigMap.DefaultMode > 0 {
				configMapVolumeSource.DefaultMode = pulumi.Int(vm.ConfigMap.DefaultMode)
			}
			volumeArgs.ConfigMap = configMapVolumeSource

		case vm.Secret != nil:
			secretVolumeSource := &kubernetescorev1.SecretVolumeSourceArgs{
				SecretName: pulumi.String(vm.Secret.Name),
			}
			if vm.Secret.Key != "" {
				path := vm.Secret.Path
				if path == "" {
					path = vm.Secret.Key
				}
				secretVolumeSource.Items = kubernetescorev1.KeyToPathArray{
					&kubernetescorev1.KeyToPathArgs{
						Key:  pulumi.String(vm.Secret.Key),
						Path: pulumi.String(path),
					},
				}
			}
			if vm.Secret.DefaultMode > 0 {
				secretVolumeSource.DefaultMode = pulumi.Int(vm.Secret.DefaultMode)
			}
			volumeArgs.Secret = secretVolumeSource

		case vm.HostPath != nil:
			hostPathVolumeSource := &kubernetescorev1.HostPathVolumeSourceArgs{
				Path: pulumi.String(vm.HostPath.Path),
			}
			if vm.HostPath.Type != "" {
				hostPathVolumeSource.Type = pulumi.StringPtr(vm.HostPath.Type)
			}
			volumeArgs.HostPath = hostPathVolumeSource

		case vm.EmptyDir != nil:
			emptyDirVolumeSource := &kubernetescorev1.EmptyDirVolumeSourceArgs{}
			if vm.EmptyDir.Medium != "" {
				emptyDirVolumeSource.Medium = pulumi.String(vm.EmptyDir.Medium)
			}
			if vm.EmptyDir.SizeLimit != "" {
				emptyDirVolumeSource.SizeLimit = pulumi.String(vm.EmptyDir.SizeLimit)
			}
			volumeArgs.EmptyDir = emptyDirVolumeSource

		case vm.Pvc != nil:
			volumeArgs.PersistentVolumeClaim = &kubernetescorev1.PersistentVolumeClaimVolumeSourceArgs{
				ClaimName: pulumi.String(vm.Pvc.ClaimName),
				ReadOnly:  pulumi.Bool(vm.Pvc.ReadOnly),
			}
		}

		volumes = append(volumes, volumeArgs)
	}

	return volumeMounts, volumes
}

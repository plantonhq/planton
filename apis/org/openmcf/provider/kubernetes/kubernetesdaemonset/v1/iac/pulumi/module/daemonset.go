package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	kubernetesdaemonsetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesdaemonset/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerenv"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func daemonSet(ctx *pulumi.Context, locals *Locals, serviceAccountName string, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	target := locals.KubernetesDaemonSet

	envVarInputs := containerenv.BuildEnvVars(target.Spec.Container.App.Env, locals.EnvSecretName)

	// DaemonSets additionally inject the node name for node-level operations.
	envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
		Name: pulumi.String("K8S_NODE_NAME"),
		ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
			FieldRef: &kubernetescorev1.ObjectFieldSelectorArgs{
				FieldPath: pulumi.String("spec.nodeName"),
			},
		},
	}))

	envFromInputs := containerenv.BuildEnvFrom(target.Spec.Container.App.Env)

	// Build ports array
	portsArray := make(kubernetescorev1.ContainerPortArray, 0)
	for _, p := range target.Spec.Container.App.Ports {
		portArgs := &kubernetescorev1.ContainerPortArgs{
			Name:          pulumi.String(p.Name),
			ContainerPort: pulumi.Int(p.ContainerPort),
			Protocol:      pulumi.String(p.NetworkProtocol),
		}
		if p.HostPort > 0 {
			portArgs.HostPort = pulumi.Int(p.HostPort)
		}
		portsArray = append(portsArray, portArgs)
	}

	// Build volume mounts and volumes using the shared VolumeMount type
	volumeMounts, volumes := buildVolumeMountsAndVolumes(target.Spec.Container.App.VolumeMounts)

	// Build main container
	mainContainer := &kubernetescorev1.ContainerArgs{
		Name: pulumi.String("daemonset-container"),
		Image: pulumi.String(fmt.Sprintf("%s:%s",
			target.Spec.Container.App.Image.Repo,
			target.Spec.Container.App.Image.Tag)),
		Env:          kubernetescorev1.EnvVarArray(envVarInputs),
		EnvFrom:      envFromInputs,
		Ports:        portsArray,
		VolumeMounts: volumeMounts,
		Resources: kubernetescorev1.ResourceRequirementsArgs{
			Limits: pulumi.ToStringMap(map[string]string{
				"cpu":    target.Spec.Container.App.Resources.Limits.Cpu,
				"memory": target.Spec.Container.App.Resources.Limits.Memory,
			}),
			Requests: pulumi.ToStringMap(map[string]string{
				"cpu":    target.Spec.Container.App.Resources.Requests.Cpu,
				"memory": target.Spec.Container.App.Resources.Requests.Memory,
			}),
		},
		LivenessProbe:  buildProbe(target.Spec.Container.App.LivenessProbe),
		ReadinessProbe: buildProbe(target.Spec.Container.App.ReadinessProbe),
		StartupProbe:   buildProbe(target.Spec.Container.App.StartupProbe),
	}

	// Add command if specified
	if len(target.Spec.Container.App.Command) > 0 {
		mainContainer.Command = pulumi.ToStringArray(target.Spec.Container.App.Command)
	}

	// Add args if specified
	if len(target.Spec.Container.App.Args) > 0 {
		mainContainer.Args = pulumi.ToStringArray(target.Spec.Container.App.Args)
	}

	// Add security context if specified
	if target.Spec.Container.App.SecurityContext != nil {
		mainContainer.SecurityContext = buildSecurityContext(target.Spec.Container.App.SecurityContext)
	}

	// Build pod spec
	podSpecArgs := &kubernetescorev1.PodSpecArgs{
		Containers: kubernetescorev1.ContainerArray{mainContainer},
		Volumes:    volumes,
	}
	if serviceAccountName != "" {
		podSpecArgs.ServiceAccountName = pulumi.String(serviceAccountName)
	}

	// Add node selector if specified
	if len(target.Spec.NodeSelector) > 0 {
		podSpecArgs.NodeSelector = pulumi.ToStringMap(target.Spec.NodeSelector)
	}

	// Add tolerations if specified
	if len(target.Spec.Tolerations) > 0 {
		tolerations := make(kubernetescorev1.TolerationArray, 0)
		for _, t := range target.Spec.Tolerations {
			toleration := &kubernetescorev1.TolerationArgs{}
			if t.Key != "" {
				toleration.Key = pulumi.StringPtr(t.Key)
			}
			if t.Operator != "" {
				toleration.Operator = pulumi.StringPtr(t.Operator)
			}
			if t.Value != "" {
				toleration.Value = pulumi.StringPtr(t.Value)
			}
			if t.Effect != "" {
				toleration.Effect = pulumi.StringPtr(t.Effect)
			}
			if t.TolerationSeconds > 0 {
				toleration.TolerationSeconds = pulumi.IntPtr(int(t.TolerationSeconds))
			}
			tolerations = append(tolerations, toleration)
		}
		podSpecArgs.Tolerations = tolerations
	}

	// Add image pull secrets if configured
	if locals.ImagePullSecretData != nil {
		podSpecArgs.ImagePullSecrets = kubernetescorev1.LocalObjectReferenceArray{
			kubernetescorev1.LocalObjectReferenceArgs{
				Name: pulumi.String(locals.ImagePullSecretName),
			},
		}
	}

	// Build DaemonSet spec
	daemonSetSpec := &appsv1.DaemonSetSpecArgs{
		Selector: &metav1.LabelSelectorArgs{
			MatchLabels: pulumi.ToStringMap(locals.SelectorLabels),
		},
		Template: &kubernetescorev1.PodTemplateSpecArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Labels: pulumi.ToStringMap(locals.Labels),
			},
			Spec: podSpecArgs,
		},
	}

	// Add min ready seconds if specified
	if target.Spec.MinReadySeconds > 0 {
		daemonSetSpec.MinReadySeconds = pulumi.Int(target.Spec.MinReadySeconds)
	}

	// Add update strategy if specified
	if target.Spec.UpdateStrategy != nil {
		updateStrategy := buildUpdateStrategy(target.Spec.UpdateStrategy)
		if updateStrategy != nil {
			daemonSetSpec.UpdateStrategy = updateStrategy
		}
	}

	// Create the DaemonSet
	dsOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := appsv1.NewDaemonSet(ctx,
		target.Metadata.Name,
		&appsv1.DaemonSetArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(target.Metadata.Name),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: daemonSetSpec,
		},
		dsOpts...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create daemonset")
	}

	return nil
}

// buildProbe converts a proto Probe definition into a Pulumi Kubernetes ProbeArgs.
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

// buildSecurityContext converts the proto security context to Pulumi SecurityContextArgs.
func buildSecurityContext(sc *kubernetesdaemonsetv1.KubernetesDaemonSetSecurityContext) *kubernetescorev1.SecurityContextArgs {
	if sc == nil {
		return nil
	}

	securityContext := &kubernetescorev1.SecurityContextArgs{}

	if sc.Privileged {
		securityContext.Privileged = pulumi.Bool(sc.Privileged)
	}

	if sc.RunAsUser > 0 {
		securityContext.RunAsUser = pulumi.IntPtr(int(sc.RunAsUser))
	}

	if sc.RunAsGroup > 0 {
		securityContext.RunAsGroup = pulumi.IntPtr(int(sc.RunAsGroup))
	}

	if sc.RunAsNonRoot {
		securityContext.RunAsNonRoot = pulumi.Bool(sc.RunAsNonRoot)
	}

	if sc.ReadOnlyRootFilesystem {
		securityContext.ReadOnlyRootFilesystem = pulumi.Bool(sc.ReadOnlyRootFilesystem)
	}

	if sc.Capabilities != nil && (len(sc.Capabilities.Add) > 0 || len(sc.Capabilities.Drop) > 0) {
		capabilities := &kubernetescorev1.CapabilitiesArgs{}
		if len(sc.Capabilities.Add) > 0 {
			capabilities.Add = pulumi.ToStringArray(sc.Capabilities.Add)
		}
		if len(sc.Capabilities.Drop) > 0 {
			capabilities.Drop = pulumi.ToStringArray(sc.Capabilities.Drop)
		}
		securityContext.Capabilities = capabilities
	}

	return securityContext
}

// buildUpdateStrategy converts the proto update strategy to Pulumi DaemonSetUpdateStrategyArgs.
func buildUpdateStrategy(us *kubernetesdaemonsetv1.KubernetesDaemonSetUpdateStrategy) *appsv1.DaemonSetUpdateStrategyArgs {
	if us == nil || us.Type == "" {
		return nil
	}

	strategy := &appsv1.DaemonSetUpdateStrategyArgs{
		Type: pulumi.String(us.Type),
	}

	if us.Type == "RollingUpdate" && us.RollingUpdate != nil {
		rollingUpdate := &appsv1.RollingUpdateDaemonSetArgs{}

		if us.RollingUpdate.MaxUnavailable != "" {
			rollingUpdate.MaxUnavailable = pulumi.Any(us.RollingUpdate.MaxUnavailable)
		}

		if us.RollingUpdate.MaxSurge != "" {
			rollingUpdate.MaxSurge = pulumi.Any(us.RollingUpdate.MaxSurge)
		}

		strategy.RollingUpdate = rollingUpdate
	}

	return strategy
}

// buildVolumeMountsAndVolumes processes the volume_mounts spec and returns
// Pulumi volume mounts for the container and volumes for the pod spec.
func buildVolumeMountsAndVolumes(
	volumeMountSpecs []*kubernetesv1.VolumeMount,
) (kubernetescorev1.VolumeMountArray, kubernetescorev1.VolumeArray) {
	volumeMounts := make(kubernetescorev1.VolumeMountArray, 0)
	volumes := make(kubernetescorev1.VolumeArray, 0)

	if volumeMountSpecs == nil {
		return volumeMounts, volumes
	}

	for _, vm := range volumeMountSpecs {
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

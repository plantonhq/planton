package module

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	kubernetesv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	kubernetesstatefulsetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func statefulSet(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	headlessService *kubernetescorev1.Service, namespaceDeps []pulumi.ResourceOption) (*appsv1.StatefulSet, error) {

	target := locals.KubernetesStatefulSet

	// Create service account
	serviceAccountArgs := &kubernetescorev1.ServiceAccountArgs{
		Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
			Name:      pulumi.String(target.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
		}),
	}

	createdServiceAccount, err := kubernetescorev1.NewServiceAccount(ctx,
		target.Metadata.Name,
		serviceAccountArgs,
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to add service account")
	}

	envVarInputs := make([]kubernetescorev1.EnvVarInput, 0)

	// Add HOSTNAME env var
	envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
		Name: pulumi.String("HOSTNAME"),
		ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
			FieldRef: &kubernetescorev1.ObjectFieldSelectorArgs{
				FieldPath: pulumi.String("status.podIP"),
			},
		},
	}))

	// Add K8S_POD_ID env var
	envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
		Name: pulumi.String("K8S_POD_ID"),
		ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
			FieldRef: &kubernetescorev1.ObjectFieldSelectorArgs{
				ApiVersion: pulumi.String("v1"),
				FieldPath:  pulumi.String("metadata.name"),
			},
		},
	}))

	// Add K8S_POD_NAMESPACE env var
	envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
		Name: pulumi.String("K8S_POD_NAMESPACE"),
		ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
			FieldRef: &kubernetescorev1.ObjectFieldSelectorArgs{
				ApiVersion: pulumi.String("v1"),
				FieldPath:  pulumi.String("metadata.namespace"),
			},
		},
	}))

	if target.Spec.Container.App.Env != nil {
		if target.Spec.Container.App.Env.Variables != nil {
			// Sort keys for deterministic output
			sortedEnvVariableKeys := make([]string, 0, len(target.Spec.Container.App.Env.Variables))
			for k := range target.Spec.Container.App.Env.Variables {
				sortedEnvVariableKeys = append(sortedEnvVariableKeys, k)
			}
			sort.Strings(sortedEnvVariableKeys)

			for _, envVarKey := range sortedEnvVariableKeys {
				envVarValue := target.Spec.Container.App.Env.Variables[envVarKey]
				// Orchestrator resolves valueFrom and places result in .value
				if envVarValue.GetValue() != "" {
					envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
						Name:  pulumi.String(envVarKey),
						Value: pulumi.String(envVarValue.GetValue()),
					}))
				}
			}
		}

		if target.Spec.Container.App.Env.Secrets != nil {
			// Sort keys for deterministic output
			sortedSecretKeys := make([]string, 0, len(target.Spec.Container.App.Env.Secrets))
			for k := range target.Spec.Container.App.Env.Secrets {
				sortedSecretKeys = append(sortedSecretKeys, k)
			}
			sort.Strings(sortedSecretKeys)

			for _, secretKey := range sortedSecretKeys {
				secretValue := target.Spec.Container.App.Env.Secrets[secretKey]

				if secretValue.GetSecretRef() != nil {
					// Use external Kubernetes Secret reference
					secretRef := secretValue.GetSecretRef()
					envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
						Name: pulumi.String(secretKey),
						ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
							SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
								Name: pulumi.String(secretRef.Name),
								Key:  pulumi.String(secretRef.Key),
							},
						},
					}))
				} else if secretValue.GetValue() != "" {
					// Use the internally created secret for direct string values
					envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
						Name: pulumi.String(secretKey),
						ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
							SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
								Name: pulumi.String(locals.EnvSecretName),
								Key:  pulumi.String(secretKey),
							},
						},
					}))
				}
			}
		}
	}

	portsArray := make(kubernetescorev1.ContainerPortArray, 0)
	for _, p := range target.Spec.Container.App.Ports {
		portsArray = append(portsArray, &kubernetescorev1.ContainerPortArgs{
			Name:          pulumi.String(p.Name),
			ContainerPort: pulumi.Int(p.ContainerPort),
		})
	}

	// Build volume mounts and volumes from spec
	// This handles ConfigMaps, Secrets, HostPaths, EmptyDirs, and PVCs
	// For PVCs that reference volumeClaimTemplates, we skip creating separate volumes
	// as StatefulSet handles these automatically
	volumeMountsArray, additionalVolumes := buildVolumeMountsAndVolumes(target.Spec)

	containerInputs := make([]kubernetescorev1.ContainerInput, 0)

	mainContainer := &kubernetescorev1.ContainerArgs{
		Name: pulumi.String("app"),
		Image: pulumi.String(fmt.Sprintf("%s:%s",
			target.Spec.Container.App.Image.Repo,
			target.Spec.Container.App.Image.Tag)),
		Env:   kubernetescorev1.EnvVarArray(envVarInputs),
		Ports: portsArray,
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
		VolumeMounts:   volumeMountsArray,
	}

	// Add command if specified
	if len(target.Spec.Container.App.Command) > 0 {
		mainContainer.Command = pulumi.ToStringArray(target.Spec.Container.App.Command)
	}

	// Add args if specified
	if len(target.Spec.Container.App.Args) > 0 {
		mainContainer.Args = pulumi.ToStringArray(target.Spec.Container.App.Args)
	}

	containerInputs = append(containerInputs, kubernetescorev1.ContainerInput(mainContainer))

	podSpecArgs := &kubernetescorev1.PodSpecArgs{
		ServiceAccountName: createdServiceAccount.Metadata.Name(),
		Containers:         kubernetescorev1.ContainerArray(containerInputs),
		Volumes:            additionalVolumes,
		// Wait for 60 seconds before sending the termination signal
		TerminationGracePeriodSeconds: pulumi.IntPtr(60),
	}

	// Create image pull secret if configured
	createdImagePullSecret, err := imagePullSecret(ctx, locals, kubernetesProvider)
	if err != nil {
		return nil, err
	}
	if createdImagePullSecret != nil {
		podSpecArgs.ImagePullSecrets = kubernetescorev1.LocalObjectReferenceArray{
			kubernetescorev1.LocalObjectReferenceArgs{
				Name: createdImagePullSecret.Metadata.Name(),
			}}
	}

	// Build volume claim templates
	volumeClaimTemplates := make(kubernetescorev1.PersistentVolumeClaimTypeArray, 0)
	for _, vct := range target.Spec.VolumeClaimTemplates {
		accessModes := pulumi.StringArray{}
		if len(vct.AccessModes) == 0 {
			accessModes = append(accessModes, pulumi.String("ReadWriteOnce"))
		} else {
			for _, am := range vct.AccessModes {
				accessModes = append(accessModes, pulumi.String(am))
			}
		}

		pvcSpec := &kubernetescorev1.PersistentVolumeClaimSpecArgs{
			AccessModes: accessModes,
			Resources: &kubernetescorev1.VolumeResourceRequirementsArgs{
				Requests: pulumi.StringMap{
					"storage": pulumi.String(vct.Size),
				},
			},
		}

		if vct.StorageClass != "" {
			pvcSpec.StorageClassName = pulumi.String(vct.StorageClass)
		}

		volumeClaimTemplates = append(volumeClaimTemplates, &kubernetescorev1.PersistentVolumeClaimTypeArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String(vct.Name),
			},
			Spec: pvcSpec,
		})
	}

	// Determine replicas
	replicas := int(1)
	if target.Spec.Availability != nil && target.Spec.Availability.Replicas > 0 {
		replicas = int(target.Spec.Availability.Replicas)
	}

	// Determine pod management policy
	podManagementPolicy := "OrderedReady"
	if target.Spec.PodManagementPolicy != "" {
		podManagementPolicy = target.Spec.PodManagementPolicy
	}

	statefulSetArgs := &appsv1.StatefulSetArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(target.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
			Annotations: pulumi.StringMap{
				"pulumi.com/patchForce": pulumi.String("true"),
			},
		},
		Spec: &appsv1.StatefulSetSpecArgs{
			ServiceName:          pulumi.String(locals.HeadlessServiceName),
			Replicas:             pulumi.Int(replicas),
			PodManagementPolicy:  pulumi.String(podManagementPolicy),
			VolumeClaimTemplates: volumeClaimTemplates,
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

	ssOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{headlessService}),
	}, namespaceDeps...)
	createdStatefulSet, err := appsv1.NewStatefulSet(ctx,
		target.Metadata.Name,
		statefulSetArgs,
		ssOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add statefulset")
	}

	return createdStatefulSet, nil
}

// buildProbe converts a proto Probe definition into a Pulumi Kubernetes ProbeArgs.
func buildProbe(protoProbe *kubernetesv1.Probe) *kubernetescorev1.ProbeArgs {
	if protoProbe == nil {
		return nil
	}

	probe := &kubernetescorev1.ProbeArgs{}

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

// isVolumeClaimTemplate checks if a PVC name matches a volumeClaimTemplate name.
// StatefulSets handle volumeClaimTemplate PVCs automatically, so we should not
// create separate volumes for these.
func isVolumeClaimTemplate(pvcName string, templates []*kubernetesstatefulsetv1.KubernetesStatefulSetVolumeClaimTemplate) bool {
	for _, t := range templates {
		if t.Name == pvcName {
			return true
		}
	}
	return false
}

// buildVolumeMountsAndVolumes processes the volume_mounts spec and returns
// Pulumi volume mounts for the container and volumes for the pod spec.
// For StatefulSets, PVC mounts that reference volumeClaimTemplates are handled
// specially - the volume mount is created but no separate volume is added
// (StatefulSet manages these automatically).
func buildVolumeMountsAndVolumes(
	spec *kubernetesstatefulsetv1.KubernetesStatefulSetSpec,
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
			volumes = append(volumes, volumeArgs)

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
			volumes = append(volumes, volumeArgs)

		case vm.HostPath != nil:
			hostPathVolumeSource := &kubernetescorev1.HostPathVolumeSourceArgs{
				Path: pulumi.String(vm.HostPath.Path),
			}
			if vm.HostPath.Type != "" {
				hostPathVolumeSource.Type = pulumi.StringPtr(vm.HostPath.Type)
			}
			volumeArgs.HostPath = hostPathVolumeSource
			volumes = append(volumes, volumeArgs)

		case vm.EmptyDir != nil:
			emptyDirVolumeSource := &kubernetescorev1.EmptyDirVolumeSourceArgs{}
			if vm.EmptyDir.Medium != "" {
				emptyDirVolumeSource.Medium = pulumi.String(vm.EmptyDir.Medium)
			}
			if vm.EmptyDir.SizeLimit != "" {
				emptyDirVolumeSource.SizeLimit = pulumi.String(vm.EmptyDir.SizeLimit)
			}
			volumeArgs.EmptyDir = emptyDirVolumeSource
			volumes = append(volumes, volumeArgs)

		case vm.Pvc != nil:
			// For StatefulSets, if the PVC references a volumeClaimTemplate,
			// we only create the volume mount - the StatefulSet controller
			// handles the actual volume binding automatically
			if !isVolumeClaimTemplate(vm.Pvc.ClaimName, spec.VolumeClaimTemplates) {
				// This is an external PVC, not a volumeClaimTemplate
				volumeArgs.PersistentVolumeClaim = &kubernetescorev1.PersistentVolumeClaimVolumeSourceArgs{
					ClaimName: pulumi.String(vm.Pvc.ClaimName),
					ReadOnly:  pulumi.Bool(vm.Pvc.ReadOnly),
				}
				volumes = append(volumes, volumeArgs)
			}
			// If it IS a volumeClaimTemplate, we don't add a volume - StatefulSet handles it
		}
	}

	return volumeMounts, volumes
}

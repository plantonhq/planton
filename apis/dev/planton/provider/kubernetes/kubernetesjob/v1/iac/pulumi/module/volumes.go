package module

import (
	kubernetesv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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

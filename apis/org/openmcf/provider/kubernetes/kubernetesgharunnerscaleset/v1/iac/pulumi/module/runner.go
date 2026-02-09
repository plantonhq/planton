package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ghaRunnerScaleSet deploys the GitHub Actions Runner Scale Set
// using the official Helm chart.
func ghaRunnerScaleSet(ctx *pulumi.Context, locals *Locals, k8sProvider *kubernetes.Provider, namespaceDeps []pulumi.ResourceOption) error {
	// Create PVCs for persistent volumes
	pvcResources, err := createPersistentVolumes(ctx, locals, k8sProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "create persistent volumes")
	}

	// Build Helm values
	helmValues := buildHelmValues(locals)

	// Deploy Helm chart
	// For OCI charts, the full URL must be passed as the Chart parameter
	// (RepositoryOpts.Repo doesn't work with OCI registries in Pulumi)
	helmOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(k8sProvider),
		pulumi.DependsOn(pvcResources),
	}, namespaceDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, &helmv3.ReleaseArgs{
		Name:            pulumi.String(locals.ReleaseName),
		Namespace:       pulumi.String(locals.Namespace),
		CreateNamespace: pulumi.Bool(false), // We handle namespace creation ourselves
		Chart:           pulumi.String(HelmChartOCI),
		Version:         pulumi.String(locals.ChartVersion),
		Values:          helmValues,
	}, helmOpts...)
	if err != nil {
		return errors.Wrap(err, "deploy helm release")
	}

	return nil
}

// createPersistentVolumes creates PVCs for persistent storage.
func createPersistentVolumes(ctx *pulumi.Context, locals *Locals, k8sProvider *kubernetes.Provider, namespaceDeps []pulumi.ResourceOption) ([]pulumi.Resource, error) {
	var pvcs []pulumi.Resource

	for _, pv := range locals.PersistentVolumes {
		pvcName := fmt.Sprintf("%s-%s", locals.ReleaseName, pv.Name)

		accessModes := pv.AccessModes
		if len(accessModes) == 0 {
			accessModes = []string{"ReadWriteOnce"}
		}

		// Build spec with optional storage class
		spec := &corev1.PersistentVolumeClaimSpecArgs{
			AccessModes: pulumi.ToStringArray(accessModes),
			Resources: &corev1.VolumeResourceRequirementsArgs{
				Requests: pulumi.StringMap{
					"storage": pulumi.String(pv.Size),
				},
			},
		}
		if pv.StorageClass != "" {
			spec.StorageClassName = pulumi.String(pv.StorageClass)
		}

		pvcArgs := &corev1.PersistentVolumeClaimArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(pvcName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.KubeLabels),
			},
			Spec: spec,
		}

		pvcOpts := append([]pulumi.ResourceOption{
			pulumi.Provider(k8sProvider),
		}, namespaceDeps...)

		pvc, err := corev1.NewPersistentVolumeClaim(ctx, pvcName, pvcArgs, pvcOpts...)
		if err != nil {
			return nil, errors.Wrapf(err, "create PVC %s", pvcName)
		}
		pvcs = append(pvcs, pvc)
	}

	return pvcs, nil
}

// buildHelmValues constructs the Helm values map from the Locals struct.
func buildHelmValues(locals *Locals) pulumi.Map {
	values := pulumi.Map{
		"githubConfigUrl": pulumi.String(locals.GitHubConfigURL),
	}

	// GitHub secret configuration
	if locals.UseExistingSecret {
		values["githubConfigSecret"] = pulumi.String(locals.GitHubSecretName)
	} else if locals.PatToken != "" {
		values["githubConfigSecret"] = pulumi.Map{
			"github_token": pulumi.String(locals.PatToken),
		}
	} else if locals.GitHubAppID != "" {
		values["githubConfigSecret"] = pulumi.Map{
			"github_app_id":              pulumi.String(locals.GitHubAppID),
			"github_app_installation_id": pulumi.String(locals.GitHubAppInstallID),
			"github_app_private_key":     pulumi.String(locals.GitHubAppPrivateKey),
		}
	}

	// Scaling configuration
	values["minRunners"] = pulumi.Int(locals.MinRunners)
	values["maxRunners"] = pulumi.Int(locals.MaxRunners)

	// Runner group and name
	if locals.RunnerGroup != "" {
		values["runnerGroup"] = pulumi.String(locals.RunnerGroup)
	}
	if locals.RunnerScaleSetName != "" {
		values["runnerScaleSetName"] = pulumi.String(locals.RunnerScaleSetName)
	}

	// Container mode
	if locals.ContainerModeType != "" {
		containerMode := pulumi.Map{
			"type": pulumi.String(locals.ContainerModeType),
		}
		if locals.ContainerModeType == "kubernetes" && locals.WorkVolumeClaimSize != "" {
			workVolumeClaim := pulumi.Map{
				"accessModes": pulumi.ToStringArray(locals.WorkVolumeClaimAccessModes),
				"resources": pulumi.Map{
					"requests": pulumi.Map{
						"storage": pulumi.String(locals.WorkVolumeClaimSize),
					},
				},
			}
			if locals.WorkVolumeClaimStorageClass != "" {
				workVolumeClaim["storageClassName"] = pulumi.String(locals.WorkVolumeClaimStorageClass)
			}
			containerMode["kubernetesModeWorkVolumeClaim"] = workVolumeClaim
		}
		values["containerMode"] = containerMode
	}

	// Template spec for runner container
	template := buildTemplateSpec(locals)
	if len(template) > 0 {
		values["template"] = pulumi.Map{
			"spec": template,
		}
	}

	// Controller service account
	if locals.ControllerServiceAccountName != "" || locals.ControllerServiceAccountNamespace != "" {
		controllerSA := pulumi.Map{}
		if locals.ControllerServiceAccountName != "" {
			controllerSA["name"] = pulumi.String(locals.ControllerServiceAccountName)
		}
		if locals.ControllerServiceAccountNamespace != "" {
			controllerSA["namespace"] = pulumi.String(locals.ControllerServiceAccountNamespace)
		}
		values["controllerServiceAccount"] = controllerSA
	}

	// Labels and annotations
	if len(locals.Labels) > 0 {
		values["labels"] = pulumi.ToStringMap(locals.Labels)
	}
	if len(locals.Annotations) > 0 {
		values["annotations"] = pulumi.ToStringMap(locals.Annotations)
	}

	return values
}

// buildTemplateSpec builds the template.spec section for the runner pod.
func buildTemplateSpec(locals *Locals) pulumi.Map {
	spec := pulumi.Map{}

	// Build container spec
	container := pulumi.Map{
		"name": pulumi.String("runner"),
	}

	// Image configuration
	if locals.RunnerImageRepository != "" {
		image := locals.RunnerImageRepository
		if locals.RunnerImageTag != "" {
			image = image + ":" + locals.RunnerImageTag
		}
		container["image"] = pulumi.String(image)
	}
	if locals.RunnerImagePullPolicy != "" {
		container["imagePullPolicy"] = pulumi.String(locals.RunnerImagePullPolicy)
	}

	// Command (required for runner container)
	container["command"] = pulumi.ToStringArray([]string{"/home/runner/run.sh"})

	// Resources
	resources := pulumi.Map{}
	if locals.RunnerCpuRequests != "" || locals.RunnerMemoryRequests != "" {
		requests := pulumi.Map{}
		if locals.RunnerCpuRequests != "" {
			requests["cpu"] = pulumi.String(locals.RunnerCpuRequests)
		}
		if locals.RunnerMemoryRequests != "" {
			requests["memory"] = pulumi.String(locals.RunnerMemoryRequests)
		}
		resources["requests"] = requests
	}
	if locals.RunnerCpuLimits != "" || locals.RunnerMemoryLimits != "" {
		limits := pulumi.Map{}
		if locals.RunnerCpuLimits != "" {
			limits["cpu"] = pulumi.String(locals.RunnerCpuLimits)
		}
		if locals.RunnerMemoryLimits != "" {
			limits["memory"] = pulumi.String(locals.RunnerMemoryLimits)
		}
		resources["limits"] = limits
	}
	if len(resources) > 0 {
		container["resources"] = resources
	}

	// Environment variables
	if len(locals.RunnerEnvVars) > 0 {
		envArray := pulumi.Array{}
		for name, value := range locals.RunnerEnvVars {
			envArray = append(envArray, pulumi.Map{
				"name":  pulumi.String(name),
				"value": pulumi.String(value),
			})
		}
		container["env"] = envArray
	}

	// Volume mounts for persistent volumes
	if len(locals.PersistentVolumes) > 0 || len(locals.RunnerVolumeMounts) > 0 {
		mountArray := pulumi.Array{}

		// Add persistent volume mounts
		for _, pv := range locals.PersistentVolumes {
			mount := pulumi.Map{
				"name":      pulumi.String(pv.Name),
				"mountPath": pulumi.String(pv.MountPath),
			}
			if pv.ReadOnly {
				mount["readOnly"] = pulumi.Bool(true)
			}
			mountArray = append(mountArray, mount)
		}

		// Add additional volume mounts
		for _, vm := range locals.RunnerVolumeMounts {
			mount := pulumi.Map{
				"name":      pulumi.String(vm.Name),
				"mountPath": pulumi.String(vm.MountPath),
			}
			if vm.ReadOnly {
				mount["readOnly"] = pulumi.Bool(true)
			}
			if vm.SubPath != "" {
				mount["subPath"] = pulumi.String(vm.SubPath)
			}
			mountArray = append(mountArray, mount)
		}

		container["volumeMounts"] = mountArray
	}

	// Only add container if we have customizations beyond defaults
	if len(container) > 2 { // more than just name and command
		spec["containers"] = pulumi.Array{container}
	}

	// Volumes for persistent volumes
	if len(locals.PersistentVolumes) > 0 {
		volumeArray := pulumi.Array{}
		for _, pv := range locals.PersistentVolumes {
			pvcName := fmt.Sprintf("%s-%s", locals.ReleaseName, pv.Name)
			volume := pulumi.Map{
				"name": pulumi.String(pv.Name),
				"persistentVolumeClaim": pulumi.Map{
					"claimName": pulumi.String(pvcName),
				},
			}
			volumeArray = append(volumeArray, volume)
		}
		spec["volumes"] = volumeArray
	}

	// Image pull secrets
	if len(locals.ImagePullSecrets) > 0 {
		secrets := pulumi.Array{}
		for _, secret := range locals.ImagePullSecrets {
			secrets = append(secrets, pulumi.Map{"name": pulumi.String(secret)})
		}
		spec["imagePullSecrets"] = secrets
	}

	return spec
}

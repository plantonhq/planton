package module

import (
	"github.com/pkg/errors"
	kubernetesv1 "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesdeploymentv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesdeployment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadpod"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// deployment renders the apps/v1 Deployment. The pod template is assembled
// through the shared workloadpod builders so container and pod semantics match
// every other workload kind exactly; only Deployment-specific mechanics
// (replicas, rollout strategy, progress tracking) live here.
func deployment(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, createdImagePullSecret *kubernetescorev1.Secret,
	namespaceDeps []pulumi.ResourceOption) (*appsv1.Deployment, error) {

	spec := locals.KubernetesDeployment.Spec

	containers := workloadpod.BuildContainers(
		spec.Container.App, spec.Container.Sidecars,
		// Default name for the main container when the spec omits one. "app" matches
		// the selector-label convention and reads correctly in kubectl output.
		"app",
		locals.EnvSecretName,
	)

	// The pod volume list is the union of every container's mounts — app,
	// sidecars, and init containers all contribute.
	volumeSourceContainers := make([]*kubernetesv1.WorkloadContainer, 0, 2+len(spec.Container.Sidecars))
	volumeSourceContainers = append(volumeSourceContainers, spec.Container.App)
	volumeSourceContainers = append(volumeSourceContainers, spec.Container.Sidecars...)
	if spec.Pod != nil {
		volumeSourceContainers = append(volumeSourceContainers, spec.Pod.InitContainers...)
	}
	volumes := workloadpod.BuildVolumes(volumeSourceContainers...)

	// The module-created docker-config secret joins the pod's pull-secret list;
	// spec-listed pull secrets pass through by name.
	moduleCreatedPullSecret := ""
	if createdImagePullSecret != nil {
		moduleCreatedPullSecret = locals.ImagePullSecretName
	}

	podTemplate := workloadpod.BuildPodTemplateSpec(spec.Pod, workloadpod.PodTemplateInputs{
		Labels:               locals.Labels,
		Containers:           containers,
		Volumes:              volumes,
		ImagePullSecretNames: workloadpod.ResolveImagePullSecretNames(spec.Pod, moduleCreatedPullSecret),
	}, locals.EnvSecretName)

	deploymentSpecArgs := &appsv1.DeploymentSpecArgs{
		Selector: &metav1.LabelSelectorArgs{
			MatchLabels: pulumi.ToStringMap(locals.SelectorLabels),
		},
		Template: podTemplate,
		Strategy: buildDeploymentStrategy(getStrategy(spec.Availability)),
	}

	// Replica count: the availability block's value, defaulting to 1. While
	// autoscaling is on, the HPA owns the live count and `replicas` is only its
	// floor, so the Deployment leaves replicas out of the applied object: an
	// apply then never resets the size the HPA chose. Terraform twin:
	// locals.deployment_replicas.
	if !spec.GetAvailability().GetHorizontalPodAutoscaling().GetEnabled() {
		replicas := int32(1)
		if spec.Availability != nil && spec.Availability.Replicas != nil {
			replicas = *spec.Availability.Replicas
		}
		deploymentSpecArgs.Replicas = pulumi.Int(int(replicas))
	}

	if spec.Availability != nil {
		if spec.Availability.MinReadySeconds != nil {
			deploymentSpecArgs.MinReadySeconds = pulumi.Int(int(*spec.Availability.MinReadySeconds))
		}
		if spec.Availability.RevisionHistoryLimit != nil {
			deploymentSpecArgs.RevisionHistoryLimit = pulumi.Int(int(*spec.Availability.RevisionHistoryLimit))
		}
		if spec.Availability.ProgressDeadlineSeconds != nil {
			deploymentSpecArgs.ProgressDeadlineSeconds = pulumi.Int(int(*spec.Availability.ProgressDeadlineSeconds))
		}
		if spec.Availability.Paused {
			deploymentSpecArgs.Paused = pulumi.Bool(true)
		}
	}

	deploymentArgs := &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.KubernetesDeployment.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
			Annotations: pulumi.StringMap{
				// Force server-side apply to take ownership of fields on conflicts —
				// without it, a previously kubectl-managed deployment blocks updates.
				"pulumi.com/patchForce": pulumi.String("true"),
			},
		},
		Spec: deploymentSpecArgs,
	}

	deployOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdDeployment, err := appsv1.NewDeployment(ctx,
		locals.KubernetesDeployment.Metadata.Name,
		deploymentArgs,
		deployOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create deployment")
	}

	return createdDeployment, nil
}

func getStrategy(availability *kubernetesdeploymentv1alpha1.KubernetesDeploymentAvailability) *kubernetesdeploymentv1alpha1.KubernetesDeploymentStrategy {
	if availability == nil {
		return nil
	}
	return availability.Strategy
}

// buildDeploymentStrategy converts the strategy spec into Pulumi args.
// Returns nil when unset so Kubernetes applies its defaults (RollingUpdate, 25%/25%).
func buildDeploymentStrategy(protoStrategy *kubernetesdeploymentv1alpha1.KubernetesDeploymentStrategy) *appsv1.DeploymentStrategyArgs {
	if protoStrategy == nil {
		return nil
	}

	strategyType := protoStrategy.Type
	if strategyType == "" {
		strategyType = "RollingUpdate"
	}

	strategy := &appsv1.DeploymentStrategyArgs{
		Type: pulumi.String(strategyType),
	}

	// A Recreate strategy must NOT carry a rollingUpdate block — the API server
	// rejects the combination.
	if strategyType == "RollingUpdate" {
		rollingUpdate := &appsv1.RollingUpdateDeploymentArgs{}
		hasRollingConfig := false
		if protoStrategy.MaxUnavailable != "" {
			rollingUpdate.MaxUnavailable = parseIntOrString(protoStrategy.MaxUnavailable)
			hasRollingConfig = true
		}
		if protoStrategy.MaxSurge != "" {
			rollingUpdate.MaxSurge = parseIntOrString(protoStrategy.MaxSurge)
			hasRollingConfig = true
		}
		if hasRollingConfig {
			strategy.RollingUpdate = rollingUpdate
		}
	}

	return strategy
}

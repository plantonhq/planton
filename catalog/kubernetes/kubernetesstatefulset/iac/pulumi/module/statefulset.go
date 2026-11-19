package module

import (
	"github.com/pkg/errors"
	kubernetesv1 "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesstatefulsetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstatefulset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadpod"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadreplicas"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// statefulSet renders the apps/v1 StatefulSet. The pod template is assembled
// through the shared workloadpod builders so container and pod semantics match
// every other workload kind exactly; only StatefulSet-specific mechanics
// (stable identity, ordered rollout, per-replica storage) live here.
func statefulSet(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, createdImagePullSecret *kubernetescorev1.Secret,
	createdService *kubernetescorev1.Service, namespaceDeps []pulumi.ResourceOption) error {

	spec := locals.KubernetesStatefulSet.Spec

	containers := workloadpod.BuildContainers(
		spec.Container.App, spec.Container.Sidecars,
		// Default name for the main container when the spec omits one. "app" matches
		// the selector-label convention and reads correctly in kubectl output.
		"app",
		locals.EnvSecretName,
	)

	// The pod volume list is the union of every container's mounts — app,
	// sidecars, and init containers all contribute. Mounts whose claim name
	// matches a volumeClaimTemplate carry no pod-level volume; the StatefulSet
	// controller binds those per pod.
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

	statefulSetSpecArgs := &appsv1.StatefulSetSpecArgs{
		Selector: &metav1.LabelSelectorArgs{
			MatchLabels: pulumi.ToStringMap(locals.SelectorLabels),
		},
		Template: podTemplate,
		// The governing Service — the source of each replica's stable DNS identity.
		ServiceName:          pulumi.String(locals.KubeServiceName),
		UpdateStrategy:       buildUpdateStrategy(spec.UpdateStrategy),
		VolumeClaimTemplates: buildVolumeClaimTemplates(spec.VolumeClaimTemplates, locals),
	}

	// Replica count: scaling stateful members is application-aware work (data
	// sync, quorum changes), so there is deliberately no inline HPA on this kind.
	// Set, `replicas` is the size every apply holds; unset, Kubernetes or an
	// external autoscaler owns the size and an apply never moves the live count
	// (workloadreplicas says why leaving the field out is not enough). Terraform
	// twin: locals.statefulset_replicas.
	var declaredReplicas *int32
	if spec.Availability != nil {
		declaredReplicas = spec.Availability.Replicas
	}
	replicas := workloadreplicas.For(declaredReplicas, false)
	statefulSetSpecArgs.Replicas = replicas.Replicas()

	if spec.PodManagementPolicy != "" {
		statefulSetSpecArgs.PodManagementPolicy = pulumi.String(spec.PodManagementPolicy)
	}

	if spec.PvcRetentionPolicy != nil {
		retentionArgs := &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicyArgs{}
		if spec.PvcRetentionPolicy.WhenDeleted != "" {
			retentionArgs.WhenDeleted = pulumi.String(spec.PvcRetentionPolicy.WhenDeleted)
		}
		if spec.PvcRetentionPolicy.WhenScaled != "" {
			retentionArgs.WhenScaled = pulumi.String(spec.PvcRetentionPolicy.WhenScaled)
		}
		statefulSetSpecArgs.PersistentVolumeClaimRetentionPolicy = retentionArgs
	}

	// PARITY-EXCEPTION: the Terraform kubernetes provider's stateful_set has no
	// `ordinals` block, so the Terraform module (iac/tf/statefulset.tf) cannot
	// express spec.ordinals.start — replicas there are always numbered from the
	// default base 0. Pulumi renders it fully here.
	if spec.Ordinals != nil && spec.Ordinals.Start != nil {
		statefulSetSpecArgs.Ordinals = &appsv1.StatefulSetOrdinalsArgs{
			Start: pulumi.Int(int(*spec.Ordinals.Start)),
		}
	}

	if spec.Availability != nil {
		if spec.Availability.MinReadySeconds != nil {
			statefulSetSpecArgs.MinReadySeconds = pulumi.Int(int(*spec.Availability.MinReadySeconds))
		}
		if spec.Availability.RevisionHistoryLimit != nil {
			statefulSetSpecArgs.RevisionHistoryLimit = pulumi.Int(int(*spec.Availability.RevisionHistoryLimit))
		}
	}

	statefulSetArgs := &appsv1.StatefulSetArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.KubernetesStatefulSet.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
			Annotations: pulumi.StringMap{
				// Force server-side apply to take ownership of fields on conflicts —
				// without it, a previously kubectl-managed stateful set blocks updates.
				"pulumi.com/patchForce": pulumi.String("true"),
			},
		},
		Spec: statefulSetSpecArgs,
	}

	stsOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{createdService}),
	}, namespaceDeps...)
	stsOpts = append(stsOpts, replicas.Options()...)
	_, err := appsv1.NewStatefulSet(ctx,
		locals.KubernetesStatefulSet.Metadata.Name,
		statefulSetArgs,
		stsOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create stateful set")
	}

	return nil
}

// buildUpdateStrategy converts the update strategy spec into Pulumi args.
// Returns nil when unset so Kubernetes applies its defaults (RollingUpdate,
// one pod at a time from the highest ordinal down).
func buildUpdateStrategy(protoStrategy *kubernetesstatefulsetv1alpha1.KubernetesStatefulSetUpdateStrategy) *appsv1.StatefulSetUpdateStrategyArgs {
	if protoStrategy == nil {
		return nil
	}

	strategyType := protoStrategy.Type
	if strategyType == "" {
		strategyType = "RollingUpdate"
	}

	strategy := &appsv1.StatefulSetUpdateStrategyArgs{
		Type: pulumi.String(strategyType),
	}

	// An OnDelete strategy must NOT carry a rollingUpdate block — the API server
	// rejects the combination.
	if strategyType == "RollingUpdate" {
		rollingUpdate := &appsv1.RollingUpdateStatefulSetStrategyArgs{}
		hasRollingConfig := false
		if protoStrategy.Partition != nil {
			rollingUpdate.Partition = pulumi.Int(int(*protoStrategy.Partition))
			hasRollingConfig = true
		}
		// PARITY-EXCEPTION: the Terraform kubernetes provider's stateful_set
		// rolling_update block supports only `partition`, so the Terraform module
		// (iac/tf/statefulset.tf) cannot express max_unavailable — rollouts there
		// fall back to the one-pod-at-a-time default. Pulumi renders it fully here.
		if protoStrategy.MaxUnavailable != "" {
			rollingUpdate.MaxUnavailable = parseIntOrString(protoStrategy.MaxUnavailable)
			hasRollingConfig = true
		}
		if hasRollingConfig {
			strategy.RollingUpdate = rollingUpdate
		}
	}

	return strategy
}

// buildVolumeClaimTemplates renders the PVC templates each replica's storage is
// stamped from. The StatefulSet controller creates one PVC per template per
// replica (<template>-<name>-<ordinal>) and re-binds it across pod restarts —
// this is the mechanism that makes the storage stateful.
func buildVolumeClaimTemplates(templates []*kubernetesstatefulsetv1alpha1.KubernetesStatefulSetVolumeClaimTemplate,
	locals *Locals) kubernetescorev1.PersistentVolumeClaimTypeArray {
	if len(templates) == 0 {
		return nil
	}

	result := make(kubernetescorev1.PersistentVolumeClaimTypeArray, 0, len(templates))
	for _, t := range templates {
		claimSpec := &kubernetescorev1.PersistentVolumeClaimSpecArgs{
			Resources: &kubernetescorev1.VolumeResourceRequirementsArgs{
				Requests: pulumi.ToStringMap(map[string]string{"storage": t.Size}),
			},
		}

		// The spec documents ["ReadWriteOnce"] as the default: one node at a time
		// suits per-replica stateful storage, and it is the only mode every storage
		// driver supports.
		accessModes := t.AccessModes
		if len(accessModes) == 0 {
			accessModes = []string{"ReadWriteOnce"}
		}
		claimSpec.AccessModes = pulumi.ToStringArray(accessModes)

		if t.StorageClass != "" {
			claimSpec.StorageClassName = pulumi.StringPtr(t.StorageClass)
		}
		if t.VolumeMode != "" {
			claimSpec.VolumeMode = pulumi.StringPtr(t.VolumeMode)
		}

		result = append(result, &kubernetescorev1.PersistentVolumeClaimTypeArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:   pulumi.String(t.Name),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
			Spec: claimSpec,
		})
	}
	return result
}

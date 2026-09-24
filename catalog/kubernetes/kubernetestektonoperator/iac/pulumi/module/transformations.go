package module

import (
	"strings"

	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetestektonoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestektonoperator/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// skipAwaitTransformation marks every resource in the WORKLOADS group
// with the provider's skip-await annotation. The group applies BEFORE
// the CRDs group (the destroy-ordering chain in Resources), and nothing
// webhook-shaped in it can become ready until the CRDs exist: the
// webhook binary FATALS at startup until the tektoninstallersets CRD is
// served (webhook_init.go deletes/creates its own webhook InstallerSet
// unconditionally — verified in the operator source and live), so
// awaiting the webhook Deployment deadlocks the chain, and awaiting the
// webhook Service deadlocks one layer later on its empty endpoints
// (also verified live). Everything converges once the CRDs land
// (kubelet backoff restarts the pods); the component's E2E verifier
// owns rollout readiness. The Terraform twin sets
// wait_for_rollout = false on the same group for the same reason.
func skipAwaitTransformation() func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
	return func(state map[string]interface{}, _ ...pulumi.ResourceOption) {
		metadata, _ := state["metadata"].(map[string]interface{})
		if metadata == nil {
			metadata = map[string]interface{}{}
			state["metadata"] = metadata
		}
		annotations, _ := metadata["annotations"].(map[string]interface{})
		if annotations == nil {
			annotations = map[string]interface{}{}
			metadata["annotations"] = annotations
		}
		annotations["pulumi.com/skipAwait"] = "true"
	}
}

// autoInstallTransformation ALWAYS patches the tekton-config-defaults
// ConfigMap's AUTOINSTALL_COMPONENTS key to "false" (the release ships
// "true"). This is a design invariant, not a knob: with auto-install on,
// the operator creates its own TektonConfig named `config` (profile
// `all`) at startup — the exact object the KubernetesTekton declaration
// kind renders — and the two managers then fight over the same fields
// through server-side apply. Disabling it makes the declaration kind the
// single owner: installing the operator alone deploys no Tekton
// components. The Terraform twin performs the identical patch through
// locals merges.
func autoInstallTransformation() func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
	return func(state map[string]interface{}, _ ...pulumi.ResourceOption) {
		if state["kind"] != "ConfigMap" {
			return
		}
		metadata, _ := state["metadata"].(map[string]interface{})
		if metadata == nil || metadata["name"] != vars.ConfigDefaultsConfigMapName {
			return
		}
		data, _ := state["data"].(map[string]interface{})
		if data == nil {
			data = map[string]interface{}{}
			state["data"] = data
		}
		data["AUTOINSTALL_COMPONENTS"] = "false"
	}
}

// deploymentTransformation applies the spec's typed overrides onto the
// release manifest's two Deployments — every other document applies
// verbatim (faithful distribution). The Terraform twin performs the
// identical patches through locals merges.
//
// Patched surfaces (all optional):
//   - the operator Deployment's containers (tekton-operator-lifecycle +
//     tekton-operator-cluster-operations share one image): operator_image,
//     operator_resources,
//   - the webhook Deployment's container: webhook_image,
//     webhook_resources,
//   - both pods: nodeSelector / tolerations / imagePullSecrets,
//   - image_registry: both Deployments' manifest images move to the
//     registry (an explicit operator_image / webhook_image still wins), and
//     the lifecycle container, the one that reconciles components, gets its
//     ghcr.io IMAGE_* values moved and every images table entry appended,
//     moved, so each component the operator installs pulls from it too.
//
// images is nil when image_registry is empty.
func deploymentTransformation(spec *kubernetestektonoperatorv1alpha1.KubernetesTektonOperatorSpec, images *tektonImages) func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
	return func(state map[string]interface{}, _ ...pulumi.ResourceOption) {
		if state["kind"] != "Deployment" {
			return
		}
		metadata, _ := state["metadata"].(map[string]interface{})
		if metadata == nil {
			return
		}

		var image string
		var resources map[string]interface{}
		isOperator := false
		switch metadata["name"] {
		case vars.OperatorDeploymentName:
			image = joinImage(spec.GetOperatorImage())
			resources = resourcesMap(spec.GetOperatorResources())
			isOperator = true
		case vars.WebhookDeploymentName:
			image = joinImage(spec.GetWebhookImage())
			resources = resourcesMap(spec.GetWebhookResources())
		default:
			return
		}
		registry := spec.GetImageRegistry()

		deploymentSpec, _ := state["spec"].(map[string]interface{})
		template, _ := deploymentSpec["template"].(map[string]interface{})
		podSpec, _ := template["spec"].(map[string]interface{})
		if podSpec == nil {
			return
		}

		containers, _ := podSpec["containers"].([]interface{})
		for _, entry := range containers {
			container, _ := entry.(map[string]interface{})
			if container == nil {
				continue
			}
			if image != "" {
				container["image"] = image
			} else if manifestImage, ok := container["image"].(string); ok {
				container["image"] = mirroredImage(manifestImage, registry)
			}
			if resources != nil {
				container["resources"] = resources
			}
			if isOperator && images != nil && container["name"] == vars.LifecycleContainerName {
				container["env"] = mirroredImageEnv(container["env"], images, registry)
			}
		}

		if len(spec.GetNodeSelector()) > 0 {
			nodeSelector := map[string]interface{}{}
			for key, value := range spec.GetNodeSelector() {
				nodeSelector[key] = value
			}
			podSpec["nodeSelector"] = nodeSelector
		}
		if tolerations := tolerationsSlice(spec); tolerations != nil {
			podSpec["tolerations"] = tolerations
		}
		if names := pullSecretNames(spec); len(names) > 0 {
			pullSecrets := []interface{}{}
			for _, name := range names {
				pullSecrets = append(pullSecrets, map[string]interface{}{"name": name})
			}
			podSpec["imagePullSecrets"] = pullSecrets
		}
	}
}

// mirroredImageEnv moves the lifecycle container's IMAGE_* values to the
// registry and appends every table entry the manifest does not already set,
// in table order, so both engines render one env list.
func mirroredImageEnv(env interface{}, images *tektonImages, registry string) []interface{} {
	entries, _ := env.([]interface{})
	out := make([]interface{}, 0, len(entries)+len(images.Images))
	present := map[string]bool{}
	for _, raw := range entries {
		entry, _ := raw.(map[string]interface{})
		if entry == nil {
			out = append(out, raw)
			continue
		}
		name, _ := entry["name"].(string)
		present[name] = true
		if value, ok := entry["value"].(string); ok && strings.HasPrefix(name, "IMAGE_") {
			moved := map[string]interface{}{}
			for key, field := range entry {
				moved[key] = field
			}
			moved["value"] = mirroredImage(value, registry)
			entry = moved
		}
		out = append(out, entry)
	}
	for _, image := range images.Images {
		if present[image.Name] {
			continue
		}
		out = append(out, map[string]interface{}{
			"name":  image.Name,
			"value": mirroredImage(image.Image, registry),
		})
	}
	return out
}

// joinImage folds repo:tag; empty keeps the release manifest's
// digest-pinned image.
func joinImage(image *kubernetes.ContainerImage) string {
	if image.GetRepo() == "" && image.GetTag() == "" {
		return ""
	}
	if image.GetRepo() == "" || image.GetTag() == "" {
		return image.GetRepo() + image.GetTag()
	}
	return image.GetRepo() + ":" + image.GetTag()
}

// pullSecretNames joins image_pull_secrets with both image overrides' own
// pull_secret_name entries, deduplicated — a private image override
// naturally travels with its own pull secret (Terraform twin:
// image_pull_secret_names in locals.tf).
func pullSecretNames(spec *kubernetestektonoperatorv1alpha1.KubernetesTektonOperatorSpec) []string {
	names := append([]string{}, spec.GetImagePullSecrets()...)
	for _, extra := range []string{
		spec.GetOperatorImage().GetPullSecretName(),
		spec.GetWebhookImage().GetPullSecretName(),
	} {
		if extra == "" {
			continue
		}
		seen := false
		for _, name := range names {
			if name == extra {
				seen = true
				break
			}
		}
		if !seen {
			names = append(names, extra)
		}
	}
	return names
}

// resourcesMap renders a container resources override; nil keeps the
// release manifest's defaults (none set at the pin).
func resourcesMap(resources *kubernetes.ContainerResources) map[string]interface{} {
	if resources == nil {
		return nil
	}

	out := map[string]interface{}{}
	if requests := quantityEntries(resources.GetRequests().GetCpu(), resources.GetRequests().GetMemory()); requests != nil {
		out["requests"] = requests
	}
	if limits := quantityEntries(resources.GetLimits().GetCpu(), resources.GetLimits().GetMemory()); limits != nil {
		out["limits"] = limits
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func quantityEntries(cpu string, memory string) map[string]interface{} {
	quantities := map[string]interface{}{}
	if cpu != "" {
		quantities["cpu"] = cpu
	}
	if memory != "" {
		quantities["memory"] = memory
	}
	if len(quantities) == 0 {
		return nil
	}
	return quantities
}

// tolerationsSlice renders the pod tolerations override; nil when empty.
func tolerationsSlice(spec *kubernetestektonoperatorv1alpha1.KubernetesTektonOperatorSpec) []interface{} {
	if len(spec.GetTolerations()) == 0 {
		return nil
	}

	out := []interface{}{}
	for _, toleration := range spec.GetTolerations() {
		entry := map[string]interface{}{}
		if toleration.GetKey() != "" {
			entry["key"] = toleration.GetKey()
		}
		if toleration.GetOperator() != "" {
			entry["operator"] = toleration.GetOperator()
		}
		if toleration.GetValue() != "" {
			entry["value"] = toleration.GetValue()
		}
		if toleration.GetEffect() != "" {
			entry["effect"] = toleration.GetEffect()
		}
		if toleration.TolerationSeconds != nil {
			entry["tolerationSeconds"] = int(toleration.GetTolerationSeconds())
		}
		out = append(out, entry)
	}
	return out
}

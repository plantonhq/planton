package module

import (
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	kuberneteskeycloakoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskeycloakoperator/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespacedBundleKinds are the kinds in THIS bundle that live in a
// namespace and therefore get metadata.namespace stamped (the bundle
// ships every document WITHOUT a namespace field — upstream expects
// kustomize to set it). ClusterRole, ClusterRoleBinding and the CRDs
// are cluster-scoped and get none. The set is bundle-specific and
// pinned with the release — the Terraform twin's namespaced_kinds local
// mirrors it.
var namespacedBundleKinds = map[string]bool{
	"ServiceAccount": true,
	"Role":           true,
	"RoleBinding":    true,
	"Service":        true,
	"Deployment":     true,
}

// skipAwaitTransformation marks every resource in the WORKLOADS group
// with the provider's skip-await annotation. The group applies BEFORE
// the CRDs group (the destroy-ordering chain in Resources), and the
// operator Deployment cannot become ready until the CRDs exist: the
// JOSDK (Quarkus Operator SDK) operator crash-loops at startup until
// the k8s.keycloak.org CRDs are served, so awaiting its rollout here
// would deadlock the chain. Everything converges once the CRDs land
// (kubelet backoff restarts the pod), and no readiness gate is needed:
// the operator registers no admission webhook, so a Keycloak resource
// applied before it runs is stored as written and reconciled once it
// starts. The Terraform twin sets wait_for_rollout = false on the same
// group for the same reason.
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

// namespaceTransformation performs what upstream's kustomize overlay
// would: it stamps metadata.namespace = <spec namespace> onto every
// NAMESPACED bundle document, and rewrites every RoleBinding /
// ClusterRoleBinding ServiceAccount subject's namespace to the same
// value. Upstream bakes `namespace: keycloak` into exactly ONE
// ClusterRoleBinding subject and leaves the rest empty for kustomize —
// both wrong for a configurable namespace, so ALL subjects are
// rewritten. The Terraform twin performs the identical stamping through
// locals merges (stamped_documents).
func namespaceTransformation(namespace string) func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
	return func(state map[string]interface{}, _ ...pulumi.ResourceOption) {
		kind, _ := state["kind"].(string)

		if namespacedBundleKinds[kind] {
			metadata, _ := state["metadata"].(map[string]interface{})
			if metadata == nil {
				metadata = map[string]interface{}{}
				state["metadata"] = metadata
			}
			metadata["namespace"] = namespace
		}

		if kind == "RoleBinding" || kind == "ClusterRoleBinding" {
			subjects, _ := state["subjects"].([]interface{})
			for _, entry := range subjects {
				subject, _ := entry.(map[string]interface{})
				if subject == nil || subject["kind"] != "ServiceAccount" {
					continue
				}
				subject["namespace"] = namespace
			}
		}
	}
}

// deploymentTransformation applies the spec's typed overrides onto the
// bundle's operator Deployment — every other document applies verbatim
// (faithful distribution). The Terraform twin performs the identical
// patches through locals merges (patched_operator_deployment).
//
// Patched surfaces (all optional):
//   - the operator container image (operator_image; the tag must stay
//     at the pinned release or the CRD schema drifts),
//   - the RELATED_IMAGE_KEYCLOAK env value (default_keycloak_image —
//     the default Keycloak server image the operator stamps into
//     Keycloak StatefulSets whose declaration sets no image),
//   - the operator container resources (upstream ships requests
//     300m/450Mi, limits 700m/450Mi — the spec's proto defaults),
//   - pod nodeSelector / tolerations (spec.scheduling; upstream ships
//     none).
func deploymentTransformation(spec *kuberneteskeycloakoperatorv1alpha1.KubernetesKeycloakOperatorSpec) func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
	return func(state map[string]interface{}, _ ...pulumi.ResourceOption) {
		if state["kind"] != "Deployment" {
			return
		}
		metadata, _ := state["metadata"].(map[string]interface{})
		if metadata == nil || metadata["name"] != vars.DeploymentName {
			return
		}

		deploymentSpec, _ := state["spec"].(map[string]interface{})
		template, _ := deploymentSpec["template"].(map[string]interface{})
		podSpec, _ := template["spec"].(map[string]interface{})
		if podSpec == nil {
			return
		}

		resources := resourcesMap(spec.GetResources())
		containers, _ := podSpec["containers"].([]interface{})
		for _, entry := range containers {
			container, _ := entry.(map[string]interface{})
			if container == nil {
				continue
			}
			if spec.GetOperatorImage() != "" {
				container["image"] = spec.GetOperatorImage()
			}
			if resources != nil {
				container["resources"] = resources
			}
			if spec.GetDefaultKeycloakImage() != "" {
				env, _ := container["env"].([]interface{})
				for _, envEntry := range env {
					envVar, _ := envEntry.(map[string]interface{})
					if envVar == nil || envVar["name"] != vars.RelatedImageKeycloakEnvName {
						continue
					}
					envVar["value"] = spec.GetDefaultKeycloakImage()
				}
			}
		}

		scheduling := spec.GetScheduling()
		if len(scheduling.GetNodeSelector()) > 0 {
			nodeSelector := map[string]interface{}{}
			for key, value := range scheduling.GetNodeSelector() {
				nodeSelector[key] = value
			}
			podSpec["nodeSelector"] = nodeSelector
		}
		if tolerations := tolerationsSlice(scheduling.GetTolerations()); tolerations != nil {
			podSpec["tolerations"] = tolerations
		}
	}
}

// resourcesMap renders the operator container resources override; nil
// keeps the bundle's own values (requests 300m/450Mi, limits
// 700m/450Mi — identical to the spec's proto defaults, so an unset
// spec drifts nothing).
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
func tolerationsSlice(tolerations []*kubernetes.WorkloadToleration) []interface{} {
	if len(tolerations) == 0 {
		return nil
	}

	out := []interface{}{}
	for _, toleration := range tolerations {
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

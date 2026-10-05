// Package kubernetesannotationkeys defines the manifest annotation keys that carry
// Kubernetes-specific platform behavior.
//
// Platform-behavior signals live in metadata.annotations, never metadata.labels:
// labels are derived into cloud-provider tags by planton IaC modules, so a platform
// key there would leak internal detail onto the user's real provider resources.
//
// No annotation carries a credential or points at one: a module reads nothing
// beside its own directory, so anything a workload needs to pull, mount, or
// connect is declared on its own spec (e.g. `pod.image_registries` for a private
// registry's login) and resolved before the module runs.
package kubernetesannotationkeys

const (
	// KubeContextAnnotationKey specifies the kubectl context to use for Kubernetes deployments
	KubeContextAnnotationKey = "kubernetes.planton.dev/context"
)

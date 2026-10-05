package module

import (
	kubernetesconfigmapv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesconfigmap/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds all derived configuration and state for the module
type Locals struct {
	// Context for Pulumi operations
	Ctx *pulumi.Context

	// IaC input containing the target resource
	IacInput *kubernetesconfigmapv1alpha1.KubernetesConfigMapIacInput

	// Target configmap resource
	Target *kubernetesconfigmapv1alpha1.KubernetesConfigMap

	// Spec from the target
	Spec *kubernetesconfigmapv1alpha1.KubernetesConfigMapSpec

	// ConfigMap name (metadata.name of the ConfigMap in the cluster)
	ConfigMapName string

	// Resolved namespace the ConfigMap is created in
	Namespace string

	// Combined labels (spec labels + standard labels)
	Labels map[string]string

	// Combined annotations (spec annotations)
	Annotations map[string]string

	// Whether the configmap is immutable
	Immutable bool

	// UTF-8 configuration entries (ConfigMap `data`)
	Data map[string]string

	// Base64-encoded binary entries (ConfigMap `binaryData`)
	BinaryData map[string]string
}

// initializeLocals creates and populates the Locals struct
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetesconfigmapv1alpha1.KubernetesConfigMapIacInput) *Locals {
	locals := &Locals{
		Ctx:      ctx,
		IacInput: iacInput,
		Target:   iacInput.Target,
		Spec:     iacInput.Target.Spec,
	}

	locals.ConfigMapName = locals.Spec.Name

	// namespace is a StringValueOrRef foreign key. References are resolved to
	// literal strings before the module runs, so GetValue() returns the final
	// namespace name. When the field is omitted entirely, fall back to the
	// cluster's "default" namespace — the same behavior as kubectl without a
	// namespace flag.
	locals.Namespace = locals.Spec.GetNamespace().GetValue()
	if locals.Namespace == "" {
		locals.Namespace = "default"
	}

	locals.Immutable = locals.Spec.Immutable

	// Build labels
	locals.Labels = buildLabels(locals)

	// Build annotations
	locals.Annotations = buildAnnotations(locals)

	// data holds UTF-8 entries; binary_data holds values that are already
	// base64-encoded strings — the exact wire form Kubernetes uses for
	// binaryData — so both maps pass through unchanged.
	locals.Data = locals.Spec.Data
	locals.BinaryData = locals.Spec.BinaryData

	return locals
}

// buildLabels combines spec labels with standard labels
func buildLabels(locals *Locals) map[string]string {
	labels := make(map[string]string)

	// Add standard labels
	labels["managed-by"] = "planton"
	labels["resource"] = locals.Target.Metadata.Name
	labels["resource-kind"] = "KubernetesConfigMap"

	// Add spec labels
	for k, v := range locals.Spec.Labels {
		labels[k] = v
	}

	return labels
}

// buildAnnotations combines spec annotations
func buildAnnotations(locals *Locals) map[string]string {
	annotations := make(map[string]string)

	// Add spec annotations
	for k, v := range locals.Spec.Annotations {
		annotations[k] = v
	}

	return annotations
}

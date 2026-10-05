package module

import (
	"strconv"

	kubernetesotelcollectorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesotelcollector/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesotelcollectorv1alpha1.KubernetesOtelCollectorSpec

	// Resource-identity labels stamped on the CR and the module-created
	// namespace.
	Labels map[string]string

	// Namespace the collector runs in (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// ResourceName is metadata.name — the CR name the operator derives
	// every child name from.
	ResourceName string

	// EffectiveMode is the mode the collector actually runs in (the CRD
	// defaults an unset mode to deployment). Drives the mode-dependent
	// rendering rules and the sidecar-empties-outputs contract.
	EffectiveMode string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesotelcollectorv1alpha1.KubernetesOtelCollectorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesOtelCollector.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// The proto enum's value names ARE the CR's mode strings
	// ("deployment", "daemonset", "statefulset", "sidecar") — chosen so
	// both engines render the identical scalar.
	effectiveMode := "deployment"
	if spec.Mode != nil && spec.GetMode() != kubernetesotelcollectorv1alpha1.KubernetesOtelCollectorMode_kubernetes_otel_collector_mode_unspecified {
		effectiveMode = spec.GetMode().String()
	}

	return &Locals{
		Spec:          spec,
		Labels:        labels,
		Namespace:     spec.Namespace.GetValue(),
		ResourceName:  target.Metadata.Name,
		EffectiveMode: effectiveMode,
	}
}

// isWorkloadMode reports whether replicas/autoscaler apply — the
// deployment/statefulset modes. In daemonset/sidecar modes the (possibly
// middleware-defaulted) replicas value is IGNORED by design (the spec
// CEL's expressibility tolerance). Twin of the Terraform module's
// is_workload_mode.
func (l *Locals) isWorkloadMode() bool {
	return l.EffectiveMode == "deployment" || l.EffectiveMode == "statefulset"
}

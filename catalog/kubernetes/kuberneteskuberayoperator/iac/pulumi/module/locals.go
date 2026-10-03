package module

import (
	"strconv"

	kuberneteskuberayoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskuberayoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteskuberayoperatorv1alpha1.KubernetesKubeRayOperatorSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace the operator installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name. The module pins the chart's
	// nameOverride AND fullnameOverride to it (the chart hardcodes both
	// to "kuberay-operator" in its values), so every chart-derived name
	// hangs off this value.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// WatchNamespaces is the operator's watch scope (empty =
	// cluster-wide — the normal one-operator-per-cluster posture).
	WatchNamespaces []string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskuberayoperatorv1alpha1.KubernetesKubeRayOperatorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKubeRayOperator.String(),
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

	chartVersion := spec.GetChartVersion()
	if chartVersion == "" {
		chartVersion = vars.DefaultChartVersion
	}

	return &Locals{
		Spec:            spec,
		Labels:          labels,
		Namespace:       spec.Namespace.GetValue(),
		ReleaseName:     target.Metadata.Name,
		ChartVersion:    chartVersion,
		WatchNamespaces: spec.GetWatchNamespaces(),
	}
}

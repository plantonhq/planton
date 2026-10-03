package module

import (
	"strconv"

	kubernetesmetricsserverv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesmetricsserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesmetricsserverv1alpha1.KubernetesMetricsServerSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace metrics-server installs into (resolved literal from the
	// spec's value-or-ref; kube-system by upstream convention).
	Namespace string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Name of the registered APIService — empty when api_service.create
	// is false (the outputs contract mirrors what actually exists).
	ApiServiceName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesmetricsserverv1alpha1.KubernetesMetricsServerIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesMetricsServer.String(),
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

	// The APIService name is fixed by the API contract, not the chart (the
	// resource-metrics API has exactly this one group/version). Empty only
	// when the spec explicitly opts out of creating it (default: create).
	apiServiceName := "v1beta1.metrics.k8s.io"
	if s := spec.GetApiService(); s != nil && s.Create != nil && !s.GetCreate() {
		apiServiceName = ""
	}

	return &Locals{
		Spec:           spec,
		Labels:         labels,
		Namespace:      spec.Namespace.GetValue(),
		ChartVersion:   chartVersion,
		ApiServiceName: apiServiceName,
	}
}

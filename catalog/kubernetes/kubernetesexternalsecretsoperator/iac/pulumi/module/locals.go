package module

import (
	"strconv"

	kubernetesexternalsecretsoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesexternalsecretsoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesexternalsecretsoperatorv1alpha1.KubernetesExternalSecretsOperatorSpec

	// Resource-identity labels stamped on the namespace this module creates
	// (never injected into the chart's own resources — Helm owns those).
	Labels map[string]string

	// Namespace the operator installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name. Fixed to the chart name: one External Secrets
	// Operator per cluster is an upstream architectural constraint
	// (cluster-scoped CRDs and webhook configuration), so a
	// manifest-derived name would only invite a second broken install.
	ReleaseName string

	// Controller ServiceAccount name — the module pins it to the chart
	// name (serviceAccount.name) so cloud-side ambient-identity bindings
	// have a deterministic subject.
	ControllerServiceAccount string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesexternalsecretsoperatorv1alpha1.KubernetesExternalSecretsOperatorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesExternalSecretsOperator.String(),
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
		Spec:                     spec,
		Labels:                   labels,
		Namespace:                spec.Namespace.GetValue(),
		ReleaseName:              vars.HelmChartName,
		ControllerServiceAccount: vars.HelmChartName,
		ChartVersion:             chartVersion,
	}
}

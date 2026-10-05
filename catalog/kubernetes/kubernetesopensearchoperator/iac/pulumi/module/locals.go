package module

import (
	"strconv"

	kubernetesopensearchoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopensearchoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesopensearchoperatorv1alpha1.KubernetesOpenSearchOperatorSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace the operator installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// DeploymentName is the operator controller-manager Deployment name,
	// derived exactly as the chart's own naming does it
	// (templates/_helpers.tpl "opensearch-operator.fullname" +
	// "-controller-manager" in the deployment template): fullname is the
	// release name when it already contains the chart name, otherwise
	// "<release>-opensearch-operator", truncated to 63 characters with a
	// trailing "-" trimmed. Valid as long as helm_values does not set
	// nameOverride/fullnameOverride (documented in the README).
	DeploymentName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesopensearchoperatorv1alpha1.KubernetesOpenSearchOperatorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesOpenSearchOperator.String(),
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

	releaseName := target.Metadata.Name

	return &Locals{
		Spec:           spec,
		Labels:         labels,
		Namespace:      spec.Namespace.GetValue(),
		ReleaseName:    releaseName,
		ChartVersion:   chartVersion,
		DeploymentName: deploymentName(releaseName),
	}
}

// deploymentName renders the controller-manager Deployment name. The
// module pins the chart's fullnameOverride to the resource name (see
// buildHelmValues), so the fullname IS the release name — verified live:
// without the pin the chart's default fullname
// ("<release>-opensearch-operator") pushes its metrics Service name
// ("<fullname>-controller-manager-metrics-service") past Kubernetes'
// 63-character limit for ordinary release names (the chart truncates the
// fullname but not the names built from it, so the install fails at the
// API server). Twin of the Terraform module's local.deployment_name —
// keep byte-identical.
func deploymentName(releaseName string) string {
	return releaseName + "-controller-manager"
}

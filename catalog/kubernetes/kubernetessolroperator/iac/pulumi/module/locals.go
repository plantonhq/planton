package module

import (
	"strconv"
	"strings"

	kubernetessolroperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessolroperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetessolroperatorv1alpha1.KubernetesSolrOperatorSpec

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

	// DeploymentName replays the chart's fullname helper
	// (templates/_helpers.tpl, "solr-operator.fullname"): with no
	// nameOverride/fullnameOverride (the typed spec sets neither), the
	// Deployment is named after the release — the release name itself
	// when it already contains "solr-operator", otherwise
	// "<release>-solr-operator" — truncated to 63 chars with one
	// trailing "-" trimmed.
	DeploymentName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetessolroperatorv1alpha1.KubernetesSolrOperatorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesSolrOperator.String(),
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
		DeploymentName: chartFullname(releaseName),
	}
}

// chartFullname is the Go replay of the chart's fullname template:
//
//	{{- if contains $name .Release.Name -}}
//	{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
//	{{- else -}}
//	{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
//
// with $name = the chart name ("solr-operator") since the typed spec sets
// neither nameOverride nor fullnameOverride. A helm_values override of
// either would change the real name without being reflected here — the
// escape-hatch caveat both engines document.
func chartFullname(releaseName string) string {
	fullname := releaseName
	if !strings.Contains(releaseName, vars.HelmChartName) {
		fullname = releaseName + "-" + vars.HelmChartName
	}
	if len(fullname) > 63 {
		fullname = fullname[:63]
	}
	return strings.TrimSuffix(fullname, "-")
}

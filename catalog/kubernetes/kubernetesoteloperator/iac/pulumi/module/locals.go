package module

import (
	"strconv"

	kubernetesoteloperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesoteloperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesoteloperatorv1alpha1.KubernetesOtelOperatorSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace the operator installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name. The module pins the chart's
	// fullnameOverride to it, so every chart-derived name below hangs
	// off this value.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// WebhookService is the operator's webhook Service name
	// ("<name>-webhook", port 443) — where the API server sends
	// admission reviews and CRD conversion calls.
	WebhookService string

	// WebhookCertSecretName is the cert-manager-issued serving-cert
	// Secret ("<name>-controller-manager-service-cert") — the 33-char
	// suffix behind the 30-character name budget.
	WebhookCertSecretName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesoteloperatorv1alpha1.KubernetesOtelOperatorIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesOtelOperator.String(),
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
		Spec:                  spec,
		Labels:                labels,
		Namespace:             spec.Namespace.GetValue(),
		ReleaseName:           releaseName,
		ChartVersion:          chartVersion,
		WebhookService:        releaseName + "-webhook",
		WebhookCertSecretName: releaseName + "-controller-manager-service-cert",
	}
}

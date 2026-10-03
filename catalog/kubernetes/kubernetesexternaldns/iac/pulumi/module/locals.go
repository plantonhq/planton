package module

import (
	"strconv"

	kubernetesexternaldnsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesexternaldns/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesexternaldnsv1alpha1.KubernetesExternalDnsSpec

	// Resource-identity labels stamped on the module-created satellites
	// (namespace, credential Secrets — never injected into the chart's own
	// resources; Helm owns those).
	Labels map[string]string

	// Namespace ExternalDNS installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: multiple
	// ExternalDNS instances per cluster are a first-class pattern (one per
	// DNS provider / zone set, separated by TXT owner IDs), so each
	// manifest gets its own release.
	ReleaseName string

	// Controller ServiceAccount name. The chart creates the SA; the module
	// pins its name to metadata.name (serviceAccount.name) so cloud-side
	// identity bindings have a deterministic subject.
	ServiceAccountName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Deterministic names for credential satellites this module may
	// materialize (empty when the selected provider needs none).
	CloudflareSecretName string
	AwsSecretName        string
	GcpSecretName        string
	AzureSecretName      string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesexternaldnsv1alpha1.KubernetesExternalDnsIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesExternalDns.String(),
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
		Spec:                 spec,
		Labels:               labels,
		Namespace:            spec.Namespace.GetValue(),
		ReleaseName:          target.Metadata.Name,
		ServiceAccountName:   target.Metadata.Name,
		ChartVersion:         chartVersion,
		CloudflareSecretName: target.Metadata.Name + "-cloudflare-credentials",
		AwsSecretName:        target.Metadata.Name + "-aws-credentials",
		GcpSecretName:        target.Metadata.Name + "-gcp-credentials",
		AzureSecretName:      target.Metadata.Name + "-azure-config",
	}
}

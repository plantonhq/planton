package module

import (
	"strconv"

	kubernetescertmanagerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetescertmanager/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetescertmanagerv1alpha1.KubernetesCertManagerSpec

	// Resource-identity labels stamped on the namespace this module creates
	// (never injected into the chart's own resources — Helm owns those).
	Labels map[string]string

	// Namespace cert-manager installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name. Fixed to the chart name: one cert-manager per
	// cluster is an upstream architectural constraint (cluster-scoped CRDs
	// and webhooks), so a manifest-derived name would only invite a second
	// broken install.
	ReleaseName string

	// The chart's fullname for the controller ServiceAccount — the chart
	// derives it from the release name, so with the release named
	// "cert-manager" the ServiceAccount is "cert-manager". Exported for
	// cloud-side identity bindings (IRSA trust policy, GKE WI member,
	// Azure federated credential).
	ServiceAccountName string

	// Resolved cluster-resource namespace: spec.cluster_resource_namespace
	// when set, otherwise the installation namespace (cert-manager's own
	// default). Exported — KubernetesClusterIssuer materializes credential
	// Secrets here.
	ClusterResourceNamespace string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetescertmanagerv1alpha1.KubernetesCertManagerIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to
	// what the Terraform module stamps for the same manifest.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesCertManager.String(),
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

	namespace := spec.Namespace.GetValue()

	clusterResourceNamespace := spec.ClusterResourceNamespace
	if clusterResourceNamespace == "" {
		clusterResourceNamespace = namespace
	}

	chartVersion := spec.GetChartVersion()
	if chartVersion == "" {
		chartVersion = vars.DefaultChartVersion
	}

	return &Locals{
		Spec:                     spec,
		Labels:                   labels,
		Namespace:                namespace,
		ReleaseName:              vars.HelmChartName,
		ServiceAccountName:       vars.HelmChartName,
		ClusterResourceNamespace: clusterResourceNamespace,
		ChartVersion:             chartVersion,
	}
}

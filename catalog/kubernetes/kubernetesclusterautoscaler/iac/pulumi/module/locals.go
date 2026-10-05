package module

import (
	"fmt"
	"strconv"

	kubernetesclusterautoscalerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesclusterautoscaler/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesclusterautoscalerv1alpha1.KubernetesClusterAutoscalerSpec

	// Metadata name of the target resource — also the benign
	// autoDiscovery.clusterName gate value for the gce/kwok arms (see
	// values.go).
	ResourceName string

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace the autoscaler installs into (resolved literal from the
	// spec's value-or-ref; "kube-system" by upstream convention).
	Namespace string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// The chart's cloudProvider value for the selected oneof arm — drives
	// the per-provider template blocks in the chart's deployment.yaml AND
	// the derived object names below.
	CloudProvider string

	// ServiceAccountName is DERIVED from the chart's naming, verified in
	// templates/_helpers.tpl: with rbac.serviceAccount.name unset the
	// service account takes the fullname template, whose default name is
	// "<cloudProvider>-<chartName>" (NOT the bare chart name). That never
	// equals the release name, so fullname renders
	// "<release>-<cloudProvider>-<chartName>" — e.g. for the aws arm:
	// "cluster-autoscaler-aws-cluster-autoscaler" (well under the
	// 63-char truncation for every supported arm). This is the subject
	// cloud-side keyless bindings (IRSA trust policies, GCP WI bindings,
	// Entra federated credentials) are written against.
	ServiceAccountName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesclusterautoscalerv1alpha1.KubernetesClusterAutoscalerIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesClusterAutoscaler.String(),
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

	// The proto oneof guarantees exactly one arm; the chart key is the
	// autoscaler binary's provider name ("clusterapi", not "cluster_api").
	cloudProvider := ""
	switch {
	case spec.GetAws() != nil:
		cloudProvider = "aws"
	case spec.GetAzure() != nil:
		cloudProvider = "azure"
	case spec.GetGce() != nil:
		cloudProvider = "gce"
	case spec.GetClusterApi() != nil:
		cloudProvider = "clusterapi"
	case spec.GetCivo() != nil:
		cloudProvider = "civo"
	case spec.GetKwok() != nil:
		cloudProvider = "kwok"
	}

	return &Locals{
		Spec:          spec,
		ResourceName:  target.Metadata.Name,
		Labels:        labels,
		Namespace:     spec.Namespace.GetValue(),
		ChartVersion:  chartVersion,
		CloudProvider: cloudProvider,
		// fullname derivation verified in _helpers.tpl — see the field
		// comment above.
		ServiceAccountName: fmt.Sprintf("%s-%s-%s", vars.ReleaseName, cloudProvider, vars.HelmChartName),
	}
}

package module

import (
	"strconv"

	kubernetesciliumv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetescilium/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesciliumv1alpha1.KubernetesCiliumSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace Cilium installs into (resolved literal from the spec's
	// value-or-ref; kube-system by upstream convention).
	Namespace string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Cluster identity Cilium runs under — the spec value or the chart's
	// "default" when unset. Exported so consumers know the name this
	// cluster carries in Hubble flows and any future Cluster Mesh.
	ClusterName string

	// Name of the hubble-relay Service — fixed "hubble-relay" by the chart
	// templates (no fullname prefix), empty when relay is not deployed
	// (the outputs contract mirrors what actually exists).
	HubbleRelayServiceName string

	// Name of the hubble-ui Service — fixed "hubble-ui" by the chart
	// templates, empty when the UI is not deployed.
	HubbleUiServiceName string

	// Name of the GatewayClass Cilium's Gateway API implementation
	// registers — fixed "cilium" by the chart, empty when gateway_api is
	// off.
	GatewayClassName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesciliumv1alpha1.KubernetesCiliumIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesCilium.String(),
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

	clusterName := spec.GetClusterName()
	if clusterName == "" {
		clusterName = "default"
	}

	// Service and GatewayClass names below are FIXED by the chart templates
	// (verified against them — hubble-relay/hubble-ui Services and the
	// GatewayClass carry no release-derived prefix), so the outputs are
	// pure functions of the spec toggles: set when the component exists,
	// empty otherwise.
	hubbleRelayServiceName := ""
	hubbleUiServiceName := ""
	if hubble := spec.GetHubble(); hubble != nil {
		if hubble.GetRelay() {
			hubbleRelayServiceName = "hubble-relay"
		}
		if hubble.GetUi() {
			hubbleUiServiceName = "hubble-ui"
		}
	}

	gatewayClassName := ""
	if spec.GetGatewayApi() {
		gatewayClassName = "cilium"
	}

	return &Locals{
		Spec:                   spec,
		Labels:                 labels,
		Namespace:              spec.Namespace.GetValue(),
		ChartVersion:           chartVersion,
		ClusterName:            clusterName,
		HubbleRelayServiceName: hubbleRelayServiceName,
		HubbleUiServiceName:    hubbleUiServiceName,
		GatewayClassName:       gatewayClassName,
	}
}

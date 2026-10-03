package module

import (
	"fmt"
	"strconv"

	kubernetesargoworkflowsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesargoworkflows/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesargoworkflowsv1alpha1.KubernetesArgoWorkflowsSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace — never injected into the chart's own resources;
	// Helm owns those).
	Labels map[string]string

	// Namespace Argo Workflows installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// engines can coexist (pair with controller.instance_id so each
	// reconciles only its own Workflows).
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Whether the Argo server is enabled (spec default true); the server
	// handles are empty when it is not.
	ServerEnabled bool

	// Name of the Argo server Service — `<fullname>-server`, with the
	// fullname pinned to the resource name via fullnameOverride ("" when
	// the server is disabled).
	ServerServiceName string

	// Whether the server serves its self-signed HTTPS listener
	// (spec.server.secure); the exported endpoint follows the scheme.
	ServerSecure bool

	// In-cluster endpoint of the Argo server ("" when disabled).
	ServerKubeEndpoint string

	// Name of the ServiceAccount workflow pods run as — the identity to
	// annotate for IRSA/workload identity when workflows touch cloud
	// APIs.
	WorkflowServiceAccount string

	// kubectl one-liner for reaching the UI from a workstation ("" when
	// the server is disabled).
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesargoworkflowsv1alpha1.KubernetesArgoWorkflowsIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesArgoWorkflows.String(),
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

	namespace := spec.Namespace.GetValue()
	releaseName := target.Metadata.Name

	serverEnabled := true
	if spec.GetServer() != nil && spec.GetServer().Enabled != nil {
		serverEnabled = spec.GetServer().GetEnabled()
	}

	serverSecure := spec.GetServer().GetSecure()
	scheme := "http"
	if serverSecure {
		scheme = "https"
	}

	serverServiceName := ""
	serverKubeEndpoint := ""
	portForwardCommand := ""
	if serverEnabled {
		// fullnameOverride is pinned to metadata.name (values.go), so the
		// server Service is exactly `<name>-server`.
		serverServiceName = fmt.Sprintf("%s-server", releaseName)
		serverKubeEndpoint = fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d",
			scheme, serverServiceName, namespace, vars.ServerPort)
		portForwardCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
			serverServiceName, namespace, vars.ServerPort, vars.ServerPort)
	}

	workflowServiceAccount := spec.GetWorkflowServiceAccount()
	if workflowServiceAccount == "" {
		workflowServiceAccount = vars.DefaultWorkflowServiceAccount
	}

	return &Locals{
		Spec:                   spec,
		Labels:                 labels,
		Namespace:              namespace,
		ReleaseName:            releaseName,
		ChartVersion:           chartVersion,
		ServerEnabled:          serverEnabled,
		ServerServiceName:      serverServiceName,
		ServerSecure:           serverSecure,
		ServerKubeEndpoint:     serverKubeEndpoint,
		WorkflowServiceAccount: workflowServiceAccount,
		PortForwardCommand:     portForwardCommand,
	}
}

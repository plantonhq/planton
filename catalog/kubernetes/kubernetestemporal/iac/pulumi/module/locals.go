package module

import (
	"fmt"
	"strconv"

	kubernetestemporalv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestemporal/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the stack input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetestemporalv1alpha1.KubernetesTemporalSpec

	// Resource-identity labels: stamped on the module-created satellites
	// (the namespace) and on every pod the chart runs, through the chart's
	// own pod-label values (server.podLabels, web.podLabels,
	// admintools.podLabels, schema.podLabels) -- so a log line, a metric or
	// an alert from any Temporal pod names its organization and environment.
	// The chart's selectors are its own fixed labels; these never reach them.
	Labels map[string]string

	// Namespace Temporal installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name:
	// several Temporal clusters can coexist in one Kubernetes cluster
	// (each is a fully isolated workflow engine).
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Whether the Web UI deploys (spec default true); the web handles
	// are empty when it does not.
	WebUiEnabled bool

	// Name of the frontend Service — `<fullname>-frontend`, with the
	// fullname pinned to the resource name via fullnameOverride.
	FrontendServiceName string

	// In-cluster frontend gRPC endpoint (host:port — what Temporal SDK
	// clients and workers dial).
	FrontendEndpoint string

	// In-cluster frontend HTTP API endpoint.
	FrontendHttpEndpoint string

	// Name of the Web UI Service (`<fullname>-web`; "" when disabled).
	WebUiServiceName string

	// In-cluster Web UI endpoint ("" when disabled).
	WebUiEndpoint string

	// kubectl one-liners for workstation access.
	PortForwardFrontendCommand string
	PortForwardWebUiCommand    string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, stackInput *kubernetestemporalv1alpha1.KubernetesTemporalStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesTemporal.String(),
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

	webUiEnabled := true
	if spec.GetWebUi() != nil && spec.GetWebUi().Enabled != nil {
		webUiEnabled = spec.GetWebUi().GetEnabled()
	}

	// fullnameOverride pins the fullname to metadata.name (values.go),
	// so the frontend Service is exactly `<name>-frontend` and the Web
	// UI Service `<name>-web`.
	frontendServiceName := fmt.Sprintf("%s-frontend", releaseName)
	frontendEndpoint := fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		frontendServiceName, namespace, vars.FrontendGrpcPort)
	frontendHttpEndpoint := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		frontendServiceName, namespace, vars.FrontendHttpPort)
	portForwardFrontendCommand := fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
		frontendServiceName, namespace, vars.FrontendGrpcPort, vars.FrontendGrpcPort)

	webUiServiceName := ""
	webUiEndpoint := ""
	portForwardWebUiCommand := ""
	if webUiEnabled {
		webUiServiceName = fmt.Sprintf("%s-web", releaseName)
		webUiEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
			webUiServiceName, namespace, vars.WebPort)
		portForwardWebUiCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
			webUiServiceName, namespace, vars.WebPort, vars.WebPort)
	}

	return &Locals{
		Spec:                       spec,
		Labels:                     labels,
		Namespace:                  namespace,
		ReleaseName:                releaseName,
		ChartVersion:               chartVersion,
		WebUiEnabled:               webUiEnabled,
		FrontendServiceName:        frontendServiceName,
		FrontendEndpoint:           frontendEndpoint,
		FrontendHttpEndpoint:       frontendHttpEndpoint,
		WebUiServiceName:           webUiServiceName,
		WebUiEndpoint:              webUiEndpoint,
		PortForwardFrontendCommand: portForwardFrontendCommand,
		PortForwardWebUiCommand:    portForwardWebUiCommand,
	}
}

package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetestektonv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetestekton/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals keeps all frequently-used values in one place – similar to a Terraform "locals {}" block.
type Locals struct {
	Labels           map[string]string
	KubernetesTekton *kubernetestektonv1.KubernetesTekton

	// Computed values
	Namespace                 string
	PipelineVersion           string
	DashboardEnabled          bool
	DashboardVersion          string
	PipelineManifestURL       string
	DashboardManifestURL      string
	DashboardInternalEndpoint string
	PortForwardDashboardCmd   string
	CloudEventsSinkURL        string

	// Ingress-related
	IngressEnabled        bool
	IngressHostname       string
	CertSecretName        string
	GatewayName           string
	HttpRedirectRouteName string
	HttpsRouteName        string
	ClusterIssuerName     string
}

// Output key constants for stack exports
const (
	OpNamespace                 = "namespace"
	OpPipelineVersion           = "pipeline_version"
	OpDashboardVersion          = "dashboard_version"
	OpDashboardInternalEndpoint = "dashboard_internal_endpoint"
	OpDashboardExternalHostname = "dashboard_external_hostname"
	OpPortForwardDashboardCmd   = "port_forward_dashboard_command"
	OpCloudEventsSinkURL        = "cloud_events_sink_url"
)

// initializeLocals builds the Locals struct and exports stack outputs.
func initializeLocals(ctx *pulumi.Context,
	stackInput *kubernetestektonv1.KubernetesTektonStackInput) *Locals {

	locals := &Locals{}
	locals.KubernetesTekton = stackInput.Target
	target := stackInput.Target

	// -------------------------------- labels ---------------------------------
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesTekton.String(),
	}

	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// ---------------------------- namespace ----------------------------------
	// Tekton always uses tekton-pipelines namespace
	locals.Namespace = vars.TektonNamespace
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// ---------------------------- versions -----------------------------------
	locals.PipelineVersion = target.Spec.PipelineVersion
	if locals.PipelineVersion == "" {
		locals.PipelineVersion = "latest"
	}
	ctx.Export(OpPipelineVersion, pulumi.String(locals.PipelineVersion))

	// Check if dashboard is enabled
	locals.DashboardEnabled = target.Spec.Dashboard != nil && target.Spec.Dashboard.Enabled

	// Get dashboard version
	if target.Spec.Dashboard != nil && target.Spec.Dashboard.Version != "" {
		locals.DashboardVersion = target.Spec.Dashboard.Version
	} else {
		locals.DashboardVersion = "latest"
	}
	if locals.DashboardEnabled {
		ctx.Export(OpDashboardVersion, pulumi.String(locals.DashboardVersion))
	}

	// ---------------------------- manifest URLs ------------------------------
	locals.PipelineManifestURL = fmt.Sprintf(vars.PipelineReleaseURLTemplate, locals.PipelineVersion)
	locals.DashboardManifestURL = fmt.Sprintf(vars.DashboardReleaseURLTemplate, locals.DashboardVersion)

	// ---------------------------- endpoints ----------------------------------
	locals.DashboardInternalEndpoint = fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		vars.DashboardServiceName, locals.Namespace, vars.DashboardServicePort)

	if locals.DashboardEnabled {
		ctx.Export(OpDashboardInternalEndpoint, pulumi.String(locals.DashboardInternalEndpoint))
	}

	// ---------------------------- port-forward cmd ---------------------------
	locals.PortForwardDashboardCmd = fmt.Sprintf(
		"kubectl port-forward -n %s service/%s 9097:9097",
		locals.Namespace, vars.DashboardServiceName)

	if locals.DashboardEnabled {
		ctx.Export(OpPortForwardDashboardCmd, pulumi.String(locals.PortForwardDashboardCmd))
	}

	// ---------------------------- cloud events -------------------------------
	if target.Spec.CloudEvents != nil && target.Spec.CloudEvents.SinkUrl != "" {
		locals.CloudEventsSinkURL = target.Spec.CloudEvents.SinkUrl
		ctx.Export(OpCloudEventsSinkURL, pulumi.String(locals.CloudEventsSinkURL))
	}

	// ------------------------------- ingress ---------------------------------
	if target.Spec.Dashboard != nil &&
		target.Spec.Dashboard.Ingress != nil &&
		target.Spec.Dashboard.Ingress.Enabled &&
		target.Spec.Dashboard.Ingress.Hostname != "" {

		locals.IngressEnabled = true
		locals.IngressHostname = target.Spec.Dashboard.Ingress.Hostname
		ctx.Export(OpDashboardExternalHostname, pulumi.String(locals.IngressHostname))

		// Extract domain from hostname for ClusterIssuer name
		hostnameParts := strings.Split(locals.IngressHostname, ".")
		if len(hostnameParts) > 1 {
			locals.ClusterIssuerName = strings.Join(hostnameParts[1:], ".")
		}

		// Computed resource names
		locals.CertSecretName = fmt.Sprintf("%s-dashboard-cert", target.Metadata.Name)
		locals.GatewayName = fmt.Sprintf("%s-dashboard-external", target.Metadata.Name)
		locals.HttpRedirectRouteName = fmt.Sprintf("%s-dashboard-http-redirect", target.Metadata.Name)
		locals.HttpsRouteName = fmt.Sprintf("%s-dashboard-https", target.Metadata.Name)
	}

	return locals
}

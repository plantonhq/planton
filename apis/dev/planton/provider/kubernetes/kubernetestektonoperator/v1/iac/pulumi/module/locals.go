package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetestektonoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetestektonoperator/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesTektonOperator *kubernetestektonoperatorv1.KubernetesTektonOperator
	KubeLabels               map[string]string
	OperatorNamespace        string
	ComponentsNamespace      string
	TektonConfigName         string
	OperatorReleaseURL       string

	// Component enablement
	EnablePipelines bool
	EnableTriggers  bool
	EnableDashboard bool

	// CloudEvents configuration
	CloudEventsSinkURL string

	// Dashboard ingress configuration
	IngressEnabled        bool
	IngressHostname       string
	CertSecretName        string
	GatewayName           string
	HttpRedirectRouteName string
	HttpsRouteName        string
	ClusterIssuerName     string
}

func initializeLocals(ctx *pulumi.Context, in *kubernetestektonoperatorv1.KubernetesTektonOperatorStackInput) *Locals {
	var l Locals
	l.KubernetesTektonOperator = in.Target

	l.KubeLabels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: in.Target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesTektonOperator.String(),
	}

	if id := in.Target.Metadata.Id; id != "" {
		l.KubeLabels[kuberneteslabelkeys.ResourceId] = id
	}
	if org := in.Target.Metadata.Org; org != "" {
		l.KubeLabels[kuberneteslabelkeys.Organization] = org
	}
	if env := in.Target.Metadata.Env; env != "" {
		l.KubeLabels[kuberneteslabelkeys.Environment] = env
	}

	// Tekton Operator uses fixed namespaces managed by the operator itself:
	// - 'tekton-operator' for the operator
	// - 'tekton-pipelines' for components (Pipelines, Triggers, Dashboard)
	// These cannot be customized by the user.
	l.OperatorNamespace = vars.OperatorNamespace
	l.ComponentsNamespace = vars.ComponentsNamespace
	l.TektonConfigName = vars.TektonConfigName

	// Compute operator release URL from version, falling back to the documented
	// default when the spec field is empty. Proto documents v0.78.0 as the
	// default but the generated Go code does not substitute it at runtime.
	operatorVersion := in.Target.Spec.OperatorVersion
	if operatorVersion == "" {
		operatorVersion = "v0.78.0"
	}
	l.OperatorReleaseURL = fmt.Sprintf(vars.OperatorReleaseURLFormat, operatorVersion)

	// Get component enablement from spec
	if comp := in.Target.Spec.Components; comp != nil {
		l.EnablePipelines = comp.Pipelines
		l.EnableTriggers = comp.Triggers
		l.EnableDashboard = comp.Dashboard
	}

	// Export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(l.ComponentsNamespace))
	ctx.Export(OpTektonConfigName, pulumi.String(l.TektonConfigName))

	// Export service names based on enabled components
	if l.EnablePipelines {
		ctx.Export(OpPipelinesControllerService, pulumi.String("tekton-pipelines-controller"))
	}
	if l.EnableTriggers {
		ctx.Export(OpTriggersControllerService, pulumi.String("tekton-triggers-controller"))
	}
	if l.EnableDashboard {
		ctx.Export(OpDashboardService, pulumi.String("tekton-dashboard"))
		ctx.Export(OpDashboardPortForwardCommand, pulumi.String(
			fmt.Sprintf("kubectl port-forward svc/tekton-dashboard -n %s 9097:9097", l.ComponentsNamespace)))
	}

	// CloudEvents configuration
	if in.Target.Spec.CloudEventsSinkUrl != "" {
		l.CloudEventsSinkURL = in.Target.Spec.CloudEventsSinkUrl
		ctx.Export(OpCloudEventsSinkURL, pulumi.String(l.CloudEventsSinkURL))
	}

	// Dashboard ingress configuration
	if in.Target.Spec.DashboardIngress != nil &&
		in.Target.Spec.DashboardIngress.Enabled &&
		in.Target.Spec.DashboardIngress.Hostname != "" {

		l.IngressEnabled = true
		l.IngressHostname = in.Target.Spec.DashboardIngress.Hostname
		ctx.Export(OpDashboardExternalHostname, pulumi.String(l.IngressHostname))

		// Extract domain from hostname for ClusterIssuer name
		// e.g., "tekton-dashboard.example.com" -> "example.com"
		hostnameParts := strings.Split(l.IngressHostname, ".")
		if len(hostnameParts) > 1 {
			l.ClusterIssuerName = strings.Join(hostnameParts[1:], ".")
		}

		// Computed resource names
		l.CertSecretName = fmt.Sprintf("%s-dashboard-cert", in.Target.Metadata.Name)
		l.GatewayName = fmt.Sprintf("%s-dashboard-external", in.Target.Metadata.Name)
		l.HttpRedirectRouteName = fmt.Sprintf("%s-dashboard-http-redirect", in.Target.Metadata.Name)
		l.HttpsRouteName = fmt.Sprintf("%s-dashboard-https", in.Target.Metadata.Name)
	}

	return &l
}

package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetestemporalv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetestemporal/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals keeps all frequently-used values in one place – similar to a
// Terraform “locals {}” block.
type Locals struct {
	Namespace                   string
	Labels                      map[string]string
	KubernetesTemporal          *kubernetestemporalv1.KubernetesTemporal
	FrontendServiceName         string
	UIServiceName               string
	FrontendEndpoint            string
	UIEndpoint                  string
	PortForwardFrontendCmd      string
	PortForwardUICmd            string
	IngressFrontendGrpcHostname string
	IngressFrontendHttpHostname string
	IngressUIHostname           string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	DatabasePasswordSecretName    string
	FrontendGrpcLbServiceName     string
	FrontendHttpCertSecretName    string
	FrontendHttpGatewayName       string
	FrontendHttpRedirectRouteName string
	FrontendHttpsRouteName        string
	UiCertSecretName              string
	UiGatewayName                 string
	UiHttpRedirectRouteName       string
	UiHttpsRouteName              string
}

// initializeLocals builds the Locals struct and immediately exports the
// values required by KubernetesTemporalStackOutputs.
func initializeLocals(ctx *pulumi.Context,
	stackInput *kubernetestemporalv1.KubernetesTemporalStackInput) *Locals {

	locals := &Locals{}
	locals.KubernetesTemporal = stackInput.Target
	target := stackInput.Target

	// -------------------------------- labels ---------------------------------
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesTemporal.String(),
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

	// get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// ---------------------------- service names ------------------------------
	locals.FrontendServiceName = fmt.Sprintf("%s-frontend", target.Metadata.Name)
	locals.UIServiceName = fmt.Sprintf("%s-web", target.Metadata.Name)

	ctx.Export(OpFrontendService, pulumi.String(locals.FrontendServiceName))
	ctx.Export(OpUIService, pulumi.String(locals.UIServiceName))

	// --------------------------- cluster endpoints ---------------------------
	locals.FrontendEndpoint = fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		locals.FrontendServiceName, locals.Namespace, vars.FrontendGrpcPort)
	locals.UIEndpoint = fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		locals.UIServiceName, locals.Namespace, vars.UIPort)

	ctx.Export(OpFrontendEndpoint, pulumi.String(locals.FrontendEndpoint))
	ctx.Export(OpWebUiEndpoint, pulumi.String(locals.UIEndpoint))

	// --------------------------- port-forward cmds ---------------------------
	locals.PortForwardFrontendCmd = fmt.Sprintf(
		"kubectl port-forward -n %s service/%s 7233:7233",
		locals.Namespace, locals.FrontendServiceName)
	locals.PortForwardUICmd = fmt.Sprintf(
		"kubectl port-forward -n %s service/%s 8080:8080",
		locals.Namespace, locals.UIServiceName)

	ctx.Export(OpPortForwardFrontendCommand, pulumi.String(locals.PortForwardFrontendCmd))
	ctx.Export(OpPortForwardUICommand, pulumi.String(locals.PortForwardUICmd))

	// ------------------------------- ingress ---------------------------------
	// Frontend ingress
	if target.Spec.Ingress != nil &&
		target.Spec.Ingress.Frontend != nil &&
		target.Spec.Ingress.Frontend.Enabled {

		if target.Spec.Ingress.Frontend.GrpcHostname != "" {
			locals.IngressFrontendGrpcHostname = target.Spec.Ingress.Frontend.GrpcHostname
			ctx.Export(OpExternalFrontendHostname, pulumi.String(locals.IngressFrontendGrpcHostname))
		}

		if target.Spec.Ingress.Frontend.HttpHostname != "" {
			locals.IngressFrontendHttpHostname = target.Spec.Ingress.Frontend.HttpHostname
		}
	}

	// Web UI ingress
	if target.Spec.Ingress != nil &&
		target.Spec.Ingress.WebUi != nil &&
		target.Spec.Ingress.WebUi.Enabled &&
		target.Spec.Ingress.WebUi.Hostname != "" {

		locals.IngressUIHostname = target.Spec.Ingress.WebUi.Hostname
		ctx.Export(OpExternalUIHostname, pulumi.String(locals.IngressUIHostname))
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	locals.DatabasePasswordSecretName = fmt.Sprintf("%s-db-password", target.Metadata.Name)
	locals.FrontendGrpcLbServiceName = fmt.Sprintf("%s-frontend-grpc-lb", target.Metadata.Name)

	// Gateway API resource names derived from hostname (flattened: dots replaced with dashes)
	// Since these resources are created in istio-ingress namespace, using full domain names
	// makes them easily identifiable (e.g., "temporal-app-prod-main-ui-planton-live-external")
	if locals.IngressFrontendHttpHostname != "" {
		flattenedFrontendHttp := flattenHostname(locals.IngressFrontendHttpHostname)
		locals.FrontendHttpCertSecretName = flattenedFrontendHttp
		locals.FrontendHttpGatewayName = fmt.Sprintf("%s-external", flattenedFrontendHttp)
		locals.FrontendHttpRedirectRouteName = fmt.Sprintf("%s-redirect", flattenedFrontendHttp)
		locals.FrontendHttpsRouteName = fmt.Sprintf("%s-https", flattenedFrontendHttp)
	} else {
		locals.FrontendHttpCertSecretName = fmt.Sprintf("%s-frontend-http-cert", target.Metadata.Name)
		locals.FrontendHttpGatewayName = fmt.Sprintf("%s-frontend-http-external", target.Metadata.Name)
		locals.FrontendHttpRedirectRouteName = fmt.Sprintf("%s-frontend-http-redirect", target.Metadata.Name)
		locals.FrontendHttpsRouteName = fmt.Sprintf("%s-frontend-https", target.Metadata.Name)
	}

	if locals.IngressUIHostname != "" {
		flattenedUi := flattenHostname(locals.IngressUIHostname)
		locals.UiCertSecretName = flattenedUi
		locals.UiGatewayName = fmt.Sprintf("%s-external", flattenedUi)
		locals.UiHttpRedirectRouteName = fmt.Sprintf("%s-redirect", flattenedUi)
		locals.UiHttpsRouteName = fmt.Sprintf("%s-https", flattenedUi)
	} else {
		locals.UiCertSecretName = fmt.Sprintf("%s-ui-cert", target.Metadata.Name)
		locals.UiGatewayName = fmt.Sprintf("%s-ui-external", target.Metadata.Name)
		locals.UiHttpRedirectRouteName = fmt.Sprintf("%s-ui-http-redirect", target.Metadata.Name)
		locals.UiHttpsRouteName = fmt.Sprintf("%s-ui-https", target.Metadata.Name)
	}

	return locals
}

// flattenHostname converts a hostname to a valid Kubernetes resource name
// by replacing dots with dashes (e.g., "temporal-ui.planton.live" -> "temporal-ui-planton-live")
func flattenHostname(hostname string) string {
	return strings.ReplaceAll(hostname, ".", "-")
}

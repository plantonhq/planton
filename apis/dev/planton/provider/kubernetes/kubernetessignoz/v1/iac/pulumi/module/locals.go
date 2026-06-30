package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetessignozv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetessignoz/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesSignoz                  *kubernetessignozv1.KubernetesSignoz
	Namespace                         string
	KubernetesLabels                  map[string]string
	SignozServiceName                 string
	OtelCollectorServiceName          string
	SignozKubeServiceFqdn             string
	OtelCollectorGrpcFqdn             string
	OtelCollectorHttpFqdn             string
	KubePortForwardCommand            string
	IngressExternalHostname           string
	IngressHostnames                  []string
	IngressCertClusterIssuerName      string
	IngressCertSecretName             string
	OtelCollectorExternalHttpHostname string
	ClickhouseEndpoint                string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	SignozCertificateName       string
	SignozGatewayName           string
	SignozHTTPRouteName         string
	SignozHttpRedirectRouteName string
	OtelCertificateName         string
	OtelGatewayName             string
	OtelHTTPRouteName           string
	OtelHttpRedirectRouteName   string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetessignozv1.KubernetesSignozStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesSignoz = stackInput.Target
	target := stackInput.Target

	locals.KubernetesLabels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesSignoz.String(),
	}

	if target.Metadata.Id != "" {
		locals.KubernetesLabels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.KubernetesLabels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.KubernetesLabels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// Get namespace from spec (required field)
	locals.Namespace = target.Spec.Namespace.GetValue()

	//export namespace
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Service names
	locals.SignozServiceName = fmt.Sprintf("%s-signoz", target.Metadata.Name)
	locals.OtelCollectorServiceName = fmt.Sprintf("%s-otel-collector", target.Metadata.Name)

	//export service names
	ctx.Export(OpSignozService, pulumi.String(locals.SignozServiceName))
	ctx.Export(OpOtelCollectorService, pulumi.String(locals.OtelCollectorServiceName))

	// Kubernetes FQDNs
	locals.SignozKubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		locals.SignozServiceName, locals.Namespace, vars.SignozUIPort)
	locals.OtelCollectorGrpcFqdn = fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		locals.OtelCollectorServiceName, locals.Namespace, vars.OtelGrpcPort)
	locals.OtelCollectorHttpFqdn = fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		locals.OtelCollectorServiceName, locals.Namespace, vars.OtelHttpPort)

	//export kubernetes endpoints
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.SignozKubeServiceFqdn))
	ctx.Export(OpOtelCollectorGrpcEndpoint, pulumi.String(locals.OtelCollectorGrpcFqdn))
	ctx.Export(OpOtelCollectorHttpEndpoint, pulumi.String(locals.OtelCollectorHttpFqdn))

	// Port forward command
	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s %d:%d",
		locals.Namespace, locals.SignozServiceName, vars.SignozUIPort, vars.SignozUIPort)

	//export kube-port-forward command
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	// ClickHouse outputs (only for self-managed)
	if target.Spec.Database != nil && !target.Spec.Database.IsExternal {
		locals.ClickhouseEndpoint = fmt.Sprintf("%s-clickhouse.%s.svc.cluster.local:8123",
			target.Metadata.Name, locals.Namespace)
		ctx.Export(OpClickhouseEndpoint, pulumi.String(locals.ClickhouseEndpoint))
		ctx.Export(OpClickhouseUsername, pulumi.String("admin"))
		ctx.Export(OpClickhousePasswordSecretName, pulumi.String(fmt.Sprintf("%s-clickhouse", target.Metadata.Name)))
		ctx.Export(OpClickhousePasswordSecretKey, pulumi.String("admin-password"))
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	locals.SignozCertificateName = fmt.Sprintf("%s-signoz-cert", target.Metadata.Name)
	locals.SignozGatewayName = fmt.Sprintf("%s-signoz-external-gateway", target.Metadata.Name)
	locals.SignozHTTPRouteName = fmt.Sprintf("%s-signoz-https-route", target.Metadata.Name)
	locals.SignozHttpRedirectRouteName = fmt.Sprintf("%s-signoz-http-redirect", target.Metadata.Name)
	locals.OtelCertificateName = fmt.Sprintf("%s-otel-http-cert", target.Metadata.Name)
	locals.OtelGatewayName = fmt.Sprintf("%s-otel-http-external-gateway", target.Metadata.Name)
	locals.OtelHTTPRouteName = fmt.Sprintf("%s-otel-https-route", target.Metadata.Name)
	locals.OtelHttpRedirectRouteName = fmt.Sprintf("%s-otel-http-redirect", target.Metadata.Name)

	// Ingress configuration for SigNoz UI
	if target.Spec.Ingress != nil &&
		target.Spec.Ingress.Ui != nil &&
		target.Spec.Ingress.Ui.Enabled &&
		target.Spec.Ingress.Ui.Hostname != "" {
		locals.IngressExternalHostname = target.Spec.Ingress.Ui.Hostname
		ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

		locals.IngressHostnames = []string{
			locals.IngressExternalHostname,
		}

		// ClusterIssuer should already exist on the cluster
		// Extract domain from hostname for ClusterIssuer name
		// Typically managed by cluster administrator or created by Planton
		hostnameParts := strings.Split(locals.IngressExternalHostname, ".")
		if len(hostnameParts) > 1 {
			locals.IngressCertClusterIssuerName = strings.Join(hostnameParts[1:], ".")
		}

		// Use computed name to avoid conflicts when multiple instances share a namespace
		locals.IngressCertSecretName = locals.SignozCertificateName
	}

	// Ingress configuration for OTel Collector
	if target.Spec.Ingress != nil &&
		target.Spec.Ingress.OtelCollector != nil &&
		target.Spec.Ingress.OtelCollector.Enabled &&
		target.Spec.Ingress.OtelCollector.Hostname != "" {
		locals.OtelCollectorExternalHttpHostname = target.Spec.Ingress.OtelCollector.Hostname
		ctx.Export(OpOtelCollectorExternalHttpHostname, pulumi.String(locals.OtelCollectorExternalHttpHostname))
	}

	return locals
}

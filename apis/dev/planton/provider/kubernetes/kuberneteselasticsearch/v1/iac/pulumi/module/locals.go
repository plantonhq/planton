package module

import (
	"fmt"
	"strconv"
	"strings"

	kuberneteselasticsearchv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteselasticsearch/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	ElasticsearchIngressExternalHostname string
	ElasticsearchKubePortForwardCommand  string
	ElasticsearchKubeServiceFqdn         string
	ElasticsearchKubeServiceName         string
	Namespace                            string
	KubernetesElasticsearch              *kuberneteselasticsearchv1.KubernetesElasticsearch
	KibanaIngressExternalHostname        string
	KibanaKubePortForwardCommand         string
	KibanaKubeServiceFqdn                string
	KibanaKubeServiceName                string
	IngressHostnames                     []string
	IngressCertClusterIssuerName         string
	IngressCertSecretName                string
	Labels                               map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	IngressCertificateName             string
	ElasticsearchExternalGatewayName   string
	ElasticsearchHttpRedirectRouteName string
	ElasticsearchHttpsRouteName        string
	KibanaExternalGatewayName          string
	KibanaHttpRedirectRouteName        string
	KibanaHttpsRouteName               string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kuberneteselasticsearchv1.KubernetesElasticsearchStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesElasticsearch = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesElasticsearch.String(),
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

	ctx.Export(OpElasticsearchUsername, pulumi.String("elastic"))
	ctx.Export(OpElasticsearchPasswordSecretName, pulumi.Sprintf("%s-es-elastic-user", target.Metadata.Name))
	ctx.Export(OpElasticsearchPasswordSecretKey, pulumi.String("elastic"))

	locals.ElasticsearchKubeServiceName = fmt.Sprintf("%s-es-http", target.Metadata.Name)

	//export kubernetes service name
	ctx.Export(OpElasticsearchService, pulumi.String(locals.ElasticsearchKubeServiceName))

	locals.ElasticsearchKubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.ElasticsearchKubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(OpElasticsearchKubeEndpoint, pulumi.String(locals.ElasticsearchKubeServiceFqdn))

	locals.ElasticsearchKubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s %d:%d",
		locals.Namespace, locals.ElasticsearchKubeServiceName, vars.ElasticsearchPort, vars.ElasticsearchPort)

	//export kube-port-forward command
	ctx.Export(OpElasticsearchPortForwardCommand, pulumi.String(locals.ElasticsearchKubePortForwardCommand))

	locals.KibanaKubeServiceName = fmt.Sprintf("%s-kb-http", target.Metadata.Name)

	//export kubernetes service name
	ctx.Export(OpKibanaService, pulumi.String(locals.KibanaKubeServiceName))

	locals.KibanaKubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KibanaKubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(OpKibanaKubeEndpoint, pulumi.String(locals.KibanaKubeServiceFqdn))

	locals.KibanaKubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s %d:%d",
		locals.Namespace, locals.KibanaKubeServiceName, vars.KibanaPort, vars.KibanaPort)

	//export kube-port-forward command
	ctx.Export(OpKibanaPortForwardCommand, pulumi.String(locals.KibanaKubePortForwardCommand))

	// Elasticsearch ingress
	if target.Spec.Elasticsearch.Ingress != nil &&
		target.Spec.Elasticsearch.Ingress.Enabled &&
		target.Spec.Elasticsearch.Ingress.Hostname != "" {

		locals.ElasticsearchIngressExternalHostname = target.Spec.Elasticsearch.Ingress.Hostname
		ctx.Export(OpElasticsearchExternalHostname, pulumi.String(locals.ElasticsearchIngressExternalHostname))
		locals.IngressHostnames = append(locals.IngressHostnames, locals.ElasticsearchIngressExternalHostname)
	}

	// Kibana ingress
	if target.Spec.Kibana != nil && target.Spec.Kibana.Enabled &&
		target.Spec.Kibana.Ingress != nil &&
		target.Spec.Kibana.Ingress.Enabled &&
		target.Spec.Kibana.Ingress.Hostname != "" {

		locals.KibanaIngressExternalHostname = target.Spec.Kibana.Ingress.Hostname
		ctx.Export(OpKibanaExternalHostname, pulumi.String(locals.KibanaIngressExternalHostname))
		locals.IngressHostnames = append(locals.IngressHostnames, locals.KibanaIngressExternalHostname)
	}

	// Set certificate issuer and secret name if any ingress is enabled
	if len(locals.IngressHostnames) > 0 {
		// Use first available hostname to extract domain for cert issuer
		firstHostname := locals.ElasticsearchIngressExternalHostname
		if firstHostname == "" {
			firstHostname = locals.KibanaIngressExternalHostname
		}
		// Extract domain from hostname for cert issuer
		parts := strings.Split(firstHostname, ".")
		if len(parts) > 1 {
			locals.IngressCertClusterIssuerName = strings.Join(parts[1:], ".")
		}
		// Use metadata.name prefix for cert secret to avoid conflicts in shared namespaces
		locals.IngressCertSecretName = fmt.Sprintf("%s-ingress-cert", target.Metadata.Name)
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	locals.IngressCertificateName = fmt.Sprintf("%s-ingress-cert", target.Metadata.Name)
	locals.ElasticsearchExternalGatewayName = fmt.Sprintf("%s-es-external-gateway", target.Metadata.Name)
	locals.ElasticsearchHttpRedirectRouteName = fmt.Sprintf("%s-es-http-redirect", target.Metadata.Name)
	locals.ElasticsearchHttpsRouteName = fmt.Sprintf("%s-es-https-route", target.Metadata.Name)
	locals.KibanaExternalGatewayName = fmt.Sprintf("%s-kb-external-gateway", target.Metadata.Name)
	locals.KibanaHttpRedirectRouteName = fmt.Sprintf("%s-kb-http-redirect", target.Metadata.Name)
	locals.KibanaHttpsRouteName = fmt.Sprintf("%s-kb-https-route", target.Metadata.Name)

	return locals
}

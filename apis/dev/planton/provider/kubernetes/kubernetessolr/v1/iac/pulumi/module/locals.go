package module

import (
	"fmt"
	"strconv"

	kubernetessolrv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetessolr/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressCertClusterIssuerName string
	IngressCertSecretName        string
	IngressExternalHostname      string
	IngressHostnames             []string
	IngressInternalHostname      string
	KubePortForwardCommand       string
	KubeServiceFqdn              string
	KubeServiceName              string
	Namespace                    string
	KubernetesSolr               *kubernetessolrv1.KubernetesSolr
	Labels                       map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	CertificateName               string
	ExternalGatewayName           string
	HttpExternalRedirectRouteName string
	HttpsExternalRouteName        string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetessolrv1.KubernetesSolrStackInput) *Locals {
	locals := &Locals{}
	//assign value for the locals variable to make it available across the project
	locals.KubernetesSolr = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesSolr.String(),
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

	// Get namespace from spec (required field)
	locals.Namespace = target.Spec.Namespace.GetValue()

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	locals.KubeServiceName = fmt.Sprintf("%s-solrcloud-common", target.Metadata.Name)

	//export kubernetes service name
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf(
		"%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:8080",
		locals.Namespace, target.Metadata.Name)

	//export kube-port-forward command
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		return locals
	}

	// Use the hostname directly from spec
	locals.IngressExternalHostname = target.Spec.Ingress.Hostname

	// Internal hostname (private ingress) - prepend internal-
	locals.IngressInternalHostname = fmt.Sprintf("internal-%s", target.Spec.Ingress.Hostname)

	locals.IngressHostnames = []string{
		locals.IngressExternalHostname,
		locals.IngressInternalHostname,
	}

	//export ingress hostnames
	ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))
	ctx.Export(OpInternalHostname, pulumi.String(locals.IngressInternalHostname))

	//note: a ClusterIssuer resource should have already exist on the kubernetes-cluster.
	//this is typically taken care of by the kubernetes cluster administrator.
	//if the kubernetes-cluster is created using Planton, then the cluster-issuer name will be
	//same as the ingress-domain-name as long as the same ingress-domain-name is added to the list of
	//ingress-domain-names for the GkeCluster/EksCluster/AksCluster spec.
	// Extract the domain from hostname for certificate issuer name
	dnsDomain := extractDomainFromHostname(target.Spec.Ingress.Hostname)
	locals.IngressCertClusterIssuerName = dnsDomain

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	locals.IngressCertSecretName = fmt.Sprintf("%s-cert-tls", target.Metadata.Name)
	locals.CertificateName = fmt.Sprintf("%s-cert", target.Metadata.Name)
	locals.ExternalGatewayName = fmt.Sprintf("%s-external", target.Metadata.Name)
	locals.HttpExternalRedirectRouteName = fmt.Sprintf("%s-http-external-redirect", target.Metadata.Name)
	locals.HttpsExternalRouteName = fmt.Sprintf("%s-https-external", target.Metadata.Name)

	return locals
}

// extractDomainFromHostname extracts the domain from a hostname
// Example: "solr.example.com" -> "example.com"
func extractDomainFromHostname(hostname string) string {
	// Split by dots and take everything after the first part
	// This is a simple implementation - assumes standard domain structure
	parts := []rune(hostname)
	firstDotIndex := -1
	for i, char := range parts {
		if char == '.' {
			firstDotIndex = i
			break
		}
	}
	if firstDotIndex > 0 && firstDotIndex < len(hostname)-1 {
		return hostname[firstDotIndex+1:]
	}
	// If no dot found or dot is at the end, return the hostname as-is
	return hostname
}

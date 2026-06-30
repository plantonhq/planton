package module

import (
	"fmt"
	"strconv"

	kubernetesjenkinsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesjenkins/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesJenkins            *kubernetesjenkinsv1.KubernetesJenkins
	Namespace                    string
	IngressCertClusterIssuerName string
	IngressCertSecretName        string
	IngressInternalHostname      string
	IngressExternalHostname      string
	IngressHostnames             []string
	KubeServiceFqdn              string
	KubeServiceName              string
	KubePortForwardCommand       string
	Labels                       map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	AdminCredentialsSecretName string
	IngressCertificateName     string
	ExternalGatewayName        string
	HttpRedirectRouteName      string
	HttpsRouteName             string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesjenkinsv1.KubernetesJenkinsStackInput) *Locals {
	locals := &Locals{}
	//assign value for the local variable to make it available across the project
	locals.KubernetesJenkins = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesJenkins.String(),
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

	locals.KubeServiceName = target.Metadata.Name

	//export kubernetes service name
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "jenkins-my-ci")
	locals.AdminCredentialsSecretName = fmt.Sprintf("%s-admin-credentials", target.Metadata.Name)
	locals.IngressCertificateName = fmt.Sprintf("%s-ingress-cert", target.Metadata.Name)
	locals.ExternalGatewayName = fmt.Sprintf("%s-external", target.Metadata.Name)
	locals.HttpRedirectRouteName = fmt.Sprintf("%s-http-redirect", target.Metadata.Name)
	locals.HttpsRouteName = fmt.Sprintf("%s-https", target.Metadata.Name)

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local",
		target.Metadata.Name, locals.Namespace)

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

	locals.IngressCertSecretName = fmt.Sprintf("%s-tls", target.Metadata.Name)

	return locals
}

// extractDomainFromHostname extracts the domain from a hostname
// Example: "jenkins.example.com" -> "example.com"
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

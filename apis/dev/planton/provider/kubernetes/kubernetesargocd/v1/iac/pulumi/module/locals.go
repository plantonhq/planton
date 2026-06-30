package module

import (
	"fmt"
	"strconv"

	kubernetesargocdv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesargocd/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesArgocd             *kubernetesargocdv1.KubernetesArgocd
	Namespace                    string
	ServiceName                  string
	KubeServiceFqdn              string
	KubePortForwardCommand       string
	IngressCertClusterIssuerName string
	IngressCertSecretName        string
	IngressInternalHostname      string
	IngressExternalHostname      string
	IngressHostnames             []string
	Labels                       map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesargocdv1.KubernetesArgocdStackInput) *Locals {
	locals := &Locals{}

	// Assign value for the local variable to make it available across the project
	locals.KubernetesArgocd = stackInput.Target

	target := stackInput.Target

	// Build labels
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesArgocd.String(),
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

	// Namespace determination
	// Priority order:
	// 1. Spec.Namespace (required field)
	// 2. Override with stackInput if provided (for backward compatibility)
	// 3. Override with custom label if provided
	// 4. Default: "argo-" + metadata.name

	resourceId := target.Metadata.Name
	if target.Metadata.Id != "" {
		resourceId = target.Metadata.Id
	}

	// get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// Export namespace
	ctx.Export(Namespace, pulumi.String(locals.Namespace))

	// Service name follows Helm chart naming: <release-name>-argocd-server
	locals.ServiceName = fmt.Sprintf("%s-argocd-server", resourceId)

	// Export service name
	ctx.Export(Service, pulumi.String(locals.ServiceName))

	// Kubernetes service FQDN for internal cluster access
	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local",
		locals.ServiceName, locals.Namespace)

	// Export kubernetes endpoint
	ctx.Export(KubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	// Port-forward command for local access
	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:80",
		locals.Namespace, locals.ServiceName)

	// Export port-forward command
	ctx.Export(PortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	// Handle ingress configuration
	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		// Export empty hostnames if ingress is not enabled
		ctx.Export(ExternalHostname, pulumi.String(""))
		ctx.Export(InternalHostname, pulumi.String(""))
		return locals
	}

	// Use the hostname directly from spec
	locals.IngressExternalHostname = target.Spec.Ingress.Hostname

	// Internal hostname (private ingress) - append -internal to the hostname
	locals.IngressInternalHostname = fmt.Sprintf("internal-%s", target.Spec.Ingress.Hostname)

	locals.IngressHostnames = []string{
		locals.IngressExternalHostname,
		locals.IngressInternalHostname,
	}

	// Export ingress hostnames
	ctx.Export(ExternalHostname, pulumi.String(locals.IngressExternalHostname))
	ctx.Export(InternalHostname, pulumi.String(locals.IngressInternalHostname))

	// Certificate configuration
	// Note: a ClusterIssuer resource should already exist on the kubernetes-cluster
	// Extract the domain from hostname for certificate issuer name
	// For example: "argocd.example.com" -> "example.com"
	dnsDomain := extractDomainFromHostname(target.Spec.Ingress.Hostname)
	locals.IngressCertClusterIssuerName = dnsDomain
	// Computed TLS secret name to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	locals.IngressCertSecretName = fmt.Sprintf("%s-tls", target.Metadata.Name)

	return locals
}

// extractDomainFromHostname extracts the domain from a hostname
// Example: "argocd.example.com" -> "example.com"
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

package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetesopenfgav1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesopenfga/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesOpenFga            *kubernetesopenfgav1.KubernetesOpenFga
	Namespace                    string
	IngressCertClusterIssuerName string
	IngressCertSecretName        string
	IngressExternalHostname      string
	IngressHostnames             []string
	KubeServiceFqdn              string
	KubeServiceName              string
	KubePortForwardCommand       string
	Labels                       map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	IngressCertificateName       string
	IngressGatewayName           string
	IngressHttpRedirectRouteName string
	IngressHttpsRouteName        string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesopenfgav1.KubernetesOpenFgaStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesOpenFga = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesOpenFga.String(),
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

	locals.IngressExternalHostname = target.Spec.Ingress.Hostname

	locals.IngressHostnames = []string{
		locals.IngressExternalHostname,
	}

	//export ingress hostname
	ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

	//note: a ClusterIssuer resource should have already exist on the kubernetes-cluster.
	//this is typically taken care of by the kubernetes cluster administrator.
	//if the kubernetes-cluster is created using Planton, then the cluster-issuer name will be
	//same as the ingress-domain-name as long as the same ingress-domain-name is added to the list of
	//ingress-domain-names for the GkeCluster/EksCluster/AksCluster spec.
	// Extract domain from hostname for certificate issuer
	// Example: "openfga.example.com" -> "example.com"
	parts := strings.Split(locals.IngressExternalHostname, ".")
	if len(parts) > 1 {
		locals.IngressCertClusterIssuerName = strings.Join(parts[1:], ".")
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "openfga-my-instance")
	locals.IngressCertSecretName = fmt.Sprintf("%s-tls", target.Metadata.Name)
	locals.IngressCertificateName = fmt.Sprintf("%s-certificate", target.Metadata.Name)
	locals.IngressGatewayName = fmt.Sprintf("%s-external", target.Metadata.Name)
	locals.IngressHttpRedirectRouteName = fmt.Sprintf("%s-http-redirect", target.Metadata.Name)
	locals.IngressHttpsRouteName = fmt.Sprintf("%s-https", target.Metadata.Name)

	return locals
}

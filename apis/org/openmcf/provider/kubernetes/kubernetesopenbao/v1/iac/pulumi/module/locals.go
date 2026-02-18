package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetesopenbaov1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesopenbao/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	Namespace               string
	KubernetesOpenBao       *kubernetesopenbaov1.KubernetesOpenBao
	Labels                  map[string]string
	KubeServiceName         string
	KubeServiceFqdn         string
	KubePortForwardCommand  string
	IngressExternalHostname string
	HaEnabled               bool
	HaReplicas              int32
	ServerReplicas          int32
	HelmChartVersion        string

	IngressCertClusterIssuerName string
	IngressCertSecretName        string
	IngressHostnames             []string
	IngressCertificateName       string
	IngressGatewayName           string
	IngressHttpRedirectRouteName string
	IngressHttpsRouteName        string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesopenbaov1.KubernetesOpenBaoStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesOpenBao = stackInput.Target
	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesOpenBao.String(),
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

	// Get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// Export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Determine HA settings
	locals.HaEnabled = target.Spec.HighAvailability != nil && target.Spec.HighAvailability.Enabled
	if locals.HaEnabled && target.Spec.HighAvailability.Replicas != nil {
		locals.HaReplicas = *target.Spec.HighAvailability.Replicas
	} else if locals.HaEnabled {
		locals.HaReplicas = 3 // default HA replicas
	}

	// Server replicas
	if target.Spec.ServerContainer != nil {
		locals.ServerReplicas = target.Spec.ServerContainer.Replicas
	} else {
		locals.ServerReplicas = 1
	}

	// Export HA status
	ctx.Export(OpHaEnabled, pulumi.Bool(locals.HaEnabled))

	// Helm chart version
	if target.Spec.HelmChartVersion != nil && *target.Spec.HelmChartVersion != "" {
		locals.HelmChartVersion = *target.Spec.HelmChartVersion
	} else {
		locals.HelmChartVersion = vars.HelmChartVersion
	}

	// Compute service name - OpenBao Helm chart uses the release name
	locals.KubeServiceName = target.Metadata.Name

	// Export kubernetes service name
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))

	// Compute kubernetes FQDN
	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		locals.KubeServiceName, locals.Namespace, vars.OpenBaoPort)

	// Export kubernetes endpoint
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	// Compute API address
	apiAddress := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		locals.KubeServiceName, locals.Namespace, vars.OpenBaoPort)
	ctx.Export(OpApiAddress, pulumi.String(apiAddress))

	// Compute cluster address (for HA mode)
	clusterAddress := fmt.Sprintf("https://%s-0.%s-internal.%s.svc.cluster.local:%d",
		locals.KubeServiceName, locals.KubeServiceName, locals.Namespace, vars.OpenBaoClusterPort)
	ctx.Export(OpClusterAddress, pulumi.String(clusterAddress))

	// Compute port-forward command
	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s %d:%d",
		locals.Namespace, locals.KubeServiceName, vars.OpenBaoPort, vars.OpenBaoPort)

	// Export port-forward command
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	// Ingress configuration
	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		return locals
	}

	locals.IngressExternalHostname = target.Spec.Ingress.Hostname

	locals.IngressHostnames = []string{
		locals.IngressExternalHostname,
	}

	ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

	// Derive ClusterIssuer name from hostname domain.
	// Example: "openbao.example.com" -> "example.com"
	parts := strings.Split(locals.IngressExternalHostname, ".")
	if len(parts) > 1 {
		locals.IngressCertClusterIssuerName = strings.Join(parts[1:], ".")
	}

	locals.IngressCertSecretName = fmt.Sprintf("%s-tls", target.Metadata.Name)
	locals.IngressCertificateName = fmt.Sprintf("%s-certificate", target.Metadata.Name)
	locals.IngressGatewayName = fmt.Sprintf("%s-external", target.Metadata.Name)
	locals.IngressHttpRedirectRouteName = fmt.Sprintf("%s-http-redirect", target.Metadata.Name)
	locals.IngressHttpsRouteName = fmt.Sprintf("%s-https", target.Metadata.Name)

	return locals
}

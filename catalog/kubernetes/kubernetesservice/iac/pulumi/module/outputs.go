package module

import (
	"fmt"

	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesServiceOutputs field names.
const (
	OutputServiceName          = "service_name"
	OutputNamespace            = "namespace"
	OutputType                 = "type"
	OutputClusterIp            = "cluster_ip"
	OutputLoadBalancerIp       = "load_balancer_ip"
	OutputLoadBalancerHostname = "load_balancer_hostname"
	OutputKubeEndpoint         = "kube_endpoint"
	OutputPortForwardCommand   = "port_forward_command"
)

// exportOutputs exports the outputs from the created Kubernetes Service.
// Every output key is exported unconditionally (empty when not applicable) so
// both engines flatten the identical field set onto the outputs proto.
func exportOutputs(ctx *pulumi.Context, locals *Locals, service *kubernetescorev1.Service) error {
	isExternalName := locals.ServiceType == "ExternalName"
	isLoadBalancer := locals.ServiceType == "LoadBalancer"

	ctx.Export(OutputServiceName, pulumi.String(locals.Name))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OutputType, pulumi.String(locals.ServiceType))

	// The virtual IP: headless services carry the literal "None" in the API
	// object and ExternalName services carry nothing — both export empty, per
	// the outputs contract (the proto documents "empty for headless/ExternalName").
	if locals.Spec.GetHeadless() || isExternalName {
		ctx.Export(OutputClusterIp, pulumi.String(""))
	} else {
		ctx.Export(OutputClusterIp, service.Spec.ClusterIP().Elem())
	}

	// Load-balancer address handles: an LB provider populates ip OR hostname,
	// never reliably both — export each independently so DNS automation can
	// pick whichever is present. Pulumi's await logic waits for LB ingress on
	// LoadBalancer services, so index 0 exists by the time these resolve.
	if isLoadBalancer {
		firstIngress := service.Status.LoadBalancer().Ingress().Index(pulumi.Int(0))
		ctx.Export(OutputLoadBalancerIp, firstIngress.Ip().ApplyT(func(v *string) string {
			if v == nil {
				return ""
			}
			return *v
		}).(pulumi.StringOutput))
		ctx.Export(OutputLoadBalancerHostname, firstIngress.Hostname().ApplyT(func(v *string) string {
			if v == nil {
				return ""
			}
			return *v
		}).(pulumi.StringOutput))
	} else {
		ctx.Export(OutputLoadBalancerIp, pulumi.String(""))
		ctx.Export(OutputLoadBalancerHostname, pulumi.String(""))
	}

	// The in-cluster DNS name resolves for every type — for ExternalName it is
	// the very CNAME alias the service exists to provide.
	ctx.Export(OutputKubeEndpoint, pulumi.String(internalDnsName(locals.Name, locals.Namespace)))

	// Port-forward needs pods behind the service; an ExternalName alias has none.
	if isExternalName || len(locals.Spec.GetPorts()) == 0 {
		ctx.Export(OutputPortForwardCommand, pulumi.String(""))
	} else {
		firstPort := locals.Spec.GetPorts()[0].GetPort()
		ctx.Export(OutputPortForwardCommand, pulumi.String(fmt.Sprintf(
			"kubectl port-forward -n %s service/%s %d:%d",
			locals.Namespace, locals.Name, firstPort, firstPort)))
	}

	return nil
}

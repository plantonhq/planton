package module

import (
	kubernetesnetworkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesIngressOutputs field names.
const (
	OutputIngressName          = "ingress_name"
	OutputNamespace            = "namespace"
	OutputLoadBalancerIp       = "load_balancer_ip"
	OutputLoadBalancerHostname = "load_balancer_hostname"
	OutputFirstHost            = "first_host"
)

// exportOutputs exports the outputs from the created Ingress.
//
// The load-balancer handles read the object's status WITHOUT waiting for a
// controller (creation is skipAwait — see ingress.go): on a cluster where a
// controller reconciles quickly the values land on the same deploy; on a
// cluster with no controller they export empty, matching the object's real
// state. Every output key is exported unconditionally so both engines flatten
// the identical field set onto the outputs proto.
func exportOutputs(ctx *pulumi.Context, locals *Locals, ingress *kubernetesnetworkingv1.Ingress) error {
	ctx.Export(OutputIngressName, pulumi.String(locals.Name))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OutputFirstHost, pulumi.String(locals.FirstHost))

	// status.loadBalancer.ingress is empty until a controller claims the
	// object — every access is nil-tolerant.
	lbIngress := ingress.Status.LoadBalancer().Ingress()
	ctx.Export(OutputLoadBalancerIp, lbIngress.ApplyT(func(items []kubernetesnetworkingv1.IngressLoadBalancerIngress) string {
		if len(items) == 0 || items[0].Ip == nil {
			return ""
		}
		return *items[0].Ip
	}).(pulumi.StringOutput))
	ctx.Export(OutputLoadBalancerHostname, lbIngress.ApplyT(func(items []kubernetesnetworkingv1.IngressLoadBalancerIngress) string {
		if len(items) == 0 || items[0].Hostname == nil {
			return ""
		}
		return *items[0].Hostname
	}).(pulumi.StringOutput))

	return nil
}

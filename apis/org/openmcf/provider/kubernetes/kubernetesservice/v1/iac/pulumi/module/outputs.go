package module

import (
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesServiceStackOutputs field names.
const (
	OutputServiceName        = "service_name"
	OutputNamespace          = "namespace"
	OutputType               = "type"
	OutputClusterIP          = "cluster_ip"
	OutputLoadBalancerIngress = "load_balancer_ingress"
	OutputInternalDnsName    = "internal_dns_name"
)

// exportOutputs exports the stack outputs from the created Kubernetes Service resource.
func exportOutputs(ctx *pulumi.Context, locals *Locals, service *kubernetescorev1.Service) error {
	ctx.Export(OutputServiceName, pulumi.String(locals.Name))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OutputType, pulumi.String(locals.ServiceType))
	ctx.Export(OutputInternalDnsName, pulumi.String(internalDnsName(locals.Name, locals.Namespace)))

	// ClusterIP is populated by Kubernetes after the service is created.
	// For headless services, this will be "None". For ExternalName, it will be empty.
	ctx.Export(OutputClusterIP, service.Spec.ClusterIP().Elem())

	// LoadBalancer ingress is only populated for LoadBalancer-type services.
	// Extract the first ingress hostname or IP if available.
	ctx.Export(OutputLoadBalancerIngress, service.Status.LoadBalancer().Ingress().Index(pulumi.Int(0)).Hostname().Elem())

	return nil
}

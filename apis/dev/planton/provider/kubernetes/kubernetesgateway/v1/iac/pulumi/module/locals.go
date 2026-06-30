package module

import (
	kubernetesgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesGateway *kubernetesgatewayv1.KubernetesGateway
	// GatewayName is the Kubernetes resource name of the Gateway. It equals the
	// Planton resource metadata.name.
	GatewayName string
	// Namespace is the resolved namespace the Gateway is created in.
	Namespace string
	// GatewayClassName is the resolved name of the GatewayClass this Gateway
	// belongs to.
	GatewayClassName string
	Labels           map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetesgatewayv1.KubernetesGatewayStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	gatewayName := metadata.Name

	// namespace and gateway_class_name are StringValueOrRef foreign keys. The
	// platform middleware resolves valueFrom references to literal strings
	// before the IaC module runs, so GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()
	gatewayClassName := spec.GetGatewayClassName().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "gateway",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "gateway",
	}

	return &Locals{
		KubernetesGateway: target,
		GatewayName:       gatewayName,
		Namespace:         namespace,
		GatewayClassName:  gatewayClassName,
		Labels:            labels,
	}
}

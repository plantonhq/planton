package module

import (
	kubernetesgatewayclassv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesgatewayclass/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesGatewayClass *kubernetesgatewayclassv1.KubernetesGatewayClass
	// GatewayClassName is the cluster-scoped name of the GatewayClass resource.
	// It equals the OpenMCF resource metadata.name.
	GatewayClassName string
	ControllerName   string
	Labels           map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetesgatewayclassv1.KubernetesGatewayClassStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	// GatewayClass is cluster-scoped, so its Kubernetes resource name is simply
	// the OpenMCF resource name (no namespace qualifier).
	gatewayClassName := metadata.Name

	labels := map[string]string{
		"app.kubernetes.io/name":       "gateway-class",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "openmcf",
		"app.kubernetes.io/component":  "gateway-class",
	}

	return &Locals{
		KubernetesGatewayClass: target,
		GatewayClassName:       gatewayClassName,
		ControllerName:         spec.ControllerName,
		Labels:                 labels,
	}
}

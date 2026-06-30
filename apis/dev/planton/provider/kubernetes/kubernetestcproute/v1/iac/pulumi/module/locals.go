package module

import (
	kubernetestcproutev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetestcproute/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the resolved inputs the module operates on: the full target
// resource plus the scalar identifiers used for the resource name, namespace,
// labels, and stack outputs.
type Locals struct {
	KubernetesTcpRoute *kubernetestcproutev1.KubernetesTcpRoute
	RouteName          string
	Namespace          string
	Labels             map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetestcproutev1.KubernetesTcpRouteStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	routeName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware
	// resolves valueFrom references to literal strings before the IaC module
	// runs, so GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "tcproute",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "tcproute",
	}

	return &Locals{
		KubernetesTcpRoute: target,
		RouteName:          routeName,
		Namespace:          namespace,
		Labels:             labels,
	}
}

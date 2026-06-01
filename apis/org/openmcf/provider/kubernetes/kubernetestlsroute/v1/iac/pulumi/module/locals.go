package module

import (
	kubernetestlsroutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetestlsroute/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the resolved inputs the module operates on: the full target
// resource plus the scalar identifiers used for the resource name, namespace,
// labels, and stack outputs.
type Locals struct {
	KubernetesTlsRoute *kubernetestlsroutev1.KubernetesTlsRoute
	RouteName          string
	Namespace          string
	Labels             map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetestlsroutev1.KubernetesTlsRouteStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	routeName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware
	// resolves valueFrom references to literal strings before the IaC module
	// runs, so GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "tlsroute",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "openmcf",
		"app.kubernetes.io/component":  "tlsroute",
	}

	return &Locals{
		KubernetesTlsRoute: target,
		RouteName:          routeName,
		Namespace:          namespace,
		Labels:             labels,
	}
}

package module

import (
	kubernetestlsroutev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestlsroute/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the resolved inputs the module operates on: the full target
// resource plus the scalar identifiers used for the resource name, namespace,
// labels, and outputs.
type Locals struct {
	KubernetesTlsRoute *kubernetestlsroutev1alpha1.KubernetesTlsRoute
	RouteName          string
	Namespace          string
	Labels             map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetestlsroutev1alpha1.KubernetesTlsRouteIacInput) *Locals {
	target := iacInput.Target
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
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "tlsroute",
	}

	return &Locals{
		KubernetesTlsRoute: target,
		RouteName:          routeName,
		Namespace:          namespace,
		Labels:             labels,
	}
}

package module

import (
	kuberneteshttproutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kuberneteshttproute/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the resolved inputs the module operates on: the full target
// resource plus the scalar identifiers used for the resource name, namespace,
// labels, and stack outputs.
type Locals struct {
	KubernetesHttpRoute *kuberneteshttproutev1.KubernetesHttpRoute
	RouteName           string
	Namespace           string
	Labels              map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kuberneteshttproutev1.KubernetesHttpRouteStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	routeName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware
	// resolves valueFrom references to literal strings before the IaC module
	// runs, so GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "httproute",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "openmcf",
		"app.kubernetes.io/component":  "httproute",
	}

	return &Locals{
		KubernetesHttpRoute: target,
		RouteName:           routeName,
		Namespace:           namespace,
		Labels:              labels,
	}
}

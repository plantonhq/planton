package module

import (
	kubernetesgrpcroutev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgrpcroute/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the resolved inputs the module operates on: the full target
// resource plus the scalar identifiers used for the resource name, namespace,
// labels, and outputs.
type Locals struct {
	KubernetesGrpcRoute *kubernetesgrpcroutev1alpha1.KubernetesGrpcRoute
	RouteName           string
	Namespace           string
	Labels              map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesgrpcroutev1alpha1.KubernetesGrpcRouteIacInput) *Locals {
	target := iacInput.Target
	metadata := target.Metadata
	spec := target.Spec

	routeName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware
	// resolves valueFrom references to literal strings before the IaC module
	// runs, so GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "grpcroute",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "grpcroute",
	}

	return &Locals{
		KubernetesGrpcRoute: target,
		RouteName:           routeName,
		Namespace:           namespace,
		Labels:              labels,
	}
}

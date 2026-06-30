package module

import (
	kubernetesreferencegrantv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesreferencegrant/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the resolved inputs the module operates on: the full target
// resource plus the scalar identifiers used for the resource name, namespace,
// labels, and stack outputs.
type Locals struct {
	KubernetesReferenceGrant *kubernetesreferencegrantv1.KubernetesReferenceGrant
	ReferenceGrantName       string
	Namespace                string
	Labels                   map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetesreferencegrantv1.KubernetesReferenceGrantStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	referenceGrantName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware
	// resolves valueFrom references to literal strings before the IaC module
	// runs, so GetValue() returns the resolved value. This is the "to" namespace:
	// the grant lives alongside the resources it authorizes inbound references to.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "referencegrant",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "referencegrant",
	}

	return &Locals{
		KubernetesReferenceGrant: target,
		ReferenceGrantName:       referenceGrantName,
		Namespace:                namespace,
		Labels:                   labels,
	}
}

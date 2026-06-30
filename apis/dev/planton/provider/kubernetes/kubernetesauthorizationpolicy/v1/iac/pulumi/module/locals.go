package module

import (
	kubernetesauthorizationpolicyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesauthorizationpolicy/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesAuthorizationPolicy *kubernetesauthorizationpolicyv1.KubernetesAuthorizationPolicy
	// AuthorizationPolicyName is the Kubernetes resource name of the
	// AuthorizationPolicy. It equals the Planton resource metadata.name.
	AuthorizationPolicyName string
	// Namespace is the resolved namespace the AuthorizationPolicy is created in.
	Namespace string
	Labels    map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetesauthorizationpolicyv1.KubernetesAuthorizationPolicyStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	authorizationPolicyName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware
	// resolves valueFrom references to literal strings before the IaC module
	// runs, so GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "authorization-policy",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "authorization-policy",
	}

	return &Locals{
		KubernetesAuthorizationPolicy: target,
		AuthorizationPolicyName:       authorizationPolicyName,
		Namespace:                     namespace,
		Labels:                        labels,
	}
}

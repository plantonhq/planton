package module

import (
	kubernetesrequestauthenticationv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesrequestauthentication/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesRequestAuthentication *kubernetesrequestauthenticationv1.KubernetesRequestAuthentication
	// RequestAuthenticationName is the Kubernetes resource name of the
	// RequestAuthentication. It equals the OpenMCF resource metadata.name.
	RequestAuthenticationName string
	// Namespace is the resolved namespace the RequestAuthentication is created in.
	Namespace string
	Labels    map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetesrequestauthenticationv1.KubernetesRequestAuthenticationStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	requestAuthenticationName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware
	// resolves valueFrom references to literal strings before the IaC module
	// runs, so GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "request-authentication",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "openmcf",
		"app.kubernetes.io/component":  "request-authentication",
	}

	return &Locals{
		KubernetesRequestAuthentication: target,
		RequestAuthenticationName:       requestAuthenticationName,
		Namespace:                       namespace,
		Labels:                          labels,
	}
}

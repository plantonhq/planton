package module

import (
	kubernetespeerauthenticationv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetespeerauthentication/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesPeerAuthentication *kubernetespeerauthenticationv1.KubernetesPeerAuthentication
	// PeerAuthenticationName is the Kubernetes resource name of the
	// PeerAuthentication. It equals the Planton resource metadata.name.
	PeerAuthenticationName string
	// Namespace is the resolved namespace the PeerAuthentication is created in.
	Namespace string
	Labels    map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetespeerauthenticationv1.KubernetesPeerAuthenticationStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	peerAuthenticationName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware
	// resolves valueFrom references to literal strings before the IaC module
	// runs, so GetValue() returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "peer-authentication",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "peer-authentication",
	}

	return &Locals{
		KubernetesPeerAuthentication: target,
		PeerAuthenticationName:       peerAuthenticationName,
		Namespace:                    namespace,
		Labels:                       labels,
	}
}

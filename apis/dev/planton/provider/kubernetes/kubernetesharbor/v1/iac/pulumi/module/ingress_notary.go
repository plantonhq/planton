package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createNotaryIngress(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	// Skip if Notary ingress is not enabled
	if locals.KubernetesHarbor.Spec.Ingress == nil ||
		locals.KubernetesHarbor.Spec.Ingress.Notary == nil ||
		!locals.KubernetesHarbor.Spec.Ingress.Notary.Enabled ||
		locals.NotaryExternalHostname == "" {
		return nil
	}

	// TODO: Implement Notary ingress resources if needed
	// Notary ingress configuration would follow similar pattern to Core ingress
	// but with appropriate service name and port for Notary service

	return nil
}

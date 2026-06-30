package module

import (
	"github.com/pkg/errors"
	kuberneteskeycloakv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteskeycloak/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespace conditionally creates the Kubernetes namespace that will hold every
// resource in the KubernetesKeycloak deployment based on the create_namespace flag.
// We return the created object (or nil if not created) so that downstream helpers
// can set it as Parent and inherit the namespace automatically.
//
// If create_namespace is false, the namespace is assumed to exist and is not created.
// In this case, nil is returned and downstream resources will use the namespace name
// directly from locals.Namespace.
//
// Terraform equivalent: a standalone kubernetes_namespace resource with count.
func namespace(ctx *pulumi.Context, stackInput *kuberneteskeycloakv1.KubernetesKeycloakStackInput,
	locals *Locals, kubernetesProvider pulumi.ProviderResource) (*kubernetescorev1.Namespace, error) {

	// Only create namespace if the flag is set to true
	if !stackInput.Target.Spec.CreateNamespace {
		return nil, nil
	}

	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:   pulumi.String(locals.Namespace),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	return createdNamespace, nil
}

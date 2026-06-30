package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createNamespace creates the Kubernetes namespace resource
func createNamespace(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) (*kubernetescorev1.Namespace, error) {
	namespace, err := kubernetescorev1.NewNamespace(
		ctx,
		locals.NamespaceName,
		&kubernetescorev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:        pulumi.String(locals.NamespaceName),
				Labels:      pulumi.ToStringMap(locals.Labels),
				Annotations: pulumi.ToStringMap(locals.Annotations),
			},
		},
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create namespace %s", locals.NamespaceName)
	}

	return namespace, nil
}

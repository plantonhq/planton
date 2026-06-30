package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespace creates the Kubernetes namespace that will hold all
// Temporal resources.  The returned object is used as Parent for every
// subsequent resource so they inherit the correct namespace.
// If create_namespace is false, returns nil (namespace must already exist).
func namespace(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource) (*kubernetescorev1.Namespace, error) {

	// Only create namespace if the flag is set to true
	if !locals.KubernetesTemporal.Spec.CreateNamespace {
		return nil, nil
	}

	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.Namespace),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	return createdNamespace, nil
}

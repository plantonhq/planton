package module

import (
	"github.com/pkg/errors"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createImagePullSecret creates an image pull Secret in the target namespace,
// using the Docker credentials from locals.ImagePullSecretData.
func createImagePullSecret(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := corev1.NewSecret(ctx,
		locals.ImagePullSecretName,
		&corev1.SecretArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.ImagePullSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Type:       pulumi.String("kubernetes.io/dockerconfigjson"),
			StringData: pulumi.ToStringMap(locals.ImagePullSecretData),
		},
		opts...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create image pull secret")
	}
	return nil
}

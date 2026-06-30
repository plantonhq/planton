package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// imagePullSecret creates a Kubernetes secret for pulling private container images.
func imagePullSecret(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource) (*kubernetescorev1.Secret, error) {

	// If no image pull secret data is configured, return nil
	if locals.ImagePullSecretData == nil {
		return nil, nil
	}

	// Create image pull secret resource with computed name to avoid conflicts
	secretArgs := &kubernetescorev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.ImagePullSecretName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Type:       pulumi.String("kubernetes.io/dockerconfigjson"),
		StringData: pulumi.ToStringMap(locals.ImagePullSecretData),
	}

	createdImagePullSecret, err := kubernetescorev1.NewSecret(ctx,
		locals.ImagePullSecretName,
		secretArgs,
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create image pull secret")
	}

	return createdImagePullSecret, nil
}

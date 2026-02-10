package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createSecret creates the Kubernetes Secret resource
func createSecret(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) (*kubernetescorev1.Secret, error) {
	secret, err := kubernetescorev1.NewSecret(
		ctx,
		locals.SecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:        pulumi.String(locals.SecretName),
				Namespace:   pulumi.String(locals.SecretNamespace),
				Labels:      pulumi.ToStringMap(locals.Labels),
				Annotations: pulumi.ToStringMap(locals.Annotations),
			},
			Type:       pulumi.String(locals.SecretType),
			StringData: pulumi.ToStringMap(locals.SecretData),
			Immutable:  pulumi.Bool(locals.Immutable),
		},
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create secret %s", locals.SecretName)
	}

	return secret, nil
}

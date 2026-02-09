package module

import (
	"encoding/base64"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dbPasswordSecret creates a Kubernetes Secret containing the external DB
// password when, and only when, external_database is provided by the user
// with a string_value password (not a secret_ref).
// When secret_ref is used, no new secret is created - the existing secret is used directly.
func dbPasswordSecret(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	ext := locals.KubernetesTemporal.Spec.Database.ExternalDatabase
	if ext == nil {
		// No external DB: nothing to create.
		return nil
	}

	// Check if password is provided and is a string value (not a secret ref)
	if ext.Password == nil {
		// No password provided: nothing to create.
		return nil
	}

	// If using a secret reference, we don't need to create a new secret
	if ext.Password.GetSecretRef() != nil {
		return nil
	}

	// Only create a secret when using value
	stringValue := ext.Password.GetValue()
	if stringValue == "" {
		return nil
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(stringValue))

	secretOpts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	secretOpts = append(secretOpts, namespaceDeps...)
	_, err := kubernetescorev1.NewSecret(ctx,
		locals.DatabasePasswordSecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.DatabasePasswordSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Data: pulumi.StringMap{
				vars.DatabasePasswordSecretKey: pulumi.String(encoded),
			},
			Type: pulumi.String("Opaque"),
		}, secretOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create database password secret")
	}

	return nil
}

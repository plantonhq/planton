package module

import (
	"encoding/base64"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func adminCredentials(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	namespaceDeps []pulumi.ResourceOption) (*kubernetescorev1.Secret, error) {

	createdRandomPassword, err := random.NewRandomPassword(ctx,
		"admin-password",
		&random.RandomPasswordArgs{
			Length:     pulumi.Int(12),
			Special:    pulumi.Bool(true),
			Numeric:    pulumi.Bool(true),
			Upper:      pulumi.Bool(true),
			Lower:      pulumi.Bool(true),
			MinSpecial: pulumi.Int(3),
			MinNumeric: pulumi.Int(2),
			MinUpper:   pulumi.Int(2),
			MinLower:   pulumi.Int(2),
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get random password value")
	}

	// Encode the password in Base64
	base64Password := createdRandomPassword.Result.ApplyT(func(p string) (string, error) {
		return base64.StdEncoding.EncodeToString([]byte(p)), nil
	}).(pulumi.StringOutput)

	// Create or update the secret
	secretOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
	}, namespaceDeps...)

	createdAdminPasswordSecret, err := kubernetescorev1.NewSecret(ctx,
		locals.AdminCredentialsSecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.AdminCredentialsSecretName),
				Namespace: pulumi.String(locals.Namespace),
			},
			Data: pulumi.StringMap{
				vars.JenkinsAdminPasswordSecretKey: base64Password,
			},
		}, secretOpts...)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create jenkins admin-password secret")
	}

	//export admin credentials to outputs
	ctx.Export(OpUsername, pulumi.String(vars.JenkinsAdminUsername))
	ctx.Export(OpPasswordSecretName, createdAdminPasswordSecret.Metadata.Name())
	ctx.Export(OpPasswordSecretKey, pulumi.String(vars.JenkinsAdminPasswordSecretKey))

	return createdAdminPasswordSecret, nil
}

package module

import (
	"encoding/base64"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// adminPassword creates a random password and stores it in a Kubernetes Secret.
func adminPassword(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	createdRandomPassword, err := random.NewRandomPassword(ctx,
		locals.PasswordSecretName,
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
		return errors.Wrap(err, "failed to create random password")
	}

	// encode the password in base64
	base64Password := createdRandomPassword.Result.ApplyT(func(p string) (string, error) {
		return base64.StdEncoding.EncodeToString([]byte(p)), nil
	}).(pulumi.StringOutput)

	// create or update the secret
	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdSecret, err := kubernetescorev1.NewSecret(ctx,
		locals.PasswordSecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.PasswordSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Data: pulumi.StringMap{
				vars.RedisPasswordSecretKey: base64Password,
			},
		}, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create admin secret")
	}

	ctx.Export(OpUsername, pulumi.String("default"))
	ctx.Export(OpPasswordSecretName, createdSecret.Metadata.Name())
	ctx.Export(OpPasswordSecretKey, pulumi.String(vars.RedisPasswordSecretKey))

	return nil
}

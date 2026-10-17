package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// generatedSecretLength is the length of the generated salt, NextAuth
// secret and Slack state secret: 48 letters and digits (~285 bits, past the
// 256 bits Langfuse asks for). Generated passwords use 32. Letters and
// digits only: symbols are the one class that can need quoting in a URL or
// a shell. Twin: the Terraform module's random_password resources.
const (
	generatedSecretLength   = 48
	generatedPasswordLength = 32
)

// instanceSecretResult carries what the rest of the module reads from the
// instance Secret's generation.
type instanceSecretResult struct {
	Secret *kubernetescorev1.Secret
	// ProjectPublicKey is the seeded public key: declared, or generated.
	ProjectPublicKey pulumi.StringOutput
}

// instanceSecret creates the module-owned `<metadata.name>-secrets` Secret:
// every declared secret value plus the credentials the module generates
// for the ones the spec leaves empty. The chart's values and the
// additionalEnv channels reference its keys by secretKeyRef, so no secret
// ever lands in the Helm release.
//
// Generation-shape arguments are ignored after creation so an IMPORTED
// credential never silently regenerates (twin: lifecycle.ignore_changes in
// the Terraform module).
func instanceSecret(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.ResourceOption,
) (*instanceSecretResult, error) {
	data := pulumi.StringMap{}
	for key, value := range locals.DeclaredSecrets {
		data[key] = pulumi.String(value)
	}

	passwordIgnores := pulumi.IgnoreChanges([]string{
		"length", "special", "upper", "lower", "numeric",
		"minLower", "minNumeric", "minSpecial", "minUpper", "overrideSpecial",
	})
	generatePassword := func(name string, length int) (pulumi.StringOutput, error) {
		generated, err := random.NewRandomPassword(ctx, name, &random.RandomPasswordArgs{
			Length:  pulumi.Int(length),
			Special: pulumi.Bool(false),
		}, passwordIgnores)
		if err != nil {
			return pulumi.StringOutput{}, errors.Wrapf(err, "failed to generate %s", name)
		}
		return generated.Result, nil
	}

	if locals.Generate.Salt {
		salt, err := generatePassword("salt", generatedSecretLength)
		if err != nil {
			return nil, err
		}
		data[keySalt] = salt
	}
	if locals.Generate.EncryptionKey {
		// Langfuse requires exactly 64 hex characters: 32 random bytes,
		// hex-encoded (`openssl rand -hex 32`). The same bytes in standard
		// base64 are the form the random_bytes importer takes.
		key, err := random.NewRandomBytes(ctx, "encryption-key", &random.RandomBytesArgs{
			Length: pulumi.Int(32),
		}, pulumi.IgnoreChanges([]string{"length"}))
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate the encryption key")
		}
		data[keyEncryptionKey] = key.Hex
		data[keyEncryptionKeyStdB64] = key.Base64
	}
	if locals.Generate.NextauthSecret {
		secret, err := generatePassword("nextauth-secret", generatedSecretLength)
		if err != nil {
			return nil, err
		}
		data[keyNextauthSecret] = secret
	}
	if locals.Generate.SlackStateSecret {
		secret, err := generatePassword("slack-state-secret", generatedSecretLength)
		if err != nil {
			return nil, err
		}
		data[keySlackStateSecret] = secret
	}
	if locals.Generate.AdminPassword {
		password, err := generatePassword("init-user-password", generatedPasswordLength)
		if err != nil {
			return nil, err
		}
		data[keyInitUserPassword] = password
	}

	projectPublicKey := pulumi.String(locals.DeclaredSecrets[keyInitProjectPublicKey]).ToStringOutput()
	if locals.Generate.ProjectKeys {
		// Langfuse's key format: the fixed pk-lf-/sk-lf- prefix and 32
		// random letters and digits. Each generated part is also stored
		// alone, the value the random_password importer takes.
		publicPart, err := generatePassword("init-project-public-key", generatedPasswordLength)
		if err != nil {
			return nil, err
		}
		secretPart, err := generatePassword("init-project-secret-key", generatedPasswordLength)
		if err != nil {
			return nil, err
		}
		// The public key identifies the project; only the secret key is a
		// credential, so the output is not marked secret.
		projectPublicKey = pulumi.Unsecret(pulumi.Sprintf("pk-lf-%s", publicPart)).(pulumi.StringOutput)
		data[keyInitProjectPublicKey] = projectPublicKey
		data[keyInitProjectPublicKeySuffix] = publicPart
		data[keyInitProjectSecretKey] = pulumi.Sprintf("sk-lf-%s", secretPart)
		data[keyInitProjectSecretKeySuffix] = secretPart
	}

	created, err := kubernetescorev1.NewSecret(ctx, locals.InstanceSecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.InstanceSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			}),
			Type: pulumi.String("Opaque"),
			// ToSecret keeps every value out of previews and state.
			StringData: pulumi.ToSecret(data).(pulumi.StringMapOutput),
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependsOn...)...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the instance secret")
	}

	return &instanceSecretResult{Secret: created, ProjectPublicKey: projectPublicKey}, nil
}

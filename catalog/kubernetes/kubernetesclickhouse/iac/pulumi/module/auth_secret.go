package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// generatedPasswordLength is the module-minted password's length: letters
// and digits, no symbols — clients embed the value in connection URLs
// (clickhouse://user:password@host), where a symbol is the one class that
// can need escaping; 32 alphanumerics (~190 bits) is far past any practical
// bar. Twin: random_password.user in the Terraform module.
const generatedPasswordLength = 32

// authSecret creates the `<metadata.name>-clickhouse-auth` Secret with one
// key per provisioned user holding that user's password: the declared one
// (resolved literal from the spec's value-or-ref) when the spec carries it,
// a module-generated one when it does not. The CHI users section references
// it via valueFrom.secretKeyRef — the password never appears in the CHI
// itself. Returns nil when spec.users is empty. Terraform twin:
// random_password.user + kubernetes_secret_v1.auth with the same name, keys
// and contents.
func authSecret(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
) (*kubernetescorev1.Secret, error) {
	users := locals.Spec.GetUsers()
	if len(users) == 0 {
		return nil, nil
	}

	// The generation-shape arguments are ignored after creation so an
	// IMPORTED credential never silently regenerates: rotation stays an
	// explicit verb, never plan fallout. Twin: the Terraform module's
	// lifecycle.ignore_changes on the same argument set.
	generationShapeIgnores := pulumi.IgnoreChanges([]string{
		"length", "special", "upper", "lower", "numeric",
		"minLower", "minNumeric", "minSpecial", "minUpper", "overrideSpecial",
	})

	passwords := pulumi.StringMap{}
	for _, user := range users {
		if declared := user.GetPassword().GetValue(); declared != "" {
			passwords[user.GetName()] = pulumi.String(declared)
			continue
		}
		// One generated password per user declared without one, with a
		// logical name keyed by USER NAME — the credential's identity — so
		// the password follows its user through spec reorders (an index
		// keying would silently SWAP passwords when the list order
		// changes) and a renamed user is honestly a NEW credential.
		generated, err := random.NewRandomPassword(ctx,
			fmt.Sprintf("user-password-%s", user.GetName()),
			&random.RandomPasswordArgs{
				Length:  pulumi.Int(generatedPasswordLength),
				Special: pulumi.Bool(false),
			},
			generationShapeIgnores)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate password for ClickHouse user %s", user.GetName())
		}
		passwords[user.GetName()] = generated.Result
	}

	createdSecret, err := kubernetescorev1.NewSecret(ctx,
		locals.AuthSecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(
				&kubernetesmeta.ObjectMetaArgs{
					Name:      pulumi.String(locals.AuthSecretName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				}),
			Type: pulumi.String("Opaque"),
			// StringData lets the apiserver base64-encode; ToSecret keeps
			// the plaintext passwords out of rendered previews and state.
			StringData: pulumi.ToSecret(passwords).(pulumi.StringMapOutput),
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependencies...)...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s secret", locals.AuthSecretName)
	}

	return createdSecret, nil
}

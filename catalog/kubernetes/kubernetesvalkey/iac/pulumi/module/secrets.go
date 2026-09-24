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
// and digits, no symbols — the chart's init script reads the value from the
// mounted Secret and clients embed it in connection URLs, and a symbol is
// the one class that can need quoting in either place; 32 alphanumerics
// (~190 bits) is far past any practical bar. Twin: random_password.user in
// the Terraform module.
const generatedPasswordLength = 32

// authSecret materializes the ACL passwords as the "<metadata.name>-auth"
// Opaque Secret — ONE KEY PER USERNAME, each key's value that user's
// password: the declared one when the spec carries it, a module-generated
// one when it does not. That layout is the chart's contract for
// auth.usersExistingSecret: its init script reads
// /valkey-users-secret/<passwordKey> where passwordKey defaults to the
// username (the module leaves passwordKey unset), and its metrics exporter
// reads the "default" key the same way. Because the rendered aclUsers carry
// no inline passwords, the chart renders no auth Secret of its own — this
// Secret is the only place the credentials land, and it never transits
// chart values. Returns nil when auth is not declared. Terraform twin:
// random_password.user + kubernetes_secret_v1.auth with the same name,
// keys, and contents.
func authSecret(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.ResourceOption,
) (pulumi.Resource, error) {
	if !locals.AuthEnabled {
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

	data := pulumi.StringMap{}
	for _, user := range locals.Spec.GetAuth().GetUsers() {
		if declared := user.GetPassword(); declared != "" {
			data[user.GetName()] = pulumi.String(declared)
			continue
		}
		// One generated password per user declared without one, with a
		// logical name keyed by USERNAME — the credential's identity — so
		// the password follows its user through spec reorders (an index
		// keying would silently SWAP passwords when the list order
		// changes) and a renamed user is honestly a NEW credential.
		generated, err := random.NewRandomPassword(ctx,
			fmt.Sprintf("auth-password-%s", user.GetName()),
			&random.RandomPasswordArgs{
				Length:  pulumi.Int(generatedPasswordLength),
				Special: pulumi.Bool(false),
			},
			generationShapeIgnores)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate password for ACL user %s", user.GetName())
		}
		data[user.GetName()] = generated.Result
	}

	createdSecret, err := kubernetescorev1.NewSecret(ctx, locals.AuthSecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.AuthSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			}),
			Type:       pulumi.String("Opaque"),
			StringData: data,
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependsOn...)...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth secret")
	}

	return createdSecret, nil
}

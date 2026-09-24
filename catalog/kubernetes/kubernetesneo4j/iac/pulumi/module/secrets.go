package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// authSecretChartKey is the chart's contract: neo4j.passwordFromSecret
	// names a Secret carrying this key with value "neo4j/<password>".
	authSecretChartKey = "NEO4J_AUTH"

	// authSecretPasswordKey holds the BARE password beside the chart's
	// pair, for workloads that take a password rather than "neo4j/<pw>";
	// the password_secret output names it. Twin: the Terraform module's
	// `password` data key.
	authSecretPasswordKey = "password"

	// generatedPasswordLength: letters and digits, no symbols — the chart
	// splits "neo4j/<password>" on the slash and clients embed the value in
	// bolt URLs, and a symbol is the one class that can need quoting in
	// either place; 24 alphanumerics with two of each class is the
	// catalog's shape for a credential a person may also type. Twin:
	// random_password.admin in the Terraform module.
	generatedPasswordLength = 24
)

// authSecret materializes the admin password — declared, or generated here
// when auth is left empty — as the "<metadata.name>-auth" Opaque Secret
// carrying the chart's contract (NEO4J_AUTH = "neo4j/<password>") and a
// bare `password` key. The chart consumes it via neo4j.passwordFromSecret
// and LOOKS THE SECRET UP AT TEMPLATE TIME (its neo4j.secretName helper
// fails the install when the Secret is missing or lacks NEO4J_AUTH), so
// this Secret must exist BEFORE the Helm release — main.go wires the
// explicit dependency. Because passwordFromSecret is set, the chart renders
// no auth Secret of its own — this Secret is the only place the credential
// lands, and it never transits chart values. Returns nil only for the
// existing_secret arm, which references a Secret the user owns. Terraform
// twin: random_password.admin + kubernetes_secret_v1.auth with the same
// name, keys, and contents.
func authSecret(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.ResourceOption,
) (pulumi.Resource, error) {
	if !locals.CreateAuthSecret {
		return nil, nil
	}

	// Marked secret so the value is encrypted in the Pulumi state — twin
	// of the Terraform module's sensitive() wrap.
	var password pulumi.StringOutput
	if locals.GenerateAdminPassword {
		// The generation-shape arguments are ignored after creation so an
		// IMPORTED credential never silently regenerates: rotation stays an
		// explicit verb, never plan fallout. Twin: the Terraform module's
		// lifecycle.ignore_changes on the same argument set.
		generationShapeIgnores := pulumi.IgnoreChanges([]string{
			"length", "special", "upper", "lower", "numeric",
			"minLower", "minNumeric", "minSpecial", "minUpper", "overrideSpecial",
		})
		generated, err := random.NewRandomPassword(ctx, "admin-password",
			&random.RandomPasswordArgs{
				Length:     pulumi.Int(generatedPasswordLength),
				Special:    pulumi.Bool(false),
				MinUpper:   pulumi.Int(2),
				MinLower:   pulumi.Int(2),
				MinNumeric: pulumi.Int(2),
			},
			generationShapeIgnores)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate admin password")
		}
		password = generated.Result
	} else {
		password = pulumi.ToSecret(pulumi.String(locals.AdminPassword)).(pulumi.StringOutput)
	}

	// The pair derives from the SAME value as the bare key, so the two can
	// never drift apart.
	neo4jAuth := password.ApplyT(func(p string) string { return "neo4j/" + p }).(pulumi.StringOutput)

	createdSecret, err := kubernetescorev1.NewSecret(ctx, locals.AuthSecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.AuthSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			}),
			Type: pulumi.String("Opaque"),
			StringData: pulumi.StringMap{
				authSecretChartKey:    neo4jAuth,
				authSecretPasswordKey: password,
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependsOn...)...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth secret")
	}

	return createdSecret, nil
}

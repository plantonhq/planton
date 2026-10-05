package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// envSecret materializes every secret value of the spec (API keys, tokens,
// passwords, webhook URLs, credential headers) as the module-owned
// `<metadata.name>-env` Opaque Secret, one data key per environment
// variable name. The chart's env values reference it by secretKeyRef, so the
// relay configuration ConfigMap and the Helm values carry no secret.
//
// Returns nil when the spec holds no secret. Created before the release:
// the Deployment's secretKeyRefs must resolve on first pod start.
// Terraform twin: kubernetes_secret_v1.env with count.
func envSecret(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.ResourceOption,
) (*kubernetescorev1.Secret, error) {
	if locals.EnvSecretName == "" {
		return nil, nil
	}

	createdSecret, err := kubernetescorev1.NewSecret(ctx,
		locals.EnvSecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.EnvSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			}),
			Type:       pulumi.String("Opaque"),
			StringData: pulumi.ToStringMap(locals.Relay.SecretEnv),
		}, append([]pulumi.ResourceOption{
			pulumi.Provider(kubernetesProvider),
			// The values are secrets in Pulumi state and previews too.
			pulumi.AdditionalSecretOutputs([]string{"data", "stringData"}),
		}, dependsOn...)...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s secret", locals.EnvSecretName)
	}

	return createdSecret, nil
}

package module

import (
	"sort"

	"github.com/pkg/errors"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// secret creates a Kubernetes Secret containing secret environment variables
// from KubernetesDaemonSet.Spec.Container.App.Env.Secrets that have direct string values.
// Secrets that reference external Kubernetes Secrets are NOT included here;
// they are handled directly in the DaemonSet env var configuration.
func secret(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	dataMap := make(map[string]string)

	if locals.KubernetesDaemonSet.Spec.Container.App.Env != nil {
		secrets := locals.KubernetesDaemonSet.Spec.Container.App.Env.Secrets
		if secrets != nil && len(secrets) > 0 {
			// Sort keys for deterministic output
			sortedKeys := make([]string, 0, len(secrets))
			for k := range secrets {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)

			for _, secretKey := range sortedKeys {
				secretValue := secrets[secretKey]
				// Only add secrets that are direct string values
				if secretValue.GetValue() != "" {
					dataMap[secretKey] = secretValue.GetValue()
				}
			}
		}
	}

	// Only create the secret if there are direct string values to store
	if len(dataMap) == 0 {
		return nil
	}

	secretOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := corev1.NewSecret(ctx,
		locals.EnvSecretName,
		&corev1.SecretArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.EnvSecretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Type:       pulumi.String("Opaque"),
			StringData: pulumi.ToStringMap(dataMap),
		},
		secretOpts...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create secret resource")
	}

	return nil
}

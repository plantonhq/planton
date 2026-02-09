package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"sort"
)

// secret creates a Kubernetes Secret for environment secrets that are provided as direct string values.
// Secrets that reference external Kubernetes Secrets (via secretRef) are not included here;
// they are handled directly in the deployment as environment variable references.
func secret(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	dataMap := make(map[string]string)

	// Add all secrets that have direct string values to the data map
	if locals.KubernetesDeployment.Spec.Container.App.Env != nil {
		secrets := locals.KubernetesDeployment.Spec.Container.App.Env.Secrets
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
				// Secrets with secretRef are handled directly in the deployment
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

	// Create a standard kubernetes secret with computed name to avoid conflicts
	secretArgs := &kubernetescorev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.EnvSecretName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Type:       pulumi.String("Opaque"),
		StringData: pulumi.ToStringMap(dataMap),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := kubernetescorev1.NewSecret(ctx,
		locals.EnvSecretName,
		secretArgs,
		opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create secret resource")
	}

	return nil
}

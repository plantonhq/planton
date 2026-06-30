package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerenv"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// secret creates a Kubernetes Secret for environment secrets that are provided as direct string values.
// Secrets that reference external Kubernetes Secrets (via secretRef) are not included here;
// they are handled directly in the deployment as environment variable references.
func secret(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	dataMap := containerenv.BuildSecretData(locals.KubernetesDeployment.Spec.Container.App.Env)

	if dataMap == nil {
		return nil
	}

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

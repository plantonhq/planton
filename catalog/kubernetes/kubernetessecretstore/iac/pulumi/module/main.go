package module

import (
	"github.com/pkg/errors"
	kubernetessecretstorev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessecretstore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/externalsecretsstore"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one External Secrets Operator SecretStore (namespaced)
// plus the credential Secrets its backend needs.
//
// The CR spec is rendered by the shared externalsecretsstore builder — the
// SAME builder the KubernetesClusterSecretStore module uses, because
// upstream SecretStore and ClusterSecretStore share an identical spec and
// the two Planton kinds share the ExternalSecretsStoreConfig proto. One
// builder means the two kinds can never drift.
//
// Credential Secrets land in the store's own namespace — a namespaced
// store's secret references resolve there by default, which is exactly the
// blast-radius boundary the namespaced grain exists for.
//
// Neither engine waits for the store to reach Ready: readiness depends on
// external reachability (the cloud secrets API, Vault) that is not part of
// applying the resource. Terraform equivalent: kubectl_manifest without a
// wait_for block.
func Resources(ctx *pulumi.Context, iacInput *kubernetessecretstorev1alpha1.KubernetesSecretStoreIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	result, err := externalsecretsstore.BuildSpec(
		locals.StoreName, locals.Namespace, false, iacInput.Target.Spec.Config)
	if err != nil {
		return errors.Wrap(err, "failed to build secret store spec")
	}

	// Credential Secrets first; the CR depends on them so ESO never
	// observes a store whose secretRefs dangle.
	var secretResources []pulumi.Resource
	for _, credential := range result.Secrets {
		createdSecret, err := corev1.NewSecret(ctx, credential.Name,
			&corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(credential.Name),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				StringData: pulumi.ToSecret(pulumi.ToStringMap(credential.Data)).(pulumi.StringMapOutput),
			},
			pulumi.Provider(kubernetesProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create credential secret %s", credential.Name)
		}
		secretResources = append(secretResources, createdSecret)
	}

	opts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	if len(secretResources) > 0 {
		opts = append(opts, pulumi.DependsOn(secretResources))
	}

	_, err = apiextensions.NewCustomResource(ctx, locals.StoreName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("external-secrets.io/v1"),
			Kind:       pulumi.String("SecretStore"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.StoreName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": result.Spec,
			},
		},
		opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create secret store")
	}

	ctx.Export(OpStoreName, pulumi.String(locals.StoreName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

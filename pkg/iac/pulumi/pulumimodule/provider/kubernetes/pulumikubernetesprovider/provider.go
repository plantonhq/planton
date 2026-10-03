// Pulumi kubernetes provider construction from a Planton Kubernetes connection.
// Runs INSIDE the pulumi module process; the kubeconfig itself comes from the shared
// builder (pkg/kubernetes/kubeconfig) so this path and the tofu env-var path cannot
// drift on how a provider's credentials are wired.
package pulumikubernetesprovider

import (
	"os"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/kubernetes/execcredential"
	"github.com/plantonhq/planton/pkg/kubernetes/kubeconfig"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// GetWithKubernetesProviderConfig returns the kubernetes provider for the connection
// carried in the IaC input. A nil config falls back to the host kubeconfig (with
// optional KUBE_CTX context selection) -- the local-workflow path.
func GetWithKubernetesProviderConfig(ctx *pulumi.Context,
	kubernetesProviderConfig *kubernetesprovider.KubernetesProviderConfig,
	providerName string) (*kubernetes.Provider, error) {

	if kubernetesProviderConfig == nil {
		kubeContext := os.Getenv(kubeconfig.KubeContextEnvVar)

		providerArgs := &kubernetes.ProviderArgs{
			EnableServerSideApply: pulumi.Bool(true),
		}
		if kubeContext != "" {
			providerArgs.Context = pulumi.String(kubeContext)
		}

		provider, err := kubernetes.NewProvider(ctx, providerName, providerArgs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get new provider")
		}
		return provider, nil
	}

	// The module process cannot know the engine-spawning binary's path on its own;
	// the host advertises it through the env contract so the kubeconfig's exec
	// credential points at a binary that actually exists. The builder rejects an
	// empty path for exec-credential providers with an error naming the variable.
	kubeConfigString, err := kubeconfig.Build(kubernetesProviderConfig,
		os.Getenv(execcredential.CommandPathEnvVar))
	if err != nil {
		return nil, errors.Wrap(err, "failed to build kubeconfig from provider config")
	}

	provider, err := kubernetes.NewProvider(ctx,
		providerName,
		&kubernetes.ProviderArgs{
			EnableServerSideApply: pulumi.Bool(true),
			Kubeconfig:            pulumi.String(kubeConfigString),
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get new provider")
	}
	return provider, nil
}

// GetWithKubernetesProviderConfigUpsert is GetWithKubernetesProviderConfig
// with upsertExistingObjects enabled: server-side apply ADOPTS an existing
// object on create instead of failing AlreadyExists. The seam for
// MODULE-OWNED, retained-on-destroy resources (kept CRDs above all): a
// destroy leaves them on the cluster by design, so the next install of the
// same module must re-adopt them — the exact semantic twin of Terraform's
// kubectl_manifest server-side apply. Scope providers built here to ONLY
// those resources; ordinary resources keep the default create-conflict
// semantics through the plain provider.
func GetWithKubernetesProviderConfigUpsert(ctx *pulumi.Context,
	kubernetesProviderConfig *kubernetesprovider.KubernetesProviderConfig,
	providerName string) (*kubernetes.Provider, error) {

	if kubernetesProviderConfig == nil {
		kubeContext := os.Getenv(kubeconfig.KubeContextEnvVar)

		providerArgs := &kubernetes.ProviderArgs{
			EnableServerSideApply: pulumi.Bool(true),
			UpsertExistingObjects: pulumi.Bool(true),
			// Force field-manager conflicts on adoption — the retained
			// object was last applied by a different stack's manager.
			// Twin of Terraform's kubectl force_conflicts.
			EnablePatchForce: pulumi.Bool(true),
		}
		if kubeContext != "" {
			providerArgs.Context = pulumi.String(kubeContext)
		}

		provider, err := kubernetes.NewProvider(ctx, providerName, providerArgs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get new provider")
		}
		return provider, nil
	}

	kubeConfigString, err := kubeconfig.Build(kubernetesProviderConfig,
		os.Getenv(execcredential.CommandPathEnvVar))
	if err != nil {
		return nil, errors.Wrap(err, "failed to build kubeconfig from provider config")
	}

	provider, err := kubernetes.NewProvider(ctx,
		providerName,
		&kubernetes.ProviderArgs{
			EnableServerSideApply: pulumi.Bool(true),
			UpsertExistingObjects: pulumi.Bool(true),
			EnablePatchForce:      pulumi.Bool(true),
			Kubeconfig:            pulumi.String(kubeConfigString),
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get new provider")
	}
	return provider, nil
}

// GetWithKubernetesProviderConfigAndNamespace is
// GetWithKubernetesProviderConfig with the provider's default namespace set:
// namespaced resources that declare no metadata.namespace are applied to
// defaultNamespace, while resources with an explicit namespace — and
// cluster-scoped resources — are untouched (the provider resolves each
// kind's scope before defaulting). The seam for modules that apply
// user-authored manifests anchored to a namespace.
func GetWithKubernetesProviderConfigAndNamespace(ctx *pulumi.Context,
	kubernetesProviderConfig *kubernetesprovider.KubernetesProviderConfig,
	providerName string, defaultNamespace string) (*kubernetes.Provider, error) {

	if kubernetesProviderConfig == nil {
		kubeContext := os.Getenv(kubeconfig.KubeContextEnvVar)

		providerArgs := &kubernetes.ProviderArgs{
			EnableServerSideApply: pulumi.Bool(true),
			Namespace:             pulumi.String(defaultNamespace),
		}
		if kubeContext != "" {
			providerArgs.Context = pulumi.String(kubeContext)
		}

		provider, err := kubernetes.NewProvider(ctx, providerName, providerArgs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get new provider")
		}
		return provider, nil
	}

	kubeConfigString, err := kubeconfig.Build(kubernetesProviderConfig,
		os.Getenv(execcredential.CommandPathEnvVar))
	if err != nil {
		return nil, errors.Wrap(err, "failed to build kubeconfig from provider config")
	}

	provider, err := kubernetes.NewProvider(ctx,
		providerName,
		&kubernetes.ProviderArgs{
			EnableServerSideApply: pulumi.Bool(true),
			Kubeconfig:            pulumi.String(kubeConfigString),
			Namespace:             pulumi.String(defaultNamespace),
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get new provider")
	}
	return provider, nil
}

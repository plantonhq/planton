package module

import (
	"github.com/pkg/errors"
	kubernetesmanifestv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesmanifest/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	yamlv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesmanifestv1.KubernetesManifestStackInput) error {
	locals, err := initializeLocals(ctx, stackInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// Create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Apply the manifest YAML using yamlv2.ConfigGroup
	if err := applyManifest(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to apply manifest")
	}

	return nil
}

// applyManifest applies the raw Kubernetes manifest YAML using yamlv2.ConfigGroup
func applyManifest(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider *pulumikubernetes.Provider,
	namespaceDeps []pulumi.ResourceOption) error {

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	// Use yamlv2.ConfigGroup which handles multi-document YAML and CRD ordering
	_, err := yamlv2.NewConfigGroup(ctx, "manifest", &yamlv2.ConfigGroupArgs{
		Yaml: pulumi.StringPtr(locals.ManifestYAML),
	}, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create config group from manifest YAML")
	}

	return nil
}

package module

import (
	"github.com/pkg/errors"
	kubernetesmanifestv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesmanifest/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	yamlv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It applies the raw manifest through yaml.v2 ConfigGroup, which handles
// multi-document YAML and orders CRDs before the custom resources that use
// them.
//
// NAMESPACE SEMANTICS (parity contract with the Terraform module): the
// provider is constructed with spec.namespace as its default namespace, so
// namespaced documents that declare no metadata.namespace land there, while
// explicit namespaces and cluster-scoped documents pass through untouched —
// the provider resolves each kind's scope before defaulting. The Terraform
// module reaches the same outcome with a per-document override_namespace on
// documents that declare no namespace.
func Resources(ctx *pulumi.Context, iacInput *kubernetesmanifestv1alpha1.KubernetesManifestIacInput) error {
	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// Create kubernetes provider from the credential in the iac-input,
	// anchored to the spec namespace (see namespace semantics above).
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfigAndNamespace(ctx,
		iacInput.ProviderConfig, "kubernetes", locals.Namespace)
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ manifest ------------------------------
	if err := applyManifest(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to apply manifest")
	}

	exportOutputs(ctx, locals)
	return nil
}

// applyManifest applies the raw manifest YAML through yaml.v2 ConfigGroup.
// The manifest content is applied exactly as written — no injected labels,
// no rewritten fields; only the namespace default described on Resources.
func applyManifest(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	namespaceDeps []pulumi.ResourceOption) error {

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	// skip_await parity: SkipAwait here, wait/wait_for_rollout inverted on
	// the kubectl_manifest resources in the Terraform module. Both engines
	// default to awaiting readiness.
	_, err := yamlv2.NewConfigGroup(ctx, "manifest", &yamlv2.ConfigGroupArgs{
		Yaml:      pulumi.StringPtr(locals.Spec.GetManifestYaml()),
		SkipAwait: pulumi.Bool(locals.Spec.GetSkipAwait()),
	}, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create config group from manifest YAML")
	}

	return nil
}

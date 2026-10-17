package module

import (
	"github.com/pkg/errors"
	kuberneteslangfusev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslangfuse/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Langfuse from the official chart as a real Helm
// release, with the objects the chart does not provide:
//  1. the namespace (create_namespace),
//  2. `<name>-secrets`, the instance Secret: every declared secret value
//     and every credential the module generates (secrets.go),
//  3. the Helm release (helm_release.go), rendered from the typed spec in
//     values.go and env.go, its four bundled stores held off.
//
// The exact same resource set renders from the Terraform module - keep them
// in lockstep.
func Resources(ctx *pulumi.Context, iacInput *kuberneteslangfusev1alpha1.KubernetesLangfuseIacInput) error {
	// NAME BUDGET, checked before anything is created (Terraform twin: the
	// lifecycle precondition on helm_release.langfuse).
	if len(iacInput.Target.Metadata.Name) > vars.MaxMetadataNameLength {
		return errors.Errorf(
			"metadata.name %q is %d characters; Langfuse's name budget allows at most %d "+
				"(the chart names its worker PodDisruptionBudget <name>-worker-pdb, and an object name is a 63-character DNS label) - shorten metadata.name",
			iacInput.Target.Metadata.Name, len(iacInput.Target.Metadata.Name), vars.MaxMetadataNameLength)
	}

	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return err
	}

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ instance secret ----------------------
	secret, err := instanceSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return err
	}

	releaseDeps := []pulumi.Resource{secret.Secret}
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, createdNamespace)
	}

	// ------------------------------ helm release -------------------------
	if _, err := helmRelease(ctx, locals, kubernetesProvider,
		[]pulumi.ResourceOption{pulumi.DependsOn(releaseDeps)}); err != nil {
		return err
	}

	exportOutputs(ctx, locals, secret.ProjectPublicKey)
	return nil
}

package module

import (
	"github.com/pkg/errors"
	kubernetesopenfgav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopenfga/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs OpenFGA from the official chart as a real Helm
// release. The typed spec renders into chart values (helm_release.go);
// declared pre-shared API keys materialize into a module-owned Secret
// before the release (authn_secret.go) so key material never rides
// rendered values; the helm_values escape hatch merges last with Helm -f
// semantics — the exact semantic twin of the Terraform module's
// helm_release with values = [typed, helm_values, fullnameOverride
// re-pin].
func Resources(ctx *pulumi.Context, iacInput *kubernetesopenfgav1alpha1.KubernetesOpenFgaIacInput) error {
	// NAME BUDGET, checked before anything is created: the chart
	// truncates its fullname at 63 characters THEN derives
	// `<fullname>-migrate` for the migration Job, whose pod label value
	// also caps at 63 — a name past 55 would truncate silently or push
	// the Job's label over the limit. Fail loudly instead (Terraform
	// twin: the lifecycle precondition on helm_release.openfga).
	if len(iacInput.Target.Metadata.Name) > vars.MaxMetadataNameLength {
		return errors.Errorf(
			"metadata.name %q is %d characters; the OpenFGA chart's name budget allows at most %d "+
				"(the chart truncates its fullname at 63 and appends \"-migrate\" for the migration Job)",
			iacInput.Target.Metadata.Name, len(iacInput.Target.Metadata.Name), vars.MaxMetadataNameLength)
	}

	locals := initializeLocals(ctx, iacInput)

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

	// --------------------------- authn keys secret ------------------------
	// Materialized before the release: the Deployment's secretKeyRef
	// must resolve on first pod start or the rollout wedges.
	createdAuthnSecret, err := authnSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create authn keys secret")
	}

	var releaseDeps []pulumi.ResourceOption
	releaseDeps = append(releaseDeps, namespaceDeps...)
	if createdAuthnSecret != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdAuthnSecret}))
	}

	// ------------------------------ helm release --------------------------
	if err := helmRelease(ctx, locals, kubernetesProvider, releaseDeps); err != nil {
		return err
	}

	exportOutputs(ctx, locals)
	return nil
}

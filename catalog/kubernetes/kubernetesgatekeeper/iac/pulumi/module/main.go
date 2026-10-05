package module

import (
	"github.com/pkg/errors"
	kubernetesgatekeeperv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgatekeeper/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs OPA Gatekeeper from the official chart as a real
// Helm release. The typed spec renders into chart values (values.go); the
// helm_values escape hatch merges last with Helm -f semantics — the exact
// semantic twin of the Terraform module's helm_release with
// values = [typed, helm_values].
//
// WEBHOOK LIFECYCLE: the chart OWNS the webhook configurations as release
// objects (unlike engines that register webhooks at runtime) — uninstall
// removes them with everything else. The policy webhook is fail-open by
// default; the namespace-label check webhook is fail-closed (both typed).
//
// CRD LIFECYCLE: the engine CRDs ship in the chart's crds/ directory —
// Helm installs them once and NEVER upgrades or deletes them; the chart's
// own pre-install/pre-upgrade Job (upgradeCRDs) keeps them current on
// upgrades, and uninstall KEEPS them by design. Constraint-template CRDs
// Gatekeeper creates at runtime also survive uninstall until their
// templates are deleted.
func Resources(ctx *pulumi.Context, iacInput *kubernetesgatekeeperv1alpha1.KubernetesGatekeeperIacInput) error {
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

	var releaseDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ helm release --------------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the engine to become Ready — the chart's own
		// post-install probe job curls the webhook endpoint, so a green
		// install means the webhook actually serves. SkipAwait false is
		// Helm --wait, stated explicitly to mirror the Terraform twin's
		// `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install gatekeeper helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

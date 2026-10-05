package module

import (
	"github.com/pkg/errors"
	kubernetesgharunnerscalesetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgharunnerscaleset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys a GitHub Actions runner scale set from the official
// OCI chart as a real Helm release: the AutoscalingRunnerSet the
// controller (a KubernetesGhaRunnerScaleSetController install, the
// registry prerequisite) reconciles into a listener and ephemeral runner
// pods.
//
// SECRET DISCIPLINE: on the declared PAT / GitHub App arms the module
// materializes the credential as the `<name>-github-auth` Secret BEFORE
// the release and passes only its NAME into chart values (the chart's
// pre-defined-secret form) — credential material never rides rendered
// values. The existing-Secret arm references the user's own Secret.
//
// OCI WIRING: Pulumi's helm.v3.Release resolves OCI registries through
// the CHART REFERENCE — the joined "oci://.../<chart>" string with NO
// RepositoryOpts; the Terraform provider takes repository + bare chart
// name instead. Same chart bytes, different wiring.
//
// NO HELM WAIT: the chart's workload is the AutoscalingRunnerSet custom
// resource — the controller creates the listener AFTER the release
// returns, and the listener needs valid GitHub credentials to come up.
// Helm --wait would pass trivially (a CR is always "ready") while
// --atomic would still roll back on nothing real; the E2E verifier owns
// the listener-registered proof instead (the CR-kind precedent).
func Resources(ctx *pulumi.Context, iacInput *kubernetesgharunnerscalesetv1alpha1.KubernetesGhaRunnerScaleSetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY past the chart's own scale-set-name budget: the chart
	// template fails installs at >45 characters (a GitHub registration
	// limit) — catch it at plan/preview instead of mid-apply. The spec
	// CEL caps the explicit field; this guard covers the metadata.name
	// fallback. Twin: the Terraform module's lifecycle precondition.
	if len(locals.RunnerScaleSetName) > vars.ScaleSetNameBudget {
		return errors.Errorf(
			"runner scale set name %q is %d characters — GitHub caps registrations at %d; set spec.runner_scale_set_name (or a shorter metadata.name)",
			locals.RunnerScaleSetName, len(locals.RunnerScaleSetName), vars.ScaleSetNameBudget)
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

	var secretDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		secretDeps = append(secretDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------ credential secret ---------------------------
	createdSecret, err := githubAuthSecret(ctx, locals, kubernetesProvider, secretDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create github auth secret")
	}

	var releaseDeps []pulumi.ResourceOption
	switch {
	case createdSecret != nil:
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdSecret}))
	case createdNamespace != nil:
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
		// OCI chart reference — joined string, no RepositoryOpts (see the
		// module comment).
		Chart:   pulumi.String(vars.HelmOciRepo + "/" + vars.HelmChartName),
		Version: pulumi.String(locals.ChartVersion),
		Values:  pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// No Helm wait — see the module comment (the workload is a CR the
		// controller reconciles after the release returns).
		SkipAwait: pulumi.Bool(true),
		Timeout:   pulumi.Int(300),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install gha-runner-scale-set helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

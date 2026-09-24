package module

import (
	"github.com/pkg/errors"
	kubernetesplantonrunnerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonrunner/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys a standing Planton runner from the official
// planton-runner OCI chart as a real Helm release, so the deployed runner
// is byte-identical to a hand-installed one — the chart carries the
// load-bearing enrollment mechanics this module deliberately does NOT
// re-model: replicas pinned to 1 with a Recreate strategy (two live pods
// under one runner name would revoke each other's keys), the ephemeral
// identity volume the runner persists its minted identity into (container
// restarts reuse it; pod recreation re-joins with the token), and the
// health endpoints.
//
// SECRET DISCIPLINE: the module materializes the runner token as the
// `<name>-token` Secret BEFORE the release and passes only its NAME into
// chart values (the chart's existingSecret form) — token material never
// rides rendered values, and the escape hatch cannot move it (the
// enrollment secret wiring is re-pinned after the merge).
//
// OCI WIRING: Pulumi's helm.v3.Release resolves OCI registries through
// the CHART REFERENCE — the joined "oci://.../<chart>" string with NO
// RepositoryOpts; the Terraform provider takes repository + bare chart
// name instead. Same chart bytes, different wiring.
//
// NO HELM WAIT: the runner's readiness contract is its control-plane work
// queue, not pod liveness — a runner whose control plane is momentarily
// unreachable (or whose token is still propagating) must still deploy and
// destroy cleanly. Helm --wait would couple the install to control-plane
// reachability; the E2E verifier owns the install-level proof instead
// (the same posture as the runner appliances on the cloud substrates).
func Resources(ctx *pulumi.Context, stackInput *kubernetesplantonrunnerv1alpha1.KubernetesPlantonRunnerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// FAIL LOUDLY below the enrollment-contract floor: charts before
	// 0.4.0 silently IGNORE the enrollment values — the runner would
	// deploy with no way to join, and nothing downstream would name the
	// cause. Twin: the Terraform module's lifecycle precondition.
	ok, err := chartVersionAtLeast(locals.ChartVersion, vars.MinChartVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to parse chart version %q", locals.ChartVersion)
	}
	if !ok {
		return errors.Errorf(
			"chart version %q predates the runner's token-enrollment contract — use %s or newer (charts below it silently ignore the enrollment values)",
			locals.ChartVersion, vars.MinChartVersion)
	}

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

	var secretDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		secretDeps = append(secretDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ---------------------------- token secret ---------------------------
	createdSecret, err := tokenSecret(ctx, locals, kubernetesProvider, secretDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create token secret")
	}

	// ---------------------------- helm release ---------------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		// OCI chart reference — joined string, no RepositoryOpts (see the
		// module comment).
		Chart:   pulumi.String(locals.ChartRepository + "/" + vars.HelmChartName),
		Version: pulumi.String(locals.ChartVersion),
		Values:  pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// No Helm wait — see the module comment (readiness is the work
		// queue, never pod liveness).
		SkipAwait: pulumi.Bool(true),
		Timeout:   pulumi.Int(300),
	}

	opts := []pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{createdSecret}),
	}

	if _, err := helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...); err != nil {
		return errors.Wrap(err, "failed to install planton-runner helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

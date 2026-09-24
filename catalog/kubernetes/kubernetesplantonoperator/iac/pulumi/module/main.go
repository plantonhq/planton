package module

import (
	"github.com/pkg/errors"
	kubernetesplantonoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonoperator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Planton operator from the official
// planton-operator Helm chart (OCI, ghcr.io/plantonhq/charts) as ONE real
// Helm release, byte-identical to a hand-installed one. The chart owns its
// two definitions (PlantonPlatform, PlantonIdentityProvider) as release
// resources: they upgrade with the chart and, behind crds.keep, survive an
// uninstall. The module maps the spec's crds dials onto the chart's
// crds.enabled / crds.keep values (values.go) and applies nothing else —
// no second copy of the schema exists anywhere. The optional namespace is
// the only object the module creates itself.
//
// The release name is FIXED: the operator enforces one installation per
// cluster itself at startup (a label-matched Deployment scan), so a second
// release could only crash-loop — the fixed name makes the collision
// impossible to express instead of merely refused. Installing more
// PLATFORMS needs no second operator: one operator watches all namespaces.
//
// The typed spec renders into chart values (values.go); the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release with
// values = [typed, helm_values].
func Resources(ctx *pulumi.Context, stackInput *kubernetesplantonoperatorv1alpha1.KubernetesPlantonOperatorStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Refuse charts that do not own their definitions: below the floor the
	// crds dials would be silently dropped, which a module must never do.
	// Twin: the Terraform module's lifecycle precondition.
	ok, err := chartVersionAtLeast(locals.ChartVersion, vars.MinChartVersion)
	if err != nil {
		return errors.Wrapf(err, "failed to parse chart version %q", locals.ChartVersion)
	}
	if !ok {
		return errors.Errorf(
			"observed: spec.chart_version is %q\n"+
				"meaning: planton-operator charts older than %s install their definitions once from Helm's crds/ directory and have no crds.enabled / crds.keep values, so the crds dials of this resource would have no effect\n"+
				"next step: set spec.chart_version to %s or newer (or leave it unset for the catalog default)",
			locals.ChartVersion, vars.MinChartVersion, vars.MinChartVersion)
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

	var releaseDeps []pulumi.Resource
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, createdNamespace)
	}

	// ------------------------------ operator release ----------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	_, err = helmv3.NewRelease(ctx, vars.ReleaseName, &helmv3.ReleaseArgs{
		Name:      pulumi.String(vars.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		// OCI chart reference — joined string, no RepositoryOpts (Pulumi's
		// helm.v3.Release does not resolve oci:// through RepositoryOpts
		// the way the Terraform provider does).
		Chart:   pulumi.String(locals.ChartRepository + "/" + vars.HelmChartName),
		Version: pulumi.String(locals.ChartVersion),
		Values:  pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag). The
		// definitions are release resources the chart renders behind
		// values.crds — never skipped here: SkipCrds governs only Helm's
		// install-once crds/ directory, which this chart does not use.
		CreateNamespace: pulumi.Bool(false),
		// Wait for the operator to become Available — a manager that never
		// becomes ready (the one-per-cluster startup guard refusing beside
		// a sibling operator is THE classic case) should fail THIS deploy
		// with a readiness timeout, not surface later as PlantonPlatform
		// resources that mysteriously never reconcile.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}, append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider)},
		dependsOn(releaseDeps)...)...)
	if err != nil {
		return errors.Wrap(err, "failed to install planton-operator helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(vars.ReleaseName))

	return nil
}

// dependsOn wraps a possibly-empty dependency list into resource options
// (an empty DependsOn is a valid no-op).
func dependsOn(deps []pulumi.Resource) []pulumi.ResourceOption {
	if len(deps) == 0 {
		return nil
	}
	return []pulumi.ResourceOption{pulumi.DependsOn(deps)}
}

package module

import (
	"github.com/pkg/errors"
	kubernetesgharunnerscalesetcontrollerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgharunnerscalesetcontroller/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the GitHub Actions Runner Scale Set controller from
// the official OCI chart as a real Helm release. The typed spec renders
// into chart values (values.go); the helm_values escape hatch merges last
// with Helm -f semantics — the exact semantic twin of the Terraform
// module's helm_release with values = [typed, helm_values].
//
// OCI WIRING: Pulumi's helm.v3.Release resolves OCI registries through
// the CHART REFERENCE — the joined
// "oci://ghcr.io/actions/actions-runner-controller-charts/<chart>" string
// with NO RepositoryOpts. The Terraform provider instead takes
// repository = the registry path plus the bare chart name and joins them
// internally. Same chart bytes, different wiring — keep both sides of the
// split in lockstep.
//
// CRD LIFECYCLE: the chart installs the actions.github.com CRDs with the
// release and they are REMOVED with it — destroying the controller
// cascade-deletes every runner scale set on the cluster (the spec's CRD
// note carries the warning).
func Resources(ctx *pulumi.Context, iacInput *kubernetesgharunnerscalesetcontrollerv1alpha1.KubernetesGhaRunnerScaleSetControllerIacInput) error {
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
		// OCI chart reference — joined string, no RepositoryOpts (see the
		// module comment).
		Chart:   pulumi.String(vars.HelmOciRepo + "/" + vars.HelmChartName),
		Version: pulumi.String(locals.ChartVersion),
		Values:  pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the controller to become Ready — a manager that never
		// starts should fail THIS deploy, not the first runner scale set
		// install. SkipAwait false is Helm --wait, stated explicitly to
		// mirror the Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(300),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install gha-runner-scale-set-controller helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

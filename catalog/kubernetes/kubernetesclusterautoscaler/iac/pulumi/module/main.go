package module

import (
	"github.com/pkg/errors"
	kubernetesclusterautoscalerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesclusterautoscaler/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Kubernetes Cluster Autoscaler from the official
// Helm chart as a real Helm release. The typed spec renders into chart
// values (values.go); the helm_values escape hatch merges last with Helm -f
// semantics — the exact semantic twin of the Terraform module's
// helm_release with values = [typed, helm_values].
//
// The release name is FIXED ("cluster-autoscaler"): the autoscaler
// leader-elects and owns the cluster-wide scaling decision — a second
// installation would fight the first over every scale-up, so one
// installation per cluster is the operating model.
func Resources(ctx *pulumi.Context, iacInput *kubernetesclusterautoscalerv1alpha1.KubernetesClusterAutoscalerIacInput) error {
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
		Name:      pulumi.String(vars.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		// The values map carries the arm's credentials (AWS secret key,
		// Azure client secret, Civo API key) on their way into the chart's
		// own Secret — never log or export it.
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the Deployment to become Available — an autoscaler that
		// never becomes ready (bad cloud credentials crash-looping the
		// pod; a ServiceMonitor rendered without the Prometheus operator
		// CRDs) should fail THIS deploy with a readiness timeout, not
		// surface later as node groups that mysteriously never scale. 600s
		// because the image pull plus leader election on a busy
		// kube-system can exceed the usual 300.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, vars.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install cluster-autoscaler helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(vars.ReleaseName))
	// Derived from the chart's fullname template (see locals.go for the
	// _helpers.tpl derivation): "<release>-<cloudProvider>-<chartName>" —
	// the subject cloud-side keyless bindings are written against.
	ctx.Export(OpServiceAccountName, pulumi.String(locals.ServiceAccountName))

	return nil
}

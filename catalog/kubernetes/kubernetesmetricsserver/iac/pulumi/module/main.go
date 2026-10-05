package module

import (
	"github.com/pkg/errors"
	kubernetesmetricsserverv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesmetricsserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs metrics-server from the official Helm chart as a real
// Helm release. The typed spec renders into chart values (values.go); the
// helm_values escape hatch merges last with Helm -f semantics — the exact
// semantic twin of the Terraform module's helm_release with
// values = [typed, helm_values].
//
// The release name is FIXED ("metrics-server"): the component registers the
// cluster-wide v1beta1.metrics.k8s.io APIService, a singleton — one
// installation per cluster is an upstream constraint.
func Resources(ctx *pulumi.Context, iacInput *kubernetesmetricsserverv1alpha1.KubernetesMetricsServerIacInput) error {
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
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the Deployment to become Available — a metrics-server
		// that never becomes ready (kubelet TLS rejection is THE classic
		// cause: self-signed kubelets without kubelet_insecure_tls) should
		// fail THIS deploy with a readiness timeout, not surface later as
		// HPAs that mysteriously never scale. The readinessProbe (/readyz)
		// only passes once the first kubelet scrape succeeds, so a green
		// deploy means metrics are actually flowing.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(300),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, vars.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install metrics-server helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(vars.ReleaseName))
	// The chart's Service carries the fullname, pinned to the release name.
	ctx.Export(OpServiceName, pulumi.String(vars.ReleaseName))
	ctx.Export(OpApiServiceName, pulumi.String(locals.ApiServiceName))

	return nil
}

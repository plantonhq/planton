package module

import (
	"github.com/pkg/errors"
	kubernetesjupyterhubv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesjupyterhub/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs JupyterHub from the official Zero to JupyterHub Helm
// chart as a real Helm release. The typed spec renders into chart values
// (values.go); credentials travel through module-owned Secrets composed
// BEFORE the release (secrets.go) — the external database password and
// the sign-in secrets never appear in rendered values (which this chart
// embeds READABLE inside its own hub Secret); the helm_values escape
// hatch merges last with Helm -f semantics — the exact semantic twin of
// the Terraform module's helm_release with values = [typed, helm_values,
// re-pins].
func Resources(ctx *pulumi.Context, iacInput *kubernetesjupyterhubv1alpha1.KubernetesJupyterHubIacInput) error {
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

	// ------------------------- module-owned secrets ----------------------
	// The hub mounts these at startup (existing-secret + the sign-in
	// env source), so they must exist before the release.
	createdSecrets, err := jupyterhubSecrets(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create module-owned secrets")
	}

	releaseDependencies := make([]pulumi.Resource, 0, len(createdSecrets)+1)
	if createdNamespace != nil {
		releaseDependencies = append(releaseDependencies, createdNamespace)
	}
	releaseDependencies = append(releaseDependencies, createdSecrets...)

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
		// Wait for the hub, proxy and scheduling machinery to become
		// Ready — an install whose hub cannot reach its database or
		// whose image-pull hook cannot finish should fail THIS deploy,
		// not the first user's login. With the pre-puller hook on (the
		// chart default) the notebook-image pull to every node runs
		// inside this budget. SkipAwait false is Helm --wait, stated
		// explicitly to mirror the Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}

	opts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	if len(releaseDependencies) > 0 {
		opts = append(opts, pulumi.DependsOn(releaseDependencies))
	}

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install jupyterhub helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The service names are
// CHART-FIXED (bare names at fullnameOverride "") — the per-namespace
// singleton contract; credential handles are Secret NAMES — values stay
// in-cluster.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	// The shared-password handle is honest: on OAuth/native arms no
	// module-owned sign-in credential exists, so the handle exports
	// EMPTY rather than a name that points at nothing (Terraform twin
	// exports the same empties).
	sharedSecretName, sharedSecretKey := "", ""
	if locals.AuthMethod == "shared_password" {
		sharedSecretName, sharedSecretKey = locals.SharedPasswordSecretName, locals.SharedPasswordSecretKey
	}
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpProxyPublicService, pulumi.String(vars.ProxyPublicServiceName))
	ctx.Export(OpEndpoint, pulumi.String(locals.ProxyPublicEndpoint))
	ctx.Export(OpHubService, pulumi.String(vars.HubServiceName))
	ctx.Export(OpSharedPasswordSecretName, pulumi.String(sharedSecretName))
	ctx.Export(OpSharedPasswordSecretKey, pulumi.String(sharedSecretKey))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

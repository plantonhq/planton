package module

import (
	"github.com/pkg/errors"
	kubernetessupersetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessuperset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Apache Superset from the official ASF Helm chart as
// a real Helm release. The typed spec renders into chart values
// (values.go); the chart's runtime-credential contract — one environment
// Secret every component consumes — is module-composed BEFORE the release
// (secrets.go), with referenced credentials arriving as extraEnvRaw
// secretKeyRef entries; nothing credential-bearing appears in rendered
// values; the helm_values escape hatch merges last with Helm -f semantics
// — the exact semantic twin of the Terraform module's helm_release with
// values = [typed, helm_values, re-pins].
func Resources(ctx *pulumi.Context, iacInput *kubernetessupersetv1alpha1.KubernetesSupersetIacInput) error {
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
	// Every component envFroms the environment Secret at startup, so it
	// must exist before the release.
	createdSecrets, err := supersetSecrets(ctx, locals, kubernetesProvider, namespaceDeps)
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
		Version:   pulumi.String(vars.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the web/worker rollouts and the post-install init
		// Job (schema migration + admin bootstrap run inside this
		// budget against the composed database) — an install whose
		// migration cannot reach its database should fail THIS deploy,
		// not the first login. SkipAwait false is Helm --wait, stated
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
		return errors.Wrap(err, "failed to install superset helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles — Service names and
// Secret NAMES; values stay in-cluster.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	// The secret-key handle exports the module-owned Secret on the
	// generated arm and the referenced Secret on bring-your-own — both
	// are real, existing Secrets an operator can rotate against.
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpService, pulumi.String(locals.Service))
	ctx.Export(OpEndpoint, pulumi.String(locals.Endpoint))
	ctx.Export(OpAdminUsername, pulumi.String(locals.AdminUsername))
	ctx.Export(OpAdminPasswordSecretName, pulumi.String(locals.AdminPasswordSecret))
	ctx.Export(OpAdminPasswordSecretKey, pulumi.String(locals.AdminPasswordKey))
	ctx.Export(OpEnvSecretName, pulumi.String(locals.EnvSecretName))
	ctx.Export(OpSecretKeySecretName, pulumi.String(locals.SecretKeySecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

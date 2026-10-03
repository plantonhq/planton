package module

import (
	"github.com/pkg/errors"
	kubernetestrinov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestrino/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Trino from the official trinodb Helm chart as a real
// Helm release. The typed spec renders into chart values (values.go);
// credentials travel through module-owned Secrets composed BEFORE the
// release (secrets.go) and reach the ConfigMap-rendered properties as
// `${ENV:VAR}` references (Trino's own secrets substitution) — nothing
// credential-bearing appears in rendered values; the helm_values escape
// hatch merges last with Helm -f semantics — the exact semantic twin of
// the Terraform module's helm_release with values = [typed, helm_values,
// re-pins].
func Resources(ctx *pulumi.Context, iacInput *kubernetestrinov1alpha1.KubernetesTrinoIacInput) error {
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
	// The pods consume them at startup (the password-file mount and the
	// env-sourced shared secret), so they must exist before the release.
	createdSecrets, err := trinoSecrets(ctx, locals, kubernetesProvider, namespaceDeps)
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
		// Wait for the coordinator and worker rollouts — an install
		// whose coordinator cannot start (bad catalog properties, a
		// missing referenced Secret) should fail THIS deploy, not the
		// first query. SkipAwait false is Helm --wait, stated
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
		return errors.Wrap(err, "failed to install trino helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles — Service names and
// Secret NAMES; values stay in-cluster.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	// The credential handles are honest: on the bring-your-own
	// password-file arm (and with auth disabled) no module-owned admin
	// exists, so the handles export EMPTY rather than names that point
	// at nothing (Terraform twin exports the same empties).
	adminSecretName, adminSecretKey := "", ""
	if locals.ModuleOwnedPasswordDb {
		adminSecretName, adminSecretKey = locals.PasswordDbSecretName, vars.AdminPasswordKey
	}
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpCoordinatorService, pulumi.String(locals.CoordinatorService))
	ctx.Export(OpCoordinatorEndpoint, pulumi.String(locals.CoordinatorEndpoint))
	ctx.Export(OpAdminUsername, pulumi.String(locals.AdminUsername))
	ctx.Export(OpAdminPasswordSecretName, pulumi.String(adminSecretName))
	ctx.Export(OpAdminPasswordSecretKey, pulumi.String(adminSecretKey))
	ctx.Export(OpWorkerService, pulumi.String(locals.WorkerService))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

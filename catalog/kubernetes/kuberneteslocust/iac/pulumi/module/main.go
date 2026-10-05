package module

import (
	"github.com/pkg/errors"
	kuberneteslocustv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslocust/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Locust from the deliveryhero Helm chart (the
// OFFICIAL locustio/locust image) as a real Helm release. The typed spec
// renders into chart values (values.go); the test scripts and the
// module-owned login backend travel through ConfigMaps composed BEFORE
// the release (scripts.go); the login credential lives in a module-owned
// Secret projected as files (secrets.go) — nothing credential-bearing
// appears in rendered values or process arguments; the helm_values
// escape hatch merges last with Helm -f semantics — the exact semantic
// twin of the Terraform module's helm_release with values = [typed,
// helm_values, re-pins].
//
// OCI WIRING: the chart's serving home is the OCI registry (the classic
// index stalls at 0.31.6) — Pulumi takes the joined
// "oci://ghcr.io/deliveryhero/helm-charts/locust" string as the chart
// reference with NO RepositoryOpts; the Terraform twin passes
// repository + bare chart name. Same chart bytes, different wiring.
func Resources(ctx *pulumi.Context, iacInput *kuberneteslocustv1alpha1.KubernetesLocustIacInput) error {
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

	// --------------------- module-owned scripts + secret ------------------
	// The pods mount them at startup, so they must exist before the
	// release.
	createdConfigMaps, err := scriptConfigMaps(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create script configmaps")
	}
	createdSecrets, err := webAuthSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create web-ui auth secret")
	}

	releaseDependencies := make([]pulumi.Resource, 0, len(createdConfigMaps)+len(createdSecrets)+1)
	if createdNamespace != nil {
		releaseDependencies = append(releaseDependencies, createdNamespace)
	}
	releaseDependencies = append(releaseDependencies, createdConfigMaps...)
	releaseDependencies = append(releaseDependencies, createdSecrets...)

	// ------------------------------ helm release --------------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		// OCI chart reference — joined string, no RepositoryOpts (see
		// the module comment above).
		Chart:   pulumi.String(vars.HelmOciRepo + "/" + vars.HelmChartName),
		Version: pulumi.String(vars.ChartVersion),
		Values:  pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the master and worker rollouts — an install whose
		// pods cannot start (a broken locustfile import, a missing
		// referenced Secret, a failing pip install) should fail THIS
		// deploy, not the first test run. SkipAwait false is Helm
		// --wait, stated explicitly to mirror the Terraform twin's
		// `wait = true`.
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
		return errors.Wrap(err, "failed to install locust helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles — Service names and
// Secret NAMES; values stay in-cluster.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	// The credential handles are honest: with the login disabled (or a
	// headless run, which starts no web UI) no module-owned credential
	// exists, so the handles export EMPTY rather than names that point
	// at nothing (Terraform twin exports the same empties).
	authSecretName, authSecretKey, username := "", "", ""
	if locals.WebLoginEnabled {
		authSecretName, authSecretKey, username = locals.AuthSecretName, vars.AuthPasswordKey, locals.WebUsername
	}
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpMasterService, pulumi.String(locals.MasterService))
	ctx.Export(OpWebEndpoint, pulumi.String(locals.WebEndpoint))
	ctx.Export(OpMasterBindEndpoint, pulumi.String(locals.MasterBindEndpoint))
	ctx.Export(OpWebUiUsername, pulumi.String(username))
	ctx.Export(OpWebUiPasswordSecretName, pulumi.String(authSecretName))
	ctx.Export(OpWebUiPasswordSecretKey, pulumi.String(authSecretKey))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

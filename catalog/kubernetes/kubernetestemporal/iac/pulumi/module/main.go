package module

import (
	"github.com/pkg/errors"
	kubernetestemporalv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestemporal/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Temporal from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); database
// passwords ride the chart's existingSecret contract (a secretKeyRef the
// server and schema Jobs resolve at runtime — never rendered as values);
// the helm_values escape hatch merges last with Helm -f semantics — the
// exact semantic twin of the Terraform module's helm_release with
// values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kubernetestemporalv1alpha1.KubernetesTemporalIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY on names past the chart's fullname budget: child names
	// are `<fullname>-<component>` and the chart's componentname helper
	// truncates the FULLNAME to fit 63 characters — the longest
	// component ("internal-frontend", 17 chars) would silently truncate
	// the fullname past 45, breaking the naming contract every exported
	// output is built on. Twin: the Terraform module's lifecycle
	// precondition.
	if len(locals.ReleaseName) > vars.FullnameBudget {
		return errors.Errorf(
			"resource name %q is %d characters — the temporal chart derives child names from it and silently truncates past %d characters, which would break the naming contract; use a name of at most %d characters",
			locals.ReleaseName, len(locals.ReleaseName), vars.FullnameBudget, vars.FullnameBudget)
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
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the cluster to become Ready — a server that never
		// starts (empty schema, unreachable database, wrong credential
		// Secret name) should fail THIS deploy, not the first workflow.
		// The pre-install schema Jobs run inside this budget too (Helm
		// hooks execute before the release resources). SkipAwait false
		// is Helm --wait, stated explicitly to mirror the Terraform
		// twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install temporal helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The frontend Service is
// `<name>-frontend` (fullnameOverride pins the fullname); the web handles
// are empty when the UI is disabled.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpFrontendService, pulumi.String(locals.FrontendServiceName))
	ctx.Export(OpFrontendEndpoint, pulumi.String(locals.FrontendEndpoint))
	ctx.Export(OpFrontendHttpEndpoint, pulumi.String(locals.FrontendHttpEndpoint))
	ctx.Export(OpWebUiService, pulumi.String(locals.WebUiServiceName))
	ctx.Export(OpWebUiEndpoint, pulumi.String(locals.WebUiEndpoint))
	ctx.Export(OpPortForwardFrontendCommand, pulumi.String(locals.PortForwardFrontendCommand))
	ctx.Export(OpPortForwardWebUiCommand, pulumi.String(locals.PortForwardWebUiCommand))
}

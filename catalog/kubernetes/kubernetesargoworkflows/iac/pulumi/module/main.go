package module

import (
	"github.com/pkg/errors"
	kubernetesargoworkflowsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesargoworkflows/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Argo Workflows from the official Helm chart as a real
// Helm release. The typed spec renders into chart values (values.go);
// artifact-store and archive-database credentials ride the chart's own
// secret SELECTORS (resolved by the workloads at runtime — never rendered
// as values); the helm_values escape hatch merges last with Helm -f
// semantics — the exact semantic twin of the Terraform module's
// helm_release with values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kubernetesargoworkflowsv1alpha1.KubernetesArgoWorkflowsIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY on names past the chart's fullname budget: every child
	// name is `<fullname>-<component>` truncated at 63 characters — the
	// longest component suffix ("-workflow-controller", 20 chars) would
	// truncate SILENTLY past 43, breaking the naming contract every
	// exported output is built on. Twin: the Terraform module's lifecycle
	// precondition.
	if len(locals.ReleaseName) > vars.FullnameBudget {
		return errors.Errorf(
			"resource name %q is %d characters — the argo-workflows chart derives child names from it and silently truncates past %d characters, which would break the naming contract; use a name of at most %d characters",
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
		// Wait for the engine to become Ready — a controller that never
		// starts (bad archive credentials Secret name, unreachable
		// database) should fail THIS deploy, not the first workflow
		// submission. The budget covers the full-schema CRD hook Job's
		// download-and-apply on the default arm. SkipAwait false is Helm
		// --wait, stated explicitly to mirror the Terraform twin's
		// `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install argo-workflows helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The server Service is
// `<name>-server` (fullnameOverride pins the fullname); both server
// handles are empty when the server is disabled. The runner ServiceAccount
// is the identity to annotate for IRSA/workload identity.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpServerService, pulumi.String(locals.ServerServiceName))
	ctx.Export(OpServerKubeEndpoint, pulumi.String(locals.ServerKubeEndpoint))
	ctx.Export(OpWorkflowServiceAccount, pulumi.String(locals.WorkflowServiceAccount))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

package module

import (
	"github.com/pkg/errors"
	kubernetesargocdv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesargocd/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Argo CD from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); the
// initial admin password stays APPLICATION-owned (Argo CD generates it at
// first start into the fixed-name `argocd-initial-admin-secret`); SSO
// client secrets ride Argo CD's `$<secret>:<key>` runtime indirection
// against labeled Secrets; the helm_values escape hatch merges last with
// Helm -f semantics — the exact semantic twin of the Terraform module's
// helm_release with values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kubernetesargocdv1alpha1.KubernetesArgocdIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY on names past the chart's fullname budget: every child
	// name is `<fullname>-<component>` truncated at 63 characters — the
	// longest component suffix ("-applicationset-controller", 26 chars)
	// would truncate SILENTLY past 37, breaking the naming contract every
	// exported output is built on. Twin: the Terraform module's lifecycle
	// precondition.
	if len(locals.ReleaseName) > vars.FullnameBudget {
		return errors.Errorf(
			"resource name %q is %d characters — the argo-cd chart derives child names from it and silently truncates past %d characters, which would break the naming contract; use a name of at most %d characters",
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
		// Wait for the control plane to become Ready — a server that
		// never starts (bad OIDC issuer, unschedulable redis-ha, missing
		// external Redis Secret) should fail THIS deploy, not the first
		// login. The budget covers all seven components plus the
		// redis-secret-init hook Job. SkipAwait false is Helm --wait,
		// stated explicitly to mirror the Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install argo-cd helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The server Service is
// `<name>-server` (fullnameOverride pins the fullname). The initial admin
// Secret's name is fixed by the APPLICATION — exported only while the
// admin user is enabled.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpServerService, pulumi.String(locals.ServerServiceName))
	ctx.Export(OpServerKubeEndpoint, pulumi.String(locals.ServerKubeEndpoint))
	ctx.Export(OpInitialAdminSecretName, pulumi.String(locals.InitialAdminSecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

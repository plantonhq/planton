package module

import (
	"github.com/pkg/errors"
	kubernetesnatsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesnats/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs NATS from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); declared
// users' passwords are module-generated into the `<name>-auth` Secret
// BEFORE the release and reach the server as secretKeyRef env vars the
// rendered config references (`$NATS_PW_<i>`) — nothing credential-bearing
// transits values; the helm_values escape hatch merges last with Helm -f
// semantics — the exact semantic twin of the Terraform module's
// helm_release with values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kubernetesnatsv1alpha1.KubernetesNatsIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// FAIL LOUDLY on names past the chart's fullname budget: derived
	// child names (`<fullname>-box-contents` is the longest, +13 chars)
	// truncate SILENTLY at 63 characters past a 50-character fullname,
	// breaking the naming contract every exported output is built on.
	// Twin: the Terraform module's lifecycle precondition.
	if len(locals.ReleaseName) > vars.FullnameBudget {
		return errors.Errorf(
			"resource name %q is %d characters — the nats chart derives child names from it and silently truncates past %d characters, which would break the naming contract; use a name of at most %d characters",
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

	var namespaceDep []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDep = append(namespaceDep, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ auth secret ---------------------------
	// Materialized BEFORE the release: the server pods mount the
	// passwords as env vars at first start.
	createdAuthSecret, err := authSecret(ctx, locals, kubernetesProvider, namespaceDep)
	if err != nil {
		return errors.Wrap(err, "failed to create auth secret")
	}

	releaseDeps := namespaceDep
	if createdAuthSecret != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdAuthSecret}))
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
		// Wait for the StatefulSet to become Ready — a server that never
		// starts (bad TLS Secret name, malformed config merge) should
		// fail THIS deploy, not the first client connection. SkipAwait
		// false is Helm --wait, stated explicitly to mirror the
		// Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(vars.HelmTimeoutSeconds),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install nats helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The client Service is
// `<name>` (fullnameOverride pins the fullname); the websocket endpoint
// and auth Secret name are empty when their features are off.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpServiceName, pulumi.String(locals.ServiceName))
	ctx.Export(OpHeadlessServiceName, pulumi.String(locals.HeadlessServiceName))
	ctx.Export(OpClientEndpoint, pulumi.String(locals.ClientEndpoint))
	ctx.Export(OpWebsocketEndpoint, pulumi.String(locals.WebsocketEndpoint))
	ctx.Export(OpAuthSecretName, pulumi.String(locals.AuthSecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

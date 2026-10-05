package module

import (
	"github.com/pkg/errors"
	kubernetessignozv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessignoz/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs SigNoz from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); the
// telemetry store is a COMPOSED ClickHouse connection (the bundled
// subchart stays permanently off — nothing ClickHouse-related installs,
// and the release carries no operator and no CRDs, so uninstall is
// ordinary object deletion); the ClickHouse password reaches the server
// as a secretKeyRef (never a rendered value); the helm_values escape
// hatch merges last with Helm -f semantics — the exact semantic twin of
// the Terraform module's helm_release with
// values = [typed, helm_values, re-pin].
func Resources(ctx *pulumi.Context, iacInput *kubernetessignozv1alpha1.KubernetesSignozIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// The collector Deployment (`<name>-otel-collector`) is the longest
	// fullname-derived child; its pod names must fit Kubernetes'
	// 63-character cap — an over-long resource name corrupts the naming
	// contract the outputs promise. Fail THIS deploy loudly instead
	// (twin: the Terraform module's plan-time precondition).
	if len(locals.ReleaseName) > vars.MaxNameLength {
		return errors.Errorf(
			"metadata.name %q is %d characters — the signoz chart's child-name budget allows at most %d "+
				"(the collector composes names like <name>-otel-collector-<replicaset>-<pod> within Kubernetes' 63-character cap)",
			locals.ReleaseName, len(locals.ReleaseName), vars.MaxNameLength)
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

	// ------------------------------ helm values ---------------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	// ------------------------------ helm release --------------------------
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
		// Wait for the whole platform to become Ready — a SigNoz whose
		// schema migration fails against the composed ClickHouse should
		// fail THIS deploy, not the first trace query. SkipAwait false is
		// Helm --wait, stated explicitly to mirror the Terraform twin's
		// `wait = true`; the migrator's own `migrate ready` init container
		// blocks until the telemetry store answers, so the budget also
		// absorbs a ClickHouse that is still coming up.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(1200),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install signoz helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. Every child name
// derives from the fullname pinned via fullnameOverride (the release
// name); the ClickHouse handles mirror the declared connection.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpService, pulumi.String(locals.Service))
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
	ctx.Export(OpOtelCollectorService, pulumi.String(locals.OtelCollectorService))
	ctx.Export(OpOtlpGrpcEndpoint, pulumi.String(locals.OtlpGrpcEndpoint))
	ctx.Export(OpOtlpHttpEndpoint, pulumi.String(locals.OtlpHttpEndpoint))
	ctx.Export(OpClickHouseEndpoint, pulumi.String(locals.ClickHouseEndpoint))
	ctx.Export(OpClickHouseUsername, pulumi.String(locals.ClickHouseUsername))
	ctx.Export(OpClickHousePasswordSecretName, pulumi.String(locals.ClickHousePasswordSecretName))
	ctx.Export(OpClickHousePasswordSecretKey, pulumi.String(locals.ClickHousePasswordSecretKey))
}

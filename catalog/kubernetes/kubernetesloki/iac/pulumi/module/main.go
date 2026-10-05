package module

import (
	"github.com/pkg/errors"
	kuberneteslokiv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesloki/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Grafana Loki from the official Helm chart as a real
// Helm release. The typed spec renders into chart values (values.go);
// exactly one deployment mode renders with every other mode's workloads
// zeroed; declared object-store credentials ride environment variables
// sourced from Secrets (the r2 arm's from the module-owned
// `<name>-r2-credentials` Secret) so no credential ever lands in the
// chart's rendered configuration; the helm_values escape hatch merges last with Helm -f
// semantics — the exact semantic twin of the Terraform module's
// helm_release with values = [typed, helm_values, re-pin].
func Resources(ctx *pulumi.Context, iacInput *kuberneteslokiv1alpha1.KubernetesLokiIacInput) error {
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

	// ------------------------ r2 credentials Secret -------------------------
	// Created BEFORE the release: Loki's credential variables read it
	// through secretKeyRef, and a pod whose referenced key is missing never
	// starts. The values arrive resolved from references; the provider
	// stores Secret data as Pulumi secrets.
	if r2Arm := locals.R2; r2Arm != nil {
		createdSecret, err := kubernetescorev1.NewSecret(ctx,
			locals.R2CredentialsSecretName,
			&kubernetescorev1.SecretArgs{
				Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
					Name:      pulumi.String(locals.R2CredentialsSecretName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				}),
				StringData: pulumi.ToSecret(pulumi.ToStringMap(r2Arm.SecretData)).(pulumi.StringMapOutput),
			},
			append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)...)
		if err != nil {
			return errors.Wrap(err, "failed to create r2 credentials secret")
		}
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdSecret}))
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
		// Wait for Loki to become Ready — a log store whose ingesters
		// never bind their storage or whose gateway never starts should
		// fail THIS deploy, not the first push. SkipAwait false is Helm
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
		return errors.Wrap(err, "failed to install loki helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. Every child name derives
// from the fullname pinned to the resource name via fullnameOverride, so the
// gateway Service and the Loki HTTP Service are deterministic.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpGatewayService, pulumi.String(locals.GatewayService))
	ctx.Export(OpGatewayEndpoint, pulumi.String(locals.GatewayEndpoint))
	ctx.Export(OpOtlpPushEndpoint, pulumi.String(locals.OtlpPushEndpoint))
	ctx.Export(OpLokiService, pulumi.String(locals.LokiService))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

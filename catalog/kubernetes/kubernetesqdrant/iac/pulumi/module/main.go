package module

import (
	"github.com/pkg/errors"
	kubernetesqdrantv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesqdrant/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Qdrant from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); API keys
// stay chart-owned (generated once via the chart's lookup, or read from an
// existing Secret AT TEMPLATE TIME); the helm_values escape hatch merges
// last with Helm -f semantics — the exact semantic twin of the Terraform
// module's helm_release with values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kubernetesqdrantv1alpha1.KubernetesQdrantIacInput) error {
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
		// Wait for the cluster to become Ready — a database that never
		// starts (bad image, unschedulable pod, unbindable volume) should
		// fail THIS deploy, not the first client connection. Multi-node
		// consensus bootstrap is quick; storage recovery on big volumes
		// is what the generous budget covers. SkipAwait false is Helm
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
		return errors.Wrap(err, "failed to install qdrant helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The service name is the
// chart's main ClusterIP Service — qdrant.fullname, pinned to the resource
// name via fullnameOverride. The API-key Secret names point at the
// chart-owned `<name>-apikey` Secret (keys api-key / read-only-api-key),
// which the chart populates for the generate and existing-secret arms
// alike.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpServiceName, pulumi.String(locals.ServiceName))
	ctx.Export(OpHttpEndpoint, pulumi.String(locals.HttpEndpoint))
	ctx.Export(OpGrpcEndpoint, pulumi.String(locals.GrpcEndpoint))
	ctx.Export(OpApiKeySecretName, pulumi.String(locals.ApiKeySecretName))
	ctx.Export(OpReadOnlyApiKeySecretName, pulumi.String(locals.ReadOnlyApiKeySecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

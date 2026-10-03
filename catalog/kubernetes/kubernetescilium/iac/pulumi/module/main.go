package module

import (
	"github.com/pkg/errors"
	kubernetesciliumv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetescilium/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Cilium from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); the
// helm_values escape hatch merges last with Helm -f semantics — the exact
// semantic twin of the Terraform module's helm_release with
// values = [typed, helm_values].
//
// The release name is FIXED ("cilium"): Cilium is the node dataplane — the
// agent DaemonSet, operator, and generated CNI configuration are cluster
// singletons, so one dataplane per cluster is an upstream constraint.
func Resources(ctx *pulumi.Context, iacInput *kubernetesciliumv1alpha1.KubernetesCiliumIacInput) error {
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
		Name:      pulumi.String(vars.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for the whole dataplane to come up. 600s (not the default
		// 300) because the install path is heavier than an ordinary
		// workload chart: the agent DaemonSet must roll out on EVERY node
		// plus the operator, and on a fresh cluster nodes transition
		// NotReady->Ready only as Cilium wires each one — the rollout
		// itself unblocks scheduling. A dataplane that never converges
		// should fail THIS deploy, not surface later as pods stuck in
		// ContainerCreating.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, vars.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install cilium helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(vars.ReleaseName))
	ctx.Export(OpClusterName, pulumi.String(locals.ClusterName))
	// Fixed chart-template names (empty when the component is off) —
	// resolved in locals.
	ctx.Export(OpHubbleRelayServiceName, pulumi.String(locals.HubbleRelayServiceName))
	ctx.Export(OpHubbleUiServiceName, pulumi.String(locals.HubbleUiServiceName))
	ctx.Export(OpGatewayClassName, pulumi.String(locals.GatewayClassName))

	return nil
}

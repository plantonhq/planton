package module

import (
	"github.com/pkg/errors"
	kubernetesopensearchv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopensearch/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one operator-managed OpenSearch cluster:
//
//  1. the namespace (optional, create_namespace),
//  2. the OpenSearchCluster resource itself (the typed SDK catches
//     field/structure drift against the pinned CRD at compile time) — the
//     ONLY custom resource: node StatefulSets, Services, TLS Secrets, the
//     admin bootstrap and the optional Dashboards deployment are all
//     operator-created from it. No ingress resources — exposure
//     composes from first-class kinds referencing the exported handles.
func Resources(ctx *pulumi.Context, iacInput *kubernetesopensearchv1alpha1.KubernetesOpenSearchIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	if _, err := createCluster(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create opensearch cluster")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The credential handle
// points at the operator-generated `<name>-admin-password` Secret — empty
// when a custom security config replaces the operator bootstrap. Dashboards
// handles are populated only when dashboards are enabled.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpClusterName, pulumi.String(locals.ClusterName))
	ctx.Export(OpServiceName, pulumi.String(locals.ClusterName))
	ctx.Export(OpHttpEndpoint, pulumi.String(locals.HttpEndpoint))
	ctx.Export(OpAdminCredentialsSecretName, pulumi.String(locals.AdminCredentialsSecretName))
	ctx.Export(OpDashboardsServiceName, pulumi.String(locals.DashboardsServiceName))
	ctx.Export(OpDashboardsEndpoint, pulumi.String(locals.DashboardsEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

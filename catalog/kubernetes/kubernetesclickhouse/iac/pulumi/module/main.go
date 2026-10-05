package module

import (
	"github.com/pkg/errors"
	kubernetesclickhousev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesclickhouse/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one operator-managed ClickHouse cluster:
//
//  1. the namespace (optional, create_namespace),
//  2. the auth Secret (only when spec.users is non-empty) — one key per
//     user; CHI users reference it via valueFrom.secretKeyRef,
//  3. the managed ClickHouseKeeperInstallation (only when the coordination
//     contract calls for one),
//  4. the ClickHouseInstallation itself (the typed SDK catches
//     field/structure drift against the pinned CRD at compile time) — every
//     host StatefulSet, Service and generated ConfigMap is operator-created
//     from it. No ingress resources — exposure composes from first-class
//     kinds referencing the exported handles.
func Resources(ctx *pulumi.Context, iacInput *kubernetesclickhousev1alpha1.KubernetesClickHouseIacInput) error {
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

	createdAuthSecret, err := authSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create auth secret")
	}

	if _, err := keeper(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create clickhouse keeper installation")
	}

	if _, err := installation(ctx, locals, kubernetesProvider, createdAuthSecret, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create clickhouse installation")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The auth-secret handle
// is empty when no users are declared; the Keeper handles are empty unless
// the module deployed a managed Keeper (external/none coordination exports
// nothing to manage).
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpChiName, pulumi.String(locals.ChiName))
	ctx.Export(OpClusterName, pulumi.String(locals.ClusterName))
	ctx.Export(OpServiceName, pulumi.String(locals.ServiceName))
	ctx.Export(OpTcpEndpoint, pulumi.String(locals.TcpEndpoint))
	ctx.Export(OpHttpEndpoint, pulumi.String(locals.HttpEndpoint))
	ctx.Export(OpAuthSecretName, pulumi.String(locals.AuthSecretName))
	ctx.Export(OpKeeperName, pulumi.String(locals.KeeperName))
	ctx.Export(OpKeeperServiceName, pulumi.String(locals.KeeperServiceName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

package module

import (
	"github.com/pkg/errors"
	kubernetessolrv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessolr/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one Apache Solr Operator-managed SolrCloud cluster:
//
//  1. the namespace (optional),
//  2. the SolrCloud resource itself (the typed SDK catches field/structure
//     drift against the pinned CRD at compile time).
//
// Everything else — the node StatefulSet, the provided ZooKeeper ensemble,
// services, PVCs, the basic-auth bootstrap Secret, Ingress exposure — is
// the operator's to create from the SolrCloud spec; the module renders the
// CR and exports the operator's deterministic names.
func Resources(ctx *pulumi.Context, iacInput *kubernetessolrv1alpha1.KubernetesSolrIacInput) error {
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

	if _, err := createSolrCloud(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create solrcloud resource")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles — all derived from the
// operator's naming contract (see locals.go), never read back from the
// cluster.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpClusterName, pulumi.String(locals.ClusterName))
	ctx.Export(OpCommonServiceName, pulumi.String(locals.CommonServiceName))
	ctx.Export(OpInternalEndpoint, pulumi.String(locals.InternalEndpoint))
	ctx.Export(OpBasicAuthSecretName, pulumi.String(locals.BasicAuthSecretName))
	ctx.Export(OpZookeeperConnectionString, pulumi.String(locals.ZookeeperConnectionString))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}

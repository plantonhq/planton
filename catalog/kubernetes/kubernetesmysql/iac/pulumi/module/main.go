package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesmysqlv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesmysql/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one Percona-operator-managed MySQL (XtraDB Cluster):
//
//  1. the namespace (optional),
//  2. declared-credential Secrets (user passwords, backup-storage keys)
//     — secrets always travel via secret references, never inline in a
//     custom resource,
//  3. the PerconaXtraDBCluster resource itself (an untyped CustomResource
//     whose spec body twins the Terraform module's mysql_manifest —
//     cluster.go records why the typed crd2pulumi path is unusable),
//  4. outputs for composition (proxy Services, root password Secret).
//
// Ordering matters only for the namespace (everything is namespaced)
// and for credential Secrets (the operator reads them at reconcile time).
func Resources(ctx *pulumi.Context, iacInput *kubernetesmysqlv1alpha1.KubernetesMysqlIacInput) error {
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

	credentialSecrets, err := createCredentialSecrets(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create credential secrets")
	}

	clusterDeps := namespaceDeps
	if len(credentialSecrets) > 0 {
		clusterDeps = append(clusterDeps, pulumi.DependsOn(credentialSecrets))
	}

	if _, err := createCluster(ctx, locals, kubernetesProvider, clusterDeps); err != nil {
		return errors.Wrap(err, "failed to create percona xtradb cluster")
	}

	exportOutputs(ctx, locals)
	return nil
}

func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpClusterName, pulumi.String(locals.ClusterName))
	ctx.Export(OpPrimaryService, pulumi.String(locals.PrimaryServiceName))
	ctx.Export(OpReplicasService, pulumi.String(locals.ReplicasServiceName))
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(fmt.Sprintf(
		"kubectl port-forward svc/%s -n %s 3306:3306",
		locals.PrimaryServiceName, locals.Namespace)))
	ctx.Export(OpRootPasswordSecretName, pulumi.String(locals.RootPasswordSecretName))
	ctx.Export(OpRootPasswordSecretKey, pulumi.String("root"))
}

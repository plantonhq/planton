package module

import (
	"github.com/pkg/errors"
	awsdocumentdbv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsdocumentdb/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/docdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the DocumentDB cluster and its folded compute. The
// cluster is the shared-storage brain (endpoints, credentials, backups,
// encryption, engine lifecycle); the spec's instances list is the compute
// that serves queries, each entry managed as its own provider resource
// keyed by name. Subnets, security groups, and KMS keys compose by
// reference -- this module never creates or mutates a resource that
// deserves to be its own node.
func Resources(ctx *pulumi.Context, iacInput *awsdocumentdbv1alpha1.AwsDocumentDbIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless
	// web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsDocumentDb.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create subnet group")
	}

	createdParameterGroup, err := clusterParameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create cluster parameter group")
	}

	createdCluster, err := docdbCluster(ctx, locals, provider, createdSubnetGroup, createdParameterGroup)
	if err != nil {
		return errors.Wrap(err, "failed to create DocumentDB cluster")
	}

	createdInstances, err := clusterInstances(ctx, locals, provider, createdCluster)
	if err != nil {
		return errors.Wrap(err, "failed to create cluster instances")
	}

	ctx.Export(OpClusterIdentifier, createdCluster.ClusterIdentifier)
	ctx.Export(OpArn, createdCluster.Arn)
	ctx.Export(OpClusterResourceId, createdCluster.ClusterResourceId)
	ctx.Export(OpEndpoint, createdCluster.Endpoint)
	ctx.Export(OpReaderEndpoint, createdCluster.ReaderEndpoint)
	ctx.Export(OpPort, createdCluster.Port)
	ctx.Export(OpHostedZoneId, createdCluster.HostedZoneId)
	ctx.Export(OpEngineVersionActual, createdCluster.EngineVersion)

	// The AWS-managed master-user secret exists only when
	// manage_master_user_password is on; export empty otherwise so the
	// output shape is stable across both password strategies.
	ctx.Export(OpMasterUserSecretArn, createdCluster.MasterUserSecrets.ApplyT(func(secrets []docdb.ClusterMasterUserSecret) string {
		if len(secrets) == 0 || secrets[0].SecretArn == nil {
			return ""
		}
		return *secrets[0].SecretArn
	}).(pulumi.StringOutput))

	ctx.Export(OpDbSubnetGroupName, createdCluster.DbSubnetGroupName)
	ctx.Export(OpDbClusterParameterGroupName, createdCluster.DbClusterParameterGroupName)

	// Per-instance endpoints in spec order -- empty for headless shapes
	// (restores and global-cluster members created without instances).
	instanceEndpoints := make(pulumi.StringArray, 0, len(createdInstances))
	for _, createdInstance := range createdInstances {
		instanceEndpoints = append(instanceEndpoints, createdInstance.Endpoint)
	}
	ctx.Export(OpInstanceEndpoints, instanceEndpoints)

	return nil
}

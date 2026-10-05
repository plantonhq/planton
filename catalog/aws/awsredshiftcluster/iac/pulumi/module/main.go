package module

import (
	"github.com/pkg/errors"
	awsredshiftclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsredshiftcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the Redshift cluster and its folded settings. The
// cluster composes onto its neighbors instead of embedding them: subnets,
// security groups, IAM roles, KMS keys, and the Elastic IP attach by
// reference, and warehouse ingress rules live on the referenced
// AwsSecurityGroup nodes -- this module never creates or mutates a
// resource that deserves to be its own node. Audit logging and
// cross-region snapshot copy are cluster settings keyed by the cluster
// itself, so they are managed here rather than modeled as standalone
// kinds.
func Resources(ctx *pulumi.Context, iacInput *awsredshiftclusterv1alpha1.AwsRedshiftClusterIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless
	// web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsRedshiftCluster.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create subnet group")
	}

	createdParameterGroup, err := clusterParameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create parameter group")
	}

	createdCluster, err := redshiftCluster(ctx, locals, provider, createdSubnetGroup, createdParameterGroup)
	if err != nil {
		return errors.Wrap(err, "failed to create Redshift cluster")
	}

	if err := clusterLogging(ctx, locals, provider, createdCluster); err != nil {
		return errors.Wrap(err, "failed to configure audit logging")
	}

	if err := clusterSnapshotCopy(ctx, locals, provider, createdCluster); err != nil {
		return errors.Wrap(err, "failed to configure snapshot copy")
	}

	if err := snapshotScheduleAssociation(ctx, locals, provider, createdCluster); err != nil {
		return errors.Wrap(err, "failed to associate snapshot schedule")
	}

	createdUsageLimitIds, err := usageLimits(ctx, locals, provider, createdCluster)
	if err != nil {
		return errors.Wrap(err, "failed to create usage limits")
	}

	if err := scheduledActions(ctx, locals, provider, createdCluster); err != nil {
		return errors.Wrap(err, "failed to create scheduled actions")
	}

	createdEndpointAddresses, err := endpointAccesses(ctx, locals, provider, createdCluster)
	if err != nil {
		return errors.Wrap(err, "failed to create endpoint accesses")
	}

	if err := endpointAuthorizations(ctx, locals, provider, createdCluster); err != nil {
		return errors.Wrap(err, "failed to create endpoint authorizations")
	}

	ctx.Export(OpClusterIdentifier, createdCluster.ClusterIdentifier)
	ctx.Export(OpClusterArn, createdCluster.Arn)
	ctx.Export(OpClusterNamespaceArn, createdCluster.ClusterNamespaceArn)
	ctx.Export(OpEndpoint, createdCluster.Endpoint)
	ctx.Export(OpDnsName, createdCluster.DnsName)

	// Empty when the spec omitted database_name: AWS creates its
	// documented default initial database ("dev") but DescribeClusters
	// echoes no name back, so the attribute stays unset (live-verified
	// on both engines).
	ctx.Export(OpDatabaseName, createdCluster.DatabaseName)
	ctx.Export(OpPort, createdCluster.Port)

	// Group names come from the cluster's own attributes so the output
	// shape is identical whether the groups are managed here, referenced,
	// or left to the Redshift defaults.
	ctx.Export(OpSubnetGroupName, createdCluster.ClusterSubnetGroupName)
	ctx.Export(OpParameterGroupName, createdCluster.ClusterParameterGroupName)

	// The AWS-managed admin-password secret exists only when
	// manage_master_password is on; the attribute resolves to "" otherwise,
	// so the output shape is stable across both password strategies.
	ctx.Export(OpMasterPasswordSecretArn, createdCluster.MasterPasswordSecretArn)

	// Per-satellite maps: endpoint addresses and AWS-generated usage-limit
	// IDs, keyed identically on both engines (imports and out-of-band CLI
	// operations address entries by these keys).
	ctx.Export(OpEndpointAccessAddresses, createdEndpointAddresses)
	ctx.Export(OpUsageLimitIds, createdUsageLimitIds)

	return nil
}

package module

import (
	"github.com/pkg/errors"
	awsredshiftclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsredshiftcluster/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS Redshift Cluster related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsredshiftclusterv1.AwsRedshiftClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsRedshiftCluster.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// Security group (ingress from SGs and/or CIDRs)
	createdSg, err := securityGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "security group")
	}

	// Subnet group (only when subnetIds provided and no name supplied)
	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "subnet group")
	}

	// Parameter group (when inline parameters provided)
	createdParamGroup, err := parameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "parameter group")
	}

	// Create the Redshift Cluster
	cluster, err := redshiftCluster(ctx, locals, provider, createdSg, createdSubnetGroup, createdParamGroup)
	if err != nil {
		return errors.Wrap(err, "redshift cluster")
	}

	// Logging (conditional on spec.Logging)
	err = redshiftLogging(ctx, locals, provider, cluster)
	if err != nil {
		return errors.Wrap(err, "redshift logging")
	}

	// Export outputs as defined in AwsRedshiftClusterStackOutputs
	ctx.Export(OpClusterIdentifier, cluster.ID())
	ctx.Export(OpClusterArn, cluster.Arn)
	ctx.Export(OpClusterNamespaceArn, cluster.ClusterNamespaceArn)
	ctx.Export(OpEndpoint, cluster.Endpoint)
	ctx.Export(OpDnsName, cluster.DnsName)
	ctx.Export(OpDatabaseName, cluster.DatabaseName)
	ctx.Export(OpPort, cluster.Port)
	ctx.Export(OpMasterPasswordSecretArn, cluster.MasterPasswordSecretArn)
	if createdSubnetGroup != nil {
		ctx.Export(OpSubnetGroupName, createdSubnetGroup.Name)
	}
	if createdSg != nil {
		ctx.Export(OpSecurityGroupId, createdSg.ID())
	}
	if createdParamGroup != nil {
		ctx.Export(OpParameterGroupName, createdParamGroup.Name)
	}
	return nil
}

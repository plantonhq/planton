package module

import (
	"github.com/pkg/errors"
	awsredshiftclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsredshiftcluster/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS Redshift Cluster related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsredshiftclusterv1.AwsRedshiftClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsRedshiftCluster.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsRedshiftCluster.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
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

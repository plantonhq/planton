package module

import (
	"github.com/pkg/errors"
	awsmskclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmskcluster/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS MSK Cluster related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsmskclusterv1.AwsMskClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsMskCluster.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsMskCluster.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	// Managed security group (ingress from SGs and/or CIDRs on Kafka + ZooKeeper ports)
	createdSg, err := securityGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "security group")
	}

	// Inline MSK Configuration (when server_properties provided)
	createdConfig, err := configuration(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "msk configuration")
	}

	// MSK Cluster
	mskCluster, err := cluster(ctx, locals, provider, createdSg, createdConfig)
	if err != nil {
		return errors.Wrap(err, "msk cluster")
	}

	// Export outputs
	ctx.Export(OpClusterArn, mskCluster.Arn)
	ctx.Export(OpClusterName, mskCluster.ClusterName)
	ctx.Export(OpClusterUuid, mskCluster.ClusterUuid)
	ctx.Export(OpCurrentVersion, mskCluster.CurrentVersion)
	ctx.Export(OpBootstrapBrokers, mskCluster.BootstrapBrokers)
	ctx.Export(OpBootstrapBrokersTls, mskCluster.BootstrapBrokersTls)
	ctx.Export(OpBootstrapBrokersSaslIam, mskCluster.BootstrapBrokersSaslIam)
	ctx.Export(OpBootstrapBrokersSaslScram, mskCluster.BootstrapBrokersSaslScram)
	ctx.Export(OpBootstrapBrokersPublicTls, mskCluster.BootstrapBrokersPublicTls)
	ctx.Export(OpBootstrapBrokersPublicSaslIam, mskCluster.BootstrapBrokersPublicSaslIam)
	ctx.Export(OpBootstrapBrokersPublicSaslScram, mskCluster.BootstrapBrokersPublicSaslScram)
	ctx.Export(OpZookeeperConnectString, mskCluster.ZookeeperConnectString)
	ctx.Export(OpZookeeperConnectStringTls, mskCluster.ZookeeperConnectStringTls)
	if createdSg != nil {
		ctx.Export(OpSecurityGroupId, createdSg.ID())
	}
	if createdConfig != nil {
		ctx.Export(OpConfigurationArn, createdConfig.Arn)
	}

	return nil
}

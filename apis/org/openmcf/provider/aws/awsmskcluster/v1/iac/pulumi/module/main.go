package module

import (
	"github.com/pkg/errors"
	awsmskclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmskcluster/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS MSK Cluster related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsmskclusterv1.AwsMskClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsMskCluster.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
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

package module

import (
	"fmt"

	"github.com/pkg/errors"
	awsneptuneclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsneptunecluster/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS Neptune cluster related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsneptuneclusterv1.AwsNeptuneClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsNeptuneCluster.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsNeptuneCluster.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	createdSg, err := securityGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "security group")
	}

	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "subnet group")
	}

	createdParamGroup, err := clusterParameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "cluster parameter group")
	}

	cluster, err := neptuneCluster(ctx, locals, provider, createdSg, createdSubnetGroup, createdParamGroup)
	if err != nil {
		return errors.Wrap(err, "neptune cluster")
	}

	err = neptuneClusterInstances(ctx, locals, provider, cluster)
	if err != nil {
		return errors.Wrap(err, "neptune cluster instances")
	}

	ctx.Export(OpClusterEndpoint, cluster.Endpoint)
	ctx.Export(OpClusterReaderEndpoint, cluster.ReaderEndpoint)
	ctx.Export(OpClusterId, cluster.ID())
	ctx.Export(OpClusterArn, cluster.Arn)
	ctx.Export(OpClusterResourceId, cluster.ClusterResourceId)
	ctx.Export(OpClusterPort, cluster.Port)
	ctx.Export(OpHostedZoneId, cluster.HostedZoneId)

	if createdSubnetGroup != nil {
		ctx.Export(OpDbSubnetGroupName, createdSubnetGroup.Name)
	}
	if createdSg != nil {
		ctx.Export(OpSecurityGroupId, createdSg.ID())
	}
	if createdParamGroup != nil {
		ctx.Export(OpClusterParameterGroupName, createdParamGroup.Name)
	}

	return nil
}

// getEffectivePort returns the port to use, either from spec or default (8182).
func getEffectivePort(spec *awsneptuneclusterv1.AwsNeptuneClusterSpec) int {
	port := spec.GetPort()
	if port == 0 {
		return 8182
	}
	return int(port)
}

// getParameterGroupFamily derives the Neptune parameter group family from the engine version.
// Neptune families: "neptune1", "neptune1.2", "neptune1.3"
func getParameterGroupFamily(engineVersion string) string {
	if engineVersion == "" {
		return "neptune1.3"
	}
	if len(engineVersion) >= 3 {
		major := engineVersion[:3]
		return fmt.Sprintf("neptune%s", major)
	}
	return fmt.Sprintf("neptune%s", engineVersion)
}

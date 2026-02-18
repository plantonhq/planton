package module

import (
	"fmt"

	"github.com/pkg/errors"
	awsdocumentdbv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsdocumentdb/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS DocumentDB cluster related resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsdocumentdbv1.AwsDocumentDbStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.AwsDocumentDb.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.AwsDocumentDb.Spec.Region),
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

	// Cluster parameter group (when parameters provided)
	createdParamGroup, err := clusterParameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "cluster parameter group")
	}

	// Create the DocumentDB Cluster
	cluster, err := docdbCluster(ctx, locals, provider, createdSg, createdSubnetGroup, createdParamGroup)
	if err != nil {
		return errors.Wrap(err, "documentdb cluster")
	}

	// Create cluster instances
	err = docdbClusterInstances(ctx, locals, provider, cluster)
	if err != nil {
		return errors.Wrap(err, "documentdb cluster instances")
	}

	// Export outputs as defined in AwsDocumentDbStackOutputs
	ctx.Export(OpClusterEndpoint, cluster.Endpoint)
	ctx.Export(OpClusterReaderEndpoint, cluster.ReaderEndpoint)
	ctx.Export(OpClusterId, cluster.ID())
	ctx.Export(OpClusterArn, cluster.Arn)
	ctx.Export(OpClusterPort, cluster.Port)
	ctx.Export(OpClusterResourceId, cluster.ClusterResourceId)

	// Build connection string
	connectionString := pulumi.Sprintf("mongodb://%s:%s@%s:%d/?tls=true&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false",
		locals.AwsDocumentDb.Spec.GetMasterUsername(),
		"<password>",
		cluster.Endpoint,
		locals.AwsDocumentDb.Spec.GetPort())
	ctx.Export(OpConnectionString, connectionString)

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

// getDefaultPort returns the default DocumentDB port
func getDefaultPort() int32 {
	return 27017
}

// getEffectivePort returns the port to use, either from spec or default
func getEffectivePort(spec *awsdocumentdbv1.AwsDocumentDbSpec) int {
	port := spec.GetPort()
	if port == 0 {
		return int(getDefaultPort())
	}
	return int(port)
}

// getEngineFamily returns the DocumentDB engine family for parameter groups
func getEngineFamily(engineVersion string) string {
	// DocumentDB parameter group family is based on engine version
	// docdb5.0, docdb4.0, docdb3.6
	if engineVersion == "" {
		return "docdb5.0"
	}
	return fmt.Sprintf("docdb%s", engineVersion[:3])
}

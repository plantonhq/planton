package module

import (
	"github.com/pkg/errors"
	awsrediselasticachev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsrediselasticache/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the ElastiCache replication group and its folded
// subnet and parameter groups. Subnets, security groups, KMS keys, and
// RBAC user groups compose by reference -- this module never creates or
// mutates a resource that deserves to be its own node.
func Resources(ctx *pulumi.Context, iacInput *awsrediselasticachev1alpha1.AwsRedisElasticacheIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless
	// web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsRedisElasticache.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create subnet group")
	}

	createdParameterGroup, err := parameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create parameter group")
	}

	createdReplicationGroup, err := replicationGroup(ctx, locals, provider, createdSubnetGroup, createdParameterGroup)
	if err != nil {
		return errors.Wrap(err, "failed to create replication group")
	}

	ctx.Export(OpReplicationGroupId, createdReplicationGroup.ReplicationGroupId)
	ctx.Export(OpPrimaryEndpointAddress, createdReplicationGroup.PrimaryEndpointAddress)
	ctx.Export(OpReaderEndpointAddress, createdReplicationGroup.ReaderEndpointAddress)
	ctx.Export(OpConfigurationEndpointAddress, createdReplicationGroup.ConfigurationEndpointAddress)
	ctx.Export(OpArn, createdReplicationGroup.Arn)
	ctx.Export(OpPort, createdReplicationGroup.Port)
	ctx.Export(OpEngineVersionActual, createdReplicationGroup.EngineVersionActual)

	// Subnet and parameter group names are exported only when this module
	// created them -- external arms are referenced, not owned, here.
	if createdSubnetGroup != nil {
		ctx.Export(OpSubnetGroupName, createdSubnetGroup.Name)
	} else {
		ctx.Export(OpSubnetGroupName, pulumi.String(""))
	}
	if createdParameterGroup != nil {
		ctx.Export(OpParameterGroupName, createdParameterGroup.Name)
	} else {
		ctx.Export(OpParameterGroupName, pulumi.String(""))
	}

	return nil
}

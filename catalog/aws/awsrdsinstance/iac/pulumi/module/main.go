package module

import (
	"github.com/pkg/errors"
	awsrdsinstancev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsrdsinstance/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the RDS DB instance: a standalone database server
// with its own EBS-backed storage -- optionally Multi-AZ with a
// synchronous standby, or a read replica of another instance. Subnets,
// security groups, KMS keys, and the monitoring role compose by
// reference -- this module never creates or mutates a resource that
// deserves to be its own node.
func Resources(ctx *pulumi.Context, iacInput *awsrdsinstancev1alpha1.AwsRdsInstanceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static keys, keyless
	// web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsRdsInstance.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdSubnetGroup, err := subnetGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create subnet group")
	}

	createdParameterGroup, err := parameterGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create DB parameter group")
	}

	createdOptionGroup, err := optionGroup(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create option group")
	}

	createdInstance, err := rdsInstance(ctx, locals, provider, createdSubnetGroup, createdParameterGroup, createdOptionGroup)
	if err != nil {
		return errors.Wrap(err, "failed to create RDS instance")
	}

	if err := roleAssociations(ctx, locals, provider, createdInstance); err != nil {
		return errors.Wrap(err, "failed to create role associations")
	}

	ctx.Export(OpInstanceIdentifier, createdInstance.Identifier)
	ctx.Export(OpArn, createdInstance.Arn)
	ctx.Export(OpResourceId, createdInstance.ResourceId)
	ctx.Export(OpEndpoint, createdInstance.Endpoint)
	ctx.Export(OpAddress, createdInstance.Address)
	ctx.Export(OpPort, createdInstance.Port)
	ctx.Export(OpHostedZoneId, createdInstance.HostedZoneId)
	ctx.Export(OpEngineVersionActual, createdInstance.EngineVersionActual)

	// The AWS-managed master-user secret exists only when
	// manage_master_user_password is on; export empty otherwise so the
	// output shape is stable across both password strategies.
	ctx.Export(OpMasterUserSecretArn, createdInstance.MasterUserSecrets.ApplyT(func(secrets []rds.InstanceMasterUserSecret) string {
		if len(secrets) == 0 || secrets[0].SecretArn == nil {
			return ""
		}
		return *secrets[0].SecretArn
	}).(pulumi.StringOutput))

	ctx.Export(OpDbSubnetGroupName, createdInstance.DbSubnetGroupName)

	// The groups in use -- the resource echoes back whichever group won
	// (managed inline, referenced, or the engine default's name).
	ctx.Export(OpDbParameterGroupName, createdInstance.ParameterGroupName)
	ctx.Export(OpOptionGroupName, createdInstance.OptionGroupName)

	return nil
}

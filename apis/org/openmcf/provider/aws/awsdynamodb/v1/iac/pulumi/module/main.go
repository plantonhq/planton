package module

import (
	"github.com/pkg/errors"
	awsdynamodbv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsdynamodb/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates DynamoDB table creation and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsdynamodbv1.AwsDynamodbStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	tbl, err := createTable(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "dynamodb table")
	}

	// Export outputs mapping to AwsDynamodbStackOutputs
	ctx.Export(OpTableName, tbl.Table.Name)
	ctx.Export(OpTableArn, tbl.Table.Arn)
	ctx.Export(OpTableId, tbl.Table.ID())
	ctx.Export(OpStreamArn, tbl.Table.StreamArn)
	ctx.Export(OpStreamLabel, tbl.Table.StreamLabel)

	return nil
}

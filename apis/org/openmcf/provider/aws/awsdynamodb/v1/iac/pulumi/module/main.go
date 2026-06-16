package module

import (
	"github.com/pkg/errors"
	awsdynamodbv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsdynamodb/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates DynamoDB table creation and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsdynamodbv1.AwsDynamodbStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
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

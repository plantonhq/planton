package module

import (
	"github.com/pkg/errors"
	awsdynamodbv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsdynamodb/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the DynamoDB table and its table-scoped
// satellites (resource policy, Kinesis change-data destination,
// contributor insights). The table composes onto its neighbors instead
// of embedding them: the KMS key that encrypts it, the Kinesis stream
// that receives its change data, and the S3 bucket an import seeds it
// from all attach by reference -- this module never creates or mutates
// a resource that deserves to be its own node.
func Resources(ctx *pulumi.Context, iacInput *awsdynamodbv1alpha1.AwsDynamodbIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared
	// builder, which resolves the right credential mechanism (static
	// keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsDynamodb.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdTable, err := table(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create DynamoDB table")
	}

	if err := satellites(ctx, locals, provider, createdTable); err != nil {
		return errors.Wrap(err, "failed to create DynamoDB table satellites")
	}

	if err := autoscaling(ctx, locals, provider, createdTable); err != nil {
		return errors.Wrap(err, "failed to configure DynamoDB capacity autoscaling")
	}

	// Stream outputs resolve to "" when streams are disabled -- the SDK
	// string outputs carry the zero value without ApplyT.
	ctx.Export(OpTableName, createdTable.Name)
	ctx.Export(OpTableArn, createdTable.Arn)
	ctx.Export(OpTableId, createdTable.ID())
	ctx.Export(OpStreamArn, createdTable.StreamArn)
	ctx.Export(OpStreamLabel, createdTable.StreamLabel)

	return nil
}

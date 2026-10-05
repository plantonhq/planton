package module

import (
	"github.com/pkg/errors"
	awsbatchjobdefinitionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbatchjobdefinition/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources registers the AWS Batch job definition and exports its outputs.
func Resources(ctx *pulumi.Context, iacInput *awsbatchjobdefinitionv1alpha1.AwsBatchJobDefinitionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsBatchJobDefinition.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdJobDefinition, err := jobDefinition(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "job definition")
	}

	// --- Exports ---
	// The revision-carrying ARN is the primary handle: it changes on every
	// registered revision, which is what rolls referencing consumers.
	ctx.Export(OpJobDefinitionArn, createdJobDefinition.Arn)
	ctx.Export(OpArnWithoutRevision, createdJobDefinition.ArnPrefix)
	ctx.Export(OpJobDefinitionName, createdJobDefinition.Name)
	ctx.Export(OpRevision, createdJobDefinition.Revision)

	return nil
}

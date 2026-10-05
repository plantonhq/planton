package module

import (
	"github.com/pkg/errors"
	awsbatchjobqueuev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbatchjobqueue/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the AWS Batch job queue and exports its outputs.
func Resources(ctx *pulumi.Context, iacInput *awsbatchjobqueuev1alpha1.AwsBatchJobQueueIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsBatchJobQueue.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdQueue, err := jobQueue(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "job queue")
	}

	// --- Exports ---
	ctx.Export(OpJobQueueArn, createdQueue.Arn)
	ctx.Export(OpJobQueueName, createdQueue.Name)

	return nil
}

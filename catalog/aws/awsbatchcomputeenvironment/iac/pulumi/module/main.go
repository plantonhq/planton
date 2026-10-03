package module

import (
	"github.com/pkg/errors"
	awsbatchcomputeenvironmentv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbatchcomputeenvironment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the AWS Batch compute environment and exports its
// outputs. Job queues and scheduling policies are separate resources
// (AwsBatchJobQueue / AwsBatchSchedulingPolicy) that compose onto the
// environment through its exported ARN.
func Resources(ctx *pulumi.Context, iacInput *awsbatchcomputeenvironmentv1alpha1.AwsBatchComputeEnvironmentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsBatchComputeEnvironment.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdCe, err := computeEnvironment(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "compute environment")
	}

	// --- Exports ---
	ctx.Export(OpComputeEnvironmentArn, createdCe.Arn)
	ctx.Export(OpComputeEnvironmentName, createdCe.Name)
	ctx.Export(OpEcsClusterArn, createdCe.EcsClusterArn)
	ctx.Export(OpStatus, createdCe.Status)

	return nil
}

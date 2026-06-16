package module

import (
	"fmt"

	"github.com/pkg/errors"
	awsbatchcomputeenvironmentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsbatchcomputeenvironment/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources orchestrates creation of AWS Batch resources and exports outputs.
func Resources(ctx *pulumi.Context, stackInput *awsbatchcomputeenvironmentv1.AwsBatchComputeEnvironmentStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.AwsBatchComputeEnvironment.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	// 1. Scheduling policy (optional, created first because job queues reference it)
	createdSchedulingPolicy, err := schedulingPolicy(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "scheduling policy")
	}

	// 2. Compute environment (the primary infrastructure resource)
	createdCe, err := computeEnvironment(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "compute environment")
	}

	// 3. Job queues (reference compute environment and optional scheduling policy)
	spec := locals.AwsBatchComputeEnvironment.Spec
	for i, jq := range spec.JobQueues {
		err := jobQueue(ctx, locals, provider, createdCe, createdSchedulingPolicy, jq, i)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("job queue %s", jq.Name))
		}
	}

	// Export outputs
	ctx.Export(OpComputeEnvironmentArn, createdCe.Arn)
	ctx.Export(OpComputeEnvironmentName, createdCe.Name)
	ctx.Export(OpEcsClusterArn, createdCe.EcsClusterArn)
	ctx.Export(OpStatus, createdCe.Status)

	if createdSchedulingPolicy != nil {
		ctx.Export(OpSchedulingPolicyArn, createdSchedulingPolicy.Arn)
	}

	return nil
}

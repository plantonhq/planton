package module

import (
	"fmt"

	"github.com/pkg/errors"
	awsbatchcomputeenvironmentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsbatchcomputeenvironment/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/batch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func jobQueue(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	ce *batch.ComputeEnvironment,
	sp *batch.SchedulingPolicy,
	jq *awsbatchcomputeenvironmentv1.AwsBatchJobQueue,
	index int,
) error {
	args := &batch.JobQueueArgs{
		Name:     pulumi.String(jq.Name),
		State:    pulumi.String(jq.GetState()),
		Priority: pulumi.Int(jq.Priority),
		Tags:     pulumi.ToStringMap(locals.Labels),
		ComputeEnvironmentOrders: batch.JobQueueComputeEnvironmentOrderArray{
			&batch.JobQueueComputeEnvironmentOrderArgs{
				ComputeEnvironment: ce.Arn,
				Order:              pulumi.Int(1),
			},
		},
	}

	// Attach scheduling policy if one was created
	if sp != nil {
		args.SchedulingPolicyArn = sp.Arn
	}

	// Job state time limit actions
	if len(jq.JobStateTimeLimitActions) > 0 {
		var actions batch.JobQueueJobStateTimeLimitActionArray
		for _, action := range jq.JobStateTimeLimitActions {
			actions = append(actions, &batch.JobQueueJobStateTimeLimitActionArgs{
				Action:         pulumi.String(action.Action),
				MaxTimeSeconds: pulumi.Int(action.MaxTimeSeconds),
				Reason:         pulumi.String(action.Reason),
				State:          pulumi.String(action.State),
			})
		}
		args.JobStateTimeLimitActions = actions
	}

	resourceName := fmt.Sprintf("batch-job-queue-%d", index)

	createdJq, err := batch.NewJobQueue(ctx, resourceName, args,
		pulumi.Provider(provider),
		pulumi.DependsOn([]pulumi.Resource{ce}),
	)
	if err != nil {
		return errors.Wrapf(err, "create job queue %s", jq.Name)
	}

	// Export individual queue ARN as job_queue_arns.<name>
	outputKey := fmt.Sprintf("job_queue_arns.%s", jq.Name)
	ctx.Export(outputKey, createdJq.Arn)

	return nil
}

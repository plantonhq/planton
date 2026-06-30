package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/batch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// schedulingPolicy creates a fair-share scheduling policy when configured in
// the spec. Returns nil if no scheduling policy is defined.
func schedulingPolicy(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
) (*batch.SchedulingPolicy, error) {
	spec := locals.AwsBatchComputeEnvironment.Spec
	if spec.SchedulingPolicy == nil {
		return nil, nil
	}

	sp := spec.SchedulingPolicy

	fairSharePolicy := &batch.SchedulingPolicyFairSharePolicyArgs{}

	if sp.ComputeReservation != nil {
		fairSharePolicy.ComputeReservation = pulumi.IntPtr(int(sp.GetComputeReservation()))
	}
	if sp.ShareDecaySeconds != nil {
		fairSharePolicy.ShareDecaySeconds = pulumi.IntPtr(int(sp.GetShareDecaySeconds()))
	}

	if len(sp.ShareDistributions) > 0 {
		var distributions batch.SchedulingPolicyFairSharePolicyShareDistributionArray
		for _, sd := range sp.ShareDistributions {
			dist := &batch.SchedulingPolicyFairSharePolicyShareDistributionArgs{
				ShareIdentifier: pulumi.String(sd.ShareIdentifier),
			}
			if sd.WeightFactor != 0 {
				dist.WeightFactor = pulumi.Float64Ptr(sd.WeightFactor)
			}
			distributions = append(distributions, dist)
		}
		fairSharePolicy.ShareDistributions = distributions
	}

	policyName := locals.AwsBatchComputeEnvironment.Metadata.Id + "-scheduling-policy"

	created, err := batch.NewSchedulingPolicy(ctx, "batch-scheduling-policy", &batch.SchedulingPolicyArgs{
		Name:            pulumi.StringPtr(policyName),
		FairSharePolicy: fairSharePolicy,
		Tags:            pulumi.ToStringMap(locals.Labels),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "create batch scheduling policy")
	}

	return created, nil
}

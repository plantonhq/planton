package module

import (
	"github.com/pkg/errors"
	awseksfargateprofilev1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksfargateprofile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the Fargate profile. The cluster attaches by
// reference, the pod execution role is a referenced AwsIamRole that
// carries its own policies (this module never modifies a role it merely
// references), and the subnets are referenced AwsSubnet nodes -- private
// only, which AWS enforces at create time.
//
// The entire profile is create-time immutable in AWS, and AWS serializes
// profile operations per cluster (one create or delete at a time) --
// both engines simply wait; no ordering knobs are needed here.
func Resources(ctx *pulumi.Context, iacInput *awseksfargateprofilev1alpha1.AwsEksFargateProfileIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEksFargateProfile.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	spec := locals.AwsEksFargateProfile.Spec

	subnetIds := make(pulumi.StringArray, 0, len(spec.SubnetIds))
	for _, subnet := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(subnet.GetValue()))
	}

	// A pod runs on Fargate when it matches ANY selector; within one
	// selector the namespace AND every label must match.
	selectors := make(eks.FargateProfileSelectorArray, 0, len(spec.Selectors))
	for _, selector := range spec.Selectors {
		selectorArgs := &eks.FargateProfileSelectorArgs{
			Namespace: pulumi.String(selector.Namespace),
		}
		if len(selector.Labels) > 0 {
			selectorArgs.Labels = pulumi.ToStringMap(selector.Labels)
		}
		selectors = append(selectors, selectorArgs)
	}

	created, err := eks.NewFargateProfile(ctx, locals.FargateProfileName, &eks.FargateProfileArgs{
		FargateProfileName:  pulumi.StringPtr(locals.FargateProfileName),
		ClusterName:         pulumi.String(spec.ClusterName.GetValue()),
		PodExecutionRoleArn: pulumi.String(spec.PodExecutionRoleArn.GetValue()),
		SubnetIds:           subnetIds,
		Selectors:           selectors,
		Tags:                pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create EKS Fargate profile")
	}

	ctx.Export(OpFargateProfileArn, created.Arn)
	ctx.Export(OpFargateProfileName, created.FargateProfileName)
	ctx.Export(OpStatus, created.Status)

	return nil
}

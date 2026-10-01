package module

import (
	"github.com/pkg/errors"
	gcpgcsbucketiammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgcsbucketiammember/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// iamMember applies the single ADDITIVE grant ON the bucket: one role, to
// one member, on one bucket. Additive means the provider merges this (role,
// member) pair into the bucket's IAM policy without touching any other
// member's bindings on the same role, and destroy subtracts only this pair.
//
// Every argument is immutable (ForceNew): an IAM grant has no update -- any
// change replaces it, which is also how the API behaves.
func iamMember(ctx *pulumi.Context, target *gcpgcsbucketiammemberv1alpha1.GcpGcsBucketIamMember, gcpProvider *gcp.Provider) error {
	spec := target.Spec

	args := &storage.BucketIAMMemberArgs{
		Bucket: pulumi.String(spec.Bucket.GetValue()),
		Role:   pulumi.String(spec.Role.GetValue()),
		Member: pulumi.String(spec.Member.GetValue()),
	}

	// An IAM Condition is part of the grant's identity: the same role granted
	// with and without a condition are two independent bindings in the policy.
	if spec.Condition != nil {
		conditionArgs := &storage.BucketIAMMemberConditionArgs{
			Title:      pulumi.String(spec.Condition.Title),
			Expression: pulumi.String(spec.Condition.Expression),
		}
		if spec.Condition.Description != "" {
			conditionArgs.Description = pulumi.String(spec.Condition.Description)
		}
		args.Condition = conditionArgs
	}

	createdMember, err := storage.NewBucketIAMMember(ctx, "iam-member", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create gcs bucket IAM member")
	}

	ctx.Export(OpBucket, createdMember.Bucket)
	ctx.Export(OpRole, createdMember.Role)
	ctx.Export(OpMember, createdMember.Member)
	ctx.Export(OpEtag, createdMember.Etag)

	return nil
}

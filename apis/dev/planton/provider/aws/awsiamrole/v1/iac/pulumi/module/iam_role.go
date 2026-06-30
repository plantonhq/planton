package module

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func iamRole(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	roleName := locals.AwsIamRole.Metadata.Name
	spec := locals.AwsIamRole.Spec

	trustPolicyString, err := structToJSONString(spec.TrustPolicy)
	if err != nil {
		return errors.Wrap(err, "failed to marshal trust policy JSON")
	}

	// Create the core IAM role
	iamRole, err := iam.NewRole(ctx, roleName, &iam.RoleArgs{
		Name:             pulumi.String(roleName),
		AssumeRolePolicy: pulumi.String(trustPolicyString),
		Description:      pulumi.String(spec.Description),
		Path:             pulumi.String(spec.Path),
		Tags:             pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create IAM role")
	}

	// Attach managed policy ARNs
	for idx, policyArn := range spec.ManagedPolicyArns {
		attachName := fmt.Sprintf("%s-attach-%d", roleName, idx)
		_, err := iam.NewRolePolicyAttachment(ctx, attachName, &iam.RolePolicyAttachmentArgs{
			Role:      iamRole.Name,
			PolicyArn: pulumi.String(policyArn),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to attach policy ARN %s", policyArn)
		}
	}

	// Inline policies
	for policyName, inlineStruct := range spec.InlinePolicies {
		inlinePolicyString, err := structToJSONString(inlineStruct)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal inline policy for %s", policyName)
		}

		inlineName := fmt.Sprintf("%s-inline-%s", roleName, policyName)
		_, err = iam.NewRolePolicy(ctx, inlineName, &iam.RolePolicyArgs{
			Role:   iamRole.Name,
			Policy: pulumi.String(inlinePolicyString),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to create inline policy %s", policyName)
		}
	}

	// Always create an instance profile that wraps this role. Instance profiles
	// are free and idempotent in AWS, and EC2 requires one (not a bare role) to
	// assume a role. Exporting its ARN lets an EC2 instance reference the role via
	// iam_instance_profile_arn -> AwsIamRole/status.outputs.instance_profile_arn.
	instanceProfile, err := iam.NewInstanceProfile(ctx, roleName, &iam.InstanceProfileArgs{
		Name: pulumi.String(roleName),
		Role: iamRole.Name,
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create IAM instance profile")
	}

	// Export final outputs
	ctx.Export(OpRoleArn, iamRole.Arn)
	ctx.Export(OpRoleName, iamRole.Name)
	ctx.Export(OpInstanceProfileArn, instanceProfile.Arn)
	ctx.Export(OpInstanceProfileName, instanceProfile.Name)

	return nil
}

// structToJSONString converts a google.protobuf.Struct to a raw JSON string.
func structToJSONString(s *structpb.Struct) (string, error) {
	if s == nil {
		return "{}", nil
	}
	m := s.AsMap()
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

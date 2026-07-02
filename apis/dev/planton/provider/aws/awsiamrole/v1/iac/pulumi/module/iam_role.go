package module

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/valuefrom"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

// iamRole provisions the role and its policy wiring. An IAM role is an
// assumable identity: the trust policy controls WHO can assume it, the
// attached/inline policies control WHAT it can do once assumed, and an
// optional permissions boundary caps the maximum it can ever do. Name and
// path are create-only (changing them replaces the role); everything else
// updates in place.
func iamRole(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) error {
	roleName := locals.AwsIamRole.Metadata.Name
	spec := locals.AwsIamRole.Spec

	// trust_policy is a free-form JSON object (google.protobuf.Struct);
	// iam.Role wants assume_role_policy as a JSON string, so encode it here.
	trustPolicyString, err := structToJSONString(spec.TrustPolicy)
	if err != nil {
		return errors.Wrap(err, "failed to marshal trust policy JSON")
	}

	roleArgs := &iam.RoleArgs{
		Name:             pulumi.String(roleName),
		AssumeRolePolicy: pulumi.String(trustPolicyString),
		Tags:             pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.Description != "" {
		roleArgs.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Path != "" {
		roleArgs.Path = pulumi.StringPtr(spec.Path)
	}
	// 0 means "unset" (proto3 zero value); AWS then applies its 3600s default.
	if spec.MaxSessionDuration != 0 {
		roleArgs.MaxSessionDuration = pulumi.IntPtr(int(spec.MaxSessionDuration))
	}
	// The boundary is a ceiling, not a grant: effective permissions are the
	// intersection of this policy and the role's permission policies. A
	// valueFrom reference is resolved to the AwsIamPolicy's policy_arn before
	// the module runs.
	if spec.PermissionsBoundary.GetValue() != "" {
		roleArgs.PermissionsBoundary = pulumi.StringPtr(spec.PermissionsBoundary.GetValue())
	}
	// When enabled, deletion force-detaches policies still attached to the
	// role (including attachments made outside this resource) instead of
	// failing.
	if spec.ForceDetachPolicies {
		roleArgs.ForceDetachPolicies = pulumi.BoolPtr(true)
	}

	createdRole, err := iam.NewRole(ctx, roleName, roleArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create IAM role")
	}

	// Each managed-policy attachment is its own resource (not the deprecated
	// exclusive managed_policy_arns role argument) so attachments reconcile
	// individually: adding or removing an entry attaches or detaches just that
	// policy, and attachments made outside this resource are left alone.
	// valueFrom references were resolved to policy ARNs before the module ran.
	for idx, policyArn := range valuefrom.ToStringArray(spec.ManagedPolicyArns) {
		attachName := fmt.Sprintf("%s-attach-%d", roleName, idx)
		_, err := iam.NewRolePolicyAttachment(ctx, attachName, &iam.RolePolicyAttachmentArgs{
			Role:      createdRole.Name,
			PolicyArn: pulumi.String(policyArn),
		}, pulumi.Provider(provider), pulumi.Parent(createdRole))
		if err != nil {
			return errors.Wrapf(err, "failed to attach policy ARN %s", policyArn)
		}
	}

	// Inline policies live and die with the role -- permissions unique to this
	// role that would be noise as standalone AwsIamPolicy resources.
	for policyName, inlineStruct := range spec.InlinePolicies {
		inlinePolicyString, err := structToJSONString(inlineStruct)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal inline policy for %s", policyName)
		}

		inlineName := fmt.Sprintf("%s-inline-%s", roleName, policyName)
		_, err = iam.NewRolePolicy(ctx, inlineName, &iam.RolePolicyArgs{
			Name:   pulumi.String(policyName),
			Role:   createdRole.Name,
			Policy: pulumi.String(inlinePolicyString),
		}, pulumi.Provider(provider), pulumi.Parent(createdRole))
		if err != nil {
			return errors.Wrapf(err, "failed to create inline policy %s", policyName)
		}
	}

	ctx.Export(OpRoleArn, createdRole.Arn)
	ctx.Export(OpRoleName, createdRole.Name)
	ctx.Export(OpRoleId, createdRole.UniqueId)

	return nil
}

// structToJSONString converts a google.protobuf.Struct to a raw JSON string.
func structToJSONString(s *structpb.Struct) (string, error) {
	if s == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(s.AsMap())
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

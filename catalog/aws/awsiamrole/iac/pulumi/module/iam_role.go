package module

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	iamrolev1 "github.com/plantonhq/planton/catalog/aws/awsiamrole/v1alpha1"
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

	// The trust oneof: exactly one of trust_policy (free-form JSON) or
	// oidc_trust (typed federated trust) is set — spec-enforced. iam.Role
	// wants assume_role_policy as a JSON string either way: the oidc arm
	// composes the web-identity document from the provider's outputs
	// (references are resolved before the module runs), the free-form arm
	// encodes the user's Struct as-is.
	var trustPolicyString string
	var err error
	if oidcTrust := spec.GetOidcTrust(); oidcTrust != nil {
		trustPolicyString, err = composeOidcTrustPolicy(oidcTrust)
		if err != nil {
			return errors.Wrap(err, "failed to compose oidc trust policy JSON")
		}
	} else {
		trustPolicyString, err = structToJSONString(spec.GetTrustPolicy())
		if err != nil {
			return errors.Wrap(err, "failed to marshal trust policy JSON")
		}
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
	// failing. Always sent -- explicit false included -- so both engines pin
	// the value in state identically (the Terraform module renders it
	// unconditionally).
	roleArgs.ForceDetachPolicies = pulumi.BoolPtr(spec.ForceDetachPolicies)

	createdRole, err := iam.NewRole(ctx, roleName, roleArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create IAM role")
	}

	// Each managed-policy attachment is its own resource (not the deprecated
	// exclusive managed_policy_arns role argument) so attachments reconcile
	// individually: adding or removing an entry attaches or detaches just that
	// policy, and attachments made outside this resource are left alone.
	// valueFrom references were resolved to policy ARNs before the module ran.
	// Attachments are keyed by the policy ARN itself (sanitized for the
	// logical name), never the list index -- mirrors the Terraform module's
	// for_each = toset(...): reordering managed_policy_arns must be a no-op,
	// not a transient detach/re-attach on a live role.
	for _, policyArn := range valuefrom.ToStringArray(spec.ManagedPolicyArns) {
		attachName := fmt.Sprintf("%s-attach-%s", roleName, sanitizeForLogicalName(policyArn))
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

// sanitizeForLogicalName reduces an ARN to a stable Pulumi logical-name
// segment: every character outside [A-Za-z0-9] becomes '-'. The full ARN
// stays in the name so distinct policies can never collide (bare policy
// names repeat across paths and accounts).
func sanitizeForLogicalName(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, s)
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

// composeOidcTrustPolicy builds the sts:AssumeRoleWithWebIdentity trust
// document from the typed oidc_trust arm, byte-equivalent in structure to the
// Terraform module's composition. One statement PER CONDITION OPERATOR: IAM
// ANDs condition operators inside a statement and ORs across statements, so
// exact subjects (StringEquals) and wildcard subjects (StringLike) ride
// separate statements — mixing both on the `sub` key in one statement would
// require a token to satisfy both at once.
func composeOidcTrustPolicy(oidcTrust *iamrolev1.AwsIamRoleOidcTrust) (string, error) {
	providerArn := oidcTrust.GetProviderArn().GetValue()
	// IAM's condition keys are the bare issuer (<host>/<path>:sub). A provider
	// URL handed over with its scheme (typed by hand, or read from a provider
	// built before its output dropped the scheme) is stripped, or the condition
	// would never match and every sign-in would be refused.
	providerUrl := issuerWithoutScheme(oidcTrust.GetProviderUrl().GetValue())

	// The audience condition defaults to sts.amazonaws.com — the audience EKS
	// IRSA and GitHub Actions both present — so the common case needs no
	// explicit audiences.
	audiences := oidcTrust.GetAudiences()
	if len(audiences) == 0 {
		audiences = []string{"sts.amazonaws.com"}
	}

	statements := make([]interface{}, 0, 2)
	if subjects := oidcTrust.GetSubjects(); len(subjects) > 0 {
		statements = append(statements, map[string]interface{}{
			"Effect":    "Allow",
			"Principal": map[string]interface{}{"Federated": providerArn},
			"Action":    "sts:AssumeRoleWithWebIdentity",
			"Condition": map[string]interface{}{
				"StringEquals": map[string]interface{}{
					providerUrl + ":sub": subjects,
					providerUrl + ":aud": audiences,
				},
			},
		})
	}
	if wildcardSubjects := oidcTrust.GetWildcardSubjects(); len(wildcardSubjects) > 0 {
		statements = append(statements, map[string]interface{}{
			"Effect":    "Allow",
			"Principal": map[string]interface{}{"Federated": providerArn},
			"Action":    "sts:AssumeRoleWithWebIdentity",
			"Condition": map[string]interface{}{
				// The audience stays an exact match even on the
				// wildcard-subject statement — only the subject pattern is
				// fuzzy.
				"StringEquals": map[string]interface{}{
					providerUrl + ":aud": audiences,
				},
				"StringLike": map[string]interface{}{
					providerUrl + ":sub": wildcardSubjects,
				},
			},
		})
	}

	bytes, err := json.Marshal(map[string]interface{}{
		"Version":   "2012-10-17",
		"Statement": statements,
	})
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// issuerWithoutScheme is an issuer URL as IAM's condition keys name it: the
// host and path, with no https:// (or http://) in front.
func issuerWithoutScheme(url string) string {
	return strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
}

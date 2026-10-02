package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/iam"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// denyPolicy provisions the IAM deny policy.
//
// GCP's API identifies the attach point by its URL-ENCODED full resource
// name (e.g. cloudresourcemanager.googleapis.com%2Fprojects%2Fmy-project).
// The module renders that string from the spec's typed parent message so
// manifests never hand-assemble it — the mis-encoding class this kind
// exists to remove. The only character needing encoding in a full resource
// name is "/", identically to Terraform's urlencode(), which keeps the two
// engines byte-compatible.
//
// No API-enablement resource here: deny policies attach to projects,
// folders, or organizations through the always-on IAM v2 surface, and
// creating them requires org-level permissions in any case.
func denyPolicy(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpIamDenyPolicy.Spec

	parent, err := resolveParent(ctx, locals, gcpProvider)
	if err != nil {
		return err
	}

	rules := iam.DenyPolicyRuleArray{}
	for _, rule := range spec.Rules {
		denyRuleArgs := &iam.DenyPolicyRuleDenyRuleArgs{}
		if len(rule.DenyRule.DeniedPrincipals) > 0 {
			denyRuleArgs.DeniedPrincipals = pulumi.ToStringArray(rule.DenyRule.DeniedPrincipals)
		}
		if len(rule.DenyRule.ExceptionPrincipals) > 0 {
			denyRuleArgs.ExceptionPrincipals = pulumi.ToStringArray(rule.DenyRule.ExceptionPrincipals)
		}
		if len(rule.DenyRule.DeniedPermissions) > 0 {
			denyRuleArgs.DeniedPermissions = pulumi.ToStringArray(rule.DenyRule.DeniedPermissions)
		}
		if len(rule.DenyRule.ExceptionPermissions) > 0 {
			denyRuleArgs.ExceptionPermissions = pulumi.ToStringArray(rule.DenyRule.ExceptionPermissions)
		}
		if rule.DenyRule.DenialCondition != nil {
			conditionArgs := &iam.DenyPolicyRuleDenyRuleDenialConditionArgs{
				Expression: pulumi.String(rule.DenyRule.DenialCondition.Expression),
			}
			if rule.DenyRule.DenialCondition.Title != "" {
				conditionArgs.Title = pulumi.StringPtr(rule.DenyRule.DenialCondition.Title)
			}
			if rule.DenyRule.DenialCondition.Description != "" {
				conditionArgs.Description = pulumi.StringPtr(rule.DenyRule.DenialCondition.Description)
			}
			if rule.DenyRule.DenialCondition.Location != "" {
				conditionArgs.Location = pulumi.StringPtr(rule.DenyRule.DenialCondition.Location)
			}
			denyRuleArgs.DenialCondition = conditionArgs
		}
		ruleArgs := &iam.DenyPolicyRuleArgs{
			DenyRule: denyRuleArgs,
		}
		if rule.Description != "" {
			ruleArgs.Description = pulumi.StringPtr(rule.Description)
		}
		rules = append(rules, ruleArgs)
	}

	args := &iam.DenyPolicyArgs{
		Name:   pulumi.String(locals.PolicyName),
		Parent: pulumi.String(parent),
		Rules:  rules,
	}
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdPolicy, err := iam.NewDenyPolicy(ctx, "deny-policy", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create deny policy")
	}

	ctx.Export(OpPolicyName, createdPolicy.ID().ToStringOutput())
	ctx.Export(OpEtag, createdPolicy.Etag)

	return nil
}

// resolveParent renders the URL-encoded full resource name of the attach
// point. An empty parent message falls back to the provider's default
// project, read from the provider's own resolved configuration — gated to
// that one case so every plan that names its attach point runs
// credential-free (the same grain as the project-IAM-member kind).
func resolveParent(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) (string, error) {
	parent := locals.GcpIamDenyPolicy.Spec.Parent

	if parent != nil && parent.FolderId.GetValue() != "" {
		return encodeParent("cloudresourcemanager.googleapis.com/folders/" +
			strings.TrimPrefix(parent.FolderId.GetValue(), "folders/")), nil
	}
	if parent != nil && parent.OrganizationId != "" {
		return encodeParent("cloudresourcemanager.googleapis.com/organizations/" +
			strings.TrimPrefix(parent.OrganizationId, "organizations/")), nil
	}

	projectId := ""
	if parent != nil {
		projectId = parent.ProjectId.GetValue()
	}
	if projectId == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return "", errors.Wrap(err, "failed to read provider client config for the default project")
		}
		if clientConfig.Project == "" {
			return "", errors.New("parent is empty and the provider has no default project configured")
		}
		projectId = clientConfig.Project
	}
	return encodeParent("cloudresourcemanager.googleapis.com/projects/" + projectId), nil
}

// encodeParent URL-encodes a full resource name. "/" is the only character
// in a full resource name that needs encoding — replacing it directly is
// byte-identical to Terraform's urlencode() on this charset, which keeps
// the two engines' rendered parents comparable.
func encodeParent(fullResourceName string) string {
	return strings.ReplaceAll(fullResourceName, "/", "%2F")
}

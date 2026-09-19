package module

import (
	"github.com/pkg/errors"
	gcporgpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcporgpolicy/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/orgpolicy"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// orgPolicy provisions the organization policy.
//
// The policy's identity is its name, `{parent}/policies/{constraint}`,
// assembled here from the rendered scope and the constraint; both are
// immutable, so a change to either recreates the policy. An empty scope
// resolves the provider's default project once through the provider's
// client config (the same fallback the Terraform module expresses with
// data.google_project), so a manifest that names no scope governs the
// project it is deployed into.
//
// PARITY: Google's API models a rule's verdict as booleans in a one-of,
// and the provider flattens that into the tri-state strings "TRUE" /
// "FALSE" / unset. The spec keeps the API's shape (a oneof of bools) and
// this module renders the string form for exactly the arm that is set --
// the Terraform module does the same in locals.tf -- so `enforce: false`
// reaches Google as "FALSE" (a real rule) and an unset arm is never sent.
func orgPolicy(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpOrgPolicy.Spec

	parent := locals.Parent
	if parent == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to resolve the provider's default project for the policy scope")
		}
		if clientConfig.Project == "" {
			return errors.New("the policy names no scope and the provider has no default project -- set spec.scope or configure a project")
		}
		parent = "projects/" + clientConfig.Project
	}

	args := &orgpolicy.PolicyArgs{
		Name:   pulumi.String(parent + "/policies/" + locals.Constraint),
		Parent: pulumi.String(parent),
	}
	if spec.Policy != nil {
		args.Spec = policySpecArgs(spec.Policy)
	}
	if spec.DryRunPolicy != nil {
		args.DryRunSpec = dryRunSpecArgs(spec.DryRunPolicy)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdPolicy, err := orgpolicy.NewPolicy(ctx, locals.GcpOrgPolicy.Metadata.Name, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create organization policy")
	}

	ctx.Export(OpName, createdPolicy.Name)
	ctx.Export(OpEtag, createdPolicy.Etag)

	return nil
}

// tristate renders a set boolean arm as the provider's string form.
func tristate(v bool) pulumi.StringPtrInput {
	if v {
		return pulumi.StringPtr("TRUE")
	}
	return pulumi.StringPtr("FALSE")
}

// policySpecArgs maps the enforced rule set onto the provider's `spec`
// block. Optional scalars are sent only when set so the provider's
// defaults stay the provider's.
func policySpecArgs(set *gcporgpolicyv1alpha1.GcpOrgPolicyRuleSet) *orgpolicy.PolicySpecArgs {
	out := &orgpolicy.PolicySpecArgs{}
	if set.InheritFromParent {
		out.InheritFromParent = pulumi.BoolPtr(true)
	}
	if set.Reset_ {
		out.Reset = pulumi.BoolPtr(true)
	}
	if len(set.Rules) > 0 {
		rules := orgpolicy.PolicySpecRuleArray{}
		for _, rule := range set.Rules {
			r := &orgpolicy.PolicySpecRuleArgs{}
			switch kind := rule.Kind.(type) {
			case *gcporgpolicyv1alpha1.GcpOrgPolicyRule_AllowAll:
				r.AllowAll = tristate(kind.AllowAll)
			case *gcporgpolicyv1alpha1.GcpOrgPolicyRule_DenyAll:
				r.DenyAll = tristate(kind.DenyAll)
			case *gcporgpolicyv1alpha1.GcpOrgPolicyRule_Enforce:
				r.Enforce = tristate(kind.Enforce)
			case *gcporgpolicyv1alpha1.GcpOrgPolicyRule_Values:
				values := &orgpolicy.PolicySpecRuleValuesArgs{}
				if len(kind.Values.AllowedValues) > 0 {
					values.AllowedValues = pulumi.ToStringArray(kind.Values.AllowedValues)
				}
				if len(kind.Values.DeniedValues) > 0 {
					values.DeniedValues = pulumi.ToStringArray(kind.Values.DeniedValues)
				}
				r.Values = values
			}
			if rule.Condition != nil {
				r.Condition = &orgpolicy.PolicySpecRuleConditionArgs{
					Expression:  pulumi.String(rule.Condition.Expression),
					Title:       optionalString(rule.Condition.Title),
					Description: optionalString(rule.Condition.Description),
					Location:    optionalString(rule.Condition.Location),
				}
			}
			if rule.Parameters != "" {
				r.Parameters = pulumi.StringPtr(rule.Parameters)
			}
			rules = append(rules, r)
		}
		out.Rules = rules
	}
	return out
}

// dryRunSpecArgs is policySpecArgs for the provider's `dry_run_spec` block,
// which the SDK types separately although the shape is identical.
func dryRunSpecArgs(set *gcporgpolicyv1alpha1.GcpOrgPolicyRuleSet) *orgpolicy.PolicyDryRunSpecArgs {
	out := &orgpolicy.PolicyDryRunSpecArgs{}
	if set.InheritFromParent {
		out.InheritFromParent = pulumi.BoolPtr(true)
	}
	if set.Reset_ {
		out.Reset = pulumi.BoolPtr(true)
	}
	if len(set.Rules) > 0 {
		rules := orgpolicy.PolicyDryRunSpecRuleArray{}
		for _, rule := range set.Rules {
			r := &orgpolicy.PolicyDryRunSpecRuleArgs{}
			switch kind := rule.Kind.(type) {
			case *gcporgpolicyv1alpha1.GcpOrgPolicyRule_AllowAll:
				r.AllowAll = tristate(kind.AllowAll)
			case *gcporgpolicyv1alpha1.GcpOrgPolicyRule_DenyAll:
				r.DenyAll = tristate(kind.DenyAll)
			case *gcporgpolicyv1alpha1.GcpOrgPolicyRule_Enforce:
				r.Enforce = tristate(kind.Enforce)
			case *gcporgpolicyv1alpha1.GcpOrgPolicyRule_Values:
				values := &orgpolicy.PolicyDryRunSpecRuleValuesArgs{}
				if len(kind.Values.AllowedValues) > 0 {
					values.AllowedValues = pulumi.ToStringArray(kind.Values.AllowedValues)
				}
				if len(kind.Values.DeniedValues) > 0 {
					values.DeniedValues = pulumi.ToStringArray(kind.Values.DeniedValues)
				}
				r.Values = values
			}
			if rule.Condition != nil {
				r.Condition = &orgpolicy.PolicyDryRunSpecRuleConditionArgs{
					Expression:  pulumi.String(rule.Condition.Expression),
					Title:       optionalString(rule.Condition.Title),
					Description: optionalString(rule.Condition.Description),
					Location:    optionalString(rule.Condition.Location),
				}
			}
			if rule.Parameters != "" {
				r.Parameters = pulumi.StringPtr(rule.Parameters)
			}
			rules = append(rules, r)
		}
		out.Rules = rules
	}
	return out
}

// optionalString sends a string only when the spec set it.
func optionalString(v string) pulumi.StringPtrInput {
	if v == "" {
		return nil
	}
	return pulumi.StringPtr(v)
}

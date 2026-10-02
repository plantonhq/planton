package gcporgpolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpOrgPolicySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: name}},
	}
}

var _ = ginkgo.Describe("GcpOrgPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// A boolean constraint enforced on the provider's default project: the
	// smallest policy that does something.
	minimal := func() *GcpOrgPolicy {
		return &GcpOrgPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpOrgPolicy",
			Metadata: &shared.CloudResourceMetadata{
				Name: "disable-serial-port",
			},
			Spec: &GcpOrgPolicySpec{
				Constraint: "compute.disableSerialPortAccess",
				Policy: &GcpOrgPolicyRuleSet{
					Rules: []*GcpOrgPolicyRule{
						{Kind: &GcpOrgPolicyRule_Enforce{Enforce: true}},
					},
				},
			},
		}
	}

	expectError := func(target *GcpOrgPolicy, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a boolean constraint enforced on the default project", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept each scope arm alone: project, folder, organization", func() {
		scopes := []*GcpOrgPolicyScope{
			{ProjectId: litRef("my-gcp-project-123")},
			{ProjectId: nameRef("landing-zone-project")},
			{FolderId: litRef("987654321098")},
			{FolderId: nameRef("environments")},
			{OrganizationId: "123456789012"},
			{},
		}
		for _, scope := range scopes {
			target := minimal()
			target.Spec.Scope = scope
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept a list constraint with values, allow_all, or deny_all, and a managed constraint with parameters", func() {
		target := minimal()
		target.Spec.Constraint = "gcp.resourceLocations"
		target.Spec.Policy = &GcpOrgPolicyRuleSet{
			InheritFromParent: true,
			Rules: []*GcpOrgPolicyRule{
				{Kind: &GcpOrgPolicyRule_Values{Values: &GcpOrgPolicyRuleValues{AllowedValues: []string{"in:us-locations"}, DeniedValues: []string{"us-west4"}}}},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.Policy.Rules = []*GcpOrgPolicyRule{{Kind: &GcpOrgPolicyRule_AllowAll{AllowAll: true}}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.Policy.Rules = []*GcpOrgPolicyRule{{Kind: &GcpOrgPolicyRule_DenyAll{DenyAll: true}}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.Constraint = "iam.managed.allowedPolicyMembers"
		target.Spec.Policy.Rules = []*GcpOrgPolicyRule{{
			Kind:       &GcpOrgPolicyRule_Enforce{Enforce: true},
			Parameters: `{"allowedMemberSubjects": ["principalSet://iam.googleapis.com/..."]}`,
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept enforce: false as a real rule (the child-scope relaxation)", func() {
		target := minimal()
		target.Spec.Policy.Rules = []*GcpOrgPolicyRule{{Kind: &GcpOrgPolicyRule_Enforce{Enforce: false}}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a conditional rule beside an unconditional one, mirrored in a dry-run policy", func() {
		target := minimal()
		rules := []*GcpOrgPolicyRule{
			{Kind: &GcpOrgPolicyRule_Enforce{Enforce: true}},
			{
				Kind: &GcpOrgPolicyRule_Enforce{Enforce: false},
				Condition: &GcpOrgPolicyRuleCondition{
					Expression:  "resource.matchTag('123456789012/environment', 'sandbox')",
					Title:       "sandbox exemption",
					Description: "sandbox projects may use the serial console",
				},
			},
		}
		target.Spec.Policy = &GcpOrgPolicyRuleSet{Rules: rules}
		target.Spec.DryRunPolicy = &GcpOrgPolicyRuleSet{Rules: rules}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a reset policy and a dry-run-only policy", func() {
		target := minimal()
		target.Spec.Policy = &GcpOrgPolicyRuleSet{Reset_: true}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target = minimal()
		target.Spec.DryRunPolicy = target.Spec.Policy
		target.Spec.Policy = nil
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a custom constraint by reference or as the custom.<name> literal", func() {
		target := minimal()
		target.Spec.Constraint = ""
		target.Spec.CustomConstraint = nameRef("deny-public-gke-nodes")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.CustomConstraint = litRef("custom.denyPublicGkeNodes")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every deletion_policy value and the empty default", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a policy naming both a constraint and a custom constraint, or neither", func() {
		target := minimal()
		target.Spec.CustomConstraint = litRef("custom.denyPublicGkeNodes")
		expectError(target, "exactly one of constraint")

		target.Spec.Constraint = ""
		target.Spec.CustomConstraint = nil
		expectError(target, "exactly one of constraint")
	})

	ginkgo.It("should reject a constraint that is not a dotted name", func() {
		for _, c := range []string{"disableSerialPortAccess", "constraints/compute.disableSerialPortAccess", "compute."} {
			target := minimal()
			target.Spec.Constraint = c
			expectError(target, "dotted constraint name")
		}
	})

	ginkgo.It("should reject two scope arms", func() {
		target := minimal()
		target.Spec.Scope = &GcpOrgPolicyScope{ProjectId: litRef("my-gcp-project-123"), OrganizationId: "123456789012"}
		expectError(target, "at most one of project_id, folder_id, or organization_id")
	})

	ginkgo.It("should reject a non-numeric organization ID", func() {
		target := minimal()
		target.Spec.Scope = &GcpOrgPolicyScope{OrganizationId: "organizations/123456789012"}
		expectError(target, "numeric organization ID")
	})

	ginkgo.It("should reject a rule without a verdict", func() {
		target := minimal()
		target.Spec.Policy.Rules = []*GcpOrgPolicyRule{{Condition: &GcpOrgPolicyRuleCondition{Expression: "resource.matchTag('1/a', 'b')"}}}
		expectError(target, "kind")
	})

	ginkgo.It("should reject a reset policy that also carries rules or inherit_from_parent", func() {
		target := minimal()
		target.Spec.Policy = &GcpOrgPolicyRuleSet{Reset_: true, Rules: []*GcpOrgPolicyRule{{Kind: &GcpOrgPolicyRule_Enforce{Enforce: true}}}}
		expectError(target, "reset: true must carry no rules")

		target.Spec.Policy = &GcpOrgPolicyRuleSet{Reset_: true, InheritFromParent: true}
		expectError(target, "reset: true must carry no rules")
	})

	ginkgo.It("should reject a condition without an expression and an empty list value", func() {
		target := minimal()
		target.Spec.Policy.Rules[0].Condition = &GcpOrgPolicyRuleCondition{Title: "no expression"}
		expectError(target, "expression")

		target = minimal()
		target.Spec.Policy.Rules = []*GcpOrgPolicyRule{{Kind: &GcpOrgPolicyRule_Values{Values: &GcpOrgPolicyRuleValues{AllowedValues: []string{""}}}}}
		expectError(target, "allowed_values")
	})

	ginkgo.It("should reject parameters that are not a JSON object", func() {
		target := minimal()
		target.Spec.Policy.Rules[0].Parameters = `["us-east1"]`
		expectError(target, "parameters must be a JSON object")
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})

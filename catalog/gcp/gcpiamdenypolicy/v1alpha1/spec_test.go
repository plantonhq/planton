package gcpiamdenypolicyv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpIamDenyPolicySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpIamDenyPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpIamDenyPolicy {
		return &GcpIamDenyPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpIamDenyPolicy",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-deny-policy",
			},
			Spec: &GcpIamDenyPolicySpec{
				Rules: []*GcpIamDenyPolicyRule{{
					DenyRule: &GcpIamDenyPolicyDenyRule{
						DeniedPrincipals:  []string{"principalSet://goog/public:all"},
						DeniedPermissions: []string{"secretmanager.googleapis.com/versions.access"},
					},
				}},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal deny policy", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept each parent arm alone", func() {
		target := minimal()
		target.Spec.Parent = &GcpIamDenyPolicyParent{ProjectId: litRef("my-gcp-project-123")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.Parent = &GcpIamDenyPolicyParent{FolderId: litRef("123456789")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.Parent = &GcpIamDenyPolicyParent{OrganizationId: "987654321"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an empty parent (ambient project)", func() {
		target := minimal()
		target.Spec.Parent = &GcpIamDenyPolicyParent{}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept exceptions and a denial condition", func() {
		target := minimal()
		target.Spec.PolicyName = "guard-secrets"
		target.Spec.DisplayName = "Guard break-glass secrets"
		target.Spec.Rules[0].Description = "Nobody reads unseal keys directly"
		target.Spec.Rules[0].DenyRule.ExceptionPrincipals = []string{
			"principal://iam.googleapis.com/projects/-/serviceAccounts/breakglass@p.iam.gserviceaccount.com",
		}
		target.Spec.Rules[0].DenyRule.ExceptionPermissions = []string{
			"secretmanager.googleapis.com/versions.list",
		}
		target.Spec.Rules[0].DenyRule.DenialCondition = &GcpIamDenyPolicyCondition{
			Expression:  "!resource.matchTag('12345678/env', 'sandbox')",
			Title:       "everywhere except sandbox",
			Description: "Sandbox projects are exempt from the guard",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a policy without rules", func() {
		target := minimal()
		target.Spec.Rules = nil
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a rule without a deny_rule body", func() {
		target := minimal()
		target.Spec.Rules = []*GcpIamDenyPolicyRule{{Description: "empty"}}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a denial condition without an expression", func() {
		target := minimal()
		target.Spec.Rules[0].DenyRule.DenialCondition = &GcpIamDenyPolicyCondition{
			Title: "no expression",
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject two parent arms together", func() {
		target := minimal()
		target.Spec.Parent = &GcpIamDenyPolicyParent{
			ProjectId: litRef("my-gcp-project-123"),
			FolderId:  litRef("123456789"),
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "at most one")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "ABANDON")).To(gomega.BeTrue())
	})
})

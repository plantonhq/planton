package cloudflarezerotrustaccesspolicyv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func ref(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func emailRule(email string) *CloudflareAccessRule {
	return &CloudflareAccessRule{Rule: &CloudflareAccessRule_Email{Email: &AccessRuleEmail{Email: email}}}
}

func validPolicy() *CloudflareZeroTrustAccessPolicy {
	return &CloudflareZeroTrustAccessPolicy{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareZeroTrustAccessPolicy",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-policy"},
		Spec: &CloudflareZeroTrustAccessPolicySpec{
			AccountId: validAccountID,
			Name:      "allow-staff",
			Decision:  CloudflareZeroTrustAccessPolicyDecision_allow,
			Include:   []*CloudflareAccessRule{emailRule("jane@example.com")},
		},
	}
}

func TestCloudflareZeroTrustAccessPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustAccessPolicySpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareZeroTrustAccessPolicySpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal allow policy", func() {
			gomega.Expect(protovalidate.Validate(validPolicy())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a policy with session, approval, mfa, and connection rules", func() {
			in := validPolicy()
			in.Spec.SessionDuration = proto("8h")
			in.Spec.ApprovalRequired = true
			in.Spec.ApprovalGroups = []*CloudflareZeroTrustAccessPolicyApprovalGroup{
				{ApprovalsNeeded: 1, EmailAddresses: []string{"approver@example.com"}},
			}
			in.Spec.PurposeJustificationRequired = true
			in.Spec.PurposeJustificationPrompt = "Why do you need access?"
			in.Spec.MfaConfig = &CloudflareZeroTrustAccessPolicyMfaConfig{
				AllowedAuthenticators: []CloudflareZeroTrustAccessPolicyMfaConfig_Authenticator{
					CloudflareZeroTrustAccessPolicyMfaConfig_totp,
				},
			}
			in.Spec.ConnectionRules = &CloudflareZeroTrustAccessPolicyConnectionRules{
				Rdp: &CloudflareZeroTrustAccessPolicyRdpRules{
					AllowedClipboardLocalToRemoteFormats: []string{"text"},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a non_identity decision with a service token rule", func() {
			in := validPolicy()
			in.Spec.Decision = CloudflareZeroTrustAccessPolicyDecision_non_identity
			in.Spec.Include = []*CloudflareAccessRule{
				{Rule: &CloudflareAccessRule_ServiceToken{ServiceToken: &AccessRuleServiceToken{TokenId: ref("token-abc")}}},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing account_id", func() {
			in := validPolicy()
			in.Spec.AccountId = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a non-hex account_id", func() {
			in := validPolicy()
			in.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unspecified decision", func() {
			in := validPolicy()
			in.Spec.Decision = CloudflareZeroTrustAccessPolicyDecision_decision_unspecified
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty include list", func() {
			in := validPolicy()
			in.Spec.Include = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a rule with no variant set", func() {
			in := validPolicy()
			in.Spec.Include = []*CloudflareAccessRule{{}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid session_duration", func() {
			in := validPolicy()
			in.Spec.SessionDuration = proto("forever")
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an approval group needing zero approvals", func() {
			in := validPolicy()
			in.Spec.ApprovalGroups = []*CloudflareZeroTrustAccessPolicyApprovalGroup{
				{ApprovalsNeeded: 0, EmailAddresses: []string{"a@example.com"}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an rdp clipboard format other than text", func() {
			in := validPolicy()
			in.Spec.ConnectionRules = &CloudflareZeroTrustAccessPolicyConnectionRules{
				Rdp: &CloudflareZeroTrustAccessPolicyRdpRules{
					AllowedClipboardLocalToRemoteFormats: []string{"image"},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})

func proto(s string) *string { return &s }

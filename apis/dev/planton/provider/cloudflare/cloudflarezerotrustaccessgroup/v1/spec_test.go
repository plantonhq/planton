package cloudflarezerotrustaccessgroupv1

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

func validGroup() *CloudflareZeroTrustAccessGroup {
	return &CloudflareZeroTrustAccessGroup{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareZeroTrustAccessGroup",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-group"},
		Spec: &CloudflareZeroTrustAccessGroupSpec{
			AccountId: validAccountID,
			Name:      "engineering",
			Include:   []*CloudflareAccessRule{emailRule("jane@example.com")},
		},
	}
}

func TestCloudflareZeroTrustAccessGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustAccessGroupSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareZeroTrustAccessGroupSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal account-scoped group", func() {
			gomega.Expect(protovalidate.Validate(validGroup())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a zone-scoped group", func() {
			in := validGroup()
			in.Spec.AccountId = ""
			in.Spec.ZoneId = ref("0123456789abcdef0123456789abcdef")
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts exclude and require rules", func() {
			in := validGroup()
			in.Spec.Exclude = []*CloudflareAccessRule{
				{Rule: &CloudflareAccessRule_EmailDomain{EmailDomain: &AccessRuleEmailDomain{Domain: "contractor.example.com"}}},
			}
			in.Spec.Require = []*CloudflareAccessRule{
				{Rule: &CloudflareAccessRule_Geo{Geo: &AccessRuleGeo{CountryCode: "US"}}},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a group-of-groups reference", func() {
			in := validGroup()
			in.Spec.Include = []*CloudflareAccessRule{
				{Rule: &CloudflareAccessRule_Group{Group: &AccessRuleGroupRef{Id: ref("group-abc")}}},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a user-risk-score rule", func() {
			in := validGroup()
			in.Spec.Include = []*CloudflareAccessRule{
				{Rule: &CloudflareAccessRule_UserRiskScore{UserRiskScore: &AccessRuleUserRiskScore{
					UserRiskScore: []AccessRuleUserRiskScore_Level{AccessRuleUserRiskScore_low, AccessRuleUserRiskScore_medium},
				}}},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an everyone rule", func() {
			in := validGroup()
			in.Spec.Include = []*CloudflareAccessRule{
				{Rule: &CloudflareAccessRule_Everyone{Everyone: &AccessRuleEmpty{}}},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects setting both account_id and zone_id", func() {
			in := validGroup()
			in.Spec.ZoneId = ref("0123456789abcdef0123456789abcdef")
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects setting neither account_id nor zone_id", func() {
			in := validGroup()
			in.Spec.AccountId = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a non-hex account_id", func() {
			in := validGroup()
			in.Spec.AccountId = "not-a-valid-account-id-string!!!"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty include list", func() {
			in := validGroup()
			in.Spec.Include = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a rule with no variant set", func() {
			in := validGroup()
			in.Spec.Include = []*CloudflareAccessRule{{}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid email in an email rule", func() {
			in := validGroup()
			in.Spec.Include = []*CloudflareAccessRule{emailRule("not-an-email")}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a geo rule with a non-two-letter country code", func() {
			in := validGroup()
			in.Spec.Include = []*CloudflareAccessRule{
				{Rule: &CloudflareAccessRule_Geo{Geo: &AccessRuleGeo{CountryCode: "USA"}}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a user-risk-score rule with no levels", func() {
			in := validGroup()
			in.Spec.Include = []*CloudflareAccessRule{
				{Rule: &CloudflareAccessRule_UserRiskScore{UserRiskScore: &AccessRuleUserRiskScore{}}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})

package azurefrontdoorsecuritypolicyv1alpha1

import (
	"fmt"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureFrontDoorSecurityPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorSecurityPolicySpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const profileId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd"

const firewallPolicyId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/frontDoorWebApplicationFirewallPolicies/edgewaf"

const endpointId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd/afdEndpoints/web"

const customDomainId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd/customDomains/www-example-com"

// minimal valid spec: the WAF associated with one endpoint's default
// domain.
func minimalSpec() *AzureFrontDoorSecurityPolicy {
	return &AzureFrontDoorSecurityPolicy{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorSecurityPolicy",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-front-door-security-policy",
		},
		Spec: &AzureFrontDoorSecurityPolicySpec{
			ProfileId:          literal(profileId),
			SecurityPolicyName: "edge-waf-attach",
			FirewallPolicyId:   literal(firewallPolicyId),
			DomainIds:          []*foreignkeyv1.StringValueOrRef{literal(endpointId)},
		},
	}
}

var _ = ginkgo.Describe("AzureFrontDoorSecurityPolicySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for the minimal spec", func() {
			err := protovalidate.Validate(minimalSpec())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a mixed list of endpoint and custom-domain IDs", func() {
			input := minimalSpec()
			input.Spec.DomainIds = []*foreignkeyv1.StringValueOrRef{
				literal(endpointId),
				literal(customDomainId),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a single-character alphanumeric name", func() {
			input := minimalSpec()
			input.Spec.SecurityPolicyName = "w"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("required fields", func() {

		ginkgo.It("should reject a missing profile reference", func() {
			input := minimalSpec()
			input.Spec.ProfileId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing WAF policy reference", func() {
			input := minimalSpec()
			input.Spec.FirewallPolicyId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a present-but-empty reference (the StringValueOrRef non-empty rule)", func() {
			input := minimalSpec()
			input.Spec.FirewallPolicyId = literal("")
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty domain list", func() {
			input := minimalSpec()
			input.Spec.DomainIds = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			input := minimalSpec()
			input.Spec.SecurityPolicyName = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("security_policy_name format", func() {

		ginkgo.It("should reject a name starting with a hyphen", func() {
			input := minimalSpec()
			input.Spec.SecurityPolicyName = "-edge-waf"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a name ending with a hyphen", func() {
			input := minimalSpec()
			input.Spec.SecurityPolicyName = "edge-waf-"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a name with characters beyond letters, digits, and hyphens", func() {
			input := minimalSpec()
			input.Spec.SecurityPolicyName = "edge_waf"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("domain list bounds", func() {

		ginkgo.It("should accept exactly 500 domains (the Premium cap)", func() {
			domains := make([]*foreignkeyv1.StringValueOrRef, 0, 500)
			for i := 0; i < 500; i++ {
				domains = append(domains, literal(fmt.Sprintf("%s%d", endpointId, i)))
			}
			input := minimalSpec()
			input.Spec.DomainIds = domains
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject more than 500 domains", func() {
			domains := make([]*foreignkeyv1.StringValueOrRef, 0, 501)
			for i := 0; i < 501; i++ {
				domains = append(domains, literal(fmt.Sprintf("%s%d", endpointId, i)))
			}
			input := minimalSpec()
			input.Spec.DomainIds = domains
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})

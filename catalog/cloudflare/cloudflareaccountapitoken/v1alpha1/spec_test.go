package cloudflareaccountapitokenv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestCloudflareAccountApiTokenSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareAccountApiTokenSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func validToken(spec *CloudflareAccountApiTokenSpec) *CloudflareAccountApiToken {
	return &CloudflareAccountApiToken{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareAccountApiToken",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-account-api-token",
		},
		Spec: spec,
	}
}

func wholeAccountPolicy() *CloudflareAccountApiTokenPolicy {
	return &CloudflareAccountApiTokenPolicy{
		Effect:             "allow",
		PermissionGroupIds: []string{"c8fed203ed3043cba015a93ad1616f1f"},
		Resources: map[string]*CloudflareAccountApiTokenResourceScope{
			"com.cloudflare.api.account." + testAccountID: {Permission: "*"},
		},
	}
}

func baseSpec() *CloudflareAccountApiTokenSpec {
	return &CloudflareAccountApiTokenSpec{
		AccountId: testAccountID,
		Name:      "readonly-dns",
		Policies:  []*CloudflareAccountApiTokenPolicy{wholeAccountPolicy()},
	}
}

var _ = ginkgo.Describe("CloudflareAccountApiTokenSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a whole-account allow policy", func() {
			gomega.Expect(protovalidate.Validate(validToken(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a nested sub-resource scoping", func() {
			spec := baseSpec()
			spec.Policies[0].Resources = map[string]*CloudflareAccountApiTokenResourceScope{
				"com.cloudflare.api.account." + testAccountID: {
					Subresources: map[string]string{"com.cloudflare.api.account.zone.*": "*"},
				},
			}
			gomega.Expect(protovalidate.Validate(validToken(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a validity window, condition, and status", func() {
			spec := baseSpec()
			spec.NotBefore = "2026-09-01T00:00:00Z"
			spec.ExpiresOn = "2027-09-01T00:00:00Z"
			spec.Status = "disabled"
			spec.Condition = &CloudflareAccountApiTokenCondition{
				RequestIp: &CloudflareAccountApiTokenRequestIp{
					InCidrs:    []string{"198.51.100.0/24"},
					NotInCidrs: []string{"198.51.100.128/25"},
				},
			}
			gomega.Expect(protovalidate.Validate(validToken(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a deny policy alongside an allow policy", func() {
			spec := baseSpec()
			deny := wholeAccountPolicy()
			deny.Effect = "deny"
			deny.Resources = map[string]*CloudflareAccountApiTokenResourceScope{
				"com.cloudflare.api.account.zone.023e105f4ecef8ad9ca31a8372d0c353": {Permission: "*"},
			}
			spec.Policies = append(spec.Policies, deny)
			gomega.Expect(protovalidate.Validate(validToken(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an empty policies list", func() {
			spec := baseSpec()
			spec.Policies = nil
			gomega.Expect(protovalidate.Validate(validToken(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown effect", func() {
			spec := baseSpec()
			spec.Policies[0].Effect = "block"
			gomega.Expect(protovalidate.Validate(validToken(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a policy without permission groups", func() {
			spec := baseSpec()
			spec.Policies[0].PermissionGroupIds = nil
			gomega.Expect(protovalidate.Validate(validToken(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a policy without resources", func() {
			spec := baseSpec()
			spec.Policies[0].Resources = nil
			gomega.Expect(protovalidate.Validate(validToken(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a resource scope with both grant forms", func() {
			spec := baseSpec()
			spec.Policies[0].Resources = map[string]*CloudflareAccountApiTokenResourceScope{
				"com.cloudflare.api.account." + testAccountID: {
					Permission:   "*",
					Subresources: map[string]string{"com.cloudflare.api.account.zone.*": "*"},
				},
			}
			gomega.Expect(protovalidate.Validate(validToken(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a resource scope with neither grant form", func() {
			spec := baseSpec()
			spec.Policies[0].Resources = map[string]*CloudflareAccountApiTokenResourceScope{
				"com.cloudflare.api.account." + testAccountID: {},
			}
			gomega.Expect(protovalidate.Validate(validToken(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed expiry", func() {
			spec := baseSpec()
			spec.ExpiresOn = "2027-09-01"
			gomega.Expect(protovalidate.Validate(validToken(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an inverted validity window", func() {
			spec := baseSpec()
			spec.NotBefore = "2027-09-01T00:00:00Z"
			spec.ExpiresOn = "2026-09-01T00:00:00Z"
			gomega.Expect(protovalidate.Validate(validToken(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a server-only status", func() {
			spec := baseSpec()
			spec.Status = "expired"
			gomega.Expect(protovalidate.Validate(validToken(spec))).NotTo(gomega.BeNil())
		})
	})
})

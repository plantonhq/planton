package cloudflarezerotrustaccessapplicationv1

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

func validApp() *CloudflareZeroTrustAccessApplication {
	return &CloudflareZeroTrustAccessApplication{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareZeroTrustAccessApplication",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-app"},
		Spec: &CloudflareZeroTrustAccessApplicationSpec{
			AccountId: validAccountID,
			Name:      "internal-app",
			Type:      CloudflareZeroTrustAccessApplicationType_self_hosted,
			Domain:    "app.example.com",
			Policies: []*CloudflareZeroTrustAccessApplicationPolicyRef{
				{Policy: ref("policy-abc"), Precedence: 1},
			},
		},
	}
}

func TestCloudflareZeroTrustAccessApplicationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustAccessApplicationSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareZeroTrustAccessApplicationSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a self-hosted app", func() {
			gomega.Expect(protovalidate.Validate(validApp())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a zone-scoped app", func() {
			in := validApp()
			in.Spec.AccountId = ""
			in.Spec.ZoneId = ref("0123456789abcdef0123456789abcdef")
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an app_launcher app without a domain", func() {
			in := validApp()
			in.Spec.Type = CloudflareZeroTrustAccessApplicationType_app_launcher
			in.Spec.Domain = ""
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a SaaS (OIDC) app", func() {
			in := validApp()
			in.Spec.Type = CloudflareZeroTrustAccessApplicationType_saas
			in.Spec.Domain = ""
			in.Spec.SaasApp = &CloudflareZeroTrustAccessSaasApp{
				AuthType:     CloudflareZeroTrustAccessSaasApp_oidc,
				RedirectUris: []string{"https://saas.example.com/callback"},
				GrantTypes:   []CloudflareZeroTrustAccessSaasApp_GrantType{CloudflareZeroTrustAccessSaasApp_authorization_code},
				Scopes:       []CloudflareZeroTrustAccessSaasScope{CloudflareZeroTrustAccessSaasScope_openid, CloudflareZeroTrustAccessSaasScope_email},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts cors headers and a scim config", func() {
			in := validApp()
			in.Spec.CorsHeaders = &CloudflareZeroTrustAccessApplicationCorsHeaders{
				AllowedMethods: []CloudflareZeroTrustAccessApplicationCorsHeaders_Method{
					CloudflareZeroTrustAccessApplicationCorsHeaders_GET,
				},
				AllowedOrigins: []string{"https://example.com"},
				MaxAge:         600,
			}
			in.Spec.ScimConfig = &CloudflareZeroTrustAccessScimConfig{
				IdpUid:    ref("idp-123"),
				RemoteUri: "https://scim.example.com/v2",
				Authentication: &CloudflareZeroTrustAccessScimAuthentication{
					Scheme: CloudflareZeroTrustAccessScimAuthentication_oauthbearertoken,
					Token:  "super-secret-token",
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an infrastructure app with target criteria", func() {
			in := validApp()
			in.Spec.Type = CloudflareZeroTrustAccessApplicationType_infrastructure
			in.Spec.Domain = ""
			in.Spec.TargetCriteria = []*CloudflareZeroTrustAccessApplicationTargetCriteria{
				{
					Port:             22,
					Protocol:         CloudflareZeroTrustAccessApplicationTargetCriteria_SSH,
					TargetAttributes: []*CloudflareZeroTrustAccessApplicationTargetAttribute{{Name: "hostname", Values: []string{"db-1"}}},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects setting both account_id and zone_id", func() {
			in := validApp()
			in.Spec.ZoneId = ref("0123456789abcdef0123456789abcdef")
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects setting neither account_id nor zone_id", func() {
			in := validApp()
			in.Spec.AccountId = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a self_hosted app without a domain", func() {
			in := validApp()
			in.Spec.Domain = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a non-hex account_id", func() {
			in := validApp()
			in.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid same_site_cookie_attribute", func() {
			in := validApp()
			in.Spec.SameSiteCookieAttribute = "weird"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a cors max_age above the allowed range", func() {
			in := validApp()
			in.Spec.CorsHeaders = &CloudflareZeroTrustAccessApplicationCorsHeaders{MaxAge: 99999}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a target criteria with an invalid protocol", func() {
			in := validApp()
			in.Spec.Type = CloudflareZeroTrustAccessApplicationType_infrastructure
			in.Spec.Domain = ""
			in.Spec.TargetCriteria = []*CloudflareZeroTrustAccessApplicationTargetCriteria{
				{
					Port:             22,
					Protocol:         CloudflareZeroTrustAccessApplicationTargetCriteria_protocol_unspecified,
					TargetAttributes: []*CloudflareZeroTrustAccessApplicationTargetAttribute{{Name: "hostname", Values: []string{"db-1"}}},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a policy ref without a policy id", func() {
			in := validApp()
			in.Spec.Policies = []*CloudflareZeroTrustAccessApplicationPolicyRef{{}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})

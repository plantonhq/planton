package azurefrontdoorroutev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureFrontDoorRouteSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorRouteSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const endpointId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd/afdEndpoints/web"
const originGroupId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd/originGroups/api-backends"
const originId = originGroupId + "/origins/primary"

// minimal valid spec: an HTTPS-redirecting catch-all route (both
// protocols, matching Azure's defaults).
func minimalSpec() *AzureFrontDoorRoute {
	return &AzureFrontDoorRoute{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorRoute",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-front-door-route",
		},
		Spec: &AzureFrontDoorRouteSpec{
			EndpointId:      literal(endpointId),
			RouteName:       "test-route",
			OriginGroupId:   literal(originGroupId),
			PatternsToMatch: []string{"/*"},
			SupportedProtocols: []AzureFrontDoorRouteProtocol{
				AzureFrontDoorRouteProtocol_HTTP,
				AzureFrontDoorRouteProtocol_HTTPS,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureFrontDoorRouteSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal route", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept origin ids for deploy ordering", func() {
			input := minimalSpec()
			input.Spec.OriginIds = []*foreignkeyv1.StringValueOrRef{literal(originId)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept HTTPS-only when the redirect is disabled", func() {
			input := minimalSpec()
			input.Spec.SupportedProtocols = []AzureFrontDoorRouteProtocol{AzureFrontDoorRouteProtocol_HTTPS}
			input.Spec.HttpsRedirectEnabled = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept HTTP-only when the redirect is disabled", func() {
			input := minimalSpec()
			input.Spec.SupportedProtocols = []AzureFrontDoorRouteProtocol{AzureFrontDoorRouteProtocol_HTTP}
			input.Spec.HttpsRedirectEnabled = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every forwarding protocol", func() {
			for _, protocol := range []AzureFrontDoorRouteForwardingProtocol{
				AzureFrontDoorRouteForwardingProtocol_MATCH_REQUEST,
				AzureFrontDoorRouteForwardingProtocol_HTTP_ONLY,
				AzureFrontDoorRouteForwardingProtocol_HTTPS_ONLY,
			} {
				input := minimalSpec()
				input.Spec.ForwardingProtocol = protocol
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "forwarding protocol %v must be accepted", protocol)
			}
		})

		ginkgo.It("should accept multiple patterns", func() {
			input := minimalSpec()
			input.Spec.PatternsToMatch = []string{"/api/*", "/static/*", "/"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an origin path and disabled state", func() {
			input := minimalSpec()
			input.Spec.OriginPath = proto.String("/site1")
			input.Spec.Enabled = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full cache block", func() {
			input := minimalSpec()
			input.Spec.Cache = &AzureFrontDoorRouteCache{
				QueryStringCachingBehavior: AzureFrontDoorRouteQueryStringCachingBehavior_INCLUDE_SPECIFIED_QUERY_STRINGS,
				QueryStrings:               []string{"page", "sort"},
				CompressionEnabled:         proto.Bool(true),
				ContentTypesToCompress:     []string{"text/html", "application/json", "image/svg+xml"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every query-string caching behavior", func() {
			for _, behavior := range []AzureFrontDoorRouteQueryStringCachingBehavior{
				AzureFrontDoorRouteQueryStringCachingBehavior_IGNORE_QUERY_STRING,
				AzureFrontDoorRouteQueryStringCachingBehavior_USE_QUERY_STRING,
				AzureFrontDoorRouteQueryStringCachingBehavior_IGNORE_SPECIFIED_QUERY_STRINGS,
				AzureFrontDoorRouteQueryStringCachingBehavior_INCLUDE_SPECIFIED_QUERY_STRINGS,
			} {
				input := minimalSpec()
				input.Spec.Cache = &AzureFrontDoorRouteCache{QueryStringCachingBehavior: behavior}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "behavior %v must be accepted", behavior)
			}
		})

		ginkgo.It("should accept route name boundaries (2 and 90 characters)", func() {
			input := minimalSpec()
			input.Spec.RouteName = "ab"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.RouteName = "a" + strings.Repeat("b", 88) + "c"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept attached rule sets", func() {
			input := minimalSpec()
			input.Spec.RuleSetIds = []*foreignkeyv1.StringValueOrRef{
				literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd/ruleSets/deliverypolicy"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept custom domains with the default domain disabled", func() {
			input := minimalSpec()
			input.Spec.CustomDomainIds = []*foreignkeyv1.StringValueOrRef{
				literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd/customDomains/www-example-com"),
			}
			input.Spec.LinkToDefaultDomain = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing endpoint reference", func() {
			input := minimalSpec()
			input.Spec.EndpointId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing origin group reference", func() {
			input := minimalSpec()
			input.Spec.OriginGroupId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing route name", func() {
			input := minimalSpec()
			input.Spec.RouteName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a route name over 90 characters", func() {
			input := minimalSpec()
			input.Spec.RouteName = "a" + strings.Repeat("b", 89) + "c"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a route name with a trailing hyphen", func() {
			input := minimalSpec()
			input.Spec.RouteName = "route-"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject empty patterns", func() {
			input := minimalSpec()
			input.Spec.PatternsToMatch = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a pattern not starting with '/'", func() {
			input := minimalSpec()
			input.Spec.PatternsToMatch = []string{"api/*"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject empty supported protocols", func() {
			input := minimalSpec()
			input.Spec.SupportedProtocols = nil
			input.Spec.HttpsRedirectEnabled = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate supported protocols", func() {
			input := minimalSpec()
			input.Spec.SupportedProtocols = []AzureFrontDoorRouteProtocol{
				AzureFrontDoorRouteProtocol_HTTPS,
				AzureFrontDoorRouteProtocol_HTTPS,
			}
			input.Spec.HttpsRedirectEnabled = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unspecified supported protocol", func() {
			input := minimalSpec()
			input.Spec.SupportedProtocols = []AzureFrontDoorRouteProtocol{
				AzureFrontDoorRouteProtocol_azure_front_door_route_protocol_unspecified,
			}
			input.Spec.HttpsRedirectEnabled = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject HTTPS-only with the redirect left at its default (enabled)", func() {
			input := minimalSpec()
			input.Spec.SupportedProtocols = []AzureFrontDoorRouteProtocol{AzureFrontDoorRouteProtocol_HTTPS}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject HTTP-only with the redirect explicitly enabled", func() {
			input := minimalSpec()
			input.Spec.SupportedProtocols = []AzureFrontDoorRouteProtocol{AzureFrontDoorRouteProtocol_HTTP}
			input.Spec.HttpsRedirectEnabled = proto.Bool(true)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined forwarding protocol", func() {
			input := minimalSpec()
			input.Spec.ForwardingProtocol = AzureFrontDoorRouteForwardingProtocol(99)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty origin path when set", func() {
			input := minimalSpec()
			input.Spec.OriginPath = proto.String("")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a query string containing a comma", func() {
			input := minimalSpec()
			input.Spec.Cache = &AzureFrontDoorRouteCache{
				QueryStrings: []string{"page,sort"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a content type outside Azure's compression list", func() {
			input := minimalSpec()
			input.Spec.Cache = &AzureFrontDoorRouteCache{
				ContentTypesToCompress: []string{"video/mp4"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined query-string caching behavior", func() {
			input := minimalSpec()
			input.Spec.Cache = &AzureFrontDoorRouteCache{
				QueryStringCachingBehavior: AzureFrontDoorRouteQueryStringCachingBehavior(99),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject disabling the default domain without custom domains", func() {
			input := minimalSpec()
			input.Spec.LinkToDefaultDomain = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong kind", func() {
			input := minimalSpec()
			input.Kind = "AzureFrontDoorRoutes"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing metadata", func() {
			input := minimalSpec()
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})

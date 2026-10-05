package azurefrontdoorrulesetv1alpha1

import (
	"fmt"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureFrontDoorRuleSetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorRuleSetSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const profileId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd"

const originGroupId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd/originGroups/api-backends"

// minimal valid spec: an empty rule set (a placeholder routes can
// already attach).
func minimalSpec() *AzureFrontDoorRuleSet {
	return &AzureFrontDoorRuleSet{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorRuleSet",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-front-door-rule-set",
		},
		Spec: &AzureFrontDoorRuleSetSpec{
			ProfileId:   literal(profileId),
			RuleSetName: "deliverypolicy",
		},
	}
}

// headerRule is the simplest valid rule: no conditions, one response
// header action.
func headerRule(name string, order int32) *AzureFrontDoorRule {
	return &AzureFrontDoorRule{
		Name:  name,
		Order: order,
		Actions: &AzureFrontDoorRuleActions{
			ResponseHeaders: []*AzureFrontDoorRuleHeaderAction{{
				HeaderAction: AzureFrontDoorRuleHeaderActionType_OVERWRITE,
				HeaderName:   "Strict-Transport-Security",
				Value:        "max-age=31536000; includeSubDomains",
			}},
		},
	}
}

func withRules(rules ...*AzureFrontDoorRule) *AzureFrontDoorRuleSet {
	input := minimalSpec()
	input.Spec.Rules = rules
	return input
}

var _ = ginkgo.Describe("AzureFrontDoorRuleSetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept an empty rule set", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a condition-less security-headers rule", func() {
			gomega.Expect(protovalidate.Validate(withRules(headerRule("securityheaders", 1)))).To(gomega.BeNil())
		})

		ginkgo.It("should accept rule set name boundaries (1 and 60 characters)", func() {
			input := minimalSpec()
			input.Spec.RuleSetName = "a"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.RuleSetName = "a234567890b234567890c234567890d234567890e234567890f23456789"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a redirect rule that stops evaluation", func() {
			rule := &AzureFrontDoorRule{
				Name:            "apexredirect",
				Order:           1,
				BehaviorOnMatch: AzureFrontDoorRuleBehaviorOnMatch_STOP,
				Conditions: &AzureFrontDoorRuleConditions{
					HostName: []*AzureFrontDoorRuleHostNameCondition{{
						Operator:    AzureFrontDoorRuleOperator_EQUAL,
						MatchValues: []string{"example.com"},
						Transforms:  []AzureFrontDoorRuleTransform{AzureFrontDoorRuleTransform_LOWERCASE},
					}},
				},
				Actions: &AzureFrontDoorRuleActions{
					UrlRedirect: &AzureFrontDoorRuleUrlRedirectAction{
						RedirectType:        AzureFrontDoorRuleRedirectType_PERMANENT_REDIRECT,
						RedirectProtocol:    AzureFrontDoorRuleForwardingProtocol_HTTPS_ONLY,
						DestinationHostname: "www.example.com",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a rewrite rule preserving the unmatched path", func() {
			rule := &AzureFrontDoorRule{
				Name:  "apiversionrewrite",
				Order: 2,
				Conditions: &AzureFrontDoorRuleConditions{
					UrlPath: []*AzureFrontDoorRuleUrlPathCondition{{
						Operator:    AzureFrontDoorRuleOperator_BEGINS_WITH,
						MatchValues: []string{"v1/"},
					}},
				},
				Actions: &AzureFrontDoorRuleActions{
					UrlRewrite: &AzureFrontDoorRuleUrlRewriteAction{
						SourcePattern:         "/v1",
						Destination:           "/api/v1",
						PreserveUnmatchedPath: true,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full caching override with specified query strings", func() {
			enabled := true
			rule := &AzureFrontDoorRule{
				Name:  "staticcache",
				Order: 3,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior:              AzureFrontDoorRuleCacheBehavior_OVERRIDE_ALWAYS,
						CacheDuration:              "1.12:00:00",
						QueryStringCachingBehavior: AzureFrontDoorRuleQueryStringCachingBehavior_INCLUDE_SPECIFIED_QUERY_STRINGS,
						QueryStringParameters:      []string{"page", "lang"},
						CompressionEnabled:         &enabled,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept an HONOR_ORIGIN override without a duration", func() {
			rule := &AzureFrontDoorRule{
				Name:  "honororigin",
				Order: 4,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior:              AzureFrontDoorRuleCacheBehavior_HONOR_ORIGIN,
						QueryStringCachingBehavior: AzureFrontDoorRuleQueryStringCachingBehavior_IGNORE_QUERY_STRING,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a bare DISABLED caching override", func() {
			rule := &AzureFrontDoorRule{
				Name:  "nocache",
				Order: 5,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior: AzureFrontDoorRuleCacheBehavior_DISABLED,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept an origin override paired with a forwarding protocol", func() {
			rule := &AzureFrontDoorRule{
				Name:  "canaryorigin",
				Order: 6,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						OriginGroupId:      literal(originGroupId),
						ForwardingProtocol: AzureFrontDoorRuleForwardingProtocol_HTTPS_ONLY,
						CacheBehavior:      AzureFrontDoorRuleCacheBehavior_DISABLED,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept the ANY operator with empty match values", func() {
			rule := headerRule("anyquerystring", 7)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				QueryString: []*AzureFrontDoorRuleQueryStringCondition{{
					Operator: AzureFrontDoorRuleOperator_ANY,
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a GEO_MATCH remote address condition with country codes", func() {
			rule := headerRule("geoblock", 8)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RemoteAddress: []*AzureFrontDoorRuleRemoteAddressCondition{{
					Operator:        AzureFrontDoorRuleOperator_GEO_MATCH,
					NegateCondition: true,
					MatchValues:     []string{"US", "DE"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept an unspecified remote address operator (deploys IP_MATCH) with CIDRs", func() {
			rule := headerRule("cidrmatch", 9)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RemoteAddress: []*AzureFrontDoorRuleRemoteAddressCondition{{
					MatchValues: []string{"203.0.113.0/24"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept the closed-vocabulary conditions", func() {
			rule := headerRule("closedvocab", 10)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RequestMethod: []*AzureFrontDoorRuleRequestMethodCondition{{
					MatchValues: []string{"GET", "POST"},
				}},
				RequestScheme: []*AzureFrontDoorRuleRequestSchemeCondition{{
					MatchValue: strPtr("HTTPS"),
				}},
				HttpVersion: []*AzureFrontDoorRuleHttpVersionCondition{{
					MatchValues: []string{"2.0", "1.1"},
				}},
				IsDevice: []*AzureFrontDoorRuleIsDeviceCondition{{
					MatchValue: "Mobile",
				}},
				ServerPort: []*AzureFrontDoorRuleServerPortCondition{{
					Operator:    AzureFrontDoorRuleOperator_EQUAL,
					MatchValues: []string{"443"},
				}},
				SslProtocol: []*AzureFrontDoorRuleSslProtocolCondition{{
					MatchValues: []string{"TLSv1.2"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept the WILDCARD operator on a url path condition", func() {
			rule := headerRule("wildcardpath", 11)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				UrlPath: []*AzureFrontDoorRuleUrlPathCondition{{
					Operator:    AzureFrontDoorRuleOperator_WILDCARD,
					MatchValues: []string{"files/*/download"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept exactly 10 conditions on one rule", func() {
			rule := headerRule("tenconditions", 12)
			headers := make([]*AzureFrontDoorRuleRequestHeaderCondition, 10)
			for i := range headers {
				headers[i] = &AzureFrontDoorRuleRequestHeaderCondition{
					HeaderName:  fmt.Sprintf("X-Header-%d", i),
					Operator:    AzureFrontDoorRuleOperator_EQUAL,
					MatchValues: []string{"v"},
				}
			}
			rule.Conditions = &AzureFrontDoorRuleConditions{RequestHeader: headers}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})

		ginkgo.It("should accept exactly 5 actions on one rule", func() {
			rule := &AzureFrontDoorRule{
				Name:  "fiveactions",
				Order: 13,
				Actions: &AzureFrontDoorRuleActions{
					UrlRewrite: &AzureFrontDoorRuleUrlRewriteAction{
						SourcePattern: "/",
						Destination:   "/app",
					},
					RequestHeaders: []*AzureFrontDoorRuleHeaderAction{
						{HeaderAction: AzureFrontDoorRuleHeaderActionType_APPEND, HeaderName: "X-Fwd", Value: "1"},
						{HeaderAction: AzureFrontDoorRuleHeaderActionType_DELETE, HeaderName: "X-Debug"},
					},
					ResponseHeaders: []*AzureFrontDoorRuleHeaderAction{
						{HeaderAction: AzureFrontDoorRuleHeaderActionType_OVERWRITE, HeaderName: "X-Frame-Options", Value: "DENY"},
					},
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior: AzureFrontDoorRuleCacheBehavior_DISABLED,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing profile reference", func() {
			input := minimalSpec()
			input.Spec.ProfileId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule set name with hyphens", func() {
			input := minimalSpec()
			input.Spec.RuleSetName = "delivery-policy"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule set name starting with a digit", func() {
			input := minimalSpec()
			input.Spec.RuleSetName = "1policy"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule set name over 60 characters", func() {
			input := minimalSpec()
			input.Spec.RuleSetName = "a234567890b234567890c234567890d234567890e234567890f2345678901"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate rule names in one set", func() {
			gomega.Expect(protovalidate.Validate(withRules(headerRule("dup", 1), headerRule("dup", 2)))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule name with hyphens", func() {
			gomega.Expect(protovalidate.Validate(withRules(headerRule("security-headers", 1)))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a negative rule order", func() {
			gomega.Expect(protovalidate.Validate(withRules(headerRule("negorder", -1)))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule without actions", func() {
			rule := headerRule("noactions", 1)
			rule.Actions = nil
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule with an empty actions block", func() {
			rule := headerRule("emptyactions", 1)
			rule.Actions = &AzureFrontDoorRuleActions{}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject 11 conditions on one rule", func() {
			rule := headerRule("elevenconditions", 1)
			headers := make([]*AzureFrontDoorRuleRequestHeaderCondition, 11)
			for i := range headers {
				headers[i] = &AzureFrontDoorRuleRequestHeaderCondition{
					HeaderName:  fmt.Sprintf("X-Header-%d", i),
					Operator:    AzureFrontDoorRuleOperator_EQUAL,
					MatchValues: []string{"v"},
				}
			}
			rule.Conditions = &AzureFrontDoorRuleConditions{RequestHeader: headers}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject 6 actions on one rule", func() {
			rule := headerRule("sixactions", 1)
			rule.Actions.RequestHeaders = []*AzureFrontDoorRuleHeaderAction{
				{HeaderAction: AzureFrontDoorRuleHeaderActionType_APPEND, HeaderName: "A", Value: "1"},
				{HeaderAction: AzureFrontDoorRuleHeaderActionType_APPEND, HeaderName: "B", Value: "1"},
				{HeaderAction: AzureFrontDoorRuleHeaderActionType_APPEND, HeaderName: "C", Value: "1"},
				{HeaderAction: AzureFrontDoorRuleHeaderActionType_APPEND, HeaderName: "D", Value: "1"},
				{HeaderAction: AzureFrontDoorRuleHeaderActionType_APPEND, HeaderName: "E", Value: "1"},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a redirect and a rewrite on one rule", func() {
			rule := &AzureFrontDoorRule{
				Name:  "redirectandrewrite",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					UrlRedirect: &AzureFrontDoorRuleUrlRedirectAction{
						RedirectType:        AzureFrontDoorRuleRedirectType_MOVED,
						DestinationHostname: "www.example.com",
					},
					UrlRewrite: &AzureFrontDoorRuleUrlRewriteAction{
						SourcePattern: "/",
						Destination:   "/app",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the ANY operator with match values", func() {
			rule := headerRule("anywithvalues", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				QueryString: []*AzureFrontDoorRuleQueryStringCondition{{
					Operator:    AzureFrontDoorRuleOperator_ANY,
					MatchValues: []string{"v"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a comparison operator without match values", func() {
			rule := headerRule("equalnovalues", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				QueryString: []*AzureFrontDoorRuleQueryStringCondition{{
					Operator: AzureFrontDoorRuleOperator_EQUAL,
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unspecified operator on a standard condition", func() {
			rule := headerRule("nooperator", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				QueryString: []*AzureFrontDoorRuleQueryStringCondition{{
					MatchValues: []string{"v"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the WILDCARD operator outside url path conditions", func() {
			rule := headerRule("wildcardhostname", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				HostName: []*AzureFrontDoorRuleHostNameCondition{{
					Operator:    AzureFrontDoorRuleOperator_WILDCARD,
					MatchValues: []string{"*.example.com"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the ANY operator on a request body condition", func() {
			rule := headerRule("bodyany", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RequestBody: []*AzureFrontDoorRuleRequestBodyCondition{{
					Operator:    AzureFrontDoorRuleOperator_ANY,
					MatchValues: []string{"payload"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a lowercase GEO_MATCH country code", func() {
			rule := headerRule("geobadcode", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RemoteAddress: []*AzureFrontDoorRuleRemoteAddressCondition{{
					Operator:    AzureFrontDoorRuleOperator_GEO_MATCH,
					MatchValues: []string{"us"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an EQUAL operator on a remote address condition", func() {
			rule := headerRule("remoteequal", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RemoteAddress: []*AzureFrontDoorRuleRemoteAddressCondition{{
					Operator:    AzureFrontDoorRuleOperator_EQUAL,
					MatchValues: []string{"203.0.113.1"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a GEO_MATCH operator on a socket address condition", func() {
			rule := headerRule("socketgeo", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				SocketAddress: []*AzureFrontDoorRuleSocketAddressCondition{{
					Operator:    AzureFrontDoorRuleOperator_GEO_MATCH,
					MatchValues: []string{"US"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the ANY operator on a remote address condition", func() {
			rule := headerRule("remoteany", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RemoteAddress: []*AzureFrontDoorRuleRemoteAddressCondition{{
					Operator: AzureFrontDoorRuleOperator_ANY,
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the ANY operator on a socket address condition", func() {
			rule := headerRule("socketany", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				SocketAddress: []*AzureFrontDoorRuleSocketAddressCondition{{
					Operator: AzureFrontDoorRuleOperator_ANY,
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a remote address condition with no match values", func() {
			rule := headerRule("remotenovalues", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RemoteAddress: []*AzureFrontDoorRuleRemoteAddressCondition{{
					Operator: AzureFrontDoorRuleOperator_IP_MATCH,
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule combining a redirect with a route-configuration override", func() {
			rule := &AzureFrontDoorRule{
				Name:  "redirectplusoverride",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					UrlRedirect: &AzureFrontDoorRuleUrlRedirectAction{
						RedirectType:        AzureFrontDoorRuleRedirectType_PERMANENT_REDIRECT,
						DestinationHostname: "www.example.com",
					},
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior: AzureFrontDoorRuleCacheBehavior_DISABLED,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown request method value", func() {
			rule := headerRule("badmethod", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RequestMethod: []*AzureFrontDoorRuleRequestMethodCondition{{
					MatchValues: []string{"PATCH"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty request method condition", func() {
			rule := headerRule("emptymethods", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RequestMethod: []*AzureFrontDoorRuleRequestMethodCondition{{}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown http version value", func() {
			rule := headerRule("badhttpversion", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				HttpVersion: []*AzureFrontDoorRuleHttpVersionCondition{{
					MatchValues: []string{"3.0"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown ssl protocol value", func() {
			rule := headerRule("badsslprotocol", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				SslProtocol: []*AzureFrontDoorRuleSslProtocolCondition{{
					MatchValues: []string{"TLSv1.3"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown server port value", func() {
			rule := headerRule("badserverport", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				ServerPort: []*AzureFrontDoorRuleServerPortCondition{{
					Operator:    AzureFrontDoorRuleOperator_EQUAL,
					MatchValues: []string{"8080"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown device class", func() {
			rule := headerRule("baddevice", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				IsDevice: []*AzureFrontDoorRuleIsDeviceCondition{{
					MatchValue: "Tablet",
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown request scheme", func() {
			rule := headerRule("badscheme", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RequestScheme: []*AzureFrontDoorRuleRequestSchemeCondition{{
					MatchValue: strPtr("FTP"),
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject more than 25 match values", func() {
			rule := headerRule("toomanyvalues", 1)
			values := make([]string, 26)
			for i := range values {
				values[i] = fmt.Sprintf("v%d", i)
			}
			rule.Conditions = &AzureFrontDoorRuleConditions{
				QueryString: []*AzureFrontDoorRuleQueryStringCondition{{
					Operator:    AzureFrontDoorRuleOperator_EQUAL,
					MatchValues: values,
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate transforms", func() {
			rule := headerRule("duptransforms", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				QueryString: []*AzureFrontDoorRuleQueryStringCondition{{
					Operator:    AzureFrontDoorRuleOperator_EQUAL,
					MatchValues: []string{"v"},
					Transforms: []AzureFrontDoorRuleTransform{
						AzureFrontDoorRuleTransform_LOWERCASE,
						AzureFrontDoorRuleTransform_LOWERCASE,
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a DELETE header action carrying a value", func() {
			rule := headerRule("deletewithvalue", 1)
			rule.Actions.ResponseHeaders[0] = &AzureFrontDoorRuleHeaderAction{
				HeaderAction: AzureFrontDoorRuleHeaderActionType_DELETE,
				HeaderName:   "X-Debug",
				Value:        "1",
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an APPEND header action without a value", func() {
			rule := headerRule("appendnovalue", 1)
			rule.Actions.ResponseHeaders[0] = &AzureFrontDoorRuleHeaderAction{
				HeaderAction: AzureFrontDoorRuleHeaderActionType_APPEND,
				HeaderName:   "X-Fwd",
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a redirect destination path without a leading slash", func() {
			rule := &AzureFrontDoorRule{
				Name:  "badredirectpath",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					UrlRedirect: &AzureFrontDoorRuleUrlRedirectAction{
						RedirectType:        AzureFrontDoorRuleRedirectType_MOVED,
						DestinationHostname: "www.example.com",
						DestinationPath:     "landing",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a redirect query string starting with '?'", func() {
			rule := &AzureFrontDoorRule{
				Name:  "badredirectquery",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					UrlRedirect: &AzureFrontDoorRuleUrlRedirectAction{
						RedirectType:        AzureFrontDoorRuleRedirectType_MOVED,
						DestinationHostname: "www.example.com",
						QueryString:         "?src=redirect",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a redirect without a redirect type", func() {
			rule := &AzureFrontDoorRule{
				Name:  "notype",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					UrlRedirect: &AzureFrontDoorRuleUrlRedirectAction{
						DestinationHostname: "www.example.com",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an origin override without a forwarding protocol", func() {
			rule := &AzureFrontDoorRule{
				Name:  "originnoprotocol",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						OriginGroupId: literal(originGroupId),
						CacheBehavior: AzureFrontDoorRuleCacheBehavior_DISABLED,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a forwarding protocol without an origin override", func() {
			rule := &AzureFrontDoorRule{
				Name:  "protocolnoorigin",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						ForwardingProtocol: AzureFrontDoorRuleForwardingProtocol_HTTPS_ONLY,
						CacheBehavior:      AzureFrontDoorRuleCacheBehavior_DISABLED,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an override without a cache behavior", func() {
			rule := &AzureFrontDoorRule{
				Name:  "nobehavior",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a DISABLED override carrying a cache duration", func() {
			rule := &AzureFrontDoorRule{
				Name:  "disabledwithduration",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior: AzureFrontDoorRuleCacheBehavior_DISABLED,
						CacheDuration: "00:05:00",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an OVERRIDE_ALWAYS override without a duration", func() {
			rule := &AzureFrontDoorRule{
				Name:  "overridenoduration",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior:              AzureFrontDoorRuleCacheBehavior_OVERRIDE_ALWAYS,
						QueryStringCachingBehavior: AzureFrontDoorRuleQueryStringCachingBehavior_IGNORE_QUERY_STRING,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an HONOR_ORIGIN override carrying a duration", func() {
			rule := &AzureFrontDoorRule{
				Name:  "honorwithduration",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior:              AzureFrontDoorRuleCacheBehavior_HONOR_ORIGIN,
						CacheDuration:              "00:05:00",
						QueryStringCachingBehavior: AzureFrontDoorRuleQueryStringCachingBehavior_IGNORE_QUERY_STRING,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a caching override without a query string caching behavior", func() {
			rule := &AzureFrontDoorRule{
				Name:  "noqscb",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior: AzureFrontDoorRuleCacheBehavior_OVERRIDE_ALWAYS,
						CacheDuration: "00:05:00",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject INCLUDE_SPECIFIED_QUERY_STRINGS without parameters", func() {
			rule := &AzureFrontDoorRule{
				Name:  "includenospecified",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior:              AzureFrontDoorRuleCacheBehavior_OVERRIDE_ALWAYS,
						CacheDuration:              "00:05:00",
						QueryStringCachingBehavior: AzureFrontDoorRuleQueryStringCachingBehavior_INCLUDE_SPECIFIED_QUERY_STRINGS,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject IGNORE_QUERY_STRING with parameters", func() {
			rule := &AzureFrontDoorRule{
				Name:  "ignorewithparams",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior:              AzureFrontDoorRuleCacheBehavior_OVERRIDE_ALWAYS,
						CacheDuration:              "00:05:00",
						QueryStringCachingBehavior: AzureFrontDoorRuleQueryStringCachingBehavior_IGNORE_QUERY_STRING,
						QueryStringParameters:      []string{"page"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a cache duration starting with '0.'", func() {
			rule := &AzureFrontDoorRule{
				Name:  "zerodayduration",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior:              AzureFrontDoorRuleCacheBehavior_OVERRIDE_ALWAYS,
						CacheDuration:              "0.12:00:00",
						QueryStringCachingBehavior: AzureFrontDoorRuleQueryStringCachingBehavior_IGNORE_QUERY_STRING,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed cache duration", func() {
			rule := &AzureFrontDoorRule{
				Name:  "badduration",
				Order: 1,
				Actions: &AzureFrontDoorRuleActions{
					RouteConfigurationOverride: &AzureFrontDoorRuleRouteConfigurationOverrideAction{
						CacheBehavior:              AzureFrontDoorRuleCacheBehavior_OVERRIDE_ALWAYS,
						CacheDuration:              "12h",
						QueryStringCachingBehavior: AzureFrontDoorRuleQueryStringCachingBehavior_IGNORE_QUERY_STRING,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty request body condition", func() {
			rule := headerRule("emptybody", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				RequestBody: []*AzureFrontDoorRuleRequestBodyCondition{{
					Operator: AzureFrontDoorRuleOperator_CONTAINS,
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a post args condition without a name", func() {
			rule := headerRule("nopostargsname", 1)
			rule.Conditions = &AzureFrontDoorRuleConditions{
				PostArgs: []*AzureFrontDoorRulePostArgsCondition{{
					Operator:    AzureFrontDoorRuleOperator_EQUAL,
					MatchValues: []string{"v"},
				}},
			}
			gomega.Expect(protovalidate.Validate(withRules(rule))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong kind", func() {
			input := minimalSpec()
			input.Kind = "AzureFrontDoorRuleSets"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing metadata", func() {
			input := minimalSpec()
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})

func strPtr(s string) *string { return &s }

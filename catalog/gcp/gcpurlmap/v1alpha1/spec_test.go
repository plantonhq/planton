package gcpurlmapv1alpha1

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

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpUrlMapSpec Suite")
}

func backendSelfLink() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: "https://www.googleapis.com/compute/v1/projects/p/global/backendServices/web",
		},
	}
}

var _ = ginkgo.Describe("GcpUrlMapSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpUrlMap {
		return &GcpUrlMap{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpUrlMap",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-url-map",
			},
			Spec: &GcpUrlMapSpec{
				DefaultService: backendSelfLink(),
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec with default_service", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept default_url_redirect for apex-to-www", func() {
		target := minimal()
		target.Spec.DefaultService = nil
		target.Spec.DefaultUrlRedirect = &GcpUrlMapUrlRedirect{
			HostRedirect:  "www.example.com",
			HttpsRedirect: true,
			StripQuery:    false,
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept default_route_action with weighted backends", func() {
		target := minimal()
		target.Spec.DefaultService = nil
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			WeightedBackendServices: []*GcpUrlMapWeightedBackendService{
				{BackendService: backendSelfLink(), Weight: 900},
				{BackendService: backendSelfLink(), Weight: 100},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept host rules and path_rules fan-out", func() {
		target := minimal()
		target.Spec.HostRules = []*GcpUrlMapHostRule{
			{Hosts: []string{"www.example.com"}, PathMatcher: "main"},
		}
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name:           "main",
				DefaultService: backendSelfLink(),
				PathRules: []*GcpUrlMapPathRule{
					{Paths: []string{"/api/*"}, Service: backendSelfLink()},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept route_rules with rich match rules", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "api",
				RouteRules: []*GcpUrlMapRouteRule{
					{
						Priority: 1,
						Service:  backendSelfLink(),
						MatchRules: []*GcpUrlMapRouteRuleMatchRule{
							{
								PrefixMatch: "/v1",
								HeaderMatches: []*GcpUrlMapHeaderMatch{
									{HeaderName: "X-Api-Version", ExactMatch: "1"},
								},
							},
						},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept path_template_rewrite inside a route rule route_action", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "tmpl",
				RouteRules: []*GcpUrlMapRouteRule{
					{
						Priority: 1,
						RouteAction: &GcpUrlMapRouteAction{
							WeightedBackendServices: []*GcpUrlMapWeightedBackendService{
								{BackendService: backendSelfLink(), Weight: 1000},
							},
							UrlRewrite: &GcpUrlMapUrlRewrite{
								PathTemplateRewrite: "/v2/{country}",
							},
						},
						MatchRules: []*GcpUrlMapRouteRuleMatchRule{
							{PathTemplateMatch: "/v1/{country}/**"},
						},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept service plus a sub-policy-only route_action on a path rule", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "resilient",
				PathRules: []*GcpUrlMapPathRule{
					{
						Paths:   []string{"/api/*"},
						Service: backendSelfLink(),
						RouteAction: &GcpUrlMapRouteAction{
							Timeout: &GcpUrlMapDuration{Seconds: proto.Int64(30)},
							RetryPolicy: &GcpUrlMapRetryPolicy{
								NumRetries:      3,
								RetryConditions: []string{"5xx", "connect-failure"},
								PerTryTimeout:   &GcpUrlMapDuration{Seconds: proto.Int64(5)},
							},
						},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept default_service plus a sub-policy-only default_route_action", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			CorsPolicy: &GcpUrlMapCorsPolicy{
				AllowOrigins: []string{"https://app.example.com"},
				AllowMethods: []string{"GET", "POST"},
				MaxAge:       3600,
			},
			MaxStreamDuration: &GcpUrlMapDuration{Seconds: proto.Int64(60)},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the full traffic-management surface on a route rule", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "full",
				RouteRules: []*GcpUrlMapRouteRule{
					{
						Priority: 1,
						RouteAction: &GcpUrlMapRouteAction{
							WeightedBackendServices: []*GcpUrlMapWeightedBackendService{
								{
									BackendService: backendSelfLink(),
									Weight:         900,
									HeaderAction: &GcpUrlMapHeaderAction{
										ResponseHeadersToAdd: []*GcpUrlMapHeaderValue{
											{HeaderName: "X-Arm", HeaderValue: "stable", Replace: true},
										},
									},
								},
								{BackendService: backendSelfLink(), Weight: 100},
							},
							Timeout: &GcpUrlMapDuration{Seconds: proto.Int64(15)},
							RequestMirrorPolicy: &GcpUrlMapRequestMirrorPolicy{
								BackendService: backendSelfLink(),
							},
							FaultInjectionPolicy: &GcpUrlMapFaultInjectionPolicy{
								Abort: &GcpUrlMapFaultAbort{HttpStatus: 503, Percentage: 0.5},
								Delay: &GcpUrlMapFaultDelay{
									FixedDelay: &GcpUrlMapDuration{Nanos: proto.Int32(250000000)},
									Percentage: 1,
								},
							},
						},
						MatchRules: []*GcpUrlMapRouteRuleMatchRule{{PrefixMatch: "/"}},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a CDN cache policy with key policy and negative caching", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "cdn",
				PathRules: []*GcpUrlMapPathRule{
					{
						Paths:   []string{"/static/*"},
						Service: backendSelfLink(),
						RouteAction: &GcpUrlMapRouteAction{
							CachePolicy: &GcpUrlMapCachePolicy{
								CacheMode:       "CACHE_ALL_STATIC",
								DefaultTtl:      &GcpUrlMapDuration{Seconds: proto.Int64(3600)},
								MaxTtl:          &GcpUrlMapDuration{Seconds: proto.Int64(86400)},
								ClientTtl:       &GcpUrlMapDuration{Seconds: proto.Int64(600)},
								ServeWhileStale: &GcpUrlMapDuration{Seconds: proto.Int64(86400)},
								NegativeCaching: true,
								NegativeCachingPolicy: []*GcpUrlMapNegativeCachingPolicy{
									{Code: 404, Ttl: &GcpUrlMapDuration{Seconds: proto.Int64(120)}},
								},
								CacheKeyPolicy: &GcpUrlMapCacheKeyPolicy{
									IncludeHost:             true,
									IncludeProtocol:         true,
									IncludedQueryParameters: []string{"version"},
								},
							},
						},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept deletion_policy PREVENT", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept header_action at the URL map level", func() {
		target := minimal()
		target.Spec.HeaderAction = &GcpUrlMapHeaderAction{
			RequestHeadersToAdd: []*GcpUrlMapHeaderValue{
				{HeaderName: "X-Edge", HeaderValue: "global", Replace: true},
			},
			RequestHeadersToRemove: []string{"X-Internal"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept custom error response policy with override code", func() {
		target := minimal()
		target.Spec.DefaultCustomErrorResponsePolicy = &GcpUrlMapCustomErrorResponsePolicy{
			ErrorService: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "https://www.googleapis.com/compute/v1/projects/p/global/backendBuckets/errors",
				},
			},
			ErrorResponseRules: []*GcpUrlMapCustomErrorResponseRule{
				{MatchResponseCodes: []string{"404", "5xx"}, Path: "/errors/404.html", OverrideResponseCode: 200},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept routing self-tests", func() {
		target := minimal()
		target.Spec.Tests = []*GcpUrlMapTest{
			{
				Host:        "www.example.com",
				Path:        "/",
				Service:     backendSelfLink(),
				Description: "home resolves to web backend",
				Headers:     []*GcpUrlMapTestHeader{{Name: "Accept", Value: "text/html"}},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept metadata filters on route rule match rules", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "mesh",
				RouteRules: []*GcpUrlMapRouteRule{
					{
						Priority: 10,
						Service:  backendSelfLink(),
						MatchRules: []*GcpUrlMapRouteRuleMatchRule{
							{
								PrefixMatch: "/",
								MetadataFilters: []*GcpUrlMapMetadataFilter{
									{
										FilterMatchCriteria: "MATCH_ALL",
										FilterLabels: []*GcpUrlMapMetadataFilterLabel{
											{Name: "region", Value: "us"},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a path matcher with default_url_redirect only", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "redirect-matcher",
				DefaultUrlRedirect: &GcpUrlMapUrlRedirect{
					PrefixRedirect: "/landing",
					StripQuery:     true,
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept path_rule route_action with weighted backends", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "canary",
				PathRules: []*GcpUrlMapPathRule{
					{
						Paths: []string{"/release/*"},
						RouteAction: &GcpUrlMapRouteAction{
							WeightedBackendServices: []*GcpUrlMapWeightedBackendService{
								{BackendService: backendSelfLink(), Weight: 500},
								{BackendService: backendSelfLink(), Weight: 500},
							},
						},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a spec with no default target", func() {
		target := minimal()
		target.Spec.DefaultService = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "exactly one default target")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject both default_service and default_url_redirect", func() {
		target := minimal()
		target.Spec.DefaultUrlRedirect = &GcpUrlMapUrlRedirect{HttpsRedirect: true, StripQuery: false}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject both default_route_action and default_url_redirect", func() {
		target := minimal()
		target.Spec.DefaultService = nil
		target.Spec.DefaultUrlRedirect = &GcpUrlMapUrlRedirect{StripQuery: false}
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			WeightedBackendServices: []*GcpUrlMapWeightedBackendService{
				{BackendService: backendSelfLink(), Weight: 100},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "mutually exclusive")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject default_route_action without weighted backends at top level", func() {
		target := minimal()
		target.Spec.DefaultService = nil
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			UrlRewrite: &GcpUrlMapUrlRewrite{HostRewrite: "internal.example.com"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid url_map_name", func() {
		target := minimal()
		target.Spec.UrlMapName = "Invalid_Name"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "RFC1035")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject path_redirect and prefix_redirect together", func() {
		target := minimal()
		target.Spec.DefaultService = nil
		target.Spec.DefaultUrlRedirect = &GcpUrlMapUrlRedirect{
			PathRedirect:   "/new",
			PrefixRedirect: "/old",
			StripQuery:     false,
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "mutually exclusive")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid redirect_response_code", func() {
		target := minimal()
		target.Spec.DefaultService = nil
		target.Spec.DefaultUrlRedirect = &GcpUrlMapUrlRedirect{
			RedirectResponseCode: "INVALID",
			StripQuery:           false,
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a path matcher with both path_rules and route_rules", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "mixed",
				PathRules: []*GcpUrlMapPathRule{
					{Paths: []string{"/a/*"}, Service: backendSelfLink()},
				},
				RouteRules: []*GcpUrlMapRouteRule{
					{Priority: 1, Service: backendSelfLink(), MatchRules: []*GcpUrlMapRouteRuleMatchRule{{PrefixMatch: "/b"}}},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "either path_rules or route_rules")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a path rule with no target", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "bad-rule",
				PathRules: []*GcpUrlMapPathRule{
					{Paths: []string{"/orphan/*"}},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a path rule with multiple targets", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "multi",
				PathRules: []*GcpUrlMapPathRule{
					{
						Paths:       []string{"/both/*"},
						Service:     backendSelfLink(),
						UrlRedirect: &GcpUrlMapUrlRedirect{StripQuery: false},
					},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a route rule with no target", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "empty-route",
				RouteRules: []*GcpUrlMapRouteRule{
					{Priority: 1, MatchRules: []*GcpUrlMapRouteRuleMatchRule{{PrefixMatch: "/x"}}},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject multiple path matchers on one match rule", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "bad-match",
				RouteRules: []*GcpUrlMapRouteRule{
					{
						Priority: 1,
						Service:  backendSelfLink(),
						MatchRules: []*GcpUrlMapRouteRuleMatchRule{
							{PrefixMatch: "/a", FullPathMatch: "/b"},
						},
					},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject path_prefix_rewrite and path_template_rewrite together", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "bad-rewrite",
				RouteRules: []*GcpUrlMapRouteRule{
					{
						Priority: 1,
						RouteAction: &GcpUrlMapRouteAction{
							UrlRewrite: &GcpUrlMapUrlRewrite{
								PathPrefixRewrite:   "/v2",
								PathTemplateRewrite: "/v2/{id}",
							},
						},
						MatchRules: []*GcpUrlMapRouteRuleMatchRule{{PrefixMatch: "/v1"}},
					},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a path matcher with multiple default targets", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name:           "two-defaults",
				DefaultService: backendSelfLink(),
				DefaultUrlRedirect: &GcpUrlMapUrlRedirect{
					StripQuery: false,
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid filter_match_criteria", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "bad-filter",
				RouteRules: []*GcpUrlMapRouteRule{
					{
						Priority: 1,
						Service:  backendSelfLink(),
						MatchRules: []*GcpUrlMapRouteRuleMatchRule{
							{
								PrefixMatch: "/",
								MetadataFilters: []*GcpUrlMapMetadataFilter{
									{
										FilterMatchCriteria: "MATCH_SOME",
										FilterLabels:        []*GcpUrlMapMetadataFilterLabel{{Name: "k", Value: "v"}},
									},
								},
							},
						},
					},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject override_response_code outside 200-599", func() {
		target := minimal()
		target.Spec.DefaultCustomErrorResponsePolicy = &GcpUrlMapCustomErrorResponsePolicy{
			ErrorResponseRules: []*GcpUrlMapCustomErrorResponseRule{
				{MatchResponseCodes: []string{"404"}, Path: "/e.html", OverrideResponseCode: 199},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject weighted backend weight above 1000", func() {
		target := minimal()
		target.Spec.DefaultService = nil
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			WeightedBackendServices: []*GcpUrlMapWeightedBackendService{
				{BackendService: backendSelfLink(), Weight: 1001},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a route rule without match_rules", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "no-match",
				RouteRules: []*GcpUrlMapRouteRule{
					{Priority: 1, Service: backendSelfLink()},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "DELETE, PREVENT, ABANDON")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an unknown retry condition", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			RetryPolicy: &GcpUrlMapRetryPolicy{
				RetryConditions: []string{"5xx", "timeout"},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "retry condition")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid cache_mode", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			CachePolicy: &GcpUrlMapCachePolicy{CacheMode: "CACHE_EVERYTHING"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject TTLs with cache_mode USE_ORIGIN_HEADERS", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			CachePolicy: &GcpUrlMapCachePolicy{
				CacheMode:  "USE_ORIGIN_HEADERS",
				DefaultTtl: &GcpUrlMapDuration{Seconds: proto.Int64(3600)},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "USE_ORIGIN_HEADERS")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a negative caching code GCP does not accept", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			CachePolicy: &GcpUrlMapCachePolicy{
				NegativeCaching: true,
				NegativeCachingPolicy: []*GcpUrlMapNegativeCachingPolicy{
					{Code: 418, Ttl: &GcpUrlMapDuration{Seconds: proto.Int64(60)}},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject both included and excluded cache key query parameters", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			CachePolicy: &GcpUrlMapCachePolicy{
				CacheKeyPolicy: &GcpUrlMapCacheKeyPolicy{
					IncludedQueryParameters: []string{"a"},
					ExcludedQueryParameters: []string{"b"},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject fault abort percentage above 100", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			FaultInjectionPolicy: &GcpUrlMapFaultInjectionPolicy{
				Abort: &GcpUrlMapFaultAbort{HttpStatus: 503, Percentage: 150},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject fault abort http_status below 200", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			FaultInjectionPolicy: &GcpUrlMapFaultInjectionPolicy{
				Abort: &GcpUrlMapFaultAbort{HttpStatus: 199, Percentage: 1},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an explicit zero duration (a TTL of 0s is a real setting)", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			Timeout: &GcpUrlMapDuration{Seconds: proto.Int64(0), Nanos: proto.Int32(0)},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject duration nanos above the nanosecond bound", func() {
		target := minimal()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			Timeout: &GcpUrlMapDuration{Seconds: proto.Int64(1), Nanos: proto.Int32(1000000000)},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject path_template_rewrite in the top-level default route action", func() {
		target := minimal()
		target.Spec.DefaultService = nil
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			WeightedBackendServices: []*GcpUrlMapWeightedBackendService{
				{BackendService: backendSelfLink(), Weight: 1000},
			},
			UrlRewrite: &GcpUrlMapUrlRewrite{PathTemplateRewrite: "/v2/{id}"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "route rule")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject path_template_rewrite in a path rule route action", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "bad-tmpl",
				PathRules: []*GcpUrlMapPathRule{
					{
						Paths:   []string{"/x/*"},
						Service: backendSelfLink(),
						RouteAction: &GcpUrlMapRouteAction{
							UrlRewrite: &GcpUrlMapUrlRewrite{PathTemplateRewrite: "/y/{id}"},
						},
					},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject path_template_rewrite in a path matcher default route action", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name:           "bad-matcher-tmpl",
				DefaultService: backendSelfLink(),
				DefaultRouteAction: &GcpUrlMapRouteAction{
					UrlRewrite: &GcpUrlMapUrlRewrite{PathTemplateRewrite: "/v2/{id}"},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "route rule")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a route rule combining route_action and url_redirect", func() {
		target := minimal()
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name: "conflict",
				RouteRules: []*GcpUrlMapRouteRule{
					{
						Priority:    1,
						UrlRedirect: &GcpUrlMapUrlRedirect{HttpsRedirect: true, StripQuery: false},
						RouteAction: &GcpUrlMapRouteAction{
							Timeout: &GcpUrlMapDuration{Seconds: proto.Int64(5)},
						},
						MatchRules: []*GcpUrlMapRouteRuleMatchRule{{PrefixMatch: "/"}},
					},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Regional arm ────────────────

	regionalBackend := func() *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/backendServices/web",
			},
		}
	}

	regional := func() *GcpUrlMap {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.DefaultService = regionalBackend()
		return target
	}

	ginkgo.It("should accept a minimal regional URL map", func() {
		gomega.Expect(validator.Validate(regional())).To(gomega.Succeed())
	})

	ginkgo.It("should accept path_template_rewrite in a path matcher default route action on a regional map", func() {
		target := regional()
		target.Spec.HostRules = []*GcpUrlMapHostRule{{Hosts: []string{"api.example.com"}, PathMatcher: "api"}}
		target.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name:           "api",
				DefaultService: regionalBackend(),
				DefaultRouteAction: &GcpUrlMapRouteAction{
					UrlRewrite: &GcpUrlMapUrlRewrite{PathTemplateRewrite: "/v2/{id}"},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional routing test that names its service", func() {
		target := regional()
		target.Spec.Tests = []*GcpUrlMapTest{
			{Host: "api.example.com", Path: "/v1/users", Service: regionalBackend()},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed region", func() {
		target := regional()
		target.Spec.Region = "US-CENTRAL1"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject cache_policy anywhere on a regional map", func() {
		top := regional()
		top.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			CachePolicy: &GcpUrlMapCachePolicy{CacheMode: "CACHE_ALL_STATIC"},
		}
		err := validator.Validate(top)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "cache_policy (Cloud CDN route caching) exists only on a global")).To(gomega.BeTrue())

		nested := regional()
		nested.Spec.PathMatchers = []*GcpUrlMapPathMatcher{
			{
				Name:           "cached",
				DefaultService: regionalBackend(),
				RouteRules: []*GcpUrlMapRouteRule{
					{
						Priority:   1,
						Service:    regionalBackend(),
						MatchRules: []*GcpUrlMapRouteRuleMatchRule{{PrefixMatch: "/static"}},
						RouteAction: &GcpUrlMapRouteAction{
							CachePolicy: &GcpUrlMapCachePolicy{CacheMode: "FORCE_CACHE_ALL"},
						},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(nested)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject custom error response policies on a regional map", func() {
		target := regional()
		target.Spec.DefaultCustomErrorResponsePolicy = &GcpUrlMapCustomErrorResponsePolicy{
			ErrorResponseRules: []*GcpUrlMapCustomErrorResponseRule{
				{MatchResponseCodes: []string{"5xx"}, Path: "/errors/5xx.html"},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "custom error response policies exist only on a global")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject max_stream_duration on a regional map's default route action", func() {
		target := regional()
		target.Spec.DefaultRouteAction = &GcpUrlMapRouteAction{
			MaxStreamDuration: &GcpUrlMapDuration{Seconds: proto.Int64(60)},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "max_stream_duration exists only on a global")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a regional routing test with headers or without a service", func() {
		withHeaders := regional()
		withHeaders.Spec.Tests = []*GcpUrlMapTest{
			{
				Host:    "api.example.com",
				Path:    "/v1",
				Service: regionalBackend(),
				Headers: []*GcpUrlMapTestHeader{{Name: "X-Debug", Value: "1"}},
			},
		}
		err := validator.Validate(withHeaders)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "regional URL map every routing test")).To(gomega.BeTrue())

		redirectOnly := regional()
		redirectOnly.Spec.Tests = []*GcpUrlMapTest{
			{Host: "api.example.com", Path: "/old", ExpectedRedirectResponseCode: 301},
		}
		gomega.Expect(validator.Validate(redirectOnly)).ToNot(gomega.Succeed())
	})
})

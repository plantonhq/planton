package cloudflarerulesetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func boolPtr(b bool) *bool { return &b }
func rulesetKindPtr(k CloudflareRulesetSpec_RulesetKind) *CloudflareRulesetSpec_RulesetKind {
	return &k
}

func TestCloudflareRulesetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareRulesetSpec Validation Suite")
}

func validResource() *CloudflareRuleset {
	return &CloudflareRuleset{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareRuleset",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-origin-rule"},
		Spec: &CloudflareRulesetSpec{
			ZoneId:      &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "zone-abc123"}},
			RulesetKind: rulesetKindPtr(CloudflareRulesetSpec_zone),
			Phase:       CloudflareRulesetSpec_http_request_origin,
			Name:        "Route app traffic to K8s",
			Rules: []*CloudflareRulesetRule{
				{
					Expression: "not http.request.uri.path starts_with \"/docs\"",
					Action:     CloudflareRulesetRule_route,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						HostHeader: "planton.ai",
						Origin:     &CloudflareRulesetOrigin{Host: "k8s-lb.example.com", Port: 443},
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("CloudflareRulesetSpec Validation", func() {

	// ---- Positive cases ----

	ginkgo.Describe("Valid inputs", func() {

		ginkgo.It("accepts a minimal origin rule with zone_id", func() {
			gomega.Expect(protovalidate.Validate(validResource())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a ruleset with account_id instead of zone_id", func() {
			r := validResource()
			r.Spec.ZoneId = nil
			r.Spec.AccountId = "acct-xyz789"
			r.Spec.RulesetKind = rulesetKindPtr(CloudflareRulesetSpec_root)
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a block rule", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_request_firewall_custom
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "ip.src eq 192.0.2.1",
					Action:     CloudflareRulesetRule_block,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						Response: &CloudflareRulesetResponse{
							StatusCode:  403,
							Content:     "Forbidden",
							ContentType: "text/plain",
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts multiple rules", func() {
			r := validResource()
			r.Spec.Rules = append(r.Spec.Rules, &CloudflareRulesetRule{
				Expression: "http.request.uri.path starts_with \"/api\"",
				Action:     CloudflareRulesetRule_route,
				Enabled:    boolPtr(true),
			})
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a skip rule with phases and products", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_request_firewall_custom
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "http.host eq \"example.com\"",
					Action:     CloudflareRulesetRule_skip,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						Phases:   []string{"http_request_firewall_managed"},
						Products: []string{"waf"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a rule without action_parameters", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_request_firewall_custom
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "true",
					Action:     CloudflareRulesetRule_challenge,
					Enabled:    boolPtr(true),
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a rate-limit rule", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_ratelimit
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "true",
					Action:     CloudflareRulesetRule_block,
					Enabled:    boolPtr(true),
					Ratelimit: &CloudflareRulesetRatelimit{
						Characteristics:   []string{"ip.src", "cf.colo.id"},
						Period:            60,
						RequestsPerPeriod: 100,
						MitigationTimeout: 600,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a set_config rule with config settings", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_config_settings
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "true",
					Action:     CloudflareRulesetRule_set_config,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						Ssl:           "full",
						SecurityLevel: "high",
						Polish:        "lossless",
						RocketLoader:  boolPtr(true),
						Autominify:    &CloudflareRulesetAutominify{Css: boolPtr(true), Js: boolPtr(true)},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an advanced cache rule with custom cache key", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_request_cache_settings
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "true",
					Action:     CloudflareRulesetRule_set_cache_settings,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						Cache:   true,
						EdgeTtl: &CloudflareRulesetEdgeTtl{Mode: "override_origin", DefaultTtl: 7200},
						CacheKey: &CloudflareRulesetCacheKey{
							CacheByDeviceType: boolPtr(true),
							CustomKey: &CloudflareRulesetCacheKeyCustomKey{
								QueryString: &CloudflareRulesetCacheKeyQueryString{
									Include: &CloudflareRulesetCacheKeyQueryStringFilter{All: boolPtr(true)},
								},
							},
						},
						CacheReserve: &CloudflareRulesetCacheReserve{Eligible: true, MinimumFileSize: 1024},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a set_cache_control rule with directives", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_response_headers_transform
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "true",
					Action:     CloudflareRulesetRule_set_cache_control,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						MaxAge:  &CloudflareRulesetCacheControlValue{Operation: "set", Value: 3600},
						Private: &CloudflareRulesetCacheControlQualifiers{Operation: "set", Qualifiers: []string{"Set-Cookie"}},
						NoStore: &CloudflareRulesetCacheControlFlag{Operation: "set"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a log_custom_field rule", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_log_custom_fields
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "true",
					Action:     CloudflareRulesetRule_log_custom_field,
					Enabled:    boolPtr(true),
					ActionParameters: &CloudflareRulesetActionParameters{
						RequestFields:  []*CloudflareRulesetLogField{{Name: "cf-ray"}},
						ResponseFields: []*CloudflareRulesetLogResponseField{{Name: "x-trace-id", PreserveDuplicates: true}},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an exposed-credential-check rule", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_request_firewall_custom
			r.Spec.Rules = []*CloudflareRulesetRule{
				{
					Expression: "http.request.method eq \"POST\"",
					Action:     CloudflareRulesetRule_managed_challenge,
					Enabled:    boolPtr(true),
					ExposedCredentialCheck: &CloudflareRulesetExposedCredentialCheck{
						UsernameExpression: "url_decode(http.request.body.form[\"user\"][0])",
						PasswordExpression: "url_decode(http.request.body.form[\"pass\"][0])",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})
	})

	// ---- Negative cases ----

	ginkgo.Describe("Invalid inputs", func() {

		ginkgo.It("rejects when neither zone_id nor account_id is set", func() {
			r := validResource()
			r.Spec.ZoneId = nil
			r.Spec.AccountId = ""
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("zone_id or account_id"))
		})

		ginkgo.It("rejects when both zone_id and account_id are set", func() {
			r := validResource()
			r.Spec.AccountId = "acct-xyz789"
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("zone_id or account_id"))
		})

		ginkgo.It("rejects unspecified phase", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_phase_unspecified
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("phase"))
		})

		ginkgo.It("rejects empty name", func() {
			r := validResource()
			r.Spec.Name = ""
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects empty rules list", func() {
			r := validResource()
			r.Spec.Rules = nil
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a rule with empty expression", func() {
			r := validResource()
			r.Spec.Rules[0].Expression = ""
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a rule with unspecified action", func() {
			r := validResource()
			r.Spec.Rules[0].Action = CloudflareRulesetRule_action_unspecified
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("action"))
		})

		ginkgo.It("rejects an invalid ssl mode", func() {
			r := validResource()
			r.Spec.Rules[0].ActionParameters = &CloudflareRulesetActionParameters{Ssl: "totally-secure"}
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("ssl"))
		})

		ginkgo.It("rejects an invalid security_level", func() {
			r := validResource()
			r.Spec.Rules[0].ActionParameters = &CloudflareRulesetActionParameters{SecurityLevel: "paranoid"}
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("security_level"))
		})

		ginkgo.It("rejects an invalid polish value", func() {
			r := validResource()
			r.Spec.Rules[0].ActionParameters = &CloudflareRulesetActionParameters{Polish: "max"}
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("polish"))
		})

		ginkgo.It("rejects an invalid content_type for serve_error", func() {
			r := validResource()
			r.Spec.Rules[0].ActionParameters = &CloudflareRulesetActionParameters{ContentType: "text/markdown"}
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("content_type"))
		})

		ginkgo.It("rejects an invalid set_cache_tags operation", func() {
			r := validResource()
			r.Spec.Rules[0].ActionParameters = &CloudflareRulesetActionParameters{Operation: "append"}
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("operation"))
		})

		ginkgo.It("rejects an invalid edge_ttl mode", func() {
			r := validResource()
			r.Spec.Rules[0].ActionParameters = &CloudflareRulesetActionParameters{
				EdgeTtl: &CloudflareRulesetEdgeTtl{Mode: "always_cache"},
			}
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("mode"))
		})

		ginkgo.It("rejects an invalid header operation", func() {
			r := validResource()
			r.Spec.Rules[0].ActionParameters = &CloudflareRulesetActionParameters{
				Headers: map[string]*CloudflareRulesetHeader{
					"X-Foo": {Operation: "upsert", Value: "bar"},
				},
			}
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("operation"))
		})

		ginkgo.It("rejects an invalid overrides sensitivity_level", func() {
			r := validResource()
			r.Spec.Rules[0].ActionParameters = &CloudflareRulesetActionParameters{
				Overrides: &CloudflareRulesetOverrides{SensitivityLevel: "extreme"},
			}
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("sensitivity_level"))
		})

		ginkgo.It("rejects a ratelimit with empty characteristics", func() {
			r := validResource()
			r.Spec.Phase = CloudflareRulesetSpec_http_ratelimit
			r.Spec.Rules[0].Action = CloudflareRulesetRule_block
			r.Spec.Rules[0].ActionParameters = nil
			r.Spec.Rules[0].Ratelimit = &CloudflareRulesetRatelimit{Characteristics: nil, Period: 60}
			err := protovalidate.Validate(r)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})

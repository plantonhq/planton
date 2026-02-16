package gcpcloudarmorpolicyv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudArmorPolicySpec Suite")
}

var _ = ginkgo.Describe("GcpCloudArmorPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCloudArmorPolicy {
		return &GcpCloudArmorPolicy{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpCloudArmorPolicy",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-policy",
			},
			Spec: &GcpCloudArmorPolicySpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Rules: []*GcpCloudArmorRule{
					{
						Action:   "allow",
						Priority: 2147483647,
						Match: &GcpCloudArmorRuleMatch{
							VersionedExpr: "SRC_IPS_V1",
							SrcIpRanges:   []string{"*"},
						},
					},
				},
			},
		}
	}

	ipAllowRule := func(priority int32, ranges []string) *GcpCloudArmorRule {
		return &GcpCloudArmorRule{
			Action:   "allow",
			Priority: priority,
			Match: &GcpCloudArmorRuleMatch{
				VersionedExpr: "SRC_IPS_V1",
				SrcIpRanges:   ranges,
			},
		}
	}

	ipDenyRule := func(priority int32, ranges []string) *GcpCloudArmorRule {
		return &GcpCloudArmorRule{
			Action:   "deny(403)",
			Priority: priority,
			Match: &GcpCloudArmorRuleMatch{
				VersionedExpr: "SRC_IPS_V1",
				SrcIpRanges:   ranges,
			},
		}
	}

	celRule := func(priority int32, expr string, action string) *GcpCloudArmorRule {
		return &GcpCloudArmorRule{
			Action:   action,
			Priority: priority,
			Match: &GcpCloudArmorRuleMatch{
				Expression: expr,
			},
		}
	}

	// Suppress unused variable warnings.
	_ = ipAllowRule
	_ = ipDenyRule
	_ = celRule

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec with default allow rule", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with policy_name", func() {
		msg := minimal()
		msg.Spec.PolicyName = "my-security-policy"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with description", func() {
		msg := minimal()
		msg.Spec.Description = "WAF policy protecting production APIs"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with CLOUD_ARMOR type", func() {
		msg := minimal()
		msg.Spec.Type = "CLOUD_ARMOR"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with CLOUD_ARMOR_EDGE type", func() {
		msg := minimal()
		msg.Spec.Type = "CLOUD_ARMOR_EDGE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with CLOUD_ARMOR_INTERNAL_SERVICE type", func() {
		msg := minimal()
		msg.Spec.Type = "CLOUD_ARMOR_INTERNAL_SERVICE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept IP allowlist with multiple CIDR ranges", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			ipAllowRule(1000, []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}),
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept IP denylist rule", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			ipDenyRule(1000, []string{"203.0.113.0/24"}),
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept CEL expression rule for geo-blocking", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			celRule(1000, "origin.region_code == 'CN'", "deny(403)"),
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept throttle rule with rate limit options", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "throttle",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
					SrcIpRanges:   []string{"*"},
				},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction: "allow",
					ExceedAction:  "deny(429)",
					EnforceOnKey:  "IP",
					RateLimitThreshold: &GcpCloudArmorRateThreshold{
						Count:       100,
						IntervalSec: 60,
					},
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept rate_based_ban rule with ban threshold", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "rate_based_ban",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
					SrcIpRanges:   []string{"*"},
				},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction: "allow",
					ExceedAction:  "deny(403)",
					EnforceOnKey:  "IP",
					RateLimitThreshold: &GcpCloudArmorRateThreshold{
						Count:       500,
						IntervalSec: 300,
					},
					BanThreshold: &GcpCloudArmorRateThreshold{
						Count:       1000,
						IntervalSec: 600,
					},
					BanDurationSec: 3600,
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept redirect rule with GOOGLE_RECAPTCHA", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "redirect",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					Expression: "origin.region_code != 'US'",
				},
				RedirectOptions: &GcpCloudArmorRedirectConfig{
					Type: "GOOGLE_RECAPTCHA",
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept redirect rule with EXTERNAL_302 and target", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "redirect",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
					SrcIpRanges:   []string{"*"},
				},
				RedirectOptions: &GcpCloudArmorRedirectConfig{
					Type:   "EXTERNAL_302",
					Target: "https://sorry.example.com/blocked",
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept rule with preview mode enabled", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "deny(403)",
				Priority: 1000,
				Preview:  true,
				Match: &GcpCloudArmorRuleMatch{
					Expression: "request.path.matches('/admin/.*')",
				},
				Description: "Block admin paths (preview)",
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept rule with header action", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "allow",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
					SrcIpRanges:   []string{"10.0.0.0/8"},
				},
				HeaderAction: &GcpCloudArmorHeaderAction{
					RequestHeadersToAdds: []*GcpCloudArmorRequestHeader{
						{HeaderName: "X-Cloud-Armor-Verified", HeaderValue: "true"},
						{HeaderName: "X-Source-Network", HeaderValue: "internal"},
					},
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept rule with preconfigured WAF exclusion", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "deny(403)",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					Expression: "evaluatePreconfiguredWaf('sqli-v33-stable')",
				},
				PreconfiguredWafConfig: &GcpCloudArmorPreconfiguredWafConfig{
					Exclusions: []*GcpCloudArmorWafExclusion{
						{
							TargetRuleSet: "sqli-v33-stable",
							RequestUris: []*GcpCloudArmorWafExclusionFieldParams{
								{Operator: "STARTS_WITH", Value: "/api/search"},
							},
						},
					},
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept adaptive protection with layer 7 DDoS defense", func() {
		msg := minimal()
		msg.Spec.AdaptiveProtectionConfig = &GcpCloudArmorAdaptiveProtectionConfig{
			EnableLayer_7DdosDefense: true,
			RuleVisibility:           "STANDARD",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept adaptive protection with PREMIUM visibility", func() {
		msg := minimal()
		msg.Spec.AdaptiveProtectionConfig = &GcpCloudArmorAdaptiveProtectionConfig{
			EnableLayer_7DdosDefense: true,
			RuleVisibility:           "PREMIUM",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept advanced options with JSON parsing and verbose logging", func() {
		msg := minimal()
		msg.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{
			JsonParsing:                "STANDARD",
			LogLevel:                   "VERBOSE",
			UserIpRequestHeaders:       []string{"X-Real-IP", "X-Forwarded-For"},
			RequestBodyInspectionSize: "32KB",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept advanced options with GraphQL parsing", func() {
		msg := minimal()
		msg.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{
			JsonParsing: "STANDARD_WITH_GRAPHQL",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept rate limit with HTTP_HEADER enforce key", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "throttle",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
					SrcIpRanges:   []string{"*"},
				},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction:    "allow",
					ExceedAction:     "deny(429)",
					EnforceOnKey:     "HTTP_HEADER",
					EnforceOnKeyName: "X-API-Key",
					RateLimitThreshold: &GcpCloudArmorRateThreshold{
						Count:       50,
						IntervalSec: 60,
					},
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept rate limit with exceed redirect to reCAPTCHA", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "throttle",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
					SrcIpRanges:   []string{"*"},
				},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction: "allow",
					ExceedAction:  "redirect",
					EnforceOnKey:  "IP",
					RateLimitThreshold: &GcpCloudArmorRateThreshold{
						Count:       200,
						IntervalSec: 120,
					},
					ExceedRedirectOptions: &GcpCloudArmorRedirectConfig{
						Type: "GOOGLE_RECAPTCHA",
					},
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept deny(404) action", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "deny(404)",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
					SrcIpRanges:   []string{"192.0.2.0/24"},
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept deny(502) action", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "deny(502)",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
					SrcIpRanges:   []string{"198.51.100.0/24"},
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept WAF exclusion with EQUALS_ANY operator (no value)", func() {
		msg := minimal()
		msg.Spec.Rules = append([]*GcpCloudArmorRule{
			{
				Action:   "deny(403)",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					Expression: "evaluatePreconfiguredWaf('xss-v33-stable')",
				},
				PreconfiguredWafConfig: &GcpCloudArmorPreconfiguredWafConfig{
					Exclusions: []*GcpCloudArmorWafExclusion{
						{
							TargetRuleSet:  "xss-v33-stable",
							TargetRuleIds:  []string{"941100", "941110"},
							RequestHeaders: []*GcpCloudArmorWafExclusionFieldParams{
								{Operator: "EQUALS_ANY"},
							},
						},
					},
				},
			},
		}, msg.Spec.Rules...)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept full-featured spec with multiple rules and all options", func() {
		msg := minimal()
		msg.Spec.PolicyName = "production-waf"
		msg.Spec.Description = "Production WAF policy with rate limiting and DDoS protection"
		msg.Spec.Type = "CLOUD_ARMOR"
		msg.Spec.AdaptiveProtectionConfig = &GcpCloudArmorAdaptiveProtectionConfig{
			EnableLayer_7DdosDefense: true,
			RuleVisibility:           "STANDARD",
		}
		msg.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{
			JsonParsing:                "STANDARD",
			LogLevel:                   "VERBOSE",
			RequestBodyInspectionSize: "32KB",
		}
		msg.Spec.Rules = []*GcpCloudArmorRule{
			ipAllowRule(100, []string{"10.0.0.0/8"}),
			celRule(200, "origin.region_code == 'CN' || origin.region_code == 'RU'", "deny(403)"),
			{
				Action:   "throttle",
				Priority: 300,
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction:  "allow",
					ExceedAction:   "deny(429)",
					EnforceOnKey:   "IP",
					RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
				},
			},
			{
				Action:   "deny(403)",
				Priority: 2147483647,
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty rules (IaC modules will add default)", func() {
		msg := minimal()
		msg.Spec.Rules = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid policy_name (uppercase)", func() {
		msg := minimal()
		msg.Spec.PolicyName = "MyPolicy"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid policy_name (starts with digit)", func() {
		msg := minimal()
		msg.Spec.PolicyName = "1-policy"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid policy_name (ends with hyphen)", func() {
		msg := minimal()
		msg.Spec.PolicyName = "my-policy-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid type", func() {
		msg := minimal()
		msg.Spec.Type = "CLOUD_ARMOR_NETWORK"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject rule with missing action", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Priority: 1000,
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject rule with invalid action", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "block",
				Priority: 1000,
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject rule with missing match", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "allow",
				Priority: 1000,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject match with both versioned_expr and expression", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "allow",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
					SrcIpRanges:   []string{"10.0.0.0/8"},
					Expression:    "origin.region_code == 'US'",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject match with neither versioned_expr nor expression", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "allow",
				Priority: 1000,
				Match:    &GcpCloudArmorRuleMatch{},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject versioned_expr without src_ip_ranges", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "allow",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V1",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid versioned_expr value", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "allow",
				Priority: 1000,
				Match: &GcpCloudArmorRuleMatch{
					VersionedExpr: "SRC_IPS_V2",
					SrcIpRanges:   []string{"*"},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject rule description exceeding 64 characters", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:      "allow",
				Priority:    1000,
				Description: "This description is intentionally very long to exceed the sixty-four character maximum allowed by GCP",
				Match:       &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid redirect type", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "redirect",
				Priority: 1000,
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RedirectOptions: &GcpCloudArmorRedirectConfig{
					Type: "INTERNAL_301",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid enforce_on_key", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "throttle",
				Priority: 1000,
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction:  "allow",
					ExceedAction:   "deny(429)",
					EnforceOnKey:   "INVALID_KEY",
					RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid conform_action", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "throttle",
				Priority: 1000,
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction:  "deny",
					ExceedAction:   "deny(429)",
					RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid exceed_action", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "throttle",
				Priority: 1000,
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction:  "allow",
					ExceedAction:   "deny(500)",
					RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject rate_limit_options with missing threshold", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "throttle",
				Priority: 1000,
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction: "allow",
					ExceedAction:  "deny(429)",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid json_parsing value", func() {
		msg := minimal()
		msg.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{
			JsonParsing: "FULL",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid log_level value", func() {
		msg := minimal()
		msg.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{
			LogLevel: "DEBUG",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid request_body_inspection_size", func() {
		msg := minimal()
		msg.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{
			RequestBodyInspectionSize: "128KB",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid rule_visibility value", func() {
		msg := minimal()
		msg.Spec.AdaptiveProtectionConfig = &GcpCloudArmorAdaptiveProtectionConfig{
			EnableLayer_7DdosDefense: true,
			RuleVisibility:           "ENTERPRISE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid WAF exclusion operator", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "deny(403)",
				Priority: 1000,
				Match:    &GcpCloudArmorRuleMatch{Expression: "evaluatePreconfiguredWaf('sqli-v33-stable')"},
				PreconfiguredWafConfig: &GcpCloudArmorPreconfiguredWafConfig{
					Exclusions: []*GcpCloudArmorWafExclusion{
						{
							TargetRuleSet: "sqli-v33-stable",
							RequestUris: []*GcpCloudArmorWafExclusionFieldParams{
								{Operator: "REGEX", Value: "/api/.*"},
							},
						},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject description exceeding 2048 characters", func() {
		msg := minimal()
		longDesc := make([]byte, 2049)
		for i := range longDesc {
			longDesc[i] = 'a'
		}
		msg.Spec.Description = string(longDesc)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})

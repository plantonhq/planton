package gcpcloudarmorpolicyv1alpha1

import (
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
			ApiVersion: "gcp.planton.dev/v1alpha1",
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
						Priority: proto.Int32(2147483647),
						Match: &GcpCloudArmorRuleMatch{
							VersionedExpr: "SRC_IPS_V1",
							SrcIpRanges:   []string{"*"},
						},
					},
				},
			},
		}
	}

	// A regional CLOUD_ARMOR_NETWORK policy on the provider's own example
	// shape: STANDARD DDoS protection, one two-byte TCP field at offset 8, a
	// network-match rule on a source range and that field, and the default
	// rule matching every packet through an empty network_match.
	regionalNetwork := func() *GcpCloudArmorPolicy {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		msg.Spec.Type = "CLOUD_ARMOR_NETWORK"
		msg.Spec.DdosProtectionConfig = &GcpCloudArmorDdosProtectionConfig{DdosProtection: "STANDARD"}
		msg.Spec.UserDefinedFields = []*GcpCloudArmorUserDefinedField{
			{Name: "SIG1_AT_0", Base: "TCP", Offset: proto.Int32(8), Size: proto.Int32(2), Mask: "0x8F00"},
		}
		msg.Spec.Rules = []*GcpCloudArmorRule{
			{
				Action:   "allow",
				Priority: proto.Int32(100),
				Preview:  true,
				NetworkMatch: &GcpCloudArmorNetworkMatch{
					SrcIpRanges:       []string{"10.10.0.0/16"},
					SrcAsns:           []int64{15169},
					UserDefinedFields: []*GcpCloudArmorNetworkMatchUserDefinedField{{Name: "SIG1_AT_0", Values: []string{"0x8F00"}}},
				},
			},
			{
				Action:       "deny(403)",
				Priority:     proto.Int32(2147483647),
				NetworkMatch: &GcpCloudArmorNetworkMatch{},
			},
		}
		return msg
	}

	ipAllowRule := func(priority int32, ranges []string) *GcpCloudArmorRule {
		return &GcpCloudArmorRule{
			Action:   "allow",
			Priority: proto.Int32(priority),
			Match: &GcpCloudArmorRuleMatch{
				VersionedExpr: "SRC_IPS_V1",
				SrcIpRanges:   ranges,
			},
		}
	}

	ipDenyRule := func(priority int32, ranges []string) *GcpCloudArmorRule {
		return &GcpCloudArmorRule{
			Action:   "deny(403)",
			Priority: proto.Int32(priority),
			Match: &GcpCloudArmorRuleMatch{
				VersionedExpr: "SRC_IPS_V1",
				SrcIpRanges:   ranges,
			},
		}
	}

	celRule := func(priority int32, expr string, action string) *GcpCloudArmorRule {
		return &GcpCloudArmorRule{
			Action:   action,
			Priority: proto.Int32(priority),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
			JsonParsing:          "STANDARD",
			LogLevel:             "VERBOSE",
			UserIpRequestHeaders: []string{"X-Real-IP", "X-Forwarded-For"},
			JsonCustomConfig: &GcpCloudArmorJsonCustomConfig{
				ContentTypes: []string{"application/vnd.api+json"},
			},
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
				Match: &GcpCloudArmorRuleMatch{
					Expression: "evaluatePreconfiguredWaf('xss-v33-stable')",
				},
				PreconfiguredWafConfig: &GcpCloudArmorPreconfiguredWafConfig{
					Exclusions: []*GcpCloudArmorWafExclusion{
						{
							TargetRuleSet: "xss-v33-stable",
							TargetRuleIds: []string{"941100", "941110"},
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
			JsonParsing: "STANDARD",
			LogLevel:    "VERBOSE",
		}
		msg.Spec.Rules = []*GcpCloudArmorRule{
			ipAllowRule(100, []string{"10.0.0.0/8"}),
			celRule(200, "origin.region_code == 'CN' || origin.region_code == 'RU'", "deny(403)"),
			{
				Action:   "throttle",
				Priority: proto.Int32(300),
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction:      "allow",
					ExceedAction:       "deny(429)",
					EnforceOnKey:       "IP",
					RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
				},
			},
			{
				Action:   "deny(403)",
				Priority: proto.Int32(2147483647),
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty rules (the API auto-adds the default rule)", func() {
		msg := minimal()
		msg.Spec.Rules = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept user labels", func() {
		target := minimal()
		target.Spec.Labels = map[string]string{"team": "platform"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each request_body_inspection_size value", func() {
		for _, v := range []string{"8KB", "16KB", "32KB", "48KB", "64KB"} {
			target := minimal()
			target.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{
				RequestBodyInspectionSize: v,
			}
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject an invalid request_body_inspection_size", func() {
		target := minimal()
		target.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{
			RequestBodyInspectionSize: "128KB",
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept an omitted project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority:    proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction:      "allow",
					ExceedAction:       "deny(429)",
					EnforceOnKey:       "INVALID_KEY",
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
				Priority: proto.Int32(1000),
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction:      "deny",
					ExceedAction:       "deny(429)",
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
				Priority: proto.Int32(1000),
				Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
				RateLimitOptions: &GcpCloudArmorRateLimitOptions{
					ConformAction:      "allow",
					ExceedAction:       "deny(500)",
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
				Priority: proto.Int32(1000),
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
				Priority: proto.Int32(1000),
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

	// ──────────────── Default-rule and priority contracts ────────────────

	ginkgo.It("should reject a non-empty rule set without the default rule", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpCloudArmorRule{ipAllowRule(1000, []string{"10.0.0.0/8"})}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject duplicate rule priorities", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules,
			ipAllowRule(1000, []string{"10.0.0.0/8"}),
			ipDenyRule(1000, []string{"192.0.2.0/24"}),
		)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a negative rule priority", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, ipAllowRule(-1, []string{"10.0.0.0/8"}))
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Action/options coherence ────────────────

	ginkgo.It("should reject throttle without rate_limit_options", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "throttle",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject rate_limit_options on a plain allow rule", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "allow",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RateLimitOptions: &GcpCloudArmorRateLimitOptions{
				ConformAction:      "allow",
				ExceedAction:       "deny(429)",
				RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject the redirect action without redirect_options", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "redirect",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a redirect rule with an EXTERNAL_302 target", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "redirect",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RedirectOptions: &GcpCloudArmorRedirectConfig{
				Type:   "EXTERNAL_302",
				Target: "https://blocked.example.com",
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject EXTERNAL_302 redirect without a target", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "redirect",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RedirectOptions: &GcpCloudArmorRedirectConfig{
				Type: "EXTERNAL_302",
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject GOOGLE_RECAPTCHA redirect with a target", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "redirect",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RedirectOptions: &GcpCloudArmorRedirectConfig{
				Type:   "GOOGLE_RECAPTCHA",
				Target: "https://blocked.example.com",
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── reCAPTCHA expression options ────────────────

	ginkgo.It("should accept expr_options with reCAPTCHA site keys on a CEL match", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "allow",
			Priority: proto.Int32(1000),
			Match: &GcpCloudArmorRuleMatch{
				Expression: "token.recaptcha_action.score > 0.5",
				ExprOptions: &GcpCloudArmorRecaptchaOptions{
					ActionTokenSiteKeys: []string{"6LcA1234567890abcdefghijklmnopqrstuvwxyz"},
				},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject expr_options on an IP-based match", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "allow",
			Priority: proto.Int32(1000),
			Match: &GcpCloudArmorRuleMatch{
				VersionedExpr: "SRC_IPS_V1",
				SrcIpRanges:   []string{"*"},
				ExprOptions: &GcpCloudArmorRecaptchaOptions{
					ActionTokenSiteKeys: []string{"key"},
				},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject empty expr_options", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "allow",
			Priority: proto.Int32(1000),
			Match: &GcpCloudArmorRuleMatch{
				Expression:  "token.recaptcha_action.score > 0.5",
				ExprOptions: &GcpCloudArmorRecaptchaOptions{},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a policy-level reCAPTCHA redirect site key", func() {
		msg := minimal()
		msg.Spec.RecaptchaOptionsConfig = &GcpCloudArmorRecaptchaOptionsConfig{
			RedirectSiteKey: "6LcA1234567890abcdefghijklmnopqrstuvwxyz",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Composite rate-limit keys ────────────────

	ginkgo.It("should accept a composite rate-limit key (IP + HTTP_PATH)", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "throttle",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RateLimitOptions: &GcpCloudArmorRateLimitOptions{
				ConformAction: "allow",
				ExceedAction:  "deny(429)",
				EnforceOnKeyConfigs: []*GcpCloudArmorEnforceOnKeyConfig{
					{EnforceOnKeyType: "IP"},
					{EnforceOnKeyType: "HTTP_PATH"},
				},
				RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject enforce_on_key together with enforce_on_key_configs", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "throttle",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RateLimitOptions: &GcpCloudArmorRateLimitOptions{
				ConformAction: "allow",
				ExceedAction:  "deny(429)",
				EnforceOnKey:  "IP",
				EnforceOnKeyConfigs: []*GcpCloudArmorEnforceOnKeyConfig{
					{EnforceOnKeyType: "HTTP_PATH"},
				},
				RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a composite key component naming a header without HTTP_HEADER type", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "throttle",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RateLimitOptions: &GcpCloudArmorRateLimitOptions{
				ConformAction: "allow",
				ExceedAction:  "deny(429)",
				EnforceOnKeyConfigs: []*GcpCloudArmorEnforceOnKeyConfig{
					{EnforceOnKeyType: "IP", EnforceOnKeyName: "X-API-Key"},
				},
				RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject enforce_on_key HTTP_HEADER without a key name", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "throttle",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RateLimitOptions: &GcpCloudArmorRateLimitOptions{
				ConformAction:      "allow",
				ExceedAction:       "deny(429)",
				EnforceOnKey:       "HTTP_HEADER",
				RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a ban duration below 60 seconds", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "rate_based_ban",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RateLimitOptions: &GcpCloudArmorRateLimitOptions{
				ConformAction:      "allow",
				ExceedAction:       "deny(429)",
				BanDurationSec:     30,
				RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an interval outside the API's allowed set", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "throttle",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RateLimitOptions: &GcpCloudArmorRateLimitOptions{
				ConformAction:      "allow",
				ExceedAction:       "deny(429)",
				RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 90},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Adaptive protection thresholds ────────────────

	ginkgo.It("should accept adaptive protection threshold configs", func() {
		confidence := 0.8
		impacted := 0.1
		expiration := int32(3600)
		msg := minimal()
		msg.Spec.AdaptiveProtectionConfig = &GcpCloudArmorAdaptiveProtectionConfig{
			EnableLayer_7DdosDefense: true,
			ThresholdConfigs: []*GcpCloudArmorThresholdConfig{
				{
					Name:                                "per-host",
					AutoDeployConfidenceThreshold:       &confidence,
					AutoDeployImpactedBaselineThreshold: &impacted,
					AutoDeployExpirationSec:             &expiration,
					TrafficGranularityConfigs: []*GcpCloudArmorTrafficGranularityConfig{
						{Type: "HTTP_HEADER_HOST", EnableEachUniqueValue: true},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject threshold configs without layer-7 defense enabled", func() {
		msg := minimal()
		msg.Spec.AdaptiveProtectionConfig = &GcpCloudArmorAdaptiveProtectionConfig{
			ThresholdConfigs: []*GcpCloudArmorThresholdConfig{
				{Name: "per-host"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a granularity config with both value and each-unique-value", func() {
		msg := minimal()
		msg.Spec.AdaptiveProtectionConfig = &GcpCloudArmorAdaptiveProtectionConfig{
			EnableLayer_7DdosDefense: true,
			ThresholdConfigs: []*GcpCloudArmorThresholdConfig{
				{
					Name: "per-host",
					TrafficGranularityConfigs: []*GcpCloudArmorTrafficGranularityConfig{
						{Type: "HTTP_HEADER_HOST", Value: "api.example.com", EnableEachUniqueValue: true},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an out-of-range auto-deploy confidence threshold", func() {
		confidence := 1.5
		msg := minimal()
		msg.Spec.AdaptiveProtectionConfig = &GcpCloudArmorAdaptiveProtectionConfig{
			EnableLayer_7DdosDefense: true,
			ThresholdConfigs: []*GcpCloudArmorThresholdConfig{
				{Name: "per-host", AutoDeployConfidenceThreshold: &confidence},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── JSON custom config ────────────────

	ginkgo.It("should reject json_custom_config without JSON parsing enabled", func() {
		msg := minimal()
		msg.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{
			JsonCustomConfig: &GcpCloudArmorJsonCustomConfig{
				ContentTypes: []string{"application/vnd.api+json"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject WAF exclusion value with EQUALS_ANY", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "deny(403)",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{Expression: "evaluatePreconfiguredWaf('sqli-v33-stable')"},
			PreconfiguredWafConfig: &GcpCloudArmorPreconfiguredWafConfig{
				Exclusions: []*GcpCloudArmorWafExclusion{
					{
						TargetRuleSet: "sqli-v33-stable",
						RequestHeaders: []*GcpCloudArmorWafExclusionFieldParams{
							{Operator: "EQUALS_ANY", Value: "should-not-be-here"},
						},
					},
				},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a WAF exclusion EQUALS operator without a value", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:   "deny(403)",
			Priority: proto.Int32(1000),
			Match:    &GcpCloudArmorRuleMatch{Expression: "evaluatePreconfiguredWaf('sqli-v33-stable')"},
			PreconfiguredWafConfig: &GcpCloudArmorPreconfiguredWafConfig{
				Exclusions: []*GcpCloudArmorWafExclusion{
					{
						TargetRuleSet: "sqli-v33-stable",
						RequestHeaders: []*GcpCloudArmorWafExclusionFieldParams{
							{Operator: "EQUALS"},
						},
					},
				},
			},
		})
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ── Priority presence: 0 is Google's highest priority and a legal value ──

	ginkgo.It("should accept a rule at priority 0", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, ipDenyRule(0, []string{"203.0.113.0/24"}))
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a rule with no priority set", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action: "deny(403)",
			Match:  &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"203.0.113.0/24"}},
		})
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a rule priority above 2147483647", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, ipDenyRule(2147483647, []string{"203.0.113.0/24"}))
		// The default rule already holds 2147483647; duplicate priority fails.
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	// ── The match contract: exactly one of match / network_match per rule ──

	ginkgo.It("should reject a rule with neither match nor network_match", func() {
		msg := minimal()
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{Action: "allow", Priority: proto.Int32(100)})
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a rule with both match and network_match", func() {
		msg := regionalNetwork()
		msg.Spec.Rules[0].Match = &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	// ── The regional arm ──

	ginkgo.It("should accept a regional CLOUD_ARMOR policy with HTTP rules", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		msg.Spec.Rules = append(msg.Spec.Rules, celRule(100, "origin.region_code == 'US'", "allow"))
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional CLOUD_ARMOR_NETWORK policy with DDoS protection, user-defined fields, and a network match", func() {
		gomega.Expect(validator.Validate(regionalNetwork())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional CLOUD_ARMOR_NETWORK policy enrolled through a network edge security service", func() {
		msg := regionalNetwork()
		msg.Spec.DdosProtectionConfig.DdosProtection = "ADVANCED"
		msg.Spec.NetworkEdgeSecurityService = &GcpCloudArmorNetworkEdgeSecurityService{Name: "edge-us-central1", Description: "advanced DDoS enrollment"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid region name", func() {
		msg := minimal()
		msg.Spec.Region = "US_Central1"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject labels on a regional policy", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		msg.Spec.Labels = map[string]string{"team": "security"}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject adaptive protection on a regional policy", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		msg.Spec.AdaptiveProtectionConfig = &GcpCloudArmorAdaptiveProtectionConfig{EnableLayer_7DdosDefense: true}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject reCAPTCHA options on a regional policy", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		msg.Spec.RecaptchaOptionsConfig = &GcpCloudArmorRecaptchaOptionsConfig{RedirectSiteKey: "6Lc-key"}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject request_body_inspection_size on a regional policy but accept the other advanced options", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		msg.Spec.AdvancedOptionsConfig = &GcpCloudArmorAdvancedOptionsConfig{JsonParsing: "STANDARD", LogLevel: "VERBOSE"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.AdvancedOptionsConfig.RequestBodyInspectionSize = "16KB"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject the CLOUD_ARMOR_INTERNAL_SERVICE type on a regional policy", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		msg.Spec.Type = "CLOUD_ARMOR_INTERNAL_SERVICE"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject header_action on a regional policy", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		r := ipAllowRule(100, []string{"10.0.0.0/8"})
		r.HeaderAction = &GcpCloudArmorHeaderAction{RequestHeadersToAdds: []*GcpCloudArmorRequestHeader{{HeaderName: "X-Trusted", HeaderValue: "1"}}}
		msg.Spec.Rules = append(msg.Spec.Rules, r)
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject the redirect action on a regional policy", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		msg.Spec.Rules = append(msg.Spec.Rules, &GcpCloudArmorRule{
			Action:          "redirect",
			Priority:        proto.Int32(100),
			Match:           &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RedirectOptions: &GcpCloudArmorRedirectConfig{Type: "EXTERNAL_302", Target: "https://example.com/blocked"},
		})
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a rate limit that exceeds to redirect on a regional policy but accept deny", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		r := &GcpCloudArmorRule{
			Action:   "throttle",
			Priority: proto.Int32(100),
			Match:    &GcpCloudArmorRuleMatch{VersionedExpr: "SRC_IPS_V1", SrcIpRanges: []string{"*"}},
			RateLimitOptions: &GcpCloudArmorRateLimitOptions{
				ConformAction:      "allow",
				ExceedAction:       "deny(429)",
				EnforceOnKey:       "IP",
				RateLimitThreshold: &GcpCloudArmorRateThreshold{Count: 100, IntervalSec: 60},
			},
		}
		msg.Spec.Rules = append(msg.Spec.Rules, r)
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		r.RateLimitOptions.ExceedAction = "redirect"
		r.RateLimitOptions.ExceedRedirectOptions = &GcpCloudArmorRedirectConfig{Type: "GOOGLE_RECAPTCHA"}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject expr_options on a regional policy", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		r := celRule(100, "token.recaptcha_action.score < 0.5", "deny(403)")
		r.Match.ExprOptions = &GcpCloudArmorRecaptchaOptions{ActionTokenSiteKeys: []string{"6Lc-key"}}
		msg.Spec.Rules = append(msg.Spec.Rules, r)
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject the CLOUD_ARMOR_NETWORK type on a global policy", func() {
		msg := regionalNetwork()
		msg.Spec.Region = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject ddos_protection_config on a regional CLOUD_ARMOR policy", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1"
		msg.Spec.DdosProtectionConfig = &GcpCloudArmorDdosProtectionConfig{DdosProtection: "STANDARD"}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid ddos_protection level", func() {
		msg := regionalNetwork()
		msg.Spec.DdosProtectionConfig.DdosProtection = "PREMIUM"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject user_defined_fields on a global policy", func() {
		msg := minimal()
		msg.Spec.UserDefinedFields = []*GcpCloudArmorUserDefinedField{{Name: "SIG1_AT_0", Base: "TCP", Offset: proto.Int32(8), Size: proto.Int32(2)}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a user-defined field with an invalid base, size, or mask", func() {
		msg := regionalNetwork()
		msg.Spec.UserDefinedFields[0].Base = "ICMP"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = regionalNetwork()
		msg.Spec.UserDefinedFields[0].Size = proto.Int32(5)
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg = regionalNetwork()
		msg.Spec.UserDefinedFields[0].Mask = "8F00"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a user-defined field at offset 0", func() {
		msg := regionalNetwork()
		msg.Spec.UserDefinedFields[0].Offset = proto.Int32(0)
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject duplicate user-defined field names", func() {
		msg := regionalNetwork()
		msg.Spec.UserDefinedFields = append(msg.Spec.UserDefinedFields, &GcpCloudArmorUserDefinedField{Name: "SIG1_AT_0", Base: "UDP", Offset: proto.Int32(4), Size: proto.Int32(4)})
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a network match naming a user-defined field the policy never defined", func() {
		msg := regionalNetwork()
		msg.Spec.Rules[0].NetworkMatch.UserDefinedFields[0].Name = "SIG_UNDEFINED"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a network match user-defined field without values", func() {
		msg := regionalNetwork()
		msg.Spec.Rules[0].NetworkMatch.UserDefinedFields[0].Values = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject network_match on a regional CLOUD_ARMOR policy", func() {
		msg := regionalNetwork()
		msg.Spec.Type = "CLOUD_ARMOR"
		msg.Spec.DdosProtectionConfig = nil
		msg.Spec.UserDefinedFields = nil
		msg.Spec.Rules[0].NetworkMatch.UserDefinedFields = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a network edge security service without ddos_protection_config", func() {
		msg := regionalNetwork()
		msg.Spec.DdosProtectionConfig = nil
		msg.Spec.NetworkEdgeSecurityService = &GcpCloudArmorNetworkEdgeSecurityService{}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a network edge security service with an invalid name", func() {
		msg := regionalNetwork()
		msg.Spec.NetworkEdgeSecurityService = &GcpCloudArmorNetworkEdgeSecurityService{Name: "Edge_Service"}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})
})

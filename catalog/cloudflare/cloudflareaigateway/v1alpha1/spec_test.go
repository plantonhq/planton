package cloudflareaigatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareAiGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareAiGatewaySpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validGateway(spec *CloudflareAiGatewaySpec) *CloudflareAiGateway {
	return &CloudflareAiGateway{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareAiGateway",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ai-gateway",
		},
		Spec: spec,
	}
}

// baseSpec carries the five required scalars every gateway must state.
func baseSpec() *CloudflareAiGatewaySpec {
	return &CloudflareAiGatewaySpec{
		AccountId:               testAccountID,
		GatewayId:               "prod-llm-gateway",
		CacheInvalidateOnUpdate: proto.Bool(true),
		CacheTtl:                proto.Int64(300),
		CollectLogs:             proto.Bool(true),
		RateLimitingInterval:    proto.Int64(60),
		RateLimitingLimit:       proto.Int64(1000),
	}
}

var _ = ginkgo.Describe("CloudflareAiGatewaySpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept the five required scalars alone", func() {
			gomega.Expect(protovalidate.Validate(validGateway(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept zero for cache and rate-limiting (disabled)", func() {
			spec := baseSpec()
			spec.CacheTtl = proto.Int64(0)
			spec.RateLimitingInterval = proto.Int64(0)
			spec.RateLimitingLimit = proto.Int64(0)
			gomega.Expect(protovalidate.Validate(validGateway(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept retry, log management, guardrails, and spend limits together", func() {
			spec := baseSpec()
			spec.RateLimitingTechnique = "sliding"
			spec.Retry = &CloudflareAiGatewayRetry{Backoff: "exponential", Delay: proto.Int64(1000), MaxAttempts: proto.Int64(3)}
			spec.LogManagement = &CloudflareAiGatewayLogManagement{MaxRecords: proto.Int64(100000), Strategy: "DELETE_OLDEST"}
			spec.Guardrails = &CloudflareAiGatewayGuardrails{
				Prompt:   &CloudflareAiGatewayGuardrailsControls{P1: "BLOCK", S1: "FLAG"},
				Response: &CloudflareAiGatewayGuardrailsControls{},
			}
			spec.SpendLimits = &CloudflareAiGatewaySpendLimits{
				Enabled: proto.Bool(true),
				Rules: []*CloudflareAiGatewaySpendLimitsRule{{
					Id:        "daily-cap",
					Limit:     proto.Float64(50),
					LimitType: "cost",
					Window:    proto.Int64(86400),
				}},
			}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a dynamic route with a conditional graph", func() {
			spec := baseSpec()
			spec.DynamicRoutes = []*CloudflareAiGatewayDynamicRoute{{
				Name: "cheap-first",
				Elements: []*CloudflareAiGatewayRouteElement{
					{
						Id:   "start",
						Type: "start",
						Outputs: &CloudflareAiGatewayRouteElementOutputs{
							Next: &CloudflareAiGatewayRouteElementBranch{ElementId: "check"},
						},
					},
					{
						Id:   "check",
						Type: "conditional",
						Outputs: &CloudflareAiGatewayRouteElementOutputs{
							OnTrue:  &CloudflareAiGatewayRouteElementBranch{ElementId: "cheap"},
							OnFalse: &CloudflareAiGatewayRouteElementBranch{ElementId: "smart"},
						},
						Properties: &CloudflareAiGatewayRouteElementProperties{
							Conditions: `{"metadata.tier": {"$eq": "free"}}`,
						},
					},
					{
						Id:   "cheap",
						Type: "model",
						Outputs: &CloudflareAiGatewayRouteElementOutputs{
							Success: &CloudflareAiGatewayRouteElementBranch{ElementId: "done"},
						},
						Properties: &CloudflareAiGatewayRouteElementProperties{
							Model:    "@cf/meta/llama-3.1-8b-instruct",
							Provider: "workers-ai",
						},
					},
					{
						Id:   "smart",
						Type: "model",
						Outputs: &CloudflareAiGatewayRouteElementOutputs{
							Success: &CloudflareAiGatewayRouteElementBranch{ElementId: "done"},
						},
						Properties: &CloudflareAiGatewayRouteElementProperties{
							Model:    "gpt-4o",
							Provider: "openai",
						},
					},
					{
						Id:      "done",
						Type:    "end",
						Outputs: &CloudflareAiGatewayRouteElementOutputs{},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a spec missing the required scalars", func() {
			input := validGateway(&CloudflareAiGatewaySpec{
				AccountId: testAccountID,
				GatewayId: "prod-llm-gateway",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown rate-limiting technique", func() {
			spec := baseSpec()
			spec.RateLimitingTechnique = "rolling"
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a retry delay above 5000ms", func() {
			spec := baseSpec()
			spec.Retry = &CloudflareAiGatewayRetry{Delay: proto.Int64(6000)}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a log cap below 10000 records", func() {
			spec := baseSpec()
			spec.LogManagement = &CloudflareAiGatewayLogManagement{MaxRecords: proto.Int64(500)}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a guardrail control outside FLAG/BLOCK", func() {
			spec := baseSpec()
			spec.Guardrails = &CloudflareAiGatewayGuardrails{
				Prompt:   &CloudflareAiGatewayGuardrailsControls{S3: "DENY"},
				Response: &CloudflareAiGatewayGuardrailsControls{},
			}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject guardrails missing the response side", func() {
			spec := baseSpec()
			spec.Guardrails = &CloudflareAiGatewayGuardrails{
				Prompt: &CloudflareAiGatewayGuardrailsControls{P1: "FLAG"},
			}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate spend-limit rule ids -- omitted ids collapsed into one at the API", func() {
			spec := baseSpec()
			spec.SpendLimits = &CloudflareAiGatewaySpendLimits{
				Rules: []*CloudflareAiGatewaySpendLimitsRule{
					{Id: "cap", Limit: proto.Float64(50), LimitType: "cost", Window: proto.Int64(86400)},
					{Id: "cap", Limit: proto.Float64(500), LimitType: "cost", Window: proto.Int64(2592000)},
				},
			}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a spend-limit rule without an id", func() {
			spec := baseSpec()
			spec.SpendLimits = &CloudflareAiGatewaySpendLimits{
				Rules: []*CloudflareAiGatewaySpendLimitsRule{
					{Limit: proto.Float64(50), LimitType: "cost", Window: proto.Int64(86400)},
				},
			}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a DLP policy checking an unknown direction", func() {
			spec := baseSpec()
			spec.Dlp = &CloudflareAiGatewayDlp{
				Enabled: proto.Bool(true),
				Policies: []*CloudflareAiGatewayDlpPolicy{{
					Id:       "pii",
					Enabled:  proto.Bool(true),
					Action:   "BLOCK",
					Check:    []string{"BOTH"},
					Profiles: []string{"profile-1"},
				}},
			}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a stripe block without its credential", func() {
			spec := baseSpec()
			spec.Stripe = &CloudflareAiGatewayStripe{
				UsageEvents: []*CloudflareAiGatewayStripeUsageEvent{{Payload: "{}"}},
			}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a route element without a type", func() {
			spec := baseSpec()
			spec.DynamicRoutes = []*CloudflareAiGatewayDynamicRoute{{
				Name: "broken",
				Elements: []*CloudflareAiGatewayRouteElement{{
					Id:      "start",
					Outputs: &CloudflareAiGatewayRouteElementOutputs{},
				}},
			}}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown route element type", func() {
			spec := baseSpec()
			spec.DynamicRoutes = []*CloudflareAiGatewayDynamicRoute{{
				Name: "broken",
				Elements: []*CloudflareAiGatewayRouteElement{{
					Id:      "start",
					Type:    "splitter",
					Outputs: &CloudflareAiGatewayRouteElementOutputs{},
				}},
			}}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an otel destination without a url", func() {
			spec := baseSpec()
			spec.Otel = []*CloudflareAiGatewayOtel{{
				Authorization: literal("Bearer test"),
			}}
			gomega.Expect(protovalidate.Validate(validGateway(spec))).NotTo(gomega.BeNil())
		})
	})
})

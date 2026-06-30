package gcpfirewallrulev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestGcpFirewallRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpFirewallRuleSpec Validation Suite")
}

// stringPtr returns a pointer to the given string.
func stringPtr(s string) *string {
	return &s
}

// int32Ptr returns a pointer to the given int32.
func int32Ptr(i int32) *int32 {
	return &i
}

// validFirewallRule returns a minimal valid GcpFirewallRule for testing.
func validFirewallRule() *GcpFirewallRule {
	priority := int32(1000)
	return &GcpFirewallRule{
		ApiVersion: "gcp.planton.dev/v1",
		Kind:       "GcpFirewallRule",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-firewall-rule",
		},
		Spec: &GcpFirewallRuleSpec{
			ProjectId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
			},
			Network: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "default"},
			},
			RuleName:  "allow-http-ingress",
			Direction: "INGRESS",
			Action:    "ALLOW",
			Rules: []*GcpFirewallProtocolPort{
				{Protocol: "tcp", Ports: []string{"80", "443"}},
			},
			Priority:     &priority,
			SourceRanges: []string{"0.0.0.0/0"},
		},
	}
}

var _ = ginkgo.Describe("GcpFirewallRuleSpec Validations", func() {

	ginkgo.Describe("Valid inputs", func() {
		ginkgo.It("should accept a minimal valid INGRESS ALLOW rule", func() {
			input := validFirewallRule()
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an EGRESS DENY rule without source fields", func() {
			input := validFirewallRule()
			input.Spec.Direction = "EGRESS"
			input.Spec.Action = "DENY"
			input.Spec.SourceRanges = nil
			input.Spec.DestinationRanges = []string{"0.0.0.0/0"}
			input.Spec.Rules = []*GcpFirewallProtocolPort{
				{Protocol: "all"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an INGRESS rule with source_tags instead of source_ranges", func() {
			input := validFirewallRule()
			input.Spec.SourceRanges = nil
			input.Spec.SourceTags = []string{"web-server"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an INGRESS rule with source_service_accounts", func() {
			input := validFirewallRule()
			input.Spec.SourceRanges = nil
			input.Spec.SourceServiceAccounts = []string{"my-sa@my-project.iam.gserviceaccount.com"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a rule with log_config", func() {
			input := validFirewallRule()
			input.Spec.LogConfig = &GcpFirewallLogConfig{
				Metadata: "INCLUDE_ALL_METADATA",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a disabled rule", func() {
			input := validFirewallRule()
			input.Spec.Disabled = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Invalid inputs", func() {
		ginkgo.It("should reject when rule_name is missing", func() {
			input := validFirewallRule()
			input.Spec.RuleName = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when rule_name has invalid format (uppercase)", func() {
			input := validFirewallRule()
			input.Spec.RuleName = "INVALID-NAME"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when direction is invalid", func() {
			input := validFirewallRule()
			input.Spec.Direction = "INVALID"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when action is invalid", func() {
			input := validFirewallRule()
			input.Spec.Action = "REJECT"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when rules is empty", func() {
			input := validFirewallRule()
			input.Spec.Rules = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject INGRESS without any source field", func() {
			input := validFirewallRule()
			input.Spec.SourceRanges = nil
			input.Spec.SourceTags = nil
			input.Spec.SourceServiceAccounts = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject when tags and service accounts are mixed", func() {
			input := validFirewallRule()
			input.Spec.SourceTags = []string{"web"}
			input.Spec.TargetServiceAccounts = []string{"my-sa@project.iam.gserviceaccount.com"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject priority below 0", func() {
			input := validFirewallRule()
			p := int32(-1)
			input.Spec.Priority = &p
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject priority above 65535", func() {
			input := validFirewallRule()
			p := int32(65536)
			input.Spec.Priority = &p
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject more than 10 source_service_accounts", func() {
			input := validFirewallRule()
			input.Spec.SourceRanges = nil
			input.Spec.SourceServiceAccounts = []string{
				"sa1@p.iam.gserviceaccount.com", "sa2@p.iam.gserviceaccount.com",
				"sa3@p.iam.gserviceaccount.com", "sa4@p.iam.gserviceaccount.com",
				"sa5@p.iam.gserviceaccount.com", "sa6@p.iam.gserviceaccount.com",
				"sa7@p.iam.gserviceaccount.com", "sa8@p.iam.gserviceaccount.com",
				"sa9@p.iam.gserviceaccount.com", "sa10@p.iam.gserviceaccount.com",
				"sa11@p.iam.gserviceaccount.com",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject invalid log_config metadata", func() {
			input := validFirewallRule()
			input.Spec.LogConfig = &GcpFirewallLogConfig{
				Metadata: "INVALID",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})

package cloudflarebotmanagementv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareBotManagementSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareBotManagementSpec Custom Validation Tests")
}

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0da42c8d2132a9ddaf714f9e7c920711"},
	}
}

func validBotManagement(spec *CloudflareBotManagementSpec) *CloudflareBotManagement {
	return &CloudflareBotManagement{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareBotManagement",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-bot-management",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareBotManagementSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a free-plan fight-mode toggle", func() {
			input := validBotManagement(&CloudflareBotManagementSpec{
				ZoneId:    zoneRef(),
				FightMode: proto.Bool(true),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an SBFM configuration", func() {
			input := validBotManagement(&CloudflareBotManagementSpec{
				ZoneId:                       zoneRef(),
				SbfmDefinitelyAutomated:      proto.String("managed_challenge"),
				SbfmLikelyAutomated:          proto.String("allow"),
				SbfmVerifiedBots:             proto.String("allow"),
				SbfmStaticResourceProtection: proto.Bool(false),
				OptimizeWordpress:            proto.Bool(true),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept AI-crawler controls", func() {
			input := validBotManagement(&CloudflareBotManagementSpec{
				ZoneId:             zoneRef(),
				AiBotsProtection:   proto.String("only_on_ad_pages"),
				CrawlerProtection:  proto.String("enabled"),
				CfRobotsVariant:    proto.String("policy_only"),
				IsRobotsTxtManaged: proto.Bool(true),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a spec that manages nothing", func() {
			input := validBotManagement(&CloudflareBotManagementSpec{
				ZoneId: zoneRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing zone_id", func() {
			input := validBotManagement(&CloudflareBotManagementSpec{
				FightMode: proto.Bool(true),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown SBFM action", func() {
			input := validBotManagement(&CloudflareBotManagementSpec{
				ZoneId:                  zoneRef(),
				SbfmDefinitelyAutomated: proto.String("challenge"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject managed_challenge on verified bots (allow|block only)", func() {
			input := validBotManagement(&CloudflareBotManagementSpec{
				ZoneId:           zoneRef(),
				SbfmVerifiedBots: proto.String("managed_challenge"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown ai_bots_protection value", func() {
			input := validBotManagement(&CloudflareBotManagementSpec{
				ZoneId:           zoneRef(),
				AiBotsProtection: proto.String("enabled"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})

package cloudflarewaitingroomv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareWaitingRoomSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareWaitingRoomSpec Custom Validation Tests")
}

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0da42c8d2132a9ddaf714f9e7c920711"},
	}
}

func validRoom(spec *CloudflareWaitingRoomSpec) *CloudflareWaitingRoom {
	return &CloudflareWaitingRoom{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareWaitingRoom",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-waiting-room",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareWaitingRoomSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal room", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:            zoneRef(),
				Name:              "launch-queue",
				Host:              "shop.example.com",
				NewUsersPerMinute: 200,
				TotalActiveUsers:  200,
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a fully-shaped room with bypass rules", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:                  zoneRef(),
				Name:                    "onsale-queue",
				Host:                    "tickets.example.com",
				Path:                    proto.String("/onsale"),
				NewUsersPerMinute:       500,
				TotalActiveUsers:        2000,
				SessionDuration:         proto.Int32(10),
				QueueingMethod:          proto.String("random"),
				QueueingStatusCode:      proto.Int32(202),
				DefaultTemplateLanguage: proto.String("de-DE"),
				CookieAttributes: &CloudflareWaitingRoomCookieAttributes{
					Samesite: proto.String("lax"),
					Secure:   proto.String("always"),
				},
				CookieSuffix: "onsale",
				AdditionalRoutes: []*CloudflareWaitingRoomRoute{
					{Host: "www.example.com", Path: proto.String("/onsale")},
				},
				EnabledOriginCommands: []string{"revoke"},
				TurnstileAction:       proto.String("infinite_queue"),
				TurnstileMode:         proto.String("visible_managed"),
				BypassRules: []*CloudflareWaitingRoomBypassRule{
					{
						Expression:  `ip.src in {203.0.113.0/24}`,
						Description: "office network skips the queue",
					},
					{
						Expression: `http.request.uri.path eq "/healthz"`,
						Enabled:    proto.Bool(false),
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject new_users_per_minute below Cloudflare's floor of 200", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:            zoneRef(),
				Name:              "launch-queue",
				Host:              "shop.example.com",
				NewUsersPerMinute: 100,
				TotalActiveUsers:  200,
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a session duration above 30 minutes", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:            zoneRef(),
				Name:              "launch-queue",
				Host:              "shop.example.com",
				NewUsersPerMinute: 200,
				TotalActiveUsers:  200,
				SessionDuration:   proto.Int32(45),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown queueing method", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:            zoneRef(),
				Name:              "launch-queue",
				Host:              "shop.example.com",
				NewUsersPerMinute: 200,
				TotalActiveUsers:  200,
				QueueingMethod:    proto.String("lifo"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a disallowed queue status code", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:             zoneRef(),
				Name:               "launch-queue",
				Host:               "shop.example.com",
				NewUsersPerMinute:  200,
				TotalActiveUsers:   200,
				QueueingStatusCode: proto.Int32(503),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unsupported template language", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:                  zoneRef(),
				Name:                    "launch-queue",
				Host:                    "shop.example.com",
				NewUsersPerMinute:       200,
				TotalActiveUsers:        200,
				DefaultTemplateLanguage: proto.String("en-GB"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown origin command", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:                zoneRef(),
				Name:                  "launch-queue",
				Host:                  "shop.example.com",
				NewUsersPerMinute:     200,
				TotalActiveUsers:      200,
				EnabledOriginCommands: []string{"pause"},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a bypass rule without an expression", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:            zoneRef(),
				Name:              "launch-queue",
				Host:              "shop.example.com",
				NewUsersPerMinute: 200,
				TotalActiveUsers:  200,
				BypassRules: []*CloudflareWaitingRoomBypassRule{
					{Description: "missing expression"},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing host", func() {
			input := validRoom(&CloudflareWaitingRoomSpec{
				ZoneId:            zoneRef(),
				Name:              "launch-queue",
				NewUsersPerMinute: 200,
				TotalActiveUsers:  200,
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})

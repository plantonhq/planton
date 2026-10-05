package cloudflarewaitingroomeventv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareWaitingRoomEventSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareWaitingRoomEventSpec Custom Validation Tests")
}

func roomRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "699d98642c564d2e855e9661899b7252"},
	}
}

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0da42c8d2132a9ddaf714f9e7c920711"},
	}
}

func validEvent(spec *CloudflareWaitingRoomEventSpec) *CloudflareWaitingRoomEvent {
	return &CloudflareWaitingRoomEvent{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareWaitingRoomEvent",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-waiting-room-event",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareWaitingRoomEventSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal event window", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				WaitingRoomId:  roomRef(),
				ZoneId:         zoneRef(),
				Name:           "product-launch",
				EventStartTime: "2026-09-01T10:00:00Z",
				EventEndTime:   "2026-09-01T14:00:00Z",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a prequeued shuffled event with overrides", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				WaitingRoomId:       roomRef(),
				ZoneId:              zoneRef(),
				Name:                "ticket-onsale",
				EventStartTime:      "2026-09-01T10:00:00Z",
				EventEndTime:        "2026-09-01T14:00:00Z",
				PrequeueStartTime:   "2026-09-01T09:30:00Z",
				ShuffleAtEventStart: proto.Bool(true),
				NewUsersPerMinute:   proto.Int32(300),
				TotalActiveUsers:    proto.Int32(1500),
				QueueingMethod:      proto.String("random"),
				SessionDuration:     proto.Int32(5),
				TurnstileAction:     proto.String("log"),
				TurnstileMode:       proto.String("invisible"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an end time before the start time", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				WaitingRoomId:  roomRef(),
				ZoneId:         zoneRef(),
				Name:           "product-launch",
				EventStartTime: "2026-09-01T14:00:00Z",
				EventEndTime:   "2026-09-01T10:00:00Z",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a window shorter than one minute", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				WaitingRoomId:  roomRef(),
				ZoneId:         zoneRef(),
				Name:           "product-launch",
				EventStartTime: "2026-09-01T10:00:00Z",
				EventEndTime:   "2026-09-01T10:00:30Z",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a prequeue closer than five minutes to the start", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				WaitingRoomId:     roomRef(),
				ZoneId:            zoneRef(),
				Name:              "product-launch",
				EventStartTime:    "2026-09-01T10:00:00Z",
				EventEndTime:      "2026-09-01T14:00:00Z",
				PrequeueStartTime: "2026-09-01T09:58:00Z",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject shuffle without a prequeue", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				WaitingRoomId:       roomRef(),
				ZoneId:              zoneRef(),
				Name:                "product-launch",
				EventStartTime:      "2026-09-01T10:00:00Z",
				EventEndTime:        "2026-09-01T14:00:00Z",
				ShuffleAtEventStart: proto.Bool(true),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject new_users_per_minute without total_active_users", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				WaitingRoomId:     roomRef(),
				ZoneId:            zoneRef(),
				Name:              "product-launch",
				EventStartTime:    "2026-09-01T10:00:00Z",
				EventEndTime:      "2026-09-01T14:00:00Z",
				NewUsersPerMinute: proto.Int32(300),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed start time", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				WaitingRoomId:  roomRef(),
				ZoneId:         zoneRef(),
				Name:           "product-launch",
				EventStartTime: "September 1st, 10am",
				EventEndTime:   "2026-09-01T14:00:00Z",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown override queueing method", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				WaitingRoomId:  roomRef(),
				ZoneId:         zoneRef(),
				Name:           "product-launch",
				EventStartTime: "2026-09-01T10:00:00Z",
				EventEndTime:   "2026-09-01T14:00:00Z",
				QueueingMethod: proto.String("lifo"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing waiting_room_id", func() {
			input := validEvent(&CloudflareWaitingRoomEventSpec{
				ZoneId:         zoneRef(),
				Name:           "product-launch",
				EventStartTime: "2026-09-01T10:00:00Z",
				EventEndTime:   "2026-09-01T14:00:00Z",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})

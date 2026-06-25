package cloudflareemailroutingzonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func value(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validZone() *CloudflareEmailRoutingZone {
	return &CloudflareEmailRoutingZone{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareEmailRoutingZone",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-email-zone"},
		Spec: &CloudflareEmailRoutingZoneSpec{
			ZoneId: value("023e105f4ecef8ad9ca31a8372d0c353"),
		},
	}
}

func TestCloudflareEmailRoutingZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareEmailRoutingZoneSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareEmailRoutingZoneSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal zone (no catch-all)", func() {
			gomega.Expect(protovalidate.Validate(validZone())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a drop catch-all", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Enabled: true,
				Type:    CloudflareEmailRoutingCatchAllActionType_drop,
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a forward catch-all with forward_to", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Enabled:   true,
				Type:      CloudflareEmailRoutingCatchAllActionType_forward,
				ForwardTo: []*foreignkeyv1.StringValueOrRef{value("ops@example.com")},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a worker catch-all with worker", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Type:   CloudflareEmailRoutingCatchAllActionType_worker,
				Worker: value("email-router"),
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing zone_id", func() {
			in := validZone()
			in.Spec.ZoneId = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a catch-all with unspecified type", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{Enabled: true}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a forward catch-all without forward_to", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Type: CloudflareEmailRoutingCatchAllActionType_forward,
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a worker catch-all without worker", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Type: CloudflareEmailRoutingCatchAllActionType_worker,
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})

package cloudflareemailroutingzonev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func value(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validZone() *CloudflareEmailRoutingZone {
	return &CloudflareEmailRoutingZone{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareEmailRoutingZone",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-email-zone"},
		Spec: &CloudflareEmailRoutingZoneSpec{
			ZoneId: value("023e105f4ecef8ad9ca31a8372d0c353"),
		},
	}
}

func dropAction() *CloudflareEmailRoutingCatchAllAction {
	return &CloudflareEmailRoutingCatchAllAction{Type: CloudflareEmailRoutingCatchAllActionType_drop}
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
				Actions: []*CloudflareEmailRoutingCatchAllAction{dropAction()},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a named forward catch-all with forward_to", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Enabled: true,
				Name:    "fallback-forward",
				Actions: []*CloudflareEmailRoutingCatchAllAction{{
					Type:      CloudflareEmailRoutingCatchAllActionType_forward,
					ForwardTo: []*foreignkeyv1.StringValueOrRef{value("ops@example.com")},
				}},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a multi-action catch-all (forward AND worker)", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Enabled: true,
				Actions: []*CloudflareEmailRoutingCatchAllAction{
					{
						Type:      CloudflareEmailRoutingCatchAllActionType_forward,
						ForwardTo: []*foreignkeyv1.StringValueOrRef{value("ops@example.com")},
					},
					{
						Type:   CloudflareEmailRoutingCatchAllActionType_worker,
						Worker: value("email-router"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a subdomain dns_name when lock_dns_records is true", func() {
			in := validZone()
			in.Spec.LockDnsRecords = true
			in.Spec.DnsName = "mail.example.com"
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing zone_id", func() {
			in := validZone()
			in.Spec.ZoneId = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a catch-all with no actions", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{Enabled: true}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a catch-all action with unspecified type", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Enabled: true,
				Actions: []*CloudflareEmailRoutingCatchAllAction{{}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a forward catch-all action without forward_to", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Actions: []*CloudflareEmailRoutingCatchAllAction{{
					Type: CloudflareEmailRoutingCatchAllActionType_forward,
				}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a worker catch-all action without worker", func() {
			in := validZone()
			in.Spec.CatchAll = &CloudflareEmailRoutingZoneCatchAll{
				Actions: []*CloudflareEmailRoutingCatchAllAction{{
					Type: CloudflareEmailRoutingCatchAllActionType_worker,
				}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects dns_name without lock_dns_records", func() {
			in := validZone()
			in.Spec.DnsName = "mail.example.com"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})

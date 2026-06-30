package gcpdnszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestGcpDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDnsZoneSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpDnsZoneSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_dns_zone", func() {

			ginkgo.It("should not return a validation error for minimal valid fields with literal value", func() {
				input := &GcpDnsZone{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-zone",
					},
					Spec: &GcpDnsZoneSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when project_id is missing", func() {
				input := &GcpDnsZone{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-zone",
					},
					Spec: &GcpDnsZoneSpec{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

package cloudflareauthenticatedoriginpullsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareAuthenticatedOriginPullsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareAuthenticatedOriginPullsSpec Custom Validation Tests")
}

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0da42c8d2132a9ddaf714f9e7c920711"},
	}
}

func certRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "2458ce5a-0c35-4c7f-82c7-8e9487d3ff60"},
	}
}

func validAop(spec *CloudflareAuthenticatedOriginPullsSpec) *CloudflareAuthenticatedOriginPulls {
	return &CloudflareAuthenticatedOriginPulls{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareAuthenticatedOriginPulls",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-aop",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareAuthenticatedOriginPullsSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept managing only the zone toggle", func() {
			input := validAop(&CloudflareAuthenticatedOriginPullsSpec{
				ZoneId:      zoneRef(),
				ZoneEnabled: proto.Bool(true),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an explicit zone_enabled false -- presence, not truthiness, satisfies the wall", func() {
			input := validAop(&CloudflareAuthenticatedOriginPullsSpec{
				ZoneId:      zoneRef(),
				ZoneEnabled: proto.Bool(false),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept associations without the zone toggle", func() {
			input := validAop(&CloudflareAuthenticatedOriginPullsSpec{
				ZoneId: zoneRef(),
				HostnameAssociations: []*CloudflareAuthenticatedOriginPullsHostnameAssociation{
					{
						Hostname:      "app.example.com",
						CertificateId: certRef(),
					},
					{
						Hostname:      "api.example.com",
						CertificateId: certRef(),
						Enabled:       proto.Bool(false),
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a spec managing nothing", func() {
			input := validAop(&CloudflareAuthenticatedOriginPullsSpec{
				ZoneId: zoneRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an association with an empty hostname", func() {
			input := validAop(&CloudflareAuthenticatedOriginPullsSpec{
				ZoneId: zoneRef(),
				HostnameAssociations: []*CloudflareAuthenticatedOriginPullsHostnameAssociation{
					{CertificateId: certRef()},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an association without a certificate -- Cloudflare requires the id on every association write (400 code 1404, measured live)", func() {
			input := validAop(&CloudflareAuthenticatedOriginPullsSpec{
				ZoneId: zoneRef(),
				HostnameAssociations: []*CloudflareAuthenticatedOriginPullsHostnameAssociation{
					{Hostname: "app.example.com"},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing zone reference", func() {
			input := validAop(&CloudflareAuthenticatedOriginPullsSpec{
				ZoneEnabled: proto.Bool(true),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})

package cloudflarezerotrustaccessinfrastructuretargetv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareZeroTrustAccessInfrastructureTargetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustAccessInfrastructureTargetSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validTarget(spec *CloudflareZeroTrustAccessInfrastructureTargetSpec) *CloudflareZeroTrustAccessInfrastructureTarget {
	return &CloudflareZeroTrustAccessInfrastructureTarget{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustAccessInfrastructureTarget",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-target",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareZeroTrustAccessInfrastructureTargetSpec {
	return &CloudflareZeroTrustAccessInfrastructureTargetSpec{
		AccountId: testAccountID,
		Hostname:  "prod-db-1",
		Ip: &CloudflareZeroTrustAccessInfrastructureTargetIp{
			Ipv4: &CloudflareZeroTrustAccessInfrastructureTargetIpInfo{
				IpAddr: "10.0.10.5",
			},
		},
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustAccessInfrastructureTargetSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept an IPv4-only target", func() {
			gomega.Expect(protovalidate.Validate(validTarget(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept an IPv6-only target", func() {
			spec := baseSpec()
			spec.Ip = &CloudflareZeroTrustAccessInfrastructureTargetIp{
				Ipv6: &CloudflareZeroTrustAccessInfrastructureTargetIpInfo{
					IpAddr: "2001:db8::5",
				},
			}
			gomega.Expect(protovalidate.Validate(validTarget(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept both families with a virtual network reference", func() {
			spec := baseSpec()
			spec.Ip.Ipv4.VirtualNetworkId = literal("f70ff985-a4ef-4643-bbbc-4a0ed4fc8415")
			spec.Ip.Ipv6 = &CloudflareZeroTrustAccessInfrastructureTargetIpInfo{
				IpAddr: "2001:db8::5",
			}
			gomega.Expect(protovalidate.Validate(validTarget(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a dotted hostname", func() {
			spec := baseSpec()
			spec.Hostname = "db1.internal.example"
			gomega.Expect(protovalidate.Validate(validTarget(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an empty ip block -- a target needs at least one address family", func() {
			spec := baseSpec()
			spec.Ip = &CloudflareZeroTrustAccessInfrastructureTargetIp{}
			gomega.Expect(protovalidate.Validate(validTarget(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing ip block", func() {
			spec := baseSpec()
			spec.Ip = nil
			gomega.Expect(protovalidate.Validate(validTarget(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a declared family without an address", func() {
			spec := baseSpec()
			spec.Ip.Ipv4.IpAddr = ""
			gomega.Expect(protovalidate.Validate(validTarget(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a hostname starting with a dash", func() {
			spec := baseSpec()
			spec.Hostname = "-prod-db"
			gomega.Expect(protovalidate.Validate(validTarget(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a hostname with spaces", func() {
			spec := baseSpec()
			spec.Hostname = "prod db"
			gomega.Expect(protovalidate.Validate(validTarget(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a hostname over 255 characters", func() {
			spec := baseSpec()
			spec.Hostname = strings.Repeat("a", 256)
			gomega.Expect(protovalidate.Validate(validTarget(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account_id", func() {
			spec := baseSpec()
			spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(validTarget(spec))).NotTo(gomega.BeNil())
		})
	})
})

package awswafipsetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAwsWafIpSetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsWafIpSetSpec Validation Suite")
}

// helper to create a minimal valid AwsWafIpSet wrapper.
func minimalIpSet(spec *AwsWafIpSetSpec) *AwsWafIpSet {
	return &AwsWafIpSet{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsWafIpSet",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-ip-set"},
		Spec:       spec,
	}
}

// helper to create a minimal valid spec.
func minimalSpec() *AwsWafIpSetSpec {
	return &AwsWafIpSetSpec{
		Region:           "us-west-2",
		Scope:            "REGIONAL",
		IpAddressVersion: "IPV4",
	}
}

var _ = ginkgo.Describe("AwsWafIpSetSpec validations", func() {

	ginkgo.It("accepts a minimal REGIONAL IPv4 set with no addresses", func() {
		gomega.Expect(protovalidate.Validate(minimalIpSet(minimalSpec()))).To(gomega.BeNil())
	})

	ginkgo.It("accepts CIDR addresses for both host and range entries", func() {
		spec := minimalSpec()
		spec.Addresses = []string{"203.0.113.0/24", "198.51.100.44/32"}
		spec.Description = "Corporate office egress ranges"
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).To(gomega.BeNil())
	})

	ginkgo.It("accepts an IPv6 set in CLOUDFRONT scope in us-east-1", func() {
		spec := minimalSpec()
		spec.Scope = "CLOUDFRONT"
		spec.Region = "us-east-1"
		spec.IpAddressVersion = "IPV6"
		spec.Addresses = []string{"2001:db8::/32"}
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).To(gomega.BeNil())
	})

	ginkgo.It("rejects CLOUDFRONT scope outside us-east-1", func() {
		spec := minimalSpec()
		spec.Scope = "CLOUDFRONT"
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid scope", func() {
		spec := minimalSpec()
		spec.Scope = "GLOBAL"
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid address version", func() {
		spec := minimalSpec()
		spec.IpAddressVersion = "DUAL"
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a bare address without a CIDR suffix", func() {
		spec := minimalSpec()
		spec.Addresses = []string{"192.0.2.44"}
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a missing region", func() {
		spec := minimalSpec()
		spec.Region = ""
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a description over 256 characters", func() {
		spec := minimalSpec()
		for i := 0; i < 26; i++ {
			spec.Description += "0123456789"
		}
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a description with characters outside AWS's WAF charset", func() {
		spec := minimalSpec()
		spec.Description = "Corporate ranges (E2E)"
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a two-character description", func() {
		spec := minimalSpec()
		spec.Description = "ab"
		gomega.Expect(protovalidate.Validate(minimalIpSet(spec))).NotTo(gomega.BeNil())
	})
})

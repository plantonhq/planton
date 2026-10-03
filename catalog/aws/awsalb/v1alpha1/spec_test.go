package awsalbv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsAlbSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsAlbSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidAlb is the common case: a region and two subnets in different
// availability zones.
func minimalValidAlb() *AwsAlb {
	return &AwsAlb{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsAlb",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "demo-alb",
		},
		Spec: &AwsAlbSpec{
			Region: "us-west-2",
			Subnets: []*foreignkeyv1.StringValueOrRef{
				literalRef("subnet-12345678"),
				literalRef("subnet-12345679"),
			},
		},
	}
}

var _ = ginkgo.Describe("AwsAlbSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_alb", func() {

			ginkgo.It("should not return a validation error for a minimal ALB", func() {
				err := protovalidate.Validate(minimalValidAlb())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a fully tuned ALB", func() {
				input := minimalValidAlb()
				input.Spec.SecurityGroups = []*foreignkeyv1.StringValueOrRef{literalRef("sg-0123456789abcdef0")}
				input.Spec.Internal = true
				input.Spec.IpAddressType = "dualstack"
				input.Spec.DeleteProtectionEnabled = true
				input.Spec.IdleTimeoutSeconds = 120
				input.Spec.ClientKeepAliveSeconds = 7200
				http2 := false
				input.Spec.Http2Enabled = &http2
				input.Spec.WafFailOpenEnabled = true
				input.Spec.ZonalShiftEnabled = true
				input.Spec.DropInvalidHeaderFields = true
				input.Spec.PreserveHostHeader = true
				input.Spec.XffClientPortEnabled = true
				input.Spec.XffHeaderProcessingMode = "preserve"
				input.Spec.DesyncMitigationMode = "strictest"
				input.Spec.TlsVersionAndCipherSuiteHeadersEnabled = true
				input.Spec.AccessLogs = &AwsAlbLogDelivery{
					Bucket: literalRef("demo-alb-logs"),
					Prefix: "alb/demo",
				}
				input.Spec.ConnectionLogs = &AwsAlbLogDelivery{Bucket: literalRef("demo-alb-logs")}
				input.Spec.HealthCheckLogs = &AwsAlbLogDelivery{Bucket: literalRef("demo-alb-logs")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a capacity reservation", func() {
				input := minimalValidAlb()
				input.Spec.MinimumLoadBalancerCapacityUnits = 500
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an IPAM-addressed ALB", func() {
				input := minimalValidAlb()
				input.Spec.Ipv4IpamPoolId = literalRef("ipam-pool-0123456789abcdef0")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with DNS configured", func() {
				input := minimalValidAlb()
				input.Spec.Dns = &AwsAlbDns{
					Enabled:       true,
					Route53ZoneId: literalRef("Z0123456789ABCDEF"),
					Hostnames:     []string{"app.example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_alb", func() {

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidAlb()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty", func() {
				input := minimalValidAlb()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error with fewer than two subnets", func() {
				input := minimalValidAlb()
				input.Spec.Subnets = []*foreignkeyv1.StringValueOrRef{literalRef("subnet-12345678")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid ip_address_type", func() {
				input := minimalValidAlb()
				input.Spec.IpAddressType = "ipv6"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range idle timeout", func() {
				input := minimalValidAlb()
				input.Spec.IdleTimeoutSeconds = 4001
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for negative capacity units", func() {
				input := minimalValidAlb()
				input.Spec.MinimumLoadBalancerCapacityUnits = -100
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range client keep-alive", func() {
				input := minimalValidAlb()
				input.Spec.ClientKeepAliveSeconds = 30
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid xff_header_processing_mode", func() {
				input := minimalValidAlb()
				input.Spec.XffHeaderProcessingMode = "rewrite"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid desync_mitigation_mode", func() {
				input := minimalValidAlb()
				input.Spec.DesyncMitigationMode = "aggressive"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for access logs without a bucket", func() {
				input := minimalValidAlb()
				input.Spec.AccessLogs = &AwsAlbLogDelivery{Prefix: "alb/demo"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate DNS hostnames", func() {
				input := minimalValidAlb()
				input.Spec.Dns = &AwsAlbDns{
					Enabled:       true,
					Route53ZoneId: literalRef("Z0123456789ABCDEF"),
					Hostnames:     []string{"app.example.com", "app.example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

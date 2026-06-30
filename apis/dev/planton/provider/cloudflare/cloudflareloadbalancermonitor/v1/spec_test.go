package cloudflareloadbalancermonitorv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func validMonitor() *CloudflareLoadBalancerMonitor {
	return &CloudflareLoadBalancerMonitor{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareLoadBalancerMonitor",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-monitor"},
		Spec: &CloudflareLoadBalancerMonitorSpec{
			AccountId:     validAccountID,
			Type:          CloudflareLoadBalancerMonitorType_https,
			Path:          "/healthz",
			ExpectedCodes: "2xx",
		},
	}
}

func TestCloudflareLoadBalancerMonitorSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareLoadBalancerMonitorSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareLoadBalancerMonitorSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal http monitor", func() {
			gomega.Expect(protovalidate.Validate(validMonitor())).To(gomega.BeNil())
		})

		ginkgo.It("accepts an http monitor with headers and tuning", func() {
			in := validMonitor()
			in.Spec.Headers = []*CloudflareLoadBalancerMonitorHeader{
				{Name: "Host", Values: []string{"app.example.com"}},
			}
			in.Spec.Interval = 30
			in.Spec.Timeout = 3
			in.Spec.Retries = 1
			in.Spec.ConsecutiveUp = 2
			in.Spec.ConsecutiveDown = 3
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a tcp monitor with a port", func() {
			in := validMonitor()
			in.Spec.Type = CloudflareLoadBalancerMonitorType_tcp
			in.Spec.Port = 5432
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an icmp_ping monitor without a port", func() {
			in := validMonitor()
			in.Spec.Type = CloudflareLoadBalancerMonitorType_icmp_ping
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			in := validMonitor()
			in.Spec.AccountId = "not-a-valid-account-id-string!!!"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a tcp monitor without a port", func() {
			in := validMonitor()
			in.Spec.Type = CloudflareLoadBalancerMonitorType_tcp
			in.Spec.Port = 0
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an smtp monitor without a port", func() {
			in := validMonitor()
			in.Spec.Type = CloudflareLoadBalancerMonitorType_smtp
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a port above the valid range", func() {
			in := validMonitor()
			in.Spec.Port = 70000
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a negative interval", func() {
			in := validMonitor()
			in.Spec.Interval = -1
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a header with no values", func() {
			in := validMonitor()
			in.Spec.Headers = []*CloudflareLoadBalancerMonitorHeader{{Name: "Host"}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})

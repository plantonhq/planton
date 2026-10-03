package azuretrafficmanagerprofilev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureTrafficManagerProfileSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureTrafficManagerProfileSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

func boolPtr(v bool) *bool { return &v }

// validResource returns a minimal valid Performance profile that
// individual cases mutate into the shape under test.
func validResource() *AzureTrafficManagerProfile {
	return &AzureTrafficManagerProfile{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureTrafficManagerProfile",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-profile",
		},
		Spec: &AzureTrafficManagerProfileSpec{
			ResourceGroup: literal("platform-rg"),
			Name:          "app-director",
			RoutingMethod: "Performance",
			DnsConfig: &AzureTrafficManagerDnsConfig{
				RelativeName: "contoso-app",
			},
			MonitorConfig: &AzureTrafficManagerMonitorConfig{
				Protocol: "HTTPS",
				Port:     int32Ptr(443),
			},
		},
	}
}

var _ = ginkgo.Describe("AzureTrafficManagerProfileSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_traffic_manager_profile", func() {

			ginkgo.It("should not return a validation error for the minimal Performance profile", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every routing method with its required companions", func() {
				for _, method := range []string{"Performance", "Priority", "Weighted", "Geographic", "Subnet"} {
					input := validResource()
					input.Spec.RoutingMethod = method
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "method %q should be valid", method)
				}
			})

			ginkgo.It("should accept a MultiValue profile carrying max_return", func() {
				input := validResource()
				input.Spec.RoutingMethod = "MultiValue"
				input.Spec.MaxReturn = int32Ptr(3)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the fast probe interval with an explicit narrowed timeout", func() {
				input := validResource()
				input.Spec.MonitorConfig.IntervalInSeconds = int32Ptr(10)
				input.Spec.MonitorConfig.TimeoutInSeconds = int32Ptr(7)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full monitor shape: path, ranges, headers, failures", func() {
				input := validResource()
				input.Spec.MonitorConfig.Protocol = "HTTP"
				input.Spec.MonitorConfig.Port = int32Ptr(80)
				input.Spec.MonitorConfig.Path = "/healthz"
				input.Spec.MonitorConfig.ExpectedStatusCodeRanges = []string{"200-299", "301-301"}
				input.Spec.MonitorConfig.ToleratedNumberOfFailures = int32Ptr(0)
				input.Spec.MonitorConfig.CustomHeaders = []*AzureTrafficManagerCustomHeader{
					{Name: "Host", Value: "app.contoso.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a parked profile with traffic view and tags", func() {
				input := validResource()
				input.Spec.Enabled = boolPtr(false)
				input.Spec.TrafficViewEnabled = true
				input.Spec.Tags = map[string]string{"cost-center": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit dns ttl", func() {
				input := validResource()
				input.Spec.DnsConfig.TtlSeconds = int32Ptr(30)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_traffic_manager_profile", func() {

			ginkgo.It("should return a validation error for an unknown routing method", func() {
				input := validResource()
				input.Spec.RoutingMethod = "RoundRobin"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a MultiValue profile without max_return", func() {
				input := validResource()
				input.Spec.RoutingMethod = "MultiValue"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "max_return")).To(gomega.BeTrue())
			})

			ginkgo.It("should return a validation error for max_return beyond 8", func() {
				input := validResource()
				input.Spec.RoutingMethod = "MultiValue"
				input.Spec.MaxReturn = int32Ptr(9)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when dns_config is missing", func() {
				input := validResource()
				input.Spec.DnsConfig = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when monitor_config is missing", func() {
				input := validResource()
				input.Spec.MonitorConfig = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an off-cadence probe interval", func() {
				input := validResource()
				input.Spec.MonitorConfig.IntervalInSeconds = int32Ptr(20)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for the fast interval without an explicit timeout", func() {
				input := validResource()
				input.Spec.MonitorConfig.IntervalInSeconds = int32Ptr(10)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "fast interval")).To(gomega.BeTrue())
			})

			ginkgo.It("should return a validation error for the fast interval with the unfitting default timeout", func() {
				input := validResource()
				input.Spec.MonitorConfig.IntervalInSeconds = int32Ptr(10)
				input.Spec.MonitorConfig.TimeoutInSeconds = int32Ptr(10)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a probe timeout below 5", func() {
				input := validResource()
				input.Spec.MonitorConfig.TimeoutInSeconds = int32Ptr(4)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed status-code range", func() {
				for _, r := range []string{"200-", "abc", "200_299", "20-299"} {
					input := validResource()
					input.Spec.MonitorConfig.ExpectedStatusCodeRanges = []string{r}
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil(), "range %q should be rejected", r)
				}
			})

			ginkgo.It("should return a validation error for a probe header without a value", func() {
				input := validResource()
				input.Spec.MonitorConfig.CustomHeaders = []*AzureTrafficManagerCustomHeader{{Name: "Host"}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a probe port of 0", func() {
				input := validResource()
				input.Spec.MonitorConfig.Port = int32Ptr(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for tolerated failures beyond 9", func() {
				input := validResource()
				input.Spec.MonitorConfig.ToleratedNumberOfFailures = int32Ptr(10)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the dns relative name is missing", func() {
				input := validResource()
				input.Spec.DnsConfig.RelativeName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

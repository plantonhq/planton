package azuretrafficmanagerendpointv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureTrafficManagerEndpointSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureTrafficManagerEndpointSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

func boolPtr(v bool) *bool { return &v }

const (
	testProfileId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/trafficManagerProfiles/app-director"
	testPublicIpId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/publicIPAddresses/app-eastus-pip"
)

// validResource returns a minimal valid external endpoint that
// individual cases mutate into the shape under test.
func validResource() *AzureTrafficManagerEndpoint {
	return &AzureTrafficManagerEndpoint{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureTrafficManagerEndpoint",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-endpoint",
		},
		Spec: &AzureTrafficManagerEndpointSpec{
			ProfileId: literal(testProfileId),
			Name:      "eastus-app",
			External: &AzureTrafficManagerExternalEndpoint{
				Target: literal("app-eastus.contoso.com"),
			},
		},
	}
}

var _ = ginkgo.Describe("AzureTrafficManagerEndpointSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_traffic_manager_endpoint", func() {

			ginkgo.It("should not return a validation error for the minimal external endpoint", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an azure endpoint targeting a public IP by ARM id", func() {
				input := validResource()
				input.Spec.External = nil
				input.Spec.Azure = &AzureTrafficManagerAzureEndpoint{
					TargetResourceId: literal(testPublicIpId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a nested endpoint composing a child profile", func() {
				input := validResource()
				input.Spec.External = nil
				input.Spec.Nested = &AzureTrafficManagerNestedEndpoint{
					TargetProfileId:       literal(testProfileId + "-child"),
					MinimumChildEndpoints: int32Ptr(1),
					EndpointLocation:      "eastus",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full shared surface: weight, priority, geo, subnets, headers", func() {
				input := validResource()
				input.Spec.Weight = int32Ptr(100)
				input.Spec.Priority = int32Ptr(1)
				input.Spec.Enabled = boolPtr(false)
				input.Spec.GeoMappings = []string{"GEO-NA", "US"}
				input.Spec.Subnets = []*AzureTrafficManagerEndpointSubnet{
					{First: "203.0.113.0", Scope: int32Ptr(24)},
					{First: "198.51.100.1", Last: "198.51.100.20"},
					{First: "192.0.2.7"},
				}
				input.Spec.CustomHeaders = []*AzureTrafficManagerEndpointCustomHeader{
					{Name: "Host", Value: "app-eastus.contoso.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an external endpoint with location and always-serve", func() {
				input := validResource()
				input.Spec.External.EndpointLocation = "eastus"
				input.Spec.External.AlwaysServeEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a nested endpoint with IPv4/IPv6 health floors", func() {
				input := validResource()
				input.Spec.External = nil
				input.Spec.Nested = &AzureTrafficManagerNestedEndpoint{
					TargetProfileId:                   literal(testProfileId + "-child"),
					MinimumChildEndpoints:             int32Ptr(2),
					MinimumRequiredChildEndpointsIpv4: int32Ptr(1),
					MinimumRequiredChildEndpointsIpv6: int32Ptr(1),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_traffic_manager_endpoint", func() {

			ginkgo.It("should return a validation error when no variant is set", func() {
				input := validResource()
				input.Spec.External = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "exactly one endpoint variant")).To(gomega.BeTrue())
			})

			ginkgo.It("should return a validation error when two variants are set", func() {
				input := validResource()
				input.Spec.Azure = &AzureTrafficManagerAzureEndpoint{
					TargetResourceId: literal(testPublicIpId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "exactly one endpoint variant")).To(gomega.BeTrue())
			})

			ginkgo.It("should return a validation error for an azure endpoint without a target", func() {
				input := validResource()
				input.Spec.External = nil
				input.Spec.Azure = &AzureTrafficManagerAzureEndpoint{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an external endpoint without a target", func() {
				input := validResource()
				input.Spec.External = &AzureTrafficManagerExternalEndpoint{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a nested endpoint without its child floor", func() {
				input := validResource()
				input.Spec.External = nil
				input.Spec.Nested = &AzureTrafficManagerNestedEndpoint{
					TargetProfileId: literal(testProfileId + "-child"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a child floor of 0", func() {
				input := validResource()
				input.Spec.External = nil
				input.Spec.Nested = &AzureTrafficManagerNestedEndpoint{
					TargetProfileId:       literal(testProfileId + "-child"),
					MinimumChildEndpoints: int32Ptr(0),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for weight outside 1-1000", func() {
				for _, w := range []int32{0, 1001} {
					input := validResource()
					input.Spec.Weight = int32Ptr(w)
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil(), "weight %d should be rejected", w)
				}
			})

			ginkgo.It("should return a validation error for priority beyond 1000", func() {
				input := validResource()
				input.Spec.Priority = int32Ptr(1001)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed subnet first address", func() {
				input := validResource()
				input.Spec.Subnets = []*AzureTrafficManagerEndpointSubnet{{First: "not-an-ip"}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a subnet scope beyond 32", func() {
				input := validResource()
				input.Spec.Subnets = []*AzureTrafficManagerEndpointSubnet{{First: "203.0.113.0", Scope: int32Ptr(33)}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an empty geo mapping code", func() {
				input := validResource()
				input.Spec.GeoMappings = []string{""}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a probe header without a value", func() {
				input := validResource()
				input.Spec.CustomHeaders = []*AzureTrafficManagerEndpointCustomHeader{{Name: "Host"}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the profile reference is missing", func() {
				input := validResource()
				input.Spec.ProfileId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the endpoint name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

package civoipaddressv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	civoprovider "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestCivoIpAddressSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoIpAddressSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoIpAddressSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal valid configuration", func() {
			ginkgo.It("should not return a validation error with just required region", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ip",
					},
					Spec: &CivoIpAddressSpec{
						Region: civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept configuration with description", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "Production web server IP",
						Region:      civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("different regions", func() {
			ginkgo.It("should accept LON1 region", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "london-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "London IP",
						Region:      civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept NYC1 region", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "newyork-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "New York IP",
						Region:      civoprovider.CivoRegion_nyc1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept FRA1 region", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "frankfurt-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "Frankfurt IP",
						Region:      civoprovider.CivoRegion_fra1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept PHX1 region", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "phoenix-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "Phoenix IP",
						Region:      civoprovider.CivoRegion_phx1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept MUM1 region", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "mumbai-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "Mumbai IP",
						Region:      civoprovider.CivoRegion_mum1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("description field variations", func() {
			ginkgo.It("should accept short description", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "short-desc-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "Web IP",
						Region:      civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept description at max length (100 chars)", func() {
				// Exactly 100 characters
				maxLengthDesc := strings.Repeat("a", 100)
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "max-desc-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: maxLengthDesc,
						Region:      civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept empty description (optional field)", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "empty-desc-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "",
						Region:      civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept description with special characters", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "special-chars-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "Prod-Web #1 (LoadBalancer) - us-east-1a",
						Region:      civoprovider.CivoRegion_nyc1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept description with numbers", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "numbered-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "Load Balancer 001 - Production Tier 1",
						Region:      civoprovider.CivoRegion_fra1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("real-world use cases", func() {
			ginkgo.It("should accept web server IP configuration", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "web-server-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "Production web server static IP",
						Region:      civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept load balancer IP configuration", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "k8s-lb-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "Kubernetes LoadBalancer ingress IP",
						Region:      civoprovider.CivoRegion_nyc1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept API gateway IP configuration", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "api-gateway-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "API Gateway static endpoint",
						Region:      civoprovider.CivoRegion_fra1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept VPN endpoint IP configuration", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "vpn-endpoint-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "WireGuard VPN endpoint",
						Region:      civoprovider.CivoRegion_mum1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required fields", func() {
			ginkgo.It("should return a validation error for missing region", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-region-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "IP without region",
						// Region is missing (required field)
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing metadata", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata:   nil, // Missing required metadata
					Spec: &CivoIpAddressSpec{
						Region: civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing spec", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-spec-ip",
					},
					Spec: nil, // Missing required spec
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("description validation", func() {
			ginkgo.It("should return a validation error for description exceeding max length", func() {
				// 101 characters (exceeds max_len = 100)
				tooLongDesc := strings.Repeat("a", 101)
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "long-desc-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: tooLongDesc,
						Region:      civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for very long description", func() {
				// 200 characters (way over limit)
				veryLongDesc := strings.Repeat("a", 200)
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "very-long-desc-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: veryLongDesc,
						Region:      civoprovider.CivoRegion_nyc1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid API version and kind", func() {
			ginkgo.It("should return a validation error for wrong api_version", func() {
				input := &CivoIpAddress{
					ApiVersion: "wrong.api.version/v1", // Wrong value
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-api-ip",
					},
					Spec: &CivoIpAddressSpec{
						Region: civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for wrong kind", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "WrongKind", // Wrong kind
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-kind-ip",
					},
					Spec: &CivoIpAddressSpec{
						Region: civoprovider.CivoRegion_lon1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid region values", func() {
			ginkgo.It("should return a validation error for unspecified region", func() {
				input := &CivoIpAddress{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoIpAddress",
					Metadata: &shared.CloudResourceMetadata{
						Name: "unspecified-region-ip",
					},
					Spec: &CivoIpAddressSpec{
						Description: "IP with unspecified region",
						Region:      civoprovider.CivoRegion_civo_region_unspecified,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

package gcprouternatv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestGcpRouterNatSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpRouterNatSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpRouterNatSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal valid fields", func() {
			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "test-nat",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with VPC self-link (full URL)", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-vpc",
							},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "test-nat",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with VPC reference", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name:      "test-vpc",
									FieldPath: "status.outputs.network_self_link",
								},
							},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "test-nat",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with specific subnets", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "test-nat",
						SubnetworkSelfLinks: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "projects/test-project/regions/us-central1/subnetworks/subnet-1",
								},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "projects/test-project/regions/us-central1/subnetworks/subnet-2",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with manual NAT IPs", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "test-nat",
						NatIpNames: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "nat-ip-1"},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "nat-ip-2"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with log filter ERRORS_ONLY", func() {
			ginkgo.It("should not return a validation error", func() {
				logFilter := GcpRouterNatLogFilter_ERRORS_ONLY
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "test-nat",
						LogFilter:  &logFilter,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with log filter ALL", func() {
			ginkgo.It("should not return a validation error", func() {
				logFilter := GcpRouterNatLogFilter_ALL
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "test-nat",
						LogFilter:  &logFilter,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with log filter DISABLED", func() {
			ginkgo.It("should not return a validation error", func() {
				logFilter := GcpRouterNatLogFilter_DISABLED
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "test-nat",
						LogFilter:  &logFilter,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with all optional fields", func() {
			ginkgo.It("should not return a validation error", func() {
				logFilter := GcpRouterNatLogFilter_ERRORS_ONLY
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
						Org:  "test-org",
						Env:  "test-env",
					},
					Spec: &GcpRouterNatSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "test-nat",
						SubnetworkSelfLinks: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "projects/test-project/regions/us-central1/subnetworks/subnet-1",
								},
							},
						},
						NatIpNames: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "nat-ip-1"},
							},
						},
						LogFilter: &logFilter,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing vpc_self_link", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						Region: "us-central1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing region", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Spec: &GcpRouterNatSpec{
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region: "us-central1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing spec", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing router_name", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:  "us-central1",
						NatName: "test-nat",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing nat_name", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "test-router",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid router_name format", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "INVALID-ROUTER", // uppercase not allowed
						NatName:    "test-nat",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid nat_name format", func() {
			ginkgo.It("should return a validation error", func() {
				input := &GcpRouterNat{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpRouterNat",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-router-nat",
					},
					Spec: &GcpRouterNatSpec{
						VpcSelfLink: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/test-project-123/global/networks/test-vpc"},
						},
						Region:     "us-central1",
						RouterName: "test-router",
						NatName:    "INVALID-NAT", // uppercase not allowed
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

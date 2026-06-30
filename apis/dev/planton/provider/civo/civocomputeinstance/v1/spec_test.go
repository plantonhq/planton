package civocomputeinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkey "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestCivoComputeInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoComputeInstanceSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoComputeInstanceSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("compute instance with minimal valid fields", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "test-instance",
						Region:       civo.CivoRegion_nyc1,
						Size:         "g3.small",
						Image:        "ubuntu-jammy",
						Network: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "network-123"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-web",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "prod-web",
						Region:       civo.CivoRegion_lon1,
						Size:         "g3.medium",
						Image:        "ubuntu-jammy",
						Network: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "network-456"},
						},
						SshKeyIds: []string{"ssh-key-789"},
						FirewallIds: []*foreignkey.StringValueOrRef{
							{LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "firewall-abc"}},
						},
						VolumeIds: []*foreignkey.StringValueOrRef{
							{LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "volume-def"}},
						},
						ReservedIpId: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "ip-ghi"},
						},
						Tags:     []string{"env:prod", "service:web"},
						UserData: "#!/bin/bash\napt-get update",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("instance_name validation", func() {

			ginkgo.It("should return a validation error when instance_name is empty", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "",
						Region:       civo.CivoRegion_nyc1,
						Size:         "g3.small",
						Image:        "ubuntu-jammy",
						Network: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "network-123"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when instance_name contains uppercase", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "Test-Instance",
						Region:       civo.CivoRegion_nyc1,
						Size:         "g3.small",
						Image:        "ubuntu-jammy",
						Network: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "network-123"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when instance_name is too long", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "this-is-a-very-long-instance-name-that-exceeds-sixty-three-characters-limit",
						Region:       civo.CivoRegion_nyc1,
						Size:         "g3.small",
						Image:        "ubuntu-jammy",
						Network: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "network-123"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("required fields validation", func() {

			ginkgo.It("should return a validation error when region is not set", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "test-instance",
						Size:         "g3.small",
						Image:        "ubuntu-jammy",
						Network: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "network-123"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size is empty", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "test-instance",
						Region:       civo.CivoRegion_nyc1,
						Size:         "",
						Image:        "ubuntu-jammy",
						Network: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "network-123"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network is not set", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "test-instance",
						Region:       civo.CivoRegion_nyc1,
						Size:         "g3.small",
						Image:        "ubuntu-jammy",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("user_data validation", func() {

			ginkgo.It("should return a validation error when user_data exceeds 32KB", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "test-instance",
						Region:       civo.CivoRegion_nyc1,
						Size:         "g3.small",
						Image:        "ubuntu-jammy",
						Network: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "network-123"},
						},
						UserData: string(make([]byte, 33000)), // Exceeds 32KB limit
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("tags validation", func() {

			ginkgo.It("should return a validation error when tags are not unique", func() {
				input := &CivoComputeInstance{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoComputeInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
					Spec: &CivoComputeInstanceSpec{
						InstanceName: "test-instance",
						Region:       civo.CivoRegion_nyc1,
						Size:         "g3.small",
						Image:        "ubuntu-jammy",
						Network: &foreignkey.StringValueOrRef{
							LiteralOrRef: &foreignkey.StringValueOrRef_Value{Value: "network-123"},
						},
						Tags: []string{"env:prod", "env:prod"}, // duplicate tags
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

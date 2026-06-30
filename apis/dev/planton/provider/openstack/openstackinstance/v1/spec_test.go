package openstackinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOpenStackInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackInstanceSpec Validation Tests")
}

// newStringValueOrRef is a helper to create a literal StringValueOrRef.
func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

// newValueFromRef is a helper to create a value_from StringValueOrRef.
func newValueFromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Name: name,
			},
		},
	}
}

// minimalValidInstance returns a minimal valid OpenStackInstance with
// flavor_name, image_name, and one network attachment.
func minimalValidInstance() *OpenStackInstance {
	return &OpenStackInstance{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackInstance",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-instance",
		},
		Spec: &OpenStackInstanceSpec{
			FlavorName: "m1.medium",
			ImageName:  "ubuntu-22.04",
			Networks: []*InstanceNetwork{
				{
					Uuid: newStringValueOrRef("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
				},
			},
		},
	}
}

var _ = ginkgo.Describe("OpenStackInstanceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_instance", func() {

			ginkgo.It("should not return a validation error for minimal valid instance (flavor_name + image_name + network)", func() {
				input := minimalValidInstance()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when using flavor_id instead of flavor_name", func() {
				input := minimalValidInstance()
				input.Spec.FlavorName = ""
				input.Spec.FlavorId = "12345-flavor-uuid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when using image_id instead of image_name", func() {
				input := minimalValidInstance()
				input.Spec.ImageName = ""
				input.Spec.ImageId = "12345-image-uuid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with key_pair FK literal", func() {
				input := minimalValidInstance()
				input.Spec.KeyPair = newStringValueOrRef("my-keypair")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with key_pair FK value_from", func() {
				input := minimalValidInstance()
				input.Spec.KeyPair = newValueFromRef("my-keypair")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with security_groups FK", func() {
				input := minimalValidInstance()
				input.Spec.SecurityGroups = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("web-sg"),
					newValueFromRef("ssh-sg"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with port-based network", func() {
				input := minimalValidInstance()
				input.Spec.Networks = []*InstanceNetwork{
					{
						Port: newStringValueOrRef("port-uuid-here"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with network value_from refs", func() {
				input := minimalValidInstance()
				input.Spec.Networks = []*InstanceNetwork{
					{
						Uuid: newValueFromRef("my-network"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with multiple networks", func() {
				input := minimalValidInstance()
				input.Spec.Networks = []*InstanceNetwork{
					{
						Uuid:          newValueFromRef("app-network"),
						AccessNetwork: true,
					},
					{
						Uuid: newValueFromRef("mgmt-network"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with network and fixed_ip_v4", func() {
				input := minimalValidInstance()
				input.Spec.Networks = []*InstanceNetwork{
					{
						Uuid:      newStringValueOrRef("network-uuid"),
						FixedIpV4: "192.168.1.100",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with block_device (boot from volume)", func() {
				input := minimalValidInstance()
				input.Spec.ImageName = ""
				input.Spec.ImageId = ""
				input.Spec.BlockDevice = []*BlockDevice{
					{
						SourceType:          "image",
						Uuid:                "image-uuid",
						DestinationType:     "volume",
						BootIndex:           0,
						VolumeSize:          20,
						DeleteOnTermination: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with multiple block_devices", func() {
				input := minimalValidInstance()
				input.Spec.BlockDevice = []*BlockDevice{
					{
						SourceType:          "image",
						Uuid:                "image-uuid",
						DestinationType:     "volume",
						BootIndex:           0,
						VolumeSize:          20,
						DeleteOnTermination: true,
					},
					{
						SourceType:      "blank",
						DestinationType: "volume",
						BootIndex:       -1,
						VolumeSize:      100,
						VolumeType:      "SSD",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with user_data", func() {
				input := minimalValidInstance()
				input.Spec.UserData = "#!/bin/bash\necho hello"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with metadata", func() {
				input := minimalValidInstance()
				input.Spec.Metadata = map[string]string{
					"environment": "dev",
					"team":        "platform",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with config_drive enabled", func() {
				input := minimalValidInstance()
				configDrive := true
				input.Spec.ConfigDrive = &configDrive
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with server_group_id FK", func() {
				input := minimalValidInstance()
				input.Spec.ServerGroupId = newValueFromRef("ha-group")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with availability_zone", func() {
				input := minimalValidInstance()
				input.Spec.AvailabilityZone = "nova"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with tags", func() {
				input := minimalValidInstance()
				input.Spec.Tags = []string{"env:dev", "team:platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for instance with region override", func() {
				input := minimalValidInstance()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for fully-specified instance", func() {
				configDrive := true
				input := &OpenStackInstance{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-instance",
						Org:  "acme-corp",
						Env:  "production",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &OpenStackInstanceSpec{
						FlavorName: "m1.xlarge",
						ImageName:  "ubuntu-22.04",
						KeyPair:    newValueFromRef("prod-keypair"),
						Networks: []*InstanceNetwork{
							{
								Port:          newValueFromRef("app-port"),
								AccessNetwork: true,
							},
						},
						SecurityGroups: []*foreignkeyv1.StringValueOrRef{
							newValueFromRef("web-sg"),
							newValueFromRef("ssh-sg"),
						},
						BlockDevice: []*BlockDevice{
							{
								SourceType:          "image",
								Uuid:                "image-uuid",
								DestinationType:     "volume",
								BootIndex:           0,
								VolumeSize:          50,
								DeleteOnTermination: true,
								VolumeType:          "SSD",
							},
						},
						UserData:         "#!/bin/bash\napt-get update",
						Metadata:         map[string]string{"role": "webserver"},
						ConfigDrive:      &configDrive,
						ServerGroupId:    newValueFromRef("ha-group"),
						AvailabilityZone: "nova",
						Tags:             []string{"production", "managed"},
						Region:           "RegionOne",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_instance", func() {

			ginkgo.It("should return a validation error when both flavor_name and flavor_id are set", func() {
				input := minimalValidInstance()
				input.Spec.FlavorId = "flavor-uuid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither flavor_name nor flavor_id is set", func() {
				input := minimalValidInstance()
				input.Spec.FlavorName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when networks is empty", func() {
				input := minimalValidInstance()
				input.Spec.Networks = []*InstanceNetwork{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network has both uuid and port", func() {
				input := minimalValidInstance()
				input.Spec.Networks = []*InstanceNetwork{
					{
						Uuid: newStringValueOrRef("network-uuid"),
						Port: newStringValueOrRef("port-uuid"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network has neither uuid nor port", func() {
				input := minimalValidInstance()
				input.Spec.Networks = []*InstanceNetwork{
					{
						FixedIpV4: "192.168.1.100",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when block_device has invalid source_type", func() {
				input := minimalValidInstance()
				input.Spec.BlockDevice = []*BlockDevice{
					{
						SourceType: "invalid",
						Uuid:       "some-uuid",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when block_device has empty source_type", func() {
				input := minimalValidInstance()
				input.Spec.BlockDevice = []*BlockDevice{
					{
						SourceType: "",
						Uuid:       "some-uuid",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when tags contain duplicates", func() {
				input := minimalValidInstance()
				input.Spec.Tags = []string{"env:dev", "env:dev"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidInstance()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidInstance()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidInstance()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackInstance{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-instance",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

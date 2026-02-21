package ocicomputeinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciComputeInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciComputeInstanceSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidInstance() *OciComputeInstance {
	return &OciComputeInstance{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciComputeInstance",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-instance",
		},
		Spec: &OciComputeInstanceSpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			AvailabilityDomain: "Ixxj:US-ASHBURN-AD-1",
			Shape:              "VM.Standard.E4.Flex",
			SourceDetails: &OciComputeInstanceSpec_SourceDetails{
				SourceType: OciComputeInstanceSpec_SourceDetails_image,
				SourceId:   "ocid1.image.oc1.iad.example",
			},
			CreateVnicDetails: &OciComputeInstanceSpec_CreateVnicDetails{
				SubnetId: newStringValueOrRef("ocid1.subnet.oc1.iad.example"),
			},
		},
	}
}

var _ = ginkgo.Describe("OciComputeInstanceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_compute_instance", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidInstance()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name set", func() {
				input := minimalValidInstance()
				input.Spec.DisplayName = "My Web Server"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with flex shape config", func() {
				input := minimalValidInstance()
				ocpus := float32(2.0)
				mem := float32(16.0)
				input.Spec.ShapeConfig = &OciComputeInstanceSpec_ShapeConfig{
					Ocpus:       &ocpus,
					MemoryInGbs: &mem,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with burstable shape config", func() {
				input := minimalValidInstance()
				ocpus := float32(1.0)
				mem := float32(8.0)
				input.Spec.ShapeConfig = &OciComputeInstanceSpec_ShapeConfig{
					Ocpus:                   &ocpus,
					MemoryInGbs:             &mem,
					BaselineOcpuUtilization: "BASELINE_1_8",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full VNIC details", func() {
				input := minimalValidInstance()
				assignPublic := true
				skipSrcDst := false
				assignDns := true
				input.Spec.CreateVnicDetails = &OciComputeInstanceSpec_CreateVnicDetails{
					SubnetId:               newStringValueOrRef("ocid1.subnet.oc1.iad.example"),
					NsgIds:                 []*foreignkeyv1.StringValueOrRef{newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.example")},
					AssignPublicIp:         &assignPublic,
					DisplayName:            "primary-vnic",
					HostnameLabel:          "webserver1",
					PrivateIp:              "10.0.1.50",
					SkipSourceDestCheck:    &skipSrcDst,
					AssignPrivateDnsRecord: &assignDns,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with metadata (SSH keys and user_data)", func() {
				input := minimalValidInstance()
				input.Spec.Metadata = map[string]string{
					"ssh_authorized_keys": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
					"user_data":           "IyEvYmluL2Jhc2g=",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with boot volume source type", func() {
				input := minimalValidInstance()
				input.Spec.SourceDetails = &OciComputeInstanceSpec_SourceDetails{
					SourceType: OciComputeInstanceSpec_SourceDetails_boot_volume,
					SourceId:   "ocid1.bootvolume.oc1.iad.example",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with source details including boot volume size and KMS key", func() {
				input := minimalValidInstance()
				bvSize := int64(100)
				bvVpus := int64(20)
				input.Spec.SourceDetails = &OciComputeInstanceSpec_SourceDetails{
					SourceType:          OciComputeInstanceSpec_SourceDetails_image,
					SourceId:            "ocid1.image.oc1.iad.example",
					BootVolumeSizeInGbs: &bvSize,
					BootVolumeVpusPerGb: &bvVpus,
					KmsKeyId:            newStringValueOrRef("ocid1.key.oc1.iad.example"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with agent config", func() {
				input := minimalValidInstance()
				mgmtDisabled := true
				input.Spec.AgentConfig = &OciComputeInstanceSpec_AgentConfig{
					IsManagementDisabled: &mgmtDisabled,
					PluginsConfig: []*OciComputeInstanceSpec_AgentConfig_PluginConfig{
						{
							Name:         "Vulnerability Scanning",
							DesiredState: OciComputeInstanceSpec_AgentConfig_PluginConfig_enabled,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with preemptible config", func() {
				input := minimalValidInstance()
				preserveBoot := true
				input.Spec.PreemptibleInstanceConfig = &OciComputeInstanceSpec_PreemptibleInstanceConfig{
					PreserveBootVolume: &preserveBoot,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with platform config (VM)", func() {
				input := minimalValidInstance()
				secureBoot := true
				measuredBoot := true
				input.Spec.PlatformConfig = &OciComputeInstanceSpec_PlatformConfig{
					Type:                  OciComputeInstanceSpec_PlatformConfig_amd_vm,
					IsSecureBootEnabled:   &secureBoot,
					IsMeasuredBootEnabled: &measuredBoot,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional top-level fields", func() {
				input := minimalValidInstance()
				pvEncrypt := true
				liveMigrate := true
				legacyImds := true
				input.Spec.FaultDomain = "FAULT-DOMAIN-2"
				input.Spec.IsPvEncryptionInTransitEnabled = &pvEncrypt
				input.Spec.AvailabilityConfig = &OciComputeInstanceSpec_AvailabilityConfig{
					IsLiveMigrationPreferred: &liveMigrate,
					RecoveryAction:           OciComputeInstanceSpec_AvailabilityConfig_restore_instance,
				}
				input.Spec.InstanceOptions = &OciComputeInstanceSpec_InstanceOptions{
					AreLegacyImdsEndpointsDisabled: &legacyImds,
				}
				input.Spec.CapacityReservationId = newStringValueOrRef("ocid1.capacityreservation.oc1.iad.example")
				input.Spec.DedicatedVmHostId = newStringValueOrRef("ocid1.dedicatedvmhost.oc1.iad.example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with launch options", func() {
				input := minimalValidInstance()
				pvEncryptLaunch := true
				consistentNaming := true
				input.Spec.LaunchOptions = &OciComputeInstanceSpec_LaunchOptions{
					BootVolumeType:                  "PARAVIRTUALIZED",
					NetworkType:                     "PARAVIRTUALIZED",
					Firmware:                        OciComputeInstanceSpec_LaunchOptions_uefi_64,
					IsPvEncryptionInTransitEnabled:  &pvEncryptLaunch,
					IsConsistentVolumeNamingEnabled: &consistentNaming,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidInstance()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with subnet_id via value_from ref", func() {
				input := minimalValidInstance()
				input.Spec.CreateVnicDetails.SubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-subnet",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with 5 NSG IDs (max allowed)", func() {
				input := minimalValidInstance()
				input.Spec.CreateVnicDetails.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.iad.1"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.2"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.3"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.4"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.5"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_compute_instance", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidInstance()
				input.ApiVersion = "wrong.openmcf.org/v1"
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
				input := &OciComputeInstance{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciComputeInstance",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-instance"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidInstance()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when availability_domain is empty", func() {
				input := minimalValidInstance()
				input.Spec.AvailabilityDomain = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape is empty", func() {
				input := minimalValidInstance()
				input.Spec.Shape = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when source_details is missing", func() {
				input := minimalValidInstance()
				input.Spec.SourceDetails = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when source_type is unspecified", func() {
				input := minimalValidInstance()
				input.Spec.SourceDetails.SourceType = OciComputeInstanceSpec_SourceDetails_source_type_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when source_id is empty", func() {
				input := minimalValidInstance()
				input.Spec.SourceDetails.SourceId = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when create_vnic_details is missing", func() {
				input := minimalValidInstance()
				input.Spec.CreateVnicDetails = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := minimalValidInstance()
				input.Spec.CreateVnicDetails.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when nsg_ids exceeds 5", func() {
				input := minimalValidInstance()
				input.Spec.CreateVnicDetails.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.iad.1"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.2"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.3"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.4"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.5"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.6"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when agent_config plugin name is empty", func() {
				input := minimalValidInstance()
				input.Spec.AgentConfig = &OciComputeInstanceSpec_AgentConfig{
					PluginsConfig: []*OciComputeInstanceSpec_AgentConfig_PluginConfig{
						{
							Name:         "",
							DesiredState: OciComputeInstanceSpec_AgentConfig_PluginConfig_enabled,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when agent_config plugin desired_state is unspecified", func() {
				input := minimalValidInstance()
				input.Spec.AgentConfig = &OciComputeInstanceSpec_AgentConfig{
					PluginsConfig: []*OciComputeInstanceSpec_AgentConfig_PluginConfig{
						{
							Name:         "Vulnerability Scanning",
							DesiredState: OciComputeInstanceSpec_AgentConfig_PluginConfig_desired_state_unspecified,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when platform_config type is unspecified", func() {
				input := minimalValidInstance()
				secureBoot := true
				input.Spec.PlatformConfig = &OciComputeInstanceSpec_PlatformConfig{
					Type:                OciComputeInstanceSpec_PlatformConfig_platform_type_unspecified,
					IsSecureBootEnabled: &secureBoot,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

package awslaunchtemplatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsLaunchTemplateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsLaunchTemplateSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidLaunchTemplate is the common case: an AMI + instance type
// blueprint (the shape every ASG-backed web fleet starts from).
func minimalValidLaunchTemplate() *AwsLaunchTemplate {
	return &AwsLaunchTemplate{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsLaunchTemplate",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "web",
		},
		Spec: &AwsLaunchTemplateSpec{
			Region:       "us-west-2",
			ImageId:      "ami-0123456789abcdef0",
			InstanceType: "t3.small",
		},
	}
}

var _ = ginkgo.Describe("AwsLaunchTemplateSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_launch_template", func() {

			ginkgo.It("should not return a validation error for a minimal launch template", func() {
				err := protovalidate.Validate(minimalValidLaunchTemplate())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a hardened web-server template", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.Description = "amzn2023 + IMDSv2 + gp3"
				input.Spec.InstanceProfile = literalRef("arn:aws:iam::123456789012:instance-profile/web")
				input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{literalRef("sg-0123456789abcdef0")}
				input.Spec.EbsOptimized = true
				input.Spec.DetailedMonitoring = true
				input.Spec.UserData = "#!/bin/bash\necho hello"
				input.Spec.MetadataOptions = &AwsLaunchTemplateMetadataOptions{
					HttpEndpoint:            "enabled",
					HttpTokens:              "required",
					HttpPutResponseHopLimit: 2,
					InstanceMetadataTags:    "enabled",
				}
				deleteOnTermination := true
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName: "/dev/xvda",
						Ebs: &AwsLaunchTemplateEbs{
							VolumeSizeGb:        50,
							VolumeType:          "gp3",
							Iops:                4000,
							ThroughputMibps:     250,
							Encrypted:           true,
							KmsKeyId:            literalRef("arn:aws:kms:us-west-2:123456789012:key/abc"),
							DeleteOnTermination: &deleteOnTermination,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an attribute-based spot template", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.InstanceType = ""
				input.Spec.InstanceRequirements = &AwsLaunchTemplateInstanceRequirements{
					MemoryMib:                             &AwsLaunchTemplateIntRange{Min: 4096, Max: 16384},
					VcpuCount:                             &AwsLaunchTemplateIntRange{Min: 2, Max: 8},
					CpuManufacturers:                      []string{"intel", "amd"},
					InstanceGenerations:                   []string{"current"},
					BareMetal:                             "excluded",
					BurstablePerformance:                  "included",
					SpotMaxPricePercentageOverLowestPrice: 100,
				}
				input.Spec.SpotOptions = &AwsLaunchTemplateSpotOptions{
					SpotInstanceType:             "one-time",
					InstanceInterruptionBehavior: "terminate",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an explicit network interface", func() {
				input := minimalValidLaunchTemplate()
				associatePublicIp := false
				input.Spec.NetworkInterfaces = []*AwsLaunchTemplateNetworkInterface{
					{
						DeviceIndex:              0,
						InterfaceType:            "interface",
						AssociatePublicIpAddress: &associatePublicIp,
						SubnetId:                 literalRef("subnet-0123456789abcdef0"),
						SecurityGroupIds:         []*foreignkeyv1.StringValueOrRef{literalRef("sg-0123456789abcdef0")},
						Ipv4PrefixCount:          4,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a reserved-fleet capacity-block template", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.MarketType = "capacity-block"
				input.Spec.CapacityReservation = &AwsLaunchTemplateCapacityReservation{
					CapacityReservationId: "cr-0123456789abcdef0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a dedicated-host BYOL template", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.Placement = &AwsLaunchTemplatePlacement{
					Tenancy:  "host",
					HostId:   "h-0123456789abcdef0",
					Affinity: "host",
				}
				input.Spec.LicenseConfigurationArns = []string{
					"arn:aws:license-manager:us-west-2:123456789012:license-configuration:lic-abc",
				}
				input.Spec.CpuOptions = &AwsLaunchTemplateCpuOptions{
					CoreCount:            8,
					ThreadsPerCore:       1,
					NestedVirtualization: "enabled",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for advanced interface networking", func() {
				input := minimalValidLaunchTemplate()
				primaryIpv6 := true
				carrierIp := true
				input.Spec.BandwidthWeighting = "vpc-1"
				input.Spec.NetworkInterfaces = []*AwsLaunchTemplateNetworkInterface{
					{
						DeviceIndex: 0,
						ConnectionTracking: &AwsLaunchTemplateConnectionTracking{
							TcpEstablishedTimeoutSeconds: 3600,
							UdpStreamTimeoutSeconds:      120,
							UdpTimeoutSeconds:            45,
						},
						EnaSrd: &AwsLaunchTemplateEnaSrd{
							Enabled:    true,
							UdpEnabled: true,
						},
						EnaQueueCount:             4,
						PrimaryIpv6:               &primaryIpv6,
						AssociateCarrierIpAddress: &carrierIp,
					},
				}
				input.Spec.SecondaryInterfaces = []*AwsLaunchTemplateSecondaryInterface{
					{
						DeviceIndex:           1,
						NetworkCardIndex:      1,
						DeleteOnTermination:   true,
						SecondarySubnetId:     literalRef("subnet-0123456789abcdef1"),
						PrivateIpAddressCount: 2,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a snapshot-hydrated volume", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName: "/dev/xvdb",
						Ebs: &AwsLaunchTemplateEbs{
							SnapshotId:                    "snap-0123456789abcdef0",
							VolumeType:                    "gp3",
							VolumeInitializationRateMibps: 200,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for placement and cpu tuning", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.Placement = &AwsLaunchTemplatePlacement{
					GroupName: "cluster-pg",
					Tenancy:   "default",
				}
				input.Spec.CpuOptions = &AwsLaunchTemplateCpuOptions{
					CoreCount:      8,
					ThreadsPerCore: 1,
				}
				input.Spec.CpuCredits = "standard"
				input.Spec.AutoRecovery = "default"
				input.Spec.InstanceInitiatedShutdownBehavior = "terminate"
				input.Spec.PrivateDnsNameOptions = &AwsLaunchTemplatePrivateDnsNameOptions{
					HostnameType:                 "resource-name",
					EnableResourceNameDnsARecord: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a partial template without image or type", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.ImageId = ""
				input.Spec.InstanceType = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("spec-level rules", func() {

			ginkgo.It("should return an error when region is empty", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when instance_type and instance_requirements are both set", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.InstanceRequirements = &AwsLaunchTemplateInstanceRequirements{
					MemoryMib: &AwsLaunchTemplateIntRange{Min: 4096},
					VcpuCount: &AwsLaunchTemplateIntRange{Min: 2},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a malformed image id", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.ImageId = "img-123"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an invalid cpu_credits value", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.CpuCredits = "turbo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an invalid auto_recovery value", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.AutoRecovery = "on"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an invalid shutdown behavior", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.InstanceInitiatedShutdownBehavior = "hibernate"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when enclave and hibernation are both enabled", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.EnclaveEnabled = true
				input.Spec.HibernationEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when top-level security groups combine with explicit interfaces", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{literalRef("sg-0123456789abcdef0")}
				input.Spec.NetworkInterfaces = []*AwsLaunchTemplateNetworkInterface{
					{DeviceIndex: 0},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("instance requirements rules", func() {

			ginkgo.It("should return an error when memory_mib is missing", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.InstanceType = ""
				input.Spec.InstanceRequirements = &AwsLaunchTemplateInstanceRequirements{
					VcpuCount: &AwsLaunchTemplateIntRange{Min: 2},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when allowed and excluded instance types are both set", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.InstanceType = ""
				input.Spec.InstanceRequirements = &AwsLaunchTemplateInstanceRequirements{
					MemoryMib:             &AwsLaunchTemplateIntRange{Min: 4096},
					VcpuCount:             &AwsLaunchTemplateIntRange{Min: 2},
					AllowedInstanceTypes:  []string{"m5.*"},
					ExcludedInstanceTypes: []string{"t2.*"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when both spot price protections are set", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.InstanceType = ""
				input.Spec.InstanceRequirements = &AwsLaunchTemplateInstanceRequirements{
					MemoryMib:                             &AwsLaunchTemplateIntRange{Min: 4096},
					VcpuCount:                             &AwsLaunchTemplateIntRange{Min: 2},
					SpotMaxPricePercentageOverLowestPrice: 100,
					MaxSpotPriceAsPercentageOfOptimalOnDemandPrice: 80,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an inverted integer range", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.InstanceType = ""
				input.Spec.InstanceRequirements = &AwsLaunchTemplateInstanceRequirements{
					MemoryMib: &AwsLaunchTemplateIntRange{Min: 8192, Max: 4096},
					VcpuCount: &AwsLaunchTemplateIntRange{Min: 2},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an invalid bare_metal value", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.InstanceType = ""
				input.Spec.InstanceRequirements = &AwsLaunchTemplateInstanceRequirements{
					MemoryMib: &AwsLaunchTemplateIntRange{Min: 4096},
					VcpuCount: &AwsLaunchTemplateIntRange{Min: 2},
					BareMetal: "yes",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("block device mapping rules", func() {

			ginkgo.It("should return an error when device_name is missing", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{Ebs: &AwsLaunchTemplateEbs{VolumeSizeGb: 50}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when ebs combines with virtual_name", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName:  "/dev/sdf",
						VirtualName: "ephemeral0",
						Ebs:         &AwsLaunchTemplateEbs{VolumeSizeGb: 50},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an invalid volume type", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName: "/dev/xvda",
						Ebs:        &AwsLaunchTemplateEbs{VolumeType: "gp4"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for throughput on a non-gp3 volume", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName: "/dev/xvda",
						Ebs:        &AwsLaunchTemplateEbs{VolumeType: "gp2", ThroughputMibps: 250},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should accept gp3 throughput in the raised 1001-2000 range", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName: "/dev/xvda",
						Ebs:        &AwsLaunchTemplateEbs{VolumeType: "gp3", ThroughputMibps: 1500},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return an error for gp3 throughput above 2000", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName: "/dev/xvda",
						Ebs:        &AwsLaunchTemplateEbs{VolumeType: "gp3", ThroughputMibps: 2500},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for iops on an unsupported volume type", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName: "/dev/xvda",
						Ebs:        &AwsLaunchTemplateEbs{VolumeType: "st1", Iops: 4000},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("network interface rules", func() {

			ginkgo.It("should return an error for an invalid interface type", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.NetworkInterfaces = []*AwsLaunchTemplateNetworkInterface{
					{InterfaceType: "trunk"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when ipv4 count combines with explicit addresses", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.NetworkInterfaces = []*AwsLaunchTemplateNetworkInterface{
					{Ipv4AddressCount: 2, Ipv4Addresses: []string{"10.0.1.5"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when ipv6 prefix count combines with explicit prefixes", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.NetworkInterfaces = []*AwsLaunchTemplateNetworkInterface{
					{Ipv6PrefixCount: 2, Ipv6Prefixes: []string{"2600:1f13::/80"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("metadata options rules", func() {

			ginkgo.It("should return an error for an invalid http_tokens value", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.MetadataOptions = &AwsLaunchTemplateMetadataOptions{HttpTokens: "v2"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for a hop limit out of range", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.MetadataOptions = &AwsLaunchTemplateMetadataOptions{HttpPutResponseHopLimit: 65}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("placement and cpu rules", func() {

			ginkgo.It("should return an error for an invalid tenancy", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.Placement = &AwsLaunchTemplatePlacement{Tenancy: "shared"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for invalid threads_per_core", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.CpuOptions = &AwsLaunchTemplateCpuOptions{ThreadsPerCore: 4}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when group_name and group_id are both set", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.Placement = &AwsLaunchTemplatePlacement{
					GroupName: "cluster-pg",
					GroupId:   "pg-0123456789abcdef0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when host targeting is used without host tenancy", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.Placement = &AwsLaunchTemplatePlacement{
					Tenancy: "dedicated",
					HostId:  "h-0123456789abcdef0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when host_id and host_resource_group_arn are both set", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.Placement = &AwsLaunchTemplatePlacement{
					Tenancy:              "host",
					HostId:               "h-0123456789abcdef0",
					HostResourceGroupArn: "arn:aws:resource-groups:us-west-2:123456789012:group/hosts",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an invalid nested_virtualization value", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.CpuOptions = &AwsLaunchTemplateCpuOptions{NestedVirtualization: "on"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("market and capacity reservation rules", func() {

			ginkgo.It("should return an error for an unknown market type", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.MarketType = "reserved"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when spot_options is combined with the capacity-block market", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.MarketType = "capacity-block"
				input.Spec.CapacityReservation = &AwsLaunchTemplateCapacityReservation{
					CapacityReservationId: "cr-0123456789abcdef0",
				}
				input.Spec.SpotOptions = &AwsLaunchTemplateSpotOptions{SpotInstanceType: "one-time"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for the capacity-block market without a reservation target", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.MarketType = "capacity-block"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when reservation id and resource group are both set", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.CapacityReservation = &AwsLaunchTemplateCapacityReservation{
					CapacityReservationId:               "cr-0123456789abcdef0",
					CapacityReservationResourceGroupArn: "arn:aws:resource-groups:us-west-2:123456789012:group/crs",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an unknown capacity reservation preference", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.CapacityReservation = &AwsLaunchTemplateCapacityReservation{Preference: "always"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("advanced networking rules", func() {

			ginkgo.It("should return an error for an invalid bandwidth weighting", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BandwidthWeighting = "ebs-2"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when SRD UDP is enabled without SRD itself", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.NetworkInterfaces = []*AwsLaunchTemplateNetworkInterface{
					{EnaSrd: &AwsLaunchTemplateEnaSrd{UdpEnabled: true}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an out-of-range UDP timeout", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.NetworkInterfaces = []*AwsLaunchTemplateNetworkInterface{
					{ConnectionTracking: &AwsLaunchTemplateConnectionTracking{UdpTimeoutSeconds: 120}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when a secondary interface sets both IP count and addresses", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.SecondaryInterfaces = []*AwsLaunchTemplateSecondaryInterface{
					{
						PrivateIpAddressCount: 2,
						PrivateIpAddresses:    []string{"10.0.1.10"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("volume initialization rules", func() {

			ginkgo.It("should return an error for an initialization rate without a snapshot", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName: "/dev/xvdb",
						Ebs:        &AwsLaunchTemplateEbs{VolumeInitializationRateMibps: 200},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error for an out-of-range initialization rate", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.BlockDeviceMappings = []*AwsLaunchTemplateBlockDeviceMapping{
					{
						DeviceName: "/dev/xvdb",
						Ebs: &AwsLaunchTemplateEbs{
							SnapshotId:                    "snap-0123456789abcdef0",
							VolumeInitializationRateMibps: 500,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("spot options rules", func() {

			ginkgo.It("should return an error for an invalid spot instance type", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.SpotOptions = &AwsLaunchTemplateSpotOptions{SpotInstanceType: "recurring"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when stop behavior is used with a one-time request", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.SpotOptions = &AwsLaunchTemplateSpotOptions{
					SpotInstanceType:             "one-time",
					InstanceInterruptionBehavior: "stop",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when valid_until is used without a persistent request", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.SpotOptions = &AwsLaunchTemplateSpotOptions{ValidUntil: "2027-01-01T00:00:00Z"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("private dns rules", func() {

			ginkgo.It("should return an error for an invalid hostname type", func() {
				input := minimalValidLaunchTemplate()
				input.Spec.PrivateDnsNameOptions = &AwsLaunchTemplatePrivateDnsNameOptions{HostnameType: "dns-name"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

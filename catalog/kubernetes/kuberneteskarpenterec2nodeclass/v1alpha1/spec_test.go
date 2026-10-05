package kuberneteskarpenterec2nodeclassv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestKubernetesKarpenterEc2NodeClass(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKarpenterEc2NodeClass Suite")
}

func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }
func int32Ptr(i int32) *int32    { return &i }
func int64Ptr(i int64) *int64    { return &i }

// aliasTerm returns the simplest AMI selection arm: a single alias term.
func aliasTerm(alias string) []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm {
	return []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{{Alias: stringPtr(alias)}}
}

// discoveryTags returns the conventional karpenter.sh/discovery tag selector.
func discoveryTags() map[string]string {
	return map[string]string{"karpenter.sh/discovery": "my-eks-cluster"}
}

var _ = ginkgo.Describe("KubernetesKarpenterEc2NodeClass Validation Tests", func() {
	var input *KubernetesKarpenterEc2NodeClass

	ginkgo.BeforeEach(func() {
		input = &KubernetesKarpenterEc2NodeClass{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKarpenterEc2NodeClass",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "default-al2023",
			},
			Spec: &KubernetesKarpenterEc2NodeClassSpec{
				AmiSelectorTerms: aliasTerm("al2023@v20240807"),
				SubnetSelectorTerms: []*KubernetesKarpenterEc2NodeClassSubnetSelectorTerm{
					{Tags: discoveryTags()},
				},
				SecurityGroupSelectorTerms: []*KubernetesKarpenterEc2NodeClassSecurityGroupSelectorTerm{
					{Tags: discoveryTags()},
				},
				Role: stringPtr("KarpenterNodeRole-my-eks-cluster"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal alias + role + discovery tags should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("instance_profile arm (without role) should be valid", func() {
			input.Spec.Role = nil
			input.Spec.InstanceProfile = stringPtr("my-preexisting-instance-profile")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("alias with matching ami_family should be valid", func() {
			input.Spec.AmiFamily = stringPtr("AL2023")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("alias with ami_family 'Custom' should be valid", func() {
			input.Spec.AmiFamily = stringPtr("Custom")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("windows alias with 'latest' should be valid", func() {
			input.Spec.AmiSelectorTerms = aliasTerm("windows2022@latest")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("id-based AMI term with ami_family should be valid", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Id: stringPtr("ami-0123456789abcdef0")},
			}
			input.Spec.AmiFamily = stringPtr("AL2023")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("name+owner AMI term with ami_family should be valid", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Name: stringPtr("my-golden-ami-*"), Owner: stringPtr("self")},
			}
			input.Spec.AmiFamily = stringPtr("Custom")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("ssm-parameter AMI term with ami_family should be valid", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{SsmParameter: stringPtr("/my/custom/ami/id")},
			}
			input.Spec.AmiFamily = stringPtr("AL2023")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("subnet and security-group selection by explicit id should be valid", func() {
			input.Spec.SubnetSelectorTerms = []*KubernetesKarpenterEc2NodeClassSubnetSelectorTerm{
				{Id: stringPtr("subnet-0a1b2c3d4e5f67890")},
			}
			input.Spec.SecurityGroupSelectorTerms = []*KubernetesKarpenterEc2NodeClassSecurityGroupSelectorTerm{
				{Id: stringPtr("sg-0a1b2c3d4e5f67890")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("security-group selection by name should be valid", func() {
			input.Spec.SecurityGroupSelectorTerms = []*KubernetesKarpenterEc2NodeClassSecurityGroupSelectorTerm{
				{Name: stringPtr("my-nodes-sg")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("EBS mapping with volume_size only should be valid", func() {
			input.Spec.BlockDeviceMappings = []*KubernetesKarpenterEc2NodeClassBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					Ebs:        &KubernetesKarpenterEc2NodeClassEbs{VolumeSize: stringPtr("100Gi")},
					RootVolume: boolPtr(true),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("EBS mapping with snapshot_id and volume_initialization_rate should be valid", func() {
			input.Spec.BlockDeviceMappings = []*KubernetesKarpenterEc2NodeClassBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					Ebs: &KubernetesKarpenterEc2NodeClassEbs{
						SnapshotId:               stringPtr("snap-0123456789abcdef0"),
						VolumeInitializationRate: int32Ptr(200),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("capacity-reservation term by id should be valid", func() {
			input.Spec.CapacityReservationSelectorTerms = []*KubernetesKarpenterEc2NodeClassCapacityReservationSelectorTerm{
				{Id: stringPtr("cr-0123456789abcdef0")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("capacity-reservation term by tags + owner_id + match criteria should be valid", func() {
			input.Spec.CapacityReservationSelectorTerms = []*KubernetesKarpenterEc2NodeClassCapacityReservationSelectorTerm{
				{
					Tags:                  map[string]string{"reservation-group": "batch"},
					OwnerId:               stringPtr("123456789012"),
					InstanceMatchCriteria: stringPtr("targeted"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("connection_tracking with one timeout should be valid", func() {
			input.Spec.ConnectionTracking = &KubernetesKarpenterEc2NodeClassConnectionTracking{
				TcpEstablishedTimeout: int32Ptr(86400),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("kubelet with paired soft eviction and ordered image GC should be valid", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				EvictionSoft:                map[string]string{"memory.available": "10%"},
				EvictionSoftGracePeriod:     map[string]string{"memory.available": "1m0s"},
				ImageGcHighThresholdPercent: int32Ptr(85),
				ImageGcLowThresholdPercent:  int32Ptr(80),
				KubeReserved:                map[string]string{"cpu": "200m", "memory": "512Mi"},
				SystemReserved:              map[string]string{"cpu": "100m"},
				MaxPods:                     int32Ptr(110),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full IMDS block with CRD-default values should be valid", func() {
			input.Spec.MetadataOptions = &KubernetesKarpenterEc2NodeClassMetadataOptions{
				HttpEndpoint:            stringPtr("enabled"),
				HttpProtocolIpv6:        stringPtr("disabled"),
				HttpPutResponseHopLimit: int64Ptr(1),
				HttpTokens:              stringPtr("required"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("network interfaces with a primary plus one efa-only per card should be valid", func() {
			input.Spec.NetworkInterfaces = []*KubernetesKarpenterEc2NodeClassNetworkInterface{
				{DeviceIndex: 0, InterfaceType: "interface", NetworkCardIndex: 0},
				{DeviceIndex: 1, InterfaceType: "efa-only", NetworkCardIndex: 0},
				{DeviceIndex: 0, InterfaceType: "efa-only", NetworkCardIndex: 1},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("placement group by name should be valid", func() {
			input.Spec.PlacementGroupSelector = &KubernetesKarpenterEc2NodeClassPlacementGroupSelector{
				Name: stringPtr("my-cluster-pg"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("placement group by id should be valid", func() {
			input.Spec.PlacementGroupSelector = &KubernetesKarpenterEc2NodeClassPlacementGroupSelector{
				Id: stringPtr("pg-0123456789abcdef0"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("unrestricted EC2 tags should be valid", func() {
			input.Spec.Tags = map[string]string{"team": "platform", "environment": "dev"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("instance_store_policy 'RAID0' should be valid", func() {
			input.Spec.InstanceStorePolicy = stringPtr("RAID0")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("role XOR instance_profile", func() {
		ginkgo.It("both role and instance_profile set should fail", func() {
			input.Spec.InstanceProfile = stringPtr("my-profile")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("neither role nor instance_profile set should fail", func() {
			input.Spec.Role = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("empty role string should fail (min_len)", func() {
			input.Spec.Role = stringPtr("")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("ami_family rules", func() {
		ginkgo.It("id-based AMI term without ami_family should fail (family not inferable)", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Id: stringPtr("ami-0123456789abcdef0")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("alias with a different family in ami_family should fail", func() {
			input.Spec.AmiFamily = stringPtr("Bottlerocket")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("ami_family outside the enum should fail", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Id: stringPtr("ami-0123456789abcdef0")},
			}
			input.Spec.AmiFamily = stringPtr("Ubuntu")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("AMI selector term rules", func() {
		ginkgo.It("empty ami_selector_terms should fail (min_items=1)", func() {
			input.Spec.AmiSelectorTerms = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("term with only owner should fail (owner is not a standalone selector)", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Owner: stringPtr("self")},
			}
			input.Spec.AmiFamily = stringPtr("AL2023")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("id combined with name should fail (id exclusive)", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Id: stringPtr("ami-0123456789abcdef0"), Name: stringPtr("my-ami")},
			}
			input.Spec.AmiFamily = stringPtr("AL2023")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("alias combined with tags should fail (alias exclusive)", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Alias: stringPtr("al2023@latest"), Tags: map[string]string{"team": "x"}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("alias term alongside a second term should fail (alias must be the only term)", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Alias: stringPtr("al2023@latest")},
				{Id: stringPtr("ami-0123456789abcdef0")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("alias with unknown family should fail the format rule", func() {
			input.Spec.AmiSelectorTerms = aliasTerm("ubuntu@latest")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("windows alias with a pinned version should fail (only 'latest')", func() {
			input.Spec.AmiSelectorTerms = aliasTerm("windows2022@v20240807")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("AMI id not matching the ami- pattern should fail", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Id: stringPtr("ami-ABCDEF")},
			}
			input.Spec.AmiFamily = stringPtr("AL2023")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("AMI selector tags with an empty value should fail", func() {
			input.Spec.AmiSelectorTerms = []*KubernetesKarpenterEc2NodeClassAmiSelectorTerm{
				{Tags: map[string]string{"team": ""}},
			}
			input.Spec.AmiFamily = stringPtr("AL2023")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("subnet selector term rules", func() {
		ginkgo.It("empty subnet_selector_terms should fail (min_items=1)", func() {
			input.Spec.SubnetSelectorTerms = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("term with neither id nor tags should fail", func() {
			input.Spec.SubnetSelectorTerms = []*KubernetesKarpenterEc2NodeClassSubnetSelectorTerm{{}}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("term with both id and tags should fail (id exclusive)", func() {
			input.Spec.SubnetSelectorTerms = []*KubernetesKarpenterEc2NodeClassSubnetSelectorTerm{
				{Id: stringPtr("subnet-0a1b2c3d4e5f67890"), Tags: discoveryTags()},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("subnet id not matching the subnet- pattern should fail", func() {
			input.Spec.SubnetSelectorTerms = []*KubernetesKarpenterEc2NodeClassSubnetSelectorTerm{
				{Id: stringPtr("vpc-0a1b2c3d")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("subnet selector tags with an empty key should fail", func() {
			input.Spec.SubnetSelectorTerms = []*KubernetesKarpenterEc2NodeClassSubnetSelectorTerm{
				{Tags: map[string]string{"": "x"}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("security-group selector term rules", func() {
		ginkgo.It("empty security_group_selector_terms should fail (min_items=1)", func() {
			input.Spec.SecurityGroupSelectorTerms = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("term with no selector at all should fail", func() {
			input.Spec.SecurityGroupSelectorTerms = []*KubernetesKarpenterEc2NodeClassSecurityGroupSelectorTerm{{}}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("term with both id and tags should fail (id exclusive)", func() {
			input.Spec.SecurityGroupSelectorTerms = []*KubernetesKarpenterEc2NodeClassSecurityGroupSelectorTerm{
				{Id: stringPtr("sg-0a1b2c3d4e5f67890"), Tags: discoveryTags()},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("term with both name and id should fail (name exclusive)", func() {
			input.Spec.SecurityGroupSelectorTerms = []*KubernetesKarpenterEc2NodeClassSecurityGroupSelectorTerm{
				{Name: stringPtr("my-sg"), Id: stringPtr("sg-0a1b2c3d4e5f67890")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("security-group id not matching the sg- pattern should fail", func() {
			input.Spec.SecurityGroupSelectorTerms = []*KubernetesKarpenterEc2NodeClassSecurityGroupSelectorTerm{
				{Id: stringPtr("sg-UPPER")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("block device mapping / EBS rules", func() {
		ginkgo.It("EBS with neither snapshot_id nor volume_size should fail", func() {
			input.Spec.BlockDeviceMappings = []*KubernetesKarpenterEc2NodeClassBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					Ebs:        &KubernetesKarpenterEc2NodeClassEbs{Encrypted: boolPtr(true)},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("volume_initialization_rate without snapshot_id should fail", func() {
			input.Spec.BlockDeviceMappings = []*KubernetesKarpenterEc2NodeClassBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					Ebs: &KubernetesKarpenterEc2NodeClassEbs{
						VolumeSize:               stringPtr("100Gi"),
						VolumeInitializationRate: int32Ptr(200),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("volume_initialization_rate outside 100-300 should fail", func() {
			input.Spec.BlockDeviceMappings = []*KubernetesKarpenterEc2NodeClassBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					Ebs: &KubernetesKarpenterEc2NodeClassEbs{
						SnapshotId:               stringPtr("snap-0123456789abcdef0"),
						VolumeInitializationRate: int32Ptr(400),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("throughput outside 125-1000 should fail", func() {
			input.Spec.BlockDeviceMappings = []*KubernetesKarpenterEc2NodeClassBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					Ebs: &KubernetesKarpenterEc2NodeClassEbs{
						VolumeSize: stringPtr("100Gi"),
						Throughput: int64Ptr(2000),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("volume_size without a Gi/G/Ti/T unit should fail the quantity pattern", func() {
			input.Spec.BlockDeviceMappings = []*KubernetesKarpenterEc2NodeClassBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					Ebs:        &KubernetesKarpenterEc2NodeClassEbs{VolumeSize: stringPtr("100")},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("volume_type outside the enum should fail", func() {
			input.Spec.BlockDeviceMappings = []*KubernetesKarpenterEc2NodeClassBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					Ebs: &KubernetesKarpenterEc2NodeClassEbs{
						VolumeSize: stringPtr("100Gi"),
						VolumeType: stringPtr("gp4"),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("two mappings both marked root_volume should fail", func() {
			input.Spec.BlockDeviceMappings = []*KubernetesKarpenterEc2NodeClassBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					Ebs:        &KubernetesKarpenterEc2NodeClassEbs{VolumeSize: stringPtr("100Gi")},
					RootVolume: boolPtr(true),
				},
				{
					DeviceName: "/dev/xvdb",
					Ebs:        &KubernetesKarpenterEc2NodeClassEbs{VolumeSize: stringPtr("200Gi")},
					RootVolume: boolPtr(true),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("capacity-reservation selector term rules", func() {
		ginkgo.It("term with only owner_id should fail (needs id, tags or match criteria)", func() {
			input.Spec.CapacityReservationSelectorTerms = []*KubernetesKarpenterEc2NodeClassCapacityReservationSelectorTerm{
				{OwnerId: stringPtr("123456789012")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("term with id and tags should fail (id exclusive)", func() {
			input.Spec.CapacityReservationSelectorTerms = []*KubernetesKarpenterEc2NodeClassCapacityReservationSelectorTerm{
				{Id: stringPtr("cr-0123456789abcdef0"), Tags: map[string]string{"x": "y"}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("instance_match_criteria outside open/targeted should fail", func() {
			input.Spec.CapacityReservationSelectorTerms = []*KubernetesKarpenterEc2NodeClassCapacityReservationSelectorTerm{
				{InstanceMatchCriteria: stringPtr("any")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("owner_id that is not 12 digits should fail", func() {
			input.Spec.CapacityReservationSelectorTerms = []*KubernetesKarpenterEc2NodeClassCapacityReservationSelectorTerm{
				{Tags: map[string]string{"x": "y"}, OwnerId: stringPtr("12345")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("connection tracking rules", func() {
		ginkgo.It("empty connection_tracking block should fail (at least one timeout)", func() {
			input.Spec.ConnectionTracking = &KubernetesKarpenterEc2NodeClassConnectionTracking{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("tcp_established_timeout below 60 should fail", func() {
			input.Spec.ConnectionTracking = &KubernetesKarpenterEc2NodeClassConnectionTracking{
				TcpEstablishedTimeout: int32Ptr(59),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("udp_stream_timeout above 180 should fail", func() {
			input.Spec.ConnectionTracking = &KubernetesKarpenterEc2NodeClassConnectionTracking{
				UdpStreamTimeout: int32Ptr(200),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("udp_timeout below 30 should fail", func() {
			input.Spec.ConnectionTracking = &KubernetesKarpenterEc2NodeClassConnectionTracking{
				UdpTimeout: int32Ptr(20),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("kubelet rules", func() {
		ginkgo.It("eviction_hard with an unknown signal should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				EvictionHard: map[string]string{"disk.available": "10%"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("eviction_soft with an unknown signal should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				EvictionSoft:            map[string]string{"disk.available": "10%"},
				EvictionSoftGracePeriod: map[string]string{"disk.available": "1m0s"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("image GC high threshold not greater than low should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				ImageGcHighThresholdPercent: int32Ptr(50),
				ImageGcLowThresholdPercent:  int32Ptr(50),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("image GC threshold above 100 should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				ImageGcHighThresholdPercent: int32Ptr(120),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("eviction_soft signal without a matching grace period should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				EvictionSoft: map[string]string{"memory.available": "10%"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("grace period without a matching eviction_soft signal should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				EvictionSoftGracePeriod: map[string]string{"memory.available": "1m0s"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("kube_reserved with an unknown resource key should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				KubeReserved: map[string]string{"gpu": "1"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("kube_reserved with a negative quantity should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				KubeReserved: map[string]string{"cpu": "-100m"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("system_reserved with an unknown resource key should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				SystemReserved: map[string]string{"network": "1"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("negative max_pods should fail", func() {
			input.Spec.Kubelet = &KubernetesKarpenterEc2NodeClassKubelet{
				MaxPods: int32Ptr(-1),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("IMDS metadata options rules", func() {
		ginkgo.It("http_endpoint outside enabled/disabled should fail", func() {
			input.Spec.MetadataOptions = &KubernetesKarpenterEc2NodeClassMetadataOptions{
				HttpEndpoint: stringPtr("on"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("http_protocol_ipv6 outside enabled/disabled should fail", func() {
			input.Spec.MetadataOptions = &KubernetesKarpenterEc2NodeClassMetadataOptions{
				HttpProtocolIpv6: stringPtr("dual-stack"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("http_tokens outside required/optional should fail", func() {
			input.Spec.MetadataOptions = &KubernetesKarpenterEc2NodeClassMetadataOptions{
				HttpTokens: stringPtr("v2"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("hop limit outside 1-64 should fail", func() {
			input.Spec.MetadataOptions = &KubernetesKarpenterEc2NodeClassMetadataOptions{
				HttpPutResponseHopLimit: int64Ptr(65),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("network interface rules", func() {
		ginkgo.It("interfaces without a primary at device 0 / card 0 should fail", func() {
			input.Spec.NetworkInterfaces = []*KubernetesKarpenterEc2NodeClassNetworkInterface{
				{DeviceIndex: 1, InterfaceType: "interface", NetworkCardIndex: 0},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate (card, device) pairs should fail", func() {
			input.Spec.NetworkInterfaces = []*KubernetesKarpenterEc2NodeClassNetworkInterface{
				{DeviceIndex: 0, InterfaceType: "interface", NetworkCardIndex: 0},
				{DeviceIndex: 0, InterfaceType: "efa-only", NetworkCardIndex: 0},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("two efa-only interfaces on the same card should fail", func() {
			input.Spec.NetworkInterfaces = []*KubernetesKarpenterEc2NodeClassNetworkInterface{
				{DeviceIndex: 0, InterfaceType: "interface", NetworkCardIndex: 0},
				{DeviceIndex: 1, InterfaceType: "efa-only", NetworkCardIndex: 0},
				{DeviceIndex: 2, InterfaceType: "efa-only", NetworkCardIndex: 0},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("interface_type outside interface/efa-only should fail", func() {
			input.Spec.NetworkInterfaces = []*KubernetesKarpenterEc2NodeClassNetworkInterface{
				{DeviceIndex: 0, InterfaceType: "efa", NetworkCardIndex: 0},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("missing interface_type should fail (required)", func() {
			input.Spec.NetworkInterfaces = []*KubernetesKarpenterEc2NodeClassNetworkInterface{
				{DeviceIndex: 0, NetworkCardIndex: 0},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("placement group rules", func() {
		ginkgo.It("both name and id set should fail (exactly one)", func() {
			input.Spec.PlacementGroupSelector = &KubernetesKarpenterEc2NodeClassPlacementGroupSelector{
				Name: stringPtr("my-pg"),
				Id:   stringPtr("pg-0123456789abcdef0"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("neither name nor id set should fail (exactly one)", func() {
			input.Spec.PlacementGroupSelector = &KubernetesKarpenterEc2NodeClassPlacementGroupSelector{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("placement-group id not matching the pg- pattern should fail", func() {
			input.Spec.PlacementGroupSelector = &KubernetesKarpenterEc2NodeClassPlacementGroupSelector{
				Id: stringPtr("placement-123"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("restricted EC2 tag keys", func() {
		ginkgo.It("kubernetes.io/cluster/* key should fail", func() {
			input.Spec.Tags = map[string]string{"kubernetes.io/cluster/my-cluster": "owned"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("karpenter.sh/nodepool key should fail", func() {
			input.Spec.Tags = map[string]string{"karpenter.sh/nodepool": "default"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("karpenter.sh/nodeclaim key should fail", func() {
			input.Spec.Tags = map[string]string{"karpenter.sh/nodeclaim": "x"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("karpenter.k8s.aws/ec2nodeclass key should fail", func() {
			input.Spec.Tags = map[string]string{"karpenter.k8s.aws/ec2nodeclass": "x"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("eks:eks-cluster-name key should fail", func() {
			input.Spec.Tags = map[string]string{"eks:eks-cluster-name": "my-cluster"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("misc scalar rules", func() {
		ginkgo.It("instance_store_policy other than RAID0 should fail", func() {
			input.Spec.InstanceStorePolicy = stringPtr("RAID1")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("negative ip_prefix_count should fail", func() {
			input.Spec.IpPrefixCount = int32Ptr(-1)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})

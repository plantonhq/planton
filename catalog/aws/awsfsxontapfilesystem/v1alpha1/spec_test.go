package awsfsxontapfilesystemv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	shared "github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsFsxOntapFileSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsFsxOntapFileSystemSpec Validation Suite")
}

func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

var _ = ginkgo.Describe("AwsFsxOntapFileSystemSpec validations", func() {
	var spec *AwsFsxOntapFileSystemSpec

	// The implicit deployment type is SINGLE_AZ_2 with one HA pair, whose
	// per-HA-pair throughput tiers are 384/768/1536/3072/6144.
	ginkgo.BeforeEach(func() {
		spec = &AwsFsxOntapFileSystemSpec{
			Region:                      "us-west-2",
			SubnetIds:                   []*foreignkeyv1.StringValueOrRef{strRef("subnet-abc123")},
			StorageCapacityGib:          1024,
			ThroughputCapacityPerHaPair: int32Ptr(384),
		}
	})

	multiAzBase := func(deploymentType string) {
		spec.DeploymentType = stringPtr(deploymentType)
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.PreferredSubnetId = strRef("subnet-abc123")
	}

	// -------------------------------------------------------------------------
	// Happy path — valid configurations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_1 with the whole-file-system throughput arm", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_1")
		spec.ThroughputCapacityPerHaPair = nil
		spec.ThroughputCapacity = int32Ptr(128)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_1 with a first-generation per-HA-pair value", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_1")
		spec.ThroughputCapacityPerHaPair = int32Ptr(128)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_2 explicitly", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_2")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MULTI_AZ_1 with two subnets, a preferred subnet, and a gen1 tier", func() {
		multiAzBase("MULTI_AZ_1")
		spec.ThroughputCapacityPerHaPair = int32Ptr(512)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MULTI_AZ_2 with two subnets, a preferred subnet, and a gen2 tier", func() {
		multiAzBase("MULTI_AZ_2")
		spec.ThroughputCapacityPerHaPair = int32Ptr(768)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multi-AZ with route_table_ids and endpoint_ip_address_range", func() {
		multiAzBase("MULTI_AZ_1")
		spec.ThroughputCapacityPerHaPair = int32Ptr(512)
		spec.RouteTableIds = []*foreignkeyv1.StringValueOrRef{strRef("rtb-abc123")}
		spec.EndpointIpAddressRange = "198.19.255.0/24"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_2 scale-out with multiple HA pairs", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_2")
		spec.HaPairs = int32Ptr(6)
		spec.StorageCapacityGib = 12288
		spec.ThroughputCapacityPerHaPair = int32Ptr(1536)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts the maximum scale-out shape (12 HA pairs, 1 PiB)", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_2")
		spec.HaPairs = int32Ptr(12)
		spec.StorageCapacityGib = 1048576
		spec.ThroughputCapacityPerHaPair = int32Ptr(6144)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts fsx_admin_password within valid range", func() {
		spec.FsxAdminPassword = "MyP@ss12"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts fsx_admin_password at 50 characters", func() {
		spec.FsxAdminPassword = "Abcdefgh12345678901234567890123456789012345678901!"
		gomega.Expect(len(spec.FsxAdminPassword)).To(gomega.Equal(50))
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full production SINGLE_AZ_2 configuration", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_2")
		spec.StorageCapacityGib = 4096
		spec.StorageType = stringPtr("SSD")
		spec.ThroughputCapacityPerHaPair = int32Ptr(768)
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{strRef("sg-abc123")}
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
		spec.FsxAdminPassword = "OntapAdmin2024!"
		spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
			Mode: stringPtr("USER_PROVISIONED"),
			Iops: 100000,
		}
		spec.AutomaticBackupRetentionDays = int32Ptr(7)
		spec.DailyAutomaticBackupStartTime = "05:00"
		spec.WeeklyMaintenanceStartTime = "7:02:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full production MULTI_AZ_2 configuration", func() {
		multiAzBase("MULTI_AZ_2")
		spec.StorageCapacityGib = 8192
		spec.StorageType = stringPtr("SSD")
		spec.ThroughputCapacityPerHaPair = int32Ptr(1536)
		spec.EndpointIpAddressRange = "198.19.255.0/24"
		spec.RouteTableIds = []*foreignkeyv1.StringValueOrRef{
			strRef("rtb-abc123"),
			strRef("rtb-def456"),
		}
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{strRef("sg-abc123")}
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
		spec.FsxAdminPassword = "OntapAdmin2024!"
		spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
			Mode: stringPtr("AUTOMATIC"),
		}
		spec.AutomaticBackupRetentionDays = int32Ptr(14)
		spec.DailyAutomaticBackupStartTime = "03:00"
		spec.WeeklyMaintenanceStartTime = "1:05:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts every gen2 single-HA-pair throughput tier", func() {
		for _, tp := range []int32{384, 768, 1536, 3072, 6144} {
			spec.ThroughputCapacityPerHaPair = int32Ptr(tp)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil(), "throughput %d should be valid", tp)
		}
	})

	ginkgo.It("accepts every gen1 per-HA-pair throughput tier on SINGLE_AZ_1", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_1")
		for _, tp := range []int32{128, 256, 512, 1024, 2048, 4096} {
			spec.ThroughputCapacityPerHaPair = int32Ptr(tp)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil(), "throughput %d should be valid", tp)
		}
	})

	ginkgo.It("accepts every whole-file-system throughput tier", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_1")
		spec.ThroughputCapacityPerHaPair = nil
		for _, tp := range []int32{128, 256, 512, 1024, 2048, 4096} {
			spec.ThroughputCapacity = int32Ptr(tp)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil(), "throughput %d should be valid", tp)
		}
	})

	// -------------------------------------------------------------------------
	// Disk IOPS configuration happy paths
	// -------------------------------------------------------------------------

	ginkgo.It("accepts AUTOMATIC IOPS mode", func() {
		spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
			Mode: stringPtr("AUTOMATIC"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts USER_PROVISIONED IOPS mode with iops at the ceiling", func() {
		spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
			Mode: stringPtr("USER_PROVISIONED"),
			Iops: 2400000,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level validation failures
	// -------------------------------------------------------------------------

	ginkgo.Context("field-level validations", func() {
		ginkgo.It("fails when subnet_ids is empty", func() {
			spec.SubnetIds = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when subnet_ids has more than two entries", func() {
			spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-a"), strRef("subnet-b"), strRef("subnet-c"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when storage_capacity_gib is below minimum (1024)", func() {
			spec.StorageCapacityGib = 512
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when storage_capacity_gib is zero", func() {
			spec.StorageCapacityGib = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when ha_pairs is zero", func() {
			spec.HaPairs = int32Ptr(0)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when ha_pairs exceeds 12", func() {
			spec.HaPairs = int32Ptr(13)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when automatic_backup_retention_days is negative", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(-1)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when automatic_backup_retention_days exceeds 90", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(91)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when disk iops exceeds 2,400,000", func() {
			spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
				Mode: stringPtr("USER_PROVISIONED"),
				Iops: 2400001,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: deployment_type_valid
	// -------------------------------------------------------------------------

	ginkgo.Context("deployment_type_valid", func() {
		ginkgo.It("fails when deployment_type is invalid", func() {
			spec.DeploymentType = stringPtr("INVALID_TYPE")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when deployment_type is lowercase", func() {
			spec.DeploymentType = stringPtr("single_az_2")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails for a Lustre deployment type", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: storage_type_valid — ONTAP is SSD-only
	// -------------------------------------------------------------------------

	ginkgo.Context("storage_type_valid", func() {
		ginkgo.It("accepts explicit SSD", func() {
			spec.StorageType = stringPtr("SSD")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails for HDD (a Windows/Lustre storage class)", func() {
			spec.StorageType = stringPtr("HDD")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails for INTELLIGENT_TIERING (an OpenZFS/Lustre storage class)", func() {
			spec.StorageType = stringPtr("INTELLIGENT_TIERING")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when storage_type is lowercase", func() {
			spec.StorageType = stringPtr("ssd")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: throughput_exactly_one_arm
	// -------------------------------------------------------------------------

	ginkgo.Context("throughput_exactly_one_arm", func() {
		ginkgo.It("fails when both throughput arms are set", func() {
			spec.ThroughputCapacity = int32Ptr(1024)
			spec.ThroughputCapacityPerHaPair = int32Ptr(384)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when neither throughput arm is set", func() {
			spec.ThroughputCapacity = nil
			spec.ThroughputCapacityPerHaPair = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: throughput value sets
	// -------------------------------------------------------------------------

	ginkgo.Context("throughput value sets", func() {
		ginkgo.It("fails when throughput_capacity uses a per-HA-pair-only tier", func() {
			spec.ThroughputCapacityPerHaPair = nil
			spec.ThroughputCapacity = int32Ptr(384)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when a gen2 single-pair file system uses a gen1 tier (128)", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.ThroughputCapacityPerHaPair = int32Ptr(128)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when the implicit SINGLE_AZ_2 default uses a gen1 tier", func() {
			spec.ThroughputCapacityPerHaPair = int32Ptr(1024)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when SINGLE_AZ_1 uses a gen2 tier (384)", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_1")
			spec.ThroughputCapacityPerHaPair = int32Ptr(384)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when scale-out (ha_pairs > 1) uses a single-pair tier (384)", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.HaPairs = int32Ptr(2)
			spec.StorageCapacityGib = 4096
			spec.ThroughputCapacityPerHaPair = int32Ptr(384)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when throughput is not any valid tier", func() {
			spec.ThroughputCapacityPerHaPair = int32Ptr(100)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: ha_pairs_scale_out_single_az_2_only
	// -------------------------------------------------------------------------

	ginkgo.Context("ha_pairs_scale_out_single_az_2_only", func() {
		ginkgo.It("fails when ha_pairs > 1 on SINGLE_AZ_1", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_1")
			spec.HaPairs = int32Ptr(2)
			spec.StorageCapacityGib = 4096
			spec.ThroughputCapacityPerHaPair = int32Ptr(1024)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when ha_pairs > 1 on MULTI_AZ_1", func() {
			multiAzBase("MULTI_AZ_1")
			spec.HaPairs = int32Ptr(2)
			spec.StorageCapacityGib = 4096
			spec.ThroughputCapacityPerHaPair = int32Ptr(1024)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when ha_pairs > 1 on MULTI_AZ_2", func() {
			multiAzBase("MULTI_AZ_2")
			spec.HaPairs = int32Ptr(3)
			spec.StorageCapacityGib = 4096
			spec.ThroughputCapacityPerHaPair = int32Ptr(1536)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts ha_pairs = 1 on MULTI_AZ_1", func() {
			multiAzBase("MULTI_AZ_1")
			spec.HaPairs = int32Ptr(1)
			spec.ThroughputCapacityPerHaPair = int32Ptr(512)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: storage capacity contracts
	// -------------------------------------------------------------------------

	ginkgo.Context("storage capacity contracts", func() {
		ginkgo.It("fails when capacity is below 1024 GiB per HA pair", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.HaPairs = int32Ptr(4)
			spec.StorageCapacityGib = 2048 // needs >= 4096
			spec.ThroughputCapacityPerHaPair = int32Ptr(1536)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when capacity exceeds 512 TiB with a single HA pair", func() {
			spec.StorageCapacityGib = 600000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when a gen1 file system exceeds 192 TiB", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_1")
			spec.ThroughputCapacityPerHaPair = int32Ptr(1024)
			spec.StorageCapacityGib = 200000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts 192 TiB exactly on gen1", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_1")
			spec.ThroughputCapacityPerHaPair = int32Ptr(1024)
			spec.StorageCapacityGib = 196608
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: subnet-count contracts
	// -------------------------------------------------------------------------

	ginkgo.Context("subnet-count contracts", func() {
		ginkgo.It("fails when a single-AZ file system has two subnets", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-abc123"),
				strRef("subnet-def456"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when a multi-AZ file system has one subnet", func() {
			spec.DeploymentType = stringPtr("MULTI_AZ_2")
			spec.PreferredSubnetId = strRef("subnet-abc123")
			spec.ThroughputCapacityPerHaPair = int32Ptr(768)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: preferred_subnet_multi_az_contract
	// -------------------------------------------------------------------------

	ginkgo.Context("preferred_subnet_multi_az_contract", func() {
		ginkgo.It("fails when preferred_subnet_id is set on SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.PreferredSubnetId = strRef("subnet-abc123")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when preferred_subnet_id is set on the implicit default", func() {
			spec.PreferredSubnetId = strRef("subnet-abc123")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when a multi-AZ file system omits preferred_subnet_id", func() {
			spec.DeploymentType = stringPtr("MULTI_AZ_1")
			spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-abc123"),
				strRef("subnet-def456"),
			}
			spec.ThroughputCapacityPerHaPair = int32Ptr(512)
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts nil preferred_subnet_id on SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.PreferredSubnetId = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: endpoint_ip_range contracts
	// -------------------------------------------------------------------------

	ginkgo.Context("endpoint_ip_range contracts", func() {
		ginkgo.It("fails when endpoint_ip_address_range is set on SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.EndpointIpAddressRange = "198.19.255.0/24"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when endpoint_ip_address_range is not CIDR-shaped", func() {
			multiAzBase("MULTI_AZ_1")
			spec.ThroughputCapacityPerHaPair = int32Ptr(512)
			spec.EndpointIpAddressRange = "198.19.255.0"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts empty endpoint_ip_address_range on SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.EndpointIpAddressRange = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: route_tables_require_multi_az
	// -------------------------------------------------------------------------

	ginkgo.Context("route_tables_require_multi_az", func() {
		ginkgo.It("fails when route_table_ids is set on SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.RouteTableIds = []*foreignkeyv1.StringValueOrRef{strRef("rtb-abc123")}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when route_table_ids is set on SINGLE_AZ_1", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_1")
			spec.ThroughputCapacityPerHaPair = int32Ptr(128)
			spec.RouteTableIds = []*foreignkeyv1.StringValueOrRef{strRef("rtb-abc123")}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts empty route_table_ids on SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.RouteTableIds = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: admin_password_length
	// -------------------------------------------------------------------------

	ginkgo.Context("admin_password_length", func() {
		ginkgo.It("fails when fsx_admin_password is too short (7 chars)", func() {
			spec.FsxAdminPassword = "Short1!"
			gomega.Expect(len(spec.FsxAdminPassword)).To(gomega.Equal(7))
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when fsx_admin_password is too long (51 chars)", func() {
			spec.FsxAdminPassword = "Abcdefgh123456789012345678901234567890123456789012!"
			gomega.Expect(len(spec.FsxAdminPassword)).To(gomega.Equal(51))
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts empty fsx_admin_password (optional)", func() {
			spec.FsxAdminPassword = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts fsx_admin_password at exactly 8 characters", func() {
			spec.FsxAdminPassword = "Exact8!!"
			gomega.Expect(len(spec.FsxAdminPassword)).To(gomega.Equal(8))
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: backup_time_requires_retention + time formats
	// -------------------------------------------------------------------------

	ginkgo.Context("backup and maintenance windows", func() {
		ginkgo.It("fails when backup time is set but retention is 0", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(0)
			spec.DailyAutomaticBackupStartTime = "05:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when backup time is set without explicit retention", func() {
			spec.DailyAutomaticBackupStartTime = "05:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when backup time is not HH:MM", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(7)
			spec.DailyAutomaticBackupStartTime = "5:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when backup hour is out of range", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(7)
			spec.DailyAutomaticBackupStartTime = "25:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when maintenance day is out of range", func() {
			spec.WeeklyMaintenanceStartTime = "8:02:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts backup time with retention > 0", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(7)
			spec.DailyAutomaticBackupStartTime = "05:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: iops mode contracts (nested message)
	// -------------------------------------------------------------------------

	ginkgo.Context("disk IOPS contracts", func() {
		ginkgo.It("fails when IOPS mode is invalid", func() {
			spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
				Mode: stringPtr("MANUAL"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when iops is set with AUTOMATIC mode", func() {
			spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
				Mode: stringPtr("AUTOMATIC"),
				Iops: 50000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when iops is set without explicit mode", func() {
			spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
				Iops: 50000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts zero iops with AUTOMATIC mode", func() {
			spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
				Mode: stringPtr("AUTOMATIC"),
				Iops: 0,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// API envelope validations
	// -------------------------------------------------------------------------

	ginkgo.Context("API envelope", func() {
		ginkgo.It("validates a complete API resource", func() {
			resource := &AwsFsxOntapFileSystem{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "AwsFsxOntapFileSystem",
				Metadata: &shared.CatalogObjectMetadata{
					Name: "my-ontap-fs",
					Id:   "awsfxo-test-123",
					Org:  "test-org",
					Env:  "dev",
				},
				Spec: spec,
			}
			err := protovalidate.Validate(resource)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails with wrong api_version", func() {
			resource := &AwsFsxOntapFileSystem{
				ApiVersion: "wrong/v1",
				Kind:       "AwsFsxOntapFileSystem",
				Metadata: &shared.CatalogObjectMetadata{
					Name: "my-ontap-fs",
					Id:   "awsfxo-test-123",
					Org:  "test-org",
					Env:  "dev",
				},
				Spec: spec,
			}
			err := protovalidate.Validate(resource)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails with wrong kind", func() {
			resource := &AwsFsxOntapFileSystem{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "WrongKind",
				Metadata: &shared.CatalogObjectMetadata{
					Name: "my-ontap-fs",
					Id:   "awsfxo-test-123",
					Org:  "test-org",
					Env:  "dev",
				},
				Spec: spec,
			}
			err := protovalidate.Validate(resource)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})

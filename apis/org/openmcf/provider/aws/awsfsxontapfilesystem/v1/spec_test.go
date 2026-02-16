package awsfsxontapfilesystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	shared "github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
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

	ginkgo.BeforeEach(func() {
		spec = &AwsFsxOntapFileSystemSpec{
			SubnetIds:                    []*foreignkeyv1.StringValueOrRef{strRef("subnet-abc123")},
			StorageCapacityGib:           1024,
			ThroughputCapacityPerHaPair:  128,
		}
	})

	// -------------------------------------------------------------------------
	// Happy path — valid configurations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_1 deployment type", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_1")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_2 deployment type", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_2")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MULTI_AZ_1 deployment type with two subnets", func() {
		spec.DeploymentType = stringPtr("MULTI_AZ_1")
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.PreferredSubnetId = strRef("subnet-abc123")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MULTI_AZ_2 deployment type with two subnets", func() {
		spec.DeploymentType = stringPtr("MULTI_AZ_2")
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.PreferredSubnetId = strRef("subnet-abc123")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multi-AZ with route_table_ids and endpoint_ip_address_range", func() {
		spec.DeploymentType = stringPtr("MULTI_AZ_1")
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.PreferredSubnetId = strRef("subnet-abc123")
		spec.RouteTableIds = []*foreignkeyv1.StringValueOrRef{strRef("rtb-abc123")}
		spec.EndpointIpAddressRange = "198.19.255.0/24"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_2 with multiple HA pairs", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_2")
		spec.HaPairs = int32Ptr(6)
		spec.ThroughputCapacityPerHaPair = 512
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_1 with multiple HA pairs", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_1")
		spec.HaPairs = int32Ptr(12)
		spec.ThroughputCapacityPerHaPair = 256
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts HDD storage type", func() {
		spec.StorageType = stringPtr("HDD")
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
		spec.StorageCapacityGib = 2048
		spec.StorageType = stringPtr("SSD")
		spec.ThroughputCapacityPerHaPair = 512
		spec.HaPairs = int32Ptr(2)
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{strRef("sg-abc123")}
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
		spec.FsxAdminPassword = "OntapAdmin2024!"
		spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
			Mode: stringPtr("USER_PROVISIONED"),
			Iops: 100000,
		}
		spec.AutomaticBackupRetentionDays = int32Ptr(7)
		spec.DailyAutomaticBackupStartTime = "05:00"
		spec.CopyTagsToBackups = true
		spec.WeeklyMaintenanceStartTime = "7:02:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full production MULTI_AZ_2 configuration", func() {
		spec.DeploymentType = stringPtr("MULTI_AZ_2")
		spec.StorageCapacityGib = 4096
		spec.StorageType = stringPtr("SSD")
		spec.ThroughputCapacityPerHaPair = 1024
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.PreferredSubnetId = strRef("subnet-abc123")
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
		spec.CopyTagsToBackups = true
		spec.SkipFinalBackup = (*bool)(nil)
		spec.WeeklyMaintenanceStartTime = "1:05:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts all valid throughput_capacity_per_ha_pair values", func() {
		for _, tp := range []int32{128, 256, 384, 512, 768, 1024, 1536, 2048, 3072, 4096, 6144} {
			spec.ThroughputCapacityPerHaPair = tp
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

	ginkgo.It("accepts USER_PROVISIONED IOPS mode with iops", func() {
		spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
			Mode: stringPtr("USER_PROVISIONED"),
			Iops: 80000,
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

		ginkgo.It("fails when throughput_capacity_per_ha_pair is zero", func() {
			spec.ThroughputCapacityPerHaPair = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when throughput_capacity_per_ha_pair is negative", func() {
			spec.ThroughputCapacityPerHaPair = -1
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

		ginkgo.It("fails for OpenZFS deployment type MULTI_AZ_1 without multi-AZ fields", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: storage_type_valid
	// -------------------------------------------------------------------------

	ginkgo.Context("storage_type_valid", func() {
		ginkgo.It("fails when storage_type is invalid", func() {
			spec.StorageType = stringPtr("NVME")
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
	// CEL: throughput_valid
	// -------------------------------------------------------------------------

	ginkgo.Context("throughput_valid", func() {
		ginkgo.It("fails when throughput is not a valid tier", func() {
			spec.ThroughputCapacityPerHaPair = 100
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when throughput is 160 (valid for OpenZFS, not ONTAP)", func() {
			spec.ThroughputCapacityPerHaPair = 160
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when throughput is 64 (valid for Lustre, not ONTAP)", func() {
			spec.ThroughputCapacityPerHaPair = 64
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: ha_pairs_single_az_only
	// -------------------------------------------------------------------------

	ginkgo.Context("ha_pairs_single_az_only", func() {
		ginkgo.It("fails when ha_pairs > 1 on MULTI_AZ_1", func() {
			spec.DeploymentType = stringPtr("MULTI_AZ_1")
			spec.HaPairs = int32Ptr(2)
			spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-abc123"),
				strRef("subnet-def456"),
			}
			spec.PreferredSubnetId = strRef("subnet-abc123")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when ha_pairs > 1 on MULTI_AZ_2", func() {
			spec.DeploymentType = stringPtr("MULTI_AZ_2")
			spec.HaPairs = int32Ptr(3)
			spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-abc123"),
				strRef("subnet-def456"),
			}
			spec.PreferredSubnetId = strRef("subnet-abc123")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts ha_pairs = 1 on MULTI_AZ_1", func() {
			spec.DeploymentType = stringPtr("MULTI_AZ_1")
			spec.HaPairs = int32Ptr(1)
			spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-abc123"),
				strRef("subnet-def456"),
			}
			spec.PreferredSubnetId = strRef("subnet-abc123")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: preferred_subnet_requires_multi_az
	// -------------------------------------------------------------------------

	ginkgo.Context("preferred_subnet_requires_multi_az", func() {
		ginkgo.It("fails when preferred_subnet_id is set on SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.PreferredSubnetId = strRef("subnet-abc123")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when preferred_subnet_id is set on SINGLE_AZ_1", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_1")
			spec.PreferredSubnetId = strRef("subnet-abc123")
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
	// CEL: endpoint_ip_range_requires_multi_az
	// -------------------------------------------------------------------------

	ginkgo.Context("endpoint_ip_range_requires_multi_az", func() {
		ginkgo.It("fails when endpoint_ip_address_range is set on SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.EndpointIpAddressRange = "198.19.255.0/24"
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
	// CEL: backup_time_requires_retention
	// -------------------------------------------------------------------------

	ginkgo.Context("backup_time_requires_retention", func() {
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

		ginkgo.It("accepts backup time with retention > 0", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(7)
			spec.DailyAutomaticBackupStartTime = "05:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts empty backup time with any retention", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(7)
			spec.DailyAutomaticBackupStartTime = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: iops_mode_valid (nested message)
	// -------------------------------------------------------------------------

	ginkgo.Context("iops_mode_valid", func() {
		ginkgo.It("fails when IOPS mode is invalid", func() {
			spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
				Mode: stringPtr("MANUAL"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when IOPS mode is lowercase", func() {
			spec.DiskIopsConfiguration = &AwsFsxOntapFileSystemDiskIopsConfiguration{
				Mode: stringPtr("automatic"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: iops_requires_user_provisioned (nested message)
	// -------------------------------------------------------------------------

	ginkgo.Context("iops_requires_user_provisioned", func() {
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
				ApiVersion: "aws.openmcf.org/v1",
				Kind:       "AwsFsxOntapFileSystem",
				Metadata: &shared.CloudResourceMetadata{
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
				Metadata: &shared.CloudResourceMetadata{
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
				ApiVersion: "aws.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
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

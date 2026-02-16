package awsfsxopenzfsfilesystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestAwsFsxOpenzfsFileSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsFsxOpenzfsFileSystemSpec Validation Suite")
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

var _ = ginkgo.Describe("AwsFsxOpenzfsFileSystemSpec validations", func() {
	var spec *AwsFsxOpenzfsFileSystemSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: subnet_ids (at least 1), storage_capacity_gib >= 64,
		// throughput_capacity > 0.
		spec = &AwsFsxOpenzfsFileSystemSpec{
			SubnetIds:          []*foreignkeyv1.StringValueOrRef{strRef("subnet-abc123")},
			StorageCapacityGib: 256,
			ThroughputCapacity: 160,
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_1 deployment type", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_1")
		spec.ThroughputCapacity = 64
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SINGLE_AZ_2 deployment type", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_2")
		spec.ThroughputCapacity = 160
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
		spec.ThroughputCapacity = 160
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MULTI_AZ_1 with route_table_ids", func() {
		spec.DeploymentType = stringPtr("MULTI_AZ_1")
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.PreferredSubnetId = strRef("subnet-abc123")
		spec.RouteTableIds = []*foreignkeyv1.StringValueOrRef{strRef("rtb-abc123")}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MULTI_AZ_1 with endpoint_ip_address_range", func() {
		spec.DeploymentType = stringPtr("MULTI_AZ_1")
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.PreferredSubnetId = strRef("subnet-abc123")
		spec.EndpointIpAddressRange = "198.19.255.0/24"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full production SINGLE_AZ_2 configuration", func() {
		spec.DeploymentType = stringPtr("SINGLE_AZ_2")
		spec.StorageCapacityGib = 1024
		spec.ThroughputCapacity = 640
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{strRef("sg-abc123")}
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
		spec.DiskIopsConfiguration = &AwsFsxOpenzfsFileSystemDiskIopsConfiguration{
			Mode: stringPtr("AUTOMATIC"),
		}
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			DataCompressionType: stringPtr("ZSTD"),
			RecordSizeKib:       int32Ptr(128),
			NfsExports: &AwsFsxOpenzfsFileSystemNfsExports{
				ClientConfigurations: []*AwsFsxOpenzfsFileSystemNfsClientConfiguration{
					{Clients: "*", Options: []string{"rw", "crossmnt", "no_root_squash"}},
				},
			},
		}
		spec.AutomaticBackupRetentionDays = int32Ptr(7)
		spec.DailyAutomaticBackupStartTime = "05:00"
		spec.CopyTagsToBackups = true
		spec.CopyTagsToVolumes = true
		spec.WeeklyMaintenanceStartTime = "1:05:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full production MULTI_AZ_1 configuration", func() {
		spec.DeploymentType = stringPtr("MULTI_AZ_1")
		spec.StorageCapacityGib = 2048
		spec.ThroughputCapacity = 1280
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
		spec.DiskIopsConfiguration = &AwsFsxOpenzfsFileSystemDiskIopsConfiguration{
			Mode: stringPtr("USER_PROVISIONED"),
			Iops: 100000,
		}
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			DataCompressionType: stringPtr("ZSTD"),
			RecordSizeKib:       int32Ptr(128),
			CopyTagsToSnapshots: true,
			NfsExports: &AwsFsxOpenzfsFileSystemNfsExports{
				ClientConfigurations: []*AwsFsxOpenzfsFileSystemNfsClientConfiguration{
					{Clients: "10.0.0.0/16", Options: []string{"rw", "crossmnt"}},
				},
			},
			UserAndGroupQuotas: []*AwsFsxOpenzfsFileSystemUserAndGroupQuota{
				{Id: 1000, StorageCapacityQuotaGib: 100, Type: "USER"},
				{Id: 1001, StorageCapacityQuotaGib: 200, Type: "GROUP"},
			},
		}
		spec.AutomaticBackupRetentionDays = int32Ptr(14)
		spec.DailyAutomaticBackupStartTime = "03:00"
		spec.CopyTagsToBackups = true
		spec.CopyTagsToVolumes = true
		spec.WeeklyMaintenanceStartTime = "7:02:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Root volume configuration happy paths
	// -------------------------------------------------------------------------

	ginkgo.It("accepts ZSTD compression on root volume", func() {
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			DataCompressionType: stringPtr("ZSTD"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts LZ4 compression on root volume", func() {
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			DataCompressionType: stringPtr("LZ4"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts NONE compression on root volume", func() {
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			DataCompressionType: stringPtr("NONE"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts record_size_kib = 4", func() {
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			RecordSizeKib: int32Ptr(4),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts record_size_kib = 1024", func() {
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			RecordSizeKib: int32Ptr(1024),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts NFS exports with multiple client configurations", func() {
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			NfsExports: &AwsFsxOpenzfsFileSystemNfsExports{
				ClientConfigurations: []*AwsFsxOpenzfsFileSystemNfsClientConfiguration{
					{Clients: "*", Options: []string{"rw", "no_root_squash"}},
					{Clients: "10.0.0.0/8", Options: []string{"ro", "root_squash"}},
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts user and group quotas", func() {
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			UserAndGroupQuotas: []*AwsFsxOpenzfsFileSystemUserAndGroupQuota{
				{Id: 0, StorageCapacityQuotaGib: 50, Type: "USER"},
				{Id: 1000, StorageCapacityQuotaGib: 100, Type: "GROUP"},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts read-only root volume", func() {
		spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
			ReadOnly: true,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Disk IOPS configuration happy paths
	// -------------------------------------------------------------------------

	ginkgo.It("accepts AUTOMATIC IOPS mode", func() {
		spec.DiskIopsConfiguration = &AwsFsxOpenzfsFileSystemDiskIopsConfiguration{
			Mode: stringPtr("AUTOMATIC"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts USER_PROVISIONED IOPS mode with iops", func() {
		spec.DiskIopsConfiguration = &AwsFsxOpenzfsFileSystemDiskIopsConfiguration{
			Mode: stringPtr("USER_PROVISIONED"),
			Iops: 50000,
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

		ginkgo.It("fails when storage_capacity_gib is below minimum", func() {
			spec.StorageCapacityGib = 32
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when storage_capacity_gib is zero", func() {
			spec.StorageCapacityGib = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when throughput_capacity is zero", func() {
			spec.ThroughputCapacity = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when throughput_capacity is negative", func() {
			spec.ThroughputCapacity = -1
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

		ginkgo.It("fails for Lustre deployment type SCRATCH_2", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
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
			spec.ThroughputCapacity = 64
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
			spec.ThroughputCapacity = 64
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
	// CEL: iops_mode_valid (nested message)
	// -------------------------------------------------------------------------

	ginkgo.Context("iops_mode_valid", func() {
		ginkgo.It("fails when IOPS mode is invalid", func() {
			spec.DiskIopsConfiguration = &AwsFsxOpenzfsFileSystemDiskIopsConfiguration{
				Mode: stringPtr("MANUAL"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when IOPS mode is lowercase", func() {
			spec.DiskIopsConfiguration = &AwsFsxOpenzfsFileSystemDiskIopsConfiguration{
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
			spec.DiskIopsConfiguration = &AwsFsxOpenzfsFileSystemDiskIopsConfiguration{
				Mode: stringPtr("AUTOMATIC"),
				Iops: 50000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when iops is set without explicit mode", func() {
			spec.DiskIopsConfiguration = &AwsFsxOpenzfsFileSystemDiskIopsConfiguration{
				Iops: 50000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts zero iops with AUTOMATIC mode", func() {
			spec.DiskIopsConfiguration = &AwsFsxOpenzfsFileSystemDiskIopsConfiguration{
				Mode: stringPtr("AUTOMATIC"),
				Iops: 0,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: data_compression_type_valid (root volume)
	// -------------------------------------------------------------------------

	ginkgo.Context("data_compression_type_valid", func() {
		ginkgo.It("fails when compression type is invalid", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				DataCompressionType: stringPtr("GZIP"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when compression type is lowercase", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				DataCompressionType: stringPtr("zstd"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: record_size_kib_valid (root volume)
	// -------------------------------------------------------------------------

	ginkgo.Context("record_size_kib_valid", func() {
		ginkgo.It("fails when record_size_kib is not a valid value", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				RecordSizeKib: int32Ptr(100),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when record_size_kib is 2", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				RecordSizeKib: int32Ptr(2),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when record_size_kib is 2048", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				RecordSizeKib: int32Ptr(2048),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts all valid record_size_kib values", func() {
			for _, size := range []int32{4, 8, 16, 32, 64, 128, 256, 512, 1024} {
				spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
					RecordSizeKib: int32Ptr(size),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil(), "record_size_kib %d should be valid", size)
			}
		})
	})

	// -------------------------------------------------------------------------
	// CEL: quota_type_valid (nested message)
	// -------------------------------------------------------------------------

	ginkgo.Context("quota_type_valid", func() {
		ginkgo.It("fails when quota type is invalid", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				UserAndGroupQuotas: []*AwsFsxOpenzfsFileSystemUserAndGroupQuota{
					{Id: 1000, StorageCapacityQuotaGib: 100, Type: "INVALID"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when quota type is lowercase", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				UserAndGroupQuotas: []*AwsFsxOpenzfsFileSystemUserAndGroupQuota{
					{Id: 1000, StorageCapacityQuotaGib: 100, Type: "user"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when quota type is empty", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				UserAndGroupQuotas: []*AwsFsxOpenzfsFileSystemUserAndGroupQuota{
					{Id: 1000, StorageCapacityQuotaGib: 100, Type: ""},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts USER quota type", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				UserAndGroupQuotas: []*AwsFsxOpenzfsFileSystemUserAndGroupQuota{
					{Id: 1000, StorageCapacityQuotaGib: 100, Type: "USER"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts GROUP quota type", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				UserAndGroupQuotas: []*AwsFsxOpenzfsFileSystemUserAndGroupQuota{
					{Id: 100, StorageCapacityQuotaGib: 500, Type: "GROUP"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// NFS client configuration validations
	// -------------------------------------------------------------------------

	ginkgo.Context("NFS client configuration", func() {
		ginkgo.It("fails when clients is empty", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				NfsExports: &AwsFsxOpenzfsFileSystemNfsExports{
					ClientConfigurations: []*AwsFsxOpenzfsFileSystemNfsClientConfiguration{
						{Clients: "", Options: []string{"rw"}},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when options is empty", func() {
			spec.RootVolumeConfiguration = &AwsFsxOpenzfsFileSystemRootVolumeConfiguration{
				NfsExports: &AwsFsxOpenzfsFileSystemNfsExports{
					ClientConfigurations: []*AwsFsxOpenzfsFileSystemNfsClientConfiguration{
						{Clients: "*", Options: []string{}},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})

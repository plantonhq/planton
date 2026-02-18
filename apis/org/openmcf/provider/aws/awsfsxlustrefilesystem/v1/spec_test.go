package awsfsxlustrefilesystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsFsxLustreFileSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsFsxLustreFileSystemSpec Validation Suite")
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

var _ = ginkgo.Describe("AwsFsxLustreFileSystemSpec validations", func() {
	var spec *AwsFsxLustreFileSystemSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: subnet_id required, storage_capacity_gib >= 1200.
		spec = &AwsFsxLustreFileSystemSpec{
			Region:             "us-west-2",
			SubnetId:           strRef("subnet-abc123"),
			StorageCapacityGib: 1200,
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SCRATCH_1 deployment type", func() {
		spec.DeploymentType = stringPtr("SCRATCH_1")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SCRATCH_2 deployment type", func() {
		spec.DeploymentType = stringPtr("SCRATCH_2")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts PERSISTENT_1 deployment type with throughput", func() {
		spec.DeploymentType = stringPtr("PERSISTENT_1")
		spec.PerUnitStorageThroughput = 50
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts PERSISTENT_2 deployment type with throughput", func() {
		spec.DeploymentType = stringPtr("PERSISTENT_2")
		spec.PerUnitStorageThroughput = 125
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts HDD storage type with PERSISTENT_1", func() {
		spec.DeploymentType = stringPtr("PERSISTENT_1")
		spec.StorageType = stringPtr("HDD")
		spec.StorageCapacityGib = 6000
		spec.PerUnitStorageThroughput = 12
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts SSD storage type", func() {
		spec.StorageType = stringPtr("SSD")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts LZ4 data compression type", func() {
		spec.DataCompressionType = stringPtr("LZ4")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts NONE data compression type", func() {
		spec.DataCompressionType = stringPtr("NONE")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts import_path on SCRATCH_2", func() {
		spec.DeploymentType = stringPtr("SCRATCH_2")
		spec.ImportPath = "s3://my-bucket/data"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts import_path and export_path on SCRATCH_1", func() {
		spec.DeploymentType = stringPtr("SCRATCH_1")
		spec.ImportPath = "s3://my-bucket/input"
		spec.ExportPath = "s3://my-bucket/output"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts metadata_configuration on PERSISTENT_2", func() {
		spec.DeploymentType = stringPtr("PERSISTENT_2")
		spec.PerUnitStorageThroughput = 125
		spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts metadata_configuration with USER_PROVISIONED mode and iops", func() {
		spec.DeploymentType = stringPtr("PERSISTENT_2")
		spec.PerUnitStorageThroughput = 125
		spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
			Mode: stringPtr("USER_PROVISIONED"),
			Iops: 3000,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts log_configuration with valid level", func() {
		spec.LogConfiguration = &AwsFsxLustreFileSystemLogConfiguration{
			Destination: strRef("arn:aws:logs:us-east-1:123456789012:log-group:/aws/fsx/lustre"),
			Level:       stringPtr("WARN_ERROR"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts log_configuration with DISABLED level", func() {
		spec.LogConfiguration = &AwsFsxLustreFileSystemLogConfiguration{
			Level: stringPtr("DISABLED"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts log_configuration with WARN_ONLY level", func() {
		spec.LogConfiguration = &AwsFsxLustreFileSystemLogConfiguration{
			Level: stringPtr("WARN_ONLY"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts log_configuration with ERROR_ONLY level", func() {
		spec.LogConfiguration = &AwsFsxLustreFileSystemLogConfiguration{
			Level: stringPtr("ERROR_ONLY"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready PERSISTENT_2 configuration", func() {
		spec.DeploymentType = stringPtr("PERSISTENT_2")
		spec.StorageCapacityGib = 2400
		spec.StorageType = stringPtr("SSD")
		spec.PerUnitStorageThroughput = 250
		spec.DataCompressionType = stringPtr("LZ4")
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{strRef("sg-abc123")}
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
		spec.AutomaticBackupRetentionDays = int32Ptr(7)
		spec.DailyAutomaticBackupStartTime = "05:00"
		spec.CopyTagsToBackups = true
		spec.WeeklyMaintenanceStartTime = "1:05:00"
		spec.LogConfiguration = &AwsFsxLustreFileSystemLogConfiguration{
			Destination: strRef("arn:aws:logs:us-east-1:123456789012:log-group:/aws/fsx/lustre"),
			Level:       stringPtr("WARN_ERROR"),
		}
		spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
			Mode: stringPtr("AUTOMATIC"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level validation failures
	// -------------------------------------------------------------------------

	ginkgo.Context("field-level validations", func() {
		ginkgo.It("fails when subnet_id is missing", func() {
			spec.SubnetId = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when storage_capacity_gib is below minimum", func() {
			spec.StorageCapacityGib = 100
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when storage_capacity_gib is zero", func() {
			spec.StorageCapacityGib = 0
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

		ginkgo.It("fails when deployment_type is lowercase scratch_2", func() {
			spec.DeploymentType = stringPtr("scratch_2")
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
	// CEL: hdd_requires_persistent_1
	// -------------------------------------------------------------------------

	ginkgo.Context("hdd_requires_persistent_1", func() {
		ginkgo.It("fails when HDD is used with SCRATCH_2", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			spec.StorageType = stringPtr("HDD")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when HDD is used with PERSISTENT_2", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.StorageType = stringPtr("HDD")
			spec.PerUnitStorageThroughput = 125
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when HDD is used with SCRATCH_1", func() {
			spec.DeploymentType = stringPtr("SCRATCH_1")
			spec.StorageType = stringPtr("HDD")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: throughput_requires_persistent
	// -------------------------------------------------------------------------

	ginkgo.Context("throughput_requires_persistent", func() {
		ginkgo.It("fails when per_unit_storage_throughput is set with SCRATCH_2", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			spec.PerUnitStorageThroughput = 50
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when per_unit_storage_throughput is set with SCRATCH_1", func() {
			spec.DeploymentType = stringPtr("SCRATCH_1")
			spec.PerUnitStorageThroughput = 50
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts zero throughput with SCRATCH deployment", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			spec.PerUnitStorageThroughput = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: data_compression_type_valid
	// -------------------------------------------------------------------------

	ginkgo.Context("data_compression_type_valid", func() {
		ginkgo.It("fails when data_compression_type is invalid", func() {
			spec.DataCompressionType = stringPtr("GZIP")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when data_compression_type is lowercase lz4", func() {
			spec.DataCompressionType = stringPtr("lz4")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: import_path_requires_scratch
	// -------------------------------------------------------------------------

	ginkgo.Context("import_path_requires_scratch", func() {
		ginkgo.It("fails when import_path is set on PERSISTENT_1", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_1")
			spec.PerUnitStorageThroughput = 50
			spec.ImportPath = "s3://my-bucket/data"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when import_path is set on PERSISTENT_2", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.PerUnitStorageThroughput = 125
			spec.ImportPath = "s3://my-bucket/data"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts empty import_path on PERSISTENT_1", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_1")
			spec.PerUnitStorageThroughput = 50
			spec.ImportPath = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: export_requires_import
	// -------------------------------------------------------------------------

	ginkgo.Context("export_requires_import", func() {
		ginkgo.It("fails when export_path is set without import_path", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			spec.ExportPath = "s3://my-bucket/output"
			spec.ImportPath = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts export_path when import_path is set", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			spec.ImportPath = "s3://my-bucket/input"
			spec.ExportPath = "s3://my-bucket/output"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts empty export_path without import_path", func() {
			spec.ExportPath = ""
			spec.ImportPath = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: metadata_config_requires_persistent_2
	// -------------------------------------------------------------------------

	ginkgo.Context("metadata_config_requires_persistent_2", func() {
		ginkgo.It("fails when metadata_configuration is set on PERSISTENT_1", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_1")
			spec.PerUnitStorageThroughput = 50
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when metadata_configuration is set on SCRATCH_2", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts nil metadata_configuration on SCRATCH_2", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			spec.MetadataConfiguration = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: log_level_valid (nested message)
	// -------------------------------------------------------------------------

	ginkgo.Context("log_level_valid", func() {
		ginkgo.It("fails when log level is invalid", func() {
			spec.LogConfiguration = &AwsFsxLustreFileSystemLogConfiguration{
				Level: stringPtr("DEBUG"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when log level is lowercase", func() {
			spec.LogConfiguration = &AwsFsxLustreFileSystemLogConfiguration{
				Level: stringPtr("warn_error"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: metadata_mode_valid (nested message)
	// -------------------------------------------------------------------------

	ginkgo.Context("metadata_mode_valid", func() {
		ginkgo.It("fails when metadata mode is invalid", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.PerUnitStorageThroughput = 125
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
				Mode: stringPtr("MANUAL"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when metadata mode is lowercase", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.PerUnitStorageThroughput = 125
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
				Mode: stringPtr("automatic"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts AUTOMATIC mode", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.PerUnitStorageThroughput = 125
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
				Mode: stringPtr("AUTOMATIC"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts USER_PROVISIONED mode", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.PerUnitStorageThroughput = 125
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
				Mode: stringPtr("USER_PROVISIONED"),
				Iops: 3000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: iops_requires_user_provisioned (nested message)
	// -------------------------------------------------------------------------

	ginkgo.Context("iops_requires_user_provisioned", func() {
		ginkgo.It("fails when iops is set with AUTOMATIC mode", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.PerUnitStorageThroughput = 125
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
				Mode: stringPtr("AUTOMATIC"),
				Iops: 3000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when iops is set without explicit mode (defaults to AUTOMATIC)", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.PerUnitStorageThroughput = 125
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
				Iops: 3000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts zero iops with AUTOMATIC mode", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.PerUnitStorageThroughput = 125
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
				Mode: stringPtr("AUTOMATIC"),
				Iops: 0,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts iops with USER_PROVISIONED mode", func() {
			spec.DeploymentType = stringPtr("PERSISTENT_2")
			spec.PerUnitStorageThroughput = 125
			spec.MetadataConfiguration = &AwsFsxLustreFileSystemMetadataConfiguration{
				Mode: stringPtr("USER_PROVISIONED"),
				Iops: 6000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})

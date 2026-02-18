package awsfsxwindowsfilesystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsFsxWindowsFileSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsFsxWindowsFileSystemSpec Validation Suite")
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

func boolPtr(b bool) *bool {
	return &b
}

var _ = ginkgo.Describe("AwsFsxWindowsFileSystemSpec validations", func() {
	var spec *AwsFsxWindowsFileSystemSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec using AWS Managed AD (simplest path):
		// - subnet_ids (at least 1)
		// - storage_capacity_gib >= 32
		// - throughput_capacity in valid set
		// - active_directory_id set (satisfies ad_required)
		spec = &AwsFsxWindowsFileSystemSpec{
			Region:             "us-west-2",
			SubnetIds:          []*foreignkeyv1.StringValueOrRef{strRef("subnet-abc123")},
			StorageCapacityGib: 256,
			ThroughputCapacity: 64,
			ActiveDirectoryId:  strRef("d-1234567890"),
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec with AWS Managed AD", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a valid spec with self-managed AD (direct credentials)", func() {
		spec.ActiveDirectoryId = nil
		spec.SelfManagedActiveDirectory = &AwsFsxWindowsFileSystemSelfManagedActiveDirectory{
			DomainName: "corp.example.com",
			DnsIps:     []string{"10.0.0.1", "10.0.0.2"},
			Username:   "admin",
			Password:   "P@ssw0rd!",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a valid spec with self-managed AD (Secrets Manager)", func() {
		spec.ActiveDirectoryId = nil
		spec.SelfManagedActiveDirectory = &AwsFsxWindowsFileSystemSelfManagedActiveDirectory{
			DomainName:                        "corp.example.com",
			DnsIps:                            []string{"10.0.0.1"},
			DomainJoinServiceAccountSecretArn: strRef("arn:aws:secretsmanager:us-east-1:123456789012:secret:fsx-ad-creds"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MULTI_AZ_1 with preferred_subnet_id", func() {
		spec.DeploymentType = stringPtr("MULTI_AZ_1")
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.PreferredSubnetId = strRef("subnet-abc123")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with audit logging enabled", func() {
		spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
			FileAccessAuditLogLevel:      stringPtr("SUCCESS_AND_FAILURE"),
			FileShareAccessAuditLogLevel: stringPtr("FAILURE_ONLY"),
			AuditLogDestination:          strRef("arn:aws:logs:us-east-1:123456789012:log-group:/aws/fsx/windows"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with USER_PROVISIONED disk IOPS", func() {
		spec.DiskIopsConfiguration = &AwsFsxWindowsFileSystemDiskIopsConfiguration{
			Mode: stringPtr("USER_PROVISIONED"),
			Iops: 100000,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full production configuration", func() {
		spec.DeploymentType = stringPtr("MULTI_AZ_1")
		spec.StorageCapacityGib = 2048
		spec.StorageType = stringPtr("SSD")
		spec.ThroughputCapacity = 2048
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-abc123"),
			strRef("subnet-def456"),
		}
		spec.PreferredSubnetId = strRef("subnet-abc123")
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{strRef("sg-abc123")}
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
		spec.DiskIopsConfiguration = &AwsFsxWindowsFileSystemDiskIopsConfiguration{
			Mode: stringPtr("USER_PROVISIONED"),
			Iops: 200000,
		}
		spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
			FileAccessAuditLogLevel:      stringPtr("SUCCESS_AND_FAILURE"),
			FileShareAccessAuditLogLevel: stringPtr("SUCCESS_AND_FAILURE"),
			AuditLogDestination:          strRef("arn:aws:logs:us-east-1:123456789012:log-group:/aws/fsx/windows"),
		}
		spec.AutomaticBackupRetentionDays = int32Ptr(30)
		spec.DailyAutomaticBackupStartTime = "03:00"
		spec.CopyTagsToBackups = true
		spec.SkipFinalBackup = boolPtr(false)
		spec.WeeklyMaintenanceStartTime = "7:02:00"
		spec.Aliases = []string{"finance.corp.example.com", "hr.corp.example.com"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Deployment type valid values
	// -------------------------------------------------------------------------

	ginkgo.Context("deployment_type_valid", func() {
		ginkgo.It("accepts SINGLE_AZ_1", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_1")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts MULTI_AZ_1", func() {
			spec.DeploymentType = stringPtr("MULTI_AZ_1")
			spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-abc123"),
				strRef("subnet-def456"),
			}
			spec.PreferredSubnetId = strRef("subnet-abc123")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails when deployment_type is invalid", func() {
			spec.DeploymentType = stringPtr("INVALID_TYPE")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("deployment_type"))
		})

		ginkgo.It("fails when deployment_type is lowercase", func() {
			spec.DeploymentType = stringPtr("single_az_2")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails for OpenZFS deployment type SINGLE_AZ_1 variant", func() {
			spec.DeploymentType = stringPtr("SCRATCH_2")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// Storage type valid values
	// -------------------------------------------------------------------------

	ginkgo.Context("storage_type_valid", func() {
		ginkgo.It("accepts SSD", func() {
			spec.StorageType = stringPtr("SSD")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts HDD with compatible deployment and storage", func() {
			spec.StorageType = stringPtr("HDD")
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.StorageCapacityGib = 2000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails when storage_type is invalid", func() {
			spec.StorageType = stringPtr("NVME")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("storage_type"))
		})

		ginkgo.It("fails when storage_type is lowercase", func() {
			spec.StorageType = stringPtr("ssd")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// HDD compatibility rules
	// -------------------------------------------------------------------------

	ginkgo.Context("hdd_requires_compatible_deployment", func() {
		ginkgo.It("fails when HDD with SINGLE_AZ_1", func() {
			spec.StorageType = stringPtr("HDD")
			spec.DeploymentType = stringPtr("SINGLE_AZ_1")
			spec.StorageCapacityGib = 2000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("HDD"))
		})

		ginkgo.It("accepts HDD with SINGLE_AZ_2", func() {
			spec.StorageType = stringPtr("HDD")
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.StorageCapacityGib = 2000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts HDD with MULTI_AZ_1", func() {
			spec.StorageType = stringPtr("HDD")
			spec.DeploymentType = stringPtr("MULTI_AZ_1")
			spec.StorageCapacityGib = 2000
			spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-abc123"),
				strRef("subnet-def456"),
			}
			spec.PreferredSubnetId = strRef("subnet-abc123")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("hdd_minimum_storage", func() {
		ginkgo.It("fails when HDD with less than 2000 GiB", func() {
			spec.StorageType = stringPtr("HDD")
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.StorageCapacityGib = 1999
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("2000"))
		})

		ginkgo.It("accepts HDD with exactly 2000 GiB", func() {
			spec.StorageType = stringPtr("HDD")
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.StorageCapacityGib = 2000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// Throughput capacity valid values
	// -------------------------------------------------------------------------

	ginkgo.Context("throughput_capacity_valid", func() {
		ginkgo.It("accepts all valid throughput values", func() {
			for _, tp := range []int32{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4608, 6144, 9216, 12288} {
				spec.ThroughputCapacity = tp
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil(), "throughput_capacity %d should be valid", tp)
			}
		})

		ginkgo.It("fails when throughput_capacity is 100 (invalid value)", func() {
			spec.ThroughputCapacity = 100
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("throughput_capacity"))
		})

		ginkgo.It("fails when throughput_capacity is 48 (between valid values)", func() {
			spec.ThroughputCapacity = 48
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
	// preferred_subnet_requires_multi_az
	// -------------------------------------------------------------------------

	ginkgo.Context("preferred_subnet_requires_multi_az", func() {
		ginkgo.It("fails when preferred_subnet_id is set on SINGLE_AZ_2", func() {
			spec.DeploymentType = stringPtr("SINGLE_AZ_2")
			spec.PreferredSubnetId = strRef("subnet-abc123")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("preferred_subnet_id"))
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
	// ad_required: exactly one of active_directory_id or self_managed_active_directory
	// -------------------------------------------------------------------------

	ginkgo.Context("ad_required", func() {
		ginkgo.It("fails when neither AD is specified", func() {
			spec.ActiveDirectoryId = nil
			spec.SelfManagedActiveDirectory = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("active_directory"))
		})

		ginkgo.It("fails when both AD types are specified", func() {
			spec.ActiveDirectoryId = strRef("d-1234567890")
			spec.SelfManagedActiveDirectory = &AwsFsxWindowsFileSystemSelfManagedActiveDirectory{
				DomainName:                        "corp.example.com",
				DnsIps:                            []string{"10.0.0.1"},
				DomainJoinServiceAccountSecretArn: strRef("arn:aws:secretsmanager:us-east-1:123456789012:secret:creds"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts only active_directory_id", func() {
			spec.ActiveDirectoryId = strRef("d-1234567890")
			spec.SelfManagedActiveDirectory = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts only self_managed_active_directory", func() {
			spec.ActiveDirectoryId = nil
			spec.SelfManagedActiveDirectory = &AwsFsxWindowsFileSystemSelfManagedActiveDirectory{
				DomainName:                        "corp.example.com",
				DnsIps:                            []string{"10.0.0.1"},
				DomainJoinServiceAccountSecretArn: strRef("arn:aws:secretsmanager:us-east-1:123456789012:secret:creds"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// Self-managed Active Directory validations
	// -------------------------------------------------------------------------

	ginkgo.Context("self_managed_active_directory", func() {
		ginkgo.BeforeEach(func() {
			spec.ActiveDirectoryId = nil
			spec.SelfManagedActiveDirectory = &AwsFsxWindowsFileSystemSelfManagedActiveDirectory{
				DomainName: "corp.example.com",
				DnsIps:     []string{"10.0.0.1"},
				Username:   "admin",
				Password:   "P@ssw0rd!",
			}
		})

		ginkgo.It("fails when domain_name is empty", func() {
			spec.SelfManagedActiveDirectory.DomainName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when dns_ips is empty", func() {
			spec.SelfManagedActiveDirectory.DnsIps = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts dns_ips with one IP", func() {
			spec.SelfManagedActiveDirectory.DnsIps = []string{"10.0.0.1"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts dns_ips with two IPs", func() {
			spec.SelfManagedActiveDirectory.DnsIps = []string{"10.0.0.1", "10.0.0.2"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails when dns_ips has more than 2 IPs", func() {
			spec.SelfManagedActiveDirectory.DnsIps = []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when username is set without password", func() {
			spec.SelfManagedActiveDirectory.Username = "admin"
			spec.SelfManagedActiveDirectory.Password = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("credentials"))
		})

		ginkgo.It("fails when password is set without username", func() {
			spec.SelfManagedActiveDirectory.Username = ""
			spec.SelfManagedActiveDirectory.Password = "P@ssw0rd!"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("credentials"))
		})

		ginkgo.It("fails when both direct credentials and secret_arn are set", func() {
			spec.SelfManagedActiveDirectory.Username = "admin"
			spec.SelfManagedActiveDirectory.Password = "P@ssw0rd!"
			spec.SelfManagedActiveDirectory.DomainJoinServiceAccountSecretArn = strRef("arn:aws:secretsmanager:us-east-1:123456789012:secret:creds")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("not both"))
		})

		ginkgo.It("fails when neither direct credentials nor secret_arn are set", func() {
			spec.SelfManagedActiveDirectory.Username = ""
			spec.SelfManagedActiveDirectory.Password = ""
			spec.SelfManagedActiveDirectory.DomainJoinServiceAccountSecretArn = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts with optional file_system_administrators_group", func() {
			spec.SelfManagedActiveDirectory.FileSystemAdministratorsGroup = stringPtr("FSxAdmins")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts with organizational_unit_distinguished_name", func() {
			spec.SelfManagedActiveDirectory.OrganizationalUnitDistinguishedName = "OU=FSx,DC=corp,DC=example,DC=com"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// Audit log configuration
	// -------------------------------------------------------------------------

	ginkgo.Context("audit_log_configuration", func() {
		ginkgo.It("accepts all valid file_access_audit_log_level values", func() {
			for _, level := range []string{"DISABLED", "SUCCESS_ONLY", "FAILURE_ONLY", "SUCCESS_AND_FAILURE"} {
				spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
					FileAccessAuditLogLevel: stringPtr(level),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil(), "file_access_audit_log_level %q should be valid", level)
			}
		})

		ginkgo.It("accepts all valid file_share_access_audit_log_level values", func() {
			for _, level := range []string{"DISABLED", "SUCCESS_ONLY", "FAILURE_ONLY", "SUCCESS_AND_FAILURE"} {
				spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
					FileShareAccessAuditLogLevel: stringPtr(level),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil(), "file_share_access_audit_log_level %q should be valid", level)
			}
		})

		ginkgo.It("fails when file_access_audit_log_level is invalid", func() {
			spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
				FileAccessAuditLogLevel: stringPtr("ALL"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("file_access_audit_log_level"))
		})

		ginkgo.It("fails when file_share_access_audit_log_level is invalid", func() {
			spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
				FileShareAccessAuditLogLevel: stringPtr("VERBOSE"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("file_share_access_audit_log_level"))
		})

		ginkgo.It("fails when audit_log_destination is set with both levels DISABLED", func() {
			spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
				FileAccessAuditLogLevel:      stringPtr("DISABLED"),
				FileShareAccessAuditLogLevel: stringPtr("DISABLED"),
				AuditLogDestination:          strRef("arn:aws:logs:us-east-1:123456789012:log-group:/aws/fsx/windows"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("audit_log_destination"))
		})

		ginkgo.It("accepts audit_log_destination when file_access level is enabled", func() {
			spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
				FileAccessAuditLogLevel:      stringPtr("SUCCESS_ONLY"),
				FileShareAccessAuditLogLevel: stringPtr("DISABLED"),
				AuditLogDestination:          strRef("arn:aws:logs:us-east-1:123456789012:log-group:/aws/fsx/windows"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts audit_log_destination when file_share level is enabled", func() {
			spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
				FileAccessAuditLogLevel:      stringPtr("DISABLED"),
				FileShareAccessAuditLogLevel: stringPtr("FAILURE_ONLY"),
				AuditLogDestination:          strRef("arn:aws:logs:us-east-1:123456789012:log-group:/aws/fsx/windows"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts audit config without destination when levels are enabled", func() {
			spec.AuditLogConfiguration = &AwsFsxWindowsFileSystemAuditLogConfiguration{
				FileAccessAuditLogLevel:      stringPtr("SUCCESS_AND_FAILURE"),
				FileShareAccessAuditLogLevel: stringPtr("SUCCESS_AND_FAILURE"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// Disk IOPS configuration
	// -------------------------------------------------------------------------

	ginkgo.Context("disk_iops_configuration", func() {
		ginkgo.It("accepts AUTOMATIC mode", func() {
			spec.DiskIopsConfiguration = &AwsFsxWindowsFileSystemDiskIopsConfiguration{
				Mode: stringPtr("AUTOMATIC"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts USER_PROVISIONED mode with iops", func() {
			spec.DiskIopsConfiguration = &AwsFsxWindowsFileSystemDiskIopsConfiguration{
				Mode: stringPtr("USER_PROVISIONED"),
				Iops: 50000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails when mode is invalid", func() {
			spec.DiskIopsConfiguration = &AwsFsxWindowsFileSystemDiskIopsConfiguration{
				Mode: stringPtr("MANUAL"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("mode"))
		})

		ginkgo.It("fails when mode is lowercase", func() {
			spec.DiskIopsConfiguration = &AwsFsxWindowsFileSystemDiskIopsConfiguration{
				Mode: stringPtr("automatic"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when iops is set with AUTOMATIC mode", func() {
			spec.DiskIopsConfiguration = &AwsFsxWindowsFileSystemDiskIopsConfiguration{
				Mode: stringPtr("AUTOMATIC"),
				Iops: 50000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("iops"))
		})

		ginkgo.It("fails when iops is set without explicit mode", func() {
			spec.DiskIopsConfiguration = &AwsFsxWindowsFileSystemDiskIopsConfiguration{
				Iops: 50000,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts zero iops with AUTOMATIC mode", func() {
			spec.DiskIopsConfiguration = &AwsFsxWindowsFileSystemDiskIopsConfiguration{
				Mode: stringPtr("AUTOMATIC"),
				Iops: 0,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// Backup configuration
	// -------------------------------------------------------------------------

	ginkgo.Context("backup_time_requires_retention", func() {
		ginkgo.It("fails when backup start time set without retention days", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(0)
			spec.DailyAutomaticBackupStartTime = "03:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("daily_automatic_backup_start_time"))
		})

		ginkgo.It("accepts backup start time with positive retention days", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(7)
			spec.DailyAutomaticBackupStartTime = "03:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts empty backup start time with zero retention", func() {
			spec.AutomaticBackupRetentionDays = int32Ptr(0)
			spec.DailyAutomaticBackupStartTime = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
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

		ginkgo.It("fails when subnet_ids is an empty slice", func() {
			spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when storage_capacity_gib is below 32", func() {
			spec.StorageCapacityGib = 31
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts storage_capacity_gib exactly at 32", func() {
			spec.StorageCapacityGib = 32
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails when storage_capacity_gib is zero", func() {
			spec.StorageCapacityGib = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// Aliases boundary
	// -------------------------------------------------------------------------

	ginkgo.Context("aliases", func() {
		ginkgo.It("accepts 50 aliases (boundary)", func() {
			aliases := make([]string, 50)
			for i := range aliases {
				aliases[i] = "alias" + string(rune('a'+i%26)) + ".corp.example.com"
			}
			spec.Aliases = aliases
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails when aliases exceed 50", func() {
			aliases := make([]string, 51)
			for i := range aliases {
				aliases[i] = "alias" + string(rune('a'+i%26)) + ".corp.example.com"
			}
			spec.Aliases = aliases
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts empty aliases", func() {
			spec.Aliases = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// Optional field defaults
	// -------------------------------------------------------------------------

	ginkgo.Context("optional field defaults", func() {
		ginkgo.It("accepts spec without setting optional deployment_type", func() {
			spec.DeploymentType = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec without setting optional storage_type", func() {
			spec.StorageType = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec without setting optional automatic_backup_retention_days", func() {
			spec.AutomaticBackupRetentionDays = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec without setting optional skip_final_backup", func() {
			spec.SkipFinalBackup = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// Miscellaneous valid configurations
	// -------------------------------------------------------------------------

	ginkgo.Context("miscellaneous", func() {
		ginkgo.It("accepts security_group_ids", func() {
			spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
				strRef("sg-abc123"),
				strRef("sg-def456"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts kms_key_id", func() {
			spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/test-key")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts copy_tags_to_backups true", func() {
			spec.CopyTagsToBackups = true
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts weekly_maintenance_start_time", func() {
			spec.WeeklyMaintenanceStartTime = "7:02:00"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})

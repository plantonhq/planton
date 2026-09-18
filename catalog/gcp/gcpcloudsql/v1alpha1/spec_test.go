package gcpcloudsqlv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudSqlSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func fromRef(kind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Name: name,
			},
		},
	}
}

func ptr(s string) *string {
	return &s
}

func intPtr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}

var _ = ginkgo.Describe("GcpCloudSqlSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimalPostgres := func() *GcpCloudSql {
		return &GcpCloudSql{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudSql",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-postgres",
			},
			Spec: &GcpCloudSqlSpec{
				InstanceName:    "orders-db",
				Region:          "us-central1",
				DatabaseEngine:  "POSTGRESQL",
				DatabaseVersion: "POSTGRES_16",
				Tier:            "db-custom-2-7680",
			},
		}
	}

	expectValid := func(r *GcpCloudSql) {
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	}

	expectInvalid := func(r *GcpCloudSql, substr string) {
		err := validator.Validate(r)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), substr)).To(
			gomega.BeTrue(), "expected error to contain %q, got: %s", substr, err)
	}

	ginkgo.Context("required fields", func() {
		ginkgo.It("accepts a minimal PostgreSQL spec", func() {
			expectValid(minimalPostgres())
		})

		ginkgo.It("rejects a missing instance_name", func() {
			r := minimalPostgres()
			r.Spec.InstanceName = ""
			expectInvalid(r, "instance_name")
		})

		ginkgo.It("rejects an invalid instance_name", func() {
			r := minimalPostgres()
			r.Spec.InstanceName = "Bad_Name"
			expectInvalid(r, "instance_name")
		})

		ginkgo.It("rejects a missing region", func() {
			r := minimalPostgres()
			r.Spec.Region = ""
			expectInvalid(r, "region")
		})

		ginkgo.It("rejects a malformed region", func() {
			r := minimalPostgres()
			r.Spec.Region = "uscentral"
			expectInvalid(r, "region")
		})

		ginkgo.It("rejects a missing database_engine", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = ""
			expectInvalid(r, "database_engine")
		})

		ginkgo.It("rejects an unknown database_engine", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "ORACLE"
			expectInvalid(r, "database_engine")
		})

		ginkgo.It("rejects a missing tier", func() {
			r := minimalPostgres()
			r.Spec.Tier = ""
			expectInvalid(r, "tier")
		})

		ginkgo.It("accepts a project_id reference", func() {
			r := minimalPostgres()
			r.Spec.ProjectId = fromRef("GcpProject", "my-project", "status.outputs.project_id")
			expectValid(r)
		})
	})

	ginkgo.Context("engine and version coherence", func() {
		ginkgo.It("rejects a MySQL version on a PostgreSQL engine", func() {
			r := minimalPostgres()
			r.Spec.DatabaseVersion = "MYSQL_8_0"
			expectInvalid(r, "must match database_engine")
		})

		ginkgo.It("accepts MySQL engine with a MySQL version", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "MYSQL"
			r.Spec.DatabaseVersion = "MYSQL_8_0"
			expectValid(r)
		})

		ginkgo.It("rejects SQL Server without a root password", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "SQLSERVER"
			r.Spec.DatabaseVersion = "SQLSERVER_2022_STANDARD"
			expectInvalid(r, "required for SQL Server")
		})

		ginkgo.It("accepts SQL Server with a root password", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "SQLSERVER"
			r.Spec.DatabaseVersion = "SQLSERVER_2022_STANDARD"
			r.Spec.RootPassword = "SqlServer123!"
			expectValid(r)
		})

		ginkgo.It("rejects a short root_password", func() {
			r := minimalPostgres()
			r.Spec.RootPassword = "short"
			expectInvalid(r, "root_password")
		})
	})

	ginkgo.Context("edition and data cache", func() {
		ginkgo.It("rejects an unknown edition", func() {
			r := minimalPostgres()
			r.Spec.Edition = ptr("PREMIUM")
			expectInvalid(r, "edition")
		})

		ginkgo.It("rejects data cache on ENTERPRISE", func() {
			r := minimalPostgres()
			r.Spec.DataCacheEnabled = true
			expectInvalid(r, "requires edition ENTERPRISE_PLUS")
		})

		ginkgo.It("accepts data cache on ENTERPRISE_PLUS", func() {
			r := minimalPostgres()
			r.Spec.Edition = ptr("ENTERPRISE_PLUS")
			r.Spec.DataCacheEnabled = true
			expectValid(r)
		})
	})

	ginkgo.Context("availability and backups", func() {
		ginkgo.It("rejects REGIONAL without backups", func() {
			r := minimalPostgres()
			r.Spec.AvailabilityType = ptr("REGIONAL")
			expectInvalid(r, "requires automated backups")
		})

		ginkgo.It("accepts REGIONAL with backups enabled", func() {
			r := minimalPostgres()
			r.Spec.AvailabilityType = ptr("REGIONAL")
			r.Spec.Backup = &GcpCloudSqlBackup{Enabled: true, PointInTimeRecoveryEnabled: true}
			expectValid(r)
		})

		ginkgo.It("rejects REGIONAL MySQL without binary logs", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "MYSQL"
			r.Spec.DatabaseVersion = "MYSQL_8_0"
			r.Spec.AvailabilityType = ptr("REGIONAL")
			r.Spec.Backup = &GcpCloudSqlBackup{Enabled: true}
			expectInvalid(r, "requires backup.binary_log_enabled")
		})

		ginkgo.It("accepts REGIONAL MySQL with binary logs", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "MYSQL"
			r.Spec.DatabaseVersion = "MYSQL_8_0"
			r.Spec.AvailabilityType = ptr("REGIONAL")
			r.Spec.Backup = &GcpCloudSqlBackup{Enabled: true, BinaryLogEnabled: true}
			expectValid(r)
		})

		ginkgo.It("rejects an unknown availability_type", func() {
			r := minimalPostgres()
			r.Spec.AvailabilityType = ptr("MULTI_REGION")
			expectInvalid(r, "availability_type")
		})

		ginkgo.It("rejects PITR on MySQL", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "MYSQL"
			r.Spec.DatabaseVersion = "MYSQL_8_0"
			r.Spec.Backup = &GcpCloudSqlBackup{Enabled: true, PointInTimeRecoveryEnabled: true}
			expectInvalid(r, "POSTGRESQL and SQLSERVER only")
		})

		ginkgo.It("rejects binary logs on PostgreSQL", func() {
			r := minimalPostgres()
			r.Spec.Backup = &GcpCloudSqlBackup{Enabled: true, BinaryLogEnabled: true}
			expectInvalid(r, "applies to MYSQL only")
		})

		ginkgo.It("rejects backup settings without enabled", func() {
			r := minimalPostgres()
			r.Spec.Backup = &GcpCloudSqlBackup{StartTime: "03:00"}
			expectInvalid(r, "require backup.enabled")
		})

		ginkgo.It("rejects a malformed start_time", func() {
			r := minimalPostgres()
			r.Spec.Backup = &GcpCloudSqlBackup{Enabled: true, StartTime: "25:00"}
			expectInvalid(r, "start_time")
		})

		ginkgo.It("rejects transaction log retention above 35 days", func() {
			r := minimalPostgres()
			r.Spec.Backup = &GcpCloudSqlBackup{Enabled: true, TransactionLogRetentionDays: intPtr(40)}
			expectInvalid(r, "transaction_log_retention_days")
		})

		ginkgo.It("accepts a full backup block", func() {
			r := minimalPostgres()
			r.Spec.Backup = &GcpCloudSqlBackup{
				Enabled:                     true,
				StartTime:                   "03:00",
				Location:                    "us",
				PointInTimeRecoveryEnabled:  true,
				TransactionLogRetentionDays: intPtr(7),
				RetainedBackups:             intPtr(14),
			}
			expectValid(r)
		})
	})

	ginkgo.Context("disk", func() {
		ginkgo.It("rejects an unknown disk type", func() {
			r := minimalPostgres()
			r.Spec.Disk = &GcpCloudSqlDisk{Type: ptr("LOCAL_SSD")}
			expectInvalid(r, "disk")
		})

		ginkgo.It("rejects a disk below 10 GB", func() {
			r := minimalPostgres()
			r.Spec.Disk = &GcpCloudSqlDisk{SizeGb: intPtr(5)}
			expectInvalid(r, "size_gb")
		})

		ginkgo.It("rejects a hyperdisk below 20 GB", func() {
			r := minimalPostgres()
			r.Spec.Disk = &GcpCloudSqlDisk{Type: ptr("HYPERDISK_BALANCED"), SizeGb: intPtr(10)}
			expectInvalid(r, "HYPERDISK_BALANCED disks require")
		})

		ginkgo.It("accepts a hyperdisk at 20 GB", func() {
			r := minimalPostgres()
			r.Spec.Disk = &GcpCloudSqlDisk{Type: ptr("HYPERDISK_BALANCED"), SizeGb: intPtr(20)}
			expectValid(r)
		})
	})

	ginkgo.Context("network", func() {
		ginkgo.It("rejects a network block with no connectivity path", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{}
			expectInvalid(r, "at least one connectivity path")
		})

		ginkgo.It("accepts private-only connectivity via a reference", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				PrivateNetwork: fromRef("GcpVpcNetwork", "my-vpc", "status.outputs.network_id"),
			}
			expectValid(r)
		})

		ginkgo.It("accepts public-only connectivity", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{Ipv4Enabled: true}
			expectValid(r)
		})

		ginkgo.It("accepts PSC-only connectivity", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Psc: &GcpCloudSqlPscConfig{Enabled: true},
			}
			expectValid(r)
		})

		ginkgo.It("rejects authorized networks without ipv4", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				PrivateNetwork:     litRef("projects/p/global/networks/n"),
				AuthorizedNetworks: []*GcpCloudSqlAuthorizedNetwork{{Value: "10.0.0.0/8"}},
			}
			expectInvalid(r, "apply to the public IP")
		})

		ginkgo.It("rejects a malformed authorized network CIDR", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled:        true,
				AuthorizedNetworks: []*GcpCloudSqlAuthorizedNetwork{{Value: "not-a-cidr"}},
			}
			expectInvalid(r, "value")
		})

		ginkgo.It("rejects allocated_ip_range without a private network", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled:      true,
				AllocatedIpRange: "psa-range",
			}
			expectInvalid(r, "allocated_ip_range applies to private IP")
		})

		ginkgo.It("rejects private path without a private network", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled:                             true,
				EnablePrivatePathForGoogleCloudServices: true,
			}
			expectInvalid(r, "enable_private_path_for_google_cloud_services applies")
		})

		ginkgo.It("rejects an unknown ssl_mode", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{Ipv4Enabled: true, SslMode: "TLS_ONLY"}
			expectInvalid(r, "ssl_mode")
		})

		ginkgo.It("rejects CUSTOMER_MANAGED_CAS_CA without a pool", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled:  true,
				ServerCaMode: "CUSTOMER_MANAGED_CAS_CA",
			}
			expectInvalid(r, "server_ca_pool is required when")
		})

		ginkgo.It("rejects a pool without CUSTOMER_MANAGED_CAS_CA", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled:  true,
				ServerCaPool: "projects/p/locations/l/caPools/pool",
			}
			expectInvalid(r, "server_ca_pool applies only")
		})

		ginkgo.It("rejects PSC settings when psc is disabled", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled: true,
				Psc: &GcpCloudSqlPscConfig{
					AllowedConsumerProjects: []string{"consumer-project"},
				},
			}
			expectInvalid(r, "PSC settings apply only")
		})

		ginkgo.It("accepts a full PSC block", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Psc: &GcpCloudSqlPscConfig{
					Enabled:                 true,
					AllowedConsumerProjects: []string{"consumer-project"},
					AutoConnections: []*GcpCloudSqlPscAutoConnection{
						{ConsumerNetwork: "projects/p/global/networks/n"},
					},
				},
			}
			expectValid(r)
		})
	})

	ginkgo.Context("maintenance", func() {
		ginkgo.It("rejects a maintenance window without a day", func() {
			r := minimalPostgres()
			r.Spec.MaintenanceWindow = &GcpCloudSqlMaintenanceWindow{}
			expectInvalid(r, "day")
		})

		ginkgo.It("rejects an out-of-range day", func() {
			r := minimalPostgres()
			r.Spec.MaintenanceWindow = &GcpCloudSqlMaintenanceWindow{Day: 8}
			expectInvalid(r, "day")
		})

		ginkgo.It("accepts hour zero (midnight)", func() {
			r := minimalPostgres()
			r.Spec.MaintenanceWindow = &GcpCloudSqlMaintenanceWindow{Day: 7, Hour: intPtr(0)}
			expectValid(r)
		})

		ginkgo.It("rejects an unknown update_track", func() {
			r := minimalPostgres()
			r.Spec.MaintenanceWindow = &GcpCloudSqlMaintenanceWindow{Day: 7, UpdateTrack: "weekly"}
			expectInvalid(r, "update_track")
		})

		ginkgo.It("accepts the week5 update track", func() {
			r := minimalPostgres()
			r.Spec.MaintenanceWindow = &GcpCloudSqlMaintenanceWindow{Day: 7, UpdateTrack: "week5"}
			expectValid(r)
		})

		ginkgo.It("rejects a deny period missing its time", func() {
			r := minimalPostgres()
			r.Spec.DenyMaintenancePeriod = &GcpCloudSqlDenyMaintenancePeriod{
				StartDate: "2026-12-20",
				EndDate:   "2027-01-05",
			}
			expectInvalid(r, "time")
		})

		ginkgo.It("accepts a recurring deny period", func() {
			r := minimalPostgres()
			r.Spec.DenyMaintenancePeriod = &GcpCloudSqlDenyMaintenancePeriod{
				StartDate: "12-20",
				EndDate:   "01-05",
				Time:      "00:00:00",
			}
			expectValid(r)
		})
	})

	ginkgo.Context("insights", func() {
		ginkgo.It("rejects insights settings without enablement", func() {
			r := minimalPostgres()
			r.Spec.InsightsConfig = &GcpCloudSqlInsightsConfig{
				RecordApplicationTags: true,
			}
			expectInvalid(r, "insights settings apply only")
		})

		ginkgo.It("accepts a full insights block", func() {
			r := minimalPostgres()
			r.Spec.InsightsConfig = &GcpCloudSqlInsightsConfig{
				QueryInsightsEnabled:  true,
				QueryStringLength:     intPtr(2048),
				RecordApplicationTags: true,
				RecordClientAddress:   true,
				QueryPlansPerMinute:   intPtr(10),
			}
			expectValid(r)
		})

		ginkgo.It("rejects an out-of-range query_string_length", func() {
			r := minimalPostgres()
			r.Spec.InsightsConfig = &GcpCloudSqlInsightsConfig{
				QueryInsightsEnabled: true,
				QueryStringLength:    intPtr(9000),
			}
			expectInvalid(r, "query_string_length")
		})
	})

	ginkgo.Context("SQL Server-only surface", func() {
		ginkgo.It("rejects time_zone on PostgreSQL", func() {
			r := minimalPostgres()
			r.Spec.TimeZone = "Pacific Standard Time"
			expectInvalid(r, "time_zone applies to SQL Server")
		})

		ginkgo.It("rejects collation on MySQL", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "MYSQL"
			r.Spec.DatabaseVersion = "MYSQL_8_0"
			r.Spec.Collation = "SQL_Latin1_General_CP1_CI_AS"
			expectInvalid(r, "collation (server-level) applies")
		})

		ginkgo.It("rejects threads_per_core on PostgreSQL", func() {
			r := minimalPostgres()
			r.Spec.ThreadsPerCore = intPtr(2)
			expectInvalid(r, "threads_per_core applies to SQL Server")
		})

		ginkgo.It("rejects an audit config on PostgreSQL", func() {
			r := minimalPostgres()
			r.Spec.SqlServerAuditConfig = &GcpCloudSqlSqlServerAuditConfig{Bucket: "gs://audit"}
			expectInvalid(r, "sql_server_audit_config applies")
		})

		ginkgo.It("accepts the full SQL Server surface", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "SQLSERVER"
			r.Spec.DatabaseVersion = "SQLSERVER_2022_STANDARD"
			r.Spec.RootPassword = "SqlServer123!"
			r.Spec.TimeZone = "Pacific Standard Time"
			r.Spec.Collation = "SQL_Latin1_General_CP1_CI_AS"
			r.Spec.ThreadsPerCore = intPtr(1)
			r.Spec.ActiveDirectory = &GcpCloudSqlActiveDirectoryConfig{Domain: "ad.example.com"}
			r.Spec.SqlServerAuditConfig = &GcpCloudSqlSqlServerAuditConfig{
				Bucket:            "gs://audit-bucket",
				RetentionInterval: "86400s",
				UploadInterval:    "1800s",
			}
			expectValid(r)
		})

		ginkgo.It("rejects a non-gs audit bucket", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "SQLSERVER"
			r.Spec.DatabaseVersion = "SQLSERVER_2022_STANDARD"
			r.Spec.RootPassword = "SqlServer123!"
			r.Spec.SqlServerAuditConfig = &GcpCloudSqlSqlServerAuditConfig{Bucket: "audit-bucket"}
			expectInvalid(r, "gs:// URI")
		})
	})

	ginkgo.Context("password validation policy", func() {
		ginkgo.It("rejects password_change_interval on MySQL", func() {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "MYSQL"
			r.Spec.DatabaseVersion = "MYSQL_8_0"
			r.Spec.PasswordValidationPolicy = &GcpCloudSqlPasswordValidationPolicy{
				EnablePasswordPolicy:   true,
				PasswordChangeInterval: "3600s",
			}
			expectInvalid(r, "applies to POSTGRESQL instances only")
		})

		ginkgo.It("accepts a full policy on PostgreSQL", func() {
			r := minimalPostgres()
			r.Spec.PasswordValidationPolicy = &GcpCloudSqlPasswordValidationPolicy{
				EnablePasswordPolicy:      true,
				MinLength:                 intPtr(12),
				Complexity:                "COMPLEXITY_DEFAULT",
				ReuseInterval:             intPtr(5),
				DisallowUsernameSubstring: true,
				PasswordChangeInterval:    "3600s",
			}
			expectValid(r)
		})

		ginkgo.It("rejects a malformed change interval", func() {
			r := minimalPostgres()
			r.Spec.PasswordValidationPolicy = &GcpCloudSqlPasswordValidationPolicy{
				EnablePasswordPolicy:   true,
				PasswordChangeInterval: "1h",
			}
			expectInvalid(r, "seconds duration")
		})
	})

	ginkgo.Context("replicas", func() {
		ginkgo.It("rejects replica_configuration without a primary", func() {
			r := minimalPostgres()
			r.Spec.ReplicaConfiguration = &GcpCloudSqlReplicaConfiguration{FailoverTarget: true}
			expectInvalid(r, "set master_instance_name")
		})

		ginkgo.It("accepts a replica referencing its primary", func() {
			r := minimalPostgres()
			r.Spec.MasterInstanceName = fromRef("GcpCloudSql", "orders-db-primary", "status.outputs.instance_name")
			r.Spec.ReplicaConfiguration = &GcpCloudSqlReplicaConfiguration{FailoverTarget: false}
			expectValid(r)
		})

		ginkgo.It("accepts a literal primary name", func() {
			r := minimalPostgres()
			r.Spec.MasterInstanceName = litRef("orders-db-primary")
			expectValid(r)
		})

		ginkgo.It("rejects a non-gs dump file path", func() {
			r := minimalPostgres()
			r.Spec.MasterInstanceName = litRef("orders-db-primary")
			r.Spec.ReplicaConfiguration = &GcpCloudSqlReplicaConfiguration{DumpFilePath: "/tmp/dump.sql"}
			expectInvalid(r, "gs:// URI")
		})
	})

	ginkgo.Context("placement", func() {
		ginkgo.It("rejects a secondary zone on a ZONAL instance", func() {
			r := minimalPostgres()
			r.Spec.LocationPreference = &GcpCloudSqlLocationPreference{
				Zone:          "us-central1-a",
				SecondaryZone: "us-central1-b",
			}
			expectInvalid(r, "secondary_zone applies only")
		})

		ginkgo.It("accepts a secondary zone on a REGIONAL instance", func() {
			r := minimalPostgres()
			r.Spec.AvailabilityType = ptr("REGIONAL")
			r.Spec.Backup = &GcpCloudSqlBackup{Enabled: true}
			r.Spec.LocationPreference = &GcpCloudSqlLocationPreference{
				Zone:          "us-central1-a",
				SecondaryZone: "us-central1-b",
			}
			expectValid(r)
		})
	})

	ginkgo.Context("misc coherence", func() {
		ginkgo.It("rejects an unknown activation_policy", func() {
			r := minimalPostgres()
			r.Spec.ActivationPolicy = ptr("SOMETIMES")
			expectInvalid(r, "activation_policy")
		})

		ginkgo.It("rejects an unknown connector_enforcement", func() {
			r := minimalPostgres()
			r.Spec.ConnectorEnforcement = "MAYBE"
			expectInvalid(r, "connector_enforcement")
		})

		ginkgo.It("accepts a CMEK key reference", func() {
			r := minimalPostgres()
			r.Spec.EncryptionKeyName = fromRef("GcpKmsKey", "sql-cmek", "status.outputs.key_id")
			expectValid(r)
		})

		ginkgo.It("accepts a production-grade full spec", func() {
			r := minimalPostgres()
			r.Spec.Edition = ptr("ENTERPRISE_PLUS")
			r.Spec.AvailabilityType = ptr("REGIONAL")
			r.Spec.Disk = &GcpCloudSqlDisk{
				Type:            ptr("PD_SSD"),
				SizeGb:          intPtr(100),
				AutoResize:      boolPtr(true),
				AutoResizeLimit: intPtr(500),
			}
			r.Spec.Network = &GcpCloudSqlNetwork{
				PrivateNetwork: fromRef("GcpVpcNetwork", "prod-vpc", "status.outputs.network_id"),
				SslMode:        "ENCRYPTED_ONLY",
			}
			r.Spec.Backup = &GcpCloudSqlBackup{
				Enabled:                     true,
				StartTime:                   "03:00",
				PointInTimeRecoveryEnabled:  true,
				TransactionLogRetentionDays: intPtr(14),
				RetainedBackups:             intPtr(30),
			}
			r.Spec.MaintenanceWindow = &GcpCloudSqlMaintenanceWindow{
				Day: 7, Hour: intPtr(3), UpdateTrack: "stable",
			}
			r.Spec.InsightsConfig = &GcpCloudSqlInsightsConfig{QueryInsightsEnabled: true}
			r.Spec.DataCacheEnabled = true
			r.Spec.DatabaseFlags = map[string]string{"max_connections": "500"}
			r.Spec.DeletionProtection = true
			r.Spec.DeletionProtectionEnabled = true
			r.Spec.RetainBackupsOnDelete = true
			r.Spec.RootPassword = "SuperSecret123!"
			expectValid(r)
		})
	})

	ginkgo.Context("deletion policy", func() {
		ginkgo.It("accepts every valid deletion_policy", func() {
			for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				r := minimalPostgres()
				r.Spec.DeletionPolicy = policy
				expectValid(r)
			}
		})

		ginkgo.It("rejects an unknown deletion_policy", func() {
			r := minimalPostgres()
			r.Spec.DeletionPolicy = "RETAIN"
			expectInvalid(r, "deletion_policy")
		})
	})

	ginkgo.Context("read pools", func() {
		ginkgo.It("accepts a read pool with node_count", func() {
			r := minimalPostgres()
			r.Spec.InstanceType = "READ_POOL_INSTANCE"
			r.Spec.NodeCount = intPtr(3)
			expectValid(r)
		})

		ginkgo.It("accepts a read pool with auto scaling", func() {
			r := minimalPostgres()
			r.Spec.InstanceType = "READ_POOL_INSTANCE"
			r.Spec.ReadPoolAutoScale = &GcpCloudSqlReadPoolAutoScale{
				Enabled:      true,
				MinNodeCount: intPtr(1),
				MaxNodeCount: intPtr(5),
				TargetMetrics: []*GcpCloudSqlReadPoolTargetMetric{
					{Metric: "cpu_utilization", TargetValue: 0.6},
				},
			}
			expectValid(r)
		})

		ginkgo.It("rejects node_count without READ_POOL_INSTANCE", func() {
			r := minimalPostgres()
			r.Spec.NodeCount = intPtr(3)
			expectInvalid(r, "READ_POOL_INSTANCE")
		})

		ginkgo.It("rejects read_pool_auto_scale without READ_POOL_INSTANCE", func() {
			r := minimalPostgres()
			r.Spec.ReadPoolAutoScale = &GcpCloudSqlReadPoolAutoScale{Enabled: true}
			expectInvalid(r, "READ_POOL_INSTANCE")
		})

		ginkgo.It("rejects inverted auto-scale bounds", func() {
			r := minimalPostgres()
			r.Spec.InstanceType = "READ_POOL_INSTANCE"
			r.Spec.ReadPoolAutoScale = &GcpCloudSqlReadPoolAutoScale{
				Enabled:      true,
				MinNodeCount: intPtr(5),
				MaxNodeCount: intPtr(2),
			}
			expectInvalid(r, "max_node_count")
		})

		ginkgo.It("rejects an unknown instance_type", func() {
			r := minimalPostgres()
			r.Spec.InstanceType = "POOL"
			expectInvalid(r, "instance_type")
		})
	})

	ginkgo.Context("restore sources", func() {
		ginkgo.It("accepts a clone from a source instance", func() {
			r := minimalPostgres()
			r.Spec.Clone = &GcpCloudSqlClone{
				SourceInstanceName: "orders-db-prod",
				PointInTime:        "2026-01-01T00:00:00Z",
				PreferredZone:      "us-central1-a",
			}
			expectValid(r)
		})

		ginkgo.It("rejects a clone without a source instance", func() {
			r := minimalPostgres()
			r.Spec.Clone = &GcpCloudSqlClone{PointInTime: "2026-01-01T00:00:00Z"}
			expectInvalid(r, "source_instance_name")
		})

		ginkgo.It("accepts a backup-run restore context", func() {
			r := minimalPostgres()
			r.Spec.RestoreBackupContext = &GcpCloudSqlRestoreBackupContext{
				BackupRunId: 1234567890,
				InstanceId:  "orders-db-prod",
			}
			expectValid(r)
		})

		ginkgo.It("accepts a Backup and DR point-in-time restore context", func() {
			r := minimalPostgres()
			r.Spec.PointInTimeRestoreContext = &GcpCloudSqlPointInTimeRestoreContext{
				Datasource:  "projects/p/locations/us-central1/backupVaults/v/dataSources/d",
				PointInTime: "2026-01-01T00:00:00Z",
			}
			expectValid(r)
		})

		ginkgo.It("rejects a PITR restore context without a datasource", func() {
			r := minimalPostgres()
			r.Spec.PointInTimeRestoreContext = &GcpCloudSqlPointInTimeRestoreContext{
				PointInTime: "2026-01-01T00:00:00Z",
			}
			expectInvalid(r, "datasource")
		})
	})

	ginkgo.Context("final backup", func() {
		ginkgo.It("accepts an enabled final backup with description", func() {
			r := minimalPostgres()
			r.Spec.FinalBackup = &GcpCloudSqlFinalBackup{
				Enabled:       true,
				RetentionDays: intPtr(30),
				Description:   "decommission snapshot",
			}
			expectValid(r)
		})

		ginkgo.It("rejects a description on a disabled final backup", func() {
			r := minimalPostgres()
			r.Spec.FinalBackup = &GcpCloudSqlFinalBackup{Description: "orphaned note"}
			expectInvalid(r, "final_backup")
		})
	})

	ginkgo.Context("Active Directory and Entra ID", func() {
		sqlServer := func() *GcpCloudSql {
			r := minimalPostgres()
			r.Spec.DatabaseEngine = "SQLSERVER"
			r.Spec.DatabaseVersion = "SQLSERVER_2022_STANDARD"
			r.Spec.RootPassword = "SqlServer123!"
			return r
		}

		ginkgo.It("accepts a managed AD join on SQL Server", func() {
			r := sqlServer()
			r.Spec.ActiveDirectory = &GcpCloudSqlActiveDirectoryConfig{Domain: "ad.example.com"}
			expectValid(r)
		})

		ginkgo.It("accepts a customer-managed AD join with bootstrap fields", func() {
			r := sqlServer()
			r.Spec.ActiveDirectory = &GcpCloudSqlActiveDirectoryConfig{
				Domain:                    "corp.example.com",
				Mode:                      "CUSTOMER_MANAGED_ACTIVE_DIRECTORY",
				DnsServers:                []string{"10.0.0.2", "10.0.0.3"},
				AdminCredentialSecretName: "projects/p/secrets/ad-admin",
				OrganizationalUnit:        "OU=SQL,DC=corp,DC=example,DC=com",
			}
			expectValid(r)
		})

		ginkgo.It("rejects AD on a non-SQL-Server engine", func() {
			r := minimalPostgres()
			r.Spec.ActiveDirectory = &GcpCloudSqlActiveDirectoryConfig{Domain: "ad.example.com"}
			expectInvalid(r, "SQL Server")
		})

		ginkgo.It("rejects customer-managed bootstrap fields in managed mode", func() {
			r := sqlServer()
			r.Spec.ActiveDirectory = &GcpCloudSqlActiveDirectoryConfig{
				Domain:     "ad.example.com",
				DnsServers: []string{"10.0.0.2"},
			}
			expectInvalid(r, "CUSTOMER_MANAGED_ACTIVE_DIRECTORY")
		})

		ginkgo.It("accepts a paired Entra ID config on SQL Server", func() {
			r := sqlServer()
			r.Spec.EntraId = &GcpCloudSqlEntraIdConfig{
				ApplicationId: "11111111-1111-1111-1111-111111111111",
				TenantId:      "22222222-2222-2222-2222-222222222222",
			}
			expectValid(r)
		})

		ginkgo.It("rejects an Entra ID config missing its tenant", func() {
			r := sqlServer()
			r.Spec.EntraId = &GcpCloudSqlEntraIdConfig{ApplicationId: "1111"}
			expectInvalid(r, "tenant_id")
		})

		ginkgo.It("rejects Entra ID on a non-SQL-Server engine", func() {
			r := minimalPostgres()
			r.Spec.EntraId = &GcpCloudSqlEntraIdConfig{
				ApplicationId: "1111",
				TenantId:      "2222",
			}
			expectInvalid(r, "SQL Server")
		})
	})

	ginkgo.Context("disk provisioned performance", func() {
		ginkgo.It("accepts IOPS and throughput on a hyperdisk", func() {
			r := minimalPostgres()
			r.Spec.Disk = &GcpCloudSqlDisk{
				Type:                  ptr("HYPERDISK_BALANCED"),
				SizeGb:                intPtr(100),
				ProvisionedIops:       int64Ptr(3000),
				ProvisionedThroughput: int64Ptr(290),
			}
			expectValid(r)
		})

		ginkgo.It("rejects provisioned IOPS on a PD_SSD disk", func() {
			r := minimalPostgres()
			r.Spec.Disk = &GcpCloudSqlDisk{
				Type:            ptr("PD_SSD"),
				SizeGb:          intPtr(100),
				ProvisionedIops: int64Ptr(3000),
			}
			expectInvalid(r, "HYPERDISK_BALANCED")
		})
	})

	ginkgo.Context("network additions", func() {
		ginkgo.It("accepts automatic certificate rotation with a CAS CA", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled:                   true,
				ServerCaMode:                  "GOOGLE_MANAGED_CAS_CA",
				ServerCertificateRotationMode: "AUTOMATIC_ROTATION_DURING_MAINTENANCE",
			}
			expectValid(r)
		})

		ginkgo.It("rejects automatic certificate rotation without a CAS CA", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled:                   true,
				ServerCertificateRotationMode: "AUTOMATIC_ROTATION_DURING_MAINTENANCE",
			}
			expectInvalid(r, "server_ca_mode")
		})

		ginkgo.It("rejects PSC DNS toggles without psc.enabled", func() {
			r := minimalPostgres()
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled: true,
				Psc:         &GcpCloudSqlPscConfig{AutoDnsEnabled: true},
			}
			expectInvalid(r, "psc.enabled")
		})
	})

	ginkgo.Context("backup retention unit", func() {
		ginkgo.It("accepts COUNT", func() {
			r := minimalPostgres()
			r.Spec.Backup = &GcpCloudSqlBackup{
				Enabled:         true,
				RetainedBackups: intPtr(14),
				RetentionUnit:   "COUNT",
			}
			expectValid(r)
		})

		ginkgo.It("rejects an unknown retention unit", func() {
			r := minimalPostgres()
			r.Spec.Backup = &GcpCloudSqlBackup{
				Enabled:       true,
				RetentionUnit: "DAYS",
			}
			expectInvalid(r, "retention_unit")
		})
	})

	ginkgo.Context("data API access", func() {
		ginkgo.It("accepts the two documented postures", func() {
			for _, v := range []string{"ALLOW_DATA_API", "DISALLOW_DATA_API"} {
				r := minimalPostgres()
				r.Spec.DataApiAccess = v
				expectValid(r)
			}
		})

		ginkgo.It("rejects an unknown posture", func() {
			r := minimalPostgres()
			r.Spec.DataApiAccess = "OPEN"
			expectInvalid(r, "data_api_access")
		})
	})

	ginkgo.Context("replica lag recreation threshold", func() {
		ginkgo.It("accepts the documented bounds", func() {
			for _, v := range []int32{300, 31536000} {
				r := minimalPostgres()
				r.Spec.ReplicationLagMaxSeconds = intPtr(v)
				expectValid(r)
			}
		})

		ginkgo.It("rejects a lag below five minutes", func() {
			r := minimalPostgres()
			r.Spec.ReplicationLagMaxSeconds = intPtr(299)
			expectInvalid(r, "replication_lag_max_seconds")
		})

		ginkgo.It("rejects a lag above one year", func() {
			r := minimalPostgres()
			r.Spec.ReplicationLagMaxSeconds = intPtr(31536001)
			expectInvalid(r, "replication_lag_max_seconds")
		})
	})

	ginkgo.Context("input-only opt-ins and the network-architecture switch", func() {
		ginkgo.It("accepts the opt-ins together", func() {
			r := minimalPostgres()
			r.Spec.SwitchTransactionLogsToCloudStorageEnabled = true
			r.Spec.IncludeReplicasForMajorVersionUpgrade = true
			enforce := true
			r.Spec.EnforceNewSqlNetworkArchitecture = &enforce
			expectValid(r)
		})
	})

	ginkgo.Context("PSC auto-connection policy", func() {
		ginkgo.It("accepts the policy switch with PSC enabled", func() {
			r := minimalPostgres()
			policy := true
			r.Spec.Network = &GcpCloudSqlNetwork{
				Psc: &GcpCloudSqlPscConfig{Enabled: true, AutoConnectionPolicyEnabled: &policy},
			}
			expectValid(r)
		})

		ginkgo.It("rejects the policy switch without psc.enabled", func() {
			r := minimalPostgres()
			policy := false
			r.Spec.Network = &GcpCloudSqlNetwork{
				Ipv4Enabled: true,
				Psc:         &GcpCloudSqlPscConfig{AutoConnectionPolicyEnabled: &policy},
			}
			expectInvalid(r, "psc.enabled")
		})
	})
})

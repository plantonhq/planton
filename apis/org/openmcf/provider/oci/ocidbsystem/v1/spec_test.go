package ocidbsystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciDbSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciDbSystemSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidDbSystem() *OciDbSystem {
	return &OciDbSystem{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciDbSystem",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-db-system",
		},
		Spec: &OciDbSystemSpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			AvailabilityDomain: "Uocm:PHX-AD-1",
			Shape:              "VM.Standard2.4",
			SubnetId:           newStringValueOrRef("ocid1.subnet.oc1.phx.example"),
			SshPublicKeys:      []string{"ssh-rsa AAAAB3...example"},
			Hostname:           "testdbhost",
			DbHome: &OciDbSystemSpec_DbHome{
				DbVersion: "19.0.0.0",
				Database: &OciDbSystemSpec_Database{
					AdminPassword: "WelcomePass#123",
					DbName:        "testdb",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("OciDbSystemSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_database_db_system", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidDbSystem()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name set", func() {
				input := minimalValidDbSystem()
				input.Spec.DisplayName = "Production Oracle DB"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with standard_edition", func() {
				input := minimalValidDbSystem()
				input.Spec.DatabaseEdition = OciDbSystemSpec_standard_edition
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with enterprise_edition", func() {
				input := minimalValidDbSystem()
				input.Spec.DatabaseEdition = OciDbSystemSpec_enterprise_edition
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with enterprise_edition_high_performance", func() {
				input := minimalValidDbSystem()
				input.Spec.DatabaseEdition = OciDbSystemSpec_enterprise_edition_high_performance
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with enterprise_edition_extreme_performance", func() {
				input := minimalValidDbSystem()
				input.Spec.DatabaseEdition = OciDbSystemSpec_enterprise_edition_extreme_performance
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with bring_your_own_license", func() {
				input := minimalValidDbSystem()
				input.Spec.LicenseModel = OciDbSystemSpec_bring_your_own_license
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with license_included", func() {
				input := minimalValidDbSystem()
				input.Spec.LicenseModel = OciDbSystemSpec_license_included
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with cpu_core_count and storage", func() {
				input := minimalValidDbSystem()
				input.Spec.CpuCoreCount = 4
				input.Spec.DataStorageSizeInGb = 512
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with disk redundancy normal", func() {
				input := minimalValidDbSystem()
				input.Spec.DiskRedundancy = OciDbSystemSpec_normal
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with disk redundancy high", func() {
				input := minimalValidDbSystem()
				input.Spec.DiskRedundancy = OciDbSystemSpec_high
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with 2-node RAC config", func() {
				input := minimalValidDbSystem()
				input.Spec.NodeCount = 2
				input.Spec.ClusterName = "raccluster"
				input.Spec.FaultDomains = []string{"FAULT-DOMAIN-1", "FAULT-DOMAIN-2"}
				input.Spec.DatabaseEdition = OciDbSystemSpec_enterprise_edition_extreme_performance
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with nsg_ids", func() {
				input := minimalValidDbSystem()
				input.Spec.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.phx.example1"),
					newStringValueOrRef("ocid1.nsg.oc1.phx.example2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backup subnet and NSGs", func() {
				input := minimalValidDbSystem()
				input.Spec.BackupSubnetId = newStringValueOrRef("ocid1.subnet.oc1.phx.backup")
				input.Spec.BackupNetworkNsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.phx.backup1"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with KMS key", func() {
				input := minimalValidDbSystem()
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1.phx.example")
				input.Spec.KmsKeyVersionId = "ocid1.keyversion.oc1.phx.example"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with time_zone set", func() {
				input := minimalValidDbSystem()
				input.Spec.TimeZone = "US/Pacific"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with sparse_diskgroup set", func() {
				input := minimalValidDbSystem()
				sparse := true
				input.Spec.SparseDiskgroup = &sparse
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with storage_volume_performance_mode", func() {
				input := minimalValidDbSystem()
				input.Spec.StorageVolumePerformanceMode = OciDbSystemSpec_high_performance
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with data_collection_options", func() {
				input := minimalValidDbSystem()
				diag := true
				health := true
				incident := false
				input.Spec.DataCollectionOptions = &OciDbSystemSpec_DataCollectionOptions{
					IsDiagnosticsEventsEnabled: &diag,
					IsHealthMonitoringEnabled:  &health,
					IsIncidentLogsEnabled:      &incident,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with db_system_options ASM", func() {
				input := minimalValidDbSystem()
				input.Spec.DbSystemOptions = &OciDbSystemSpec_DbSystemOptions{
					StorageManagement: OciDbSystemSpec_asm,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with db_system_options LVM", func() {
				input := minimalValidDbSystem()
				input.Spec.DbSystemOptions = &OciDbSystemSpec_DbSystemOptions{
					StorageManagement: OciDbSystemSpec_lvm,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maintenance_window no_preference", func() {
				input := minimalValidDbSystem()
				input.Spec.MaintenanceWindowDetails = &OciDbSystemSpec_MaintenanceWindowDetails{
					Preference: OciDbSystemSpec_no_preference,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maintenance_window custom_preference", func() {
				input := minimalValidDbSystem()
				customTimeout := true
				monthlyPatching := false
				input.Spec.MaintenanceWindowDetails = &OciDbSystemSpec_MaintenanceWindowDetails{
					Preference:                     OciDbSystemSpec_custom_preference,
					PatchingMode:                   OciDbSystemSpec_rolling,
					LeadTimeInWeeks:                1,
					Months:                         []string{"JANUARY", "APRIL", "JULY", "OCTOBER"},
					WeeksOfMonth:                   []int32{1},
					DaysOfWeek:                     []string{"MONDAY"},
					HoursOfDay:                     []int32{4},
					CustomActionTimeoutInMins:       120,
					IsCustomActionTimeoutEnabled:   &customTimeout,
					IsMonthlyPatchingEnabled:       &monthlyPatching,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with database_software_image_id instead of db_version", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.DbVersion = ""
				input.Spec.DbHome.DatabaseSoftwareImageId = newStringValueOrRef("ocid1.databasesoftwareimage.oc1.phx.example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with db_home display_name", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.DisplayName = "dbhome19c"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with database character sets", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.Database.CharacterSet = "AL32UTF8"
				input.Spec.DbHome.Database.NcharacterSet = "AL16UTF16"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with database pdb_name", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.Database.PdbName = "mypdb"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with database backup config", func() {
				input := minimalValidDbSystem()
				autoBackup := true
				input.Spec.DbHome.Database.DbBackupConfig = &OciDbSystemSpec_DbBackupConfig{
					AutoBackupEnabled:    &autoBackup,
					AutoBackupWindow:     "SLOT_TWO",
					RecoveryWindowInDays: 30,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with database TDE encryption", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.Database.KmsKeyId = newStringValueOrRef("ocid1.key.oc1.phx.tdekey")
				input.Spec.DbHome.Database.KmsKeyVersionId = "ocid1.keyversion.oc1.phx.tde1"
				input.Spec.DbHome.Database.VaultId = newStringValueOrRef("ocid1.vault.oc1.phx.tdevault")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidDbSystem()
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
				input := minimalValidDbSystem()
				input.Spec.SubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-private-subnet",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				input := minimalValidDbSystem()
				sparse := false
				diag := true
				health := true
				incident := true
				autoBackup := true
				customTimeout := true
				monthlyPatching := false
				input.Spec.DisplayName = "Full Config DB"
				input.Spec.CpuCoreCount = 8
				input.Spec.DatabaseEdition = OciDbSystemSpec_enterprise_edition
				input.Spec.LicenseModel = OciDbSystemSpec_bring_your_own_license
				input.Spec.DataStorageSizeInGb = 1024
				input.Spec.DataStoragePercentage = 80
				input.Spec.DiskRedundancy = OciDbSystemSpec_normal
				input.Spec.NodeCount = 1
				input.Spec.Domain = "db.example.com"
				input.Spec.TimeZone = "UTC"
				input.Spec.SparseDiskgroup = &sparse
				input.Spec.StorageVolumePerformanceMode = OciDbSystemSpec_balanced
				input.Spec.PrivateIp = "10.0.1.100"
				input.Spec.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.phx.example"),
				}
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1.phx.example")
				input.Spec.DataCollectionOptions = &OciDbSystemSpec_DataCollectionOptions{
					IsDiagnosticsEventsEnabled: &diag,
					IsHealthMonitoringEnabled:  &health,
					IsIncidentLogsEnabled:      &incident,
				}
				input.Spec.DbSystemOptions = &OciDbSystemSpec_DbSystemOptions{
					StorageManagement: OciDbSystemSpec_asm,
				}
				input.Spec.MaintenanceWindowDetails = &OciDbSystemSpec_MaintenanceWindowDetails{
					Preference:                   OciDbSystemSpec_custom_preference,
					PatchingMode:                 OciDbSystemSpec_rolling,
					LeadTimeInWeeks:              2,
					Months:                       []string{"MARCH"},
					WeeksOfMonth:                 []int32{2},
					DaysOfWeek:                   []string{"WEDNESDAY"},
					HoursOfDay:                   []int32{2},
					CustomActionTimeoutInMins:    60,
					IsCustomActionTimeoutEnabled: &customTimeout,
					IsMonthlyPatchingEnabled:     &monthlyPatching,
				}
				input.Spec.DbHome.DisplayName = "dbhome19"
				input.Spec.DbHome.Database.CharacterSet = "AL32UTF8"
				input.Spec.DbHome.Database.NcharacterSet = "AL16UTF16"
				input.Spec.DbHome.Database.PdbName = "mypdb"
				input.Spec.DbHome.Database.DbDomain = "db.example.com"
				input.Spec.DbHome.Database.KmsKeyId = newStringValueOrRef("ocid1.key.oc1.phx.tde")
				input.Spec.DbHome.Database.VaultId = newStringValueOrRef("ocid1.vault.oc1.phx.tde")
				input.Spec.DbHome.Database.DbBackupConfig = &OciDbSystemSpec_DbBackupConfig{
					AutoBackupEnabled:    &autoBackup,
					AutoBackupWindow:     "SLOT_FOUR",
					RecoveryWindowInDays: 45,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_database_db_system", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidDbSystem()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidDbSystem()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidDbSystem()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciDbSystem{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciDbSystem",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-db"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidDbSystem()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when availability_domain is empty", func() {
				input := minimalValidDbSystem()
				input.Spec.AvailabilityDomain = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape is empty", func() {
				input := minimalValidDbSystem()
				input.Spec.Shape = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := minimalValidDbSystem()
				input.Spec.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ssh_public_keys is empty", func() {
				input := minimalValidDbSystem()
				input.Spec.SshPublicKeys = []string{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when hostname is empty", func() {
				input := minimalValidDbSystem()
				input.Spec.Hostname = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_home is missing", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_home.database is missing", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.Database = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when admin_password is too short", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.Database.AdminPassword = "x"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_name is empty", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.Database.DbName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_name starts with a number", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.Database.DbName = "1testdb"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_name contains special characters", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.Database.DbName = "test-db"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_name exceeds 30 characters", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.Database.DbName = "abcdefghijklmnopqrstuvwxyz12345"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both db_version and database_software_image_id are set", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.DbVersion = "19.0.0.0"
				input.Spec.DbHome.DatabaseSoftwareImageId = newStringValueOrRef("ocid1.databasesoftwareimage.oc1.phx.example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither db_version nor database_software_image_id is set", func() {
				input := minimalValidDbSystem()
				input.Spec.DbHome.DbVersion = ""
				input.Spec.DbHome.DatabaseSoftwareImageId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

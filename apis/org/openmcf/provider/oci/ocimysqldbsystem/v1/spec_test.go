package ocimysqldbsystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestOciMysqlDbSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciMysqlDbSystemSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidMysql() *OciMysqlDbSystem {
	return &OciMysqlDbSystem{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciMysqlDbSystem",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-mysql",
		},
		Spec: &OciMysqlDbSystemSpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			AvailabilityDomain: "Uocm:PHX-AD-1",
			ShapeName:          "MySQL.VM.Standard.E4.4.64GB",
			SubnetId:           newStringValueOrRef("ocid1.subnet.oc1.phx.example"),
		},
	}
}

var _ = ginkgo.Describe("OciMysqlDbSystemSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_mysql_db_system", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidMysql()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name and description", func() {
				input := minimalValidMysql()
				input.Spec.DisplayName = "Production MySQL"
				input.Spec.Description = "Primary production MySQL database"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with admin credentials", func() {
				input := minimalValidMysql()
				input.Spec.AdminUsername = "admin"
				input.Spec.AdminPassword = "MySecure@Pass1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HA enabled", func() {
				input := minimalValidMysql()
				ha := true
				input.Spec.IsHighlyAvailable = &ha
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with mysql_version", func() {
				input := minimalValidMysql()
				input.Spec.MysqlVersion = "8.0.36"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with configuration_id", func() {
				input := minimalValidMysql()
				input.Spec.ConfigurationId = newStringValueOrRef("ocid1.mysqlconfiguration.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with networking details", func() {
				input := minimalValidMysql()
				input.Spec.HostnameLabel = "mydb"
				input.Spec.IpAddress = "10.0.1.50"
				input.Spec.FaultDomain = "FAULT-DOMAIN-1"
				input.Spec.Port = 3306
				input.Spec.PortX = 33060
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with crash_recovery and database_management", func() {
				input := minimalValidMysql()
				input.Spec.CrashRecovery = "ENABLED"
				input.Spec.DatabaseManagement = "ENABLED"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with NSG IDs", func() {
				input := minimalValidMysql()
				input.Spec.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.phx.1"),
					newStringValueOrRef("ocid1.nsg.oc1.phx.2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with data storage", func() {
				input := minimalValidMysql()
				autoExpand := true
				input.Spec.DataStorage = &OciMysqlDbSystemSpec_DataStorage{
					DataStorageSizeInGb:        50,
					IsAutoExpandStorageEnabled: &autoExpand,
					MaxStorageSizeInGbs:        65536,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backup policy enabled", func() {
				input := minimalValidMysql()
				enabled := true
				pitrEnabled := true
				input.Spec.BackupPolicy = &OciMysqlDbSystemSpec_BackupPolicy{
					IsEnabled:       &enabled,
					RetentionInDays: 7,
					WindowStartTime: "03:00",
					PitrPolicy: &OciMysqlDbSystemSpec_PitrPolicy{
						IsEnabled: &pitrEnabled,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backup policy disabled", func() {
				input := minimalValidMysql()
				disabled := false
				input.Spec.BackupPolicy = &OciMysqlDbSystemSpec_BackupPolicy{
					IsEnabled: &disabled,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maintenance window", func() {
				input := minimalValidMysql()
				input.Spec.Maintenance = &OciMysqlDbSystemSpec_Maintenance{
					WindowStartTime:         "mon 10:00",
					MaintenanceScheduleType: OciMysqlDbSystemSpec_regular,
					VersionPreference:       OciMysqlDbSystemSpec_newest,
					VersionTrackPreference:  OciMysqlDbSystemSpec_long_term_support,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with early maintenance schedule", func() {
				input := minimalValidMysql()
				input.Spec.Maintenance = &OciMysqlDbSystemSpec_Maintenance{
					WindowStartTime:         "wed 02:00",
					MaintenanceScheduleType: OciMysqlDbSystemSpec_early,
					VersionPreference:       OciMysqlDbSystemSpec_oldest,
					VersionTrackPreference:  OciMysqlDbSystemSpec_innovation,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with deletion policy", func() {
				input := minimalValidMysql()
				deleteProtected := true
				input.Spec.DeletionPolicy = &OciMysqlDbSystemSpec_DeletionPolicy{
					AutomaticBackupRetention: "RETAIN",
					FinalBackup:              "REQUIRE_FINAL_BACKUP",
					IsDeleteProtected:        &deleteProtected,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with system encryption", func() {
				input := minimalValidMysql()
				input.Spec.EncryptData = &OciMysqlDbSystemSpec_EncryptData{
					KeyGenerationType: OciMysqlDbSystemSpec_system,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with BYOK encryption", func() {
				input := minimalValidMysql()
				input.Spec.EncryptData = &OciMysqlDbSystemSpec_EncryptData{
					KeyGenerationType: OciMysqlDbSystemSpec_byok,
					KeyId:             newStringValueOrRef("ocid1.key.oc1.phx.example"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with system secure connections", func() {
				input := minimalValidMysql()
				input.Spec.SecureConnections = &OciMysqlDbSystemSpec_SecureConnections{
					CertificateGenerationType: OciMysqlDbSystemSpec_system_cert,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with BYOC secure connections", func() {
				input := minimalValidMysql()
				input.Spec.SecureConnections = &OciMysqlDbSystemSpec_SecureConnections{
					CertificateGenerationType: OciMysqlDbSystemSpec_byoc,
					CertificateId:             newStringValueOrRef("ocid1.certificate.oc1.phx.example"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with customer contacts", func() {
				input := minimalValidMysql()
				input.Spec.CustomerContacts = []*OciMysqlDbSystemSpec_CustomerContact{
					{Email: "ops@example.com"},
					{Email: "dba@example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with read endpoint", func() {
				input := minimalValidMysql()
				readEnabled := true
				input.Spec.ReadEndpoint = &OciMysqlDbSystemSpec_ReadEndpoint{
					IsEnabled:                 &readEnabled,
					ReadEndpointHostnameLabel: "mydb-read",
					ReadEndpointIpAddress:     "10.0.1.51",
					ExcludeIps:                []string{"10.0.1.100"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with database console", func() {
				input := minimalValidMysql()
				input.Spec.DatabaseConsole = &OciMysqlDbSystemSpec_DatabaseConsole{
					Status: OciMysqlDbSystemSpec_enabled,
					Port:   8443,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with REST config", func() {
				input := minimalValidMysql()
				input.Spec.Rest = &OciMysqlDbSystemSpec_Rest{
					Configuration: "ENABLED",
					Port:          8444,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidMysql()
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

			ginkgo.It("should not return a validation error with subnet_id via valueFrom ref", func() {
				input := minimalValidMysql()
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

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidMysql()
				ha := true
				autoExpand := true
				backupEnabled := true
				pitrEnabled := true
				deleteProtected := false
				readEnabled := true
				input.Spec.DisplayName = "Production MySQL HeatWave"
				input.Spec.Description = "Primary production MySQL database"
				input.Spec.AdminUsername = "admin"
				input.Spec.AdminPassword = "MySecure@Pass1"
				input.Spec.MysqlVersion = "8.0.36"
				input.Spec.ConfigurationId = newStringValueOrRef("ocid1.mysqlconfiguration.oc1..example")
				input.Spec.IsHighlyAvailable = &ha
				input.Spec.HostnameLabel = "mydb"
				input.Spec.IpAddress = "10.0.1.50"
				input.Spec.FaultDomain = "FAULT-DOMAIN-1"
				input.Spec.Port = 3306
				input.Spec.PortX = 33060
				input.Spec.CrashRecovery = "ENABLED"
				input.Spec.DatabaseManagement = "ENABLED"
				input.Spec.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.phx.1"),
				}
				input.Spec.DataStorage = &OciMysqlDbSystemSpec_DataStorage{
					DataStorageSizeInGb:        100,
					IsAutoExpandStorageEnabled: &autoExpand,
					MaxStorageSizeInGbs:        65536,
				}
				input.Spec.BackupPolicy = &OciMysqlDbSystemSpec_BackupPolicy{
					IsEnabled:       &backupEnabled,
					RetentionInDays: 30,
					WindowStartTime: "03:00",
					PitrPolicy: &OciMysqlDbSystemSpec_PitrPolicy{
						IsEnabled: &pitrEnabled,
					},
				}
				input.Spec.Maintenance = &OciMysqlDbSystemSpec_Maintenance{
					WindowStartTime:         "sun 02:00",
					MaintenanceScheduleType: OciMysqlDbSystemSpec_regular,
					VersionPreference:       OciMysqlDbSystemSpec_newest,
					VersionTrackPreference:  OciMysqlDbSystemSpec_long_term_support,
				}
				input.Spec.DeletionPolicy = &OciMysqlDbSystemSpec_DeletionPolicy{
					AutomaticBackupRetention: "RETAIN",
					FinalBackup:              "SKIP_FINAL_BACKUP",
					IsDeleteProtected:        &deleteProtected,
				}
				input.Spec.EncryptData = &OciMysqlDbSystemSpec_EncryptData{
					KeyGenerationType: OciMysqlDbSystemSpec_system,
				}
				input.Spec.SecureConnections = &OciMysqlDbSystemSpec_SecureConnections{
					CertificateGenerationType: OciMysqlDbSystemSpec_system_cert,
				}
				input.Spec.CustomerContacts = []*OciMysqlDbSystemSpec_CustomerContact{
					{Email: "ops@example.com"},
				}
				input.Spec.ReadEndpoint = &OciMysqlDbSystemSpec_ReadEndpoint{
					IsEnabled:                 &readEnabled,
					ReadEndpointHostnameLabel: "mydb-read",
				}
				input.Spec.DatabaseConsole = &OciMysqlDbSystemSpec_DatabaseConsole{
					Status: OciMysqlDbSystemSpec_enabled,
					Port:   8443,
				}
				input.Spec.Rest = &OciMysqlDbSystemSpec_Rest{
					Configuration: "ENABLED",
					Port:          8444,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_mysql_db_system", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidMysql()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidMysql()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidMysql()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciMysqlDbSystem{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciMysqlDbSystem",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-mysql"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidMysql()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when availability_domain is empty", func() {
				input := minimalValidMysql()
				input.Spec.AvailabilityDomain = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape_name is empty", func() {
				input := minimalValidMysql()
				input.Spec.ShapeName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := minimalValidMysql()
				input.Spec.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when BYOK encryption has no key_id", func() {
				input := minimalValidMysql()
				input.Spec.EncryptData = &OciMysqlDbSystemSpec_EncryptData{
					KeyGenerationType: OciMysqlDbSystemSpec_byok,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when BYOC secure connections has no certificate_id", func() {
				input := minimalValidMysql()
				input.Spec.SecureConnections = &OciMysqlDbSystemSpec_SecureConnections{
					CertificateGenerationType: OciMysqlDbSystemSpec_byoc,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when customer contact email is empty", func() {
				input := minimalValidMysql()
				input.Spec.CustomerContacts = []*OciMysqlDbSystemSpec_CustomerContact{
					{Email: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when customer contacts exceed 10", func() {
				input := minimalValidMysql()
				contacts := make([]*OciMysqlDbSystemSpec_CustomerContact, 11)
				for i := 0; i < 11; i++ {
					contacts[i] = &OciMysqlDbSystemSpec_CustomerContact{Email: "user@example.com"}
				}
				input.Spec.CustomerContacts = contacts
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when maintenance window_start_time is empty", func() {
				input := minimalValidMysql()
				input.Spec.Maintenance = &OciMysqlDbSystemSpec_Maintenance{
					WindowStartTime: "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

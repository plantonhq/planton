package ocipostgresqldbsystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciPostgresqlDbSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciPostgresqlDbSystemSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidPostgresql() *OciPostgresqlDbSystem {
	return &OciPostgresqlDbSystem{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciPostgresqlDbSystem",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-pg",
		},
		Spec: &OciPostgresqlDbSystemSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			DbVersion:     "16",
			Shape:         "VM.Standard.E4.Flex",
			NetworkDetails: &OciPostgresqlDbSystemSpec_NetworkDetails{
				SubnetId: newStringValueOrRef("ocid1.subnet.oc1.phx.example"),
			},
			StorageDetails: &OciPostgresqlDbSystemSpec_StorageDetails{
				IsRegionallyDurable: true,
			},
		},
	}
}

var _ = ginkgo.Describe("OciPostgresqlDbSystemSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_psql_db_system", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidPostgresql()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name and description", func() {
				input := minimalValidPostgresql()
				input.Spec.DisplayName = "Production PostgreSQL"
				input.Spec.Description = "Primary production PostgreSQL database"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with flex shape fields", func() {
				input := minimalValidPostgresql()
				input.Spec.InstanceOcpuCount = 4
				input.Spec.InstanceMemorySizeInGbs = 64
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with instance_count", func() {
				input := minimalValidPostgresql()
				input.Spec.InstanceCount = 3
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with plain_text credentials", func() {
				input := minimalValidPostgresql()
				input.Spec.Credentials = &OciPostgresqlDbSystemSpec_Credentials{
					Username: "pgadmin",
					PasswordDetails: &OciPostgresqlDbSystemSpec_PasswordDetails{
						PasswordType: OciPostgresqlDbSystemSpec_plain_text,
						Password:     "MySecure@Pass1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with vault_secret credentials", func() {
				input := minimalValidPostgresql()
				input.Spec.Credentials = &OciPostgresqlDbSystemSpec_Credentials{
					Username: "pgadmin",
					PasswordDetails: &OciPostgresqlDbSystemSpec_PasswordDetails{
						PasswordType:  OciPostgresqlDbSystemSpec_vault_secret,
						SecretId:      newStringValueOrRef("ocid1.vaultsecret.oc1.phx.example"),
						SecretVersion: "1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with reader endpoint enabled", func() {
				input := minimalValidPostgresql()
				readerEnabled := true
				input.Spec.NetworkDetails.IsReaderEndpointEnabled = &readerEnabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with NSG IDs", func() {
				input := minimalValidPostgresql()
				input.Spec.NetworkDetails.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.phx.1"),
					newStringValueOrRef("ocid1.nsg.oc1.phx.2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with primary endpoint private IP", func() {
				input := minimalValidPostgresql()
				input.Spec.NetworkDetails.PrimaryDbEndpointPrivateIp = "10.0.1.50"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with AD-local storage", func() {
				input := minimalValidPostgresql()
				input.Spec.StorageDetails.IsRegionallyDurable = false
				input.Spec.StorageDetails.AvailabilityDomain = "Uocm:PHX-AD-1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with IOPS", func() {
				input := minimalValidPostgresql()
				input.Spec.StorageDetails.Iops = 75000
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with daily backup policy", func() {
				input := minimalValidPostgresql()
				input.Spec.ManagementPolicy = &OciPostgresqlDbSystemSpec_ManagementPolicy{
					BackupPolicy: &OciPostgresqlDbSystemSpec_BackupPolicy{
						Kind:          OciPostgresqlDbSystemSpec_daily,
						BackupStart:   "02:00",
						RetentionDays: 7,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with weekly backup policy", func() {
				input := minimalValidPostgresql()
				input.Spec.ManagementPolicy = &OciPostgresqlDbSystemSpec_ManagementPolicy{
					BackupPolicy: &OciPostgresqlDbSystemSpec_BackupPolicy{
						Kind:          OciPostgresqlDbSystemSpec_weekly,
						BackupStart:   "03:00",
						RetentionDays: 14,
						DaysOfTheWeek: []string{"MONDAY", "THURSDAY"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with monthly backup policy", func() {
				input := minimalValidPostgresql()
				input.Spec.ManagementPolicy = &OciPostgresqlDbSystemSpec_ManagementPolicy{
					BackupPolicy: &OciPostgresqlDbSystemSpec_BackupPolicy{
						Kind:           OciPostgresqlDbSystemSpec_monthly,
						BackupStart:    "04:00",
						RetentionDays:  30,
						DaysOfTheMonth: []int32{1, 15},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backup kind none", func() {
				input := minimalValidPostgresql()
				input.Spec.ManagementPolicy = &OciPostgresqlDbSystemSpec_ManagementPolicy{
					BackupPolicy: &OciPostgresqlDbSystemSpec_BackupPolicy{
						Kind: OciPostgresqlDbSystemSpec_none,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maintenance_window_start", func() {
				input := minimalValidPostgresql()
				input.Spec.ManagementPolicy = &OciPostgresqlDbSystemSpec_ManagementPolicy{
					MaintenanceWindowStart: "tue 02:00:00",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with config_id", func() {
				input := minimalValidPostgresql()
				input.Spec.ConfigId = newStringValueOrRef("ocid1.psqlconfiguration.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with instances_details", func() {
				input := minimalValidPostgresql()
				input.Spec.InstanceCount = 2
				input.Spec.InstancesDetails = []*OciPostgresqlDbSystemSpec_InstanceDetails{
					{DisplayName: "primary", Description: "Primary node", PrivateIp: "10.0.1.10"},
					{DisplayName: "replica-1", Description: "Read replica", PrivateIp: "10.0.1.11"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidPostgresql()
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
				input := minimalValidPostgresql()
				input.Spec.NetworkDetails.SubnetId = &foreignkeyv1.StringValueOrRef{
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
				input := minimalValidPostgresql()
				readerEnabled := true
				input.Spec.DisplayName = "Production PostgreSQL 16"
				input.Spec.Description = "Primary production PostgreSQL database"
				input.Spec.InstanceOcpuCount = 8
				input.Spec.InstanceMemorySizeInGbs = 128
				input.Spec.InstanceCount = 3
				input.Spec.NetworkDetails.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.phx.1"),
				}
				input.Spec.NetworkDetails.IsReaderEndpointEnabled = &readerEnabled
				input.Spec.NetworkDetails.PrimaryDbEndpointPrivateIp = "10.0.1.50"
				input.Spec.StorageDetails.Iops = 75000
				input.Spec.Credentials = &OciPostgresqlDbSystemSpec_Credentials{
					Username: "pgadmin",
					PasswordDetails: &OciPostgresqlDbSystemSpec_PasswordDetails{
						PasswordType: OciPostgresqlDbSystemSpec_plain_text,
						Password:     "MySecure@Pass1",
					},
				}
				input.Spec.ManagementPolicy = &OciPostgresqlDbSystemSpec_ManagementPolicy{
					BackupPolicy: &OciPostgresqlDbSystemSpec_BackupPolicy{
						Kind:          OciPostgresqlDbSystemSpec_daily,
						BackupStart:   "02:00",
						RetentionDays: 14,
					},
					MaintenanceWindowStart: "sun 03:00:00",
				}
				input.Spec.ConfigId = newStringValueOrRef("ocid1.psqlconfiguration.oc1..example")
				input.Spec.InstancesDetails = []*OciPostgresqlDbSystemSpec_InstanceDetails{
					{DisplayName: "primary", PrivateIp: "10.0.1.50"},
					{DisplayName: "replica-1", PrivateIp: "10.0.1.51"},
					{DisplayName: "replica-2", PrivateIp: "10.0.1.52"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_psql_db_system", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidPostgresql()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidPostgresql()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidPostgresql()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciPostgresqlDbSystem{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciPostgresqlDbSystem",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-pg"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidPostgresql()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_version is empty", func() {
				input := minimalValidPostgresql()
				input.Spec.DbVersion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape is empty", func() {
				input := minimalValidPostgresql()
				input.Spec.Shape = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network_details is missing", func() {
				input := minimalValidPostgresql()
				input.Spec.NetworkDetails = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network_details.subnet_id is missing", func() {
				input := minimalValidPostgresql()
				input.Spec.NetworkDetails.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage_details is missing", func() {
				input := minimalValidPostgresql()
				input.Spec.StorageDetails = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for plain_text credentials without password", func() {
				input := minimalValidPostgresql()
				input.Spec.Credentials = &OciPostgresqlDbSystemSpec_Credentials{
					Username: "pgadmin",
					PasswordDetails: &OciPostgresqlDbSystemSpec_PasswordDetails{
						PasswordType: OciPostgresqlDbSystemSpec_plain_text,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for vault_secret credentials without secret_id", func() {
				input := minimalValidPostgresql()
				input.Spec.Credentials = &OciPostgresqlDbSystemSpec_Credentials{
					Username: "pgadmin",
					PasswordDetails: &OciPostgresqlDbSystemSpec_PasswordDetails{
						PasswordType: OciPostgresqlDbSystemSpec_vault_secret,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when credentials username is empty", func() {
				input := minimalValidPostgresql()
				input.Spec.Credentials = &OciPostgresqlDbSystemSpec_Credentials{
					Username: "",
					PasswordDetails: &OciPostgresqlDbSystemSpec_PasswordDetails{
						PasswordType: OciPostgresqlDbSystemSpec_plain_text,
						Password:     "MySecure@Pass1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

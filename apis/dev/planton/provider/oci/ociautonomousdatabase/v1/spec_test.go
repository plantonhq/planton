package ociautonomousdatabasev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciAutonomousDatabaseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciAutonomousDatabaseSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidAdb() *OciAutonomousDatabase {
	return &OciAutonomousDatabase{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciAutonomousDatabase",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-adb",
		},
		Spec: &OciAutonomousDatabaseSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			DbName:        "testdb01",
		},
	}
}

var _ = ginkgo.Describe("OciAutonomousDatabaseSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_autonomous_database", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidAdb()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name set", func() {
				input := minimalValidAdb()
				input.Spec.DisplayName = "My ATP Database"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with OLTP workload", func() {
				input := minimalValidAdb()
				input.Spec.DbWorkload = OciAutonomousDatabaseSpec_oltp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with DW workload", func() {
				input := minimalValidAdb()
				input.Spec.DbWorkload = OciAutonomousDatabaseSpec_dw
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with AJD workload", func() {
				input := minimalValidAdb()
				input.Spec.DbWorkload = OciAutonomousDatabaseSpec_ajd
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with APEX workload", func() {
				input := minimalValidAdb()
				input.Spec.DbWorkload = OciAutonomousDatabaseSpec_apex
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with LH workload", func() {
				input := minimalValidAdb()
				input.Spec.DbWorkload = OciAutonomousDatabaseSpec_lh
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ECPU compute model", func() {
				input := minimalValidAdb()
				computeCount := float32(4.0)
				input.Spec.ComputeModel = OciAutonomousDatabaseSpec_ecpu
				input.Spec.ComputeCount = &computeCount
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with OCPU compute model", func() {
				input := minimalValidAdb()
				computeCount := float32(2.0)
				input.Spec.ComputeModel = OciAutonomousDatabaseSpec_ocpu
				input.Spec.ComputeCount = &computeCount
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with storage in TBs", func() {
				input := minimalValidAdb()
				input.Spec.DataStorageSizeInTbs = 2
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with storage in GBs", func() {
				input := minimalValidAdb()
				input.Spec.DataStorageSizeInGb = 256
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with admin_password set", func() {
				input := minimalValidAdb()
				input.Spec.AdminPassword = "MySecure@Pass1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with secret_id set", func() {
				input := minimalValidAdb()
				input.Spec.SecretId = newStringValueOrRef("ocid1.vaultsecret.oc1..example")
				input.Spec.SecretVersionNumber = 3
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with private endpoint config", func() {
				input := minimalValidAdb()
				mtls := true
				input.Spec.SubnetId = newStringValueOrRef("ocid1.subnet.oc1.iad.example")
				input.Spec.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.iad.example"),
				}
				input.Spec.PrivateEndpointLabel = "myatp"
				input.Spec.PrivateEndpointIp = "10.0.1.50"
				input.Spec.IsMtlsConnectionRequired = &mtls
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with encryption config", func() {
				input := minimalValidAdb()
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1.iad.example")
				input.Spec.VaultId = newStringValueOrRef("ocid1.vault.oc1.iad.example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with free tier", func() {
				input := minimalValidAdb()
				freeTier := true
				input.Spec.IsFreeTier = &freeTier
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with dedicated infrastructure", func() {
				input := minimalValidAdb()
				dedicated := true
				input.Spec.IsDedicated = &dedicated
				input.Spec.AutonomousContainerDatabaseId = newStringValueOrRef("ocid1.autonomouscontainerdatabase.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with customer contacts", func() {
				input := minimalValidAdb()
				input.Spec.CustomerContacts = []*OciAutonomousDatabaseSpec_CustomerContact{
					{Email: "ops@example.com"},
					{Email: "dba@example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidAdb()
				computeCount := float32(4.0)
				autoScale := true
				storageAutoScale := false
				dataGuard := true
				devTier := false
				input.Spec.DisplayName = "Production ATP"
				input.Spec.DbWorkload = OciAutonomousDatabaseSpec_oltp
				input.Spec.DbVersion = "23ai"
				input.Spec.DatabaseEdition = OciAutonomousDatabaseSpec_enterprise_edition
				input.Spec.LicenseModel = OciAutonomousDatabaseSpec_bring_your_own_license
				input.Spec.CharacterSet = "AL32UTF8"
				input.Spec.NcharacterSet = "AL16UTF16"
				input.Spec.ComputeModel = OciAutonomousDatabaseSpec_ecpu
				input.Spec.ComputeCount = &computeCount
				input.Spec.DataStorageSizeInTbs = 2
				input.Spec.IsAutoScalingEnabled = &autoScale
				input.Spec.IsAutoScalingForStorageEnabled = &storageAutoScale
				input.Spec.AdminPassword = "MySecure@Pass1"
				input.Spec.SubnetId = newStringValueOrRef("ocid1.subnet.oc1.iad.example")
				input.Spec.WhitelistedIps = []string{"10.0.0.0/16"}
				input.Spec.BackupRetentionPeriodInDays = 60
				input.Spec.IsLocalDataGuardEnabled = &dataGuard
				input.Spec.AutonomousMaintenanceScheduleType = OciAutonomousDatabaseSpec_early
				input.Spec.IsDevTier = &devTier
				input.Spec.CustomerContacts = []*OciAutonomousDatabaseSpec_CustomerContact{
					{Email: "ops@example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidAdb()
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
				input := minimalValidAdb()
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

			ginkgo.It("should not return a validation error with 5 NSG IDs (max allowed)", func() {
				input := minimalValidAdb()
				input.Spec.SubnetId = newStringValueOrRef("ocid1.subnet.oc1.iad.example")
				input.Spec.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.iad.1"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.2"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.3"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.4"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.5"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with whitelisted IPs", func() {
				input := minimalValidAdb()
				acl := true
				input.Spec.WhitelistedIps = []string{"10.0.0.0/16", "192.168.1.0/24", "ocid1.vcn.oc1..example"}
				input.Spec.IsAccessControlEnabled = &acl
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_autonomous_database", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidAdb()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidAdb()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidAdb()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciAutonomousDatabase{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciAutonomousDatabase",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-adb"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidAdb()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_name is empty", func() {
				input := minimalValidAdb()
				input.Spec.DbName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_name starts with a number", func() {
				input := minimalValidAdb()
				input.Spec.DbName = "1testdb"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_name contains special characters", func() {
				input := minimalValidAdb()
				input.Spec.DbName = "test-db"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when db_name exceeds 30 characters", func() {
				input := minimalValidAdb()
				input.Spec.DbName = "abcdefghijklmnopqrstuvwxyz12345"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when nsg_ids exceeds 5", func() {
				input := minimalValidAdb()
				input.Spec.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1.iad.1"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.2"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.3"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.4"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.5"),
					newStringValueOrRef("ocid1.nsg.oc1.iad.6"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when customer contact email is empty", func() {
				input := minimalValidAdb()
				input.Spec.CustomerContacts = []*OciAutonomousDatabaseSpec_CustomerContact{
					{Email: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both storage sizes are set", func() {
				input := minimalValidAdb()
				input.Spec.DataStorageSizeInTbs = 2
				input.Spec.DataStorageSizeInGb = 512
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both admin_password and secret_id are set", func() {
				input := minimalValidAdb()
				input.Spec.AdminPassword = "MySecure@Pass1"
				input.Spec.SecretId = newStringValueOrRef("ocid1.vaultsecret.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

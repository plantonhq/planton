package azuremysqlflexibleserverv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureMysqlFlexibleServerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMysqlFlexibleServerSpec Validation Tests")
}

// helper to create a minimal valid spec (public access mode)
func minimalPublicSpec() *AzureMysqlFlexibleServer {
	return &AzureMysqlFlexibleServer{
		ApiVersion: "azure.planton.dev/v1",
		Kind:       "AzureMysqlFlexibleServer",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-mysql",
		},
		Spec: &AzureMysqlFlexibleServerSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name:               "test-mysql-server",
			AdministratorLogin: "mysqladmin",
			AdministratorPassword: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "P@ssw0rd1234!",
				},
			},
			SkuName:       "GP_Standard_D2ds_v4",
			StorageSizeGb: 20,
		},
	}
}

var _ = ginkgo.Describe("AzureMysqlFlexibleServerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_mysql_flexible_server", func() {

			ginkgo.It("should not return a validation error for a minimal public server", func() {
				input := minimalPublicSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with VNet integration", func() {
				input := minimalPublicSpec()
				input.Spec.DelegatedSubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/mysql-subnet",
					},
				}
				input.Spec.PrivateDnsZoneId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/privateDnsZones/privatelink.mysql.database.azure.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with databases", func() {
				input := minimalPublicSpec()
				input.Spec.Databases = []*AzureMysqlDatabase{
					{Name: "myapp"},
					{Name: "analytics"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with databases and custom charset", func() {
				charset := "latin1"
				collation := "latin1_swedish_ci"
				input := minimalPublicSpec()
				input.Spec.Databases = []*AzureMysqlDatabase{
					{Name: "legacy", Charset: &charset, Collation: &collation},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with firewall rules", func() {
				input := minimalPublicSpec()
				input.Spec.FirewallRules = []*AzureMysqlFirewallRule{
					{Name: "allow-office", StartIpAddress: "203.0.113.0", EndIpAddress: "203.0.113.255"},
					{Name: "allow-azure", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with ZoneRedundant HA", func() {
				input := minimalPublicSpec()
				input.Spec.SkuName = "GP_Standard_D4ds_v4"
				input.Spec.Zone = "1"
				input.Spec.HighAvailability = &AzureMysqlHighAvailability{
					Mode:                    "ZoneRedundant",
					StandbyAvailabilityZone: "2",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with SameZone HA", func() {
				input := minimalPublicSpec()
				input.Spec.SkuName = "GP_Standard_D2ds_v4"
				input.Spec.HighAvailability = &AzureMysqlHighAvailability{
					Mode: "SameZone",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				version := "8.4"
				autoGrow := false
				backupDays := int32(35)
				geoBackup := true
				input := minimalPublicSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.Version = &version
				input.Spec.SkuName = "GP_Standard_D8ds_v4"
				input.Spec.StorageSizeGb = 1024
				input.Spec.AutoGrowEnabled = &autoGrow
				input.Spec.Zone = "1"
				input.Spec.HighAvailability = &AzureMysqlHighAvailability{
					Mode:                    "ZoneRedundant",
					StandbyAvailabilityZone: "3",
				}
				input.Spec.BackupRetentionDays = &backupDays
				input.Spec.GeoRedundantBackupEnabled = &geoBackup
				input.Spec.Databases = []*AzureMysqlDatabase{
					{Name: "app"},
					{Name: "reporting"},
				}
				input.Spec.FirewallRules = []*AzureMysqlFirewallRule{
					{Name: "allow-vpn", StartIpAddress: "10.0.0.1", EndIpAddress: "10.0.0.1"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for each valid MySQL version", func() {
				versions := []string{"5.7", "8.0.21", "8.4"}
				for _, v := range versions {
					ver := v
					input := minimalPublicSpec()
					input.Spec.Version = &ver
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error when server name starts with a number", func() {
				input := minimalPublicSpec()
				input.Spec.Name = "1-mysql-server"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for resource_group", func() {
				input := minimalPublicSpec()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureResourceGroup,
							Name:      "shared-rg",
							FieldPath: "status.outputs.resource_group_name",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for password", func() {
				input := minimalPublicSpec()
				input.Spec.AdministratorPassword = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureKeyVault,
							Name:      "mysql-secrets",
							FieldPath: "status.outputs.secret_value",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when backup_retention_days is 1 (MySQL minimum)", func() {
				days := int32(1)
				input := minimalPublicSpec()
				input.Spec.BackupRetentionDays = &days
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for large storage (16 TB)", func() {
				input := minimalPublicSpec()
				input.Spec.StorageSizeGb = 16384
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_mysql_flexible_server", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := minimalPublicSpec()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := minimalPublicSpec()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := minimalPublicSpec()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is too short", func() {
				input := minimalPublicSpec()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 63 characters", func() {
				tooLongName := "a"
				for len(tooLongName) < 64 {
					tooLongName += "b"
				}
				input := minimalPublicSpec()
				input.Spec.Name = tooLongName
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains uppercase", func() {
				input := minimalPublicSpec()
				input.Spec.Name = "Invalid-Name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name ends with a hyphen", func() {
				input := minimalPublicSpec()
				input.Spec.Name = "mysql-server-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when administrator_login is missing", func() {
				input := minimalPublicSpec()
				input.Spec.AdministratorLogin = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when administrator_login exceeds 32 characters", func() {
				longLogin := "a"
				for len(longLogin) < 33 {
					longLogin += "b"
				}
				input := minimalPublicSpec()
				input.Spec.AdministratorLogin = longLogin
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when administrator_password is missing", func() {
				input := minimalPublicSpec()
				input.Spec.AdministratorPassword = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when sku_name is missing", func() {
				input := minimalPublicSpec()
				input.Spec.SkuName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage_size_gb is below minimum", func() {
				input := minimalPublicSpec()
				input.Spec.StorageSizeGb = 10
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage_size_gb is zero", func() {
				input := minimalPublicSpec()
				input.Spec.StorageSizeGb = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when version is invalid", func() {
				invalidVersion := "5.6"
				input := minimalPublicSpec()
				input.Spec.Version = &invalidVersion
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when version is 8.0 (must be exact 8.0.21)", func() {
				invalidVersion := "8.0"
				input := minimalPublicSpec()
				input.Spec.Version = &invalidVersion
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when HA mode is invalid", func() {
				input := minimalPublicSpec()
				input.Spec.HighAvailability = &AzureMysqlHighAvailability{
					Mode: "InvalidMode",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when HA mode is missing", func() {
				input := minimalPublicSpec()
				input.Spec.HighAvailability = &AzureMysqlHighAvailability{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backup_retention_days is zero", func() {
				days := int32(0)
				input := minimalPublicSpec()
				input.Spec.BackupRetentionDays = &days
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backup_retention_days exceeds maximum", func() {
				days := int32(40)
				input := minimalPublicSpec()
				input.Spec.BackupRetentionDays = &days
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalPublicSpec()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureMysqlFlexibleServer{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureMysqlFlexibleServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-mysql",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := minimalPublicSpec()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := minimalPublicSpec()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when database name is missing", func() {
				input := minimalPublicSpec()
				input.Spec.Databases = []*AzureMysqlDatabase{
					{Name: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule name is missing", func() {
				input := minimalPublicSpec()
				input.Spec.FirewallRules = []*AzureMysqlFirewallRule{
					{Name: "", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule start_ip is missing", func() {
				input := minimalPublicSpec()
				input.Spec.FirewallRules = []*AzureMysqlFirewallRule{
					{Name: "bad-rule", StartIpAddress: "", EndIpAddress: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule end_ip is missing", func() {
				input := minimalPublicSpec()
				input.Spec.FirewallRules = []*AzureMysqlFirewallRule{
					{Name: "bad-rule", StartIpAddress: "0.0.0.0", EndIpAddress: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

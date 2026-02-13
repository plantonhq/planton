package azurepostgresqlflexibleserverv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzurePostgresqlFlexibleServerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePostgresqlFlexibleServerSpec Validation Tests")
}

// helper to create a minimal valid spec (public access mode)
func minimalPublicSpec() *AzurePostgresqlFlexibleServer {
	return &AzurePostgresqlFlexibleServer{
		ApiVersion: "azure.openmcf.org/v1",
		Kind:       "AzurePostgresqlFlexibleServer",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-pg",
		},
		Spec: &AzurePostgresqlFlexibleServerSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name:               "test-pg-server",
			AdministratorLogin: "pgadmin",
			AdministratorPassword: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "P@ssw0rd1234!",
				},
			},
			SkuName:   "GP_Standard_D2s_v3",
			StorageMb: 32768,
		},
	}
}

var _ = ginkgo.Describe("AzurePostgresqlFlexibleServerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_postgresql_flexible_server", func() {

			ginkgo.It("should not return a validation error for a minimal public server", func() {
				input := minimalPublicSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with VNet integration", func() {
				input := minimalPublicSpec()
				input.Spec.DelegatedSubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/pg-subnet",
					},
				}
				input.Spec.PrivateDnsZoneId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with databases", func() {
				input := minimalPublicSpec()
				input.Spec.Databases = []*AzurePostgresqlDatabase{
					{Name: "myapp"},
					{Name: "analytics"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with databases and custom charset", func() {
				charset := "SQL_ASCII"
				collation := "C"
				input := minimalPublicSpec()
				input.Spec.Databases = []*AzurePostgresqlDatabase{
					{Name: "legacy", Charset: &charset, Collation: &collation},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with firewall rules", func() {
				input := minimalPublicSpec()
				input.Spec.FirewallRules = []*AzurePostgresqlFirewallRule{
					{Name: "allow-office", StartIpAddress: "203.0.113.0", EndIpAddress: "203.0.113.255"},
					{Name: "allow-azure", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with ZoneRedundant HA", func() {
				input := minimalPublicSpec()
				input.Spec.SkuName = "GP_Standard_D4s_v3"
				input.Spec.Zone = "1"
				input.Spec.HighAvailability = &AzurePostgresqlHighAvailability{
					Mode:                    "ZoneRedundant",
					StandbyAvailabilityZone: "2",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with SameZone HA", func() {
				input := minimalPublicSpec()
				input.Spec.SkuName = "GP_Standard_D2s_v3"
				input.Spec.HighAvailability = &AzurePostgresqlHighAvailability{
					Mode: "SameZone",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				version := "17"
				autoGrow := true
				backupDays := int32(35)
				geoBackup := true
				input := minimalPublicSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.Version = &version
				input.Spec.SkuName = "GP_Standard_D8s_v3"
				input.Spec.StorageMb = 1048576
				input.Spec.AutoGrowEnabled = &autoGrow
				input.Spec.Zone = "1"
				input.Spec.HighAvailability = &AzurePostgresqlHighAvailability{
					Mode:                    "ZoneRedundant",
					StandbyAvailabilityZone: "3",
				}
				input.Spec.BackupRetentionDays = &backupDays
				input.Spec.GeoRedundantBackupEnabled = &geoBackup
				input.Spec.Databases = []*AzurePostgresqlDatabase{
					{Name: "app"},
					{Name: "reporting"},
				}
				input.Spec.FirewallRules = []*AzurePostgresqlFirewallRule{
					{Name: "allow-vpn", StartIpAddress: "10.0.0.1", EndIpAddress: "10.0.0.1"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for each valid PostgreSQL version", func() {
				versions := []string{"12", "13", "14", "15", "16", "17"}
				for _, v := range versions {
					ver := v
					input := minimalPublicSpec()
					input.Spec.Version = &ver
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
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
				// Use AzureKeyVault as the kind since it's a valid enum value.
				// In practice, the password source could be any resource that outputs a secret.
				input.Spec.AdministratorPassword = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureKeyVault,
							Name:      "pg-secrets",
							FieldPath: "status.outputs.secret_value",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_postgresql_flexible_server", func() {

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

			ginkgo.It("should return a validation error when name starts with a number", func() {
				input := minimalPublicSpec()
				input.Spec.Name = "1-invalid-name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains uppercase", func() {
				input := minimalPublicSpec()
				input.Spec.Name = "Invalid-Name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when administrator_login is missing", func() {
				input := minimalPublicSpec()
				input.Spec.AdministratorLogin = ""
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

			ginkgo.It("should return a validation error when storage_mb is below minimum", func() {
				input := minimalPublicSpec()
				input.Spec.StorageMb = 16384
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage_mb is zero", func() {
				input := minimalPublicSpec()
				input.Spec.StorageMb = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when version is invalid", func() {
				invalidVersion := "11"
				input := minimalPublicSpec()
				input.Spec.Version = &invalidVersion
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when version is 18 (beta)", func() {
				betaVersion := "18"
				input := minimalPublicSpec()
				input.Spec.Version = &betaVersion
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when HA mode is invalid", func() {
				input := minimalPublicSpec()
				input.Spec.HighAvailability = &AzurePostgresqlHighAvailability{
					Mode: "InvalidMode",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when HA mode is missing", func() {
				input := minimalPublicSpec()
				input.Spec.HighAvailability = &AzurePostgresqlHighAvailability{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backup_retention_days is below minimum", func() {
				days := int32(3)
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
				input := &AzurePostgresqlFlexibleServer{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePostgresqlFlexibleServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pg",
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
				input.Spec.Databases = []*AzurePostgresqlDatabase{
					{Name: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule name is missing", func() {
				input := minimalPublicSpec()
				input.Spec.FirewallRules = []*AzurePostgresqlFirewallRule{
					{Name: "", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule start_ip is missing", func() {
				input := minimalPublicSpec()
				input.Spec.FirewallRules = []*AzurePostgresqlFirewallRule{
					{Name: "bad-rule", StartIpAddress: "", EndIpAddress: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule end_ip is missing", func() {
				input := minimalPublicSpec()
				input.Spec.FirewallRules = []*AzurePostgresqlFirewallRule{
					{Name: "bad-rule", StartIpAddress: "0.0.0.0", EndIpAddress: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

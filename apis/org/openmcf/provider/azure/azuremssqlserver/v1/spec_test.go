package azuremssqlserverv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzureMssqlServerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMssqlServerSpec Validation Tests")
}

// helper to create a minimal valid spec (public access mode, no databases)
func minimalSpec() *AzureMssqlServer {
	return &AzureMssqlServer{
		ApiVersion: "azure.openmcf.org/v1",
		Kind:       "AzureMssqlServer",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-sql",
		},
		Spec: &AzureMssqlServerSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name:               "test-sql-server",
			AdministratorLogin: "sqladmin",
			AdministratorPassword: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "P@ssw0rd1234!",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMssqlServerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_mssql_server", func() {

			ginkgo.It("should not return a validation error for a minimal server", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with databases", func() {
				input := minimalSpec()
				input.Spec.Databases = []*AzureMssqlDatabase{
					{Name: "myapp", SkuName: "S0"},
					{Name: "analytics", SkuName: "GP_Gen5_2"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a database with all optional fields", func() {
				maxSize := int32(250)
				collation := "SQL_Latin1_General_CP1_CS_AS"
				zoneRedundant := true
				licenseType := "BasePrice"
				storageType := "Zone"
				input := minimalSpec()
				input.Spec.Databases = []*AzureMssqlDatabase{
					{
						Name:               "enterprise",
						SkuName:            "BC_Gen5_4",
						MaxSizeGb:          &maxSize,
						Collation:          &collation,
						ZoneRedundant:      &zoneRedundant,
						LicenseType:        &licenseType,
						StorageAccountType: &storageType,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a server with firewall rules", func() {
				input := minimalSpec()
				input.Spec.FirewallRules = []*AzureMssqlFirewallRule{
					{Name: "allow-office", StartIpAddress: "203.0.113.0", EndIpAddress: "203.0.113.255"},
					{Name: "allow-azure", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional server fields set", func() {
				version := "12.0"
				tls := "1.2"
				publicAccess := false
				connPolicy := "Redirect"
				input := minimalSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.Version = &version
				input.Spec.MinimumTlsVersion = &tls
				input.Spec.PublicNetworkAccessEnabled = &publicAccess
				input.Spec.ConnectionPolicy = &connPolicy
				input.Spec.Databases = []*AzureMssqlDatabase{
					{Name: "app", SkuName: "GP_Gen5_2"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for each valid version", func() {
				versions := []string{"2.0", "12.0"}
				for _, v := range versions {
					ver := v
					input := minimalSpec()
					input.Spec.Version = &ver
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for each valid TLS version", func() {
				tlsVersions := []string{"1.0", "1.2"}
				for _, v := range tlsVersions {
					ver := v
					input := minimalSpec()
					input.Spec.MinimumTlsVersion = &ver
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for each valid connection policy", func() {
				policies := []string{"Default", "Proxy", "Redirect"}
				for _, p := range policies {
					pol := p
					input := minimalSpec()
					input.Spec.ConnectionPolicy = &pol
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for each valid license type", func() {
				types := []string{"BasePrice", "LicenseIncluded"}
				for _, lt := range types {
					licType := lt
					input := minimalSpec()
					input.Spec.Databases = []*AzureMssqlDatabase{
						{Name: "testdb", SkuName: "GP_Gen5_2", LicenseType: &licType},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error for each valid storage account type", func() {
				types := []string{"Geo", "GeoZone", "Local", "Zone"}
				for _, st := range types {
					stType := st
					input := minimalSpec()
					input.Spec.Databases = []*AzureMssqlDatabase{
						{Name: "testdb", SkuName: "S0", StorageAccountType: &stType},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error with valueFrom reference for resource_group", func() {
				input := minimalSpec()
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
				input := minimalSpec()
				input.Spec.AdministratorPassword = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureKeyVault,
							Name:      "sql-secrets",
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
		ginkgo.Context("azure_mssql_server", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := minimalSpec()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := minimalSpec()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is too short", func() {
				input := minimalSpec()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 63 characters", func() {
				tooLongName := "a"
				for len(tooLongName) < 64 {
					tooLongName += "b"
				}
				input := minimalSpec()
				input.Spec.Name = tooLongName
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a number", func() {
				input := minimalSpec()
				input.Spec.Name = "1-invalid-name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains uppercase", func() {
				input := minimalSpec()
				input.Spec.Name = "Invalid-Name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when administrator_login is missing", func() {
				input := minimalSpec()
				input.Spec.AdministratorLogin = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when administrator_password is missing", func() {
				input := minimalSpec()
				input.Spec.AdministratorPassword = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when version is invalid", func() {
				invalidVersion := "15.0"
				input := minimalSpec()
				input.Spec.Version = &invalidVersion
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when minimum_tls_version is invalid", func() {
				invalidTls := "1.1"
				input := minimalSpec()
				input.Spec.MinimumTlsVersion = &invalidTls
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when connection_policy is invalid", func() {
				invalidPolicy := "Direct"
				input := minimalSpec()
				input.Spec.ConnectionPolicy = &invalidPolicy
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when database name is missing", func() {
				input := minimalSpec()
				input.Spec.Databases = []*AzureMssqlDatabase{
					{Name: "", SkuName: "S0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when database sku_name is missing", func() {
				input := minimalSpec()
				input.Spec.Databases = []*AzureMssqlDatabase{
					{Name: "testdb", SkuName: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when database max_size_gb is zero", func() {
				maxSize := int32(0)
				input := minimalSpec()
				input.Spec.Databases = []*AzureMssqlDatabase{
					{Name: "testdb", SkuName: "S0", MaxSizeGb: &maxSize},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when license_type is invalid", func() {
				invalidLicense := "Free"
				input := minimalSpec()
				input.Spec.Databases = []*AzureMssqlDatabase{
					{Name: "testdb", SkuName: "S0", LicenseType: &invalidLicense},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage_account_type is invalid", func() {
				invalidStorage := "Premium"
				input := minimalSpec()
				input.Spec.Databases = []*AzureMssqlDatabase{
					{Name: "testdb", SkuName: "S0", StorageAccountType: &invalidStorage},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule name is missing", func() {
				input := minimalSpec()
				input.Spec.FirewallRules = []*AzureMssqlFirewallRule{
					{Name: "", StartIpAddress: "0.0.0.0", EndIpAddress: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule start_ip is missing", func() {
				input := minimalSpec()
				input.Spec.FirewallRules = []*AzureMssqlFirewallRule{
					{Name: "bad-rule", StartIpAddress: "", EndIpAddress: "0.0.0.0"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when firewall rule end_ip is missing", func() {
				input := minimalSpec()
				input.Spec.FirewallRules = []*AzureMssqlFirewallRule{
					{Name: "bad-rule", StartIpAddress: "0.0.0.0", EndIpAddress: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalSpec()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureMssqlServer{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureMssqlServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-sql",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := minimalSpec()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := minimalSpec()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

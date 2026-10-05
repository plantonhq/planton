package azuredatafactorylinkedservicev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureDataFactoryLinkedServiceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataFactoryLinkedServiceSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testFactoryId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DataFactory/factories/app-df"
	testKeyVaultId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.KeyVault/vaults/app-kv"
	testSpId       = "12345678-1234-1234-1234-123456789012"
)

// kvSecret builds a Key-Vault-sourced secret reference.
func kvSecret(linkedServiceName, secretName string) *AzureDataFactoryLinkedServiceKeyVaultSecretRef {
	return &AzureDataFactoryLinkedServiceKeyVaultSecretRef{
		LinkedServiceName: literal(linkedServiceName),
		SecretName:        secretName,
	}
}

// validResource returns a valid key_vault linked service (the
// simplest variant) that individual cases mutate into the shape under
// test.
func validResource() *AzureDataFactoryLinkedService {
	return &AzureDataFactoryLinkedService{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataFactoryLinkedService",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-adf-linked-service",
		},
		Spec: &AzureDataFactoryLinkedServiceSpec{
			DataFactoryId: literal(testFactoryId),
			Name:          "secrets-vault",
			KeyVault: &AzureDataFactoryLinkedServiceKeyVault{
				KeyVaultId: literal(testKeyVaultId),
			},
		},
	}
}

// withVariant clears the fixture's key_vault variant so a case can
// install a different one.
func withoutVariant() *AzureDataFactoryLinkedService {
	input := validResource()
	input.Spec.KeyVault = nil
	return input
}

var _ = ginkgo.Describe("AzureDataFactoryLinkedServiceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("variant selection -- one minimal document per variant", func() {

			ginkgo.It("should accept a key vault connection", func() {
				gomega.Expect(protovalidate.Validate(validResource())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a blob storage connection by connection string", func() {
				input := withoutVariant()
				input.Spec.AzureBlobStorage = &AzureDataFactoryLinkedServiceAzureBlobStorage{
					ConnectionString: "DefaultEndpointsProtocol=https;AccountName=app;AccountKey=secret",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a blob storage connection by service endpoint + managed identity", func() {
				input := withoutVariant()
				input.Spec.AzureBlobStorage = &AzureDataFactoryLinkedServiceAzureBlobStorage{
					ServiceEndpoint:    literal("https://app.blob.core.windows.net"),
					UseManagedIdentity: proto.Bool(true),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a databricks connection (MSI + new cluster)", func() {
				input := withoutVariant()
				input.Spec.AzureDatabricks = &AzureDataFactoryLinkedServiceAzureDatabricks{
					AdbDomain:      "https://adb-1234567890123456.7.azuredatabricks.net",
					MsiWorkspaceId: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Databricks/workspaces/app-dbw",
					NewClusterConfig: &AzureDataFactoryLinkedServiceDatabricksNewCluster{
						NodeType:       "Standard_DS3_v2",
						ClusterVersion: "16.4.x-scala2.12",
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a file storage connection", func() {
				input := withoutVariant()
				input.Spec.AzureFileStorage = &AzureDataFactoryLinkedServiceAzureFileStorage{
					ConnectionString: "DefaultEndpointsProtocol=https;AccountName=app;AccountKey=secret",
					FileShare:        "landing",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an azure function connection with a key vault key", func() {
				input := withoutVariant()
				input.Spec.AzureFunction = &AzureDataFactoryLinkedServiceAzureFunction{
					Url:         "https://app.azurewebsites.net",
					KeyVaultKey: kvSecret("secrets-vault", "fn-host-key"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an azure search connection", func() {
				input := withoutVariant()
				input.Spec.AzureSearch = &AzureDataFactoryLinkedServiceAzureSearch{
					Url:              literal("https://app-search.search.windows.net"),
					SearchServiceKey: "admin-key",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an azure sql database connection by key vault connection string", func() {
				input := withoutVariant()
				input.Spec.AzureSqlDatabase = &AzureDataFactoryLinkedServiceAzureSqlDatabase{
					KeyVaultConnectionString: kvSecret("secrets-vault", "sqldb-conn"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a table storage connection", func() {
				input := withoutVariant()
				input.Spec.AzureTableStorage = &AzureDataFactoryLinkedServiceAzureTableStorage{
					ConnectionString: "DefaultEndpointsProtocol=https;AccountName=app;AccountKey=secret",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a cosmosdb connection by account detail", func() {
				input := withoutVariant()
				input.Spec.Cosmosdb = &AzureDataFactoryLinkedServiceCosmosdb{
					AccountEndpoint: "https://app.documents.azure.com:443/",
					AccountKey:      "account-key",
					Database:        "appdb",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a cosmosdb mongo connection", func() {
				input := withoutVariant()
				input.Spec.CosmosdbMongoapi = &AzureDataFactoryLinkedServiceCosmosdbMongoapi{
					ConnectionString: "mongodb://app:key@app.documents.azure.com:10255/?ssl=true",
					Database:         "appdb",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom connection", func() {
				input := withoutVariant()
				input.Spec.Custom = &AzureDataFactoryLinkedServiceCustom{
					Type:               "RestService",
					TypePropertiesJson: `{"url":"https://api.example.com","enableServerCertificateValidation":true,"authenticationType":"Anonymous"}`,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a data lake gen2 connection with managed identity", func() {
				input := withoutVariant()
				input.Spec.DataLakeStorageGen2 = &AzureDataFactoryLinkedServiceDataLakeStorageGen2{
					Url:                literal("https://app.dfs.core.windows.net"),
					UseManagedIdentity: proto.Bool(true),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a kusto connection with managed identity", func() {
				input := withoutVariant()
				input.Spec.Kusto = &AzureDataFactoryLinkedServiceKusto{
					KustoEndpoint:      "https://appadx.westeurope.kusto.windows.net",
					KustoDatabaseName:  "appdb",
					UseManagedIdentity: proto.Bool(true),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a mysql connection", func() {
				input := withoutVariant()
				input.Spec.Mysql = &AzureDataFactoryLinkedServiceMysql{
					ConnectionString: "Server=db.example.com;Port=3306;Database=appdb;UID=app;PWD=secret",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an anonymous odata connection", func() {
				input := withoutVariant()
				input.Spec.Odata = &AzureDataFactoryLinkedServiceOdata{
					Url: "https://services.odata.org/V4/TripPinServiceRW/",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an odbc connection with basic authentication", func() {
				input := withoutVariant()
				input.Spec.Odbc = &AzureDataFactoryLinkedServiceOdbc{
					ConnectionString: "Driver={SQL Server};Server=db.internal;Database=appdb",
					BasicAuthentication: &AzureDataFactoryLinkedServiceBasicAuth{
						Username: "app",
						Password: "secret",
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a postgresql connection", func() {
				input := withoutVariant()
				input.Spec.Postgresql = &AzureDataFactoryLinkedServicePostgresql{
					ConnectionString: "Host=db.example.com;Port=5432;Database=appdb;UID=app;Password=secret",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an sftp connection with basic authentication", func() {
				input := withoutVariant()
				input.Spec.Sftp = &AzureDataFactoryLinkedServiceSftp{
					AuthenticationType: "Basic",
					Host:               "sftp.example.com",
					Port:               22,
					Username:           "app",
					Password:           "secret",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a snowflake connection with a key vault password", func() {
				input := withoutVariant()
				input.Spec.Snowflake = &AzureDataFactoryLinkedServiceSnowflake{
					ConnectionString: "jdbc:snowflake://app.snowflakecomputing.com/?user=app&db=appdb&warehouse=wh",
					KeyVaultPassword: kvSecret("secrets-vault", "snowflake-password"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a sql managed instance connection with a service principal", func() {
				input := withoutVariant()
				input.Spec.SqlManagedInstance = &AzureDataFactoryLinkedServiceSqlManagedInstance{
					ConnectionString:    "Server=app-mi.public.dns.zone;Database=appdb",
					ServicePrincipalId:  testSpId,
					ServicePrincipalKey: "sp-secret",
					Tenant:              testSpId,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a sql server connection", func() {
				input := withoutVariant()
				input.Spec.SqlServer = &AzureDataFactoryLinkedServiceSqlServer{
					ConnectionString: "Server=db.internal;Database=appdb;Integrated Security=False",
					UserName:         "app",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a synapse connection", func() {
				input := withoutVariant()
				input.Spec.Synapse = &AzureDataFactoryLinkedServiceSynapse{
					ConnectionString: "Server=app-ws.sql.azuresynapse.net;Database=appdb",
					KeyVaultPassword: kvSecret("secrets-vault", "synapse-password"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an anonymous web connection", func() {
				input := withoutVariant()
				input.Spec.Web = &AzureDataFactoryLinkedServiceWeb{
					Url:                "https://example.com/data.csv",
					AuthenticationType: "Anonymous",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("shared root fields", func() {

			ginkgo.It("should accept the full shared surface", func() {
				input := validResource()
				input.Spec.Description = "the vault every other connection resolves secrets through"
				input.Spec.Annotations = []string{"team:data"}
				input.Spec.Parameters = map[string]string{"env": "prod"}
				input.Spec.AdditionalProperties = map[string]string{"connectVia.extra": "x"}
				input.Spec.IntegrationRuntimeName = literal("self-hosted-ir")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name containing special characters mixed with letters", func() {
				input := validResource()
				input.Spec.Name = "vault-connection.v2"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("variant selection", func() {

			ginkgo.It("should reject a spec with no variant block", func() {
				input := withoutVariant()
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a spec with two variant blocks", func() {
				input := validResource()
				input.Spec.Web = &AzureDataFactoryLinkedServiceWeb{
					Url:                "https://example.com",
					AuthenticationType: "Anonymous",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("shared root fields", func() {

			ginkgo.It("should reject a missing data_factory_id", func() {
				input := validResource()
				input.Spec.DataFactoryId = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name consisting entirely of special characters", func() {
				input := validResource()
				input.Spec.Name = "-.+?/"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty annotation entry", func() {
				input := validResource()
				input.Spec.Annotations = []string{""}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("blob storage variant", func() {

			ginkgo.It("should reject a blob connection with no connection form", func() {
				input := withoutVariant()
				input.Spec.AzureBlobStorage = &AzureDataFactoryLinkedServiceAzureBlobStorage{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a blob connection with two connection forms", func() {
				input := withoutVariant()
				input.Spec.AzureBlobStorage = &AzureDataFactoryLinkedServiceAzureBlobStorage{
					ConnectionString: "DefaultEndpointsProtocol=https;AccountName=app;AccountKey=secret",
					SasUri:           "https://app.blob.core.windows.net/?sv=token",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject managed identity combined with a service principal", func() {
				input := withoutVariant()
				input.Spec.AzureBlobStorage = &AzureDataFactoryLinkedServiceAzureBlobStorage{
					ServiceEndpoint:    literal("https://app.blob.core.windows.net"),
					UseManagedIdentity: proto.Bool(true),
					ServicePrincipalId: testSpId,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed service principal id", func() {
				input := withoutVariant()
				input.Spec.AzureBlobStorage = &AzureDataFactoryLinkedServiceAzureBlobStorage{
					ServiceEndpoint:    literal("https://app.blob.core.windows.net"),
					ServicePrincipalId: "not-a-uuid",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown storage kind", func() {
				input := withoutVariant()
				input.Spec.AzureBlobStorage = &AzureDataFactoryLinkedServiceAzureBlobStorage{
					ConnectionString: "DefaultEndpointsProtocol=https;AccountName=app;AccountKey=secret",
					StorageKind:      "PremiumBlob",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("databricks variant", func() {

			ginkgo.It("should reject a databricks connection with no authentication", func() {
				input := withoutVariant()
				input.Spec.AzureDatabricks = &AzureDataFactoryLinkedServiceAzureDatabricks{
					AdbDomain:         "https://adb-1.azuredatabricks.net",
					ExistingClusterId: "0000-000000-abcdefgh",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a databricks connection with two authentications", func() {
				input := withoutVariant()
				input.Spec.AzureDatabricks = &AzureDataFactoryLinkedServiceAzureDatabricks{
					AdbDomain:         "https://adb-1.azuredatabricks.net",
					MsiWorkspaceId:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Databricks/workspaces/app-dbw",
					AccessToken:       "dapi-token",
					ExistingClusterId: "0000-000000-abcdefgh",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a databricks connection with no cluster choice", func() {
				input := withoutVariant()
				input.Spec.AzureDatabricks = &AzureDataFactoryLinkedServiceAzureDatabricks{
					AdbDomain:      "https://adb-1.azuredatabricks.net",
					MsiWorkspaceId: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Databricks/workspaces/app-dbw",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject autoscaling bounds with max below min", func() {
				input := withoutVariant()
				input.Spec.AzureDatabricks = &AzureDataFactoryLinkedServiceAzureDatabricks{
					AdbDomain:      "https://adb-1.azuredatabricks.net",
					MsiWorkspaceId: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Databricks/workspaces/app-dbw",
					NewClusterConfig: &AzureDataFactoryLinkedServiceDatabricksNewCluster{
						NodeType:           "Standard_DS3_v2",
						ClusterVersion:     "16.4.x-scala2.12",
						MinNumberOfWorkers: proto.Int32(4),
						MaxNumberOfWorkers: 2,
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("azure function variant", func() {

			ginkgo.It("should reject a function connection with no key form", func() {
				input := withoutVariant()
				input.Spec.AzureFunction = &AzureDataFactoryLinkedServiceAzureFunction{
					Url: "https://app.azurewebsites.net",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a function connection with both key forms", func() {
				input := withoutVariant()
				input.Spec.AzureFunction = &AzureDataFactoryLinkedServiceAzureFunction{
					Url:         "https://app.azurewebsites.net",
					Key:         "host-key",
					KeyVaultKey: kvSecret("secrets-vault", "fn-host-key"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("azure sql database variant", func() {

			ginkgo.It("should reject both connection forms together", func() {
				input := withoutVariant()
				input.Spec.AzureSqlDatabase = &AzureDataFactoryLinkedServiceAzureSqlDatabase{
					ConnectionString:         "Server=app.database.windows.net;Database=appdb",
					KeyVaultConnectionString: kvSecret("secrets-vault", "sqldb-conn"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service principal id without its key", func() {
				input := withoutVariant()
				input.Spec.AzureSqlDatabase = &AzureDataFactoryLinkedServiceAzureSqlDatabase{
					ConnectionString:   "Server=app.database.windows.net;Database=appdb",
					ServicePrincipalId: testSpId,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("cosmosdb variant", func() {

			ginkgo.It("should reject both connection forms together", func() {
				input := withoutVariant()
				input.Spec.Cosmosdb = &AzureDataFactoryLinkedServiceCosmosdb{
					ConnectionString: "AccountEndpoint=https://app.documents.azure.com;AccountKey=key",
					AccountEndpoint:  "https://app.documents.azure.com:443/",
					AccountKey:       "key",
					Database:         "appdb",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject incomplete account detail", func() {
				input := withoutVariant()
				input.Spec.Cosmosdb = &AzureDataFactoryLinkedServiceCosmosdb{
					AccountEndpoint: "https://app.documents.azure.com:443/",
					AccountKey:      "key",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("data lake gen2 variant", func() {

			ginkgo.It("should reject a gen2 connection with no authentication mode", func() {
				input := withoutVariant()
				input.Spec.DataLakeStorageGen2 = &AzureDataFactoryLinkedServiceDataLakeStorageGen2{
					Url: literal("https://app.dfs.core.windows.net"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject two authentication modes together", func() {
				input := withoutVariant()
				input.Spec.DataLakeStorageGen2 = &AzureDataFactoryLinkedServiceDataLakeStorageGen2{
					Url:                literal("https://app.dfs.core.windows.net"),
					UseManagedIdentity: proto.Bool(true),
					StorageAccountKey:  "account-key",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an incomplete service principal", func() {
				input := withoutVariant()
				input.Spec.DataLakeStorageGen2 = &AzureDataFactoryLinkedServiceDataLakeStorageGen2{
					Url:                literal("https://app.dfs.core.windows.net"),
					ServicePrincipalId: testSpId,
					// service_principal_key and tenant are missing.
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("kusto variant", func() {

			ginkgo.It("should reject a kusto connection with no authentication mode", func() {
				input := withoutVariant()
				input.Spec.Kusto = &AzureDataFactoryLinkedServiceKusto{
					KustoEndpoint:     "https://appadx.westeurope.kusto.windows.net",
					KustoDatabaseName: "appdb",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service principal id without its key", func() {
				input := withoutVariant()
				input.Spec.Kusto = &AzureDataFactoryLinkedServiceKusto{
					KustoEndpoint:      "https://appadx.westeurope.kusto.windows.net",
					KustoDatabaseName:  "appdb",
					ServicePrincipalId: testSpId,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a tenant declared outside the service principal mode", func() {
				input := withoutVariant()
				input.Spec.Kusto = &AzureDataFactoryLinkedServiceKusto{
					KustoEndpoint:      "https://appadx.westeurope.kusto.windows.net",
					KustoDatabaseName:  "appdb",
					UseManagedIdentity: proto.Bool(true),
					Tenant:             "tenant-id",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("mysql variant", func() {

			ginkgo.It("should reject an unknown driver version", func() {
				input := withoutVariant()
				input.Spec.Mysql = &AzureDataFactoryLinkedServiceMysql{
					ConnectionString: "Server=db;Database=appdb",
					DriverVersion:    proto.String("V3"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("basic authentication block (odata/odbc)", func() {

			ginkgo.It("should reject basic authentication without a password", func() {
				input := withoutVariant()
				input.Spec.Odata = &AzureDataFactoryLinkedServiceOdata{
					Url: "https://services.odata.org/V4/TripPinServiceRW/",
					BasicAuthentication: &AzureDataFactoryLinkedServiceBasicAuth{
						Username: "app",
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("sftp variant", func() {

			ginkgo.It("should reject a missing port", func() {
				input := withoutVariant()
				input.Spec.Sftp = &AzureDataFactoryLinkedServiceSftp{
					AuthenticationType: "Basic",
					Host:               "sftp.example.com",
					Username:           "app",
					Password:           "secret",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a password held in two homes", func() {
				input := withoutVariant()
				input.Spec.Sftp = &AzureDataFactoryLinkedServiceSftp{
					AuthenticationType: "Basic",
					Host:               "sftp.example.com",
					Port:               22,
					Username:           "app",
					Password:           "secret",
					KeyVaultPassword:   kvSecret("secrets-vault", "sftp-password"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a private key held in two homes", func() {
				input := withoutVariant()
				input.Spec.Sftp = &AzureDataFactoryLinkedServiceSftp{
					AuthenticationType:              "SshPublicKey",
					Host:                            "sftp.example.com",
					Port:                            22,
					Username:                        "app",
					PrivateKeyContentBase64:         "LS0tLS1CRUdJTg==",
					KeyVaultPrivateKeyContentBase64: kvSecret("secrets-vault", "sftp-key"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a password on an SshPublicKey connection", func() {
				input := withoutVariant()
				input.Spec.Sftp = &AzureDataFactoryLinkedServiceSftp{
					AuthenticationType: "SshPublicKey",
					Host:               "sftp.example.com",
					Port:               22,
					Username:           "app",
					Password:           "secret",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a private key on a Basic connection", func() {
				input := withoutVariant()
				input.Spec.Sftp = &AzureDataFactoryLinkedServiceSftp{
					AuthenticationType:      "Basic",
					Host:                    "sftp.example.com",
					Port:                    22,
					Username:                "app",
					Password:                "secret",
					PrivateKeyContentBase64: "LS0tLS1CRUdJTg==",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should accept a managed-secret reference in the private key", func() {
				// Sensitive fields carry secret references on consuming
				// platforms -- the base64 framing is taught in the field
				// comment, never enforced on the stored value.
				input := withoutVariant()
				input.Spec.Sftp = &AzureDataFactoryLinkedServiceSftp{
					AuthenticationType:      "SshPublicKey",
					Host:                    "sftp.example.com",
					Port:                    22,
					Username:                "app",
					PrivateKeyContentBase64: "${secrets-group/sftp/private-key}",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown authentication type", func() {
				input := withoutVariant()
				input.Spec.Sftp = &AzureDataFactoryLinkedServiceSftp{
					AuthenticationType: "Kerberos",
					Host:               "sftp.example.com",
					Port:               22,
					Username:           "app",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("sql managed instance variant", func() {

			ginkgo.It("should reject a partial service principal trio", func() {
				input := withoutVariant()
				input.Spec.SqlManagedInstance = &AzureDataFactoryLinkedServiceSqlManagedInstance{
					ConnectionString:   "Server=app-mi.public.dns.zone;Database=appdb",
					ServicePrincipalId: testSpId,
					// key and tenant are missing.
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a connection with neither connection form", func() {
				input := withoutVariant()
				input.Spec.SqlManagedInstance = &AzureDataFactoryLinkedServiceSqlManagedInstance{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("sql server variant", func() {

			ginkgo.It("should reject both connection forms together", func() {
				input := withoutVariant()
				input.Spec.SqlServer = &AzureDataFactoryLinkedServiceSqlServer{
					ConnectionString:         "Server=db.internal;Database=appdb",
					KeyVaultConnectionString: kvSecret("secrets-vault", "sql-conn"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("web variant", func() {

			ginkgo.It("should reject Basic authentication without credentials", func() {
				input := withoutVariant()
				input.Spec.Web = &AzureDataFactoryLinkedServiceWeb{
					Url:                "https://example.com",
					AuthenticationType: "Basic",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the authentication type the provider does not wire", func() {
				input := withoutVariant()
				input.Spec.Web = &AzureDataFactoryLinkedServiceWeb{
					Url:                "https://example.com",
					AuthenticationType: "ClientCertificate",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("key vault secret references", func() {

			ginkgo.It("should reject a secret reference without a secret name", func() {
				input := withoutVariant()
				input.Spec.Snowflake = &AzureDataFactoryLinkedServiceSnowflake{
					ConnectionString: "jdbc:snowflake://app.snowflakecomputing.com/?user=app",
					KeyVaultPassword: &AzureDataFactoryLinkedServiceKeyVaultSecretRef{
						LinkedServiceName: literal("secrets-vault"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a secret reference without a linked service name", func() {
				input := withoutVariant()
				input.Spec.Snowflake = &AzureDataFactoryLinkedServiceSnowflake{
					ConnectionString: "jdbc:snowflake://app.snowflakecomputing.com/?user=app",
					KeyVaultPassword: &AzureDataFactoryLinkedServiceKeyVaultSecretRef{
						SecretName: "snowflake-password",
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("custom variant", func() {

			ginkgo.It("should reject a custom connection without a type", func() {
				input := withoutVariant()
				input.Spec.Custom = &AzureDataFactoryLinkedServiceCustom{
					TypePropertiesJson: `{"url":"https://api.example.com"}`,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a custom connection without type properties", func() {
				input := withoutVariant()
				input.Spec.Custom = &AzureDataFactoryLinkedServiceCustom{
					Type: "RestService",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})

package azurelinuxwebappv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureLinuxWebAppSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureLinuxWebAppSpec Validation Tests")
}

// helper to create a pointer to a bool value
func boolPtr(b bool) *bool { return &b }

// helper to create a pointer to a string value
func stringPtr(s string) *string { return &s }

// helper to create a pointer to an int32 value
func int32Ptr(i int32) *int32 { return &i }

// helper to create a literal StringValueOrRef
func literalRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: val,
		},
	}
}

// helper to create a minimal valid spec
func minimalSpec() *AzureLinuxWebApp {
	return &AzureLinuxWebApp{
		ApiVersion: "azure.planton.dev/v1",
		Kind:       "AzureLinuxWebApp",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-webapp",
		},
		Spec: &AzureLinuxWebAppSpec{
			Region:        "eastus",
			ResourceGroup: literalRef("my-rg"),
			Name:          "my-web-app",
			ServicePlanId: literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Web/serverfarms/plan"),
			SiteConfig:    &AzureLinuxWebAppSiteConfig{},
		},
	}
}

var _ = ginkgo.Describe("AzureLinuxWebAppSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_linux_web_app", func() {

			ginkgo.It("should not return a validation error for a minimal spec", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with Python application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					PythonVersion: "3.12",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with Node.js application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					NodeVersion: "22-lts",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with .NET application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					DotnetVersion: "8.0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with PHP application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					PhpVersion: "8.3",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with Ruby application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					RubyVersion: "2.7",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with Go application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					GoVersion: "1.19",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with Java application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					JavaVersion:       "17",
					JavaServer:        "TOMCAT",
					JavaServerVersion: "10.0-java17",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with Docker application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					Docker: &AzureLinuxWebAppDockerConfig{
						RegistryUrl:      "https://myregistry.azurecr.io",
						ImageName:        "myorg/my-web-app",
						ImageTag:         "v1.0.0",
						RegistryUsername: "myuser",
						RegistryPassword: literalRef("mypassword"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with app_settings map", func() {
				input := minimalSpec()
				input.Spec.AppSettings = map[string]string{
					"MY_SETTING":      "value1",
					"ANOTHER_SETTING": "value2",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with connection_strings", func() {
				input := minimalSpec()
				input.Spec.ConnectionStrings = []*AzureLinuxWebAppConnectionString{
					{
						Name:  "MyDatabase",
						Type:  "SQLAzure",
						Value: literalRef("Server=tcp:myserver.database.windows.net;Database=mydb;"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with application_insights_connection_string (valueFrom)", func() {
				input := minimalSpec()
				input.Spec.ApplicationInsightsConnectionString = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureApplicationInsights,
							Name:      "my-appinsights",
							FieldPath: "status.outputs.connection_string",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with virtual_network_subnet_id", func() {
				input := minimalSpec()
				input.Spec.VirtualNetworkSubnetId = literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet1")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with SystemAssigned identity", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: "SystemAssigned",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with UserAssigned identity and identity_ids", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: "UserAssigned",
					IdentityIds: []*foreignkeyv1.StringValueOrRef{
						literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/my-id"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with SystemAssigned,UserAssigned identity", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: "SystemAssigned,UserAssigned",
					IdentityIds: []*foreignkeyv1.StringValueOrRef{
						literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/my-id"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with health_check_path in site_config", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.HealthCheckPath = "/api/health"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with health_check_eviction_time of 5 minutes", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = int32Ptr(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with always_on in site_config", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AlwaysOn = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with cors settings", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.Cors = &AzureLinuxWebAppCorsSettings{
					AllowedOrigins:     []string{"https://myapp.example.com", "https://admin.example.com"},
					SupportCredentials: boolPtr(true),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with IP restrictions (ip_address)", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictions = []*AzureLinuxWebAppIpRestriction{
					{
						Name:      "AllowOffice",
						Priority:  int32Ptr(100),
						Action:    stringPtr("Allow"),
						IpAddress: "203.0.113.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with IP restrictions (service_tag)", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictions = []*AzureLinuxWebAppIpRestriction{
					{
						Name:       "AllowFrontDoor",
						Priority:   int32Ptr(100),
						Action:     stringPtr("Allow"),
						ServiceTag: "AzureFrontDoor.Backend",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with storage mounts (AzureFiles)", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "data-mount",
						Type:        "AzureFiles",
						AccountName: "mystore",
						ShareName:   "myshare",
						AccessKey:   literalRef("abc123secretkey=="),
						MountPath:   "/mnt/data",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a spec with storage mounts (AzureBlob)", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "blob-mount",
						Type:        "AzureBlob",
						AccountName: "blobstore",
						ShareName:   "my-container",
						AccessKey:   literalRef("blobkey=="),
						MountPath:   "/mnt/blobs",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				input := minimalSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.AppSettings = map[string]string{
					"APP_ENV": "production",
				}
				input.Spec.ConnectionStrings = []*AzureLinuxWebAppConnectionString{
					{
						Name:  "Redis",
						Type:  "RedisCache",
						Value: literalRef("redis://my-redis:6379"),
					},
				}
				input.Spec.ApplicationInsightsConnectionString = literalRef("InstrumentationKey=abc-def-123")
				input.Spec.HttpsOnly = boolPtr(true)
				input.Spec.PublicNetworkAccessEnabled = boolPtr(true)
				input.Spec.Enabled = boolPtr(true)
				input.Spec.ClientAffinityEnabled = boolPtr(false)
				input.Spec.VirtualNetworkSubnetId = literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet1")
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: "SystemAssigned",
				}
				input.Spec.KeyVaultReferenceIdentityId = literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/kv-id")
				input.Spec.ClientCertificateEnabled = boolPtr(true)
				input.Spec.ClientCertificateMode = stringPtr("Required")
				input.Spec.ClientCertificateExclusionPaths = "/api/health;/api/status"
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					PythonVersion: "3.12",
				}
				input.Spec.SiteConfig.AlwaysOn = boolPtr(true)
				input.Spec.SiteConfig.AppCommandLine = "gunicorn --bind=0.0.0.0"
				input.Spec.SiteConfig.HealthCheckPath = "/api/health"
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = int32Ptr(5)
				input.Spec.SiteConfig.MinimumTlsVersion = stringPtr("1.2")
				input.Spec.SiteConfig.ScmMinimumTlsVersion = stringPtr("1.2")
				input.Spec.SiteConfig.WorkerCount = int32Ptr(3)
				input.Spec.SiteConfig.Http2Enabled = boolPtr(true)
				input.Spec.SiteConfig.WebsocketsEnabled = boolPtr(false)
				input.Spec.SiteConfig.Use_32BitWorker = boolPtr(false)
				input.Spec.SiteConfig.VnetRouteAllEnabled = boolPtr(true)
				input.Spec.SiteConfig.FtpsState = stringPtr("Disabled")
				input.Spec.SiteConfig.LoadBalancingMode = stringPtr("LeastRequests")
				input.Spec.SiteConfig.Cors = &AzureLinuxWebAppCorsSettings{
					AllowedOrigins: []string{"https://myapp.example.com"},
				}
				input.Spec.SiteConfig.IpRestrictions = []*AzureLinuxWebAppIpRestriction{
					{
						Name:      "AllowAll",
						Priority:  int32Ptr(100),
						Action:    stringPtr("Allow"),
						IpAddress: "0.0.0.0/0",
					},
				}
				input.Spec.SiteConfig.IpRestrictionDefaultAction = stringPtr("Deny")
				input.Spec.SiteConfig.ScmUseMainIpRestriction = boolPtr(true)
				input.Spec.SiteConfig.ContainerRegistryUseManagedIdentity = boolPtr(false)
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "logs",
						Type:        "AzureBlob",
						AccountName: "logstore",
						ShareName:   "logs-container",
						AccessKey:   literalRef("logstorekey=="),
						MountPath:   "/mnt/logs",
					},
				}
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					ApplicationLogs: &AzureLinuxWebAppApplicationLogs{
						FileSystemLevel: stringPtr("Information"),
					},
					HttpLogs: &AzureLinuxWebAppHttpLogs{
						RetentionInMb:   int32Ptr(50),
						RetentionInDays: int32Ptr(7),
					},
					FailedRequestTracing:  boolPtr(true),
					DetailedErrorMessages: boolPtr(false),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
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

			ginkgo.It("should not return a validation error with valueFrom reference for service_plan_id", func() {
				input := minimalSpec()
				input.Spec.ServicePlanId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureServicePlan,
							Name:      "my-plan",
							FieldPath: "status.outputs.plan_id",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with Docker config including registry credentials", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					Docker: &AzureLinuxWebAppDockerConfig{
						RegistryUrl:      "https://ghcr.io",
						ImageName:        "myorg/my-web-app",
						ImageTag:         "latest",
						RegistryUsername: "myuser",
						RegistryPassword: literalRef("mypassword"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with connection_string type PostgreSQL", func() {
				input := minimalSpec()
				input.Spec.ConnectionStrings = []*AzureLinuxWebAppConnectionString{
					{
						Name:  "PgConn",
						Type:  "PostgreSQL",
						Value: literalRef("host=mydb.postgres.database.azure.com;dbname=mydb"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with client_certificate_enabled and mode Required", func() {
				input := minimalSpec()
				input.Spec.ClientCertificateEnabled = boolPtr(true)
				input.Spec.ClientCertificateMode = stringPtr("Required")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with client_certificate_mode Optional", func() {
				input := minimalSpec()
				input.Spec.ClientCertificateMode = stringPtr("Optional")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with client_certificate_mode OptionalInteractiveUser", func() {
				input := minimalSpec()
				input.Spec.ClientCertificateMode = stringPtr("OptionalInteractiveUser")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ftps_state AllAllowed", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.FtpsState = stringPtr("AllAllowed")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ftps_state FtpsOnly", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.FtpsState = stringPtr("FtpsOnly")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with load_balancing_mode WeightedRoundRobin", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.LoadBalancingMode = stringPtr("WeightedRoundRobin")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with minimum_tls_version 1.3", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.MinimumTlsVersion = stringPtr("1.3")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ip_restriction_default_action Deny", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictionDefaultAction = stringPtr("Deny")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with worker_count of 1", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.WorkerCount = int32Ptr(1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with worker_count of 100", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.WorkerCount = int32Ptr(100)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for web app name with hyphens", func() {
				input := minimalSpec()
				input.Spec.Name = "my-web-app-01"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for web app name starting with a number", func() {
				input := minimalSpec()
				input.Spec.Name = "01-web-app"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a two-character name", func() {
				input := minimalSpec()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a 60-character name", func() {
				name := ""
				for len(name) < 60 {
					name += "a"
				}
				input := minimalSpec()
				input.Spec.Name = name
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with key_vault_reference_identity_id", func() {
				input := minimalSpec()
				input.Spec.KeyVaultReferenceIdentityId = literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/kv-id")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with enabled = false", func() {
				input := minimalSpec()
				input.Spec.Enabled = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with client_affinity_enabled = true", func() {
				input := minimalSpec()
				input.Spec.ClientAffinityEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with public_network_access_enabled = false", func() {
				input := minimalSpec()
				input.Spec.PublicNetworkAccessEnabled = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with logs block (application_logs + http_logs + tracing)", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					ApplicationLogs: &AzureLinuxWebAppApplicationLogs{
						FileSystemLevel: stringPtr("Warning"),
					},
					HttpLogs: &AzureLinuxWebAppHttpLogs{
						RetentionInMb:   int32Ptr(35),
						RetentionInDays: int32Ptr(3),
					},
					FailedRequestTracing:  boolPtr(true),
					DetailedErrorMessages: boolPtr(false),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with application_logs file_system_level Information", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					ApplicationLogs: &AzureLinuxWebAppApplicationLogs{
						FileSystemLevel: stringPtr("Information"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with http_logs retention_in_mb 50 and retention_in_days 7", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					HttpLogs: &AzureLinuxWebAppHttpLogs{
						RetentionInMb:   int32Ptr(50),
						RetentionInDays: int32Ptr(7),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_linux_web_app", func() {

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

			ginkgo.It("should return a validation error when name is too short (1 char)", func() {
				input := minimalSpec()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 60 characters", func() {
				tooLong := ""
				for len(tooLong) < 61 {
					tooLong += "a"
				}
				input := minimalSpec()
				input.Spec.Name = tooLong
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains spaces", func() {
				input := minimalSpec()
				input.Spec.Name = "my web app"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains special characters", func() {
				input := minimalSpec()
				input.Spec.Name = "my.web@app"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name ends with a hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "my-web-app-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "-my-web-app"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains underscores", func() {
				input := minimalSpec()
				input.Spec.Name = "my_web_app"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service_plan_id is missing", func() {
				input := minimalSpec()
				input.Spec.ServicePlanId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when site_config is missing", func() {
				input := minimalSpec()
				input.Spec.SiteConfig = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when minimum_tls_version is invalid (wrong format)", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.MinimumTlsVersion = stringPtr("tls1.2")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when minimum_tls_version is completely invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.MinimumTlsVersion = stringPtr("2.0")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ftps_state is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.FtpsState = stringPtr("enabled")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when client_certificate_mode is invalid", func() {
				input := minimalSpec()
				input.Spec.ClientCertificateMode = stringPtr("mandatory")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when identity type is invalid", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: "ManagedIdentity",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when identity type uses wrong casing", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: "systemassigned",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when connection_string type is invalid", func() {
				input := minimalSpec()
				input.Spec.ConnectionStrings = []*AzureLinuxWebAppConnectionString{
					{
						Name:  "MyConn",
						Type:  "InvalidType",
						Value: literalRef("some-value"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when connection_string name is empty", func() {
				input := minimalSpec()
				input.Spec.ConnectionStrings = []*AzureLinuxWebAppConnectionString{
					{
						Name:  "",
						Type:  "Custom",
						Value: literalRef("some-value"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when connection_string value is missing", func() {
				input := minimalSpec()
				input.Spec.ConnectionStrings = []*AzureLinuxWebAppConnectionString{
					{
						Name: "MyConn",
						Type: "Custom",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage mount type is invalid", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "bad-mount",
						Type:        "S3",
						AccountName: "mystore",
						ShareName:   "myshare",
						AccessKey:   literalRef("key=="),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage mount name is empty", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "",
						Type:        "AzureFiles",
						AccountName: "mystore",
						ShareName:   "myshare",
						AccessKey:   literalRef("key=="),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage mount account_name is empty", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "my-mount",
						Type:        "AzureFiles",
						AccountName: "",
						ShareName:   "myshare",
						AccessKey:   literalRef("key=="),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage mount share_name is empty", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "my-mount",
						Type:        "AzureFiles",
						AccountName: "mystore",
						ShareName:   "",
						AccessKey:   literalRef("key=="),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when storage mount access_key is missing", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "my-mount",
						Type:        "AzureFiles",
						AccountName: "mystore",
						ShareName:   "myshare",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when dotnet_version is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					DotnetVersion: "5.0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when node_version is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					NodeVersion: "15-lts",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when python_version is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					PythonVersion: "2.7",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when php_version is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					PhpVersion: "7.3",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ruby_version is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					RubyVersion: "3.0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when go_version is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					GoVersion: "1.20",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when java_version is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					JavaVersion: "14",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when java_server is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					JavaVersion: "17",
					JavaServer:  "WILDFLY",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when Docker config is missing registry_url", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					Docker: &AzureLinuxWebAppDockerConfig{
						ImageName: "myorg/my-web-app",
						ImageTag:  "v1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when Docker config is missing image_name", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					Docker: &AzureLinuxWebAppDockerConfig{
						RegistryUrl: "https://myregistry.azurecr.io",
						ImageTag:    "v1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when Docker config is missing image_tag", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					Docker: &AzureLinuxWebAppDockerConfig{
						RegistryUrl: "https://myregistry.azurecr.io",
						ImageName:   "myorg/my-web-app",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cors allowed_origins is empty", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.Cors = &AzureLinuxWebAppCorsSettings{
					AllowedOrigins: []string{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when worker_count is 0 (below minimum)", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.WorkerCount = int32Ptr(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when worker_count is negative", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.WorkerCount = int32Ptr(-1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when worker_count exceeds 100", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.WorkerCount = int32Ptr(101)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when load_balancing_mode is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.LoadBalancingMode = stringPtr("RoundRobin")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ip_restriction_default_action is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictionDefaultAction = stringPtr("Block")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ip_restriction action is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictions = []*AzureLinuxWebAppIpRestriction{
					{
						Name:      "BadAction",
						Priority:  int32Ptr(100),
						Action:    stringPtr("Block"),
						IpAddress: "10.0.0.0/8",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ip_restriction priority is 0", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictions = []*AzureLinuxWebAppIpRestriction{
					{
						Name:      "BadPriority",
						Priority:  int32Ptr(0),
						Action:    stringPtr("Allow"),
						IpAddress: "10.0.0.0/8",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ip_restriction priority exceeds 65000", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictions = []*AzureLinuxWebAppIpRestriction{
					{
						Name:      "BadPriority",
						Priority:  int32Ptr(65001),
						Action:    stringPtr("Allow"),
						IpAddress: "10.0.0.0/8",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when scm_ip_restriction_default_action is invalid", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ScmIpRestrictionDefaultAction = stringPtr("Reject")
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
				input := &AzureLinuxWebApp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureLinuxWebApp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-webapp",
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

			ginkgo.It("should return a validation error when health_check_eviction_time is below 2", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = int32Ptr(1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health_check_eviction_time is above 10", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = int32Ptr(11)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when http_logs retention_in_mb is below 25", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					HttpLogs: &AzureLinuxWebAppHttpLogs{
						RetentionInMb: int32Ptr(24),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when http_logs retention_in_mb is above 100", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					HttpLogs: &AzureLinuxWebAppHttpLogs{
						RetentionInMb: int32Ptr(101),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when application_logs file_system_level is invalid", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					ApplicationLogs: &AzureLinuxWebAppApplicationLogs{
						FileSystemLevel: stringPtr("Debug"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

package azurelinuxwebappv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
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

// helper to create a pointer to a float64 value
func float64Ptr(f float64) *float64 { return &f }

// helper to create a literal StringValueOrRef
func literalRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: val,
		},
	}
}

// helper to create a StringValueOrRef carrying a value_from reference
func valueFromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// helper to create a minimal valid spec
func minimalSpec() *AzureLinuxWebApp {
	return &AzureLinuxWebApp{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureLinuxWebApp",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-webapp",
		},
		Spec: &AzureLinuxWebAppSpec{
			Region:        "eastus",
			ResourceGroup: literalRef("my-rg"),
			WebAppName:    "my-web-app",
			ServicePlanId: literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Web/serverfarms/plan"),
			SiteConfig:    &AzureLinuxWebAppSiteConfig{},
		},
	}
}

// helper to create a minimal valid auth login block
func minimalAuthLogin() *AzureLinuxWebAppAuthV2Login {
	return &AzureLinuxWebAppAuthV2Login{}
}

var _ = ginkgo.Describe("AzureLinuxWebAppSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_linux_web_app", func() {

			ginkgo.It("should not return a validation error for a minimal spec", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Python application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					PythonVersion: "3.12",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Node.js application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					NodeVersion: "22-lts",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a .NET application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					DotnetVersion: "8.0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a PHP application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					PhpVersion: "8.3",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a complete Java application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					JavaVersion:       "17",
					JavaServer:        AzureLinuxWebAppJavaServer_TOMCAT,
					JavaServerVersion: "10.1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Docker application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					Docker: &AzureLinuxWebAppDockerConfig{
						RegistryUrl:      "https://myregistry.azurecr.io",
						ImageName:        "myorg/my-web-app",
						ImageTag:         "v1.2.3",
						RegistryUsername: "puller",
						RegistryPassword: literalRef("secret-password"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full site config dials", func() {
				input := minimalSpec()
				input.Spec.SiteConfig = &AzureLinuxWebAppSiteConfig{
					AlwaysOn:                     boolPtr(true),
					AppCommandLine:               "gunicorn app:app",
					ApiManagementApiId:           "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ApiManagement/service/apim/apis/api",
					ApiDefinitionUrl:             "https://example.com/openapi.json",
					DefaultDocuments:             []string{"index.html", "default.html"},
					HealthCheckPath:              "/healthz",
					HealthCheckEvictionTimeInMin: int32Ptr(5),
					MinimumTlsVersion:            AzureLinuxWebAppTlsVersion_TLS_1_3,
					ScmMinimumTlsVersion:         AzureLinuxWebAppTlsVersion_TLS_1_2,
					MinimumTlsCipherSuite:        "TLS_AES_256_GCM_SHA384",
					WorkerCount:                  int32Ptr(3),
					Http2Enabled:                 boolPtr(true),
					WebsocketsEnabled:            boolPtr(true),
					Use_32BitWorker:              boolPtr(false),
					FtpsState:                    AzureLinuxWebAppFtpsState_DISABLED,
					LoadBalancingMode:            AzureLinuxWebAppLoadBalancingMode_LEAST_RESPONSE_TIME,
					ManagedPipelineMode:          AzureLinuxWebAppManagedPipelineMode_INTEGRATED,
					LocalMysqlEnabled:            boolPtr(false),
					RemoteDebuggingEnabled:       boolPtr(false),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an auto-heal requests trigger", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AutoHealSetting = &AzureLinuxWebAppAutoHealSetting{
					Trigger: &AzureLinuxWebAppAutoHealTrigger{
						Requests: &AzureLinuxWebAppAutoHealRequestsTrigger{
							Count:    100,
							Interval: "00:05:00",
						},
					},
					MinimumProcessExecutionTime: "00:01:00",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an auto-heal status-code range trigger", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AutoHealSetting = &AzureLinuxWebAppAutoHealSetting{
					Trigger: &AzureLinuxWebAppAutoHealTrigger{
						StatusCodes: []*AzureLinuxWebAppAutoHealStatusCodeTrigger{
							{
								StatusCodeRange: "500-599",
								Count:           10,
								Interval:        "00:10:00",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an auto-heal slow-request trigger", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AutoHealSetting = &AzureLinuxWebAppAutoHealSetting{
					Trigger: &AzureLinuxWebAppAutoHealTrigger{
						SlowRequest: &AzureLinuxWebAppAutoHealSlowRequestTrigger{
							TimeTaken: "00:00:10",
							Interval:  "00:05:00",
							Count:     20,
						},
						SlowRequestWithPath: []*AzureLinuxWebAppAutoHealSlowRequestWithPathTrigger{
							{
								TimeTaken: "00:00:05",
								Interval:  "00:05:00",
								Count:     5,
								Path:      "/api/slow",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for IP restrictions with headers", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictions = []*AzureLinuxWebAppIpRestriction{
					{
						Name:       "front-door-only",
						Priority:   int32Ptr(100),
						Action:     AzureLinuxWebAppIpRestrictionAction_ALLOW,
						ServiceTag: "AzureFrontDoor.Backend",
						Headers: &AzureLinuxWebAppIpRestrictionHeaders{
							// One literal GUID and one reference (the
							// origin-lockdown seam resolving an
							// AzureFrontDoorProfile's resource_guid).
							XAzureFdid: []*foreignkeyv1.StringValueOrRef{
								literalRef("11111111-2222-3333-4444-555555555555"),
								valueFromRef("my-front-door-profile"),
							},
						},
					},
					{
						Name:      "office",
						Priority:  int32Ptr(200),
						Action:    AzureLinuxWebAppIpRestrictionAction_ALLOW,
						IpAddress: "203.0.113.0/24",
					},
				}
				input.Spec.SiteConfig.IpRestrictionDefaultAction = AzureLinuxWebAppIpRestrictionAction_DENY
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CORS without wildcard credentials", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.Cors = &AzureLinuxWebAppCorsSettings{
					AllowedOrigins:     []string{"https://myapp.example.com"},
					SupportCredentials: boolPtr(true),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a wildcard CORS origin without credentials", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.Cors = &AzureLinuxWebAppCorsSettings{
					AllowedOrigins: []string{"*"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for storage mounts", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "shared-data",
						Type:        AzureLinuxWebAppStorageMountType_AZURE_FILES,
						AccountName: "mystorageacct",
						ShareName:   "data",
						AccessKey:   literalRef("storage-key"),
						MountPath:   "/mnt/data",
					},
					{
						Name:        "static-assets",
						Type:        AzureLinuxWebAppStorageMountType_AZURE_BLOB,
						AccountName: "mystorageacct",
						ShareName:   "assets",
						AccessKey:   literalRef("storage-key"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for file-system logs", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					ApplicationLogs: &AzureLinuxWebAppApplicationLogs{
						FileSystemLevel: AzureLinuxWebAppLogLevel_INFORMATION,
					},
					HttpLogs: &AzureLinuxWebAppHttpLogs{
						FileSystem: &AzureLinuxWebAppHttpLogsFileSystem{
							RetentionInMb:   int32Ptr(50),
							RetentionInDays: int32Ptr(7),
						},
					},
					FailedRequestTracing:  boolPtr(true),
					DetailedErrorMessages: boolPtr(false),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for blob-storage logs", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					ApplicationLogs: &AzureLinuxWebAppApplicationLogs{
						FileSystemLevel: AzureLinuxWebAppLogLevel_ERROR,
						AzureBlobStorage: &AzureLinuxWebAppBlobStorageLogs{
							Level:           AzureLinuxWebAppLogLevel_WARNING,
							SasUrl:          literalRef("https://logs.blob.core.windows.net/applogs?sig=abc"),
							RetentionInDays: int32Ptr(30),
						},
					},
					HttpLogs: &AzureLinuxWebAppHttpLogs{
						AzureBlobStorage: &AzureLinuxWebAppHttpLogsBlobStorage{
							SasUrl:          literalRef("https://logs.blob.core.windows.net/httplogs?sig=abc"),
							RetentionInDays: int32Ptr(30),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a daily backup", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureLinuxWebAppBackup{
					Name:              "nightly",
					StorageAccountUrl: literalRef("https://backups.blob.core.windows.net/webapp?sig=abc"),
					Schedule: &AzureLinuxWebAppBackupSchedule{
						FrequencyInterval:    1,
						FrequencyUnit:        AzureLinuxWebAppBackupFrequencyUnit_DAY,
						KeepAtLeastOneBackup: boolPtr(true),
						RetentionPeriodDays:  int32Ptr(30),
						StartTime:            "2026-01-01T00:00:00Z",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for hourly backups", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureLinuxWebAppBackup{
					Name:              "hourly",
					StorageAccountUrl: literalRef("https://backups.blob.core.windows.net/webapp?sig=abc"),
					Enabled:           boolPtr(true),
					Schedule: &AzureLinuxWebAppBackupSchedule{
						FrequencyInterval: 12,
						FrequencyUnit:     AzureLinuxWebAppBackupFrequencyUnit_HOUR,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for sticky app settings", func() {
				input := minimalSpec()
				input.Spec.StickySettings = &AzureLinuxWebAppStickySettings{
					AppSettingNames: []string{"STAGING_ONLY_FLAG"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for sticky connection strings", func() {
				input := minimalSpec()
				input.Spec.StickySettings = &AzureLinuxWebAppStickySettings{
					ConnectionStringNames: []string{"Database"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for minimal Easy Auth v2", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login:       minimalAuthLogin(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Entra ID auth with a secret setting name", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled:           boolPtr(true),
					RequireAuthentication: boolPtr(true),
					UnauthenticatedAction: AzureLinuxWebAppUnauthenticatedAction_RETURN_401,
					Login:                 minimalAuthLogin(),
					ActiveDirectoryV2: &AzureLinuxWebAppAuthV2ActiveDirectory{
						ClientId:                "11111111-2222-3333-4444-555555555555",
						TenantAuthEndpoint:      "https://login.microsoftonline.com/v2.0/99999999-8888-7777-6666-555555555555/",
						ClientSecretSettingName: "AAD_CLIENT_SECRET",
						AllowedAudiences:        []string{"api://my-app"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Entra ID auth with a certificate thumbprint", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login:       minimalAuthLogin(),
					ActiveDirectoryV2: &AzureLinuxWebAppAuthV2ActiveDirectory{
						ClientId:                          "11111111-2222-3333-4444-555555555555",
						TenantAuthEndpoint:                "https://login.microsoftonline.com/v2.0/99999999-8888-7777-6666-555555555555/",
						ClientSecretCertificateThumbprint: "AB12CD34EF56",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a custom OIDC provider", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled:     boolPtr(true),
					DefaultProvider: "corp-idp",
					Login:           minimalAuthLogin(),
					CustomOidcV2: []*AzureLinuxWebAppAuthV2CustomOidc{
						{
							Name:                        "corp-idp",
							ClientId:                    "web-app-client",
							OpenidConfigurationEndpoint: "https://idp.example.com/.well-known/openid-configuration",
							Scopes:                      []string{"openid", "profile"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for social login providers", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login:       minimalAuthLogin(),
					GithubV2: &AzureLinuxWebAppAuthV2Github{
						ClientId:                "gh-client",
						ClientSecretSettingName: "GITHUB_SECRET",
					},
					GoogleV2: &AzureLinuxWebAppAuthV2Google{
						ClientId:                "google-client",
						ClientSecretSettingName: "GOOGLE_SECRET",
					},
					MicrosoftV2: &AzureLinuxWebAppAuthV2Microsoft{
						ClientId:                "ms-client",
						ClientSecretSettingName: "MS_SECRET",
					},
					FacebookV2: &AzureLinuxWebAppAuthV2Facebook{
						AppId:                "fb-app",
						AppSecretSettingName: "FB_SECRET",
					},
					AppleV2: &AzureLinuxWebAppAuthV2Apple{
						ClientId:                "apple-client",
						ClientSecretSettingName: "APPLE_SECRET",
					},
					TwitterV2: &AzureLinuxWebAppAuthV2Twitter{
						ConsumerKey:               "tw-key",
						ConsumerSecretSettingName: "TW_SECRET",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a custom forward proxy with header names", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled:                        boolPtr(true),
					ForwardProxyConvention:             AzureLinuxWebAppForwardProxyConvention_FORWARD_PROXY_CUSTOM,
					ForwardProxyCustomHostHeaderName:   "X-Original-Host",
					ForwardProxyCustomSchemeHeaderName: "X-Original-Scheme",
					Login:                              minimalAuthLogin(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a full login block", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login: &AzureLinuxWebAppAuthV2Login{
						LogoutEndpoint:              "/logout",
						TokenStoreEnabled:           boolPtr(true),
						TokenRefreshExtensionTime:   float64Ptr(48),
						CookieExpirationConvention:  AzureLinuxWebAppCookieExpirationConvention_FIXED_TIME,
						CookieExpirationTime:        stringPtr("12:00:00"),
						ValidateNonce:               boolPtr(true),
						NonceExpirationTime:         stringPtr("00:10:00"),
						AllowedExternalRedirectUrls: []string{"https://myapp.example.com/done"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a system-assigned identity", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: AzureLinuxWebAppIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a user-assigned identity with IDs", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: AzureLinuxWebAppIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{
						literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-identity"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for both identity types with IDs", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: AzureLinuxWebAppIdentityType_SYSTEM_AND_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{
						literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-identity"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for connection strings", func() {
				input := minimalSpec()
				input.Spec.ConnectionStrings = []*AzureLinuxWebAppConnectionString{
					{
						Name:  "Database",
						Type:  AzureLinuxWebAppConnectionStringType_POSTGRESQL,
						Value: literalRef("Host=db.example.com;Database=app"),
					},
					{
						Name:  "Cache",
						Type:  AzureLinuxWebAppConnectionStringType_REDIS_CACHE,
						Value: literalRef("cache.example.com:6380,ssl=true"),
					},
					{
						Name:  "Custom",
						Type:  AzureLinuxWebAppConnectionStringType_CUSTOM,
						Value: literalRef("custom-value"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for VNet toggles with a subnet", func() {
				input := minimalSpec()
				input.Spec.VirtualNetworkSubnetId = literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/apps")
				input.Spec.VnetImagePullEnabled = boolPtr(true)
				input.Spec.VirtualNetworkBackupRestoreEnabled = boolPtr(true)
				input.Spec.SiteConfig.VnetRouteAllEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with basic-auth publishing disabled", func() {
				input := minimalSpec()
				input.Spec.WebdeployPublishBasicAuthenticationEnabled = boolPtr(false)
				input.Spec.FtpPublishBasicAuthenticationEnabled = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a zip deploy file and client cert settings", func() {
				input := minimalSpec()
				input.Spec.ZipDeployFile = "./build/app.zip"
				input.Spec.ClientCertificateEnabled = boolPtr(true)
				input.Spec.ClientCertificateMode = AzureLinuxWebAppClientCertificateMode_REQUIRED
				input.Spec.ClientCertificateExclusionPaths = "/api/health;/api/status"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with user tags", func() {
				input := minimalSpec()
				input.Spec.Tags = map[string]string{"cost-center": "platform", "owner": "web-team"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom references", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureResourceGroup,
							Name:      "shared-rg",
							FieldPath: "status.outputs.resource_group_name",
						},
					},
				}
				input.Spec.ServicePlanId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureServicePlan,
							Name:      "shared-plan",
							FieldPath: "status.outputs.service_plan_id",
						},
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

			ginkgo.It("should return a validation error when web_app_name is missing", func() {
				input := minimalSpec()
				input.Spec.WebAppName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when web_app_name is one character", func() {
				input := minimalSpec()
				input.Spec.WebAppName = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when web_app_name starts with a hyphen", func() {
				input := minimalSpec()
				input.Spec.WebAppName = "-my-app"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when web_app_name contains invalid characters", func() {
				input := minimalSpec()
				input.Spec.WebAppName = "my_app!"
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

			ginkgo.It("should return a validation error for health check eviction without a path", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = int32Ptr(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for health check eviction out of range", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.HealthCheckPath = "/healthz"
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = int32Ptr(11)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for worker_count of zero", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.WorkerCount = int32Ptr(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for worker_count above 100", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.WorkerCount = int32Ptr(101)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid dotnet version", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					DotnetVersion: "4.8",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid cipher suite", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.MinimumTlsCipherSuite = "TLS_FAKE_SUITE"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for java_version without java_server", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					JavaVersion: "17",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for java_server without java_version", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureLinuxWebAppApplicationStack{
					JavaServer:        AzureLinuxWebAppJavaServer_TOMCAT,
					JavaServerVersion: "10.1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for CORS with empty origins", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.Cors = &AzureLinuxWebAppCorsSettings{
					AllowedOrigins: []string{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for wildcard CORS with credentials", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.Cors = &AzureLinuxWebAppCorsSettings{
					AllowedOrigins:     []string{"*"},
					SupportCredentials: boolPtr(true),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for empty sticky settings", func() {
				input := minimalSpec()
				input.Spec.StickySettings = &AzureLinuxWebAppStickySettings{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a backup without a schedule", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureLinuxWebAppBackup{
					Name:              "nightly",
					StorageAccountUrl: literalRef("https://backups.blob.core.windows.net/webapp?sig=abc"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a backup frequency interval of zero", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureLinuxWebAppBackup{
					Name:              "nightly",
					StorageAccountUrl: literalRef("https://backups.blob.core.windows.net/webapp?sig=abc"),
					Schedule: &AzureLinuxWebAppBackupSchedule{
						FrequencyInterval: 0,
						FrequencyUnit:     AzureLinuxWebAppBackupFrequencyUnit_DAY,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a backup without a frequency unit", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureLinuxWebAppBackup{
					Name:              "nightly",
					StorageAccountUrl: literalRef("https://backups.blob.core.windows.net/webapp?sig=abc"),
					Schedule: &AzureLinuxWebAppBackupSchedule{
						FrequencyInterval: 1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for http_logs with both destinations", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					HttpLogs: &AzureLinuxWebAppHttpLogs{
						FileSystem: &AzureLinuxWebAppHttpLogsFileSystem{},
						AzureBlobStorage: &AzureLinuxWebAppHttpLogsBlobStorage{
							SasUrl: literalRef("https://logs.blob.core.windows.net/httplogs?sig=abc"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for http_logs with no destination", func() {
				input := minimalSpec()
				input.Spec.Logs = &AzureLinuxWebAppLogs{
					HttpLogs: &AzureLinuxWebAppHttpLogs{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an auto-heal setting with an empty trigger", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AutoHealSetting = &AzureLinuxWebAppAutoHealSetting{
					Trigger: &AzureLinuxWebAppAutoHealTrigger{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an auto-heal requests trigger with zero count", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AutoHealSetting = &AzureLinuxWebAppAutoHealSetting{
					Trigger: &AzureLinuxWebAppAutoHealTrigger{
						Requests: &AzureLinuxWebAppAutoHealRequestsTrigger{
							Count:    0,
							Interval: "00:05:00",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed auto-heal interval", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AutoHealSetting = &AzureLinuxWebAppAutoHealSetting{
					Trigger: &AzureLinuxWebAppAutoHealTrigger{
						Requests: &AzureLinuxWebAppAutoHealRequestsTrigger{
							Count:    100,
							Interval: "5m",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed status code range", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AutoHealSetting = &AzureLinuxWebAppAutoHealSetting{
					Trigger: &AzureLinuxWebAppAutoHealTrigger{
						StatusCodes: []*AzureLinuxWebAppAutoHealStatusCodeTrigger{
							{
								StatusCodeRange: "5xx",
								Count:           10,
								Interval:        "00:10:00",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for auth settings without a login block", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for Entra ID auth with both credential forms", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login:       minimalAuthLogin(),
					ActiveDirectoryV2: &AzureLinuxWebAppAuthV2ActiveDirectory{
						ClientId:                          "11111111-2222-3333-4444-555555555555",
						TenantAuthEndpoint:                "https://login.microsoftonline.com/v2.0/99999999-8888-7777-6666-555555555555/",
						ClientSecretSettingName:           "AAD_CLIENT_SECRET",
						ClientSecretCertificateThumbprint: "AB12CD34EF56",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for custom proxy headers without the custom convention", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled:                      boolPtr(true),
					ForwardProxyCustomHostHeaderName: "X-Original-Host",
					Login:                            minimalAuthLogin(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a login block with both token store backings", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login: &AzureLinuxWebAppAuthV2Login{
						TokenStorePath:           "/tokens",
						TokenStoreSasSettingName: "TOKEN_STORE_SAS",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed cookie expiration time", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureLinuxWebAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login: &AzureLinuxWebAppAuthV2Login{
						CookieExpirationTime: stringPtr("8h"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a user-assigned identity without IDs", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: AzureLinuxWebAppIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a system-assigned identity with IDs", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{
					Type: AzureLinuxWebAppIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{
						literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-identity"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an identity without a type", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureLinuxWebAppIdentity{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for vnet_image_pull_enabled without a subnet", func() {
				input := minimalSpec()
				input.Spec.VnetImagePullEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for virtual_network_backup_restore_enabled without a subnet", func() {
				input := minimalSpec()
				input.Spec.VirtualNetworkBackupRestoreEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for vnet_route_all_enabled without a subnet", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.VnetRouteAllEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an IP restriction priority of zero", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictions = []*AzureLinuxWebAppIpRestriction{
					{
						IpAddress: "203.0.113.0/24",
						Priority:  int32Ptr(0),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for too many forwarded-for headers", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictions = []*AzureLinuxWebAppIpRestriction{
					{
						IpAddress: "203.0.113.0/24",
						Headers: &AzureLinuxWebAppIpRestrictionHeaders{
							XForwardedFor: []string{"1.1.1.1", "2.2.2.2", "3.3.3.3", "4.4.4.4", "5.5.5.5", "6.6.6.6", "7.7.7.7", "8.8.8.8", "9.9.9.9"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a storage mount without a type", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureLinuxWebAppStorageMount{
					{
						Name:        "shared-data",
						AccountName: "mystorageacct",
						ShareName:   "data",
						AccessKey:   literalRef("storage-key"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a connection string without a type", func() {
				input := minimalSpec()
				input.Spec.ConnectionStrings = []*AzureLinuxWebAppConnectionString{
					{
						Name:  "Database",
						Value: literalRef("Host=db.example.com"),
					},
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
				input := minimalSpec()
				input.Spec = nil
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

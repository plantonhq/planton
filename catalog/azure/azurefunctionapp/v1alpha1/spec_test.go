package azurefunctionappv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureFunctionAppSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFunctionAppSpec Validation Tests")
}

// helper to create a pointer to a bool value
func boolPtr(b bool) *bool { return &b }

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

// helper to create a StringValueOrRef carrying a value_from reference
func valueFromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// helper to create a minimal valid spec (account-name storage binding)
func minimalSpec() *AzureFunctionApp {
	return &AzureFunctionApp{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFunctionApp",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-fn",
		},
		Spec: &AzureFunctionAppSpec{
			Region:             "eastus",
			ResourceGroup:      literalRef("my-rg"),
			FunctionAppName:    "my-function-app",
			ServicePlanId:      literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Web/serverfarms/plan"),
			StorageAccountName: literalRef("mystorageacct"),
			SiteConfig:         &AzureFunctionAppSiteConfig{},
		},
	}
}

// helper to create a minimal valid auth login block
func minimalAuthLogin() *AzureFunctionAppAuthV2Login {
	return &AzureFunctionAppAuthV2Login{}
}

var _ = ginkgo.Describe("AzureFunctionAppSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_function_app", func() {

			ginkgo.It("should not return a validation error for a minimal spec", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Python application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureFunctionAppApplicationStack{
					PythonVersion: "3.12",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a .NET isolated application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureFunctionAppApplicationStack{
					DotnetVersion:            "8.0",
					UseDotnetIsolatedRuntime: boolPtr(true),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a Docker application stack", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureFunctionAppApplicationStack{
					Docker: &AzureFunctionAppDockerConfig{
						RegistryUrl:      "https://myregistry.azurecr.io",
						ImageName:        "myorg/my-function-app",
						ImageTag:         "v1.2.3",
						RegistryUsername: "puller",
						RegistryPassword: literalRef("secret-password"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a custom runtime", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureFunctionAppApplicationStack{
					UseCustomRuntime: boolPtr(true),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for the Key Vault storage binding", func() {
				input := minimalSpec()
				input.Spec.StorageAccountName = nil
				input.Spec.StorageKeyVaultSecretId = "https://my-vault.vault.azure.net/secrets/storage-connection"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a versioned Key Vault storage binding", func() {
				input := minimalSpec()
				input.Spec.StorageAccountName = nil
				input.Spec.StorageKeyVaultSecretId = "https://my-vault.vault.azure.net/secrets/storage-connection/0123456789abcdef"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for managed-identity storage access", func() {
				input := minimalSpec()
				input.Spec.StorageUsesManagedIdentity = boolPtr(true)
				input.Spec.Identity = &AzureFunctionAppIdentity{
					Type: AzureFunctionAppIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for an access-key storage binding", func() {
				input := minimalSpec()
				input.Spec.StorageAccountAccessKey = literalRef("storage-key")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with serverless scaling dials", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AppScaleLimit = int32Ptr(50)
				input.Spec.SiteConfig.ElasticInstanceMinimum = int32Ptr(2)
				input.Spec.SiteConfig.PreWarmedInstanceCount = int32Ptr(3)
				input.Spec.SiteConfig.RuntimeScaleMonitoringEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a daily memory quota", func() {
				input := minimalSpec()
				input.Spec.DailyMemoryTimeQuota = int32Ptr(500000)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full site config dials", func() {
				input := minimalSpec()
				input.Spec.SiteConfig = &AzureFunctionAppSiteConfig{
					AlwaysOn:                     boolPtr(true),
					AppCommandLine:               "func host start",
					ApiManagementApiId:           "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ApiManagement/service/apim/apis/api",
					ApiDefinitionUrl:             "https://example.com/openapi.json",
					DefaultDocuments:             []string{"index.html"},
					HealthCheckPath:              "/api/health",
					HealthCheckEvictionTimeInMin: int32Ptr(5),
					MinimumTlsVersion:            AzureFunctionAppTlsVersion_TLS_1_3,
					ScmMinimumTlsVersion:         AzureFunctionAppTlsVersion_TLS_1_2,
					MinimumTlsCipherSuite:        "TLS_AES_256_GCM_SHA384",
					WorkerCount:                  int32Ptr(3),
					Http2Enabled:                 boolPtr(true),
					WebsocketsEnabled:            boolPtr(false),
					Use_32BitWorker:              boolPtr(false),
					FtpsState:                    AzureFunctionAppFtpsState_DISABLED,
					LoadBalancingMode:            AzureFunctionAppLoadBalancingMode_LEAST_REQUESTS,
					ManagedPipelineMode:          AzureFunctionAppManagedPipelineMode_INTEGRATED,
					RemoteDebuggingEnabled:       boolPtr(false),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for IP restrictions with a deny default", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.IpRestrictions = []*AzureFunctionAppIpRestriction{
					{
						Name:       "front-door-only",
						Priority:   int32Ptr(100),
						Action:     AzureFunctionAppIpRestrictionAction_ALLOW,
						ServiceTag: "AzureFrontDoor.Backend",
						Headers: &AzureFunctionAppIpRestrictionHeaders{
							// One literal GUID and one reference (the
							// origin-lockdown seam resolving an
							// AzureFrontDoorProfile's resource_guid).
							XAzureFdid: []*foreignkeyv1.StringValueOrRef{
								literalRef("11111111-2222-3333-4444-555555555555"),
								valueFromRef("my-front-door-profile"),
							},
						},
					},
				}
				input.Spec.SiteConfig.IpRestrictionDefaultAction = AzureFunctionAppIpRestrictionAction_DENY
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for app service logs", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AppServiceLogs = &AzureFunctionAppAppServiceLogs{
					DiskQuotaMb:         int32Ptr(50),
					RetentionPeriodDays: int32Ptr(7),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for storage mounts", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureFunctionAppStorageMount{
					{
						Name:        "shared-data",
						Type:        AzureFunctionAppStorageMountType_AZURE_FILES,
						AccountName: "mystorageacct",
						ShareName:   "data",
						AccessKey:   literalRef("storage-key"),
						MountPath:   "/mnt/data",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a daily backup", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureFunctionAppBackup{
					Name:              "nightly",
					StorageAccountUrl: literalRef("https://backups.blob.core.windows.net/fnapp?sig=abc"),
					Schedule: &AzureFunctionAppBackupSchedule{
						FrequencyInterval:    1,
						FrequencyUnit:        AzureFunctionAppBackupFrequencyUnit_DAY,
						KeepAtLeastOneBackup: boolPtr(true),
						RetentionPeriodDays:  int32Ptr(30),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for sticky settings", func() {
				input := minimalSpec()
				input.Spec.StickySettings = &AzureFunctionAppStickySettings{
					AppSettingNames: []string{"SLOT_MARKER"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Entra ID Easy Auth", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppAuthSettingsV2{
					AuthEnabled:           boolPtr(true),
					RequireAuthentication: boolPtr(true),
					UnauthenticatedAction: AzureFunctionAppUnauthenticatedAction_RETURN_401,
					Login:                 minimalAuthLogin(),
					ActiveDirectoryV2: &AzureFunctionAppAuthV2ActiveDirectory{
						ClientId:                "11111111-2222-3333-4444-555555555555",
						TenantAuthEndpoint:      "https://login.microsoftonline.com/v2.0/99999999-8888-7777-6666-555555555555/",
						ClientSecretSettingName: "AAD_CLIENT_SECRET",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a custom OIDC provider", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login:       minimalAuthLogin(),
					CustomOidcV2: []*AzureFunctionAppAuthV2CustomOidc{
						{
							Name:                        "corp-idp",
							ClientId:                    "fn-client",
							OpenidConfigurationEndpoint: "https://idp.example.com/.well-known/openid-configuration",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for VNet toggles with a subnet", func() {
				input := minimalSpec()
				input.Spec.VirtualNetworkSubnetId = literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/fns")
				input.Spec.VnetImagePullEnabled = boolPtr(true)
				input.Spec.VirtualNetworkBackupRestoreEnabled = boolPtr(true)
				input.Spec.SiteConfig.VnetRouteAllEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with hardened publishing and user tags", func() {
				input := minimalSpec()
				input.Spec.WebdeployPublishBasicAuthenticationEnabled = boolPtr(false)
				input.Spec.FtpPublishBasicAuthenticationEnabled = boolPtr(false)
				input.Spec.ClientCertificateEnabled = boolPtr(true)
				input.Spec.ClientCertificateMode = AzureFunctionAppClientCertificateMode_REQUIRED
				input.Spec.Tags = map[string]string{"cost-center": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for identity variants", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureFunctionAppIdentity{
					Type: AzureFunctionAppIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{
						literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/fn-identity"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom references", func() {
				input := minimalSpec()
				input.Spec.ServicePlanId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureServicePlan,
							Name:      "fn-plan",
							FieldPath: "status.outputs.service_plan_id",
						},
					},
				}
				input.Spec.StorageAccountName = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureStorageAccount,
							Name:      "fn-storage",
							FieldPath: "status.outputs.storage_account_name",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for connection strings", func() {
				input := minimalSpec()
				input.Spec.ConnectionStrings = []*AzureFunctionAppConnectionString{
					{
						Name:  "ServiceBus",
						Type:  AzureFunctionAppConnectionStringType_SERVICE_BUS,
						Value: literalRef("Endpoint=sb://ns.servicebus.windows.net/..."),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_function_app", func() {

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

			ginkgo.It("should return a validation error when function_app_name is missing", func() {
				input := minimalSpec()
				input.Spec.FunctionAppName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when function_app_name starts with a hyphen", func() {
				input := minimalSpec()
				input.Spec.FunctionAppName = "-my-fn"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when service_plan_id is missing", func() {
				input := minimalSpec()
				input.Spec.ServicePlanId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when no storage binding is set", func() {
				input := minimalSpec()
				input.Spec.StorageAccountName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both storage bindings are set", func() {
				input := minimalSpec()
				input.Spec.StorageKeyVaultSecretId = "https://my-vault.vault.azure.net/secrets/storage-connection"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed Key Vault secret ID", func() {
				input := minimalSpec()
				input.Spec.StorageAccountName = nil
				input.Spec.StorageKeyVaultSecretId = "my-vault-secret"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the access key and managed identity are combined", func() {
				input := minimalSpec()
				input.Spec.StorageAccountAccessKey = literalRef("storage-key")
				input.Spec.StorageUsesManagedIdentity = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when site_config is missing", func() {
				input := minimalSpec()
				input.Spec.SiteConfig = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a negative daily memory quota", func() {
				input := minimalSpec()
				input.Spec.DailyMemoryTimeQuota = int32Ptr(-1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for health check eviction without a path", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = int32Ptr(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid python version", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.ApplicationStack = &AzureFunctionAppApplicationStack{
					PythonVersion: "2.7",
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

			ginkgo.It("should return a validation error for wildcard CORS with credentials", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.Cors = &AzureFunctionAppCorsSettings{
					AllowedOrigins:     []string{"*"},
					SupportCredentials: boolPtr(true),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for empty sticky settings", func() {
				input := minimalSpec()
				input.Spec.StickySettings = &AzureFunctionAppStickySettings{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a backup without a schedule", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureFunctionAppBackup{
					Name:              "nightly",
					StorageAccountUrl: literalRef("https://backups.blob.core.windows.net/fnapp?sig=abc"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for auth settings without a login block", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for Entra ID auth with both credential forms", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login:       minimalAuthLogin(),
					ActiveDirectoryV2: &AzureFunctionAppAuthV2ActiveDirectory{
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
				input.Spec.AuthSettingsV2 = &AzureFunctionAppAuthSettingsV2{
					AuthEnabled:                      boolPtr(true),
					ForwardProxyCustomHostHeaderName: "X-Original-Host",
					Login:                            minimalAuthLogin(),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a user-assigned identity without IDs", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureFunctionAppIdentity{
					Type: AzureFunctionAppIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a system-assigned identity with IDs", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureFunctionAppIdentity{
					Type: AzureFunctionAppIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{
						literalRef("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/fn-identity"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for vnet_image_pull_enabled without a subnet", func() {
				input := minimalSpec()
				input.Spec.VnetImagePullEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for vnet_route_all_enabled without a subnet", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.VnetRouteAllEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a storage mount without a type", func() {
				input := minimalSpec()
				input.Spec.StorageMounts = []*AzureFunctionAppStorageMount{
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

			ginkgo.It("should return a validation error for an app service logs quota out of range", func() {
				input := minimalSpec()
				input.Spec.SiteConfig.AppServiceLogs = &AzureFunctionAppAppServiceLogs{
					DiskQuotaMb: int32Ptr(200),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for virtual_network_backup_restore_enabled without a subnet", func() {
				input := minimalSpec()
				input.Spec.VirtualNetworkBackupRestoreEnabled = boolPtr(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a login block with both token store backings", func() {
				input := minimalSpec()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login: &AzureFunctionAppAuthV2Login{
						TokenStorePath:           "/home/tokens",
						TokenStoreSasSettingName: "TOKEN_STORE_SAS",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed cookie expiration time", func() {
				input := minimalSpec()
				badTime := "8 hours"
				input.Spec.AuthSettingsV2 = &AzureFunctionAppAuthSettingsV2{
					AuthEnabled: boolPtr(true),
					Login: &AzureFunctionAppAuthV2Login{
						CookieExpirationTime: &badTime,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a backup frequency interval of zero", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureFunctionAppBackup{
					Name:              "nightly",
					StorageAccountUrl: literalRef("https://backups.blob.core.windows.net/fnapp?sig=abc"),
					Schedule: &AzureFunctionAppBackupSchedule{
						FrequencyInterval: 0,
						FrequencyUnit:     AzureFunctionAppBackupFrequencyUnit_DAY,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a backup without a frequency unit", func() {
				input := minimalSpec()
				input.Spec.Backup = &AzureFunctionAppBackup{
					Name:              "nightly",
					StorageAccountUrl: literalRef("https://backups.blob.core.windows.net/fnapp?sig=abc"),
					Schedule: &AzureFunctionAppBackupSchedule{
						FrequencyInterval: 1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an identity without a type", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureFunctionAppIdentity{}
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

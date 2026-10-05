package azurefunctionappflexconsumptionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureFunctionAppFlexConsumptionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFunctionAppFlexConsumptionSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testPlanId     = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Web/serverFarms/flex-plan"
	testSubnetId   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/flex-subnet"
	testIdentityId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-mi"
	testEndpoint   = "https://appstorage.blob.core.windows.net/deployments"
)

// validResource returns a minimal valid flex consumption app that
// individual cases mutate into the shape under test.
func validResource() *AzureFunctionAppFlexConsumption {
	return &AzureFunctionAppFlexConsumption{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFunctionAppFlexConsumption",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-flex-app",
		},
		Spec: &AzureFunctionAppFlexConsumptionSpec{
			Region:                    "eastus",
			ResourceGroup:             literal("app-rg"),
			FunctionAppName:           "acme-flex-app",
			ServicePlanId:             literal(testPlanId),
			StorageContainerEndpoint:  testEndpoint,
			StorageAuthenticationType: AzureFunctionAppFlexConsumptionStorageAuthenticationType_STORAGE_ACCOUNT_CONNECTION_STRING,
			StorageAccessKey:          literal("c3VwZXItc2VjcmV0LWtleQ=="),
			RuntimeName:               AzureFunctionAppFlexConsumptionRuntimeName_NODE,
			RuntimeVersion:            "20",
			SiteConfig:                &AzureFunctionAppFlexConsumptionSiteConfig{},
		},
	}
}

var _ = ginkgo.Describe("AzureFunctionAppFlexConsumptionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_function_app_flex_consumption", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				gomega.Expect(protovalidate.Validate(validResource())).To(gomega.BeNil())
			})

			ginkgo.It("should accept every runtime", func() {
				input := validResource()
				for _, runtime := range []AzureFunctionAppFlexConsumptionRuntimeName{
					AzureFunctionAppFlexConsumptionRuntimeName_NODE,
					AzureFunctionAppFlexConsumptionRuntimeName_DOTNET_ISOLATED,
					AzureFunctionAppFlexConsumptionRuntimeName_JAVA,
					AzureFunctionAppFlexConsumptionRuntimeName_POWERSHELL,
					AzureFunctionAppFlexConsumptionRuntimeName_PYTHON,
					AzureFunctionAppFlexConsumptionRuntimeName_CUSTOM_HANDLER,
				} {
					input.Spec.RuntimeName = runtime
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "expected runtime %v to be accepted", runtime)
				}
			})

			ginkgo.It("should accept system-assigned identity storage auth without a key", func() {
				input := validResource()
				input.Spec.StorageAuthenticationType = AzureFunctionAppFlexConsumptionStorageAuthenticationType_SYSTEM_ASSIGNED_IDENTITY
				input.Spec.StorageAccessKey = nil
				input.Spec.Identity = &AzureFunctionAppFlexConsumptionIdentity{
					Type: AzureFunctionAppFlexConsumptionIdentityType_SYSTEM_ASSIGNED,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept user-assigned identity storage auth with the identity wired", func() {
				input := validResource()
				input.Spec.StorageAuthenticationType = AzureFunctionAppFlexConsumptionStorageAuthenticationType_USER_ASSIGNED_IDENTITY
				input.Spec.StorageAccessKey = nil
				input.Spec.StorageUserAssignedIdentityId = literal(testIdentityId)
				input.Spec.Identity = &AzureFunctionAppFlexConsumptionIdentity{
					Type:        AzureFunctionAppFlexConsumptionIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept the scale dials at their bounds", func() {
				input := validResource()
				input.Spec.MaximumInstanceCount = proto.Int32(1)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.MaximumInstanceCount = proto.Int32(1000)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.HttpConcurrency = proto.Int32(1)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.HttpConcurrency = proto.Int32(1000)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.InstanceMemoryInMb = proto.Int32(4096)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept always-ready pools", func() {
				input := validResource()
				input.Spec.AlwaysReady = []*AzureFunctionAppFlexConsumptionAlwaysReady{
					{Name: "http", InstanceCount: proto.Int32(2)},
					{Name: "function:ProcessOrders", InstanceCount: proto.Int32(1)},
					{Name: "durable"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fully-dialed site config", func() {
				input := validResource()
				input.Spec.VirtualNetworkSubnetId = literal(testSubnetId)
				input.Spec.SiteConfig = &AzureFunctionAppFlexConsumptionSiteConfig{
					ApiManagementApiId:            "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ApiManagement/service/apim/apis/api1",
					ApiDefinitionUrl:              "https://example.com/openapi.json",
					AppCommandLine:                "node server.js",
					ApplicationInsightsKey:        "00000000-0000-0000-0000-000000000000",
					AppServiceLogs:                &AzureFunctionAppFlexConsumptionAppServiceLogs{DiskQuotaMb: proto.Int32(50), RetentionPeriodDays: proto.Int32(7)},
					DefaultDocuments:              []string{"index.html"},
					ElasticInstanceMinimum:        proto.Int32(1),
					Http2Enabled:                  proto.Bool(true),
					HealthCheckPath:               "/api/health",
					HealthCheckEvictionTimeInMin:  proto.Int32(5),
					WorkerCount:                   proto.Int32(4),
					MinimumTlsVersion:             AzureFunctionAppFlexConsumptionTlsVersion_TLS_1_3,
					ScmMinimumTlsVersion:          AzureFunctionAppFlexConsumptionTlsVersion_TLS_1_2,
					LoadBalancingMode:             AzureFunctionAppFlexConsumptionLoadBalancingMode_WEIGHTED_ROUND_ROBIN,
					ManagedPipelineMode:           AzureFunctionAppFlexConsumptionManagedPipelineMode_INTEGRATED,
					RemoteDebuggingEnabled:        proto.Bool(true),
					RemoteDebuggingVersion:        "VS2022",
					RuntimeScaleMonitoringEnabled: proto.Bool(true),
					WebsocketsEnabled:             proto.Bool(true),
					VnetRouteAllEnabled:           proto.Bool(true),
					Cors: &AzureFunctionAppFlexConsumptionCorsSettings{
						AllowedOrigins:     []string{"https://app.example.com"},
						SupportCredentials: proto.Bool(true),
					},
					IpRestrictions: []*AzureFunctionAppFlexConsumptionIpRestriction{
						{
							Name:       "front-door-only",
							Priority:   proto.Int32(100),
							Action:     AzureFunctionAppFlexConsumptionIpRestrictionAction_ALLOW,
							ServiceTag: "AzureFrontDoor.Backend",
							Headers: &AzureFunctionAppFlexConsumptionIpRestrictionHeaders{
								XAzureFdid: []*foreignkeyv1.StringValueOrRef{literal("00000000-0000-0000-0000-000000000000")},
							},
						},
						{
							Name:      "office",
							Priority:  proto.Int32(200),
							Action:    AzureFunctionAppFlexConsumptionIpRestrictionAction_DENY,
							IpAddress: "203.0.113.0/24",
						},
					},
					IpRestrictionDefaultAction: AzureFunctionAppFlexConsumptionIpRestrictionAction_DENY,
					ScmUseMainIpRestriction:    proto.Bool(true),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept sticky settings with one non-empty list", func() {
				input := validResource()
				input.Spec.StickySettings = &AzureFunctionAppFlexConsumptionStickySettings{
					AppSettingNames: []string{"ENVIRONMENT"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept connection strings", func() {
				input := validResource()
				input.Spec.ConnectionStrings = []*AzureFunctionAppFlexConsumptionConnectionString{
					{
						Name:  "orders-db",
						Type:  AzureFunctionAppFlexConsumptionConnectionStringType_SQL_AZURE,
						Value: literal("Server=tcp:db.database.windows.net;Database=orders"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept Easy Auth v2 with Entra ID and a custom OIDC provider", func() {
				input := validResource()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppFlexConsumptionAuthSettingsV2{
					AuthEnabled:           proto.Bool(true),
					RequireAuthentication: proto.Bool(true),
					UnauthenticatedAction: AzureFunctionAppFlexConsumptionUnauthenticatedAction_RETURN_401,
					Login: &AzureFunctionAppFlexConsumptionAuthV2Login{
						TokenStoreEnabled: proto.Bool(true),
					},
					ActiveDirectoryV2: &AzureFunctionAppFlexConsumptionAuthV2ActiveDirectory{
						ClientId:                "00000000-0000-0000-0000-000000000000",
						TenantAuthEndpoint:      "https://login.microsoftonline.com/v2.0/11111111-1111-1111-1111-111111111111/",
						ClientSecretSettingName: "AAD_CLIENT_SECRET",
					},
					CustomOidcV2: []*AzureFunctionAppFlexConsumptionAuthV2CustomOidc{
						{
							Name:                        "corp-idp",
							ClientId:                    "corp-client",
							OpenidConfigurationEndpoint: "https://idp.example.com/.well-known/openid-configuration",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept the forward-proxy custom headers with the CUSTOM convention", func() {
				input := validResource()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppFlexConsumptionAuthSettingsV2{
					Login:                              &AzureFunctionAppFlexConsumptionAuthV2Login{},
					ForwardProxyConvention:             AzureFunctionAppFlexConsumptionForwardProxyConvention_FORWARD_PROXY_CUSTOM,
					ForwardProxyCustomHostHeaderName:   "X-Original-Host",
					ForwardProxyCustomSchemeHeaderName: "X-Original-Scheme",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("required fields", func() {

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing service plan", func() {
				input := validResource()
				input.Spec.ServicePlanId = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing runtime name and version", func() {
				input := validResource()
				input.Spec.RuntimeName = AzureFunctionAppFlexConsumptionRuntimeName_azure_function_app_flex_consumption_runtime_name_unspecified
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.RuntimeVersion = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing site config", func() {
				input := validResource()
				input.Spec.SiteConfig = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing storage authentication type", func() {
				input := validResource()
				input.Spec.StorageAuthenticationType = AzureFunctionAppFlexConsumptionStorageAuthenticationType_azure_function_app_flex_consumption_storage_authentication_type_unspecified
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("app name", func() {

			ginkgo.It("should reject names with illegal characters or over-length names", func() {
				input := validResource()
				input.Spec.FunctionAppName = "flex_app"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.FunctionAppName = "x" + "a123456789a123456789a123456789a123456789a123456789a123456789"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("deployment storage", func() {

			ginkgo.It("should reject a non-https or container-less endpoint", func() {
				input := validResource()
				input.Spec.StorageContainerEndpoint = "http://appstorage.blob.core.windows.net/deployments"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.StorageContainerEndpoint = "https://appstorage.blob.core.windows.net"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.StorageContainerEndpoint = "https://appstorage/deployments"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject connection-string auth without an access key", func() {
				input := validResource()
				input.Spec.StorageAccessKey = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject user-assigned identity auth without the identity id", func() {
				input := validResource()
				input.Spec.StorageAuthenticationType = AzureFunctionAppFlexConsumptionStorageAuthenticationType_USER_ASSIGNED_IDENTITY
				input.Spec.StorageAccessKey = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("scale dials", func() {

			ginkgo.It("should reject out-of-range maximum_instance_count", func() {
				input := validResource()
				input.Spec.MaximumInstanceCount = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.MaximumInstanceCount = proto.Int32(1001)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject out-of-range http_concurrency", func() {
				input := validResource()
				input.Spec.HttpConcurrency = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.HttpConcurrency = proto.Int32(1001)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an always-ready pool without a name or out of range", func() {
				input := validResource()
				input.Spec.AlwaysReady = []*AzureFunctionAppFlexConsumptionAlwaysReady{{Name: ""}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.AlwaysReady = []*AzureFunctionAppFlexConsumptionAlwaysReady{{Name: "http", InstanceCount: proto.Int32(1001)}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("site config", func() {

			ginkgo.It("should reject an unpaired health check", func() {
				input := validResource()
				input.Spec.SiteConfig.HealthCheckPath = "/api/health"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = proto.Int32(5)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an out-of-range health check eviction time", func() {
				input := validResource()
				input.Spec.SiteConfig.HealthCheckPath = "/api/health"
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = proto.Int32(1)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.SiteConfig.HealthCheckEvictionTimeInMin = proto.Int32(11)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an out-of-range worker count", func() {
				input := validResource()
				input.Spec.SiteConfig.WorkerCount = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.SiteConfig.WorkerCount = proto.Int32(101)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown remote debugging version", func() {
				input := validResource()
				input.Spec.SiteConfig.RemoteDebuggingVersion = "VS2015"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject out-of-range app service log settings", func() {
				input := validResource()
				input.Spec.SiteConfig.AppServiceLogs = &AzureFunctionAppFlexConsumptionAppServiceLogs{DiskQuotaMb: proto.Int32(24)}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.SiteConfig.AppServiceLogs = &AzureFunctionAppFlexConsumptionAppServiceLogs{DiskQuotaMb: proto.Int32(101)}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject CORS without origins and the credentials-wildcard pairing", func() {
				input := validResource()
				input.Spec.SiteConfig.Cors = &AzureFunctionAppFlexConsumptionCorsSettings{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.SiteConfig.Cors = &AzureFunctionAppFlexConsumptionCorsSettings{
					AllowedOrigins:     []string{"*"},
					SupportCredentials: proto.Bool(true),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject out-of-range ip restriction priorities and over-long header lists", func() {
				input := validResource()
				input.Spec.SiteConfig.IpRestrictions = []*AzureFunctionAppFlexConsumptionIpRestriction{
					{Name: "r", Priority: proto.Int32(0), IpAddress: "10.0.0.0/24"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.SiteConfig.IpRestrictions = []*AzureFunctionAppFlexConsumptionIpRestriction{
					{Name: "r", Priority: proto.Int32(65001), IpAddress: "10.0.0.0/24"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				nine := make([]string, 9)
				for i := range nine {
					nine[i] = "203.0.113.1/32"
				}
				input.Spec.SiteConfig.IpRestrictions = []*AzureFunctionAppFlexConsumptionIpRestriction{
					{Name: "r", IpAddress: "10.0.0.0/24", Headers: &AzureFunctionAppFlexConsumptionIpRestrictionHeaders{XForwardedFor: nine}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject vnet_route_all_enabled without a subnet", func() {
				input := validResource()
				input.Spec.SiteConfig.VnetRouteAllEnabled = proto.Bool(true)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("identity and settings blocks", func() {

			ginkgo.It("should reject a mispaired identity block", func() {
				input := validResource()
				input.Spec.Identity = &AzureFunctionAppFlexConsumptionIdentity{
					Type: AzureFunctionAppFlexConsumptionIdentityType_USER_ASSIGNED,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.Identity = &AzureFunctionAppFlexConsumptionIdentity{
					Type:        AzureFunctionAppFlexConsumptionIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject empty sticky settings", func() {
				input := validResource()
				input.Spec.StickySettings = &AzureFunctionAppFlexConsumptionStickySettings{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a connection string without a type or value", func() {
				input := validResource()
				input.Spec.ConnectionStrings = []*AzureFunctionAppFlexConsumptionConnectionString{
					{Name: "orders-db", Value: literal("Server=...")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.ConnectionStrings = []*AzureFunctionAppFlexConsumptionConnectionString{
					{Name: "orders-db", Type: AzureFunctionAppFlexConsumptionConnectionStringType_SQL_AZURE},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("auth settings v2", func() {

			ginkgo.It("should reject auth settings without the login block", func() {
				input := validResource()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppFlexConsumptionAuthSettingsV2{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Entra ID with both credential forms", func() {
				input := validResource()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppFlexConsumptionAuthSettingsV2{
					Login: &AzureFunctionAppFlexConsumptionAuthV2Login{},
					ActiveDirectoryV2: &AzureFunctionAppFlexConsumptionAuthV2ActiveDirectory{
						ClientId:                          "00000000-0000-0000-0000-000000000000",
						TenantAuthEndpoint:                "https://login.microsoftonline.com/v2.0/11111111-1111-1111-1111-111111111111/",
						ClientSecretSettingName:           "AAD_CLIENT_SECRET",
						ClientSecretCertificateThumbprint: "AABBCCDDEEFF",
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject custom proxy headers without the CUSTOM convention", func() {
				input := validResource()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppFlexConsumptionAuthSettingsV2{
					Login:                            &AzureFunctionAppFlexConsumptionAuthV2Login{},
					ForwardProxyCustomHostHeaderName: "X-Original-Host",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject both token store backings at once", func() {
				input := validResource()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppFlexConsumptionAuthSettingsV2{
					Login: &AzureFunctionAppFlexConsumptionAuthV2Login{
						TokenStorePath:           "/tokens",
						TokenStoreSasSettingName: "TOKEN_STORE_SAS",
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject malformed cookie and nonce lifetimes", func() {
				input := validResource()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppFlexConsumptionAuthSettingsV2{
					Login: &AzureFunctionAppFlexConsumptionAuthV2Login{
						CookieExpirationTime: proto.String("8h"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.AuthSettingsV2 = &AzureFunctionAppFlexConsumptionAuthSettingsV2{
					Login: &AzureFunctionAppFlexConsumptionAuthV2Login{
						NonceExpirationTime: proto.String("5m"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("api envelope", func() {

			ginkgo.It("should reject a wrong kind or api version", func() {
				input := validResource()
				input.Kind = "AzureFunctionApp"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				input = validResource()
				input.ApiVersion = "azure.planton.dev/v1"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})

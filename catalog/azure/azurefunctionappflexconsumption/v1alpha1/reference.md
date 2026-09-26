# AzureFunctionAppFlexConsumption

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureFunctionAppFlexConsumptionSpec** defines the configuration for
creating an Azure Function App on the Flex Consumption plan
(Microsoft.Web/sites kind=functionapp,linux on an FC1 plan) -- Azure's
newest serverless Functions hosting model.

Flex Consumption differs from the classic Consumption/Premium models in
three ways, and each shapes this spec:

  1. **Deployment storage is explicit.** The app's code package lives in
     a blob container YOU provide (`storage_container_endpoint`), and the
     app authenticates to it by connection string, system-assigned
     identity, or user-assigned identity
     (`storage_authentication_type`). There is no platform-managed
     content share.
  2. **Scale is per-instance and configurable.** Instances have a chosen
     memory size (`instance_memory_in_mb`), a scale-out ceiling
     (`maximum_instance_count`), optional per-instance HTTP concurrency
     (`http_concurrency`), and named always-ready instance pools
     (`always_ready`) that eliminate cold starts for specific
     functions or groups.
  3. **The runtime is declared flat.** `runtime_name` + `runtime_version`
     replace the classic application_stack block; containers are not
     supported on Flex Consumption.

**Relationship to AzureFunctionApp**: Azure models Flex Consumption
function apps as their own resource type with a distinct configuration
surface, so they are a separate kind. Classic Consumption (Y1), Elastic
Premium (EP*), and Dedicated-plan function apps are AzureFunctionApp.

**Relationship to AzureServicePlan**: the referenced plan MUST be the
FLEX_CONSUMPTION_FC1 SKU -- Azure rejects app creation on any other
tier ("the sku name is ... which is not valid for a flex consumption
function app"). One FC1 plan can host multiple flex apps.

**Authentication (Easy Auth)**: the `auth_settings_v2` block turns on
App Service's built-in authentication layer -- Azure validates identity
tokens at the front door (Entra ID, Apple, Facebook, GitHub, Google,
Microsoft account, Twitter, or any OpenID Connect provider) before
requests reach function code. Provider secrets are referenced by APP
SETTING NAME (never inline). Legacy `auth_settings` (v1) is superseded
by v2 and deliberately not modeled.

**ForceNew fields** (changing these destroys and recreates the app):
`function_app_name`, `region`, `resource_group`, `service_plan_id`.

## Example

```yaml
# Maximal offline-validation manifest for
# AzureFunctionAppFlexConsumption: every configurable arm populated
# with literal values -- the user-assigned-identity deployment-storage
# mode (the two live lanes prove the connection-string and
# system-assigned modes), all scale dials with multiple always-ready
# pools, the full site_config surface, Easy Auth v2 with EVERY identity
# provider plus a custom OIDC entry, connection strings, sticky
# settings, and the combined identity model. Drives validate-manifest
# and the offline tofu plan's rendered-value inspection; never deployed
# live.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFunctionAppFlexConsumption
metadata:
  name: flex-consumption-maximal
  org: planton-oss
  env: e2e
spec:
  region: centralus
  resourceGroup:
    value: app-rg
  functionAppName: acme-orders-flex
  servicePlanId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Web/serverFarms/flex-plan
  storageContainerEndpoint: https://acmeordersflexsa.blob.core.windows.net/deployments
  storageAuthenticationType: USER_ASSIGNED_IDENTITY
  storageUserAssignedIdentityId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/orders-mi
  runtimeName: DOTNET_ISOLATED
  runtimeVersion: "8.0"
  instanceMemoryInMb: 4096
  maximumInstanceCount: 250
  httpConcurrency: 32
  alwaysReady:
    - name: http
      instanceCount: 2
    - name: durable
      instanceCount: 1
    - name: function:ProcessOrders
      instanceCount: 1
  siteConfig:
    apiManagementApiId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ApiManagement/service/acme-apim/apis/orders
    apiDefinitionUrl: https://acme.example.com/openapi.json
    appCommandLine: ""
    applicationInsightsKey: 00000000-0000-0000-0000-000000000000
    appServiceLogs:
      diskQuotaMb: 50
      retentionPeriodDays: 7
    containerRegistryUseManagedIdentity: true
    defaultDocuments:
      - index.html
    elasticInstanceMinimum: 1
    http2Enabled: true
    ipRestrictions:
      - name: front-door-only
        priority: 100
        action: ALLOW
        serviceTag: AzureFrontDoor.Backend
        description: Only Front Door reaches the origin
        headers:
          xForwardedFor:
            - 203.0.113.0/24
          xForwardedHost:
            - orders.acme.example.com
          xAzureFdid:
            - value: 11111111-2222-3333-4444-555555555555
          xFdHealthProbe:
            - "1"
      - name: office
        priority: 200
        action: DENY
        ipAddress: 198.51.100.0/24
      - name: hub-subnet
        priority: 300
        action: ALLOW
        virtualNetworkSubnetId:
          value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/net-rg/providers/Microsoft.Network/virtualNetworks/hub-vnet/subnets/agents
    ipRestrictionDefaultAction: DENY
    scmUseMainIpRestriction: false
    scmIpRestrictions:
      - name: ci-runners
        priority: 100
        action: ALLOW
        ipAddress: 192.0.2.0/24
    scmIpRestrictionDefaultAction: DENY
    loadBalancingMode: WEIGHTED_ROUND_ROBIN
    managedPipelineMode: INTEGRATED
    remoteDebuggingEnabled: true
    remoteDebuggingVersion: VS2022
    runtimeScaleMonitoringEnabled: true
    websocketsEnabled: true
    healthCheckPath: /api/health
    healthCheckEvictionTimeInMin: 5
    workerCount: 4
    minimumTlsVersion: TLS_1_3
    scmMinimumTlsVersion: TLS_1_2
    cors:
      allowedOrigins:
        - https://orders.acme.example.com
        - https://admin.acme.example.com
      supportCredentials: true
    vnetRouteAllEnabled: true
  appSettings:
    ENVIRONMENT: production
    AAD_CLIENT_SECRET: set-me-via-key-vault-reference
    GITHUB_CLIENT_SECRET: set-me-via-key-vault-reference
  connectionStrings:
    - name: orders-db
      type: SQL_AZURE
      value:
        value: Server=tcp:orders.database.windows.net;Database=orders
    - name: cache
      type: REDIS_CACHE
      value:
        value: acme-cache.redis.cache.windows.net:6380,ssl=true
  stickySettings:
    appSettingNames:
      - ENVIRONMENT
    connectionStringNames:
      - orders-db
  applicationInsightsConnectionString:
    value: InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://centralus-2.in.applicationinsights.azure.com/
  httpsOnly: true
  publicNetworkAccessEnabled: true
  enabled: true
  clientCertificateEnabled: true
  clientCertificateMode: OPTIONAL_INTERACTIVE_USER
  clientCertificateExclusionPaths: /api/health;/api/status
  virtualNetworkSubnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/net-rg/providers/Microsoft.Network/virtualNetworks/hub-vnet/subnets/flex-integration
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/orders-mi
  webdeployPublishBasicAuthenticationEnabled: false
  authSettingsV2:
    authEnabled: true
    requireAuthentication: true
    unauthenticatedAction: RETURN_401
    defaultProvider: azureactivedirectory
    excludedPaths:
      - /api/health
    requireHttps: true
    login:
      tokenStoreEnabled: true
      tokenRefreshExtensionTime: 96
      cookieExpirationConvention: FIXED_TIME
      cookieExpirationTime: "08:00:00"
      validateNonce: true
      nonceExpirationTime: "00:05:00"
      allowedExternalRedirectUrls:
        - https://orders.acme.example.com/signed-in
    appleV2:
      clientId: com.acme.orders
      clientSecretSettingName: APPLE_CLIENT_SECRET
    activeDirectoryV2:
      clientId: 22222222-2222-2222-2222-222222222222
      tenantAuthEndpoint: https://login.microsoftonline.com/v2.0/33333333-3333-3333-3333-333333333333/
      clientSecretSettingName: AAD_CLIENT_SECRET
      allowedAudiences:
        - api://acme-orders
      allowedGroups:
        - 44444444-4444-4444-4444-444444444444
    azureStaticWebAppV2:
      clientId: 55555555-5555-5555-5555-555555555555
    customOidcV2:
      - name: corp-idp
        clientId: corp-orders-client
        openidConfigurationEndpoint: https://idp.acme.example.com/.well-known/openid-configuration
        nameClaimType: name
        scopes:
          - openid
          - profile
    facebookV2:
      appId: "666666666666666"
      appSecretSettingName: FACEBOOK_APP_SECRET
      graphApiVersion: v17.0
    githubV2:
      clientId: Iv1.acmeorders
      clientSecretSettingName: GITHUB_CLIENT_SECRET
    googleV2:
      clientId: acme-orders.apps.googleusercontent.com
      clientSecretSettingName: GOOGLE_CLIENT_SECRET
      loginScopes:
        - openid
    microsoftV2:
      clientId: 77777777-7777-7777-7777-777777777777
      clientSecretSettingName: MICROSOFT_CLIENT_SECRET
    twitterV2:
      consumerKey: acmeorderskey
      consumerSecretSettingName: TWITTER_CONSUMER_SECRET
  tags:
    cost-center: orders
    owner: platform-team
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.functionAppName` | `string` | yes |  |  |
| `spec.servicePlanId` | `string \| valueFrom` | yes |  | AzureServicePlan (`status.outputs.service_plan_id`) |
| `spec.storageContainerEndpoint` | `string` | yes |  |  |
| `spec.storageAuthenticationType` | `enum` | yes |  |  |
| `spec.storageAccessKey` | `string \| valueFrom` (sensitive) |  |  | AzureStorageAccount (`status.outputs.primary_access_key`) |
| `spec.storageUserAssignedIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.runtimeName` | `enum` | yes |  |  |
| `spec.runtimeVersion` | `string` | yes |  |  |
| `spec.instanceMemoryInMb` | `int32` |  | `2048` |  |
| `spec.maximumInstanceCount` | `int32` |  | `100` |  |
| `spec.httpConcurrency` | `int32` |  |  |  |
| `spec.alwaysReady` | `[]AzureFunctionAppFlexConsumptionAlwaysReady` |  |  |  |
| `spec.alwaysReady[].name` | `string` | yes |  |  |
| `spec.alwaysReady[].instanceCount` | `int32` |  |  |  |
| `spec.siteConfig` | `AzureFunctionAppFlexConsumptionSiteConfig` | yes |  |  |
| `spec.siteConfig.apiManagementApiId` | `string` |  |  |  |
| `spec.siteConfig.apiDefinitionUrl` | `string` |  |  |  |
| `spec.siteConfig.appCommandLine` | `string` |  |  |  |
| `spec.siteConfig.applicationInsightsKey` | `string` (sensitive) |  |  |  |
| `spec.siteConfig.appServiceLogs` | `AzureFunctionAppFlexConsumptionAppServiceLogs` |  |  |  |
| `spec.siteConfig.appServiceLogs.diskQuotaMb` | `int32` |  | `35` |  |
| `spec.siteConfig.appServiceLogs.retentionPeriodDays` | `int32` |  |  |  |
| `spec.siteConfig.containerRegistryUseManagedIdentity` | `bool` |  | `false` |  |
| `spec.siteConfig.defaultDocuments` | `[]string` |  |  |  |
| `spec.siteConfig.elasticInstanceMinimum` | `int32` |  |  |  |
| `spec.siteConfig.http2Enabled` | `bool` |  | `false` |  |
| `spec.siteConfig.ipRestrictions` | `[]AzureFunctionAppFlexConsumptionIpRestriction` |  |  |  |
| `spec.siteConfig.ipRestrictions[].name` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].priority` | `int32` |  |  |  |
| `spec.siteConfig.ipRestrictions[].action` | `enum` |  |  |  |
| `spec.siteConfig.ipRestrictions[].ipAddress` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].serviceTag` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].virtualNetworkSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.siteConfig.ipRestrictions[].description` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers` | `AzureFunctionAppFlexConsumptionIpRestrictionHeaders` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers.xForwardedFor` | `[]string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers.xForwardedHost` | `[]string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers.xAzureFdid` | `[]string \| valueFrom` |  |  | AzureFrontDoorProfile (`status.outputs.resource_guid`) |
| `spec.siteConfig.ipRestrictions[].headers.xFdHealthProbe` | `[]string` |  |  |  |
| `spec.siteConfig.ipRestrictionDefaultAction` | `enum` |  |  |  |
| `spec.siteConfig.scmUseMainIpRestriction` | `bool` |  | `false` |  |
| `spec.siteConfig.scmIpRestrictions` | `[]AzureFunctionAppFlexConsumptionIpRestriction` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].name` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].priority` | `int32` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].action` | `enum` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].ipAddress` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].serviceTag` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].virtualNetworkSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.siteConfig.scmIpRestrictions[].description` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers` | `AzureFunctionAppFlexConsumptionIpRestrictionHeaders` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers.xForwardedFor` | `[]string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers.xForwardedHost` | `[]string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers.xAzureFdid` | `[]string \| valueFrom` |  |  | AzureFrontDoorProfile (`status.outputs.resource_guid`) |
| `spec.siteConfig.scmIpRestrictions[].headers.xFdHealthProbe` | `[]string` |  |  |  |
| `spec.siteConfig.scmIpRestrictionDefaultAction` | `enum` |  |  |  |
| `spec.siteConfig.loadBalancingMode` | `enum` |  |  |  |
| `spec.siteConfig.managedPipelineMode` | `enum` |  |  |  |
| `spec.siteConfig.remoteDebuggingEnabled` | `bool` |  | `false` |  |
| `spec.siteConfig.remoteDebuggingVersion` | `string` |  |  |  |
| `spec.siteConfig.runtimeScaleMonitoringEnabled` | `bool` |  |  |  |
| `spec.siteConfig.websocketsEnabled` | `bool` |  | `false` |  |
| `spec.siteConfig.healthCheckPath` | `string` |  |  |  |
| `spec.siteConfig.healthCheckEvictionTimeInMin` | `int32` |  |  |  |
| `spec.siteConfig.workerCount` | `int32` |  |  |  |
| `spec.siteConfig.minimumTlsVersion` | `enum` |  |  |  |
| `spec.siteConfig.scmMinimumTlsVersion` | `enum` |  |  |  |
| `spec.siteConfig.cors` | `AzureFunctionAppFlexConsumptionCorsSettings` |  |  |  |
| `spec.siteConfig.cors.allowedOrigins` | `[]string` | yes |  |  |
| `spec.siteConfig.cors.supportCredentials` | `bool` |  | `false` |  |
| `spec.siteConfig.vnetRouteAllEnabled` | `bool` |  | `false` |  |
| `spec.appSettings` | `map<string, string>` |  |  |  |
| `spec.connectionStrings` | `[]AzureFunctionAppFlexConsumptionConnectionString` |  |  |  |
| `spec.connectionStrings[].name` | `string` | yes |  |  |
| `spec.connectionStrings[].type` | `enum` | yes |  |  |
| `spec.connectionStrings[].value` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.stickySettings` | `AzureFunctionAppFlexConsumptionStickySettings` |  |  |  |
| `spec.stickySettings.appSettingNames` | `[]string` |  |  |  |
| `spec.stickySettings.connectionStringNames` | `[]string` |  |  |  |
| `spec.applicationInsightsConnectionString` | `string \| valueFrom` |  |  | AzureApplicationInsights (`status.outputs.connection_string`) |
| `spec.httpsOnly` | `bool` |  | `true` |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.clientCertificateEnabled` | `bool` |  | `false` |  |
| `spec.clientCertificateMode` | `enum` |  |  |  |
| `spec.clientCertificateExclusionPaths` | `string` |  |  |  |
| `spec.virtualNetworkSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.identity` | `AzureFunctionAppFlexConsumptionIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.webdeployPublishBasicAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.zipDeployFile` | `string` |  |  |  |
| `spec.authSettingsV2` | `AzureFunctionAppFlexConsumptionAuthSettingsV2` |  |  |  |
| `spec.authSettingsV2.authEnabled` | `bool` |  | `false` |  |
| `spec.authSettingsV2.runtimeVersion` | `string` |  | `~1` |  |
| `spec.authSettingsV2.configFilePath` | `string` |  |  |  |
| `spec.authSettingsV2.requireAuthentication` | `bool` |  | `false` |  |
| `spec.authSettingsV2.unauthenticatedAction` | `enum` |  |  |  |
| `spec.authSettingsV2.defaultProvider` | `string` |  |  |  |
| `spec.authSettingsV2.excludedPaths` | `[]string` |  |  |  |
| `spec.authSettingsV2.requireHttps` | `bool` |  | `true` |  |
| `spec.authSettingsV2.httpRouteApiPrefix` | `string` |  | `/.auth` |  |
| `spec.authSettingsV2.forwardProxyConvention` | `enum` |  |  |  |
| `spec.authSettingsV2.forwardProxyCustomHostHeaderName` | `string` |  |  |  |
| `spec.authSettingsV2.forwardProxyCustomSchemeHeaderName` | `string` |  |  |  |
| `spec.authSettingsV2.login` | `AzureFunctionAppFlexConsumptionAuthV2Login` | yes |  |  |
| `spec.authSettingsV2.login.logoutEndpoint` | `string` |  |  |  |
| `spec.authSettingsV2.login.tokenStoreEnabled` | `bool` |  | `false` |  |
| `spec.authSettingsV2.login.tokenRefreshExtensionTime` | `double` |  | `72` |  |
| `spec.authSettingsV2.login.tokenStorePath` | `string` |  |  |  |
| `spec.authSettingsV2.login.tokenStoreSasSettingName` | `string` |  |  |  |
| `spec.authSettingsV2.login.preserveUrlFragmentsForLogins` | `bool` |  | `false` |  |
| `spec.authSettingsV2.login.allowedExternalRedirectUrls` | `[]string` |  |  |  |
| `spec.authSettingsV2.login.cookieExpirationConvention` | `enum` |  |  |  |
| `spec.authSettingsV2.login.cookieExpirationTime` | `string` |  | `08:00:00` |  |
| `spec.authSettingsV2.login.validateNonce` | `bool` |  | `true` |  |
| `spec.authSettingsV2.login.nonceExpirationTime` | `string` |  | `00:05:00` |  |
| `spec.authSettingsV2.appleV2` | `AzureFunctionAppFlexConsumptionAuthV2Apple` |  |  |  |
| `spec.authSettingsV2.appleV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.appleV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.activeDirectoryV2` | `AzureFunctionAppFlexConsumptionAuthV2ActiveDirectory` |  |  |  |
| `spec.authSettingsV2.activeDirectoryV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.activeDirectoryV2.tenantAuthEndpoint` | `string` | yes |  |  |
| `spec.authSettingsV2.activeDirectoryV2.clientSecretSettingName` | `string` |  |  |  |
| `spec.authSettingsV2.activeDirectoryV2.clientSecretCertificateThumbprint` | `string` |  |  |  |
| `spec.authSettingsV2.activeDirectoryV2.loginParameters` | `map<string, string>` |  |  |  |
| `spec.authSettingsV2.activeDirectoryV2.wwwAuthenticationDisabled` | `bool` |  | `false` |  |
| `spec.authSettingsV2.activeDirectoryV2.jwtAllowedGroups` | `[]string` |  |  |  |
| `spec.authSettingsV2.activeDirectoryV2.jwtAllowedClientApplications` | `[]string` |  |  |  |
| `spec.authSettingsV2.activeDirectoryV2.allowedGroups` | `[]string` |  |  |  |
| `spec.authSettingsV2.activeDirectoryV2.allowedIdentities` | `[]string` |  |  |  |
| `spec.authSettingsV2.activeDirectoryV2.allowedApplications` | `[]string` |  |  |  |
| `spec.authSettingsV2.activeDirectoryV2.allowedAudiences` | `[]string` |  |  |  |
| `spec.authSettingsV2.azureStaticWebAppV2` | `AzureFunctionAppFlexConsumptionAuthV2StaticWebApp` |  |  |  |
| `spec.authSettingsV2.azureStaticWebAppV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2` | `[]AzureFunctionAppFlexConsumptionAuthV2CustomOidc` |  |  |  |
| `spec.authSettingsV2.customOidcV2[].name` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2[].clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2[].openidConfigurationEndpoint` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2[].nameClaimType` | `string` |  |  |  |
| `spec.authSettingsV2.customOidcV2[].scopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.facebookV2` | `AzureFunctionAppFlexConsumptionAuthV2Facebook` |  |  |  |
| `spec.authSettingsV2.facebookV2.appId` | `string` | yes |  |  |
| `spec.authSettingsV2.facebookV2.appSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.facebookV2.graphApiVersion` | `string` |  |  |  |
| `spec.authSettingsV2.facebookV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.githubV2` | `AzureFunctionAppFlexConsumptionAuthV2Github` |  |  |  |
| `spec.authSettingsV2.githubV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.githubV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.githubV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.googleV2` | `AzureFunctionAppFlexConsumptionAuthV2Google` |  |  |  |
| `spec.authSettingsV2.googleV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.googleV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.googleV2.allowedAudiences` | `[]string` |  |  |  |
| `spec.authSettingsV2.googleV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.microsoftV2` | `AzureFunctionAppFlexConsumptionAuthV2Microsoft` |  |  |  |
| `spec.authSettingsV2.microsoftV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.microsoftV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.microsoftV2.allowedAudiences` | `[]string` |  |  |  |
| `spec.authSettingsV2.microsoftV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.twitterV2` | `AzureFunctionAppFlexConsumptionAuthV2Twitter` |  |  |  |
| `spec.authSettingsV2.twitterV2.consumerKey` | `string` | yes |  |  |
| `spec.authSettingsV2.twitterV2.consumerSecretSettingName` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Function App will be created.
Examples: "eastus", "westus2", "westeurope", "southeastasia".

Flex Consumption is not available in every region -- check
availability before choosing (the service plan and the app must be
in the same region).

**ForceNew**: Changing this destroys and recreates the function app.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the Function App will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

**ForceNew**: Changing this destroys and recreates the function app.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.functionAppName

`string` · required

The name of the Function App.
Must be globally unique across Azure (it forms the default hostname:
`{function_app_name}.azurewebsites.net`).

Allowed characters: alphanumeric and dashes, 1 to 60 characters
(exactly the rule Azure's own tooling enforces; names that start or
end with a dash produce awkward hostnames but are accepted).

**ForceNew**: Changing this destroys and recreates the function app.

- rule: function_app_name may only contain alphanumeric characters and dashes and must be 1 to 60 characters long
- rule: {"required":true}

### spec.servicePlanId

`string | valueFrom` · required

The App Service Plan that hosts this Function App. The plan's SKU
MUST be FLEX_CONSUMPTION_FC1 -- Azure rejects creation on any other
tier at apply time. One FC1 plan can host multiple flex apps; the
plan itself has no idle compute cost (billing follows executions and
always-ready instances).

**ForceNew**: Changing this destroys and recreates the function app.

- references: AzureServicePlan (`status.outputs.service_plan_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServicePlan, name: <that resource's name>, fieldPath: status.outputs.service_plan_id}} -- a bare string does not parse

### spec.storageContainerEndpoint

`string` · required

The HTTPS endpoint of the blob CONTAINER holding the app's code
package -- Flex Consumption's deployment storage. Format:
https://{storage-account}.blob.core.windows.net/{container}

Compose it from an AzureStorageAccount's `primary_blob_endpoint`
output plus an AzureStorageContainer's name (the container kind
deliberately exports no URL -- the endpoint is account-level).
The container must exist before the app is created; the platform
uploads the deployment package into it.

- rule: storage_container_endpoint must be an https blob container URL (https://{account}.blob.core.windows.net/{container})
- rule: {"required":true}

### spec.storageAuthenticationType

`enum` · required

How the app authenticates to the deployment storage container.

STORAGE_ACCOUNT_CONNECTION_STRING requires storage_access_key;
USER_ASSIGNED_IDENTITY requires storage_user_assigned_identity_id
(both enforced here, exactly as Azure enforces them at apply time);
SYSTEM_ASSIGNED_IDENTITY needs neither -- grant the app's
system-assigned identity "Storage Blob Data Contributor" on the
storage account instead. Live-proven: ARM does NOT check that
grant at app create -- the site object succeeds without it. The
grant is still required before a package deploy (day-2).

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_storage_authentication_type_unspecified` -- Not specified -- invalid; pick an explicit authentication method.
- `STORAGE_ACCOUNT_CONNECTION_STRING` -- Authenticate with the storage account's access key (requires storage_access_key). Azure derives the connection string and manages it as the DEPLOYMENT_STORAGE_CONNECTION_STRING app setting.
- `SYSTEM_ASSIGNED_IDENTITY` -- Authenticate as the app's system-assigned managed identity (credential-free; grant it "Storage Blob Data Contributor" on the storage account).
- `USER_ASSIGNED_IDENTITY` -- Authenticate as a user-assigned managed identity (requires storage_user_assigned_identity_id; attach the same identity via identity.identity_ids).

### spec.storageAccessKey

`string | valueFrom` · sensitive

The storage account access key, required when
storage_authentication_type is STORAGE_ACCOUNT_CONNECTION_STRING.
Defaults to referencing an AzureStorageAccount's primary_access_key
output so the binding composes in one manifest set; a literal value
or a managed-secret reference works too.

SECRET-BEARING (treated as sensitive here even though the write-only
wire field is not marked so upstream). Prefer an identity-based
storage_authentication_type where the workload supports it -- keys
are static credential material.

- references: AzureStorageAccount (`status.outputs.primary_access_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_access_key}} -- a bare string does not parse

### spec.storageUserAssignedIdentityId

`string | valueFrom`

The user-assigned managed identity that accesses the deployment
storage, required when storage_authentication_type is
USER_ASSIGNED_IDENTITY. The identity needs "Storage Blob Data
Contributor" on the storage account, and must also be attached to
the app via identity.identity_ids.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.runtimeName

`enum` · required

The language runtime the app's functions run on.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_runtime_name_unspecified` -- Not specified -- invalid; pick a runtime.
- `NODE` -- Node.js (deploys as "node").
- `DOTNET_ISOLATED` -- .NET isolated worker (deploys as "dotnet-isolated"; the in-process model does not exist on Flex Consumption).
- `JAVA` -- Java (deploys as "java").
- `POWERSHELL` -- PowerShell (deploys as "powershell").
- `PYTHON` -- Python (deploys as "python").
- `CUSTOM_HANDLER` -- Custom handler -- any language implementing the Functions custom handler protocol (deploys as "custom"; named CUSTOM_HANDLER here because proto enum values share the file scope with the connection-string type vocabulary's CUSTOM).

### spec.runtimeVersion

`string` · required

The runtime version, as Azure spells it for the chosen runtime.
Azure evolves the supported set continuously, so any non-empty value
is accepted here and validated by Azure at apply time. Current
examples: node "20"/"22", python "3.11"/"3.12", java "11"/"17"/"21",
dotnet-isolated "8.0"/"9.0", powershell "7.4".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.instanceMemoryInMb

`int32` · optional (explicit presence)

Memory available to EACH instance, in MB. Azure offers fixed sizes
(currently 512, 2048, and 4096); the accepted set is validated by
Azure at apply time. Unset deploys 2048.

- default: `2048`

### spec.maximumInstanceCount

`int32` · optional (explicit presence)

The scale-out ceiling -- the maximum number of instances the app can
fan out to. Range 1-1000. Unset deploys 100.

The sum of always_ready instance counts must stay within this
ceiling (Azure enforces it at apply time with "the total number of
always-ready instances should not exceed the maximum scale out
limit" -- a sum constraint manifest validation cannot express).

- default: `100`
- rule: {"int32":{"lte":1000,"gte":1}}

### spec.httpConcurrency

`int32` · optional (explicit presence)

Concurrent HTTP requests each instance handles before Azure scales
out further. Range 1-1000. Unset lets Azure pick the runtime's
default concurrency for the chosen instance memory size.

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.alwaysReady

`[]AzureFunctionAppFlexConsumptionAlwaysReady`

Named pools of pre-warmed instances that never scale to zero --
Flex Consumption's cold-start eliminator. Each entry names a scope
("http", "durable", "blob", or "function:{functionName}") and how
many instances stay warm for it. Always-ready instances bill for
their uptime (the app's only idle cost).

### spec.alwaysReady[].name

`string` · required

What the pool keeps warm: "http" (all HTTP triggers), "durable"
(Durable Functions), "blob" (blob triggers), or
"function:{functionName}" for one specific function. Azure
lower-cases the name on save, so treat it case-insensitively.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.alwaysReady[].instanceCount

`int32` · optional (explicit presence)

How many instances stay warm for this scope. Range 0-1000. The sum
across all always_ready entries must not exceed
maximum_instance_count (Azure enforces this at apply time).

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.siteConfig

`AzureFunctionAppFlexConsumptionSiteConfig` · required

Site-level configuration: App Insights wiring, access restrictions,
TLS floors, CORS, health checks, and operational toggles.

- rule: {"required":true}
- rule: health_check_path and health_check_eviction_time_in_min require each other (Azure pairs them both ways)

### spec.siteConfig.apiManagementApiId

`string`

The ARM ID of the API Management API this app backs. Wires the app
into an API Management gateway so the API surface is managed there.
Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/
        Microsoft.ApiManagement/service/{apim}/apis/{api}

### spec.siteConfig.apiDefinitionUrl

`string`

URL of the OpenAPI/Swagger definition describing this app's API.
Surfaces in the portal's API Definition blade and to API consumers.

### spec.siteConfig.appCommandLine

`string`

The program and any arguments used to launch this app via the
command line. (Example: "node myapp.js").

### spec.siteConfig.applicationInsightsKey

`string` · sensitive

Application Insights instrumentation key (classic).
Prefer application_insights_connection_string on the parent spec
for new deployments. This field is for backward compatibility with
apps already using the instrumentation key. Travels as the
APPINSIGHTS_INSTRUMENTATIONKEY app setting.

### spec.siteConfig.appServiceLogs

`AzureFunctionAppFlexConsumptionAppServiceLogs`

App Service logging configuration (disk quota + retention for the
file-system logs). Azure applies this on UPDATE operations only and
never returns it on read -- expect the portal, not the manifest, to
reflect drift here.

### spec.siteConfig.appServiceLogs.diskQuotaMb

`int32` · optional (explicit presence)

Disk quota for app service logs in megabytes.
Range: 25 to 100.

Default: 35

- default: `35`
- rule: {"int32":{"lte":100,"gte":25}}

### spec.siteConfig.appServiceLogs.retentionPeriodDays

`int32` · optional (explicit presence)

Log retention in days. Set to 0 for indefinite retention.

- rule: {"int32":{"gte":0}}

### spec.siteConfig.containerRegistryUseManagedIdentity

`bool` · optional (explicit presence)

Use managed identity for pulling container images from Azure
Container Registry (relevant to future container-based features;
Flex Consumption itself runs language runtimes, not containers).

Default: false

- default: `false`

### spec.siteConfig.defaultDocuments

`[]string`

Default documents served when a request maps to a directory.
Evaluated in order (e.g. ["index.html", "default.html"]).

### spec.siteConfig.elasticInstanceMinimum

`int32` · optional (explicit presence)

Minimum instance count for Elastic-Premium-style pre-warming.
Azure accepts the value on this hosting model but never returns it
on read (always-ready pools are the flex-native warm-instance
mechanism -- prefer always_ready on the spec).

- rule: {"int32":{"gte":0}}

### spec.siteConfig.http2Enabled

`bool` · optional (explicit presence)

Enable the HTTP/2 protocol.

Default: false

- default: `false`

### spec.siteConfig.ipRestrictions

`[]AzureFunctionAppFlexConsumptionIpRestriction`

IP restriction rules for the main site.
Controls which IP addresses, service tags, or subnets can access
the Function App.

### spec.siteConfig.ipRestrictions[].name

`string`

Rule name for identification.

### spec.siteConfig.ipRestrictions[].priority

`int32` · optional (explicit presence)

Rule priority. Lower numbers are evaluated first.
Range: 1 to 65000.

- rule: {"int32":{"lte":65000,"gte":1}}

### spec.siteConfig.ipRestrictions[].action

`enum`

Whether matching traffic is allowed or denied. Unset deploys ALLOW.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.ipRestrictions[].ipAddress

`string`

IP address or CIDR range (comma-separated ranges are accepted).
Example: "10.0.0.0/24", "203.0.113.50/32"

### spec.siteConfig.ipRestrictions[].serviceTag

`string`

Azure service tag.
Example: "AzureFrontDoor.Backend", "AzureCloud.WestUS"

### spec.siteConfig.ipRestrictions[].virtualNetworkSubnetId

`string | valueFrom`

Subnet ID for VNet-based access control.
Traffic from this subnet is allowed/denied based on the action.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.siteConfig.ipRestrictions[].description

`string`

Human-readable description of the rule.

### spec.siteConfig.ipRestrictions[].headers

`AzureFunctionAppFlexConsumptionIpRestrictionHeaders`

HTTP header filters for the rule.
Used with Azure Front Door or other reverse proxies to restrict
access based on request headers.

### spec.siteConfig.ipRestrictions[].headers.xForwardedFor

`[]string`

X-Forwarded-For header values to match.
Up to 8 entries. CIDR ranges are supported.

- rule: {"repeated":{"maxItems":"8"}}

### spec.siteConfig.ipRestrictions[].headers.xForwardedHost

`[]string`

X-Forwarded-Host header values to match.
Up to 8 entries.

- rule: {"repeated":{"maxItems":"8"}}

### spec.siteConfig.ipRestrictions[].headers.xAzureFdid

`[]string | valueFrom`

X-Azure-FDID (Front Door ID) header values to match, up to 8
entries. Locks the app to specific Front Door instances: pair an
ALLOW rule on the AzureFrontDoor.Backend service tag with this
filter so only YOUR profile's traffic reaches the origin -- without
it, anyone who discovers the app's default hostname bypasses the
edge (and its WAF) entirely. Each entry references an
AzureFrontDoorProfile's resource_guid output or carries the GUID as
a literal.

- references: AzureFrontDoorProfile (`status.outputs.resource_guid`)
- rule: {"repeated":{"maxItems":"8"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorProfile, name: <that resource's name>, fieldPath: status.outputs.resource_guid}} -- a bare string does not parse

### spec.siteConfig.ipRestrictions[].headers.xFdHealthProbe

`[]string`

X-FD-HealthProbe header values to match.
The only supported value is "1" (allow Front Door health probes).

- rule: {"repeated":{"maxItems":"1"}}

### spec.siteConfig.ipRestrictionDefaultAction

`enum`

Default action for traffic that matches no ip_restrictions rule.
Unset deploys ALLOW (rules act as a deny-list); DENY flips the rules
into an allow-list.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.scmUseMainIpRestriction

`bool` · optional (explicit presence)

Use the main site's IP restrictions for the SCM (Kudu) site.
When true, scm_ip_restrictions are ignored.

Default: false

- default: `false`

### spec.siteConfig.scmIpRestrictions

`[]AzureFunctionAppFlexConsumptionIpRestriction`

IP restriction rules for the SCM (Kudu) site.
Only used when scm_use_main_ip_restriction is false.

### spec.siteConfig.scmIpRestrictions[].name

`string`

Rule name for identification.

### spec.siteConfig.scmIpRestrictions[].priority

`int32` · optional (explicit presence)

Rule priority. Lower numbers are evaluated first.
Range: 1 to 65000.

- rule: {"int32":{"lte":65000,"gte":1}}

### spec.siteConfig.scmIpRestrictions[].action

`enum`

Whether matching traffic is allowed or denied. Unset deploys ALLOW.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.scmIpRestrictions[].ipAddress

`string`

IP address or CIDR range (comma-separated ranges are accepted).
Example: "10.0.0.0/24", "203.0.113.50/32"

### spec.siteConfig.scmIpRestrictions[].serviceTag

`string`

Azure service tag.
Example: "AzureFrontDoor.Backend", "AzureCloud.WestUS"

### spec.siteConfig.scmIpRestrictions[].virtualNetworkSubnetId

`string | valueFrom`

Subnet ID for VNet-based access control.
Traffic from this subnet is allowed/denied based on the action.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.siteConfig.scmIpRestrictions[].description

`string`

Human-readable description of the rule.

### spec.siteConfig.scmIpRestrictions[].headers

`AzureFunctionAppFlexConsumptionIpRestrictionHeaders`

HTTP header filters for the rule.
Used with Azure Front Door or other reverse proxies to restrict
access based on request headers.

### spec.siteConfig.scmIpRestrictions[].headers.xForwardedFor

`[]string`

X-Forwarded-For header values to match.
Up to 8 entries. CIDR ranges are supported.

- rule: {"repeated":{"maxItems":"8"}}

### spec.siteConfig.scmIpRestrictions[].headers.xForwardedHost

`[]string`

X-Forwarded-Host header values to match.
Up to 8 entries.

- rule: {"repeated":{"maxItems":"8"}}

### spec.siteConfig.scmIpRestrictions[].headers.xAzureFdid

`[]string | valueFrom`

X-Azure-FDID (Front Door ID) header values to match, up to 8
entries. Locks the app to specific Front Door instances: pair an
ALLOW rule on the AzureFrontDoor.Backend service tag with this
filter so only YOUR profile's traffic reaches the origin -- without
it, anyone who discovers the app's default hostname bypasses the
edge (and its WAF) entirely. Each entry references an
AzureFrontDoorProfile's resource_guid output or carries the GUID as
a literal.

- references: AzureFrontDoorProfile (`status.outputs.resource_guid`)
- rule: {"repeated":{"maxItems":"8"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorProfile, name: <that resource's name>, fieldPath: status.outputs.resource_guid}} -- a bare string does not parse

### spec.siteConfig.scmIpRestrictions[].headers.xFdHealthProbe

`[]string`

X-FD-HealthProbe header values to match.
The only supported value is "1" (allow Front Door health probes).

- rule: {"repeated":{"maxItems":"1"}}

### spec.siteConfig.scmIpRestrictionDefaultAction

`enum`

Default action for traffic that matches no scm_ip_restrictions rule.
Unset deploys ALLOW.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.loadBalancingMode

`enum`

Load balancing mode for distributing requests across instances.
Unset deploys LEAST_REQUESTS.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_load_balancing_mode_unspecified` -- Not specified -- deploys LEAST_REQUESTS.
- `LEAST_REQUESTS` -- Route to the instance with the fewest active requests (the default).
- `WEIGHTED_ROUND_ROBIN`
- `LEAST_RESPONSE_TIME`
- `WEIGHTED_TOTAL_TRAFFIC`
- `REQUEST_HASH`
- `PER_SITE_ROUND_ROBIN`

### spec.siteConfig.managedPipelineMode

`enum`

IIS-lineage request pipeline mode. INTEGRATED (the default) is
correct for everything modern; CLASSIC exists for legacy
compatibility only.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_managed_pipeline_mode_unspecified` -- Not specified -- deploys INTEGRATED.
- `INTEGRATED` -- The modern pipeline (the default; correct for everything current).
- `CLASSIC` -- Legacy-compatibility pipeline.

### spec.siteConfig.remoteDebuggingEnabled

`bool` · optional (explicit presence)

Enable remote debugging (Visual Studio attach). Turn off outside
active debugging sessions.

Default: false

- default: `false`

### spec.siteConfig.remoteDebuggingVersion

`string`

The Visual Studio generation remote debugging targets.
Leave empty to let Azure pick the current generation.

- rule: remote_debugging_version must be one of: VS2017, VS2019, VS2022

### spec.siteConfig.runtimeScaleMonitoringEnabled

`bool` · optional (explicit presence)

Enable runtime scale monitoring for KEDA-based triggers.
When enabled, the Functions runtime can directly monitor event
sources to make more accurate scaling decisions.

### spec.siteConfig.websocketsEnabled

`bool` · optional (explicit presence)

Enable WebSocket connections.

Default: false

- default: `false`

### spec.siteConfig.healthCheckPath

`string`

Health check endpoint path.
Azure periodically sends requests to this path and marks the
instance as unhealthy if it doesn't respond with a 200-299 status
code. Requires health_check_eviction_time_in_min (paired both ways,
exactly as Azure pairs them). Common paths: "/api/health".

### spec.siteConfig.healthCheckEvictionTimeInMin

`int32` · optional (explicit presence)

Time in minutes after which a continuously-unhealthy instance is
evicted from the load balancer. Range 2-10. Requires
health_check_path. Travels as the
WEBSITE_HEALTHCHECK_MAXPINGFAILURES app setting.

- rule: {"int32":{"lte":10,"gte":2}}

### spec.siteConfig.workerCount

`int32` · optional (explicit presence)

The number of workers for this app. Range 1-100. Unset lets the
platform manage it (instance scaling on this hosting model is
driven by the spec's maximum_instance_count and always_ready).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.siteConfig.minimumTlsVersion

`enum`

Minimum TLS version for incoming HTTPS requests. Unset deploys
TLS_1_2 (the industry floor; TLS_1_3 for maximum security).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_tls_version_unspecified` -- Not specified -- deploys TLS_1_2.
- `TLS_1_0` -- TLS 1.0 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_1` -- TLS 1.1 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_2` -- TLS 1.2 -- the industry floor (the default).
- `TLS_1_3` -- TLS 1.3 -- the strongest option; requires modern clients.

### spec.siteConfig.scmMinimumTlsVersion

`enum`

Minimum TLS version for the SCM (Kudu) site. Unset deploys TLS_1_2.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_tls_version_unspecified` -- Not specified -- deploys TLS_1_2.
- `TLS_1_0` -- TLS 1.0 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_1` -- TLS 1.1 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_2` -- TLS 1.2 -- the industry floor (the default).
- `TLS_1_3` -- TLS 1.3 -- the strongest option; requires modern clients.

### spec.siteConfig.cors

`AzureFunctionAppFlexConsumptionCorsSettings`

CORS (Cross-Origin Resource Sharing) configuration.
Controls which origins are allowed to make cross-origin requests to
the Function App's HTTP endpoints.

- rule: support_credentials cannot be used with a wildcard '*' origin

### spec.siteConfig.cors.allowedOrigins

`[]string` · required

List of origins allowed to make cross-origin requests.
Use "*" to allow all origins (not recommended for production).
Example: ["https://myapp.example.com", "https://admin.example.com"]

- rule: {"repeated":{"minItems":"1"}}

### spec.siteConfig.cors.supportCredentials

`bool` · optional (explicit presence)

Allow credentials (cookies, authorization headers) in cross-origin
requests. Cannot be combined with a wildcard origin (enforced here
-- browsers reject the pairing).

Default: false

- default: `false`

### spec.siteConfig.vnetRouteAllEnabled

`bool` · optional (explicit presence)

Route all outbound traffic from the Function App through the VNet.
Requires virtual_network_subnet_id on the spec (spec-enforced).

When false (default), only RFC1918 traffic routes through the VNet.
When true, all outbound traffic (including public internet) routes
through the VNet, enabling inspection via NSG rules or a firewall.

Default: false

- default: `false`

### spec.appSettings

`map<string, string>`

Application settings (environment variables) for the Function App.
Key-value pairs available to functions at runtime.

Azure automatically manages several settings (AzureWebJobsStorage,
DEPLOYMENT_STORAGE_CONNECTION_STRING, the App Insights keys) --
user-provided settings are merged with these system settings. Auth
provider secrets referenced by auth_settings_v2 setting names also
live here.

### spec.connectionStrings

`[]AzureFunctionAppFlexConsumptionConnectionString`

Named connection strings for database and service connections.
Each connection string has a name, type, and value. The type
determines how Azure exposes the connection in the runtime
environment.

### spec.connectionStrings[].name

`string` · required

The name of the connection string.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.connectionStrings[].type

`enum` · required

The service type. Determines the environment variable prefix Azure
exposes the value under (e.g. SQLAZURECONNSTR_, MYSQLCONNSTR_,
CUSTOMCONNSTR_).

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_connection_string_type_unspecified` -- Not specified -- invalid; pick the service type (CUSTOM for anything without a dedicated type).
- `MYSQL`
- `SQL_SERVER`
- `SQL_AZURE`
- `CUSTOM`
- `NOTIFICATION_HUB`
- `SERVICE_BUS`
- `EVENT_HUB`
- `API_HUB`
- `DOC_DB`
- `REDIS_CACHE`
- `POSTGRESQL`

### spec.connectionStrings[].value

`string | valueFrom` · required · sensitive

The connection string value. This is a sensitive credential.
Can be a literal value or a reference to a secrets manager.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.stickySettings

`AzureFunctionAppFlexConsumptionStickySettings`

Settings pinned to the production slot during slot swaps.
Named app settings and connection strings listed here do NOT move
with the app content when a staging slot is swapped into production.

- rule: sticky_settings requires at least one app_setting_names or connection_string_names entry

### spec.stickySettings.appSettingNames

`[]string`

Names of app_settings entries that stay with the production slot
during a swap.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.stickySettings.connectionStringNames

`[]string`

Names of connection_strings entries that stay with the production
slot during a swap.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.applicationInsightsConnectionString

`string | valueFrom`

Application Insights connection string for APM telemetry.
When provided, Azure automatically configures the Function App to
send telemetry (requests, dependencies, exceptions, traces) to
Application Insights. This is the recommended way to monitor
Function Apps (the legacy instrumentation key lives in site_config).

- references: AzureApplicationInsights (`status.outputs.connection_string`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationInsights, name: <that resource's name>, fieldPath: status.outputs.connection_string}} -- a bare string does not parse

### spec.httpsOnly

`bool` · optional (explicit presence)

Enforce HTTPS-only access to the Function App.
When true, all HTTP requests are redirected to HTTPS.

Default: true (secure by default, unlike the Azure API default of false)

- default: `true`

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Enable or disable public network access to the Function App.
When false, the Function App is only accessible via Private
Endpoints.

Default: true

- default: `true`

### spec.enabled

`bool` · optional (explicit presence)

Enable or disable the Function App.
When false, the app is stopped and does not run functions, but the
resource still exists. Useful for temporarily disabling an app
without deleting it.

Default: true

- default: `true`

### spec.clientCertificateEnabled

`bool` · optional (explicit presence)

Enable client certificate authentication (mutual TLS).
When true, clients must present a valid certificate to access the
app (subject to client_certificate_mode).

Default: false

- default: `false`

### spec.clientCertificateMode

`enum`

Client certificate mode -- how strictly client certificates are
enforced when client_certificate_enabled is true. Unset deploys
OPTIONAL (certificate requested but not required).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_client_certificate_mode_unspecified` -- Not specified -- deploys OPTIONAL.
- `REQUIRED` -- All requests must present a valid client certificate.
- `OPTIONAL` -- Certificate is requested but not required.
- `OPTIONAL_INTERACTIVE_USER` -- Certificate is optional for interactive (browser) users, required for non-interactive clients.

### spec.clientCertificateExclusionPaths

`string`

Paths excluded from client certificate validation.
Semicolon-separated list of paths where client certificates are not
required. Example: "/api/health;/api/status"

### spec.virtualNetworkSubnetId

`string | valueFrom`

The subnet ID for VNet integration. When provided, the Function
App's outbound traffic routes through this subnet, enabling access
to VNet-connected resources (databases, Redis, etc.) without public
endpoints. The subnet must be delegated to Microsoft.App/environments
for Flex Consumption apps.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.identity

`AzureFunctionAppFlexConsumptionIdentity`

Managed identity configuration for the Function App.
Enables the app to authenticate with Azure services (Key Vault,
Storage, ACR, etc.) without managing credentials. Identity-based
deployment storage authentication also rides on this block.

- rule: identity_ids is required when type includes USER_ASSIGNED, and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

The identity model: Azure-managed (tied to the app's lifecycle),
bring-your-own (independent lifecycle), or both.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_identity_type_unspecified` -- Not specified -- invalid; pick an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates and rotates an identity tied to the app's lifecycle.
- `USER_ASSIGNED` -- Attach pre-created AzureUserAssignedIdentity resources (independent lifecycle; shareable across apps).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned identity and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

User Assigned Identity Azure resource IDs.
Required when type includes USER_ASSIGNED.

Can be literal ARM resource IDs or references to
AzureUserAssignedIdentity outputs.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.webdeployPublishBasicAuthenticationEnabled

`bool` · optional (explicit presence)

Allow basic-auth (username/password) publishing over Web Deploy
(msdeploy). Disabling it closes the classic credential-based
deployment path and forces identity-based deployment (recommended
posture for locked-down environments).

Default: true (Azure's own default; flip to false to harden)

- default: `true`

### spec.zipDeployFile

`string`

Path to a local ZIP package to deploy on create/update (one-shot
zip deploy through the publishing endpoint). Primarily useful for
simple pipelines that produce a build artifact next to the manifest;
most production deployments push code through CI/CD instead.
Write-only: Azure never returns it, so imports cannot recover it.

### spec.authSettingsV2

`AzureFunctionAppFlexConsumptionAuthSettingsV2`

App Service built-in authentication (Easy Auth v2). When enabled,
Azure authenticates requests at the platform layer -- before they
reach function code -- against any of the configured identity
providers. Provider client secrets are referenced by APP SETTING
NAME (set the actual secret value in app_settings or via a Key
Vault reference), never inline in this block.

- rule: forward_proxy_custom_host_header_name and forward_proxy_custom_scheme_header_name require forward_proxy_convention FORWARD_PROXY_CUSTOM

### spec.authSettingsV2.authEnabled

`bool` · optional (explicit presence)

Master switch: whether the authentication/authorization layer
intercepts requests.

Default: false

- default: `false`

### spec.authSettingsV2.runtimeVersion

`string` · optional (explicit presence)

The Easy Auth middleware runtime version.

Default: "~1"

- default: `~1`

### spec.authSettingsV2.configFilePath

`string`

Path to a config file carrying the auth settings when they are
file-managed instead of ARM-managed. Rarely used.

### spec.authSettingsV2.requireAuthentication

`bool` · optional (explicit presence)

Require every request to be authenticated (subject to
excluded_paths).

Default: false

- default: `false`

### spec.authSettingsV2.unauthenticatedAction

`enum`

What happens to unauthenticated requests. Unset deploys
REDIRECT_TO_LOGIN_PAGE.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_unauthenticated_action_unspecified` -- Not specified -- deploys REDIRECT_TO_LOGIN_PAGE.
- `REDIRECT_TO_LOGIN_PAGE` -- Redirect to the login page of the default provider.
- `ALLOW_ANONYMOUS` -- Let the request through; the app decides (identity headers are populated when present).
- `RETURN_401` -- Reject with HTTP 401.
- `RETURN_403` -- Reject with HTTP 403.

### spec.authSettingsV2.defaultProvider

`string`

The provider unauthenticated requests are redirected to when
unauthenticated_action is REDIRECT_TO_LOGIN_PAGE and more than one
provider is configured: "apple", "azureactivedirectory", "facebook",
"github", "google", "twitter", or the name of a custom_oidc_v2
provider.

### spec.authSettingsV2.excludedPaths

`[]string`

Paths that skip authentication entirely (e.g. ["/api/health"]).

### spec.authSettingsV2.requireHttps

`bool` · optional (explicit presence)

Require HTTPS for authentication requests.

Default: true

- default: `true`

### spec.authSettingsV2.httpRouteApiPrefix

`string` · optional (explicit presence)

The prefix that the Easy Auth HTTP endpoints (login, logout, token
refresh) are served under.

Default: "/.auth"

- default: `/.auth`

### spec.authSettingsV2.forwardProxyConvention

`enum`

How the original request URL is derived when the app sits behind a
forward proxy. Unset deploys FORWARD_PROXY_NO_PROXY.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_forward_proxy_convention_unspecified` -- Not specified -- deploys FORWARD_PROXY_NO_PROXY.
- `FORWARD_PROXY_NO_PROXY` -- No proxy: use the request as received (the default).
- `FORWARD_PROXY_STANDARD` -- Standard X-Forwarded-Host/X-Forwarded-Proto headers.
- `FORWARD_PROXY_CUSTOM` -- Custom header names (set the two custom header name fields).

### spec.authSettingsV2.forwardProxyCustomHostHeaderName

`string`

The header carrying the original host, when forward_proxy_convention
is FORWARD_PROXY_CUSTOM.

### spec.authSettingsV2.forwardProxyCustomSchemeHeaderName

`string`

The header carrying the original scheme, when
forward_proxy_convention is FORWARD_PROXY_CUSTOM.

### spec.authSettingsV2.login

`AzureFunctionAppFlexConsumptionAuthV2Login` · required

Login/session behavior (token store, cookie expiration, nonce).
Required by Azure whenever auth_settings_v2 is configured.

- rule: {"required":true}
- rule: token_store_path and token_store_sas_setting_name are mutually exclusive

### spec.authSettingsV2.login.logoutEndpoint

`string`

The endpoint the browser is sent to on logout (clears the session).

### spec.authSettingsV2.login.tokenStoreEnabled

`bool` · optional (explicit presence)

Durably store identity tokens so the app (and the /.auth/me
endpoint) can retrieve them later. Required for token refresh.

Default: false

- default: `false`

### spec.authSettingsV2.login.tokenRefreshExtensionTime

`double` · optional (explicit presence)

Hours after session expiry that a token refresh may still succeed.

Default: 72

- default: `72`
- rule: {"double":{"gte":0}}

### spec.authSettingsV2.login.tokenStorePath

`string`

File-system path backing the token store (mutually exclusive with
the SAS-setting form).

### spec.authSettingsV2.login.tokenStoreSasSettingName

`string`

Name of the app setting holding the SAS URL of the blob container
backing the token store (mutually exclusive with the path form).

### spec.authSettingsV2.login.preserveUrlFragmentsForLogins

`bool` · optional (explicit presence)

Preserve URL fragments (#...) across the login redirect dance.

Default: false

- default: `false`

### spec.authSettingsV2.login.allowedExternalRedirectUrls

`[]string`

External URLs that post-login/logout redirects may target (in
addition to same-host URLs, which are always allowed).

### spec.authSettingsV2.login.cookieExpirationConvention

`enum`

How session cookies expire. Unset deploys FIXED_TIME.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_flex_consumption_cookie_expiration_convention_unspecified` -- Not specified -- deploys FIXED_TIME.
- `FIXED_TIME` -- Cookies expire after cookie_expiration_time (the default).
- `IDENTITY_PROVIDER_DERIVED` -- Cookie lifetime follows the identity provider's token lifetime.

### spec.authSettingsV2.login.cookieExpirationTime

`string` · optional (explicit presence)

Session cookie lifetime (hh:mm:ss) when the convention is
FIXED_TIME.

Default: "08:00:00"

- default: `08:00:00`
- rule: cookie_expiration_time must be hh:mm:ss (e.g. 08:00:00)

### spec.authSettingsV2.login.validateNonce

`bool` · optional (explicit presence)

Validate the anti-forgery nonce during the login flow. Leave on.

Default: true

- default: `true`

### spec.authSettingsV2.login.nonceExpirationTime

`string` · optional (explicit presence)

Nonce lifetime (hh:mm:ss).

Default: "00:05:00"

- default: `00:05:00`
- rule: nonce_expiration_time must be hh:mm:ss (e.g. 00:05:00)

### spec.authSettingsV2.appleV2

`AzureFunctionAppFlexConsumptionAuthV2Apple`

Sign in with Apple.

### spec.authSettingsV2.appleV2.clientId

`string` · required

The Apple Services ID (client ID).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.appleV2.clientSecretSettingName

`string` · required

Name of the app setting holding the client secret (never the secret
itself).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.activeDirectoryV2

`AzureFunctionAppFlexConsumptionAuthV2ActiveDirectory`

Microsoft Entra ID (Azure Active Directory).

- rule: client_secret_setting_name and client_secret_certificate_thumbprint are mutually exclusive

### spec.authSettingsV2.activeDirectoryV2.clientId

`string` · required

The Entra app registration's client ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.activeDirectoryV2.tenantAuthEndpoint

`string` · required

The OpenID issuer endpoint for the tenant, e.g.
https://login.microsoftonline.com/v2.0/{tenant-guid}/

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.activeDirectoryV2.clientSecretSettingName

`string`

Name of the app setting holding the client secret (mutually
exclusive with the certificate thumbprint).

### spec.authSettingsV2.activeDirectoryV2.clientSecretCertificateThumbprint

`string`

Thumbprint of the certificate used as the client credential
(mutually exclusive with the secret setting name).

### spec.authSettingsV2.activeDirectoryV2.loginParameters

`map<string, string>`

Extra parameters sent to the authorization endpoint on login
(key=value form).

### spec.authSettingsV2.activeDirectoryV2.wwwAuthenticationDisabled

`bool` · optional (explicit presence)

Suppress the WWW-Authenticate challenge header on 401 responses.

Default: false

- default: `false`

### spec.authSettingsV2.activeDirectoryV2.jwtAllowedGroups

`[]string`

Group claim values the JWT must carry to be accepted.

### spec.authSettingsV2.activeDirectoryV2.jwtAllowedClientApplications

`[]string`

Client-application claim values the JWT must carry to be accepted.

### spec.authSettingsV2.activeDirectoryV2.allowedGroups

`[]string`

Entra group object IDs allowed access.

### spec.authSettingsV2.activeDirectoryV2.allowedIdentities

`[]string`

Identity object IDs allowed access.

### spec.authSettingsV2.activeDirectoryV2.allowedApplications

`[]string`

Client (application) IDs allowed access.

### spec.authSettingsV2.activeDirectoryV2.allowedAudiences

`[]string`

Token audiences accepted in addition to the client ID.

### spec.authSettingsV2.azureStaticWebAppV2

`AzureFunctionAppFlexConsumptionAuthV2StaticWebApp`

Azure Static Web Apps authentication (when fronted by one).

### spec.authSettingsV2.azureStaticWebAppV2.clientId

`string` · required

The Static Web App's client ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.customOidcV2

`[]AzureFunctionAppFlexConsumptionAuthV2CustomOidc`

Any OpenID Connect provider(s), by name.

### spec.authSettingsV2.customOidcV2[].name

`string` · required

The provider's name (also the login route segment and the
default_provider value that selects it).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.customOidcV2[].clientId

`string` · required

The provider's client ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.customOidcV2[].openidConfigurationEndpoint

`string` · required

The provider's OpenID configuration (discovery) endpoint, e.g.
https://idp.example.com/.well-known/openid-configuration

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.customOidcV2[].nameClaimType

`string`

The claim carrying the user's display name.

### spec.authSettingsV2.customOidcV2[].scopes

`[]string`

OAuth scopes requested at login.

### spec.authSettingsV2.facebookV2

`AzureFunctionAppFlexConsumptionAuthV2Facebook`

Facebook login.

### spec.authSettingsV2.facebookV2.appId

`string` · required

The Facebook app's App ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.facebookV2.appSecretSettingName

`string` · required

Name of the app setting holding the app secret (never the secret
itself).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.facebookV2.graphApiVersion

`string`

The Facebook Graph API version used for login (e.g. "v17.0").

### spec.authSettingsV2.facebookV2.loginScopes

`[]string`

OAuth scopes requested at login.

### spec.authSettingsV2.githubV2

`AzureFunctionAppFlexConsumptionAuthV2Github`

GitHub login.

### spec.authSettingsV2.githubV2.clientId

`string` · required

The GitHub OAuth app's client ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.githubV2.clientSecretSettingName

`string` · required

Name of the app setting holding the client secret (never the secret
itself).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.githubV2.loginScopes

`[]string`

OAuth scopes requested at login.

### spec.authSettingsV2.googleV2

`AzureFunctionAppFlexConsumptionAuthV2Google`

Google login.

### spec.authSettingsV2.googleV2.clientId

`string` · required

The Google OAuth client ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.googleV2.clientSecretSettingName

`string` · required

Name of the app setting holding the client secret (never the secret
itself).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.googleV2.allowedAudiences

`[]string`

Token audiences accepted in addition to the client ID.

### spec.authSettingsV2.googleV2.loginScopes

`[]string`

OAuth scopes requested at login.

### spec.authSettingsV2.microsoftV2

`AzureFunctionAppFlexConsumptionAuthV2Microsoft`

Microsoft account (consumer) login.

### spec.authSettingsV2.microsoftV2.clientId

`string` · required

The Microsoft app registration's client ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.microsoftV2.clientSecretSettingName

`string` · required

Name of the app setting holding the client secret (never the secret
itself).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.microsoftV2.allowedAudiences

`[]string`

Token audiences accepted in addition to the client ID.

### spec.authSettingsV2.microsoftV2.loginScopes

`[]string`

OAuth scopes requested at login.

### spec.authSettingsV2.twitterV2

`AzureFunctionAppFlexConsumptionAuthV2Twitter`

Twitter login.

### spec.authSettingsV2.twitterV2.consumerKey

`string` · required

The Twitter app's consumer (API) key.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.twitterV2.consumerSecretSettingName

`string` · required

Name of the app setting holding the consumer secret (never the
secret itself).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tags

`map<string, string>`

Free-form Azure resource tags applied to the Function App, merged
over the platform's metadata-derived tags (user tags win on key
collision) -- the hooks for cost allocation, chargeback reports, and
Azure Policy governance rules that filter or group by them.
Updatable in place.

## Validation Rules

- `flex_storage_key_required_for_connection_string`: storage_access_key is required when storage_authentication_type is STORAGE_ACCOUNT_CONNECTION_STRING
- `flex_storage_uami_required_for_user_assigned`: storage_user_assigned_identity_id is required when storage_authentication_type is USER_ASSIGNED_IDENTITY
- `flex_vnet_route_all_requires_subnet`: site_config.vnet_route_all_enabled requires virtual_network_subnet_id (there is no VNet to route through otherwise)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureFunctionAppFlexConsumption, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.function_app_id` | `string` | The Azure Resource Manager ID of the Function App. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Web/sites/{name} |
| `status.outputs.default_hostname` | `string` | The default hostname of the Function App. Format: {name}.azurewebsites.net This is the primary endpoint for HTTP-triggered functions. Custom domains can be added via Azure portal or DNS configuration. |
| `status.outputs.outbound_ip_addresses` | `[]string` | Outbound IP addresses used by the Function App. These IPs should be allowed in downstream firewall rules (e.g., database firewall, third-party API whitelists). Note: outbound IPs on serverless plans are shared across the region and may change -- use possible_outbound_ip_addresses for allowlists that must survive scale events. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the system-assigned managed identity. Populated only when the app has a system-assigned identity (identity.type includes "SystemAssigned"). Used for granting RBAC roles: e.g., "Storage Blob Data Contributor" (identity-based deployment storage), "Key Vault Secrets User". |
| `status.outputs.identity_tenant_id` | `string` | The tenant ID of the system-assigned managed identity. Paired with identity_principal_id for RBAC configuration. |
| `status.outputs.custom_domain_verification_id` | `string` | The custom domain verification ID. Used when binding custom domains to the Function App. Add this value as a TXT record at `asuid.{custom-domain}` to verify domain ownership. |
| `status.outputs.kind` | `string` | The resource kind string as reported by Azure. Example: "functionapp,linux" |
| `status.outputs.possible_outbound_ip_addresses` | `[]string` | Every outbound IP address the platform could EVER route this app's traffic through (a superset of outbound_ip_addresses, which lists only the currently active set). Use THIS list for downstream firewall allowlists that must survive scale events and platform moves. |
| `status.outputs.site_credential_name` | `string` | The site-level publishing credential's username (the Kudu/SCM basic-auth user). Paired with site_credential_password; only usable while webdeploy_publish_basic_authentication_enabled is true. |
| `status.outputs.site_credential_password` | `string` (sensitive) | The site-level publishing credential's password. SECRET-BEARING: anyone holding it can deploy code to the app over Web Deploy/SCM while basic-auth publishing is enabled -- treat it like an admin password (disable the basic-auth toggle to revoke the surface entirely). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.servicePlanId` | AzureServicePlan | `status.outputs.service_plan_id` |
| `spec.storageAccessKey` | AzureStorageAccount | `status.outputs.primary_access_key` |
| `spec.storageUserAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.siteConfig.ipRestrictions[].virtualNetworkSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.siteConfig.ipRestrictions[].headers.xAzureFdid` | AzureFrontDoorProfile | `status.outputs.resource_guid` |
| `spec.siteConfig.scmIpRestrictions[].virtualNetworkSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.siteConfig.scmIpRestrictions[].headers.xAzureFdid` | AzureFrontDoorProfile | `status.outputs.resource_guid` |
| `spec.applicationInsightsConnectionString` | AzureApplicationInsights | `status.outputs.connection_string` |
| `spec.virtualNetworkSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## See Also

- [Overview](../README.md)

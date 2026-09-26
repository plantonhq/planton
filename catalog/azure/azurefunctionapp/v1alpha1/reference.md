# AzureFunctionApp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureFunctionAppSpec** defines the configuration for creating an Azure
Linux Function App (Microsoft.Web/sites kind=functionapp,linux), a
serverless compute platform for event-driven workloads.

A Function App hosts functions triggered by HTTP requests, queue
messages, timer schedules, blob storage events, and other Azure service
events. Azure Functions run on an App Service Plan, giving users control
over the compute tier (Consumption for pay-per-execution, Elastic
Premium for pre-warmed instances, or Dedicated for reserved capacity).

**Relationship to AzureServicePlan**:

Every Function App requires a Service Plan. The plan determines cost model,
scale behavior, and available features:

Consumption (Y1):
  - Pay-per-execution, scales to 0 when idle
  - Auto-scales up to 200 instances
  - 5-minute default timeout (configurable to 10 min)
  - Cold start latency on first request

Elastic Premium (EP1-EP3):
  - Pre-warmed instances eliminate cold starts
  - Scales up to 100 instances (configurable via app_scale_limit)
  - VNet integration and private endpoints supported
  - Unlimited execution duration

Dedicated (B*/S*/P*):
  - Fixed instance count (manual or auto-scale)
  - `always_on` should be true to prevent idle shutdown
  - Full App Service features (custom domains, SSL, etc.)

Azure also offers Flex Consumption (FC1 plans) as a SEPARATE resource
type with its own storage and scaling model -- it is not this kind.

**Storage requirement**:

Every Function App requires an Azure Storage Account for runtime state
(function triggers, logs, queue management). Exactly one binding form
applies: `storage_account_name` (with an access key or managed
identity), or `storage_key_vault_secret_id` (a Key Vault secret holding
the connection string). The storage account must exist before the
Function App is created.

**Application stack**:

The application_stack within site_config selects the runtime. Exactly one
runtime must be chosen: .NET, Node.js, Python, Java, PowerShell, Docker
container, or custom handler. Docker enables running custom container images
as Azure Functions.

**Authentication (Easy Auth)**: the `auth_settings_v2` block turns on
App Service's built-in authentication layer -- Azure validates identity
tokens at the front door (Entra ID, Apple, Facebook, GitHub, Google,
Microsoft account, Twitter, or any OpenID Connect provider) before
requests reach function code. Provider secrets are referenced by APP
SETTING NAME (never inline).

**Windows Function Apps** (`azurerm_windows_function_app`) are
deliberately not modeled: the platform targets Linux-first runtimes and
containers. The `AzureServicePlan` kind supports Windows plans, so the
compute tier is ready if a Windows app kind is ever added.

**ForceNew fields** (changing these destroys and recreates the function
app): `function_app_name`, `region`, `resource_group`.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFunctionApp
metadata:
  name: test-fn
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  functionAppName: planton-hack-function-app
  servicePlanId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Web/serverFarms/plan
  storageAccountName:
    value: teststorage
  storageAccountAccessKey:
    value: dGVzdC1rZXk=
  # The Consumption cost circuit breaker exercises the quota seam.
  dailyMemoryTimeQuota: 500000
  # Basic-auth publishing disabled exercises the hardened posture.
  webdeployPublishBasicAuthenticationEnabled: false
  ftpPublishBasicAuthenticationEnabled: false
  identity:
    type: SYSTEM_ASSIGNED
  stickySettings:
    appSettingNames:
      - SLOT_MARKER
  appSettings:
    SLOT_MARKER: production
  connectionStrings:
    - name: ServiceBus
      type: SERVICE_BUS
      value:
        value: Endpoint=sb://ns.servicebus.windows.net/;SharedAccessKeyName=key
  siteConfig:
    applicationStack:
      pythonVersion: "3.12"
    healthCheckPath: /api/health
    healthCheckEvictionTimeInMin: 5
    # Enum-valued dials exercise the wire maps on both engines.
    minimumTlsVersion: TLS_1_2
    ftpsState: DISABLED
    loadBalancingMode: LEAST_REQUESTS
    managedPipelineMode: INTEGRATED
    http2Enabled: true
    appScaleLimit: 50
    ipRestrictions:
      # The Front Door origin lockdown: allow only the AzureFrontDoor.Backend
      # ranges AND filter by the profile's FDID GUID (a reference-or-literal;
      # the literal form exercised here) so only YOUR profile's traffic
      # reaches the origin.
      - name: front-door-only
        priority: 100
        action: ALLOW
        serviceTag: AzureFrontDoor.Backend
        headers:
          xAzureFdid:
            - value: 11111111-2222-3333-4444-555555555555
      - name: office
        priority: 200
        action: ALLOW
        ipAddress: 203.0.113.0/24
    ipRestrictionDefaultAction: DENY
    appServiceLogs:
      diskQuotaMb: 50
      retentionPeriodDays: 7
  # Easy Auth v2 with a custom OIDC provider exercises the auth block
  # end to end without needing a real Entra app registration at plan time.
  authSettingsV2:
    authEnabled: true
    requireAuthentication: true
    unauthenticatedAction: RETURN_401
    excludedPaths:
      - /api/health
    login:
      tokenStoreEnabled: true
    customOidcV2:
      - name: corp-idp
        clientId: fn-client
        openidConfigurationEndpoint: https://idp.example.com/.well-known/openid-configuration
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.functionAppName` | `string` | yes |  |  |
| `spec.servicePlanId` | `string \| valueFrom` | yes |  | AzureServicePlan (`status.outputs.service_plan_id`) |
| `spec.storageAccountName` | `string \| valueFrom` |  |  | AzureStorageAccount (`status.outputs.storage_account_name`) |
| `spec.storageAccountAccessKey` | `string \| valueFrom` (sensitive) |  |  | AzureStorageAccount (`status.outputs.primary_access_key`) |
| `spec.storageUsesManagedIdentity` | `bool` |  | `false` |  |
| `spec.storageKeyVaultSecretId` | `string` |  |  |  |
| `spec.functionsExtensionVersion` | `string` |  | `~4` |  |
| `spec.dailyMemoryTimeQuota` | `int32` |  | `0` |  |
| `spec.siteConfig` | `AzureFunctionAppSiteConfig` | yes |  |  |
| `spec.siteConfig.applicationStack` | `AzureFunctionAppApplicationStack` |  |  |  |
| `spec.siteConfig.applicationStack.dotnetVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.useDotnetIsolatedRuntime` | `bool` |  | `false` |  |
| `spec.siteConfig.applicationStack.nodeVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.pythonVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.javaVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.powershellCoreVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.docker` | `AzureFunctionAppDockerConfig` |  |  |  |
| `spec.siteConfig.applicationStack.docker.registryUrl` | `string` | yes |  |  |
| `spec.siteConfig.applicationStack.docker.imageName` | `string` | yes |  |  |
| `spec.siteConfig.applicationStack.docker.imageTag` | `string` | yes |  |  |
| `spec.siteConfig.applicationStack.docker.registryUsername` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.docker.registryPassword` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.siteConfig.applicationStack.useCustomRuntime` | `bool` |  |  |  |
| `spec.siteConfig.alwaysOn` | `bool` |  |  |  |
| `spec.siteConfig.appCommandLine` | `string` |  |  |  |
| `spec.siteConfig.apiManagementApiId` | `string` |  |  |  |
| `spec.siteConfig.apiDefinitionUrl` | `string` |  |  |  |
| `spec.siteConfig.defaultDocuments` | `[]string` |  |  |  |
| `spec.siteConfig.healthCheckPath` | `string` |  |  |  |
| `spec.siteConfig.healthCheckEvictionTimeInMin` | `int32` |  |  |  |
| `spec.siteConfig.minimumTlsVersion` | `enum` |  |  |  |
| `spec.siteConfig.scmMinimumTlsVersion` | `enum` |  |  |  |
| `spec.siteConfig.minimumTlsCipherSuite` | `string` |  |  |  |
| `spec.siteConfig.appScaleLimit` | `int32` |  |  |  |
| `spec.siteConfig.elasticInstanceMinimum` | `int32` |  |  |  |
| `spec.siteConfig.preWarmedInstanceCount` | `int32` |  |  |  |
| `spec.siteConfig.workerCount` | `int32` |  |  |  |
| `spec.siteConfig.http2Enabled` | `bool` |  | `false` |  |
| `spec.siteConfig.websocketsEnabled` | `bool` |  | `false` |  |
| `spec.siteConfig.use32BitWorker` | `bool` |  | `false` |  |
| `spec.siteConfig.vnetRouteAllEnabled` | `bool` |  | `false` |  |
| `spec.siteConfig.ftpsState` | `enum` |  |  |  |
| `spec.siteConfig.loadBalancingMode` | `enum` |  |  |  |
| `spec.siteConfig.managedPipelineMode` | `enum` |  |  |  |
| `spec.siteConfig.remoteDebuggingEnabled` | `bool` |  | `false` |  |
| `spec.siteConfig.runtimeScaleMonitoringEnabled` | `bool` |  |  |  |
| `spec.siteConfig.cors` | `AzureFunctionAppCorsSettings` |  |  |  |
| `spec.siteConfig.cors.allowedOrigins` | `[]string` | yes |  |  |
| `spec.siteConfig.cors.supportCredentials` | `bool` |  | `false` |  |
| `spec.siteConfig.ipRestrictions` | `[]AzureFunctionAppIpRestriction` |  |  |  |
| `spec.siteConfig.ipRestrictions[].name` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].priority` | `int32` |  |  |  |
| `spec.siteConfig.ipRestrictions[].action` | `enum` |  |  |  |
| `spec.siteConfig.ipRestrictions[].ipAddress` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].serviceTag` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].virtualNetworkSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.siteConfig.ipRestrictions[].description` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers` | `AzureFunctionAppIpRestrictionHeaders` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers.xForwardedFor` | `[]string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers.xForwardedHost` | `[]string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers.xAzureFdid` | `[]string \| valueFrom` |  |  | AzureFrontDoorProfile (`status.outputs.resource_guid`) |
| `spec.siteConfig.ipRestrictions[].headers.xFdHealthProbe` | `[]string` |  |  |  |
| `spec.siteConfig.ipRestrictionDefaultAction` | `enum` |  |  |  |
| `spec.siteConfig.scmUseMainIpRestriction` | `bool` |  | `false` |  |
| `spec.siteConfig.scmIpRestrictions` | `[]AzureFunctionAppIpRestriction` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].name` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].priority` | `int32` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].action` | `enum` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].ipAddress` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].serviceTag` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].virtualNetworkSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.siteConfig.scmIpRestrictions[].description` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers` | `AzureFunctionAppIpRestrictionHeaders` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers.xForwardedFor` | `[]string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers.xForwardedHost` | `[]string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers.xAzureFdid` | `[]string \| valueFrom` |  |  | AzureFrontDoorProfile (`status.outputs.resource_guid`) |
| `spec.siteConfig.scmIpRestrictions[].headers.xFdHealthProbe` | `[]string` |  |  |  |
| `spec.siteConfig.scmIpRestrictionDefaultAction` | `enum` |  |  |  |
| `spec.siteConfig.appServiceLogs` | `AzureFunctionAppAppServiceLogs` |  |  |  |
| `spec.siteConfig.appServiceLogs.diskQuotaMb` | `int32` |  | `35` |  |
| `spec.siteConfig.appServiceLogs.retentionPeriodDays` | `int32` |  |  |  |
| `spec.siteConfig.containerRegistryUseManagedIdentity` | `bool` |  | `false` |  |
| `spec.siteConfig.containerRegistryManagedIdentityClientId` | `string` |  |  |  |
| `spec.siteConfig.applicationInsightsKey` | `string` (sensitive) |  |  |  |
| `spec.appSettings` | `map<string, string>` |  |  |  |
| `spec.connectionStrings` | `[]AzureFunctionAppConnectionString` |  |  |  |
| `spec.connectionStrings[].name` | `string` | yes |  |  |
| `spec.connectionStrings[].type` | `enum` | yes |  |  |
| `spec.connectionStrings[].value` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.stickySettings` | `AzureFunctionAppStickySettings` |  |  |  |
| `spec.stickySettings.appSettingNames` | `[]string` |  |  |  |
| `spec.stickySettings.connectionStringNames` | `[]string` |  |  |  |
| `spec.applicationInsightsConnectionString` | `string \| valueFrom` |  |  | AzureApplicationInsights (`status.outputs.connection_string`) |
| `spec.httpsOnly` | `bool` |  | `true` |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.builtinLoggingEnabled` | `bool` |  | `true` |  |
| `spec.contentShareForceDisabled` | `bool` |  | `false` |  |
| `spec.clientCertificateEnabled` | `bool` |  | `false` |  |
| `spec.clientCertificateMode` | `enum` |  |  |  |
| `spec.clientCertificateExclusionPaths` | `string` |  |  |  |
| `spec.virtualNetworkSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.vnetImagePullEnabled` | `bool` |  | `false` |  |
| `spec.virtualNetworkBackupRestoreEnabled` | `bool` |  | `false` |  |
| `spec.identity` | `AzureFunctionAppIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.keyVaultReferenceIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.webdeployPublishBasicAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.ftpPublishBasicAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.zipDeployFile` | `string` |  |  |  |
| `spec.storageMounts` | `[]AzureFunctionAppStorageMount` |  |  |  |
| `spec.storageMounts[].name` | `string` | yes |  |  |
| `spec.storageMounts[].type` | `enum` | yes |  |  |
| `spec.storageMounts[].accountName` | `string` | yes |  |  |
| `spec.storageMounts[].shareName` | `string` | yes |  |  |
| `spec.storageMounts[].accessKey` | `string \| valueFrom` (sensitive) | yes |  | AzureStorageAccount (`status.outputs.primary_access_key`) |
| `spec.storageMounts[].mountPath` | `string` |  |  |  |
| `spec.backup` | `AzureFunctionAppBackup` |  |  |  |
| `spec.backup.name` | `string` | yes |  |  |
| `spec.backup.storageAccountUrl` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.backup.enabled` | `bool` |  | `true` |  |
| `spec.backup.schedule` | `AzureFunctionAppBackupSchedule` | yes |  |  |
| `spec.backup.schedule.frequencyInterval` | `int32` | yes |  |  |
| `spec.backup.schedule.frequencyUnit` | `enum` | yes |  |  |
| `spec.backup.schedule.keepAtLeastOneBackup` | `bool` |  | `false` |  |
| `spec.backup.schedule.retentionPeriodDays` | `int32` |  | `30` |  |
| `spec.backup.schedule.startTime` | `string` |  |  |  |
| `spec.authSettingsV2` | `AzureFunctionAppAuthSettingsV2` |  |  |  |
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
| `spec.authSettingsV2.login` | `AzureFunctionAppAuthV2Login` | yes |  |  |
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
| `spec.authSettingsV2.appleV2` | `AzureFunctionAppAuthV2Apple` |  |  |  |
| `spec.authSettingsV2.appleV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.appleV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.activeDirectoryV2` | `AzureFunctionAppAuthV2ActiveDirectory` |  |  |  |
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
| `spec.authSettingsV2.azureStaticWebAppV2` | `AzureFunctionAppAuthV2StaticWebApp` |  |  |  |
| `spec.authSettingsV2.azureStaticWebAppV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2` | `[]AzureFunctionAppAuthV2CustomOidc` |  |  |  |
| `spec.authSettingsV2.customOidcV2[].name` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2[].clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2[].openidConfigurationEndpoint` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2[].nameClaimType` | `string` |  |  |  |
| `spec.authSettingsV2.customOidcV2[].scopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.facebookV2` | `AzureFunctionAppAuthV2Facebook` |  |  |  |
| `spec.authSettingsV2.facebookV2.appId` | `string` | yes |  |  |
| `spec.authSettingsV2.facebookV2.appSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.facebookV2.graphApiVersion` | `string` |  |  |  |
| `spec.authSettingsV2.facebookV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.githubV2` | `AzureFunctionAppAuthV2Github` |  |  |  |
| `spec.authSettingsV2.githubV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.githubV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.githubV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.googleV2` | `AzureFunctionAppAuthV2Google` |  |  |  |
| `spec.authSettingsV2.googleV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.googleV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.googleV2.allowedAudiences` | `[]string` |  |  |  |
| `spec.authSettingsV2.googleV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.microsoftV2` | `AzureFunctionAppAuthV2Microsoft` |  |  |  |
| `spec.authSettingsV2.microsoftV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.microsoftV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.microsoftV2.allowedAudiences` | `[]string` |  |  |  |
| `spec.authSettingsV2.microsoftV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.twitterV2` | `AzureFunctionAppAuthV2Twitter` |  |  |  |
| `spec.authSettingsV2.twitterV2.consumerKey` | `string` | yes |  |  |
| `spec.authSettingsV2.twitterV2.consumerSecretSettingName` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Function App will be created.
Examples: "eastus", "westus2", "westeurope", "southeastasia".

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

Allowed characters: alphanumeric and hyphens.
Must start and end with an alphanumeric character.
Length: 2 to 60 characters.

**ForceNew**: Changing this destroys and recreates the function app.

- rule: function_app_name must contain only alphanumeric characters and hyphens, and must start and end with an alphanumeric character
- rule: {"required":true,"string":{"minLen":"2","maxLen":"60"}}

### spec.servicePlanId

`string | valueFrom` · required

The App Service Plan that provides compute resources for this Function App.
Determines the pricing tier, scale behavior, and available features.

Consumption (Y1): pay-per-execution, auto-scale to 200 instances
Elastic Premium (EP*): pre-warmed instances, up to 100
Dedicated (B*/S*/P*): fixed or auto-scaled instances

**Conditional ForceNew**: Changing between Dynamic tiers (Consumption
to/from any other tier) forces recreation.

- references: AzureServicePlan (`status.outputs.service_plan_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServicePlan, name: <that resource's name>, fieldPath: status.outputs.service_plan_id}} -- a bare string does not parse

### spec.storageAccountName

`string | valueFrom`

The name of the Azure Storage Account used for Function App runtime state.
Azure Functions use storage for trigger management, execution logs, and
internal coordination. The storage account must already exist.

Exactly one storage binding applies: this field OR
storage_key_vault_secret_id (enforced here, exactly as the provider
enforces it).

- references: AzureStorageAccount (`status.outputs.storage_account_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_name}} -- a bare string does not parse

### spec.storageAccountAccessKey

`string | valueFrom` · sensitive

The access key for the storage account. Defaults to referencing an
AzureStorageAccount's primary_access_key output, so the binding
composes in one manifest set; a literal value or a managed-secret
reference works too. Prefer storage_uses_managed_identity where the
workload supports it -- keys are static credential material.

Conflicts with storage_uses_managed_identity (enforced here).

- references: AzureStorageAccount (`status.outputs.primary_access_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_access_key}} -- a bare string does not parse

### spec.storageUsesManagedIdentity

`bool` · optional (explicit presence)

Use the Function App's managed identity to access the storage account
instead of an access key. This is the modern, credential-free approach.

When true, the Function App's system-assigned or user-assigned identity
must have Storage Blob Data Owner and Storage Queue Data Contributor
roles on the storage account.

Conflicts with storage_account_access_key (enforced here).

Default: false

- default: `false`

### spec.storageKeyVaultSecretId

`string`

The Key Vault secret ID holding the storage account's CONNECTION
STRING -- the alternative storage binding for vault-managed
credentials. Format:
https://{vault}.vault.azure.net/secrets/{secret}[/{version}]
(an unversioned ID tracks the latest secret version; pair with
key_vault_reference_identity_id or the system-assigned identity for
vault access).

Exactly one storage binding applies: this field OR
storage_account_name (enforced here).

- rule: storage_key_vault_secret_id must be a Key Vault secret URL (https://{vault}.vault.azure.net/secrets/{name} with an optional /{version})

### spec.functionsExtensionVersion

`string` · optional (explicit presence)

The Azure Functions runtime version.
Controls which version of the Azure Functions host runs the app.

Common values: "~4" (current LTS), "~3" (legacy, EOL).
The "~" prefix enables automatic minor version updates.

Default: "~4"

- default: `~4`

### spec.dailyMemoryTimeQuota

`int32` · optional (explicit presence)

Daily compute quota in GB-seconds -- the Consumption-plan cost
circuit breaker. When the app's aggregate compute crosses the quota,
Azure stops it until the next day. 0 (the default) means unlimited.
Only meaningful on Consumption (Y1) plans.

- default: `0`
- rule: {"int32":{"gte":0}}

### spec.siteConfig

`AzureFunctionAppSiteConfig` · required

Site configuration for the Function App.
Contains the application stack (runtime), scaling settings, security
settings, and operational configuration.

- rule: {"required":true}
- rule: health_check_eviction_time_in_min requires health_check_path (there is no probe to evict on otherwise)

### spec.siteConfig.applicationStack

`AzureFunctionAppApplicationStack`

The application stack defines the runtime for the Function App.
Exactly one runtime must be specified: dotnet_version, node_version,
python_version, java_version, powershell_core_version, docker, or
use_custom_runtime.

### spec.siteConfig.applicationStack.dotnetVersion

`string`

.NET runtime version.
Uses the dotnet-isolated worker model by default in Functions v4+.

Valid values: "3.1", "6.0", "7.0", "8.0", "9.0", "10.0"

Note: .NET 6+ uses isolated worker by default. Set
use_dotnet_isolated_runtime = false only for legacy in-process apps.

- rule: dotnet_version must be one of: 3.1, 6.0, 7.0, 8.0, 9.0, 10.0

### spec.siteConfig.applicationStack.useDotnetIsolatedRuntime

`bool` · optional (explicit presence)

Use the .NET isolated worker runtime model.
The isolated model runs functions in a separate process from the host,
enabling different .NET versions and full dependency control.

Default: false (in-process model for backward compatibility, though
Microsoft recommends isolated for new apps)

- default: `false`

### spec.siteConfig.applicationStack.nodeVersion

`string`

Node.js runtime version.

Valid values: "12", "14", "16", "18", "20", "22", "24"

- rule: node_version must be one of: 12, 14, 16, 18, 20, 22, 24

### spec.siteConfig.applicationStack.pythonVersion

`string`

Python runtime version.

Valid values: "3.8", "3.9", "3.10", "3.11", "3.12", "3.13", "3.14"

- rule: python_version must be one of: 3.8, 3.9, 3.10, 3.11, 3.12, 3.13, 3.14

### spec.siteConfig.applicationStack.javaVersion

`string`

Java runtime version.

Valid values: "8", "11", "17", "21"

- rule: java_version must be one of: 8, 11, 17, 21

### spec.siteConfig.applicationStack.powershellCoreVersion

`string`

PowerShell Core runtime version.

Valid values: "7", "7.2", "7.4"

- rule: powershell_core_version must be one of: 7, 7.2, 7.4

### spec.siteConfig.applicationStack.docker

`AzureFunctionAppDockerConfig`

Docker container configuration.
Runs a custom container image as the Function App runtime.
The image must include the Azure Functions runtime.

### spec.siteConfig.applicationStack.docker.registryUrl

`string` · required

The URL of the container registry.
Examples: "https://myregistry.azurecr.io", "https://ghcr.io"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.siteConfig.applicationStack.docker.imageName

`string` · required

The container image name (without tag).
Example: "myorg/my-function-app"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.siteConfig.applicationStack.docker.imageTag

`string` · required

The container image tag.
Example: "latest", "v1.2.3", "sha-abc123"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.siteConfig.applicationStack.docker.registryUsername

`string`

Username for authenticating with the container registry.
Not needed when using managed identity for ACR authentication
(set container_registry_use_managed_identity in site_config).

### spec.siteConfig.applicationStack.docker.registryPassword

`string | valueFrom` · sensitive

Password for authenticating with the container registry.
This is a sensitive credential. Provide directly or via StringValueOrRef.

Not needed when using managed identity for ACR authentication.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.siteConfig.applicationStack.useCustomRuntime

`bool` · optional (explicit presence)

Use a custom handler runtime.
Custom handlers allow any language/runtime to work with Azure Functions
by implementing a lightweight HTTP server that communicates with the
Functions host process.

### spec.siteConfig.alwaysOn

`bool` · optional (explicit presence)

Keep the Function App always loaded in memory.
When true, the app is never unloaded due to inactivity.

Critical for Dedicated plans (B*/S*/P*) -- without this, the app
may be unloaded after idle periods, causing cold start latency.

Automatically managed on Consumption and Elastic Premium plans.
Not supported on Free (F1) tier -- Azure rejects it at apply time.

### spec.siteConfig.appCommandLine

`string`

Custom startup command for the Function App.
Overrides the default startup behavior. Useful for custom Docker
containers or runtimes that need specific initialization.

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

### spec.siteConfig.defaultDocuments

`[]string`

Default documents served when a request maps to a directory.
Evaluated in order (e.g. ["index.html", "default.html"]).

### spec.siteConfig.healthCheckPath

`string`

Health check endpoint path.
Azure periodically sends requests to this path and marks the instance
as unhealthy if it doesn't respond with a 200-299 status code.
Unhealthy instances are removed from the load balancer rotation.

Recommended for production deployments. Common paths: "/api/health",
"/healthz".

### spec.siteConfig.healthCheckEvictionTimeInMin

`int32` · optional (explicit presence)

Time in minutes after which an unhealthy instance is evicted.
Azure monitors the health check path and removes instances that
have been continuously unhealthy for this duration.

Range: 2 to 10 minutes. Requires health_check_path.

- rule: {"int32":{"lte":10,"gte":2}}

### spec.siteConfig.minimumTlsVersion

`enum`

Minimum TLS version for incoming HTTPS requests. Unset deploys
TLS_1_2 (the industry floor; TLS_1_3 for maximum security).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_tls_version_unspecified` -- Not specified -- deploys TLS_1_2.
- `TLS_1_0` -- TLS 1.0 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_1` -- TLS 1.1 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_2` -- TLS 1.2 -- the industry floor (the default).
- `TLS_1_3` -- TLS 1.3 -- the strongest option; requires modern clients.

### spec.siteConfig.scmMinimumTlsVersion

`enum`

Minimum TLS version for the SCM (Kudu) site. Unset deploys TLS_1_2.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_tls_version_unspecified` -- Not specified -- deploys TLS_1_2.
- `TLS_1_0` -- TLS 1.0 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_1` -- TLS 1.1 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_2` -- TLS 1.2 -- the industry floor (the default).
- `TLS_1_3` -- TLS 1.3 -- the strongest option; requires modern clients.

### spec.siteConfig.minimumTlsCipherSuite

`string`

The minimum TLS cipher suite the app accepts. Suites WEAKER than the
chosen one are rejected. Azure's identifiers, strongest first:
TLS_AES_256_GCM_SHA384 ... TLS_RSA_WITH_AES_128_CBC_SHA. Leave empty
to accept Azure's platform default set.

- rule: minimum_tls_cipher_suite must be one of Azure's TLS cipher suite identifiers (e.g. TLS_AES_256_GCM_SHA384, TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)

### spec.siteConfig.appScaleLimit

`int32` · optional (explicit presence)

Maximum number of workers for scale-out.
Only applicable to Consumption and Elastic Premium plans.

Consumption (Y1): Azure auto-scales up to 200, but this field caps
the maximum to control costs.
Elastic Premium (EP*): Caps the elastic scale-out range.

Ignored on Dedicated plans (use worker_count on the Service Plan instead).

- rule: {"int32":{"gte":0}}

### spec.siteConfig.elasticInstanceMinimum

`int32` · optional (explicit presence)

Minimum number of pre-warmed instances for Elastic Premium plans.
These instances are always running and ready to handle requests,
eliminating cold start latency.

Only applicable to Elastic Premium (EP*) plans. Ignored on other tiers.

- rule: {"int32":{"gte":0}}

### spec.siteConfig.preWarmedInstanceCount

`int32` · optional (explicit presence)

Number of pre-warmed instances beyond the minimum.
Pre-warmed instances sit in a "warm" state and can handle requests
faster than cold-starting new instances.

Only applicable to Elastic Premium (EP*) plans.

- rule: {"int32":{"gte":0}}

### spec.siteConfig.workerCount

`int32` · optional (explicit presence)

Number of worker instances for the Function App.
Controls how many instances are allocated on Dedicated plans.

Range: 1-100.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.siteConfig.http2Enabled

`bool` · optional (explicit presence)

Enable HTTP/2 protocol for the Function App.
HTTP/2 provides multiplexing, header compression, and server push
for improved performance.

Default: false

- default: `false`

### spec.siteConfig.websocketsEnabled

`bool` · optional (explicit presence)

Enable WebSocket connections for the Function App.

Default: false

- default: `false`

### spec.siteConfig.use32BitWorker

`bool` · optional (explicit presence)

Use a 32-bit worker process instead of 64-bit.
Reduces memory footprint but limits addressable memory to ~2 GB.

Default: false

- default: `false`

### spec.siteConfig.vnetRouteAllEnabled

`bool` · optional (explicit presence)

Route all outbound traffic from the Function App through the VNet.
Requires virtual_network_subnet_id to be set on the spec (spec-enforced).

When false (default), only RFC1918 traffic routes through the VNet.
When true, all outbound traffic (including public internet) routes
through the VNet, enabling inspection via NSG rules or a firewall.

Not supported on Consumption plans (Y1).

Default: false

- default: `false`

### spec.siteConfig.ftpsState

`enum`

FTPS state for the Function App -- the FTP deployment endpoint's TLS
posture. Unset deploys DISABLED (secure by default). Independent of
ftp_publish_basic_authentication_enabled, which controls whether the
publishing credential works at all.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_ftps_state_unspecified` -- Not specified -- deploys DISABLED (secure by default).
- `ALL_ALLOWED` -- Both plain FTP and FTPS accepted.
- `FTPS_ONLY` -- Only FTPS (encrypted) accepted.
- `DISABLED` -- The FTP endpoint is off entirely (recommended).

### spec.siteConfig.loadBalancingMode

`enum`

Load balancing mode for distributing requests across instances.
Unset deploys LEAST_REQUESTS.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_load_balancing_mode_unspecified` -- Not specified -- deploys LEAST_REQUESTS.
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

- `azure_function_app_managed_pipeline_mode_unspecified` -- Not specified -- deploys INTEGRATED.
- `INTEGRATED` -- The modern pipeline (the default; correct for everything current).
- `CLASSIC` -- Legacy-compatibility pipeline.

### spec.siteConfig.remoteDebuggingEnabled

`bool` · optional (explicit presence)

Enable remote debugging (Visual Studio attach). Azure supports the
current Visual Studio generation only; the platform lets Azure pick
the debugger version. Turn off outside active debugging sessions.

Default: false

- default: `false`

### spec.siteConfig.runtimeScaleMonitoringEnabled

`bool` · optional (explicit presence)

Enable runtime scale monitoring for KEDA-based triggers.
When enabled, the Functions runtime can directly monitor event sources
to make more accurate scaling decisions.

Supported on Elastic Premium and Dedicated plans with Functions v4+.

### spec.siteConfig.cors

`AzureFunctionAppCorsSettings`

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

Allow credentials (cookies, authorization headers) in cross-origin requests.
Cannot be combined with a wildcard origin (enforced here -- browsers
reject the pairing).

Default: false

- default: `false`

### spec.siteConfig.ipRestrictions

`[]AzureFunctionAppIpRestriction`

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

- `azure_function_app_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.ipRestrictions[].ipAddress

`string`

IP address or CIDR range.
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

`AzureFunctionAppIpRestrictionHeaders`

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

- `azure_function_app_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.scmUseMainIpRestriction

`bool` · optional (explicit presence)

Use the main site's IP restrictions for the SCM (Kudu) site.
When true, scm_ip_restrictions are ignored.

Default: false

- default: `false`

### spec.siteConfig.scmIpRestrictions

`[]AzureFunctionAppIpRestriction`

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

- `azure_function_app_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.scmIpRestrictions[].ipAddress

`string`

IP address or CIDR range.
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

`AzureFunctionAppIpRestrictionHeaders`

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

- `azure_function_app_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.appServiceLogs

`AzureFunctionAppAppServiceLogs`

App Service logging configuration (disk quota + retention for the
file-system logs).

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

Use managed identity for pulling container images from Azure Container Registry.
Requires the Function App's identity to have AcrPull role on the registry.

Default: false

- default: `false`

### spec.siteConfig.containerRegistryManagedIdentityClientId

`string`

Client ID of the managed identity used for ACR image pulls.
Only used when container_registry_use_managed_identity is true and
a user-assigned identity (not system-assigned) should be used.

### spec.siteConfig.applicationInsightsKey

`string` · sensitive

Application Insights instrumentation key (classic).
Prefer application_insights_connection_string on the parent spec
for new deployments. This field is for backward compatibility with
apps already using the instrumentation key.

### spec.appSettings

`map<string, string>`

Application settings (environment variables) for the Function App.
Key-value pairs that are available to functions at runtime via
environment variables.

Azure automatically manages several settings (AzureWebJobsStorage,
FUNCTIONS_WORKER_RUNTIME, etc.). User-provided settings are merged
with these system settings. Auth provider secrets referenced by
auth_settings_v2 setting names also live here.

### spec.connectionStrings

`[]AzureFunctionAppConnectionString`

Named connection strings for database and service connections.
Each connection string has a name, type, and value. The type determines
how Azure exposes the connection in the runtime environment.

For most use cases, app_settings is simpler. Use connection_strings when
you need Azure's native connection string management (e.g., for Entity
Framework, Azure Service Bus SDK auto-discovery).

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

- `azure_function_app_connection_string_type_unspecified` -- Not specified -- invalid; pick the service type (CUSTOM for anything without a dedicated type).
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

The connection string value. This is a sensitive credential (the
provider marks it Sensitive): supply a managed-secret reference, a
sibling resource's output reference, or a literal.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.stickySettings

`AzureFunctionAppStickySettings`

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
When provided, Azure automatically configures the Function App to send
telemetry (requests, dependencies, exceptions, traces) to Application
Insights. This is the recommended way to monitor Function Apps.

Uses the connection_string format (not the legacy instrumentation_key).

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
When false, the Function App is only accessible via VNet integration
or Private Endpoints.

Default: true

- default: `true`

### spec.enabled

`bool` · optional (explicit presence)

Enable or disable the Function App.
When false, the app is stopped and does not run functions, but the
resource still exists and incurs plan-level costs. Useful for
temporarily disabling an app without deleting it.

Default: true

- default: `true`

### spec.builtinLoggingEnabled

`bool` · optional (explicit presence)

Enable built-in logging via AzureWebJobsDashboard.
When true, Azure configures the AzureWebJobsDashboard storage connection
for the legacy Functions dashboard. When Application Insights is configured,
you may want to disable this to avoid duplicate logging and storage costs.

Default: true

- default: `true`

### spec.contentShareForceDisabled

`bool` · optional (explicit presence)

Force disable the Azure Files content share that Azure Functions
automatically creates. Set to true when using a custom deployment
method and the content share is not needed.

Default: false

- default: `false`

### spec.clientCertificateEnabled

`bool` · optional (explicit presence)

Enable client certificate authentication (mutual TLS).
When true, clients must present a valid certificate to access the app
(subject to client_certificate_mode).

Default: false

- default: `false`

### spec.clientCertificateMode

`enum`

Client certificate mode -- how strictly client certificates are
enforced when client_certificate_enabled is true. Unset deploys
OPTIONAL (certificate requested but not required).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_client_certificate_mode_unspecified` -- Not specified -- deploys OPTIONAL.
- `REQUIRED` -- All requests must present a valid client certificate.
- `OPTIONAL` -- Certificate is requested but not required.
- `OPTIONAL_INTERACTIVE_USER` -- Certificate is optional for interactive (browser) users, required for non-interactive clients.

### spec.clientCertificateExclusionPaths

`string`

Paths excluded from client certificate validation.
Semicolon-separated list of paths where client certificates are not required.
Example: "/api/health;/api/status"

### spec.virtualNetworkSubnetId

`string | valueFrom`

The subnet ID for VNet integration. When provided, the Function App's
outbound traffic routes through this subnet, enabling access to
VNet-connected resources (databases, Redis, etc.) without public endpoints.

The subnet must be delegated to Microsoft.Web/serverFarms.
Not supported on Consumption plans (Y1).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.vnetImagePullEnabled

`bool` · optional (explicit presence)

Pull container images over the VNet integration (instead of over the
public internet). Required for images in registries that are only
reachable privately (e.g. an ACR locked to a private endpoint).
Requires virtual_network_subnet_id. Azure additionally requires this
to stay enabled for apps hosted in an App Service Environment, and
rejects it on Consumption plans (apply-time contracts).

Default: false

- default: `false`

### spec.virtualNetworkBackupRestoreEnabled

`bool` · optional (explicit presence)

Route the backup/restore traffic of the `backup` block over the VNet
integration -- needed when the backup storage account is firewalled
to the VNet. Requires virtual_network_subnet_id.

Default: false

- default: `false`

### spec.identity

`AzureFunctionAppIdentity`

Managed identity configuration for the Function App.
Enables the app to authenticate with Azure services (Key Vault, Storage,
ACR, etc.) without managing credentials.

When identity is configured with SYSTEM_ASSIGNED, the function app gets
a system-assigned identity whose principal_id and tenant_id are exported
as stack outputs.

- rule: identity_ids is required when type includes USER_ASSIGNED, and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

The identity model: Azure-managed (tied to the app's lifecycle),
bring-your-own (independent lifecycle), or both.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_identity_type_unspecified` -- Not specified -- invalid; pick an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates and rotates an identity tied to the app's lifecycle.
- `USER_ASSIGNED` -- Attach pre-created AzureUserAssignedIdentity resources (independent lifecycle; shareable across apps).
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned identity and user-assigned identities.

### spec.identity.identityIds

`[]string | valueFrom`

User Assigned Identity Azure resource IDs.
Required when type includes USER_ASSIGNED.

Can be literal ARM resource IDs or references to AzureUserAssignedIdentity outputs.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.keyVaultReferenceIdentityId

`string | valueFrom`

User Assigned Identity ID for accessing Key Vault references.
When the Function App uses Key Vault references in app_settings
(e.g., `@Microsoft.KeyVault(SecretUri=...)`), this identity is used
to authenticate with Key Vault.

If not specified, the system-assigned identity is used.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.webdeployPublishBasicAuthenticationEnabled

`bool` · optional (explicit presence)

Allow basic-auth (username/password) publishing over Web Deploy
(msdeploy). Disabling both basic-auth toggles closes the classic
credential-based deployment paths and forces identity-based
deployment (recommended posture for locked-down environments).

Default: true (Azure's own default; flip to false to harden)

- default: `true`

### spec.ftpPublishBasicAuthenticationEnabled

`bool` · optional (explicit presence)

Allow basic-auth (username/password) publishing over FTP/FTPS.
Independent of site_config.ftps_state (which controls the FTP
endpoint's TLS posture); this toggle controls whether the publishing
CREDENTIAL is accepted at all.

Default: true (Azure's own default; flip to false to harden)

- default: `true`

### spec.zipDeployFile

`string`

Path to a local ZIP package to deploy on create/update (one-shot
"run-from-package" style zip deploy). Primarily useful for simple
pipelines that produce a build artifact next to the manifest; most
production deployments push code through CI/CD instead.

### spec.storageMounts

`[]AzureFunctionAppStorageMount`

Azure Storage Account mounts for the Function App.
Mounts Azure File Shares or Blob containers as directories accessible
to the function code at runtime.

### spec.storageMounts[].name

`string` · required

Unique name for this mount within the Function App.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.storageMounts[].type

`enum` · required

What is being mounted: an Azure File Share (read-write) or a Blob
container (read-only).

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_storage_mount_type_unspecified` -- Not specified -- invalid; pick the storage service.
- `AZURE_FILES` -- Azure File Share (SMB) -- read-write.
- `AZURE_BLOB` -- Azure Blob container -- read-only.

### spec.storageMounts[].accountName

`string` · required

Name of the Azure Storage Account that contains the share or container.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.storageMounts[].shareName

`string` · required

Name of the file share or blob container to mount.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.storageMounts[].accessKey

`string | valueFrom` · required · sensitive

Access key for the storage account. Defaults to referencing an
AzureStorageAccount's primary_access_key output, so the mount
composes in one manifest set; a literal value or a managed-secret
reference works too.

- references: AzureStorageAccount (`status.outputs.primary_access_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_access_key}} -- a bare string does not parse

### spec.storageMounts[].mountPath

`string`

Path inside the container where the share is mounted.
Example: "/mnt/data"

### spec.backup

`AzureFunctionAppBackup`

Scheduled backups of the app's content and configuration to an Azure
Storage container (referenced by SAS URL). Requires Standard tier or
above -- Azure rejects backup on Consumption and Basic plans at
apply time. Restore is an operational action in the portal/CLI, not
a manifest field.

### spec.backup.name

`string` · required

A name for this backup job (appears in the portal's backup list).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.backup.storageAccountUrl

`string | valueFrom` · required · sensitive

SAS URL of the storage container receiving the backups. Carries a
signed write credential -- treat like a password.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.backup.enabled

`bool` · optional (explicit presence)

Whether the backup job runs. Disable to pause the schedule without
losing its configuration.

Default: true

- default: `true`

### spec.backup.schedule

`AzureFunctionAppBackupSchedule` · required

When and how often backups run, and how long they are kept.

- rule: {"required":true}

### spec.backup.schedule.frequencyInterval

`int32` · required

How many frequency_units pass between backups (e.g. 1 DAY = daily,
12 HOUR = twice a day). Range: 1-1000.

- rule: {"required":true,"int32":{"lte":1000,"gte":1}}

### spec.backup.schedule.frequencyUnit

`enum` · required

The unit frequency_interval counts in.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_function_app_backup_frequency_unit_unspecified` -- Not specified -- invalid; pick DAY or HOUR.
- `DAY`
- `HOUR`

### spec.backup.schedule.keepAtLeastOneBackup

`bool` · optional (explicit presence)

Never delete the last remaining backup, even when retention would
expire it.

Default: false

- default: `false`

### spec.backup.schedule.retentionPeriodDays

`int32` · optional (explicit presence)

Days each backup is retained before deletion. 0 keeps backups
indefinitely.

Default: 30

- default: `30`
- rule: {"int32":{"gte":0}}

### spec.backup.schedule.startTime

`string`

When the schedule starts, RFC3339 format (e.g.
"2026-01-01T00:00:00Z"). Unset starts immediately.

### spec.authSettingsV2

`AzureFunctionAppAuthSettingsV2`

App Service built-in authentication (Easy Auth v2). When enabled,
Azure authenticates requests at the platform layer -- before they
reach function code -- against any of the configured identity
providers. Provider client secrets are referenced by APP SETTING NAME
(set the actual secret value in app_settings or via a Key Vault
reference), never inline in this block.

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

- `azure_function_app_unauthenticated_action_unspecified` -- Not specified -- deploys REDIRECT_TO_LOGIN_PAGE.
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

- `azure_function_app_forward_proxy_convention_unspecified` -- Not specified -- deploys FORWARD_PROXY_NO_PROXY.
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

`AzureFunctionAppAuthV2Login` · required

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

- `azure_function_app_cookie_expiration_convention_unspecified` -- Not specified -- deploys FIXED_TIME.
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

`AzureFunctionAppAuthV2Apple`

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

`AzureFunctionAppAuthV2ActiveDirectory`

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

`AzureFunctionAppAuthV2StaticWebApp`

Azure Static Web Apps authentication (when fronted by one).

### spec.authSettingsV2.azureStaticWebAppV2.clientId

`string` · required

The Static Web App's client ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.customOidcV2

`[]AzureFunctionAppAuthV2CustomOidc`

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

`AzureFunctionAppAuthV2Facebook`

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

`AzureFunctionAppAuthV2Github`

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

`AzureFunctionAppAuthV2Google`

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

`AzureFunctionAppAuthV2Microsoft`

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

`AzureFunctionAppAuthV2Twitter`

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

- `function_app_storage_binding_exactly_one`: exactly one storage binding applies: storage_account_name OR storage_key_vault_secret_id
- `function_app_storage_key_xor_identity`: storage_account_access_key conflicts with storage_uses_managed_identity -- pick one storage authentication method
- `function_app_vnet_image_pull_requires_subnet`: vnet_image_pull_enabled requires virtual_network_subnet_id (the image pull rides the VNet integration)
- `function_app_vnet_backup_restore_requires_subnet`: virtual_network_backup_restore_enabled requires virtual_network_subnet_id (backup traffic rides the VNet integration)
- `function_app_vnet_route_all_requires_subnet`: site_config.vnet_route_all_enabled requires virtual_network_subnet_id (there is no VNet to route through otherwise)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureFunctionApp, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.function_app_id` | `string` | The Azure Resource Manager ID of the Function App. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Web/sites/{name} |
| `status.outputs.default_hostname` | `string` | The default hostname of the Function App. Format: {name}.azurewebsites.net This is the primary endpoint for HTTP-triggered functions. Custom domains can be added via Azure portal or DNS configuration. |
| `status.outputs.outbound_ip_addresses` | `[]string` | Outbound IP addresses used by the Function App. These IPs should be allowed in downstream firewall rules (e.g., database firewall, third-party API whitelists). Note: On Consumption plans, outbound IPs are shared across the region and may change. On Dedicated/Premium plans, IPs are stable. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the system-assigned managed identity. Populated only when the Function App has a system-assigned identity (identity.type includes "SystemAssigned"). Used for granting RBAC roles: e.g., "Key Vault Secrets User", "Storage Blob Data Contributor". |
| `status.outputs.identity_tenant_id` | `string` | The tenant ID of the system-assigned managed identity. Paired with identity_principal_id for RBAC configuration. |
| `status.outputs.custom_domain_verification_id` | `string` | The custom domain verification ID. Used when binding custom domains to the Function App. Add this value as a TXT record at `asuid.{custom-domain}` to verify domain ownership. |
| `status.outputs.kind` | `string` | The resource kind string as reported by Azure. Example: "functionapp,linux" |
| `status.outputs.possible_outbound_ip_addresses` | `[]string` | Every outbound IP address the platform could EVER route this app's traffic through (a superset of outbound_ip_addresses, which lists only the currently active set). Use THIS list for downstream firewall allowlists that must survive scale events and platform moves. |
| `status.outputs.hosting_environment_id` | `string` | The ARM ID of the App Service Environment hosting the app -- set only when the app's plan runs on Isolated SKUs inside an ASE. |
| `status.outputs.site_credential_name` | `string` | The site-level publishing credential's username (the Kudu/SCM basic-auth user). Paired with site_credential_password; only usable while the basic-auth publishing toggles are enabled. |
| `status.outputs.site_credential_password` | `string` (sensitive) | The site-level publishing credential's password. SECRET-BEARING: anyone holding it can deploy code to the app over Web Deploy/SCM while basic-auth publishing is enabled -- treat it like an admin password (disable the basic-auth toggles to revoke the surface entirely). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.servicePlanId` | AzureServicePlan | `status.outputs.service_plan_id` |
| `spec.storageAccountName` | AzureStorageAccount | `status.outputs.storage_account_name` |
| `spec.storageAccountAccessKey` | AzureStorageAccount | `status.outputs.primary_access_key` |
| `spec.siteConfig.ipRestrictions[].virtualNetworkSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.siteConfig.ipRestrictions[].headers.xAzureFdid` | AzureFrontDoorProfile | `status.outputs.resource_guid` |
| `spec.siteConfig.scmIpRestrictions[].virtualNetworkSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.siteConfig.scmIpRestrictions[].headers.xAzureFdid` | AzureFrontDoorProfile | `status.outputs.resource_guid` |
| `spec.applicationInsightsConnectionString` | AzureApplicationInsights | `status.outputs.connection_string` |
| `spec.virtualNetworkSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.keyVaultReferenceIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.storageMounts[].accessKey` | AzureStorageAccount | `status.outputs.primary_access_key` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureMonitorActionGroup | `spec.azureFunctionReceivers[].functionAppResourceId` | `status.outputs.function_app_id` |

## See Also

- [Overview](../README.md)

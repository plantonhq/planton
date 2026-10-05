# AzureLinuxWebApp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureLinuxWebAppSpec** defines the configuration for creating an Azure
Linux Web App (Microsoft.Web/sites kind=app,linux), a managed web hosting
platform for running web applications on Azure App Service.

A Linux Web App hosts long-running web applications serving HTTP traffic.
Unlike serverless Function Apps that are event-driven and scale to zero,
Web Apps are designed for always-on workloads: APIs, web frontends,
background services, and containerized applications.

**Relationship to AzureServicePlan**:

Every Web App requires a Service Plan. The plan determines cost model,
compute tier, and available features:

Free/Shared (F1/D1):
  - Shared infrastructure, no Always-On, dev/test only

Basic (B1-B3):
  - Dedicated compute, custom domains, manual scaling
  - No auto-scale, no deployment slots

Standard (S1-S3):
  - Auto-scale up to 10 instances, deployment slots, daily backups
  - VNet integration, hybrid connections

Premium (P*v2/P*v3/P*v4):
  - Up to 30 instances, enhanced performance
  - VNet integration, private endpoints, zone redundancy

**Application stack**:

The application_stack within site_config selects the runtime. Exactly one
runtime must be chosen: .NET, Node.js, Python, PHP, Go, Java
(with server selection), or Docker container. Web Apps support a broader
set of runtimes compared to Function Apps, including PHP, Go, and
Java with configurable application servers (Tomcat, JBoss EAP).

**Authentication (Easy Auth)**: the `auth_settings_v2` block turns on
App Service's built-in authentication layer -- Azure validates identity
tokens at the front door (Entra ID, Apple, Facebook, GitHub, Google,
Microsoft account, Twitter, or any OpenID Connect provider) before
requests reach application code. Provider secrets are referenced by APP
SETTING NAME (never inline), so the secret values live in app_settings
or Key Vault references.

**Windows Web Apps** (`azurerm_windows_web_app`) are deliberately not
modeled: the platform targets Linux-first runtimes and containers, and
Windows remains the legacy .NET Framework path. The `AzureServicePlan`
kind supports Windows plans, so the compute tier is ready if a Windows
app kind is ever added.

**ForceNew fields** (changing these destroys and recreates the web app):
`web_app_name`, `region`, `resource_group`.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLinuxWebApp
metadata:
  name: hack-webapp
spec:
  region: eastus
  resourceGroup:
    value: hack-rg
  webAppName: planton-hack-webapp-001
  servicePlanId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/hack-rg/providers/Microsoft.Web/serverFarms/hack-plan
  httpsOnly: true
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
    - name: Database
      type: POSTGRESQL
      value:
        value: Host=db.example.com;Database=app
  siteConfig:
    applicationStack:
      pythonVersion: "3.12"
    healthCheckPath: /health
    healthCheckEvictionTimeInMin: 5
    # Enum-valued dials exercise the wire maps on both engines.
    minimumTlsVersion: TLS_1_2
    ftpsState: DISABLED
    loadBalancingMode: LEAST_REQUESTS
    managedPipelineMode: INTEGRATED
    http2Enabled: true
    defaultDocuments:
      - index.html
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
    # Auto-heal recycles on a 5xx storm; the uptime guard prevents
    # recycle loops.
    autoHealSetting:
      trigger:
        statusCodes:
          - statusCodeRange: 500-599
            count: 50
            interval: "00:05:00"
      minimumProcessExecutionTime: "00:01:00"
  logs:
    applicationLogs:
      fileSystemLevel: INFORMATION
    httpLogs:
      fileSystem:
        retentionInMb: 50
        retentionInDays: 7
    failedRequestTracing: true
  # Easy Auth v2 with a custom OIDC provider exercises the auth block
  # end to end without needing a real Entra app registration at plan time.
  authSettingsV2:
    authEnabled: true
    requireAuthentication: true
    unauthenticatedAction: RETURN_401
    excludedPaths:
      - /health
    login:
      tokenStoreEnabled: true
    customOidcV2:
      - name: corp-idp
        clientId: web-app-client
        openidConfigurationEndpoint: https://idp.example.com/.well-known/openid-configuration
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.webAppName` | `string` | yes |  |  |
| `spec.servicePlanId` | `string \| valueFrom` | yes |  | AzureServicePlan (`status.outputs.service_plan_id`) |
| `spec.siteConfig` | `AzureLinuxWebAppSiteConfig` | yes |  |  |
| `spec.siteConfig.applicationStack` | `AzureLinuxWebAppApplicationStack` |  |  |  |
| `spec.siteConfig.applicationStack.dotnetVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.nodeVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.pythonVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.phpVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.goVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.javaVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.javaServer` | `enum` |  |  |  |
| `spec.siteConfig.applicationStack.javaServerVersion` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.docker` | `AzureLinuxWebAppDockerConfig` |  |  |  |
| `spec.siteConfig.applicationStack.docker.registryUrl` | `string` | yes |  |  |
| `spec.siteConfig.applicationStack.docker.imageName` | `string` | yes |  |  |
| `spec.siteConfig.applicationStack.docker.imageTag` | `string` | yes |  |  |
| `spec.siteConfig.applicationStack.docker.registryUsername` | `string` |  |  |  |
| `spec.siteConfig.applicationStack.docker.registryPassword` | `string \| valueFrom` (sensitive) |  |  |  |
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
| `spec.siteConfig.ftpsState` | `enum` |  |  |  |
| `spec.siteConfig.workerCount` | `int32` |  |  |  |
| `spec.siteConfig.http2Enabled` | `bool` |  | `false` |  |
| `spec.siteConfig.websocketsEnabled` | `bool` |  | `false` |  |
| `spec.siteConfig.use32BitWorker` | `bool` |  | `false` |  |
| `spec.siteConfig.vnetRouteAllEnabled` | `bool` |  | `false` |  |
| `spec.siteConfig.loadBalancingMode` | `enum` |  |  |  |
| `spec.siteConfig.managedPipelineMode` | `enum` |  |  |  |
| `spec.siteConfig.localMysqlEnabled` | `bool` |  | `false` |  |
| `spec.siteConfig.remoteDebuggingEnabled` | `bool` |  | `false` |  |
| `spec.siteConfig.cors` | `AzureLinuxWebAppCorsSettings` |  |  |  |
| `spec.siteConfig.cors.allowedOrigins` | `[]string` | yes |  |  |
| `spec.siteConfig.cors.supportCredentials` | `bool` |  | `false` |  |
| `spec.siteConfig.ipRestrictions` | `[]AzureLinuxWebAppIpRestriction` |  |  |  |
| `spec.siteConfig.ipRestrictions[].name` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].priority` | `int32` |  |  |  |
| `spec.siteConfig.ipRestrictions[].action` | `enum` |  |  |  |
| `spec.siteConfig.ipRestrictions[].ipAddress` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].serviceTag` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].virtualNetworkSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.siteConfig.ipRestrictions[].description` | `string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers` | `AzureLinuxWebAppIpRestrictionHeaders` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers.xForwardedFor` | `[]string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers.xForwardedHost` | `[]string` |  |  |  |
| `spec.siteConfig.ipRestrictions[].headers.xAzureFdid` | `[]string \| valueFrom` |  |  | AzureFrontDoorProfile (`status.outputs.resource_guid`) |
| `spec.siteConfig.ipRestrictions[].headers.xFdHealthProbe` | `[]string` |  |  |  |
| `spec.siteConfig.ipRestrictionDefaultAction` | `enum` |  |  |  |
| `spec.siteConfig.scmUseMainIpRestriction` | `bool` |  | `false` |  |
| `spec.siteConfig.scmIpRestrictions` | `[]AzureLinuxWebAppIpRestriction` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].name` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].priority` | `int32` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].action` | `enum` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].ipAddress` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].serviceTag` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].virtualNetworkSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.siteConfig.scmIpRestrictions[].description` | `string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers` | `AzureLinuxWebAppIpRestrictionHeaders` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers.xForwardedFor` | `[]string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers.xForwardedHost` | `[]string` |  |  |  |
| `spec.siteConfig.scmIpRestrictions[].headers.xAzureFdid` | `[]string \| valueFrom` |  |  | AzureFrontDoorProfile (`status.outputs.resource_guid`) |
| `spec.siteConfig.scmIpRestrictions[].headers.xFdHealthProbe` | `[]string` |  |  |  |
| `spec.siteConfig.scmIpRestrictionDefaultAction` | `enum` |  |  |  |
| `spec.siteConfig.containerRegistryUseManagedIdentity` | `bool` |  | `false` |  |
| `spec.siteConfig.containerRegistryManagedIdentityClientId` | `string` |  |  |  |
| `spec.siteConfig.autoHealSetting` | `AzureLinuxWebAppAutoHealSetting` |  |  |  |
| `spec.siteConfig.autoHealSetting.trigger` | `AzureLinuxWebAppAutoHealTrigger` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.requests` | `AzureLinuxWebAppAutoHealRequestsTrigger` |  |  |  |
| `spec.siteConfig.autoHealSetting.trigger.requests.count` | `int32` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.requests.interval` | `string` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.statusCodes` | `[]AzureLinuxWebAppAutoHealStatusCodeTrigger` |  |  |  |
| `spec.siteConfig.autoHealSetting.trigger.statusCodes[].statusCodeRange` | `string` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.statusCodes[].count` | `int32` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.statusCodes[].interval` | `string` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.statusCodes[].subStatus` | `int32` |  |  |  |
| `spec.siteConfig.autoHealSetting.trigger.statusCodes[].win32StatusCode` | `int32` |  |  |  |
| `spec.siteConfig.autoHealSetting.trigger.statusCodes[].path` | `string` |  |  |  |
| `spec.siteConfig.autoHealSetting.trigger.slowRequest` | `AzureLinuxWebAppAutoHealSlowRequestTrigger` |  |  |  |
| `spec.siteConfig.autoHealSetting.trigger.slowRequest.timeTaken` | `string` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.slowRequest.interval` | `string` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.slowRequest.count` | `int32` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath` | `[]AzureLinuxWebAppAutoHealSlowRequestWithPathTrigger` |  |  |  |
| `spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath[].timeTaken` | `string` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath[].interval` | `string` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath[].count` | `int32` | yes |  |  |
| `spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath[].path` | `string` |  |  |  |
| `spec.siteConfig.autoHealSetting.minimumProcessExecutionTime` | `string` |  |  |  |
| `spec.appSettings` | `map<string, string>` |  |  |  |
| `spec.connectionStrings` | `[]AzureLinuxWebAppConnectionString` |  |  |  |
| `spec.connectionStrings[].name` | `string` | yes |  |  |
| `spec.connectionStrings[].type` | `enum` | yes |  |  |
| `spec.connectionStrings[].value` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.stickySettings` | `AzureLinuxWebAppStickySettings` |  |  |  |
| `spec.stickySettings.appSettingNames` | `[]string` |  |  |  |
| `spec.stickySettings.connectionStringNames` | `[]string` |  |  |  |
| `spec.applicationInsightsConnectionString` | `string \| valueFrom` |  |  | AzureApplicationInsights (`status.outputs.connection_string`) |
| `spec.httpsOnly` | `bool` |  | `true` |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.clientAffinityEnabled` | `bool` |  | `false` |  |
| `spec.clientCertificateEnabled` | `bool` |  | `false` |  |
| `spec.clientCertificateMode` | `enum` |  |  |  |
| `spec.clientCertificateExclusionPaths` | `string` |  |  |  |
| `spec.virtualNetworkSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.vnetImagePullEnabled` | `bool` |  | `false` |  |
| `spec.virtualNetworkBackupRestoreEnabled` | `bool` |  | `false` |  |
| `spec.identity` | `AzureLinuxWebAppIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.keyVaultReferenceIdentityId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.webdeployPublishBasicAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.ftpPublishBasicAuthenticationEnabled` | `bool` |  | `true` |  |
| `spec.zipDeployFile` | `string` |  |  |  |
| `spec.storageMounts` | `[]AzureLinuxWebAppStorageMount` |  |  |  |
| `spec.storageMounts[].name` | `string` | yes |  |  |
| `spec.storageMounts[].type` | `enum` | yes |  |  |
| `spec.storageMounts[].accountName` | `string` | yes |  |  |
| `spec.storageMounts[].shareName` | `string` | yes |  |  |
| `spec.storageMounts[].accessKey` | `string \| valueFrom` (sensitive) | yes |  | AzureStorageAccount (`status.outputs.primary_access_key`) |
| `spec.storageMounts[].mountPath` | `string` |  |  |  |
| `spec.logs` | `AzureLinuxWebAppLogs` |  |  |  |
| `spec.logs.applicationLogs` | `AzureLinuxWebAppApplicationLogs` |  |  |  |
| `spec.logs.applicationLogs.fileSystemLevel` | `enum` |  |  |  |
| `spec.logs.applicationLogs.azureBlobStorage` | `AzureLinuxWebAppBlobStorageLogs` |  |  |  |
| `spec.logs.applicationLogs.azureBlobStorage.level` | `enum` | yes |  |  |
| `spec.logs.applicationLogs.azureBlobStorage.sasUrl` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.logs.applicationLogs.azureBlobStorage.retentionInDays` | `int32` |  | `0` |  |
| `spec.logs.httpLogs` | `AzureLinuxWebAppHttpLogs` |  |  |  |
| `spec.logs.httpLogs.fileSystem` | `AzureLinuxWebAppHttpLogsFileSystem` |  |  |  |
| `spec.logs.httpLogs.fileSystem.retentionInMb` | `int32` |  | `35` |  |
| `spec.logs.httpLogs.fileSystem.retentionInDays` | `int32` |  | `0` |  |
| `spec.logs.httpLogs.azureBlobStorage` | `AzureLinuxWebAppHttpLogsBlobStorage` |  |  |  |
| `spec.logs.httpLogs.azureBlobStorage.sasUrl` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.logs.httpLogs.azureBlobStorage.retentionInDays` | `int32` |  | `0` |  |
| `spec.logs.failedRequestTracing` | `bool` |  | `false` |  |
| `spec.logs.detailedErrorMessages` | `bool` |  | `false` |  |
| `spec.backup` | `AzureLinuxWebAppBackup` |  |  |  |
| `spec.backup.name` | `string` | yes |  |  |
| `spec.backup.storageAccountUrl` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.backup.enabled` | `bool` |  | `true` |  |
| `spec.backup.schedule` | `AzureLinuxWebAppBackupSchedule` | yes |  |  |
| `spec.backup.schedule.frequencyInterval` | `int32` | yes |  |  |
| `spec.backup.schedule.frequencyUnit` | `enum` | yes |  |  |
| `spec.backup.schedule.keepAtLeastOneBackup` | `bool` |  | `false` |  |
| `spec.backup.schedule.retentionPeriodDays` | `int32` |  | `30` |  |
| `spec.backup.schedule.startTime` | `string` |  |  |  |
| `spec.authSettingsV2` | `AzureLinuxWebAppAuthSettingsV2` |  |  |  |
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
| `spec.authSettingsV2.login` | `AzureLinuxWebAppAuthV2Login` | yes |  |  |
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
| `spec.authSettingsV2.appleV2` | `AzureLinuxWebAppAuthV2Apple` |  |  |  |
| `spec.authSettingsV2.appleV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.appleV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.activeDirectoryV2` | `AzureLinuxWebAppAuthV2ActiveDirectory` |  |  |  |
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
| `spec.authSettingsV2.azureStaticWebAppV2` | `AzureLinuxWebAppAuthV2StaticWebApp` |  |  |  |
| `spec.authSettingsV2.azureStaticWebAppV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2` | `[]AzureLinuxWebAppAuthV2CustomOidc` |  |  |  |
| `spec.authSettingsV2.customOidcV2[].name` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2[].clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2[].openidConfigurationEndpoint` | `string` | yes |  |  |
| `spec.authSettingsV2.customOidcV2[].nameClaimType` | `string` |  |  |  |
| `spec.authSettingsV2.customOidcV2[].scopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.facebookV2` | `AzureLinuxWebAppAuthV2Facebook` |  |  |  |
| `spec.authSettingsV2.facebookV2.appId` | `string` | yes |  |  |
| `spec.authSettingsV2.facebookV2.appSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.facebookV2.graphApiVersion` | `string` |  |  |  |
| `spec.authSettingsV2.facebookV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.githubV2` | `AzureLinuxWebAppAuthV2Github` |  |  |  |
| `spec.authSettingsV2.githubV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.githubV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.githubV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.googleV2` | `AzureLinuxWebAppAuthV2Google` |  |  |  |
| `spec.authSettingsV2.googleV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.googleV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.googleV2.allowedAudiences` | `[]string` |  |  |  |
| `spec.authSettingsV2.googleV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.microsoftV2` | `AzureLinuxWebAppAuthV2Microsoft` |  |  |  |
| `spec.authSettingsV2.microsoftV2.clientId` | `string` | yes |  |  |
| `spec.authSettingsV2.microsoftV2.clientSecretSettingName` | `string` | yes |  |  |
| `spec.authSettingsV2.microsoftV2.allowedAudiences` | `[]string` |  |  |  |
| `spec.authSettingsV2.microsoftV2.loginScopes` | `[]string` |  |  |  |
| `spec.authSettingsV2.twitterV2` | `AzureLinuxWebAppAuthV2Twitter` |  |  |  |
| `spec.authSettingsV2.twitterV2.consumerKey` | `string` | yes |  |  |
| `spec.authSettingsV2.twitterV2.consumerSecretSettingName` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the Web App will be created.
Examples: "eastus", "westus2", "westeurope", "southeastasia".

**ForceNew**: Changing this destroys and recreates the web app.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the Web App will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

**ForceNew**: Changing this destroys and recreates the web app.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.webAppName

`string` · required

The name of the Web App.
Must be globally unique across Azure (it forms the default hostname:
`{web_app_name}.azurewebsites.net`).

Allowed characters: alphanumeric and hyphens.
Must start and end with an alphanumeric character.
Length: 2 to 60 characters.

**ForceNew**: Changing this destroys and recreates the web app.

- rule: web_app_name must contain only alphanumeric characters and hyphens, and must start and end with an alphanumeric character
- rule: {"required":true,"string":{"minLen":"2","maxLen":"60"}}

### spec.servicePlanId

`string | valueFrom` · required

The App Service Plan that provides compute resources for this Web App.
Determines the pricing tier, scale behavior, and available features.
NOT ForceNew -- apps can move between plans in the same region and
resource group.

- references: AzureServicePlan (`status.outputs.service_plan_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServicePlan, name: <that resource's name>, fieldPath: status.outputs.service_plan_id}} -- a bare string does not parse

### spec.siteConfig

`AzureLinuxWebAppSiteConfig` · required

Site configuration for the Web App.
Contains the application stack (runtime), security settings,
networking behavior, auto-heal rules, and operational configuration.

- rule: {"required":true}
- rule: health_check_eviction_time_in_min requires health_check_path (there is no probe to evict on otherwise)

### spec.siteConfig.applicationStack

`AzureLinuxWebAppApplicationStack`

The application stack defines the runtime for the Web App.
Exactly one runtime must be specified: dotnet_version, node_version,
python_version, php_version, go_version, java_version
(with java_server + java_server_version), or docker.

- rule: java_version, java_server, and java_server_version must be set together

### spec.siteConfig.applicationStack.dotnetVersion

`string`

.NET runtime version.

Valid values: "3.1", "5.0", "6.0", "7.0", "8.0", "9.0", "10.0"

- rule: dotnet_version must be one of: 3.1, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0

### spec.siteConfig.applicationStack.nodeVersion

`string`

Node.js runtime version.
Linux Web Apps use the LTS variant identifiers.

Valid values: "12-lts", "14-lts", "16-lts", "18-lts", "20-lts",
"22-lts", "24-lts"

- rule: node_version must be one of: 12-lts, 14-lts, 16-lts, 18-lts, 20-lts, 22-lts, 24-lts

### spec.siteConfig.applicationStack.pythonVersion

`string`

Python runtime version.

Valid values: "3.7", "3.8", "3.9", "3.10", "3.11", "3.12", "3.13"

- rule: python_version must be one of: 3.7, 3.8, 3.9, 3.10, 3.11, 3.12, 3.13

### spec.siteConfig.applicationStack.phpVersion

`string`

PHP runtime version.

Valid values: "7.4", "8.0", "8.1", "8.2", "8.3", "8.4"

- rule: php_version must be one of: 7.4, 8.0, 8.1, 8.2, 8.3, 8.4

### spec.siteConfig.applicationStack.goVersion

`string`

Go runtime version.

**Deprecated**: Go support on Azure App Service is limited and
not recommended for new deployments. Consider containerized deployment
via the docker block instead.

Valid values: "1.18", "1.19"

- rule: go_version must be one of: 1.18, 1.19

### spec.siteConfig.applicationStack.javaVersion

`string`

Java runtime version.
When using Java, you must also set java_server and java_server_version
to specify the application server.

Valid values: "8", "11", "17", "21"

RequiredWith: java_server, java_server_version

- rule: java_version must be one of: 8, 11, 17, 21

### spec.siteConfig.applicationStack.javaServer

`enum`

Java application server type. Specifies the servlet container or
application server runtime. Unset is only valid when java_version is
also unset.

RequiredWith: java_version, java_server_version

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_linux_web_app_java_server_unspecified` -- Not specified -- only valid when no Java runtime is selected.
- `JAVA_SE` -- Java SE: embedded server (e.g. a Spring Boot executable JAR).
- `TOMCAT` -- Apache Tomcat servlet container.
- `JBOSSEAP` -- Red Hat JBoss EAP application server.

### spec.siteConfig.applicationStack.javaServerVersion

`string`

Java application server version.
The valid values depend on the selected java_server:

JAVA_SE: "8", "11", "17", "21" (SE version, typically matches java_version)
TOMCAT: "8.5", "9.0", "10.0", "10.1" (and minor versions)
JBOSSEAP: "7", "7.4", "8.0" (and minor versions)

No CEL whitelist is applied because the valid combinations across
server types and Azure region availability are too numerous.

RequiredWith: java_version, java_server

### spec.siteConfig.applicationStack.docker

`AzureLinuxWebAppDockerConfig`

Docker container configuration.
Runs a custom container image as the Web App runtime.
Any web server that listens on the configured port (default 8080)
can be deployed as a containerized Web App.

### spec.siteConfig.applicationStack.docker.registryUrl

`string` · required

The URL of the container registry.
Examples: "https://myregistry.azurecr.io", "https://ghcr.io"

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.siteConfig.applicationStack.docker.imageName

`string` · required

The container image name (without tag).
Example: "myorg/my-web-app"

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

### spec.siteConfig.alwaysOn

`bool` · optional (explicit presence)

Keep the Web App always loaded in memory.
When true, the app is never unloaded due to inactivity.

Critical for Standard and Premium plans -- without this, the app
may be unloaded after idle periods, causing cold start latency.

Not supported on Free (F1) or Shared (D1) tiers -- Azure rejects
always_on there at apply time (the plan's SKU is not visible to
manifest validation).

### spec.siteConfig.appCommandLine

`string`

Custom startup command for the Web App.
Overrides the default startup behavior. Useful for custom Docker
containers or runtimes that need specific initialization.
Example: "gunicorn --bind=0.0.0.0 --timeout 600 app:app"

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

Recommended for production deployments. Common paths: "/health",
"/healthz", "/api/health".

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

- `azure_linux_web_app_tls_version_unspecified` -- Not specified -- deploys TLS_1_2.
- `TLS_1_0` -- TLS 1.0 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_1` -- TLS 1.1 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_2` -- TLS 1.2 -- the industry floor (the default).
- `TLS_1_3` -- TLS 1.3 -- the strongest option; requires modern clients.

### spec.siteConfig.scmMinimumTlsVersion

`enum`

Minimum TLS version for the SCM (Kudu) site. Unset deploys TLS_1_2.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_linux_web_app_tls_version_unspecified` -- Not specified -- deploys TLS_1_2.
- `TLS_1_0` -- TLS 1.0 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_1` -- TLS 1.1 -- legacy clients only; fails modern compliance baselines.
- `TLS_1_2` -- TLS 1.2 -- the industry floor (the default).
- `TLS_1_3` -- TLS 1.3 -- the strongest option; requires modern clients.

### spec.siteConfig.minimumTlsCipherSuite

`string`

The minimum TLS cipher suite the app accepts. Suites WEAKER than the
chosen one are rejected; Azure's cipher order (strongest first) is:
TLS_AES_256_GCM_SHA384, TLS_AES_128_GCM_SHA256,
TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384,
TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
TLS_RSA_WITH_AES_256_GCM_SHA384, TLS_RSA_WITH_AES_128_GCM_SHA256,
TLS_RSA_WITH_AES_256_CBC_SHA256, TLS_RSA_WITH_AES_128_CBC_SHA256,
TLS_RSA_WITH_AES_256_CBC_SHA, TLS_RSA_WITH_AES_128_CBC_SHA.
Leave empty to accept Azure's platform default set.

- rule: minimum_tls_cipher_suite must be one of Azure's TLS cipher suite identifiers (e.g. TLS_AES_256_GCM_SHA384, TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)

### spec.siteConfig.ftpsState

`enum`

FTPS state for the Web App -- the FTP deployment endpoint's TLS
posture. Unset deploys DISABLED (secure by default; Azure's own
default). Independent of ftp_publish_basic_authentication_enabled,
which controls whether the publishing credential works at all.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_linux_web_app_ftps_state_unspecified` -- Not specified -- deploys DISABLED (secure by default).
- `ALL_ALLOWED` -- Both plain FTP and FTPS accepted.
- `FTPS_ONLY` -- Only FTPS (encrypted) accepted.
- `DISABLED` -- The FTP endpoint is off entirely (recommended).

### spec.siteConfig.workerCount

`int32` · optional (explicit presence)

Number of worker instances for the Web App.
Controls how many instances are allocated.

Range: 1-100.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.siteConfig.http2Enabled

`bool` · optional (explicit presence)

Enable HTTP/2 protocol for the Web App.
HTTP/2 provides multiplexing, header compression, and server push
for improved performance.

Default: false

- default: `false`

### spec.siteConfig.websocketsEnabled

`bool` · optional (explicit presence)

Enable WebSocket connections for the Web App.

Default: false

- default: `false`

### spec.siteConfig.use32BitWorker

`bool` · optional (explicit presence)

Use a 32-bit worker process instead of 64-bit.
Reduces memory footprint but limits addressable memory to ~2 GB.

Note: The Azure provider defaults to true, but the platform overrides
this to false. 64-bit workers are recommended for production.

Default: false (opinionated override; Azure provider default is true)

- default: `false`

### spec.siteConfig.vnetRouteAllEnabled

`bool` · optional (explicit presence)

Route all outbound traffic from the Web App through the VNet.
Requires virtual_network_subnet_id to be set on the spec (spec-enforced).

When false (default), only RFC1918 traffic routes through the VNet.
When true, all outbound traffic (including public internet) routes
through the VNet, enabling inspection via NSG rules or a firewall.

Default: false

- default: `false`

### spec.siteConfig.loadBalancingMode

`enum`

Load balancing mode for distributing requests across instances.
Unset deploys LEAST_REQUESTS.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_linux_web_app_load_balancing_mode_unspecified` -- Not specified -- deploys LEAST_REQUESTS.
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

- `azure_linux_web_app_managed_pipeline_mode_unspecified` -- Not specified -- deploys INTEGRATED.
- `INTEGRATED` -- The modern pipeline (the default; correct for everything current).
- `CLASSIC` -- Legacy-compatibility pipeline.

### spec.siteConfig.localMysqlEnabled

`bool` · optional (explicit presence)

Enable the in-instance MySQL database (MySQL in-app). A single
MySQL process on the same instance as the app -- dev/test
convenience only; data does not survive instance moves and does not
scale out. Use AzureMysqlFlexibleServer for real workloads.

Default: false

- default: `false`

### spec.siteConfig.remoteDebuggingEnabled

`bool` · optional (explicit presence)

Enable remote debugging (Visual Studio attach). Azure supports the
current Visual Studio generation only; the platform lets Azure pick
the debugger version. Turn off outside active debugging sessions --
it holds the process for debugger attach.

Default: false

- default: `false`

### spec.siteConfig.cors

`AzureLinuxWebAppCorsSettings`

CORS (Cross-Origin Resource Sharing) configuration.
Controls which origins are allowed to make cross-origin requests to
the Web App's HTTP endpoints.

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

`[]AzureLinuxWebAppIpRestriction`

IP restriction rules for the main site.
Controls which IP addresses, service tags, or subnets can access
the Web App.

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

- `azure_linux_web_app_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
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

`AzureLinuxWebAppIpRestrictionHeaders`

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

- `azure_linux_web_app_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.scmUseMainIpRestriction

`bool` · optional (explicit presence)

Use the main site's IP restrictions for the SCM (Kudu) site.
When true, scm_ip_restrictions are ignored.

Default: false

- default: `false`

### spec.siteConfig.scmIpRestrictions

`[]AzureLinuxWebAppIpRestriction`

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

- `azure_linux_web_app_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
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

`AzureLinuxWebAppIpRestrictionHeaders`

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

- `azure_linux_web_app_ip_restriction_action_unspecified` -- Not specified -- deploys ALLOW.
- `ALLOW`
- `DENY`

### spec.siteConfig.containerRegistryUseManagedIdentity

`bool` · optional (explicit presence)

Use managed identity for pulling container images from Azure Container Registry.
Requires the Web App's identity to have AcrPull role on the registry.

Default: false

- default: `false`

### spec.siteConfig.containerRegistryManagedIdentityClientId

`string`

Client ID of the managed identity used for ACR image pulls.
Only used when container_registry_use_managed_identity is true and
a user-assigned identity (not system-assigned) should be used.

### spec.siteConfig.autoHealSetting

`AzureLinuxWebAppAutoHealSetting`

Auto-heal: recycle the app automatically when a trigger condition is
met (request volume, status-code patterns, or slow requests). The
only heal action Linux supports is recycling the process -- the
platform sends it implicitly, so only the trigger and the optional
process-uptime guard are configured here.

### spec.siteConfig.autoHealSetting.trigger

`AzureLinuxWebAppAutoHealTrigger` · required

The condition(s) that fire the recycle. At least one trigger kind
must be set.

- rule: {"required":true}
- rule: auto_heal_setting.trigger requires at least one of: requests, status_codes, slow_request, slow_request_with_path

### spec.siteConfig.autoHealSetting.trigger.requests

`AzureLinuxWebAppAutoHealRequestsTrigger`

Fire on total request volume: more than `count` requests within
`interval`.

### spec.siteConfig.autoHealSetting.trigger.requests.count

`int32` · required

Request count threshold. Minimum: 1.

- rule: {"required":true,"int32":{"gte":1}}

### spec.siteConfig.autoHealSetting.trigger.requests.interval

`string` · required

The sliding window, hh:mm:ss (e.g. "00:05:00").

- rule: interval must be hh:mm:ss (e.g. 00:05:00)
- rule: {"required":true}

### spec.siteConfig.autoHealSetting.trigger.statusCodes

`[]AzureLinuxWebAppAutoHealStatusCodeTrigger`

Fire on status-code patterns: more than `count` responses in the
configured code or range within `interval`.

### spec.siteConfig.autoHealSetting.trigger.statusCodes[].statusCodeRange

`string` · required

The status code ("500") or range ("500-599") to count.

- rule: status_code_range must be a status code (e.g. 500) or a range (e.g. 500-599)
- rule: {"required":true}

### spec.siteConfig.autoHealSetting.trigger.statusCodes[].count

`int32` · required

Response count threshold. Minimum: 1.

- rule: {"required":true,"int32":{"gte":1}}

### spec.siteConfig.autoHealSetting.trigger.statusCodes[].interval

`string` · required

The sliding window, hh:mm:ss.

- rule: interval must be hh:mm:ss (e.g. 00:05:00)
- rule: {"required":true}

### spec.siteConfig.autoHealSetting.trigger.statusCodes[].subStatus

`int32` · optional (explicit presence)

IIS sub-status to narrow a single status code. Only valid when
status_code_range is a single code.

- rule: {"int32":{"gte":0}}

### spec.siteConfig.autoHealSetting.trigger.statusCodes[].win32StatusCode

`int32` · optional (explicit presence)

Win32 status to narrow a single status code. Only valid when
status_code_range is a single code.

- rule: {"int32":{"gte":0}}

### spec.siteConfig.autoHealSetting.trigger.statusCodes[].path

`string`

Count only requests to this path. Empty counts all paths.

### spec.siteConfig.autoHealSetting.trigger.slowRequest

`AzureLinuxWebAppAutoHealSlowRequestTrigger`

Fire when more than `count` requests take longer than `time_taken`
within `interval`, across all paths.

### spec.siteConfig.autoHealSetting.trigger.slowRequest.timeTaken

`string` · required

Requests slower than this fire the counter, hh:mm:ss (e.g.
"00:00:10" = 10 seconds).

- rule: time_taken must be hh:mm:ss (e.g. 00:00:10)
- rule: {"required":true}

### spec.siteConfig.autoHealSetting.trigger.slowRequest.interval

`string` · required

The sliding window, hh:mm:ss.

- rule: interval must be hh:mm:ss (e.g. 00:05:00)
- rule: {"required":true}

### spec.siteConfig.autoHealSetting.trigger.slowRequest.count

`int32` · required

Slow-request count threshold. Minimum: 1.

- rule: {"required":true,"int32":{"gte":1}}

### spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath

`[]AzureLinuxWebAppAutoHealSlowRequestWithPathTrigger`

Fire on slow requests to specific paths (each entry carries its own
path filter).

### spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath[].timeTaken

`string` · required

Requests slower than this fire the counter, hh:mm:ss.

- rule: time_taken must be hh:mm:ss (e.g. 00:00:10)
- rule: {"required":true}

### spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath[].interval

`string` · required

The sliding window, hh:mm:ss.

- rule: interval must be hh:mm:ss (e.g. 00:05:00)
- rule: {"required":true}

### spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath[].count

`int32` · required

Slow-request count threshold. Minimum: 1.

- rule: {"required":true,"int32":{"gte":1}}

### spec.siteConfig.autoHealSetting.trigger.slowRequestWithPath[].path

`string`

Count only requests to this path.

### spec.siteConfig.autoHealSetting.minimumProcessExecutionTime

`string`

Minimum process uptime (hh:mm:ss) before the recycle action may
fire -- prevents recycle loops on a crashing process. Example:
"00:05:00". Unset lets Azure apply its default.

- rule: minimum_process_execution_time must be hh:mm:ss (e.g. 00:05:00)

### spec.appSettings

`map<string, string>`

Application settings (environment variables) for the Web App.
Key-value pairs that are available to the application at runtime via
environment variables.

Common use cases: database connection strings, API keys, feature flags,
third-party service configuration, and the app-setting-named secrets
that auth_settings_v2 providers reference.

### spec.connectionStrings

`[]AzureLinuxWebAppConnectionString`

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

- `azure_linux_web_app_connection_string_type_unspecified` -- Not specified -- invalid; pick the service type (CUSTOM for anything without a dedicated type).
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

`AzureLinuxWebAppStickySettings`

Settings pinned to the production slot during slot swaps.
Named app settings and connection strings listed here do NOT move with
the app content when a staging slot is swapped into production --
exactly what slot-specific values (staging database URLs, slot-scoped
API keys) need.

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
When provided, the Web App can be configured to send telemetry
(requests, dependencies, exceptions, traces) to Application Insights.

Uses the connection_string format (not the legacy instrumentation_key).

- references: AzureApplicationInsights (`status.outputs.connection_string`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationInsights, name: <that resource's name>, fieldPath: status.outputs.connection_string}} -- a bare string does not parse

### spec.httpsOnly

`bool` · optional (explicit presence)

Enforce HTTPS-only access to the Web App.
When true, all HTTP requests are redirected to HTTPS.

Default: true (secure by default, unlike the Azure API default of false)

- default: `true`

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Enable or disable public network access to the Web App.
When false, the Web App is only accessible via VNet integration
or Private Endpoints.

Default: true

- default: `true`

### spec.enabled

`bool` · optional (explicit presence)

Enable or disable the Web App.
When false, the Web App is stopped and does not serve traffic,
but the resource still exists and incurs plan-level costs.
Useful for temporarily disabling an app without deleting it.

Default: true

- default: `true`

### spec.clientAffinityEnabled

`bool` · optional (explicit presence)

Enable client affinity (ARR session affinity) for the Web App.
When true, Azure uses ARR (Application Request Routing) cookies to
route subsequent requests from a client to the same instance. This is
useful for stateful applications that store session data in memory.

For stateless apps (recommended), disable this for better load distribution.

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

- `azure_linux_web_app_client_certificate_mode_unspecified` -- Not specified -- deploys OPTIONAL.
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

The subnet ID for VNet integration. When provided, the Web App's
outbound traffic routes through this subnet, enabling access to
VNet-connected resources (databases, Redis, etc.) without public endpoints.

The subnet must be delegated to Microsoft.Web/serverFarms.
Not supported on Free (F1) and Shared (D1) tiers.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.vnetImagePullEnabled

`bool` · optional (explicit presence)

Pull container images over the VNet integration (instead of over the
public internet). Required for images in registries that are only
reachable privately (e.g. an ACR locked to a private endpoint).
Requires virtual_network_subnet_id.

Default: false

- default: `false`

### spec.virtualNetworkBackupRestoreEnabled

`bool` · optional (explicit presence)

Route the backup/restore traffic of the `backup` block over the VNet
integration -- needed when the backup storage account is firewalled to
the VNet. Requires virtual_network_subnet_id.

Default: false

- default: `false`

### spec.identity

`AzureLinuxWebAppIdentity`

Managed identity configuration for the Web App.
Enables the app to authenticate with Azure services (Key Vault, Storage,
ACR, etc.) without managing credentials.

When identity is configured with SYSTEM_ASSIGNED, the web app gets
a system-assigned identity whose principal_id and tenant_id are exported
as outputs.

- rule: identity_ids is required when type includes USER_ASSIGNED, and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

The identity model: Azure-managed (tied to the app's lifecycle),
bring-your-own (independent lifecycle), or both.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_linux_web_app_identity_type_unspecified` -- Not specified -- invalid; pick an explicit identity model.
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
When the Web App uses Key Vault references in app_settings
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

`[]AzureLinuxWebAppStorageMount`

Azure Storage Account mounts for the Web App.
Mounts Azure File Shares or Blob containers as directories accessible
to the application code at runtime.

### spec.storageMounts[].name

`string` · required

Unique name for this mount within the Web App.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.storageMounts[].type

`enum` · required

What is being mounted: an Azure File Share (read-write) or a Blob
container (read-only).

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_linux_web_app_storage_mount_type_unspecified` -- Not specified -- invalid; pick the storage service.
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

### spec.logs

`AzureLinuxWebAppLogs`

Logging configuration for the Web App.
Controls application-level logging, HTTP request logging, failed request
tracing, and detailed error messages. Unlike Function Apps where logging
is nested inside site_config, Web App logs are a top-level block.

### spec.logs.applicationLogs

`AzureLinuxWebAppApplicationLogs`

Application-level logging configuration.
Controls the verbosity of application logs written to the file system
or shipped to blob storage.

### spec.logs.applicationLogs.fileSystemLevel

`enum`

Log level for the file system logger. Unset deploys ERROR.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_linux_web_app_log_level_unspecified` -- Not specified -- deploys ERROR.
- `OFF` -- No logging.
- `ERROR` -- Only errors.
- `WARNING` -- Errors and warnings.
- `INFORMATION` -- Errors, warnings, and informational messages.
- `VERBOSE` -- All messages including debug/trace.

### spec.logs.applicationLogs.azureBlobStorage

`AzureLinuxWebAppBlobStorageLogs`

Ship application logs to an Azure Blob container instead of (or in
addition to) the instance file system -- durable, queryable log
storage that survives instance recycling.

### spec.logs.applicationLogs.azureBlobStorage.level

`enum` · required

Log level shipped to blob storage.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_linux_web_app_log_level_unspecified` -- Not specified -- deploys ERROR.
- `OFF` -- No logging.
- `ERROR` -- Only errors.
- `WARNING` -- Errors and warnings.
- `INFORMATION` -- Errors, warnings, and informational messages.
- `VERBOSE` -- All messages including debug/trace.

### spec.logs.applicationLogs.azureBlobStorage.sasUrl

`string | valueFrom` · required · sensitive

SAS URL of the blob container receiving the logs. Carries a signed
credential -- treat like a password.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.logs.applicationLogs.azureBlobStorage.retentionInDays

`int32` · optional (explicit presence)

Days to retain logs in the container. 0 keeps them indefinitely.

- default: `0`
- rule: {"int32":{"gte":0}}

### spec.logs.httpLogs

`AzureLinuxWebAppHttpLogs`

HTTP request logging configuration.
Controls retention and storage limits for HTTP request/response logs.

- rule: http_logs takes exactly one destination: file_system or azure_blob_storage

### spec.logs.httpLogs.fileSystem

`AzureLinuxWebAppHttpLogsFileSystem`

Store HTTP logs on the instance file system (rotated by size/days).

### spec.logs.httpLogs.fileSystem.retentionInMb

`int32` · optional (explicit presence)

Maximum size of HTTP log files in megabytes.
When the total log size exceeds this limit, the oldest log files
are automatically deleted.

Range: 25 to 100 MB.

Default: 35

- default: `35`
- rule: {"int32":{"lte":100,"gte":25}}

### spec.logs.httpLogs.fileSystem.retentionInDays

`int32` · optional (explicit presence)

Number of days to retain HTTP log files.
Set to 0 for indefinite retention (limited only by retention_in_mb).

Default: 0 (indefinite, limited by disk quota)

- default: `0`
- rule: {"int32":{"gte":0}}

### spec.logs.httpLogs.azureBlobStorage

`AzureLinuxWebAppHttpLogsBlobStorage`

Ship HTTP logs to an Azure Blob container (durable storage).

### spec.logs.httpLogs.azureBlobStorage.sasUrl

`string | valueFrom` · required · sensitive

SAS URL of the blob container receiving the logs. Carries a signed
credential -- treat like a password.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.logs.httpLogs.azureBlobStorage.retentionInDays

`int32` · optional (explicit presence)

Days to retain logs in the container. 0 keeps them indefinitely.

- default: `0`
- rule: {"int32":{"gte":0}}

### spec.logs.failedRequestTracing

`bool` · optional (explicit presence)

Enable failed request tracing.
When true, Azure captures detailed traces for failed HTTP requests
(status code >= 400), including the request pipeline stages, timing,
and any errors. Useful for diagnosing intermittent failures.

Default: false

- default: `false`

### spec.logs.detailedErrorMessages

`bool` · optional (explicit presence)

Enable detailed error messages in HTTP error responses.
When true, the Web App returns detailed error pages for HTTP errors
instead of generic error pages. Useful for development and debugging.

**Security note**: Disable in production to avoid leaking internal
implementation details in error responses.

Default: false

- default: `false`

### spec.backup

`AzureLinuxWebAppBackup`

Scheduled backups of the app's content and configuration to an Azure
Storage container (referenced by SAS URL). Requires Standard tier or
above. Restore is an operational action in the portal/CLI, not a
manifest field.

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

`AzureLinuxWebAppBackupSchedule` · required

When and how often backups run, and how long they are kept.

- rule: {"required":true}

### spec.backup.schedule.frequencyInterval

`int32` · required

How many frequency_units pass between backups (e.g. 1 DAY = daily,
12 HOUR = twice a day). Range: 0-1000.

- rule: {"required":true,"int32":{"lte":1000,"gte":1}}

### spec.backup.schedule.frequencyUnit

`enum` · required

The unit frequency_interval counts in.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_linux_web_app_backup_frequency_unit_unspecified` -- Not specified -- invalid; pick DAY or HOUR.
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

`AzureLinuxWebAppAuthSettingsV2`

App Service built-in authentication (Easy Auth v2). When enabled,
Azure authenticates requests at the platform layer -- before they
reach application code -- against any of the configured identity
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

- `azure_linux_web_app_unauthenticated_action_unspecified` -- Not specified -- deploys REDIRECT_TO_LOGIN_PAGE.
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

Paths that skip authentication entirely (e.g. ["/health",
"/public/*"]).

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

- `azure_linux_web_app_forward_proxy_convention_unspecified` -- Not specified -- deploys FORWARD_PROXY_NO_PROXY.
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

`AzureLinuxWebAppAuthV2Login` · required

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

- `azure_linux_web_app_cookie_expiration_convention_unspecified` -- Not specified -- deploys FIXED_TIME.
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

`AzureLinuxWebAppAuthV2Apple`

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

`AzureLinuxWebAppAuthV2ActiveDirectory`

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

`AzureLinuxWebAppAuthV2StaticWebApp`

Azure Static Web Apps authentication (when fronted by one).

### spec.authSettingsV2.azureStaticWebAppV2.clientId

`string` · required

The Static Web App's client ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.authSettingsV2.customOidcV2

`[]AzureLinuxWebAppAuthV2CustomOidc`

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

`AzureLinuxWebAppAuthV2Facebook`

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

`AzureLinuxWebAppAuthV2Github`

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

`AzureLinuxWebAppAuthV2Google`

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

`AzureLinuxWebAppAuthV2Microsoft`

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

`AzureLinuxWebAppAuthV2Twitter`

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

Free-form Azure resource tags applied to the Web App, merged over the
platform's metadata-derived tags (user tags win on key collision) --
the hooks for cost allocation, chargeback reports, and Azure Policy
governance rules that filter or group by them. Updatable in place.

## Validation Rules

- `web_app_vnet_image_pull_requires_subnet`: vnet_image_pull_enabled requires virtual_network_subnet_id (the image pull rides the VNet integration)
- `web_app_vnet_backup_restore_requires_subnet`: virtual_network_backup_restore_enabled requires virtual_network_subnet_id (backup traffic rides the VNet integration)
- `web_app_vnet_route_all_requires_subnet`: site_config.vnet_route_all_enabled requires virtual_network_subnet_id (there is no VNet to route through otherwise)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureLinuxWebApp, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.web_app_id` | `string` | The Azure Resource Manager ID of the Web App. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Web/sites/{name} |
| `status.outputs.default_hostname` | `string` | The default hostname of the Web App. Format: {name}.azurewebsites.net This is the primary endpoint for the web app. Custom domains can be added via Azure portal or DNS configuration. |
| `status.outputs.outbound_ip_addresses` | `[]string` | Outbound IP addresses used by the Web App. These IPs should be allowed in downstream firewall rules (e.g., database firewall, third-party API whitelists). Note: On shared/free tiers, outbound IPs are shared across the region and may change. On Dedicated/Premium plans, IPs are stable. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the system-assigned managed identity. Populated only when the Web App has a system-assigned identity (identity.type includes "SystemAssigned"). Used for granting RBAC roles: e.g., "Key Vault Secrets User", "Storage Blob Data Contributor". |
| `status.outputs.identity_tenant_id` | `string` | The tenant ID of the system-assigned managed identity. Paired with identity_principal_id for RBAC configuration. |
| `status.outputs.custom_domain_verification_id` | `string` | The custom domain verification ID. Used when binding custom domains to the Web App. Add this value as a TXT record at `asuid.{custom-domain}` to verify domain ownership. |
| `status.outputs.kind` | `string` | The resource kind string as reported by Azure. Example: "app,linux" |
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
| AzureFrontDoorOrigin | `spec.privateLink.privateLinkTargetId` | `status.outputs.web_app_id` |

## See Also

- [Overview](../README.md)

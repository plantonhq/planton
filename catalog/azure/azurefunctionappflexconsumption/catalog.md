# Azure Function App Flex Consumption

Deploys an Azure Function App on the Flex Consumption plan -- Azure's newest serverless Functions hosting model, with per-instance memory selection, a configurable scale-out ceiling, and always-ready instance pools that eliminate cold starts. Three things distinguish it from classic hosting: deployment storage is a blob container you provide, scale is per-instance and configurable, and the runtime is declared flat (containers are not supported). Classic Consumption, Elastic Premium, and Dedicated-plan apps are the separate Azure Function App kind; an idle flex app with no warm pools costs nothing.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Flex Consumption Function App** -- the Microsoft.Web site with its runtime declaration, deployment-storage binding, scale configuration (instance memory, fan-out ceiling, HTTP concurrency, always-ready pools), site configuration, app settings, connection strings, managed identity, and Easy Auth v2
- **Azure Tags** -- Planton-derived metadata tags merged with the manifest's `tags` (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Azure Subscription

- **An FC1 service plan** -- an AzureServicePlan with `skuName: FLEX_CONSUMPTION_FC1`; Azure rejects flex apps on any other tier, and the check runs against the LIVE plan. One FC1 plan hosts many flex apps and has no idle compute cost.
- **A blob container for deployment storage** -- an AzureStorageAccount plus an AzureStorageContainer; `storageContainerEndpoint` composes the account's blob endpoint with the container name, and the container must exist before the app is created.
- **Regional availability** -- Flex Consumption is not offered in every region; the plan and app must share a region where it is, and region is ForceNew.
- **A "Storage Blob Data Contributor" grant** (only for identity-based storage auth) -- required before package deployments work; ARM does NOT check it at app create, so a green create is not proof the package path works.

## Deploy

### Console

Open the deployment store, find **Azure Function App Flex Consumption**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the plan and storage bindings, runtime, scale dials, and site configuration. Start from the **Node HTTP API**, **Identity-Secured Worker**, or **Entra-Protected API** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFunctionAppFlexConsumption
metadata:
  name: orders-api
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-apps
      fieldPath: status.outputs.resource_group_name
  functionAppName: orders-api-acme
  region: eastus
  servicePlanId:
    valueFrom:
      kind: AzureServicePlan
      name: flex-plan
      fieldPath: status.outputs.service_plan_id
  storageContainerEndpoint: https://acmeappstorage.blob.core.windows.net/deployments
  storageAuthenticationType: STORAGE_ACCOUNT_CONNECTION_STRING
  storageAccessKey:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.primary_access_key
  runtimeName: NODE
  runtimeVersion: "20"
  maximumInstanceCount: 100
  alwaysReady:
    - name: http
      instanceCount: 1
  siteConfig:
    minimumTlsVersion: TLS_1_2
```

```shell
planton apply -f flex-function-app.yaml
```

This creates a Node 20 flex app with one always-ready HTTP instance (its only idle cost), a 100-instance fan-out ceiling, and connection-string deployment storage. An Infra Job tracks the provisioning in real time.

### InfraChart

When the plan, storage, and app deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-apps
      fieldPath: status.outputs.resource_group_name
  functionAppName: worker-acme
  region: eastus
  servicePlanId:
    valueFrom:
      kind: AzureServicePlan
      name: flex-plan
      fieldPath: status.outputs.service_plan_id
  storageContainerEndpoint: https://acmeappstorage.blob.core.windows.net/deployments
  storageAuthenticationType: SYSTEM_ASSIGNED_IDENTITY
  runtimeName: PYTHON
  runtimeVersion: "3.12"
  identity:
    type: SYSTEM_ASSIGNED
  siteConfig: {}
  applicationInsightsConnectionString:
    valueFrom:
      kind: AzureApplicationInsights
      name: apm
      fieldPath: status.outputs.connection_string
```

The InfraPipeline resolves the dependency graph, deploys the plan, storage, and Application Insights first, then provisions the app with the resolved values.

## Key Configuration

These are the most important decisions when configuring a flex app. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The plan tier is a hard gate** -- Azure rejects app creation on any plan whose SKU is not FC1, and the check runs against the LIVE plan, so a plan re-tiered after the reference was wired still fails. Keep flex apps on dedicated FC1 plans. `functionAppName` (globally unique -- it forms `{name}.azurewebsites.net`), `region`, `resourceGroup`, and `servicePlanId` are all ForceNew.

**Deployment storage: pick storageAuthenticationType deliberately** -- `STORAGE_ACCOUNT_CONNECTION_STRING` is the simplest (Azure derives the connection string from your `storageAccessKey` and manages it as a hidden app setting; a rotated key must be updated in the manifest). `SYSTEM_ASSIGNED_IDENTITY` is credential-free, but the grant is day-2 by construction -- the identity exists only after the app does, and package deployments fail until "Storage Blob Data Contributor" lands on the storage account. `USER_ASSIGNED_IDENTITY` lets you pre-grant before the app exists; the same identity must be attached via `identity.identityIds` AND named in `storageUserAssignedIdentityId`.

**Always-ready pools are the only idle cost -- and the only cold-start cure** -- everything else bills per execution. Each `alwaysReady` entry keeps N instances warm for a scope (`http`, `durable`, `blob`, or `function:{name}`) and bills for their uptime. The counts' sum must stay within `maximumInstanceCount` (Azure enforces this at apply time -- a sum rule manifest validation cannot express). Azure lower-cases pool names on save.

**The scale dials are the cost levers** -- `instanceMemoryInMb` picks the per-instance size (512, 2048, or 4096; default 2048), `maximumInstanceCount` (default 100, up to 1000) caps the fan-out bill on traffic spikes, and `httpConcurrency` sets how many requests each instance absorbs before Azure scales out further.

**The write-only class: what imports cannot recover** -- Azure never returns `storageAccessKey`, `zipDeployFile`, or `siteConfig.appServiceLogs` on reads; re-supplying them in the manifest is expected, not drift. `siteConfig.elasticInstanceMinimum` is accepted on the wire but never returned on this hosting model -- `alwaysReady` is the flex-native warm-instance mechanism, prefer it.

**Easy Auth secrets live in app settings, never in the auth block** -- every `authSettingsV2` identity provider references its client secret by APP SETTING NAME (`clientSecretSettingName`); put the actual value in `appSettings`, ideally as a Key Vault reference (`@Microsoft.KeyVault(SecretUri=...)`), and pin it with `stickySettings` if you use deployment slots. For APIs, set `unauthenticatedAction: RETURN_401`; the login-redirect default suits browser-facing apps only.

**Hardening the deployment surface** -- `webdeployPublishBasicAuthenticationEnabled: false` closes the classic username/password publishing path and forces identity-based deployment. While basic auth IS enabled, the `site_credential_password` output is a working deploy credential -- treat it like an admin password.

**Network posture** -- `virtualNetworkSubnetId` routes outbound traffic through a subnet delegated to `Microsoft.App/environments`; `siteConfig.vnetRouteAllEnabled` extends that to ALL outbound traffic (and requires the subnet). `siteConfig.ipRestrictions` gate inbound access, including the Front Door lock-down shape (`AzureFrontDoor.Backend` service tag plus an `xAzureFdid` header filter) that keeps traffic from bypassing the edge.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureServicePlan** (FC1 SKU) | `servicePlanId` | `status.outputs.service_plan_id` |
| **AzureStorageAccount** (connection-string mode) | `storageAccessKey` | `status.outputs.primary_access_key` |
| **AzureUserAssignedIdentity** (user-assigned storage auth) | `storageUserAssignedIdentityId`, `identity.identityIds` | `status.outputs.identity_id` |
| **AzureApplicationInsights** (optional) | `applicationInsightsConnectionString` | `status.outputs.connection_string` |
| **AzureSubnet** (VNet integration) | `virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| **AzureFrontDoorProfile** (edge lock-down) | `siteConfig.ipRestrictions[].headers.xAzureFdid` | `status.outputs.resource_guid` |

`storageContainerEndpoint` is a plain string composed from an AzureStorageAccount's blob endpoint plus an AzureStorageContainer's name (the container kind deliberately exports no URL).

### What This Kind Provides

After provisioning, `status.outputs` contains values consumed by users, DNS configuration, and RBAC wiring -- this is a leaf resource; no catalog kind references it downstream:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `default_hostname` | The app's FQDN (`{name}.azurewebsites.net`) | DNS records, upstream proxies, client configuration |
| `identity_principal_id` | The system-assigned identity's principal ID | RBAC grants: "Storage Blob Data Contributor" for identity-based deployment storage, "Key Vault Secrets User" |
| `possible_outbound_ip_addresses` | Every outbound IP the platform could ever route through (superset of the active set) | Downstream firewall allowlists that must survive scale events |
| `custom_domain_verification_id` | The domain-ownership token | The TXT record at `asuid.{custom-domain}` when binding custom domains |
| `site_credential_password` | The Kudu/SCM basic-auth deploy credential (secret-bearing) | CI/CD publishing while basic auth is enabled -- revoke by disabling the toggle |

## Common Patterns

**Warm-path HTTP API** -- one always-ready HTTP instance answers instantly; everything else bills per execution, with the fan-out ceiling capping spike cost. Start from the **Node HTTP API** preset.

**Credential-free worker** -- identity-based deployment storage plus basic-auth publishing disabled: no storage key or deploy password exists anywhere in the configuration, at the price of a day-2 RBAC grant before the first package deploy. Start from the **Identity-Secured Worker** preset.

**Platform-authenticated API** -- Easy Auth v2 validates Entra ID tokens before requests reach function code, with the client secret held in app settings as a Key Vault reference and telemetry wired to Application Insights. Start from the **Entra-Protected API** preset.

## Works With

- [**Azure Service Plan**](/infra-catalog/azure-service-plan) -- the FC1 plan that hosts the app
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the group the app lives in
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) / [**Azure Storage Container**](/infra-catalog/azure-storage-container) -- the deployment-storage container and its account
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- pre-grantable identities for storage auth and service access
- [**Azure Application Insights**](/infra-catalog/azure-application-insights) -- APM telemetry via the connection-string reference
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- VNet integration for outbound traffic
- [**Azure Front Door Profile**](/infra-catalog/azure-front-door-profile) -- the edge whose ID locks down inbound access
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- the RBAC grants the app's identity needs on storage and other services

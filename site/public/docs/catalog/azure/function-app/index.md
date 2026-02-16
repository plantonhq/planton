---
title: "Function App"
description: "Function App deployment documentation"
icon: "package"
order: 100
componentName: "azurefunctionapp"
---

# Azure Function App

Deploys an Azure Linux Function App -- a serverless compute platform for event-driven workloads supporting HTTP triggers, queue triggers, timer schedules, and more. The component provides full configuration of the application runtime stack, managed identity, VNet integration, Application Insights telemetry, IP restrictions, CORS, storage mounts, and connection strings.

## What Gets Created

When you deploy an AzureFunctionApp resource, OpenMCF provisions:

- **Linux Function App** -- an `appservice.LinuxFunctionApp` resource in the specified region and resource group, configured with the chosen runtime stack, storage binding, Application Insights connection, and operational settings
- **Managed Identity** -- created only when `identity` is configured, provides credential-free authentication to Azure services (Key Vault, Storage, ACR)
- **VNet Integration** -- created only when `virtualNetworkSubnetId` is set, routes outbound traffic through a VNet subnet for private connectivity
- **Azure Tags** -- resource metadata tags applied to the function app for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the function app will be created (can reference an AzureResourceGroup resource)
- **An Azure Service Plan** providing compute resources -- Consumption (`Y1`) for pay-per-execution, Elastic Premium (`EP1`-`EP3`) for pre-warmed instances, or Dedicated (`B1`-`P3v3`) for reserved capacity
- **An Azure Storage Account** for Function App runtime state (trigger management, logs, coordination)
- **A globally unique app name** -- the name becomes the hostname `{name}.azurewebsites.net`

## Quick Start

Create a file `functionapp.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: my-func
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureFunctionApp.my-func
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-func-app
  servicePlanId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Web/serverFarms/my-plan
  storageAccountName: mystorageacct
  storageAccountAccessKey: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx=="
  siteConfig:
    applicationStack:
      pythonVersion: "3.11"
```

Deploy:

```shell
openmcf apply -f functionapp.yaml
```

This creates a Python 3.11 Function App on the specified Service Plan with HTTPS-only access, TLS 1.2, and Functions runtime v4.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the function app. **ForceNew**. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. **ForceNew**. | Required |
| `name` | `string` | Globally unique app name. Becomes `{name}.azurewebsites.net`. **ForceNew**. | Required, 2-60 characters, pattern `^[a-zA-Z0-9][a-zA-Z0-9-]{0,58}[a-zA-Z0-9]$` |
| `servicePlanId` | `StringValueOrRef` | Service Plan providing compute resources. Can reference an AzureServicePlan resource via `valueFrom`. | Required |
| `storageAccountName` | `StringValueOrRef` | Storage Account name for runtime state. Can reference an AzureStorageAccount resource via `valueFrom`. | Required |
| `siteConfig` | `object` | Site configuration containing the application stack. | Required |
| `siteConfig.applicationStack` | `object` | Runtime selection. Exactly one runtime must be set: `dotnetVersion`, `nodeVersion`, `pythonVersion`, `javaVersion`, `powershellCoreVersion`, `docker`, or `useCustomRuntime`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `storageAccountAccessKey` | `StringValueOrRef` | -- | Storage Account access key (sensitive). Conflicts with `storageUsesManagedIdentity`. |
| `storageUsesManagedIdentity` | `bool` | `false` | Use managed identity for storage access instead of access key. |
| `functionsExtensionVersion` | `string` | `"~4"` | Azure Functions runtime version. |
| `httpsOnly` | `bool` | `true` | Redirect all HTTP to HTTPS. |
| `publicNetworkAccessEnabled` | `bool` | `true` | Allow public internet access. |
| `builtinLoggingEnabled` | `bool` | `true` | Enable legacy AzureWebJobsDashboard logging. |
| `applicationInsightsConnectionString` | `StringValueOrRef` | -- | App Insights connection string for APM telemetry. Can reference an AzureApplicationInsights resource via `valueFrom`. |
| `virtualNetworkSubnetId` | `StringValueOrRef` | -- | Subnet ID for VNet integration (outbound traffic). Can reference an AzureSubnet resource via `valueFrom`. |
| `identity.type` | `string` | -- | Managed identity type: `SystemAssigned`, `UserAssigned`, or `SystemAssigned,UserAssigned`. |
| `identity.identityIds` | `StringValueOrRef[]` | `[]` | User-assigned identity IDs. Can reference AzureUserAssignedIdentity resources via `valueFrom`. |
| `appSettings` | `map<string, string>` | `{}` | Application environment variables. |
| `connectionStrings` | `list` | `[]` | Named connection strings with `name`, `type`, and `value`. |
| `siteConfig.alwaysOn` | `bool` | -- | Keep app loaded in memory. Critical for Dedicated plans. |
| `siteConfig.healthCheckPath` | `string` | -- | Health check endpoint (e.g., `/api/health`). |
| `siteConfig.appScaleLimit` | `int` | -- | Maximum scale-out instances (Consumption/EP plans). |
| `siteConfig.cors.allowedOrigins` | `string[]` | -- | CORS allowed origins. |
| `siteConfig.ipRestrictions` | `list` | `[]` | IP-based access restriction rules. |

## Examples

### Python HTTP API

A Python 3.11 Function App on a Consumption plan with Application Insights:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: python-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureFunctionApp.python-api
spec:
  region: eastus
  resourceGroup: prod-rg
  name: python-api-func
  servicePlanId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Web/serverFarms/consumption-plan
  storageAccountName: prodfuncstorage
  storageAccountAccessKey: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx=="
  applicationInsightsConnectionString: "InstrumentationKey=00000000-0000-0000-0000-000000000000;IngestionEndpoint=https://eastus-0.in.applicationinsights.azure.com/"
  siteConfig:
    applicationStack:
      pythonVersion: "3.11"
    cors:
      allowedOrigins:
        - "https://myapp.example.com"
  appSettings:
    DATABASE_URL: "postgresql://..."
```

### Docker Container Function App

A containerized Function App on an Elastic Premium plan with VNet integration:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: docker-func
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureFunctionApp.docker-func
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: docker-func-app
  servicePlanId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Web/serverFarms/ep-plan
  storageAccountName: prodfuncstorage
  storageUsesManagedIdentity: true
  virtualNetworkSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/functions
  identity:
    type: SystemAssigned
  siteConfig:
    applicationStack:
      docker:
        registryUrl: https://myregistry.azurecr.io
        imageName: myorg/my-function
        imageTag: v1.2.3
    containerRegistryUseManagedIdentity: true
    alwaysOn: true
    healthCheckPath: /api/health
    vnetRouteAllEnabled: true
```

### Using Foreign Key References

Reference OpenMCF-managed resources for the service plan, storage, and monitoring:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: ref-func
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureFunctionApp.ref-func
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-func-app
  servicePlanId:
    valueFrom:
      kind: AzureServicePlan
      name: my-plan
      field: status.outputs.plan_id
  storageAccountName:
    valueFrom:
      kind: AzureStorageAccount
      name: my-storage
      field: status.outputs.storage_account_name
  storageUsesManagedIdentity: true
  applicationInsightsConnectionString:
    valueFrom:
      kind: AzureApplicationInsights
      name: my-insights
      field: status.outputs.connection_string
  identity:
    type: SystemAssigned
  siteConfig:
    applicationStack:
      nodeVersion: "20"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `function_app_id` | `string` | Azure Resource Manager ID of the Function App |
| `default_hostname` | `string` | Default hostname (`{name}.azurewebsites.net`) |
| `outbound_ip_addresses` | `string[]` | Outbound IP addresses for firewall allowlisting |
| `identity_principal_id` | `string` | System-assigned identity principal ID (when identity is configured) |
| `identity_tenant_id` | `string` | System-assigned identity tenant ID |
| `custom_domain_verification_id` | `string` | TXT record value for custom domain verification |
| `kind` | `string` | Resource kind (e.g., `functionapp,linux`) |

## Related Components

- [AzureServicePlan](/docs/catalog/azure/service-plan) -- provides the compute tier for the Function App
- [AzureStorageAccount](/docs/catalog/azure/storage-account) -- provides runtime storage for triggers and logs
- [AzureApplicationInsights](/docs/catalog/azure/application-insights) -- provides APM telemetry collection
- [AzureResourceGroup](/docs/catalog/azure/resource-group) -- provides the resource group for app placement
- [AzureSubnet](/docs/catalog/azure/subnet) -- provides VNet integration for outbound connectivity

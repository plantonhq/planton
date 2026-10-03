# Azure Function App

Deploys a Linux Function App for event-driven serverless workloads with configurable runtime stacks (.NET, Node.js, Python, Java, PowerShell, Docker), App Service Plan binding, Storage Account integration, VNet integration, managed identity, Application Insights monitoring, and IP restriction rules. The App Service Plan decides the app's cost model and capabilities up front: Consumption is pay-per-execution with cold starts and no VNet integration, while Elastic Premium and Dedicated plans unlock pre-warmed instances, Docker images, and private networking.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Linux Function App** -- a serverless compute resource in the specified Azure region, bound to an App Service Plan, with the chosen application stack, app settings, connection strings, and security configuration
- **Storage Binding** -- connects the Function App to an Azure Storage Account for trigger management, execution logs, and internal coordination, using either an access key or managed identity
- **Identity** -- created only when `identity` is configured; SystemAssigned, UserAssigned, or both for credential-free access to Key Vault, Storage, ACR, and other Azure services
- **VNet Integration** -- created only when `virtualNetworkSubnetId` is provided; routes outbound traffic through the specified subnet for private resource access
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Function App will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **An App Service Plan** that provides compute resources. Consumption (Y1) for pay-per-execution, Elastic Premium (EP*) for pre-warmed instances, or Dedicated (B*/S*/P*) for reserved capacity. Provide the plan ID directly or reference an AzureServicePlan Infra Component via ValueFromRef.
- **An Azure Storage Account** for Functions runtime state. Provide the account name directly or reference an AzureStorageAccount Infra Component via ValueFromRef.
- **Application Insights** (optional) for APM telemetry. Provide the connection string directly or reference an AzureApplicationInsights Infra Component via ValueFromRef.
- **A VNet Subnet** (optional) delegated to `Microsoft.Web/serverFarms` for VNet integration. Not supported on Consumption plans.

## Deploy

### Console

Open the deployment store, find **Azure Function App**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Python HTTP API** preset in the [Presets](#presets) tab to pre-populate a working Python 3.12 serverless API configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFunctionApp
metadata:
  name: order-processor
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  functionAppName: order-processor
  servicePlanId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.Web/serverfarms/consumption-plan"
  storageAccountName:
    value: "acmefuncstorage"
  storageAccountAccessKey:
    value: $secret/prod-func-storage-key
  siteConfig:
    applicationStack:
      pythonVersion: "3.12"
```

```shell
planton apply -f function-app.yaml
```

This creates a Python 3.12 Function App on the specified Service Plan with HTTPS enforcement, TLS 1.2, and FTPS disabled -- no VNet integration, no managed identity, no Application Insights. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Function App to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  servicePlanId:
    valueFrom:
      kind: AzureServicePlan
      name: elastic-premium
      fieldPath: status.outputs.service_plan_id
  storageAccountName:
    valueFrom:
      kind: AzureStorageAccount
      name: func-storage
      fieldPath: status.outputs.storage_account_name
  virtualNetworkSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: functions-subnet
      fieldPath: status.outputs.subnet_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, service plan, storage account, and subnet first, then provisions the Function App with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Function App. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Application stack** -- Exactly one runtime must be chosen in `siteConfig.applicationStack`: .NET (`dotnetVersion`), Node.js (`nodeVersion`), Python (`pythonVersion`), Java (`javaVersion`), PowerShell (`powershellCoreVersion`), Docker container (`docker`), or custom handler (`useCustomRuntime`). Docker-based apps require Elastic Premium or Dedicated plans.

**Service Plan tier** -- The plan determines cost model and scaling behavior. Consumption (Y1) scales to 200 instances with pay-per-execution pricing but has cold start latency. Elastic Premium (EP*) provides pre-warmed instances and VNet integration. Dedicated (B*/P*) gives fixed capacity with `alwaysOn: true` to prevent idle shutdown.

**Storage authentication** -- Functions require a Storage Account for runtime state. Provide `storageAccountAccessKey` for key-based auth, or set `storageUsesManagedIdentity: true` for the credential-free approach (requires Storage Blob Data Owner and Storage Queue Data Contributor roles on the identity).

**VNet integration** -- Set `virtualNetworkSubnetId` to route outbound traffic through a VNet subnet. Enable `siteConfig.vnetRouteAllEnabled: true` to route all traffic (including public internet) through the VNet for firewall inspection. Not supported on Consumption plans.

**Identity and Key Vault** -- Configure `identity` with SystemAssigned or UserAssigned for credential-free access to Azure services. Use Key Vault references (`@Microsoft.KeyVault(SecretUri=...)`) in `appSettings` to fetch secrets at runtime without storing them in manifests.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureServicePlan** | `servicePlanId` | `status.outputs.service_plan_id` |
| **AzureStorageAccount** | `storageAccountName` | `status.outputs.storage_account_name` |
| **AzureApplicationInsights** (optional) | `applicationInsightsConnectionString` | `status.outputs.connection_string` |
| **AzureSubnet** (optional) | `virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| **AzureUserAssignedIdentity** (optional) | `keyVaultReferenceIdentityId` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `function_app_id` | Azure Resource Manager ID of the Function App | Azure Policy assignments, diagnostic settings |
| `default_hostname` | Default FQDN (`{name}.azurewebsites.net`) | API gateway backends, DNS CNAME records |
| `outbound_ip_addresses` | Egress IP addresses used by the Function App | Database firewall rules, third-party API allowlists |
| `identity_principal_id` | Principal ID of the system-assigned managed identity | RBAC role assignments (Key Vault, Storage, ACR) |
| `identity_tenant_id` | Tenant ID of the system-assigned managed identity | RBAC configuration paired with principal ID |
| `custom_domain_verification_id` | TXT record value for custom domain verification | DNS TXT records at `asuid.{custom-domain}` |
| `possible_outbound_ip_addresses` | Every outbound IP the app may use across scale operations | The COMPLETE set for firewall rules |
| `site_credential_name` | Basic-auth publishing username (secret -- masked in the console) | Web Deploy/FTP publishing; inert when both basic-auth toggles are off |
| `site_credential_password` | Basic-auth publishing password (secret -- masked in the console) | Web Deploy/FTP publishing; inert when both basic-auth toggles are off |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Python HTTP API** -- Python 3.12 Function App with Application Insights monitoring, health check endpoint, CORS configuration, and secure defaults (HTTPS only, TLS 1.2, FTPS disabled). Suitable for REST APIs and webhook handlers on Consumption or Premium plans. Start from the **Python HTTP API** preset.

**Docker container** -- Containerized Function App running a custom Docker image from ACR with managed identity for credential-free image pulls, always-on mode, and health checks. Requires Elastic Premium or Dedicated plan. Start from the **Docker Container** preset.

**Enterprise Elastic Premium** -- Production Function App with VNet integration, managed identity for storage, Key Vault secret references, pre-warmed instances, IP restrictions, runtime scale monitoring, and full security hardening. Start from the **Enterprise Elastic Premium** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the Function App is created
- [**Azure Service Plan**](/infra-catalog/azure-service-plan) -- provides the compute tier and scaling behavior
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- provides runtime state storage for triggers and logs
- [**Azure Application Insights**](/infra-catalog/azure-application-insights) -- provides APM telemetry for monitoring and diagnostics
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- provides VNet integration for private resource access
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- provides credential-free access to Key Vault and other Azure services
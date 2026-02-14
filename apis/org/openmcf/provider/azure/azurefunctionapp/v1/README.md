# AzureFunctionApp

An Azure Linux Function App provides serverless compute for event-driven workloads -- HTTP APIs, queue processors, timer-triggered jobs, and blob-event handlers -- running on the Azure Functions runtime.

## Overview

The `AzureFunctionApp` component provisions an `azurerm_linux_function_app` resource, a serverless compute platform that executes code in response to events without managing infrastructure. It is the Azure equivalent of AWS Lambda or GCP Cloud Functions, but with a key difference: Azure Functions run on an App Service Plan, giving explicit control over the compute tier (Consumption for pay-per-execution, Elastic Premium for pre-warmed instances, or Dedicated for reserved capacity).

Every Function App requires:
- **An App Service Plan** (`AzureServicePlan`): Determines cost model, scale behavior, and features
- **A Storage Account** (`AzureStorageAccount`): For runtime state, trigger management, and execution logs
- **An application stack**: The runtime (Python, Node.js, .NET, Java, PowerShell, Docker, or custom handler)

## Key Features

- **Dual IaC support**: Both Pulumi and Terraform modules with feature parity
- **StringValueOrRef composability**: `service_plan_id`, `storage_account_name`, `virtual_network_subnet_id`, and `application_insights_connection_string` all support `valueFrom` references
- **Full-feature site_config**: Application stack selection, scaling controls, health checks, TLS settings, FTPS state, load balancing, CORS, and IP restrictions
- **Docker support**: Run custom container images as Azure Functions via the `docker` application stack
- **Managed identity**: SystemAssigned, UserAssigned, or both -- credential-free access to Azure services
- **Storage identity mode**: `storage_uses_managed_identity` for key-free storage binding
- **Connection strings**: Named, typed connection strings for database and service integrations
- **IP restrictions**: IP-based, service-tag, and VNet-based access control for both the main site and SCM (Kudu)
- **CORS**: Cross-origin resource sharing configuration for HTTP endpoints
- **Storage mounts**: Mount Azure File Shares or Blob containers as directories accessible at runtime

## When to Use

- **Event-driven processing**: Respond to queue messages, blob uploads, Event Grid events, or Cosmos DB changes
- **HTTP APIs**: Lightweight REST APIs with automatic scaling (alternative to AzureContainerApp for simple APIs)
- **Scheduled tasks**: Timer-triggered functions for cron-like jobs (billing runs, report generation, cleanup)
- **Queue processing**: Consume messages from Service Bus, Storage Queues, or Event Hubs
- **Infra charts**: Leaf resource in the `function-app-environment` infra chart (references ServicePlan, StorageAccount, AppInsights, Subnet)

## Quick Example

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: my-functions
spec:
  region: eastus
  resource_group: my-rg
  name: my-functions
  service_plan_id: /subscriptions/.../Microsoft.Web/serverfarms/my-plan
  storage_account_name: myfuncsstorage
  storage_account_access_key: <storage-access-key>
  site_config:
    application_stack:
      python_version: "3.12"
```

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region (ForceNew) |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or AzureResourceGroup ref) (ForceNew) |
| `name` | string | Yes | - | Globally unique name (`{name}.azurewebsites.net`) (ForceNew) |
| `service_plan_id` | StringValueOrRef | Yes | - | App Service Plan ARM ID (AzureServicePlan ref) |
| `storage_account_name` | StringValueOrRef | Yes | - | Storage account name (AzureStorageAccount ref) |
| `storage_account_access_key` | StringValueOrRef | No | - | Storage account access key (conflicts with managed identity) |
| `storage_uses_managed_identity` | bool | No | `false` | Use managed identity for storage (conflicts with access key) |
| `functions_extension_version` | string | No | `"~4"` | Azure Functions runtime version |
| `site_config` | SiteConfig | Yes | - | Site configuration (runtime, scaling, security) |
| `app_settings` | map | No | - | Environment variables (key-value pairs) |
| `connection_strings` | repeated | No | - | Named connection strings (name, type, value) |
| `application_insights_connection_string` | StringValueOrRef | No | - | Application Insights connection string |
| `https_only` | bool | No | `true` | Enforce HTTPS-only access |
| `public_network_access_enabled` | bool | No | `true` | Enable public network access |
| `builtin_logging_enabled` | bool | No | `true` | Enable AzureWebJobsDashboard logging |
| `virtual_network_subnet_id` | StringValueOrRef | No | - | Subnet for VNet integration (AzureSubnet ref) |
| `identity` | Identity | No | - | Managed identity (SystemAssigned, UserAssigned, or both) |
| `key_vault_reference_identity_id` | StringValueOrRef | No | - | Identity for Key Vault references |
| `client_certificate_enabled` | bool | No | `false` | Enable mTLS client certificates |
| `client_certificate_mode` | string | No | `"Optional"` | Certificate mode (Required, Optional, OptionalInteractiveUser) |
| `client_certificate_exclusion_paths` | string | No | - | Semicolon-separated paths excluded from cert validation |
| `content_share_force_disabled` | bool | No | `false` | Disable auto-created Azure Files content share |
| `storage_mounts` | repeated | No | - | Azure File Share or Blob container mounts |

## Outputs

| Output | Description |
|--------|-------------|
| `function_app_id` | ARM resource ID of the Function App |
| `default_hostname` | Default hostname (`{name}.azurewebsites.net`) |
| `outbound_ip_addresses` | Outbound IP addresses (for downstream firewall rules) |
| `identity_principal_id` | System-assigned identity principal ID (for RBAC) |
| `identity_tenant_id` | System-assigned identity tenant ID |
| `custom_domain_verification_id` | TXT record value for custom domain verification |
| `kind` | Resource kind string (e.g., `"functionapp,linux"`) |

## Downstream Usage

AzureFunctionApp is a **leaf resource** -- nothing references its outputs downstream. It consumes outputs from upstream resources via `valueFrom`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFunctionApp
metadata:
  name: event-processor
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: shared-rg
      fieldPath: status.outputs.resource_group_name
  name: event-processor
  service_plan_id:
    valueFrom:
      kind: AzureServicePlan
      name: functions-plan
      fieldPath: status.outputs.plan_id
  storage_account_name:
    valueFrom:
      kind: AzureStorageAccount
      name: func-storage
      fieldPath: status.outputs.storage_account_name
  application_insights_connection_string:
    valueFrom:
      kind: AzureApplicationInsights
      name: func-insights
      fieldPath: status.outputs.connection_string
  virtual_network_subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: functions-subnet
      fieldPath: status.outputs.subnet_id
  site_config:
    application_stack:
      python_version: "3.12"
```

## What's NOT Included (80/20 Scope)

- **auth_settings / auth_settings_v2**: Azure App Service Authentication (Easy Auth). Complex configuration surface with 20+ sub-fields. Deferred to v2 when demand materializes.
- **backup**: Automated backup configuration. Niche feature for stateful apps; most function apps are stateless.
- **sticky_settings**: App settings that don't swap during slot deployments. Requires deployment slots (not in v1).
- **Deployment slots**: Blue-green deployment via staging slots. Significant complexity; deferred to v2.
- **zip_deploy_file**: In-line ZIP deployment. Most teams use CI/CD pipelines for deployment, not in-line ZIP.
- **Windows Function Apps**: `azurerm_windows_function_app` is excluded. Linux covers the vast majority of serverless workloads.

These omissions follow the 80/20 principle: the included fields cover the vast majority of production use cases while keeping the API surface clean and maintainable.

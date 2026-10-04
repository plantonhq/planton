# AzureFunctionApp

An Azure Function App manages serverless compute infrastructure on Azure App Service (Linux), hosting event-driven functions triggered by HTTP requests, queues, timers, and Azure service events.

## Overview

The `AzureFunctionApp` component provisions an `azurerm_linux_function_app` resource. Unlike always-on Web Apps, Function Apps are event-driven and (on serverless plans) scale to zero.

Every Function App requires:
- **An App Service Plan** (`AzureServicePlan`): Consumption (`CONSUMPTION_Y1`) for pay-per-execution, Elastic Premium (`ELASTIC_PREMIUM_EP1`-`EP3`) for pre-warmed instances, or Dedicated tiers for reserved capacity
- **A storage binding**: the runtime state store -- a storage account name (with an access key or managed identity), or a Key Vault secret holding the connection string
- **An application stack**: the runtime (.NET with isolated worker, Node.js, Python, Java, PowerShell, Docker container, or custom handler)

Azure's Flex Consumption tier (FC1 plans) is a separate resource type with its own storage and scaling model, not this kind.

## Key Features

- **Dual IaC support**: Both Pulumi and Terraform modules with feature parity
- **StringValueOrRef composability**: `service_plan_id`, `resource_group`, `storage_account_name` + `storage_account_access_key`, `virtual_network_subnet_id`, and `application_insights_connection_string` all support `valueFrom` references
- **Three storage bindings**: account + access key, account + managed identity (credential-free), or Key Vault secret ID -- with the exactly-one and key-XOR-identity contracts validated at manifest time
- **Serverless scaling dials**: `app_scale_limit` (the Consumption cost cap), `elastic_instance_minimum` + `pre_warmed_instance_count` (Elastic Premium cold-start elimination), `runtime_scale_monitoring_enabled` (KEDA-based trigger scaling), and `daily_memory_time_quota` (the Consumption cost circuit breaker)
- **Built-in authentication (Easy Auth v2)**: Platform-level authentication against Microsoft Entra ID, Apple, Facebook, GitHub, Google, Microsoft account, Twitter, or any OpenID Connect provider -- validated before requests reach function code, with provider secrets referenced by app setting name
- **Scheduled backups**: Content + configuration backups to a storage container (Standard tier and above)
- **Slot-swap safety**: `sticky_settings` pins named app settings and connection strings to the production slot
- **Docker support**: Custom container images, over the public internet or the VNet (`vnet_image_pull_enabled`), with managed-identity ACR pulls
- **Managed identity**: SYSTEM_ASSIGNED, USER_ASSIGNED, or both
- **Deployment hardening**: Basic-auth publishing toggles (Web Deploy + FTP) close the classic credential-based deployment paths
- **IP restrictions**: IP-based, service-tag, and VNet-based access control for both the main site and SCM (Kudu), with allow/deny default actions
- **User tags**: Free-form Azure resource tags merged over the platform's metadata-derived tags

## When to Use

- **HTTP APIs**: Lightweight REST endpoints with pay-per-execution billing
- **Event processing**: Queue, blob, Service Bus, and Event Hub triggered workloads
- **Scheduled jobs**: Timer-triggered functions (cron-style)
- **Containerized functions**: Custom Docker images with the Functions runtime
- **Infra Charts**: Leaf resource in the `function-app-environment` Infra Chart (references ServicePlan, StorageAccount, AppInsights)

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFunctionApp
metadata:
  name: my-fn
spec:
  region: eastus
  resourceGroup:
    value: my-rg
  functionAppName: my-fn-app
  servicePlanId:
    value: /subscriptions/.../Microsoft.Web/serverfarms/my-plan
  storageAccountName:
    value: myfnstorage
  storageAccountAccessKey:
    value: "<storage-key>"
  siteConfig:
    applicationStack:
      pythonVersion: "3.12"
  httpsOnly: true
```

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region (ForceNew) |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (ForceNew) |
| `function_app_name` | string | Yes | - | Globally unique name (`{function_app_name}.azurewebsites.net`) (ForceNew) |
| `service_plan_id` | StringValueOrRef | Yes | - | App Service Plan ARM ID (AzureServicePlan ref) |
| `storage_account_name` | StringValueOrRef | One-of | - | Runtime-state storage account (XOR `storage_key_vault_secret_id`) |
| `storage_account_access_key` | StringValueOrRef | No | - | Storage access key (XOR `storage_uses_managed_identity`) |
| `storage_uses_managed_identity` | bool | No | `false` | Credential-free storage access via the app's identity |
| `storage_key_vault_secret_id` | string | One-of | - | Key Vault secret holding the storage connection string |
| `functions_extension_version` | string | No | `"~4"` | Functions host version |
| `daily_memory_time_quota` | int32 | No | `0` | Consumption cost circuit breaker (GB-seconds/day; 0 = unlimited) |
| `site_config` | SiteConfig | Yes | - | Runtime, scaling, security, and operational configuration |
| `app_settings` | map | No | - | Environment variables |
| `connection_strings` | repeated | No | - | Named connection strings (name, enum type, value) |
| `sticky_settings` | StickySettings | No | - | Settings pinned to the production slot during swaps |
| `application_insights_connection_string` | StringValueOrRef | No | - | APM telemetry destination |
| `https_only` | bool | No | `true` | Enforce HTTPS-only access |
| `public_network_access_enabled` | bool | No | `true` | Enable public network access |
| `enabled` | bool | No | `true` | Enable or disable the app |
| `builtin_logging_enabled` | bool | No | `true` | Legacy AzureWebJobsDashboard logging |
| `content_share_force_disabled` | bool | No | `false` | Skip the auto-created content share |
| `client_certificate_enabled` / `_mode` / `_exclusion_paths` | - | No | `OPTIONAL` | Mutual TLS posture |
| `virtual_network_subnet_id` | StringValueOrRef | No | - | Subnet for VNet integration (AzureSubnet ref) |
| `vnet_image_pull_enabled` | bool | No | `false` | Pull container images over the VNet |
| `virtual_network_backup_restore_enabled` | bool | No | `false` | Route backup traffic over the VNet |
| `identity` | Identity | No | - | Managed identity (SYSTEM_ASSIGNED, USER_ASSIGNED, or both) |
| `key_vault_reference_identity_id` | StringValueOrRef | No | - | Identity for Key Vault references |
| `webdeploy_publish_basic_authentication_enabled` | bool | No | `true` | Basic-auth publishing over Web Deploy |
| `ftp_publish_basic_authentication_enabled` | bool | No | `true` | Basic-auth publishing over FTP/FTPS |
| `zip_deploy_file` | string | No | - | Local ZIP package deployed on create/update |
| `storage_mounts` | repeated | No | - | Azure File Share or Blob container mounts |
| `backup` | Backup | No | - | Scheduled backups (Standard+) |
| `auth_settings_v2` | AuthSettingsV2 | No | - | Built-in authentication (Easy Auth v2) |
| `tags` | map | No | - | Free-form Azure resource tags |

## Outputs

| Output | Description |
|--------|-------------|
| `function_app_id` | ARM resource ID of the Function App |
| `default_hostname` | Default hostname (`{function_app_name}.azurewebsites.net`) |
| `outbound_ip_addresses` | Currently active outbound IP addresses |
| `possible_outbound_ip_addresses` | Every outbound IP the platform could ever use -- for durable firewall allowlists |
| `identity_principal_id` | System-assigned identity principal ID (for RBAC) |
| `identity_tenant_id` | System-assigned identity tenant ID |
| `custom_domain_verification_id` | TXT record value for custom domain verification |
| `kind` | Resource kind string (e.g., `"functionapp,linux"`) |
| `hosting_environment_id` | ARM ID of the App Service Environment (empty outside ASE) |
| `site_credential_name` | Publishing credential username (Kudu/SCM basic auth) |
| `site_credential_password` | Publishing credential password -- secret-bearing; grants deploy access while basic-auth publishing is enabled |

## Downstream Usage

AzureFunctionApp is a **leaf resource**. It consumes outputs from upstream resources via `valueFrom`:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFunctionApp
metadata:
  name: my-events-fn
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: shared-rg
      fieldPath: status.outputs.resource_group_name
  functionAppName: my-events-fn
  servicePlanId:
    valueFrom:
      kind: AzureServicePlan
      name: fn-plan
      fieldPath: status.outputs.service_plan_id
  storageAccountName:
    valueFrom:
      kind: AzureStorageAccount
      name: fn-storage
      fieldPath: status.outputs.storage_account_name
  storageAccountAccessKey:
    valueFrom:
      kind: AzureStorageAccount
      name: fn-storage
      fieldPath: status.outputs.primary_access_key
  siteConfig:
    applicationStack:
      pythonVersion: "3.12"
```

## Deliberate Scope Boundaries

- **Flex Consumption function apps** are Azure's own separate resource type (`azurerm_function_app_flex_consumption`) with a container-endpoint storage model and top-level runtime selection -- a standalone-kind candidate, not fields on this spec.
- **Legacy `auth_settings` (v1)**: superseded by `auth_settings_v2`, which this kind models fully.
- **Deployment slots** mirror the entire app surface with an independent lifecycle -- a standalone-kind candidate. `sticky_settings` is modeled here because it lives on the production app.
- **Windows Function Apps** are the legacy path; the platform targets Linux-first runtimes and containers.

## Further Reading


---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

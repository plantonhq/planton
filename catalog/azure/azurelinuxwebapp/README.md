# AzureLinuxWebApp

An Azure Linux Web App manages web hosting infrastructure on Azure App Service (Linux), providing a fully managed platform for running web applications, APIs, containerized services, and microservices.

## Overview

The `AzureLinuxWebApp` component provisions an `azurerm_linux_web_app` resource, a managed web hosting platform that runs long-running HTTP workloads on Azure App Service. It hosts always-on web applications serving HTTP traffic, unlike serverless Function Apps that are event-driven and scale to zero.

Every Web App requires:
- **An App Service Plan** (`AzureServicePlan`): Determines cost model, compute tier, and available features (Free through Premium v4)
- **An application stack**: The runtime (.NET, Node.js, Python, PHP, Go, Java with Tomcat/JBoss, or Docker container)

## Key Features

- **Dual IaC support**: Both Pulumi and Terraform modules with feature parity
- **StringValueOrRef composability**: `service_plan_id`, `resource_group`, `virtual_network_subnet_id`, and `application_insights_connection_string` all support `valueFrom` references
- **Broad runtime support**: .NET, Node.js, Python, PHP, Go, Java (SE/Tomcat/JBoss EAP), and Docker containers -- more runtimes than Function Apps. Ruby is deliberately not modeled: Azure retired the Ruby runtime on App Service; run Ruby apps as Docker containers
- **Full-feature site_config**: Application stack selection, health checks, TLS floors + minimum cipher suite, FTPS state, load balancing, request pipeline mode, CORS, IP restrictions, HTTP/2, WebSockets, API Management wiring, default documents, remote debugging
- **Auto-heal**: Recycle the app automatically on request-volume, status-code, or slow-request triggers, with a minimum-uptime guard against recycle loops
- **Built-in authentication (Easy Auth v2)**: Platform-level authentication against Microsoft Entra ID, Apple, Facebook, GitHub, Google, Microsoft account, Twitter, or any OpenID Connect provider -- validated before requests reach application code, with provider secrets referenced by app setting name
- **Scheduled backups**: Content + configuration backups to a storage container with cadence and retention controls (Standard tier and above)
- **Slot-swap safety**: `sticky_settings` pins named app settings and connection strings to the production slot during swaps
- **Docker support**: Run custom container images as Web Apps via the `docker` application stack, over the public internet or the VNet (`vnet_image_pull_enabled`)
- **Managed identity**: SYSTEM_ASSIGNED, USER_ASSIGNED, or both -- credential-free access to Azure services
- **Deployment hardening**: Basic-auth publishing toggles (Web Deploy + FTP) close the classic credential-based deployment paths
- **Connection strings**: Named, typed connection strings for database and service integrations
- **IP restrictions**: IP-based, service-tag, and VNet-based access control for both the main site and SCM (Kudu), with allow/deny default actions
- **Logging**: Application logs (file system or blob storage), HTTP logs (file system or blob storage), failed request tracing, and detailed error messages
- **Storage mounts**: Mount Azure File Shares or Blob containers as directories accessible at runtime
- **User tags**: Free-form Azure resource tags merged over the platform's metadata-derived tags

## When to Use

- **Web APIs**: REST/GraphQL APIs behind a load balancer with health checks (Node.js, Python Flask/FastAPI, .NET Web API)
- **Web applications**: Server-rendered web apps (Next.js, Django, ASP.NET MVC, Spring Boot)
- **Containerized services**: Custom Docker containers with any runtime or framework
- **Microservices**: Individual services in a microservices architecture, each with independent scaling
- **Infra charts**: Leaf resource in the `web-app-environment` infra chart (references ServicePlan, AppInsights, Subnet)

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLinuxWebApp
metadata:
  name: my-web-app
spec:
  region: eastus
  resourceGroup:
    value: my-rg
  webAppName: my-web-app
  servicePlanId:
    value: /subscriptions/.../Microsoft.Web/serverfarms/my-plan
  siteConfig:
    applicationStack:
      pythonVersion: "3.12"
    healthCheckPath: /health
  httpsOnly: true
```

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region (ForceNew) |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or AzureResourceGroup ref) (ForceNew) |
| `web_app_name` | string | Yes | - | Globally unique name (`{web_app_name}.azurewebsites.net`) (ForceNew) |
| `service_plan_id` | StringValueOrRef | Yes | - | App Service Plan ARM ID (AzureServicePlan ref) |
| `site_config` | SiteConfig | Yes | - | Site configuration (runtime, scaling, security, auto-heal) |
| `app_settings` | map | No | - | Environment variables (key-value pairs) |
| `connection_strings` | repeated | No | - | Named connection strings (name, enum type, value) |
| `sticky_settings` | StickySettings | No | - | Settings pinned to the production slot during swaps |
| `application_insights_connection_string` | StringValueOrRef | No | - | Application Insights connection string |
| `https_only` | bool | No | `true` | Enforce HTTPS-only access |
| `public_network_access_enabled` | bool | No | `true` | Enable public network access |
| `enabled` | bool | No | `true` | Enable or disable the Web App |
| `client_affinity_enabled` | bool | No | `false` | ARR session affinity (stateful apps) |
| `client_certificate_enabled` | bool | No | `false` | Enable mTLS client certificates |
| `client_certificate_mode` | enum | No | `OPTIONAL` | `REQUIRED`, `OPTIONAL`, or `OPTIONAL_INTERACTIVE_USER` |
| `client_certificate_exclusion_paths` | string | No | - | Semicolon-separated paths excluded from cert validation |
| `virtual_network_subnet_id` | StringValueOrRef | No | - | Subnet for VNet integration (AzureSubnet ref) |
| `vnet_image_pull_enabled` | bool | No | `false` | Pull container images over the VNet |
| `virtual_network_backup_restore_enabled` | bool | No | `false` | Route backup/restore traffic over the VNet |
| `identity` | Identity | No | - | Managed identity (SYSTEM_ASSIGNED, USER_ASSIGNED, or both) |
| `key_vault_reference_identity_id` | StringValueOrRef | No | - | Identity for Key Vault references |
| `webdeploy_publish_basic_authentication_enabled` | bool | No | `true` | Basic-auth publishing over Web Deploy |
| `ftp_publish_basic_authentication_enabled` | bool | No | `true` | Basic-auth publishing over FTP/FTPS |
| `zip_deploy_file` | string | No | - | Local ZIP package to deploy on create/update |
| `storage_mounts` | repeated | No | - | Azure File Share or Blob container mounts |
| `logs` | Logs | No | - | Application and HTTP logging (file system or blob storage) |
| `backup` | Backup | No | - | Scheduled backups to a storage container (Standard+) |
| `auth_settings_v2` | AuthSettingsV2 | No | - | Built-in authentication (Easy Auth v2) |
| `tags` | map | No | - | Free-form Azure resource tags |

## Outputs

| Output | Description |
|--------|-------------|
| `web_app_id` | ARM resource ID of the Web App |
| `default_hostname` | Default hostname (`{web_app_name}.azurewebsites.net`) |
| `outbound_ip_addresses` | Currently active outbound IP addresses |
| `possible_outbound_ip_addresses` | Every outbound IP the platform could ever use -- for durable firewall allowlists |
| `identity_principal_id` | System-assigned identity principal ID (for RBAC) |
| `identity_tenant_id` | System-assigned identity tenant ID |
| `custom_domain_verification_id` | TXT record value for custom domain verification |
| `kind` | Resource kind string (e.g., `"app,linux"`) |
| `hosting_environment_id` | ARM ID of the App Service Environment (empty outside ASE) |
| `site_credential_name` | Publishing credential username (Kudu/SCM basic auth) |
| `site_credential_password` | Publishing credential password -- secret-bearing; grants deploy access while basic-auth publishing is enabled |

## Downstream Usage

AzureLinuxWebApp is a **leaf resource** -- nothing references its outputs downstream. It consumes outputs from upstream resources via `valueFrom`:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLinuxWebApp
metadata:
  name: my-api
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: shared-rg
      fieldPath: status.outputs.resource_group_name
  webAppName: my-api
  servicePlanId:
    valueFrom:
      kind: AzureServicePlan
      name: web-plan
      fieldPath: status.outputs.service_plan_id
  applicationInsightsConnectionString:
    valueFrom:
      kind: AzureApplicationInsights
      name: web-insights
      fieldPath: status.outputs.connection_string
  virtualNetworkSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: web-subnet
      fieldPath: status.outputs.subnet_id
  siteConfig:
    applicationStack:
      pythonVersion: "3.12"
```

## Deliberate Scope Boundaries

- **Legacy `auth_settings` (v1)**: superseded by `auth_settings_v2`, which this kind models fully. Azure keeps v1 for backward compatibility only; new configuration should always use v2.
- **Deployment slots**: a slot mirrors the entire app surface with an independent lifecycle -- a genuine standalone-kind candidate, not a field on this spec. `sticky_settings` (which governs swap behavior) is modeled here because it lives on the production app.
- **Custom domain bindings and certificates**: separate Azure resources with their own lifecycles (hostname binding, managed certificate, certificate binding); they compose with the app rather than embed in it.
- **Windows Web Apps**: `azurerm_windows_web_app` is a separate resource for the legacy .NET Framework path; the platform targets Linux-first runtimes and containers. `AzureServicePlan` supports Windows plans, so the compute tier is ready if a Windows app kind is ever added.

## Further Reading


---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# AzureLinuxWebApp

An Azure Linux Web App manages web hosting infrastructure on Azure App Service (Linux), providing a fully managed platform for running web applications, APIs, containerized services, and microservices.

## Overview

The `AzureLinuxWebApp` component provisions an `azurerm_linux_web_app` resource, a managed web hosting platform that runs long-running HTTP workloads on Azure App Service. It is the Azure equivalent of AWS Elastic Beanstalk or GCP App Engine -- it hosts always-on web applications serving HTTP traffic, unlike serverless Function Apps that are event-driven and scale to zero.

Every Web App requires:
- **An App Service Plan** (`AzureServicePlan`): Determines cost model, compute tier, and available features (Free through Premium v3)
- **An application stack**: The runtime (.NET, Node.js, Python, PHP, Ruby, Go, Java with Tomcat/JBoss, or Docker container)

## Key Features

- **Dual IaC support**: Both Pulumi and Terraform modules with feature parity
- **StringValueOrRef composability**: `service_plan_id`, `resource_group`, `virtual_network_subnet_id`, and `application_insights_connection_string` all support `valueFrom` references
- **Broad runtime support**: .NET, Node.js, Python, PHP, Ruby, Go, Java (SE/Tomcat/JBoss EAP), and Docker containers -- more runtimes than Function Apps
- **Full-feature site_config**: Application stack selection, health checks, TLS settings, FTPS state, load balancing, CORS, IP restrictions, HTTP/2, WebSockets
- **Docker support**: Run custom container images as Web Apps via the `docker` application stack
- **Managed identity**: SystemAssigned, UserAssigned, or both -- credential-free access to Azure services
- **Connection strings**: Named, typed connection strings for database and service integrations
- **IP restrictions**: IP-based, service-tag, and VNet-based access control for both the main site and SCM (Kudu)
- **CORS**: Cross-origin resource sharing configuration for HTTP endpoints
- **Logging**: Application logs, HTTP logs, failed request tracing, and detailed error messages
- **Storage mounts**: Mount Azure File Shares or Blob containers as directories accessible at runtime

## When to Use

- **Web APIs**: REST/GraphQL APIs behind a load balancer with health checks (Node.js, Python Flask/FastAPI, .NET Web API)
- **Web applications**: Server-rendered web apps (Next.js, Django, ASP.NET MVC, Spring Boot)
- **Containerized services**: Custom Docker containers with any runtime or framework
- **Microservices**: Individual services in a microservices architecture, each with independent scaling
- **Infra charts**: Leaf resource in the `web-app-environment` infra chart (references ServicePlan, AppInsights, Subnet)

## Quick Example

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: my-web-app
spec:
  region: eastus
  resource_group: my-rg
  name: my-web-app
  service_plan_id: /subscriptions/.../Microsoft.Web/serverfarms/my-plan
  site_config:
    application_stack:
      python_version: "3.12"
    health_check_path: /health
  https_only: true
```

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region (ForceNew) |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or AzureResourceGroup ref) (ForceNew) |
| `name` | string | Yes | - | Globally unique name (`{name}.azurewebsites.net`) (ForceNew) |
| `service_plan_id` | StringValueOrRef | Yes | - | App Service Plan ARM ID (AzureServicePlan ref) |
| `site_config` | SiteConfig | Yes | - | Site configuration (runtime, scaling, security) |
| `app_settings` | map | No | - | Environment variables (key-value pairs) |
| `connection_strings` | repeated | No | - | Named connection strings (name, type, value) |
| `application_insights_connection_string` | StringValueOrRef | No | - | Application Insights connection string |
| `https_only` | bool | No | `true` | Enforce HTTPS-only access |
| `public_network_access_enabled` | bool | No | `true` | Enable public network access |
| `enabled` | bool | No | `true` | Enable or disable the Web App |
| `virtual_network_subnet_id` | StringValueOrRef | No | - | Subnet for VNet integration (AzureSubnet ref) |
| `identity` | Identity | No | - | Managed identity (SystemAssigned, UserAssigned, or both) |
| `key_vault_reference_identity_id` | StringValueOrRef | No | - | Identity for Key Vault references |
| `client_affinity_enabled` | bool | No | `false` | ARR session affinity (stateful apps) |
| `client_certificate_enabled` | bool | No | `false` | Enable mTLS client certificates |
| `client_certificate_mode` | string | No | `"Optional"` | Certificate mode (Required, Optional, OptionalInteractiveUser) |
| `client_certificate_exclusion_paths` | string | No | - | Semicolon-separated paths excluded from cert validation |
| `storage_mounts` | repeated | No | - | Azure File Share or Blob container mounts |
| `logs` | Logs | No | - | Application and HTTP logging configuration |

## Outputs

| Output | Description |
|--------|-------------|
| `web_app_id` | ARM resource ID of the Web App |
| `default_hostname` | Default hostname (`{name}.azurewebsites.net`) |
| `outbound_ip_addresses` | Outbound IP addresses (for downstream firewall rules) |
| `identity_principal_id` | System-assigned identity principal ID (for RBAC) |
| `identity_tenant_id` | System-assigned identity tenant ID |
| `custom_domain_verification_id` | TXT record value for custom domain verification |
| `kind` | Resource kind string (e.g., `"app,linux"`) |

## Downstream Usage

AzureLinuxWebApp is a **leaf resource** -- nothing references its outputs downstream. It consumes outputs from upstream resources via `valueFrom`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLinuxWebApp
metadata:
  name: my-api
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: shared-rg
      fieldPath: status.outputs.resource_group_name
  name: my-api
  service_plan_id:
    valueFrom:
      kind: AzureServicePlan
      name: web-plan
      fieldPath: status.outputs.plan_id
  application_insights_connection_string:
    valueFrom:
      kind: AzureApplicationInsights
      name: web-insights
      fieldPath: status.outputs.connection_string
  virtual_network_subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: web-subnet
      fieldPath: status.outputs.subnet_id
  site_config:
    application_stack:
      python_version: "3.12"
```

## What's NOT Included (80/20 Scope)

- **auth_settings / auth_settings_v2**: Azure App Service Authentication (Easy Auth). Complex configuration surface with 20+ sub-fields. Deferred to v2 when demand materializes.
- **backup**: Automated backup configuration. Niche feature; most web apps are stateless with CI/CD deployment.
- **sticky_settings**: App settings that don't swap during slot deployments. Requires deployment slots (not in v1).
- **Deployment slots**: Blue-green deployment via staging slots. Significant complexity; deferred to v2.
- **zip_deploy_file**: In-line ZIP deployment. Most teams use CI/CD pipelines for deployment, not in-line ZIP.
- **Windows Web Apps**: `azurerm_windows_web_app` is excluded. Linux covers the vast majority of web workloads, and the API surface is kept clean with a single component.
- **Custom domain bindings**: Custom domain binding is a separate resource (`AzureAppServiceCustomHostnameBinding`), not embedded in the Web App spec.

These omissions follow the 80/20 principle: the included fields cover the vast majority of production use cases while keeping the API surface clean and maintainable.

## Further Reading

- [examples.md](./examples.md) -- Complete YAML manifest examples for common scenarios
- [docs/README.md](./docs/README.md) -- Comprehensive research and design documentation
- [iac/pulumi/overview.md](./iac/pulumi/overview.md) -- Pulumi module architecture overview

# AzureContainerAppEnvironment

An Azure Container Apps Managed Environment is the hosting platform for Azure Container Apps -- a secure boundary that provides networking, logging, and compute isolation for one or more containerized workloads.

## Overview

The `AzureContainerAppEnvironment` component provisions an `azurerm_container_app_environment` resource, creating the execution boundary where Azure Container Apps run. It is the **foundation resource** for the `container-apps-environment` infra chart.

Think of it as the Azure equivalent of an ECS Cluster or a Kubernetes namespace: all apps within an environment share the same virtual network, logging configuration, and Dapr infrastructure.

A Container App Environment determines:
- **Networking mode**: External (public) or internal (VNet-only)
- **VNet injection**: Optional placement inside a customer-managed VNet
- **Logging**: Optional Log Analytics Workspace integration
- **Compute**: Serverless Consumption (default) plus optional dedicated workload profiles
- **Zone redundancy**: Cross-zone distribution for production resilience

## Key Features

- **Dual IaC support**: Both Pulumi and Terraform modules with feature parity
- **StringValueOrRef composability**: `resource_group`, `infrastructure_subnet_id`, and `log_analytics_workspace_id` all support `valueFrom` for infra-chart wiring
- **Auto-derived logging**: When `log_analytics_workspace_id` is provided, `logs_destination` is automatically set to `"log-analytics"` -- no extra field for the user
- **Workload profiles**: Optional dedicated compute (D4, D8, E4, E8, GPU) alongside always-available Consumption
- **Zone redundancy**: Single boolean for cross-zone HA (requires VNet injection)
- **Internal mode**: Single boolean for VNet-only access (no public internet)

## When to Use

- **Microservice platforms**: Host multiple container apps in a shared networking and logging boundary
- **VNet-integrated workloads**: Connect container apps to private databases, storage, and other VNet resources
- **Serverless containers**: Run containers without managing infrastructure (Consumption plan)
- **GPU workloads**: Provision dedicated GPU profiles (NC24-A100, NC48-A100, NC96-A100)
- **Infra charts**: Foundation resource in the `container-apps-environment` infra chart

## Networking Modes

| Mode | Subnet Required | Public Access | Use Case |
|------|----------------|---------------|----------|
| External (default) | No | Yes | Public APIs, web frontends |
| VNet-injected | Yes (/21+) | Yes | Apps needing private backend access |
| Internal | Yes (/21+) | No | Backend services, internal APIs |

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or AzureResourceGroup ref) |
| `name` | string | Yes | - | Environment name (lowercase, hyphens, 2-60 chars) |
| `infrastructure_subnet_id` | StringValueOrRef | No | - | Subnet for VNet injection (must be /21+) |
| `log_analytics_workspace_id` | StringValueOrRef | No | - | LAW for centralized logging |
| `internal_load_balancer_enabled` | bool | No | `false` | Internal-only access (requires subnet) |
| `zone_redundancy_enabled` | bool | No | `false` | Cross-zone HA (requires subnet) |
| `workload_profiles` | repeated | No | - | Dedicated compute profiles |

## Outputs

| Output | Description |
|--------|-------------|
| `environment_id` | ARM resource ID (referenced by AzureContainerApp) |
| `default_domain` | Default domain for apps ({app}.{domain}) |
| `static_ip_address` | Static IP (public or private depending on mode) |
| `platform_reserved_cidr` | Reserved IP range for platform infrastructure |
| `platform_reserved_dns_ip_address` | Internal DNS server IP |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureContainerAppEnvironment
metadata:
  name: my-apps-env
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: my-apps-env
  infrastructure_subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: apps-subnet
      fieldPath: status.outputs.subnet_id
  log_analytics_workspace_id:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: central-law
      fieldPath: status.outputs.workspace_id
```

## Downstream Usage

The `environment_id` output is referenced by Container Apps:

```yaml
# AzureContainerApp referencing this environment
apiVersion: azure.planton.dev/v1
kind: AzureContainerApp
metadata:
  name: my-api
spec:
  container_app_environment_id:
    valueFrom:
      kind: AzureContainerAppEnvironment
      name: my-apps-env
      fieldPath: status.outputs.environment_id
  name: my-api
  # ...
```

## What's NOT Included (80/20 Scope)

- **`mutual_tls_enabled`**: Public Preview feature that may increase latency and reduce throughput. Enterprise niche.
- **`public_network_access`**: Computed by Azure based on VNet configuration. Redundant with `internal_load_balancer_enabled`.
- **`infrastructure_resource_group_name`**: Azure auto-generates the platform resource group. Niche customization.
- **`dapr_application_insights_connection_string`**: Dapr-specific telemetry. Very niche, ForceNew, and write-only.
- **`identity`**: Environment-level managed identity is uncommon. Container Apps themselves use identity.
- **Storage, Dapr components, certificates, custom domains**: These are separate Terraform resources. Would be separate Planton kinds if needed.

These omissions follow the 80/20 principle: the included fields cover the vast majority of production use cases while keeping the API surface clean and maintainable.

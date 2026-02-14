---
title: "AzureContainerAppEnvironment: Research & Design Documentation"
description: "AzureContainerAppEnvironment: Research & Design Documentation deployment documentation"
icon: "package"
order: 100
componentName: "azurecontainerappenvironment"
---

# AzureContainerAppEnvironment: Research & Design Documentation

## Overview

Azure Container Apps is Microsoft's fully managed serverless container platform, built on Kubernetes and open-source technologies (Envoy, KEDA, Dapr). A **Managed Environment** (`Microsoft.App/managedEnvironments`) is the logical boundary that hosts one or more Container Apps, providing shared networking, logging, and compute configuration.

This document captures the research, design decisions, and 80/20 scoping rationale for the OpenMCF `AzureContainerAppEnvironment` component.

## Azure Deployment Landscape

### What is a Container App Environment?

A Container App Environment is:
- A secure boundary around groups of container apps
- The shared networking context (VNet, internal/external load balancing)
- The logging configuration (Log Analytics, Azure Monitor, or streaming)
- The compute capacity definition (Consumption and/or dedicated workload profiles)
- The Dapr infrastructure context (sidecar mesh)

All apps within an environment share these resources. Apps in different environments are fully isolated.

### Comparison with Other Azure Compute

| Feature | Container Apps | AKS | App Service | Functions |
|---------|---------------|-----|-------------|-----------|
| Managed K8s | Yes (hidden) | Yes (visible) | No | No |
| Scale to zero | Yes | No | No (except Y1) | Yes (Y1) |
| GPU support | Yes (dedicated) | Yes | No | No |
| VNet injection | Optional | Required | Optional | Optional |
| Dapr built-in | Yes | Manual | No | No |
| Custom domains | Yes | Via Ingress | Yes | Yes |

### When Container Apps > AKS

- Microservice teams that don't need K8s API access
- Event-driven workloads that benefit from KEDA auto-scaling
- Applications needing Dapr for state management, pub/sub, service invocation
- Teams wanting serverless containers without cluster management overhead

### When AKS > Container Apps

- Workloads needing direct Kubernetes API access
- Complex networking requirements (CNI, network policies)
- Helm chart-based deployments
- Stateful workloads needing persistent volumes

## Method Comparison

### Terraform: `azurerm_container_app_environment`

The Terraform azurerm provider (v4.x, API v2025-07-01) exposes the full surface area:

**All fields**: name, resource_group_name, location, infrastructure_subnet_id, log_analytics_workspace_id, logs_destination, internal_load_balancer_enabled, zone_redundancy_enabled, workload_profile, mutual_tls_enabled, public_network_access, infrastructure_resource_group_name, dapr_application_insights_connection_string, identity, tags.

**Computed outputs**: custom_domain_verification_id, default_domain, docker_bridge_cidr, platform_reserved_cidr, platform_reserved_dns_ip_address, static_ip_address.

### Pulumi: `containerapp.Environment`

The Pulumi Azure classic provider (v6) mirrors the Terraform schema 1:1 since it's generated from the same provider.

### Azure CLI: `az containerapp env create`

```bash
az containerapp env create \
  --name my-env \
  --resource-group my-rg \
  --location eastus \
  --infrastructure-subnet-resource-id $SUBNET_ID \
  --logs-workspace-id $WORKSPACE_ID \
  --internal-only true \
  --zone-redundant true
```

## 80/20 Scoping Decisions

### Included (covers 95%+ of production use cases)

| Field | Rationale |
|-------|-----------|
| `region` | Required for all Azure resources |
| `resource_group` | Per DD05, StringValueOrRef for composability |
| `name` | Required identifier |
| `infrastructure_subnet_id` | VNet injection is the primary production pattern |
| `log_analytics_workspace_id` | Centralized logging is essential for production |
| `internal_load_balancer_enabled` | Common pattern for backend services |
| `zone_redundancy_enabled` | Production HA requirement |
| `workload_profiles` | Dedicated compute for workloads needing guaranteed resources |

### Excluded (with justification)

| Feature | Rationale |
|---------|-----------|
| `mutual_tls_enabled` | Public Preview; may increase latency and reduce throughput in high-load scenarios. Enterprise security niche. |
| `public_network_access` | Computed by Azure based on VNet config. Cannot be "Enabled" when internal LB is on. Redundant control surface. |
| `infrastructure_resource_group_name` | Azure auto-generates. Only valid with workload profiles. Niche customization that adds confusion. |
| `dapr_application_insights_connection_string` | Dapr-specific telemetry. ForceNew + write-only (not returned by API). Very niche. |
| `identity` | Environment-level identity is uncommon. Container Apps themselves use identity via the AzureContainerApp spec. |
| `logs_destination` | Auto-derived in IaC modules from `log_analytics_workspace_id` presence. Eliminates a redundant user-facing field. |
| Storage | Separate TF resource (`azurerm_container_app_environment_storage`). Different lifecycle. |
| Dapr components | Separate TF resource (`azurerm_container_app_environment_dapr_component`). Different lifecycle. |
| Certificates | Separate TF resource (`azurerm_container_app_environment_certificate`). Different lifecycle. |
| Custom domains | Separate TF resource (`azurerm_container_app_environment_custom_domain`). Different lifecycle. |

### logs_destination Auto-Derivation

The Terraform provider has an explicit `logs_destination` field with values `"log-analytics"`, `"azure-monitor"`, or empty (streaming only).

We chose to auto-derive this in IaC modules rather than expose it as a user-facing field:
- If `log_analytics_workspace_id` is provided → `logs_destination = "log-analytics"`
- If not provided → `logs_destination = null` (streaming only)

This covers 95%+ of use cases. The `"azure-monitor"` destination is a separate integration path that can be added in v2 if demand warrants it.

### Workload Profile Consumption Auto-Addition

Azure's API always returns the "Consumption" workload profile, even if not specified. The `workload_profiles` field in our spec is exclusively for **dedicated compute profiles**. Users never need to add Consumption -- it's always available.

This design simplifies the user experience: an empty `workload_profiles` list means Consumption-only, and any entries are additional dedicated capacity.

## Best Practices for Production

1. **Always use VNet injection** for production environments (set `infrastructure_subnet_id`)
2. **Always enable Log Analytics** (set `log_analytics_workspace_id`)
3. **Enable zone redundancy** for production workloads (set `zone_redundancy_enabled: true`)
4. **Subnet sizing**: Use /21 or larger. Container Apps reserves significant IP space for platform infrastructure.
5. **Internal mode** for backend services that should never be publicly accessible
6. **Workload profiles** for workloads needing guaranteed resources or GPU access

## Infra Chart Integration

The `AzureContainerAppEnvironment` is a Layer 1 resource in the `container-apps-environment` infra chart:

```
Layer 0: AzureResourceGroup
Layer 0: AzureVpc → AzureSubnet
Layer 0: AzureLogAnalyticsWorkspace
Layer 1: AzureContainerAppEnvironment (this resource)
Layer 2: AzureContainerApp (one or more)
```

All Layer 0 dependencies are connected via `StringValueOrRef` fields, enabling the infra chart DAG to resolve deployment order automatically.

## Related Resources

- **AzureContainerApp** (R18): Individual container app workloads deployed into this environment
- **AzureSubnet**: Required for VNet-injected environments (/21 subnet)
- **AzureLogAnalyticsWorkspace**: Recommended for centralized log collection
- **AzureResourceGroup**: Required container for all Azure resources

---

**Status**: Production Ready
**API Version**: azure.openmcf.org/v1
**Terraform Resource**: `azurerm_container_app_environment`
**Pulumi Resource**: `containerapp.Environment`

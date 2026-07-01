# Azure Container Apps Environment

Serverless container platform on Azure with VNet integration and centralized monitoring.

## What This Chart Deploys

| Resource | Kind | Condition |
|----------|------|-----------|
| Resource Group | `AzureResourceGroup` | Always |
| Virtual Network | `AzureVpc` | Always |
| Container Apps Subnet | `AzureSubnet` | Always |
| Log Analytics Workspace | `AzureLogAnalyticsWorkspace` | Always |
| Application Insights | `AzureApplicationInsights` | `create_app_insights` |
| Container App Environment | `AzureContainerAppEnvironment` | Always |
| Container App | `AzureContainerApp` | `create_container_app` |

## Architecture

The chart creates a VNet-injected Container App Environment. This means all
container apps run inside a customer-managed VNet, enabling private connectivity
to databases, storage, and other VNet resources.

The Container Apps subnet requires a minimum /21 CIDR (2048 IPs) for the
Container Apps infrastructure. The default is `10.2.8.0/21`.

Log Analytics Workspace is always created for centralized log collection.
Application Insights is optional (default on) for application-level telemetry.

## Parameters

### Foundation

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Azure region | `eastus` |
| `resource_group_name` | Resource group name suffix | `cae-rg` |
| `vnet_cidr` | VNet address space | `10.2.0.0/16` |
| `default_subnet_cidr` | Default subnet CIDR | `10.2.0.0/24` |
| `container_apps_subnet_cidr` | Container Apps subnet (min /21) | `10.2.8.0/21` |

### Container App Environment

| Parameter | Description | Default |
|-----------|-------------|---------|
| `environment_name` | Environment name | `my-cae` |
| `internal_load_balancer` | Internal-only access | `false` |
| `zone_redundancy` | Enable zone redundancy | `false` |

### Container App (optional)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_container_app` | Deploy a sample app | `true` |
| `container_app_name` | App name | `my-app` |
| `container_image` | Container image | `mcr.microsoft.com/k8se/quickstart:latest` |
| `container_cpu` | CPU cores | `0.25` |
| `container_memory` | Memory | `0.5Gi` |
| `target_port` | Container port | `80` |
| `ingress_external` | Public ingress | `true` |
| `min_replicas` | Min replicas | `0` |
| `max_replicas` | Max replicas | `3` |

### Monitoring

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_app_insights` | Enable Application Insights | `true` |

## Example

Deploy a production Container Apps environment with zone redundancy:

```yaml
params:
  region: westus2
  resource_group_name: myapp-cae-rg
  environment_name: myapp-env
  zone_redundancy: true
  container_app_name: api
  container_image: myregistry.azurecr.io/api:latest
  container_cpu: "0.5"
  container_memory: 1Gi
  target_port: "8080"
  min_replicas: "1"
  max_replicas: "10"
```

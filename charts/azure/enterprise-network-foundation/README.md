# Azure Enterprise Network Foundation

Multi-tier VNet foundation with security, load balancing, and monitoring for
enterprise Azure workloads.

## What This Chart Deploys

| Resource | Kind | Condition |
|----------|------|-----------|
| Resource Group | `AzureResourceGroup` | Always |
| Virtual Network | `AzureVpc` | Always |
| Web Tier Subnet | `AzureSubnet` | Always |
| App Tier Subnet | `AzureSubnet` | Always |
| Data Tier Subnet | `AzureSubnet` | Always |
| Gateway Subnet | `AzureSubnet` | `create_app_gateway` |
| Web Tier NSG | `AzureNetworkSecurityGroup` | Always |
| App Tier NSG | `AzureNetworkSecurityGroup` | Always |
| Data Tier NSG | `AzureNetworkSecurityGroup` | Always |
| Public IP | `AzurePublicIp` | Any of NAT/AppGW/LB enabled |
| NAT Gateway | `AzureNatGateway` | `create_nat_gateway` |
| Application Gateway | `AzureApplicationGateway` | `create_app_gateway` |
| Load Balancer | `AzureLoadBalancer` | `create_load_balancer` |
| Log Analytics Workspace | `AzureLogAnalyticsWorkspace` | Always |
| Key Vault | `AzureKeyVault` | `create_key_vault` |

## Architecture

The chart implements a classic three-tier network architecture:

```
                    Internet
                       |
              ┌────────┴────────┐
              │   Public IP     │
              └────────┬────────┘
                       |
         ┌─────────────┼─────────────┐
         │             │             │
    ┌────┴────┐  ┌─────┴─────┐ ┌────┴────┐
    │   NAT   │  │  App GW   │ │   LB    │
    │ Gateway │  │   (L7)    │ │  (L4)   │
    └─────────┘  └─────┬─────┘ └────┬────┘
                       │             │
    ┌──────────────────┼─────────────┘
    │                  │
    ├── Web Tier ──────┤  (NSG: allow HTTP/S from internet)
    │                  │
    ├── App Tier ──────┤  (NSG: allow from web tier only)
    │                  │
    └── Data Tier ─────┘  (NSG: allow from app tier only,
                                deny internet outbound)
```

### Security Model

- **Web Tier NSG**: Allows HTTP (80) and HTTPS (443) from the internet,
  allows all outbound
- **App Tier NSG**: Allows port 8080 from the web tier CIDR only,
  denies all internet inbound, allows all outbound
- **Data Tier NSG**: Allows SQL (1433), PostgreSQL (5432), and Redis (6380)
  from the app tier CIDR only, denies internet inbound AND outbound

### Service Endpoints

- **Web Subnet**: `Microsoft.Web` (App Service integration)
- **App Subnet**: `Microsoft.Sql`, `Microsoft.Storage`, `Microsoft.KeyVault`
- **Data Subnet**: `Microsoft.Sql`, `Microsoft.Storage` (Private Endpoint ready)

## Default CIDR Allocation

| Subnet | CIDR | Size | Purpose |
|--------|------|------|---------|
| Default (VPC internal) | 10.0.0.0/24 | 254 IPs | Reserved |
| Web Tier | 10.0.1.0/24 | 254 IPs | Public-facing workloads |
| App Tier | 10.0.2.0/24 | 254 IPs | Application logic |
| Data Tier | 10.0.3.0/24 | 254 IPs | Databases, storage |
| Gateway | 10.0.4.0/27 | 30 IPs | Application Gateway |

## Parameters

### Foundation

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Azure region | `eastus` |
| `resource_group_name` | Resource group suffix | `network-rg` |
| `vnet_cidr` | VNet address space | `10.0.0.0/16` |
| `default_subnet_cidr` | Default subnet CIDR | `10.0.0.0/24` |

### Tier Subnets

| Parameter | Description | Default |
|-----------|-------------|---------|
| `web_subnet_cidr` | Web tier CIDR | `10.0.1.0/24` |
| `app_subnet_cidr` | App tier CIDR | `10.0.2.0/24` |
| `data_subnet_cidr` | Data tier CIDR | `10.0.3.0/24` |
| `gateway_subnet_cidr` | Gateway CIDR (min /27) | `10.0.4.0/27` |

### Networking

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_nat_gateway` | Enable NAT Gateway | `true` |
| `create_app_gateway` | Enable Application Gateway (L7) | `false` |
| `app_gateway_sku` | AppGW SKU (Standard_v2/WAF_v2) | `Standard_v2` |
| `app_gateway_capacity` | AppGW instance count | `2` |
| `create_load_balancer` | Enable Load Balancer (L4) | `false` |

### Extras

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_key_vault` | Enable Key Vault | `false` |
| `key_vault_name` | Key Vault name | `my-ent-kv` |

## Example

Deploy a production network with Application Gateway and Key Vault:

```yaml
params:
  region: westeurope
  resource_group_name: prod-network-rg
  vnet_cidr: 10.10.0.0/16
  web_subnet_cidr: 10.10.1.0/24
  app_subnet_cidr: 10.10.2.0/24
  data_subnet_cidr: 10.10.3.0/24
  gateway_subnet_cidr: 10.10.4.0/27
  create_nat_gateway: true
  create_app_gateway: true
  app_gateway_sku: WAF_v2
  app_gateway_capacity: "2"
  create_key_vault: true
  key_vault_name: prod-ent-kv
```

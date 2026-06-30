# AzurePublicIp

## Overview

`AzurePublicIp` provisions an Azure Public IP Address -- a static, internet-routable
IPv4 address that can be attached to load balancers, application gateways, NAT gateways,
and virtual machines. Public IPs are a foundational networking primitive in Azure,
sitting at Layer 0-1 of most architectures.

This component provisions Standard SKU public IPs with static allocation. Azure
retired the Basic SKU on September 30, 2025, so only Standard is supported. Standard
SKU requires static allocation, which means the IP address is assigned at creation
time and persists for the lifetime of the resource.

## Key Features

- **Standard SKU only** -- no deprecated Basic SKU baggage; clean and forward-looking
- **Static allocation** -- IP address is persistent and available immediately
- **DNS integration** -- optional `domain_name_label` creates a stable FQDN at
  `{label}.{region}.cloudapp.azure.com`
- **Zone-redundant** -- supports availability zones for production resilience
- **Idle timeout tuning** -- configurable TCP idle timeout (4-30 minutes) for
  long-lived connections (WebSocket, gRPC, database)
- **Composable outputs** -- exports `public_ip_id` for downstream `StringValueOrRef`
  wiring to load balancers, application gateways, and NAT gateways

## When to Use

- When an Azure resource (load balancer, app gateway, NAT gateway) needs a dedicated
  public IP address
- When you need a stable, persistent IP address for DNS A records
- When building enterprise network foundations with explicit IP addressing
- As part of the `enterprise-network-foundation` infra chart

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | -- | Azure region (e.g., "eastus") |
| `resource_group` | StringValueOrRef | Yes | -- | Resource group name or reference |
| `name` | string | Yes | -- | Public IP name (1-80 characters) |
| `domain_name_label` | string | No | -- | DNS label for FQDN creation |
| `zones` | repeated string | No | -- | Availability zones (e.g., ["1","2","3"]) |
| `idle_timeout_in_minutes` | int32 | No | 4 | TCP idle timeout (4-30 minutes) |

## Outputs

| Output | Description |
|--------|-------------|
| `public_ip_id` | Azure Resource Manager ID (used by downstream resources) |
| `ip_address` | Allocated static IPv4 address |
| `fqdn` | Domain name (if `domain_name_label` set) |
| `public_ip_name` | Name of the resource |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1
kind: AzurePublicIp
metadata:
  name: gateway-pip
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-gateway-pip
  domain_name_label: prod-gateway
  zones:
    - "1"
    - "2"
    - "3"
```

## Downstream Usage

Other Azure resources reference this Public IP via `StringValueOrRef`:

```yaml
# In an AzureLoadBalancer spec:
spec:
  public_ip_id:
    valueFrom:
      kind: AzurePublicIp
      name: gateway-pip
      fieldPath: status.outputs.public_ip_id
```

## What's NOT Included

- **Basic SKU** -- retired by Azure, not supported
- **Dynamic allocation** -- Standard SKU requires Static
- **IPv6** -- niche use case, can be added in a future iteration
- **Global tier** -- for cross-region LB only, niche
- **DDoS protection plan** -- separate governance concern
- **IP tags** -- advanced Azure networking feature, very niche

# AzurePublicIp

## Overview

`AzurePublicIp` provisions an Azure Public IP Address -- a static, internet-routable
address that load balancers, application gateways, NAT gateways, firewalls, and
virtual machines attach for inbound or outbound connectivity. Public IPs are a
foundational networking primitive in Azure, sitting at Layer 0-1 of most architectures.

The public IP is a first-class node in the resource graph: higher-level resources
reference it by ARM ID rather than creating their own, so one address can move
between consumers (e.g. re-pointing a frontend during a blue/green cutover)
without changing what the world has allowlisted.

Allocation is always static and is deliberately not modeled: dynamic allocation
existed only for the Basic SKU, whose creation Azure discontinued in 2025 (fully
retired September 30, 2025), and every current SKU requires static allocation.
The address is assigned at creation and persists for the lifetime of the resource.

## Key Features

- **SKU choice** -- `STANDARD` (Azure's default, the production tier every current
  architecture uses) or `STANDARD_V2` (Azure's next-generation SKU, required to
  attach the address to a StandardV2 NAT gateway)
- **Regional or Global tier** -- `REGIONAL` (default) for virtually everything;
  `GLOBAL` for cross-region load balancer frontends (requires the `STANDARD` SKU)
- **IPv4 and IPv6** -- `ipVersion` selects the address family (IPv4 default)
- **Zone-redundant** -- anchor the address to one or more availability zones;
  multiple zones make it zone-redundant, the production default
- **Prefix allocation** -- draw the address from a reserved `AzurePublicIpPrefix`
  (one contiguous, allowlistable range) instead of Microsoft's general pool
- **DNS integration** -- optional `domainNameLabel` creates a stable FQDN at
  `{label}.{region}.cloudapp.azure.com`, with an optional `domainNameLabelScope`
  reuse policy (hashed labels as a defense against dangling-DNS subdomain takeover)
- **Reverse DNS** -- `reverseFqdn` records the PTR name mail servers and
  forward-confirmed-reverse-DNS checks see
- **DDoS protection stance** -- inherit from the virtual network (default),
  `ENABLED` with a dedicated protection plan, or `DISABLED` to opt out
- **IP tags** -- Azure routing metadata attached to the address itself
  (e.g. `RoutingPreference: Internet` for cold-potato vs hot-potato transit)
- **Edge zones** -- deploy the address into a metro-local Azure Edge Zone
- **Idle timeout tuning** -- configurable TCP idle timeout (4-30 minutes) for
  long-lived connections (WebSocket, gRPC, database)
- **Composable outputs** -- exports `public_ip_id` for downstream `StringValueOrRef`
  wiring to load balancers, application gateways, and NAT gateways

## When to Use

- When an Azure resource (load balancer, app gateway, NAT gateway) needs a dedicated
  public IP address
- When you need a stable, persistent IP address for DNS A records
- When building enterprise network foundations with explicit IP addressing
- As part of the `enterprise-network-foundation` Infra Chart

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | -- | Azure region (e.g., "eastus"); must match the region of the resource it attaches to |
| `resource_group` | StringValueOrRef | Yes | -- | Resource group name or reference to an AzureResourceGroup |
| `name` | string | Yes | -- | Public IP name (1-80 characters), unique within the resource group |
| `sku` | enum | No | `STANDARD` | `STANDARD` or `STANDARD_V2`; fixed at creation |
| `sku_tier` | enum | No | `REGIONAL` | `REGIONAL` or `GLOBAL` (cross-region LB frontends; requires `STANDARD` SKU); fixed at creation |
| `ip_version` | enum | No | `IPV4` | `IPV4` or `IPV6`; fixed at creation |
| `zones` | repeated string | No | -- | Availability zones ("1", "2", "3"); multiple zones = zone-redundant; fixed at creation |
| `public_ip_prefix_id` | StringValueOrRef | No | -- | ARM ID of an AzurePublicIpPrefix to allocate the address from; fixed at creation |
| `domain_name_label` | string | No | -- | DNS label for FQDN creation (3-63 chars, lowercase letters, digits, hyphens) |
| `domain_name_label_scope` | enum | No | -- | Label reuse policy: `TENANT_REUSE`, `SUBSCRIPTION_REUSE`, `RESOURCE_GROUP_REUSE`, or `NO_REUSE`; requires `domain_name_label` |
| `reverse_fqdn` | string | No | -- | Reverse-DNS (PTR) name; the FQDN must already resolve to the address |
| `idle_timeout_in_minutes` | int32 | No | 4 | TCP idle timeout (4-30 minutes) |
| `ip_tags` | map | No | -- | Azure IP tags (routing metadata, e.g. `RoutingPreference`); fixed at creation |
| `ddos_protection_mode` | enum | No | inherit | `ENABLED` (pair with `ddos_protection_plan_id`) or `DISABLED`; default inherits from the virtual network |
| `ddos_protection_plan_id` | string | No | -- | ARM ID of the DDoS protection plan; only valid with `ddos_protection_mode: ENABLED` |
| `edge_zone` | string | No | -- | Azure Edge Zone (e.g. "losangeles"); unset deploys into the standard region |
| `tags` | map | No | -- | Free-form tags merged over the Planton-derived resource tags (user tag wins) |

## Outputs

| Output | Description |
|--------|-------------|
| `public_ip_id` | Azure Resource Manager ID (used by downstream resources) |
| `ip_address` | The allocated address; static for the resource's lifetime |
| `fqdn` | Azure-managed FQDN (only populated when `domain_name_label` is set) |
| `public_ip_name` | Name of the resource |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePublicIp
metadata:
  name: gateway-pip
  org: mycompany
  env: production
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      name: network-rg
  name: prod-gateway-pip
  domainNameLabel: prod-gateway
  domainNameLabelScope: TENANT_REUSE
  zones:
    - "1"
    - "2"
    - "3"
```

An address drawn from a reserved prefix, on the next-generation SKU:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePublicIp
metadata:
  name: nat-egress-pip
  org: mycompany
  env: production
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      name: network-rg
  name: prod-nat-egress-pip
  sku: STANDARD_V2
  publicIpPrefixId:
    valueFrom:
      name: prod-snat-prefix
```

## Downstream Usage

Other Azure resources reference this Public IP via `StringValueOrRef`:

```yaml
# In an AzureNatGateway spec:
spec:
  publicIpIds:
    - valueFrom:
        name: gateway-pip
```

## What's NOT Included

- **Basic SKU** -- retired by Azure (September 30, 2025), not supported
- **Dynamic allocation** -- every current SKU requires static allocation
- **DDoS protection plan creation** -- the plan is referenced by ARM ID;
  plans are shared, rarely-created governance resources
- **Attachment to consumers** -- load balancers, application gateways, and NAT
  gateways reference the address by ID from their own specs

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

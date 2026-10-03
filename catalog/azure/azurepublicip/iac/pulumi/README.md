# AzurePublicIp Pulumi Module

## Overview

This Pulumi module provisions an Azure Public IP Address using the Azure
Classic provider (`pulumi-azure`). It creates a single `network.PublicIp`
with static allocation (every current SKU requires it), covering the full
azurerm surface: SKU and tier, IP version, zones, prefix allocation, DNS
label with scope-based reuse, reverse FQDN, idle timeout, IP tags, DDoS
protection stance, and edge zones.

Reverse FQDN, DDoS settings, idle timeout, and tags update in place. Name,
SKU/tier, IP version, zones, prefix membership, IP tags, and edge zone are
fixed at creation -- changing any of them replaces the resource and with it
the actual address, so treat replacement as a coordinated migration (DNS,
allowlists).

Enum fields are only sent when explicitly chosen, so an unspecified spec
deploys Azure's defaults (Standard / Regional / IPv4 / region-unique label /
inherited DDoS stance) identically on both engines.

## Resources Created

- `network.PublicIp` -- the static public IP address

## Inputs

The module receives an `AzurePublicIpIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the address's ARM identity (references resolved to literals by the platform)
- `target.spec.sku` / `target.spec.sku_tier` / `target.spec.ip_version` -- STANDARD/STANDARD_V2, REGIONAL/GLOBAL, IPV4/IPV6; unset defers to Azure's defaults
- `target.spec.zones` -- availability zones; multiple zones make the address zone-redundant
- `target.spec.public_ip_prefix_id` -- the AzurePublicIpPrefix to allocate from (reference resolved to a literal ARM ID by the platform)
- `target.spec.domain_name_label` / `target.spec.domain_name_label_scope` -- the Azure-managed DNS label and its hashed-reuse policy
- `target.spec.reverse_fqdn` -- the reverse-DNS (PTR) name; the forward record must exist first
- `target.spec.idle_timeout_in_minutes` -- TCP idle timeout, 4-30 minutes (Azure defaults to 4)
- `target.spec.ip_tags` -- Azure routing metadata (e.g. RoutingPreference), not governance tags
- `target.spec.ddos_protection_mode` / `target.spec.ddos_protection_plan_id` -- per-IP DDoS stance; the plan is only valid with the ENABLED mode (pairing enforced by spec validation)
- `target.spec.edge_zone` -- optional Azure Edge Zone deployment
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `public_ip_id` | Full ARM ID of the address -- the join key downstream consumers attach through |
| `ip_address` | The allocated address; static for the resource's lifetime |
| `fqdn` | The Azure-managed FQDN; populated only when a domain name label is set |
| `public_ip_name` | The address's name as deployed |

## Local Development

```bash
make deps    # Tidy Go modules
make build   # Build module and entrypoint
make test    # Run tests
```

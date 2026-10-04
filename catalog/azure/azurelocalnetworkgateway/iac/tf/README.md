# AzureLocalNetworkGateway Terraform Module

## Overview

This Terraform module provisions a local network gateway using the
`azurerm` provider. It creates an `azurerm_local_network_gateway` --
Azure's description of the on-premises side of a site-to-site VPN: the
device's public endpoint and the address space behind it. Nothing
deploys on-premises; the object provisions in seconds and costs nothing
to keep.

The endpoint contract (exactly one of address or FQDN) and the
routing-source contract (static prefixes, BGP, or both -- never
neither) are spec-validated before the module runs. In azurerm v5 the
address space is a SET -- order is not significant.

## Resources Created

- `azurerm_local_network_gateway.main` -- the site description

## Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Planton metadata (name, org, env) |
| `spec` | object | Local network gateway specification |

Key spec fields:

| Field | Required | Description |
|-------|----------|-------------|
| `region` / `resource_group` / `name` | yes | The object's ARM identity; all ForceNew |
| `gateway_address` | one-of | The device's static public IPv4 |
| `gateway_fqdn` | one-of | The device's re-resolved public name |
| `address_spaces` | no* | CIDRs Azure routes into the tunnel (*or BGP) |
| `bgp_settings` | no* | The site's BGP speaker: ASN, tunnel-interior peering address, weight |
| `tags` | no | User tags, merged over metadata-derived tags |

## Outputs

| Output | Description |
|--------|-------------|
| `local_network_gateway_id` | Full ARM ID -- the join key connections reference |
| `local_network_gateway_name` | The description's name as deployed |

## Usage

```hcl
module "hq_site" {
  source = "./iac/tf"

  metadata = { name = "hq-datacenter", org = "mycompany", env = "production" }

  spec = {
    region          = "eastus"
    resource_group  = "network-rg"
    name            = "hq-datacenter"
    gateway_address = "198.51.100.4"
    address_spaces  = ["192.168.0.0/16"]
  }
}
```

## Behavior Notes

- Everything except name/region/resource-group updates in place.
- Editing the prefix list on a BGP-carrying site takes two ARM
  round-trips (an ARM constraint the provider sequences internally).

## Required Permissions

Least-privilege runner permissions for this kind are declared in [`../permissions.yaml`](../permissions.yaml).

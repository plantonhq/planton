# Cloudflare Zero Trust Access Infrastructure Target

## Overview

`CloudflareZeroTrustAccessInfrastructureTarget` registers an infrastructure target: a server (by hostname and private IP) that Access infrastructure applications grant SSH access to. Targets are the inventory layer -- an infrastructure application selects targets by hostname or IP, and Access brokers short-lived SSH access to them through the account's tunnels.

## Key Features

- **Inventory as code** -- every SSH-reachable server declared, named, and versioned
- **Virtual-network placement** -- per-address-family placement into a `CloudflareZeroTrustTunnelVirtualNetwork` (omit for the account default)
- **Plain lifecycle** -- real create, in-place updates, real delete

## Use Cases

**Ideal for:**

- Registering production servers behind Access infrastructure applications (SSH with short-lived certificates)
- Overlapping-CIDR estates: the same private address in different virtual networks

**Not ideal for:**

- HTTP applications -- those are `CloudflareZeroTrustAccessApplication` in its web shapes
- Machines without a tunnel path -- targets are reached through the account's Cloudflare Tunnels

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `hostname` | string | Yes | The target's hostname (≤255 chars, letters/digits/dashes/periods, alphanumeric ends). The selection surface for applications. |
| `ip` | object | Yes | At least one of `ipv4` / `ipv6`, each with `ip_addr` and an optional `virtual_network_id`. |

### Outputs

| Field | Description |
|-------|-------------|
| `target_id` | The Cloudflare-assigned UUID of the target |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessInfrastructureTarget
metadata:
  name: prod-db-1
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  hostname: prod-db-1
  ip:
    ipv4:
      ip_addr: 10.0.10.5
```

## Destroy Semantics

Destroy is a real delete: the target leaves the inventory and applications selecting it lose the route.

## Related Resources

- **CloudflareZeroTrustTunnelVirtualNetwork** -- the network segment an address belongs to
- **CloudflareZeroTrustAccessApplication** -- the infrastructure applications that select targets
- **CloudflareZeroTrustTunnel** -- the data path Access brokers connections through

## Further Reading

For operational judgment -- hostname naming as the access-control surface, default-vnet behavior -- see GUIDE.md.

## References

- [Cloudflare Access infrastructure applications](https://developers.cloudflare.com/cloudflare-one/applications/non-http/infrastructure-apps/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

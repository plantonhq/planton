# Cloudflare Zero Trust DNS Location

## Overview

`CloudflareZeroTrustDnsLocation` creates a Gateway DNS location: a named entry point (office, site, network) whose DNS traffic Gateway filters. Cloudflare assigns the location its resolver endpoints -- a DoH subdomain and destination IPs -- and Gateway policies can then match on the location.

## Key Features

- **Four endpoint types** -- DoH (with optional token gating), DoT, plain IPv4, plain IPv6, declared as one tree
- **Source-network allowlists** -- per endpoint, plus the IPv4 endpoint's top-level networks (CIDRs to /24)
- **TTL capping** -- inherit the account setting, override per location (60-36000s), or disable
- **EDNS Client Subnet** -- optional ECS support for better geo-answers

## Use Cases

**Ideal for:**

- Per-office DNS filtering: each site a location, policies matching per site
- Token-gated DoH for roaming or unmanaged devices

**Not ideal for:**

- The filtering rules themselves -- those are `CloudflareZeroTrustGatewayPolicy`
- WARP-enrolled devices -- WARP carries its own resolver path

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `name` | string | Yes | The location's name (matched by policies). |

### Key Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `endpoints` | object | ALL FOUR types (doh/dot/ipv4/ipv6) declared at once. `require_token` exists on doh only; `networks` on doh/dot/ipv6. |
| `networks` | list | Source IPv4 CIDRs (to /24) for the IPv4 endpoint. |
| `max_ttl` | object | mode inherit/override/disabled; `ttl_secs` (60-36000) required exactly with override. |
| `dns_destination_ips_id` | string | Leave unset for the shared pool auto-assign. |
| `client_default` | bool | Make this the account's default location. |

### Outputs

| Field | Description |
|-------|-------------|
| `location_id` | The Cloudflare-assigned UUID |
| `doh_subdomain` | The unique DoH subdomain clients embed |
| `ip` | The IPv4 destination for the plain-DNS endpoint |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDnsLocation
metadata:
  name: hq-office
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: hq-office
  endpoints:
    doh: { enabled: true, require_token: true }
    dot: { enabled: false }
    ipv4: { enabled: true }
    ipv6: { enabled: false }
  networks:
    - network: 203.0.113.0/24
```

## Destroy Semantics

Destroy is a real delete: the location and its assigned endpoints disappear; devices pointed at them lose resolution through Gateway.

## Related Resources

- **CloudflareZeroTrustGatewayPolicy** -- the rules that filter this location's queries
- **CloudflareZeroTrustGatewaySettings** -- the account-wide Gateway posture

## Further Reading

For operational judgment -- the full-replace update semantics, the client_default lever, token-gated DoH -- see GUIDE.md.

## References

- [Cloudflare Gateway DNS locations](https://developers.cloudflare.com/cloudflare-one/connections/connect-devices/agentless/dns/locations/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

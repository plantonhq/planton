---
title: "DNS Zone"
description: "DNS Zone deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarednszone"
---

# Cloudflare DNS Zone

Deploys a Cloudflare DNS zone with optional inline DNS records, zone-wide DNS settings, and DNSSEC. The component creates the zone, exports the assigned nameservers and (when enabled) the DNSSEC DS material, and provisions any DNS records defined in the spec.

## What Gets Created

When you deploy a CloudflareDnsZone resource, Planton provisions:

- **DNS Zone** — a `cloudflare_zone` resource attached to the specified account, with configurable type, pause state, and vanity name servers
- **DNS Records** — one `cloudflare_dns_record` per entry in the `records` list (the lean inline model; use standalone CloudflareDnsRecord resources for the full record feature set)
- **DNS Settings** — a `cloudflare_zone_dns_settings` resource when `dnsSettings` is provided (CNAME flattening, zone mode, SOA, nameservers, NS TTL)
- **DNSSEC** — a `cloudflare_zone_dnssec` resource when `dnssec.enabled` is true; the DS material is exported for entry at your registrar

## Prerequisites

- **Cloudflare credentials** configured via environment variables or Planton provider config
- **A Cloudflare account ID** with permission to create zones
- **Domain ownership** — you must control the domain and update its registrar nameservers to the values returned in stack outputs

## Quick Start

Create a file `dns-zone.yaml`:

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareDnsZone
metadata:
  name: my-zone
spec:
  zoneName: example.com
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

Deploy:

```shell
planton apply -f dns-zone.yaml
```

This creates a full (Cloudflare-hosted) DNS zone for `example.com`. Update your domain registrar's nameservers to the values in `status.outputs.nameservers` to activate the zone.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zoneName` | `string` | Fully qualified domain name for the zone (e.g., `example.com`). | Must match a valid FQDN pattern |
| `accountId` | `string` | Cloudflare account ID under which to create the zone. | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | `enum` | `full` | Zone deployment type: `full`, `partial`, `secondary`, `internal`. |
| `paused` | `bool` | `false` | When `true`, the zone is DNS-only with no Cloudflare proxy/CDN/security. |
| `vanityNameServers` | `string[]` | `[]` | Custom (vanity) name servers (Business/Enterprise plans). |
| `records` | `object[]` | `[]` | Inline DNS records. Each has `name`, `type`, `content`, and optional `proxied`, `ttl`, `priority`, `comment`. |
| `dnsSettings` | `object` | — | Zone-wide DNS settings: `flattenAllCnames`, `foundationDns`, `multiProvider`, `secondaryOverrides`, `nsTtl`, `zoneMode`, `soa`, `nameservers`, `internalDns`. |
| `dnssec` | `object` | — | DNSSEC config: `enabled`, `multiSigner`, `presigned`, `useNsec3`. When enabled, Cloudflare signs the zone and the DS material is exported. |

## Examples

### Zone with Common DNS Records

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareDnsZone
metadata:
  name: app-zone
spec:
  zoneName: myapp.com
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  records:
    - name: "@"
      type: A
      content: "203.0.113.50"
      proxied: true
      ttl: 1
    - name: www
      type: CNAME
      content: myapp.com
      proxied: true
      ttl: 1
    - name: "@"
      type: MX
      content: mail.myapp.com
      priority: 10
      ttl: 3600
```

### Zone with DNS Settings and DNSSEC

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareDnsZone
metadata:
  name: secure-zone
spec:
  zoneName: production.com
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  dnsSettings:
    flattenAllCnames: true
    zoneMode: standard
    soa:
      refresh: 10000
      retry: 2400
      expire: 604800
      minTtl: 1800
      ttl: 3600
  dnssec:
    enabled: true
    useNsec3: true
```

After apply, read `status.outputs.dnssec_ds` (and the individual digest/key-tag fields) and enter them at your registrar to complete the DNSSEC chain of trust.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | `string` | The Cloudflare Zone ID of the created DNS zone |
| `nameservers` | `string[]` | The nameserver addresses assigned to the zone |
| `status` | `string` | The zone status on Cloudflare |
| `dnssec_ds` | `string` | The full DS record to enter at the registrar (empty unless DNSSEC is enabled) |
| `dnssec_digest`, `dnssec_digest_type`, `dnssec_digest_algorithm` | `string` | DS digest material |
| `dnssec_algorithm`, `dnssec_key_tag`, `dnssec_public_key`, `dnssec_flags` | `string` | DNSKEY material |
| `dnssec_status` | `string` | DNSSEC status (empty unless enabled) |

## Related Components

- [CloudflareDnsRecord](/docs/catalog/cloudflare/dns-record) — manages individual DNS records as standalone resources with the full record feature set
- [CloudflareR2Bucket](/docs/catalog/cloudflare/r2-bucket) — references this zone via `customDomains[].zoneId` for custom domain bucket access
- [CloudflareWorker](/docs/catalog/cloudflare/worker) — commonly deployed with DNS routes pointing to Worker endpoints
- [CloudflareLoadBalancer](/docs/catalog/cloudflare/load-balancer) — load balances traffic across origins within the zone

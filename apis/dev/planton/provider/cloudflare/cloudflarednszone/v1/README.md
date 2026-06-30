# Cloudflare DNS Zone

Provision and manage Cloudflare DNS zones using Planton's unified API.

## Overview

Cloudflare DNS provides authoritative DNS served from a global anycast network, with built-in DDoS protection, zero per-query charges, and optional integrated CDN/WAF/proxy capabilities. This component creates a zone and, optionally, manages inline DNS records, zone-wide DNS settings, and DNSSEC alongside it.

## Key Features

- **Global Anycast DNS**: authoritative DNS from a worldwide edge network
- **Zone types**: full, partial (CNAME setup), secondary, and internal zones
- **Inline records**: a lean set of DNS records managed with the zone (use standalone CloudflareDnsRecord for the full record feature set)
- **Folded DNS settings**: CNAME flattening, zone mode, SOA, nameserver set, and NS TTL
- **DNSSEC**: enable Cloudflare zone signing and export the DS material for your registrar
- **Vanity name servers**: custom name servers on Business/Enterprise plans

## Prerequisites

1. **Cloudflare Account**: an active account with permission to create zones
2. **API Token**: a Cloudflare API token with `Zone:Edit` (and `DNS:Edit` for records/DNSSEC)
3. **Planton CLI**: install from [planton.dev](https://planton.dev)

## Quick Start

### Minimal Configuration

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareDnsZone
metadata:
  name: my-zone
spec:
  zone_name: "example.com"
  account_id: "your-cloudflare-account-id"
```

Deploy:

```bash
planton apply -f zone.yaml
```

The zone defaults to a `full` (Cloudflare-hosted) zone. Update your registrar's
nameservers to the values in `status.outputs.nameservers` to activate it.

### With Inline Records

```yaml
spec:
  zone_name: "example.com"
  account_id: "your-cloudflare-account-id"
  records:
    - name: "@"
      type: A
      content: "203.0.113.50"
      proxied: true
    - name: "@"
      type: MX
      content: mail.example.com
      priority: 10
```

### With DNS Settings and DNSSEC

```yaml
spec:
  zone_name: "example.com"
  account_id: "your-cloudflare-account-id"
  dns_settings:
    flatten_all_cnames: true
    zone_mode: standard
  dnssec:
    enabled: true
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone_name` | string | Fully qualified domain name (e.g., "example.com") |
| `account_id` | string | Cloudflare account ID |

### Optional Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `type` | enum | Zone type: full, partial, secondary, internal | full |
| `paused` | bool | If true, zone is DNS-only (no proxy/CDN/WAF) | false |
| `vanity_name_servers` | []string | Custom name servers (Business/Enterprise) | [] |
| `records` | object[] | Inline DNS records (name, type, content, proxied, ttl, priority, comment) | [] |
| `dns_settings` | object | Zone-wide DNS settings (see below) | — |
| `dnssec` | object | DNSSEC config: enabled, multi_signer, presigned, use_nsec3 | — |

### DNS Settings

`dns_settings` folds the zone's DNS-level options: `flatten_all_cnames`,
`foundation_dns`, `multi_provider`, `secondary_overrides`, `ns_ttl`, `zone_mode`
(standard/cdn_only/dns_only), `soa` (expire/min_ttl/mname/refresh/retry/rname/ttl),
`nameservers` (ns_set/type), and `internal_dns` (reference_zone_id).

## Outputs

| Output | Description |
|--------|-------------|
| `zone_id` | The unique identifier of the created zone |
| `nameservers` | The Cloudflare nameservers assigned to this zone |
| `status` | The zone status on Cloudflare |
| `dnssec_ds` and friends | DS record material to enter at your registrar (only when DNSSEC is enabled) |

```bash
planton output zone_id
planton output nameservers
```

## Nameserver Configuration

After creating the zone, update your domain's nameservers at your registrar to the
Cloudflare nameservers returned in `status.outputs.nameservers`, then wait for
propagation (typically 1-24 hours).

## DNSSEC

Set `dnssec.enabled: true` to have Cloudflare sign the zone. After apply, read the
`dnssec_ds` output (and the individual digest/key-tag fields) and enter them at your
registrar to complete the chain of trust. DNSSEC fully activates only once the zone
is active and the DS records are accepted by the registrar.

## Terraform and Pulumi

This component supports both Pulumi (default) and Terraform, producing identical infrastructure:

- **Pulumi**: `iac/pulumi/` — Go-based implementation
- **Terraform**: `iac/tf/` — HCL-based implementation

## Support

- **Architecture details**: [docs/README.md](docs/README.md)
- **Cloudflare DNS Docs**: [developers.cloudflare.com/dns](https://developers.cloudflare.com/dns)
- **Planton**: [planton.dev](https://planton.dev)

## License

This component is part of Planton and follows the same license.

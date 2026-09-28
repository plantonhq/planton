# Cloudflare DNS Zone

Provision and manage Cloudflare DNS zones using Planton's unified API.

## Overview

Cloudflare DNS provides authoritative DNS served from a global anycast network, with built-in DDoS protection, zero per-query charges, and optional integrated CDN/WAF/proxy capabilities. This component creates a zone and, optionally, manages inline DNS records, zone-wide DNS settings, and DNSSEC alongside it.

## Key Features

- **Global Anycast DNS**: authoritative DNS from a worldwide edge network
- **Zone types**: full, partial (CNAME setup), secondary, and internal zones
- **Inline records at full depth**: all 21 record types managed with the zone — simple records via `content` (a literal, or a reference to another resource's output), structured records (SRV, CAA, TLSA, LOC, …) via typed data blocks, plus per-record tags, settings, and private routing. Records with independent lifecycles are better modeled as standalone CloudflareDnsRecord resources; the surface is identical.
- **Folded DNS settings**: CNAME flattening, zone mode, SOA, nameserver set, and NS TTL
- **DNSSEC**: enable Cloudflare zone signing and export the DS material for your registrar
- **Zone hold**: block the zone's hostname (and optionally subdomains) from being added as a zone in any other Cloudflare account
- **Plan subscription**: subscribe the zone to a Cloudflare rate plan directly from the spec
- **Vanity name servers**: custom name servers on Business/Enterprise plans

## Prerequisites

1. **Cloudflare Account**: an active account with permission to create zones
2. **API Token**: a Cloudflare API token with `Zone:Edit` (and `DNS:Edit` for records/DNSSEC)
3. **Planton CLI**: install from [planton.dev](https://planton.dev)

## Quick Start

### Minimal Configuration

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
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
      content:
        value: "203.0.113.50"
      proxied: true
    - name: "@"
      type: MX
      content:
        value: mail.example.com
      priority: 10
```

### With Structured Records

Structured record types carry their fields in a typed block named after the
record type (exactly one of `content` or a typed block is set, and it must
match `type`):

```yaml
spec:
  zone_name: "example.com"
  account_id: "your-cloudflare-account-id"
  records:
    - name: _sip._tcp
      type: SRV
      srv:
        priority: 10
        weight: 5
        port: 5060
        target:
          value: sip.example.com
    - name: "@"
      type: CAA
      caa:
        tag: issue
        value: letsencrypt.org
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
| `records` | object[] | Inline DNS records: name, type (all 21 types), content or a typed data block (srv/caa/cert/dnskey/ds/https/loc/naptr/smimea/sshfp/svcb/tlsa/uri), proxied, ttl, priority, comment, tags, settings (ipv4_only/ipv6_only/flatten_cname), private_routing | [] |
| `dns_settings` | object | Zone-wide DNS settings (see below) | — |
| `dnssec` | object | DNSSEC config: enabled, multi_signer, presigned, use_nsec3 | — |
| `hold` | object | Zone hold: enabled, include_subdomains, hold_after (RFC3339) | — |
| `subscription` | object | Zone plan: rate_plan (free/lite/pro/pro_plus/business/enterprise/partner variants), frequency, scope. A paid plan bills real money and needs Billing Write token scope | — |

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

## Zone Hold and Plan

Set `hold.enabled: true` to protect the zone's hostname from being created as a
zone in any other Cloudflare account (add `include_subdomains: true` to extend
the hold to every subdomain) — the standard guard during account migrations.
Set `subscription.rate_plan` to subscribe the zone to a Cloudflare plan; paid
plans start billing at apply and the deploying API token needs Billing Write
scope.

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

- **Cloudflare DNS Docs**: [developers.cloudflare.com/dns](https://developers.cloudflare.com/dns)
- **Planton**: [planton.dev](https://planton.dev)

## License

This component is part of Planton and follows the same license.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

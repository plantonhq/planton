# Cloudflare DNS Record

Provision and manage individual DNS records in Cloudflare zones using Planton's unified API.

## Overview

Cloudflare DNS provides authoritative DNS served from a global anycast network, with built-in DDoS protection, zero per-query charges, and optional integrated CDN/WAF/proxy capabilities. This component manages a single DNS record within a Cloudflare-managed zone and covers the full Cloudflare record surface — every record type, structured record data, tags, and record-level settings.

A record is either **simple** (its value is a presentation-format string in `content`) or **structured** (its components are supplied through a typed `data` block). The component validates which representation a given type requires.

## Key Features

- **All 21 record types**: A, AAAA, CNAME, MX, NS, PTR, TXT, OPENPGPKEY, CAA, CERT, DNSKEY, DS, HTTPS, LOC, NAPTR, SMIMEA, SRV, SSHFP, SVCB, TLSA, URI
- **Typed structured data**: per-type `data` blocks for CAA, CERT, DNSKEY, DS, HTTPS, LOC, NAPTR, SMIMEA, SRV, SSHFP, SVCB, TLSA, and URI records
- **Orange Cloud integration**: optional Cloudflare proxy (CDN/WAF) for A, AAAA, and CNAME records
- **TTL control**: automatic (1) or custom TTL
- **Tags and settings**: custom record tags and proxied-record settings (`ipv4_only`, `ipv6_only`, `flatten_cname`)
- **Validation**: record-type/representation coherence, TTL and priority ranges, and cross-field rules

## Prerequisites

1. **Cloudflare DNS Zone**: an existing zone where records will be created (use the CloudflareDnsZone component)
2. **Zone ID**: the Cloudflare Zone ID (from CloudflareDnsZone outputs or the dashboard)
3. **API Token**: a Cloudflare API token with `DNS:Edit`
4. **Planton CLI**: install from [planton.dev](https://planton.dev)

## Quick Start

### A Record (IPv4)

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareDnsRecord
metadata:
  name: www-a-record
spec:
  zone_id: "your-zone-id-here"
  name: "www"
  type: A
  content: "192.0.2.1"
  proxied: true
```

### MX Record (Email)

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareDnsRecord
metadata:
  name: mx-primary
spec:
  zone_id: "your-zone-id-here"
  name: "@"
  type: MX
  content: "mail.example.com"
  priority: 10
```

### SRV Record (structured data)

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareDnsRecord
metadata:
  name: sip-srv
spec:
  zone_id: "your-zone-id-here"
  name: "_sip._tcp"
  type: SRV
  data:
    srv:
      priority: 10
      weight: 5
      port: 5060
      target: "sip.example.com"
```

### CAA Record (structured data)

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareDnsRecord
metadata:
  name: caa-letsencrypt
spec:
  zone_id: "your-zone-id-here"
  name: "@"
  type: CAA
  data:
    caa:
      flags: 0
      tag: issue
      value: "letsencrypt.org"
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone_id` | string-or-ref | Cloudflare Zone ID (literal or reference to a CloudflareDnsZone) |
| `name` | string | Record name (e.g., "www", "@" for the apex) |
| `type` | enum | DNS record type (one of the 21 supported types) |

Exactly one of `content` or a `data` block is required, and it must match the record `type`.

### Value Fields

| Field | Type | Description |
|-------|------|-------------|
| `content` | string | Presentation-format value for **simple** types (A, AAAA, CNAME, MX, NS, PTR, TXT, OPENPGPKEY) |
| `data` | oneof | Typed block for **structured** types (caa, cert, dnskey, ds, https, loc, naptr, smimea, srv, sshfp, svcb, tlsa, uri) |

### Optional Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `proxied` | bool | Route through Cloudflare CDN/WAF (A, AAAA, CNAME only) | false |
| `ttl` | int32 | Time to live: 0/1 (auto) or 30-86400 seconds | auto |
| `priority` | int32 | Priority for MX records (0-65535) | 0 |
| `comment` | string | Free-form note (no effect on DNS responses) | "" |
| `tags` | []string | Custom tags for organizing/filtering records | [] |
| `settings` | object | `ipv4_only`, `ipv6_only`, `flatten_cname` (proxied records only) | — |

### Record Types

Simple types use `content`:

| Type | Example `content` |
|------|-------------------|
| A | `192.0.2.1` |
| AAAA | `2001:db8::1` |
| CNAME | `www.example.com` |
| MX | `mail.example.com` (with `priority`) |
| NS | `ns1.example.com` |
| PTR | `host.example.com` |
| TXT | `v=spf1 include:_spf.google.com ~all` |
| OPENPGPKEY | base64 OpenPGP key |

Structured types use the matching `data` block: CAA, CERT, DNSKEY, DS, HTTPS, LOC, NAPTR, SMIMEA, SRV, SSHFP, SVCB, TLSA, URI. SRV/URI/HTTPS/SVCB carry their own `priority` inside `data`.

## Outputs

| Output | Description |
|--------|-------------|
| `record_id` | Unique identifier of the created DNS record |
| `record_name` | The record name as stored by Cloudflare |
| `record_type` | The DNS record type that was created |
| `proxied` | Whether the record is proxied through Cloudflare |

```bash
planton output record_id
planton output record_name
```

## Orange Cloud vs Grey Cloud

Cloudflare's signature feature is the **proxy toggle**:

- **Orange Cloud (`proxied: true`)**: traffic flows through Cloudflare — CDN caching, WAF, DDoS protection, SSL termination, and a hidden origin IP. Use for web services (www, app, api).
- **Grey Cloud (`proxied: false`)**: DNS-only resolution, direct to origin. Use for email (MX), SSH, VPN, and non-HTTP services.

Only A, AAAA, and CNAME records can be proxied.

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

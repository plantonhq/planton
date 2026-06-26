# Cloudflare DNS Record

Deploys a single DNS record into an existing Cloudflare DNS zone. The component supports all 21 Cloudflare record types — simple types whose value is a `content` string (A, AAAA, CNAME, MX, NS, PTR, TXT, OPENPGPKEY) and structured types configured through a typed `data` block (CAA, CERT, DNSKEY, DS, HTTPS, LOC, NAPTR, SMIMEA, SRV, SSHFP, SVCB, TLSA, URI) — with optional Cloudflare proxy (orange-cloud) mode for A, AAAA, and CNAME records.

## What Gets Created

When you deploy a CloudflareDnsRecord resource, OpenMCF provisions:

- **DNS Record** — a `cloudflare_dns_record` resource in the specified zone, configured with the given type, value (`content` or a `data` block), TTL, proxy setting, tags, settings, and optional priority and comment

## Prerequisites

- **Cloudflare credentials** configured via environment variables or OpenMCF provider config
- **An existing Cloudflare DNS zone** — either the zone ID as a literal string or a deployed CloudflareDnsZone resource to reference
- **Appropriate permissions** — the API token must have DNS edit access for the target zone

## Quick Start

Create a file `dns-record.yaml`:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsRecord
metadata:
  name: my-record
spec:
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: www
  type: A
  content: "203.0.113.50"
  proxied: true
```

Deploy:

```shell
openmcf apply -f dns-record.yaml
```

This creates a proxied A record for `www` in the specified zone, routing traffic through Cloudflare's CDN and WAF.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zoneId` | `StringValueOrRef` | The Cloudflare Zone ID. Accepts a literal `value` or a `valueFrom` reference to a CloudflareDnsZone. | Required |
| `name` | `string` | The record name (e.g., `www`, `api`, `@` for zone apex). | Required, non-empty |
| `type` | `enum` | The DNS record type — one of the 21 supported types. | Required, defined value |

Exactly one of `content` or a `data` block is required and must match the `type`.

### Value Fields

| Field | Type | Description |
|-------|------|-------------|
| `content` | `string` | Presentation-format value for simple types (A, AAAA, CNAME, MX, NS, PTR, TXT, OPENPGPKEY). |
| `data` | `oneof` | Typed block for structured types: `caa`, `cert`, `dnskey`, `ds`, `https`, `loc`, `naptr`, `smimea`, `srv`, `sshfp`, `svcb`, `tlsa`, `uri`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `proxied` | `bool` | `false` | Route traffic through Cloudflare's CDN/WAF (orange-cloud). Only valid for `A`, `AAAA`, `CNAME`. |
| `ttl` | `int32` | `1` (auto) | Time to live in seconds. `0`/`1` for automatic, or `30`–`86400`. |
| `priority` | `int32` | `0` | Required for `MX` records. Range 0–65535. (SRV/URI/HTTPS/SVCB carry priority inside `data`.) |
| `comment` | `string` | `""` | A note describing the record's purpose. |
| `tags` | `string[]` | `[]` | Custom tags for organizing and filtering records. |
| `settings` | `object` | — | `ipv4_only`, `ipv6_only`, `flatten_cname` (apply to proxied records only). |

### Zone ID Reference

The `zoneId` field accepts either a literal value or a cross-resource reference:

```yaml
spec:
  zoneId:
    valueFrom:
      name: my-zone
```

When using `valueFrom`, the `kind` defaults to `CloudflareDnsZone` and the `fieldPath` defaults to `status.outputs.zone_id`, so only the resource `name` is required.

## Examples

### MX Record for Mail Delivery

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsRecord
metadata:
  name: mail-mx-record
spec:
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: "@"
  type: MX
  content: aspmx.l.google.com
  priority: 1
  ttl: 3600
  comment: "Google Workspace primary"
```

### SRV Record (structured data)

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsRecord
metadata:
  name: sip-srv-record
spec:
  zoneId:
    valueFrom:
      name: prod-zone
  name: "_sip._tcp"
  type: SRV
  data:
    srv:
      priority: 10
      weight: 5
      port: 5060
      target: sip.example.com
```

### CAA Record (structured data)

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsRecord
metadata:
  name: caa-record
spec:
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: "@"
  type: CAA
  data:
    caa:
      flags: 0
      tag: issue
      value: letsencrypt.org
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `recordId` | `string` | The unique identifier of the created DNS record in Cloudflare |
| `recordName` | `string` | The record name as stored by Cloudflare |
| `recordType` | `string` | The DNS record type that was created |
| `proxied` | `bool` | Whether the record is proxied through Cloudflare |

## Related Components

- [CloudflareDnsZone](/docs/catalog/cloudflare/cloudflarednszone) — manages the parent DNS zone; its `zone_id` output can be referenced by this component via `valueFrom`
- [CloudflareR2Bucket](/docs/catalog/cloudflare/cloudflarer2bucket) — may use DNS records for custom domain access
- [CloudflareWorker](/docs/catalog/cloudflare/cloudflareworker) — commonly paired with DNS records pointing to Worker routes
- [CloudflareLoadBalancer](/docs/catalog/cloudflare/cloudflareloadbalancer) — load balances traffic across origins, often configured alongside DNS records

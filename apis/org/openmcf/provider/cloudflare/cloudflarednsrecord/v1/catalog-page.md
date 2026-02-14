# Cloudflare DNS Record

Deploys a single DNS record into an existing Cloudflare DNS zone. The component supports A, AAAA, CNAME, MX, TXT, SRV, NS, and CAA record types, with optional Cloudflare proxy (orange-cloud) mode for A, AAAA, and CNAME records.

## What Gets Created

When you deploy a CloudflareDnsRecord resource, OpenMCF provisions:

- **DNS Record** — a `cloudflare_dns_record` resource in the specified zone, configured with the given type, value, TTL, proxy setting, and optional priority and comment

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
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareDnsRecord.my-record
spec:
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: www
  type: A
  value: "203.0.113.50"
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
| `zoneId` | `StringValueOrRef` | The Cloudflare Zone ID where this DNS record will be created. Accepts a literal `value` string or a `valueFrom` reference to a CloudflareDnsZone resource. | Required |
| `name` | `string` | The name of the DNS record (e.g., `www`, `api`, `@` for zone apex). | Required, non-empty |
| `type` | `enum` | The DNS record type. One of: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `NS`, `CAA`. | Required, must be a defined value |
| `value` | `string` | The record value. For A records: an IPv4 address. For AAAA: an IPv6 address. For CNAME: a target hostname. For MX: a mail server hostname. For TXT: a text string. | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `proxied` | `bool` | `false` | Route traffic through Cloudflare's CDN/WAF (orange-cloud). Only applicable to `A`, `AAAA`, and `CNAME` records. The spec rejects `proxied: true` for other record types. |
| `ttl` | `int32` | `1` (auto) | Time to live in seconds. `1` for automatic TTL (recommended for proxied records), or `60`–`86400`. A value of `0` is treated as `1` (automatic). |
| `priority` | `int32` | `0` | Record priority. Required for `MX` records, optional for `SRV`. Range: 0–65535. |
| `comment` | `string` | `""` | A note describing the record's purpose. Maximum 100 characters. |

### Zone ID Reference

The `zoneId` field accepts either a literal value or a cross-resource reference.

**Literal value:**

```yaml
spec:
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

**Reference to a CloudflareDnsZone resource:**

```yaml
spec:
  zoneId:
    valueFrom:
      name: my-zone
```

When using `valueFrom`, the `kind` defaults to `CloudflareDnsZone` and the `fieldPath` defaults to `status.outputs.zone_id`, so only the resource `name` is required. You may also specify `env` to reference a zone deployed in a different environment.

## Examples

### Proxied A Record

An A record with Cloudflare proxy enabled, suitable for a web server:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsRecord
metadata:
  name: web-a-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareDnsRecord.web-a-record
spec:
  zoneId:
    valueFrom:
      name: prod-zone
  name: "@"
  type: A
  value: "198.51.100.10"
  proxied: true
  ttl: 1
  comment: "Production web server"
```

### MX Record for Mail Delivery

An MX record pointing to a mail server, with priority set:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsRecord
metadata:
  name: mail-mx-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareDnsRecord.mail-mx-record
spec:
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: "@"
  type: MX
  value: aspmx.l.google.com
  priority: 1
  ttl: 3600
  comment: "Google Workspace primary"
```

### TXT Record for SPF

A TXT record at the zone apex defining an SPF policy:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsRecord
metadata:
  name: spf-txt-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareDnsRecord.spf-txt-record
spec:
  zoneId:
    valueFrom:
      name: prod-zone
      env: prod
  name: "@"
  type: TXT
  value: "v=spf1 include:_spf.google.com ~all"
  ttl: 3600
  comment: "SPF for Google Workspace"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `recordId` | `string` | The unique identifier of the created DNS record in Cloudflare |
| `hostname` | `string` | The fully qualified hostname of the DNS record (e.g., `www.example.com`) |
| `recordType` | `string` | The DNS record type that was created (e.g., `A`, `CNAME`) |
| `proxied` | `bool` | Whether the record is proxied through Cloudflare |

## Related Components

- [CloudflareDnsZone](/docs/catalog/cloudflare/cloudflarednszone) — manages the parent DNS zone; its `zone_id` output can be referenced by this component via `valueFrom`
- [CloudflareR2Bucket](/docs/catalog/cloudflare/cloudflarer2bucket) — may use DNS records for custom domain access
- [CloudflareWorker](/docs/catalog/cloudflare/cloudflareworker) — commonly paired with DNS records pointing to Worker routes
- [CloudflareLoadBalancer](/docs/catalog/cloudflare/cloudflareloadbalancer) — load balances traffic across origins, often configured alongside DNS records

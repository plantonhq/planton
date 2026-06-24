# Cloudflare DNS Zone

Deploys a Cloudflare DNS zone with optional inline DNS record management. The component creates the zone, exports the assigned nameservers, and provisions any DNS records defined in the spec — supporting A, AAAA, CNAME, MX, TXT, SRV, NS, and CAA record types.

## What Gets Created

When you deploy a CloudflareDnsZone resource, OpenMCF provisions:

- **DNS Zone** — a `cloudflare_zone` resource attached to the specified Cloudflare account, with configurable pause state
- **DNS Records** — one `cloudflare_dns_record` resource per entry in the `records` list, created within the zone with support for proxied mode, custom TTL, priority, and comments

## Prerequisites

- **Cloudflare credentials** configured via environment variables or OpenMCF provider config
- **A Cloudflare account ID** with permission to create zones
- **Domain ownership** — you must own or control the domain being added as a zone, and update its registrar nameservers to the values returned in stack outputs

## Quick Start

Create a file `dns-zone.yaml`:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: my-zone
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareDnsZone.my-zone
spec:
  zoneName: example.com
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

Deploy:

```shell
openmcf apply -f dns-zone.yaml
```

This creates a DNS zone for `example.com` on the Free plan. Update your domain registrar's nameservers to the values in `status.outputs.nameservers` to activate the zone.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zoneName` | `string` | Fully qualified domain name for the zone (e.g., `example.com`). | Must match a valid FQDN pattern |
| `accountId` | `string` | Cloudflare account ID under which to create the zone. | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `plan` | `enum` | `FREE` | Subscription plan for the zone. One of: `FREE`, `PRO`, `BUSINESS`, `ENTERPRISE`. Note: plan management may require separate Cloudflare account configuration. |
| `paused` | `bool` | `false` | When `true`, creates the zone in paused (DNS-only) mode with no Cloudflare proxy, CDN, or security features active. |
| `defaultProxied` | `bool` | `false` | When `true`, new DNS records in the zone default to being proxied (orange-cloud) through Cloudflare. Note: zone-level default proxied setting may require separate configuration in the Cloudflare dashboard. |
| `records` | `object[]` | `[]` | DNS records to create within the zone. See record fields below. |
| `records[].name` | `string` | — | Record name (e.g., `www`, `api`, `@` for zone apex). Required per record. |
| `records[].type` | `enum` | — | Record type. One of: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `NS`, `CAA`. Required per record. |
| `records[].value` | `string` | — | Record value (IP address, hostname, or text depending on type). Required per record. |
| `records[].proxied` | `bool` | `false` | Route traffic through Cloudflare's CDN/WAF (orange-cloud). Only applicable to `A`, `AAAA`, and `CNAME` records. |
| `records[].ttl` | `int` | `1` | Time to live in seconds. `1` for automatic (recommended for proxied records), or `60`–`86400`. |
| `records[].priority` | `int` | `0` | Record priority. Required for `MX` records, optional for `SRV`. Range: 0–65535. |
| `records[].comment` | `string` | `""` | Optional note describing the record's purpose. Maximum 100 characters. |

## Examples

### Basic Zone

A DNS zone with no records, useful when records are managed separately or via CloudflareDnsRecord resources:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: example-zone
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareDnsZone.example-zone
spec:
  zoneName: example.com
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

### Zone with Common DNS Records

A zone with A, CNAME, and MX records for a typical web application:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: app-zone
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareDnsZone.app-zone
spec:
  zoneName: myapp.com
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  records:
    - name: "@"
      type: A
      value: "203.0.113.50"
      proxied: true
      ttl: 1
    - name: www
      type: CNAME
      value: myapp.com
      proxied: true
      ttl: 1
    - name: "@"
      type: MX
      value: mail.myapp.com
      priority: 10
      ttl: 3600
```

### Full-Featured Zone with Multiple Record Types

Production zone with proxied web records, mail configuration, SPF, and a paused initial state:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareDnsZone
metadata:
  name: prod-zone
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareDnsZone.prod-zone
spec:
  zoneName: production.com
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  paused: false
  records:
    - name: "@"
      type: A
      value: "198.51.100.10"
      proxied: true
      ttl: 1
      comment: "Production web server"
    - name: www
      type: CNAME
      value: production.com
      proxied: true
      ttl: 1
    - name: api
      type: A
      value: "198.51.100.20"
      proxied: true
      ttl: 1
      comment: "API endpoint"
    - name: "@"
      type: MX
      value: aspmx.l.google.com
      priority: 1
      ttl: 3600
      comment: "Google Workspace primary"
    - name: "@"
      type: MX
      value: alt1.aspmx.l.google.com
      priority: 5
      ttl: 3600
      comment: "Google Workspace secondary"
    - name: "@"
      type: TXT
      value: "v=spf1 include:_spf.google.com ~all"
      ttl: 3600
      comment: "SPF record for Google Workspace"
    - name: "_dmarc"
      type: TXT
      value: "v=DMARC1; p=quarantine; rua=mailto:dmarc@production.com"
      ttl: 3600
      comment: "DMARC policy"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | `string` | The Cloudflare Zone ID of the created DNS zone |
| `nameservers` | `string[]` | The nameserver addresses assigned to the zone. Update your domain registrar to use these values to activate the zone. |

## Related Components

- [CloudflareDnsRecord](/docs/catalog/cloudflare/cloudflarednsrecord) — manages individual DNS records as standalone resources, useful when records are owned by different teams
- [CloudflareR2Bucket](/docs/catalog/cloudflare/cloudflarer2bucket) — references this zone via `customDomain.zoneId` for custom domain bucket access
- [CloudflareWorker](/docs/catalog/cloudflare/cloudflareworker) — commonly deployed with DNS routes pointing to Worker endpoints
- [CloudflareLoadBalancer](/docs/catalog/cloudflare/cloudflareloadbalancer) — load balances traffic across origins within the zone

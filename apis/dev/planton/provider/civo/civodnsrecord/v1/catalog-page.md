# Civo DNS Record

Manages individual DNS records within a Civo DNS zone. The component creates a single record of any supported type (A, AAAA, CNAME, MX, TXT, SRV, NS), validates the manifest, and exposes the record ID and fully qualified hostname as stack outputs.

## What Gets Created

When you deploy a CivoDnsRecord resource, Planton provisions:

- **Civo DNS Domain Record** --- a DNS record of the specified type attached to an existing Civo DNS zone, created via the `civo.DnsDomainRecord` Pulumi resource
- **Labels** --- key-value metadata derived from `metadata.labels` applied to internal tracking (resource name, kind, organization, environment)

## Prerequisites

- **Civo credentials** configured via environment variables or Planton provider config
- **An existing Civo DNS Zone** --- either supply the zone ID directly or reference a CivoDnsZone resource using `valueFrom`
- **A valid record value** appropriate for the chosen record type (e.g., an IPv4 address for A records, a hostname for CNAME records)

## Quick Start

Create a file `civo-dns-record.yaml`:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsRecord
metadata:
  name: www-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.CivoDnsRecord.www-record
spec:
  zoneId:
    value: "a1b2c3d4-zone-id"
  name: www
  type: A
  value: "203.0.113.10"
```

Deploy:

```shell
planton apply -f civo-dns-record.yaml
```

This creates an A record for `www` in the specified DNS zone pointing to `203.0.113.10` with a default TTL of 3600 seconds (1 hour).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zoneId` | `StringValueOrRef` | The ID of the Civo DNS zone where the record will be created. Accepts a literal `value` string or a `valueFrom` reference to a CivoDnsZone resource. | Required |
| `name` | `string` | The DNS record name (e.g., `"www"`, `"api"`, `"@"` for the zone apex). | Required |
| `type` | `RecordType` | The DNS record type. One of: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `NS`. | Required; must not be `record_type_unspecified` |
| `value` | `string` | The record value. Format depends on the record type (see notes below). | Required |

**Value format by record type:**

| Type | Expected value format | Example |
|------|-----------------------|---------|
| `A` | IPv4 address | `"192.0.2.1"` |
| `AAAA` | IPv6 address | `"2001:db8::1"` |
| `CNAME` | Target hostname | `"example.com"` |
| `MX` | Mail server hostname | `"mail.example.com"` |
| `TXT` | Arbitrary text | `"v=spf1 include:_spf.google.com ~all"` |
| `SRV` | Priority weight port target | `"10 60 5060 sip.example.com"` |
| `NS` | Nameserver hostname | `"ns1.example.com"` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttl` | `int32` | `3600` | Time to live in seconds. Determines how long resolvers cache the record. Valid range: 60--86400 (1 minute to 24 hours). A value of `0` uses the default (3600). |
| `priority` | `int32` | `0` | Priority for MX and SRV records. Lower values indicate higher priority. Range: 0--65535. Required when `type` is `MX`. |

### Cross-Field Validation

- When `type` is `MX`, the `priority` field must be greater than 0.

### Zone ID Reference Syntax

The `zoneId` field supports two forms:

**Literal value** --- supply the zone ID directly:

```yaml
zoneId:
  value: "a1b2c3d4-zone-id"
```

**Resource reference** --- resolve the zone ID from a CivoDnsZone resource at deploy time. The default kind is `CivoDnsZone` and the default field path is `status.outputs.zone_id`, so only the `name` is required:

```yaml
zoneId:
  valueFrom:
    name: my-zone
```

## Examples

### A Record with Custom TTL

An A record pointing a subdomain to an IPv4 address with a 5-minute TTL:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsRecord
metadata:
  name: api-a-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CivoDnsRecord.api-a-record
spec:
  zoneId:
    valueFrom:
      name: my-zone
  name: api
  type: A
  value: "203.0.113.50"
  ttl: 300
```

### MX Record with Priority

A mail exchange record directing email to a mail server with priority 10:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsRecord
metadata:
  name: mail-mx-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CivoDnsRecord.mail-mx-record
spec:
  zoneId:
    value: "a1b2c3d4-zone-id"
  name: "@"
  type: MX
  value: "mail.example.com"
  ttl: 3600
  priority: 10
```

### CNAME Record Referencing a DNS Zone

A CNAME alias that resolves the zone ID from a CivoDnsZone resource:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoDnsRecord
metadata:
  name: docs-cname
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.CivoDnsRecord.docs-cname
spec:
  zoneId:
    valueFrom:
      name: my-zone
  name: docs
  type: CNAME
  value: "docs-hosting.example.com"
  ttl: 600
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `recordId` | `string` | Unique identifier of the DNS record, assigned by Civo |
| `hostname` | `string` | The fully qualified hostname of the DNS record (e.g., `"www.example.com"`) |
| `recordType` | `string` | The DNS record type that was created (e.g., `"A"`, `"CNAME"`) |
| `accountId` | `string` | The Civo account ID where the record was created |

## Related Components

- [CivoDnsZone](/docs/catalog/civo/civodnszone) --- the parent DNS zone where records are created; can be referenced via `valueFrom` in the `zoneId` field
- [CivoKubernetesCluster](/docs/catalog/civo/civokubernetescluster) --- Kubernetes clusters whose ingress addresses are common targets for A and CNAME records
- [CivoIpAddress](/docs/catalog/civo/civoipaddress) --- reserved IP addresses that can be used as A record values for stable endpoints

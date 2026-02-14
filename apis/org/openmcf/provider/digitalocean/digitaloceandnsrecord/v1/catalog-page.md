# DigitalOcean DNS Record

Creates a single DNS record within an existing DigitalOcean DNS zone (domain). The component supports A, AAAA, CNAME, MX, TXT, SRV, NS, and CAA record types, with type-specific fields for priority, weight, port, flags, and tag applied conditionally based on the record type.

## What Gets Created

When you deploy a DigitalOceanDnsRecord resource, OpenMCF provisions:

- **DNS Record** — a single `digitalocean_record` resource in the specified domain with the configured type, name, value, and TTL
- **Type-Specific Attributes** — `priority` is set for MX and SRV records; `weight` and `port` are set for SRV records; `flags` and `tag` are set for CAA records. These attributes are omitted for inapplicable record types.

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or OpenMCF provider config
- **An existing DigitalOcean DNS zone (domain)** managed by DigitalOcean's DNS service. The `domain` field can reference a DigitalOceanDnsZone resource via `valueFrom`.

## Quick Start

Create a file `dns-record.yaml`:

```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: www-a-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanDnsRecord.www-a-record
spec:
  domain:
    value: "example.com"
  name: "www"
  type: A
  value:
    value: "192.0.2.1"
  ttlSeconds: 3600
```

Deploy:

```shell
openmcf apply -f dns-record.yaml
```

This creates an A record pointing `www.example.com` to `192.0.2.1` with a one-hour TTL.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `domain` | `StringValueOrRef` | The DigitalOcean domain name (DNS zone) where the record will be created. Can reference a DigitalOceanDnsZone resource via `valueFrom` (default field path: `status.outputs.zone_name`). | Required |
| `name` | `string` | Hostname or subdomain for the record. Use `@` for root domain records, or specify the subdomain (e.g., `www`, `api.v1`). | Required |
| `type` | `enum` | DNS record type. Valid values: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `NS`, `CAA`. | Required, must not be `record_type_unspecified` |
| `value` | `StringValueOrRef` | The record value. Format depends on type: IPv4 address for A, IPv6 for AAAA, target hostname for CNAME/MX/NS/SRV, text string for TXT, CA domain for CAA. Can reference another resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttlSeconds` | `int32` | `1800` | Time to live in seconds. Determines how long resolvers cache the record. Range: 30--86400. |
| `priority` | `int32` | `0` | Priority for MX and SRV records. Lower values indicate higher priority. Range: 0--65535. Ignored for other record types. |
| `weight` | `int32` | `0` | Relative weight for SRV records with the same priority. Higher values receive proportionally more traffic. Range: 0--65535. Ignored for non-SRV types. |
| `port` | `int32` | — | TCP or UDP port for SRV records. Range: 0--65535. Required when `type` is `SRV`. Ignored for other types. |
| `flags` | `int32` | `0` | Flags for CAA records. `0` = non-critical (CA may ignore unknown tags), `128` = critical (CA must refuse if tag is unknown). Range: 0--255. Ignored for non-CAA types. |
| `tag` | `string` | — | Property tag for CAA records: `issue` (authorize CA), `issuewild` (authorize wildcard), or `iodef` (violation reporting URL). Required when `type` is `CAA`. Ignored for other types. |

### Cross-Field Validation

The protobuf schema enforces two cross-field rules:

- `port` must be greater than 0 when `type` is `SRV`.
- `tag` must be non-empty when `type` is `CAA`.

## Examples

### A Record

Points a subdomain to an IPv4 address:

```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: www-a-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanDnsRecord.www-a-record
spec:
  domain:
    value: "example.com"
  name: "www"
  type: A
  value:
    value: "192.0.2.1"
  ttlSeconds: 3600
```

### CNAME Record

Creates an alias from one hostname to another:

```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: blog-cname
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanDnsRecord.blog-cname
spec:
  domain:
    value: "example.com"
  name: "blog"
  type: CNAME
  value:
    value: "www.example.com."
  ttlSeconds: 3600
```

### MX Record with Priority

Routes email to a mail server with explicit priority:

```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: mail-mx
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanDnsRecord.mail-mx
spec:
  domain:
    value: "example.com"
  name: "@"
  type: MX
  value:
    value: "mail.example.com."
  ttlSeconds: 3600
  priority: 10
```

### CAA Record

Restricts which certificate authorities may issue certificates for the domain:

```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: caa-letsencrypt
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanDnsRecord.caa-letsencrypt
spec:
  domain:
    value: "example.com"
  name: "@"
  type: CAA
  value:
    value: "letsencrypt.org"
  ttlSeconds: 3600
  flags: 0
  tag: "issue"
```

### Domain Reference with valueFrom

Uses a `valueFrom` reference to resolve the domain from a DigitalOceanDnsZone resource instead of specifying it inline:

```yaml
apiVersion: digitalocean.openmcf.org/v1
kind: DigitalOceanDnsRecord
metadata:
  name: api-a-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanDnsRecord.api-a-record
spec:
  domain:
    valueFrom:
      kind: DigitalOceanDnsZone
      name: prod-zone
      fieldPath: status.outputs.zone_name
  name: "api"
  type: A
  value:
    value: "198.51.100.10"
  ttlSeconds: 300
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `record_id` | `string` | Unique identifier of the created DNS record in DigitalOcean |
| `hostname` | `string` | Fully qualified hostname (e.g., `www.example.com` or `example.com` for root records) |
| `record_type` | `string` | DNS record type that was created (`A`, `AAAA`, `CNAME`, etc.) |
| `domain` | `string` | Domain name (DNS zone) where the record was created |
| `ttl_seconds` | `int32` | TTL in seconds applied to the record |

## Related Components

- [DigitalOceanDnsZone](/docs/catalog/digitalocean/digitaloceandnszone) -- provides the domain (DNS zone) in which records are created
- [DigitalOceanDroplet](/docs/catalog/digitalocean/digitaloceandroplet) -- provisions Droplets whose IP addresses can be used as A/AAAA record values
- [DigitalOceanLoadBalancer](/docs/catalog/digitalocean/digitaloceanloadbalancer) -- provisions load balancers whose IPs or hostnames can serve as record targets

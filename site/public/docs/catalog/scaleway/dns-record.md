---
title: "DNS Record"
description: "DNS Record deployment documentation"
icon: "package"
order: 100
componentName: "scalewaydnsrecord"
---

# Scaleway DNS Record

Deploys a standalone DNS record in a Scaleway DNS zone. Designed as a DAG-friendly alternative to inline ScalewayDnsZone records, this component enables explicit dependency edges when record values come from other infrastructure resources such as Load Balancer IPs or Kapsule cluster endpoints.

## What Gets Created

When you deploy a ScalewayDnsRecord resource, OpenMCF provisions:

- **DNS Record** — a single `domain.Record` resource in the specified Scaleway DNS zone with the configured name, type, data, TTL, and optional priority

No tags are applied to the record because the Scaleway DNS API does not support resource tags. The record FQDN and `metadata.name` serve as the primary identifiers.

## Prerequisites

- **Scaleway credentials** configured via environment variables or OpenMCF provider config
- **An existing DNS zone** in Scaleway where the record will be created (can be provisioned via a ScalewayDnsZone resource or managed externally)
- **Record data** available as a literal value or as a stack output from another OpenMCF resource (e.g., a Load Balancer IP, an Instance public IP)

## Quick Start

Create a file `dns-record.yaml`:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: www-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayDnsRecord.www-record
spec:
  zoneName:
    value: example.com
  name: www
  type: A
  data:
    value: "192.0.2.1"
  ttl: 3600
```

Deploy:

```shell
openmcf apply -f dns-record.yaml
```

This creates an A record `www.example.com` pointing to `192.0.2.1` with a 1-hour TTL.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zoneName` | `StringValueOrRef` | DNS zone where the record is created (e.g., `"example.com"`). Can be a direct value or a reference to a ScalewayDnsZone output. The zone must already exist. | Required |
| `type` | `RecordType` | DNS record type. One of: `A`, `AAAA`, `ALIAS`, `CAA`, `CNAME`, `DNAME`, `MX`, `NS`, `PTR`, `SOA`, `SRV`, `TXT`, `TLSA`. Cannot be changed after creation. | Required, must not be `record_type_unspecified` |
| `data` | `StringValueOrRef` | Record value/data. Can be a literal string or a reference to another resource's output. Format depends on the record type (see type documentation below). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | `""` (zone apex) | Record name relative to the zone. Use empty string or omit for the zone apex. Cannot be changed after creation. Examples: `"www"`, `"api"`, `"_dmarc"`. |
| `ttl` | `uint32` | `3600` | Time to live in seconds. Common values: `300` (migrations), `3600` (standard), `86400` (static records). |
| `priority` | `uint32` | `0` | Priority for MX and SRV records. Lower values indicate higher priority. Ignored for other record types. |
| `keepEmptyZone` | `bool` | `true` (recommended) | When `true`, prevents the DNS zone from being deleted if this is the last record destroyed. Recommended when zones are managed by separate ScalewayDnsZone resources. |

**Data format by record type:**

| Type | Data Format | Example |
|------|-------------|---------|
| `A` | IPv4 address | `"192.0.2.1"` |
| `AAAA` | IPv6 address | `"2001:db8::1"` |
| `ALIAS` | Target hostname | `"www.example.com."` |
| `CAA` | `flags tag value` | `"0 issue \"letsencrypt.org\""` |
| `CNAME` | Target with trailing dot | `"target.example.com."` |
| `DNAME` | Delegation target | `"other.example.com."` |
| `MX` | Mail server with trailing dot | `"mail.example.com."` |
| `NS` | Nameserver with trailing dot | `"ns1.example.com."` |
| `PTR` | Pointer target | `"host.example.com."` |
| `SRV` | `weight port target` | `"10 5060 sipserver.example.com."` |
| `TXT` | Text data | `"v=spf1 include:_spf.google.com ~all"` |
| `TLSA` | `usage selector matching-type cert-data` | `"3 1 1 abcdef..."` |

## Examples

### Simple A Record

A basic A record pointing a subdomain to a static IP address:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: api-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayDnsRecord.api-record
spec:
  zoneName:
    value: example.com
  name: api
  type: A
  data:
    value: "203.0.113.50"
  ttl: 3600
```

### MX Records for Email

Two MX records for a domain with primary and backup mail servers. Each record is a separate ScalewayDnsRecord resource:

Primary mail server (`mx-primary.yaml`):

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: mx-primary
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayDnsRecord.mx-primary
spec:
  zoneName:
    value: example.com
  name: ""
  type: MX
  data:
    value: "mail.example.com."
  ttl: 3600
  priority: 1
  keepEmptyZone: true
```

Backup mail server (`mx-backup.yaml`):

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: mx-backup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayDnsRecord.mx-backup
spec:
  zoneName:
    value: example.com
  name: ""
  type: MX
  data:
    value: "mail-backup.example.com."
  ttl: 3600
  priority: 10
  keepEmptyZone: true
```

### Cross-Resource Wiring with References

An A record whose value is dynamically resolved from a ScalewayLoadBalancer output, and whose zone comes from a ScalewayDnsZone resource. This is the primary use case for standalone DNS records -- creating explicit dependency edges in infra charts:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: lb-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayDnsRecord.lb-record
spec:
  zoneName:
    valueFrom:
      kind: ScalewayDnsZone
      name: prod-zone
      fieldPath: status.outputs.zone_name
  name: app
  type: A
  data:
    valueFrom:
      kind: ScalewayLoadBalancer
      name: prod-lb
      fieldPath: status.outputs.ip_address
  ttl: 300
  keepEmptyZone: true
```

A CNAME record pointing to a Kapsule cluster endpoint:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: k8s-ingress
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayDnsRecord.k8s-ingress
spec:
  zoneName:
    value: example.com
  name: ingress
  type: CNAME
  data:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: prod-cluster
      fieldPath: status.outputs.apiserver_url
  ttl: 300
  keepEmptyZone: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `recordId` | `string` | The unique identifier of the created DNS record in Scaleway. Format: `"{dns_zone}/{record_id}"`. Used for API operations and Terraform import. |
| `fqdn` | `string` | The fully qualified domain name of the record, computed by Scaleway from the record name and zone name (e.g., `"www.example.com"`, `"example.com"`). Primary downstream output for other resources that need the full record name. |

## Related Components

- [ScalewayDnsZone](/docs/catalog/scaleway/dns-zone) — manages DNS zones where records are created; provides `zone_name` output used by `zoneName.valueFrom`
- [ScalewayLoadBalancer](/docs/catalog/scaleway/load-balancer) — deploys load balancers whose IP addresses are commonly wired into A records via `data.valueFrom`
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/kapsule-cluster) — deploys Kubernetes clusters whose endpoints can be referenced in CNAME or A records
- [ScalewayInstance](/docs/catalog/scaleway/instance) — deploys compute instances whose public IPs can be targeted by A/AAAA records

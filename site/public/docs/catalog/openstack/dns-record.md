---
title: "DNS Record"
description: "DNS Record deployment documentation"
icon: "package"
order: 100
componentName: "openstackdnsrecord"
---

# OpenStack DNS Record

Deploys a standalone DNS recordset in an OpenStack Designate zone with a configurable name, type, values, and TTL. Use this component when individual DNS records need to be independently managed or wired as explicit dependencies in InfraCharts; for managing all records within a single component, use the inline records feature of OpenStackDnsZone instead.

## What Gets Created

When you deploy an OpenStackDnsRecord resource, OpenMCF provisions:

- **DNS RecordSet** — an `openstack_dns_recordset_v2` resource in the specified Designate zone. The recordset contains one fully qualified domain name, one record type, and one or more values. Multiple values in the same recordset produce a round-robin record set.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **An existing Designate DNS zone** — provide the zone UUID directly or reference an OpenStackDnsZone resource via `valueFrom`
- **Designate DNS service enabled** in the target OpenStack project

## Quick Start

Create a file `dns-record.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsRecord
metadata:
  name: my-a-record
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackDnsRecord.my-a-record
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnsrecord/v1/iac/pulumi/module
spec:
  zoneId: 5a6b7c8d-9e0f-1a2b-3c4d-5e6f7a8b9c0d
  recordName: "www.example.com."
  type: A
  values:
    - "192.0.2.1"
```

Deploy:

```shell
openmcf apply -f dns-record.yaml
```

This creates an A record for `www.example.com.` pointing to `192.0.2.1` in the specified Designate zone.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zoneId` | `StringValueOrRef` | The ID (UUID) of the Designate zone where this record will be created. Can reference an OpenStackDnsZone resource via `valueFrom`. ForceNew: changing this recreates the record. | Required |
| `recordName` | `string` | Fully qualified domain name for this record. Must end with a trailing dot (e.g., `www.example.com.`). ForceNew: changing this recreates the record. | Must be a valid DNS name ending with `.` |
| `type` | `RecordType` | DNS record type. Supported values: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `NS`, `PTR`, `CAA`, `SOA`, `SPF`, `SSHFP`, `NAPTR`. ForceNew: changing this recreates the record. | Must be a defined enum value; cannot be unspecified |
| `values` | `string[]` | DNS record values. Format depends on record type (see examples below). Multiple values create a round-robin record set. | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttl` | `int32` | zone default | Time To Live in seconds. Determines how long resolvers cache this record. If omitted, the zone's default TTL is used. |
| `description` | `string` | — | Human-readable description of the DNS record. |
| `region` | `string` | provider default | Override the region from the provider config for this record. ForceNew: changing this recreates the record. |

## Examples

### A Record with Custom TTL

A single IPv4 address record with a 5-minute TTL:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsRecord
metadata:
  name: web-a-record
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackDnsRecord.web-a-record
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnsrecord/v1/iac/pulumi/module
spec:
  zoneId: 5a6b7c8d-9e0f-1a2b-3c4d-5e6f7a8b9c0d
  recordName: "web.example.com."
  type: A
  values:
    - "192.0.2.10"
    - "192.0.2.11"
  ttl: 300
  description: "Round-robin A records for the web tier"
```

### CNAME Record

An alias record pointing a subdomain to another hostname:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsRecord
metadata:
  name: docs-cname
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackDnsRecord.docs-cname
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnsrecord/v1/iac/pulumi/module
spec:
  zoneId: 5a6b7c8d-9e0f-1a2b-3c4d-5e6f7a8b9c0d
  recordName: "docs.example.com."
  type: CNAME
  values:
    - "lb.example.com."
  ttl: 3600
```

### MX Records for Mail Delivery

Mail exchange records with priority values for primary and backup mail servers:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsRecord
metadata:
  name: mail-mx
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackDnsRecord.mail-mx
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnsrecord/v1/iac/pulumi/module
spec:
  zoneId: 5a6b7c8d-9e0f-1a2b-3c4d-5e6f7a8b9c0d
  recordName: "example.com."
  type: MX
  values:
    - "10 mail1.example.com."
    - "20 mail2.example.com."
  ttl: 3600
  description: "Primary and backup mail servers"
```

### Using Foreign Key References

Reference an OpenMCF-managed DNS zone instead of hardcoding the zone UUID:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackDnsRecord
metadata:
  name: api-record
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackDnsRecord.api-record
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackdnsrecord/v1/iac/pulumi/module
spec:
  zoneId:
    valueFrom:
      kind: OpenStackDnsZone
      name: my-zone
      field: status.outputs.zone_id
  recordName: "api.example.com."
  type: A
  values:
    - "10.0.1.100"
  ttl: 60
  description: "API endpoint for the example.com zone"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `recordset_id` | `string` | UUID of the created DNS recordset |
| `fqdn` | `string` | Fully qualified domain name of the created DNS record (e.g., `www.example.com.`) |
| `record_type` | `string` | DNS record type that was created (e.g., `A`, `CNAME`, `MX`) |
| `values` | `string[]` | List of DNS record values that were set |
| `zone_id` | `string` | ID of the Designate zone containing this record |
| `region` | `string` | OpenStack region where the record was created |

## Related Components

- [OpenStackDnsZone](/docs/catalog/openstack/dns-zone) — provides the Designate zone where records are created; also supports inline records for simpler setups

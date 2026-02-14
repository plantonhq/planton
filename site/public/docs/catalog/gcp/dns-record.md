---
title: "DNS Record"
description: "DNS Record deployment documentation"
icon: "package"
order: 100
componentName: "gcpdnsrecord"
---

# GCP DNS Record

Deploys an individual DNS record set within an existing Google Cloud DNS Managed Zone. This component supports all standard record types (A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA, SOA), configurable TTL, and round-robin record sets with multiple values.

## What Gets Created

When you deploy a GcpDnsRecord resource, OpenMCF provisions:

- **DNS Record Set** — a `google_dns_record_set` resource in the specified managed zone, with the given type, FQDN, values, and TTL

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **An existing GCP project** — referenced via `projectId`
- **An existing Cloud DNS Managed Zone** — referenced via `managedZone`, either by direct name or as a foreign key to a GcpDnsZone resource
- **IAM permissions** to create and manage DNS record sets in the target managed zone

## Quick Start

Create a file `dns-record.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsRecord
metadata:
  name: app-a-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpDnsRecord.app-a-record
spec:
  projectId: my-gcp-project-123
  managedZone: example-zone
  type: A
  name: app.example.com.
  values:
    - 203.0.113.10
```

Deploy:

```shell
openmcf apply -f dns-record.yaml
```

This creates an A record for `app.example.com.` pointing to `203.0.113.10` with the default TTL of 300 seconds.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID where the managed zone exists. Can reference a GcpProject resource via `valueFrom`. | Required |
| `managedZone` | `StringValueOrRef` | Name of the Cloud DNS Managed Zone where the record is created. Can reference a GcpDnsZone resource via `valueFrom`. | Required |
| `type` | `RecordType` | DNS record type. One of: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `NS`, `PTR`, `CAA`, `SOA`. | Required, must be a defined enum value |
| `name` | `string` | Fully qualified domain name for the record. Must end with a trailing dot (e.g., `www.example.com.`). | Required, must match valid FQDN pattern |
| `values` | `string[]` | Record values. For A records: IPv4 addresses. For AAAA: IPv6 addresses. For CNAME: target hostname with trailing dot. Multiple values create a round-robin record set. | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttlSeconds` | `int32` | `300` | Time to live for the DNS record in seconds. Determines how long resolvers cache this record. Valid range: 1-86400. Common values: 60 (1 min), 300 (5 min), 3600 (1 hour), 86400 (1 day). |

## Examples

### Simple A Record

An A record pointing a subdomain to a single IP address:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsRecord
metadata:
  name: web-a-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpDnsRecord.web-a-record
spec:
  projectId: my-gcp-project-123
  managedZone: example-zone
  type: A
  name: www.example.com.
  values:
    - 203.0.113.10
  ttlSeconds: 300
```

### CNAME Record with Foreign Key References

A CNAME record that references OpenMCF-managed GcpProject and GcpDnsZone resources instead of hardcoding identifiers:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsRecord
metadata:
  name: docs-cname
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpDnsRecord.docs-cname
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  managedZone:
    valueFrom:
      kind: GcpDnsZone
      name: example.com
      fieldPath: status.outputs.zone_name
  type: CNAME
  name: docs.example.com.
  values:
    - example.github.io.
  ttlSeconds: 3600
```

### Round-Robin A Record with Multiple IPs

An A record with multiple values for basic load distribution across servers:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsRecord
metadata:
  name: api-round-robin
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpDnsRecord.api-round-robin
spec:
  projectId: my-prod-project-456
  managedZone: example-zone
  type: A
  name: api.example.com.
  values:
    - 203.0.113.10
    - 203.0.113.11
    - 203.0.113.12
  ttlSeconds: 60
```

### MX Record for Email Routing

An MX record configuring mail delivery with primary and backup mail servers:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsRecord
metadata:
  name: mail-mx
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpDnsRecord.mail-mx
spec:
  projectId: my-prod-project-456
  managedZone: example-zone
  type: MX
  name: example.com.
  values:
    - "10 mail.example.com."
    - "20 mail2.example.com."
  ttlSeconds: 3600
```

### TXT Record for SPF and Domain Verification

A TXT record used for email sender policy and domain ownership verification:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsRecord
metadata:
  name: spf-txt
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpDnsRecord.spf-txt
spec:
  projectId: my-prod-project-456
  managedZone: example-zone
  type: TXT
  name: example.com.
  values:
    - "v=spf1 include:_spf.google.com ~all"
  ttlSeconds: 3600
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `fqdn` | `string` | The fully qualified domain name of the created DNS record (e.g., `www.example.com.`) |
| `record_type` | `string` | The DNS record type that was created (e.g., `A`, `CNAME`, `TXT`) |
| `managed_zone` | `string` | The name of the managed zone containing this record |
| `project_id` | `string` | The GCP project ID where the record was created |
| `ttl_seconds` | `int32` | The TTL (time to live) in seconds for the DNS record |

## Related Components

- [GcpDnsZone](/docs/catalog/gcp/gcpdnszone) — creates the Cloud DNS Managed Zone where records are hosted
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project referenced by `projectId`
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — creates service accounts that can be granted DNS management permissions
- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — deploys GKE clusters whose ingress endpoints are commonly referenced by A or CNAME records

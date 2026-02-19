# Alibaba Cloud DNS Record

Creates and manages DNS records within an Alibaba Cloud Alidns-hosted domain. Supports all standard record types (A, AAAA, CNAME, MX, TXT, NS, SRV, CAA) with configurable TTL, priority, resolution lines, and record status.

## What Gets Created

When you deploy an AlicloudDnsRecord resource, OpenMCF provisions:

- **Alidns Record** -- an `alicloud_alidns_record` resource (Pulumi: `dns.AlidnsRecord`) that creates a DNS record within the specified parent domain

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or OpenMCF provider config
- **Parent domain** registered in Alidns -- either via the AlicloudDnsDomain component or manually in the console
- **OpenMCF CLI** installed with either Pulumi or Terraform (OpenTofu) backend

## Quick Start

Create a file `dns-record.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsRecord
metadata:
  name: my-record
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudDnsRecord.my-record
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: www
  type: A
  value: "203.0.113.10"
```

Deploy:

```shell
openmcf apply -f dns-record.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for provider initialization (e.g., `cn-hangzhou`). Alidns is global, but the provider requires a region. | Required; non-empty |
| `domainName` | `string` | The parent domain name (e.g., `example.com`). Must already exist in Alidns. Cannot be changed after creation. | Required; 1-253 characters |
| `rr` | `string` | Host record (subdomain part). `@` for apex, `*` for wildcard, or any valid subdomain label. | Required; 1-253 characters |
| `type` | `string` | DNS record type. | Required; one of `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV`, `CAA`, `REDIRECT_URL`, `FORWORD_URL` |
| `value` | `string` | Record value. Interpretation depends on `type` (IP for A/AAAA, domain for CNAME/MX/NS, text for TXT). | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttl` | `int32` | `600` | Time-to-live in seconds. Range depends on the Alidns plan (Free: 600-86400). |
| `priority` | `int32` | - | MX record priority, range 1 (highest) to 10 (lowest). Required when `type` is `MX`, ignored for other types. |
| `line` | `string` | `"default"` | DNS resolution line for ISP/geo-based routing. Use `"default"` for standard resolution. Must be `"default"` when `type` is `FORWORD_URL`. |
| `status` | `string` | `"ENABLE"` | Record status: `ENABLE` (active) or `DISABLE` (record exists but is not served). |
| `remark` | `string` | `""` | Description or notes for the record. Visible in the Alidns console. |

## Examples

### A Record

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsRecord
metadata:
  name: web-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudDnsRecord.web-server
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: www
  type: A
  value: "203.0.113.10"
  ttl: 600
```

### CNAME Record

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsRecord
metadata:
  name: cdn-alias
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudDnsRecord.cdn-alias
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: cdn
  type: CNAME
  value: example.com.cdn-provider.com
```

### MX Record with Priority

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsRecord
metadata:
  name: mail-primary
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AlicloudDnsRecord.mail-primary
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: "@"
  type: MX
  value: mx1.example.com
  priority: 5
  ttl: 3600
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `record_id` | `string` | The record ID assigned by Alibaba Cloud. |

## Related Components

- [AlicloudDnsDomain](/docs/catalog/alicloud/aliclouddnsdomain) -- registers the parent domain in Alidns (prerequisite for creating records)
- [AlicloudPrivateZone](/docs/catalog/alicloud/alicloudprivatezone) -- manages private DNS zones for VPC-internal resolution (separate from public Alidns)

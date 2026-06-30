---
title: "Private Zone"
description: "Private Zone deployment documentation"
icon: "package"
order: 100
componentName: "alicloudprivatednszone"
---

# AliCloud Private Zone

Provisions and manages an Alibaba Cloud Private Zone (PVTZ) for VPC-internal DNS resolution, with automated VPC attachment and inline DNS record management. Private Zone records are only resolvable within attached VPCs -- they are invisible to the public internet.

## What Gets Created

When you deploy an AliCloudPrivateDnsZone resource, Planton provisions:

- **Private Zone** -- an `alicloud_pvtz_zone` resource (Pulumi: `pvtz.Zone`) that creates the private DNS hosted zone
- **VPC Attachment** -- an `alicloud_pvtz_zone_attachment` resource (Pulumi: `pvtz.ZoneAttachment`) that binds the zone to one or more VPCs, enabling DNS resolution within those VPCs. Cross-region attachments are supported.
- **Zone Records** -- `alicloud_pvtz_zone_record` resources (Pulumi: `pvtz.ZoneRecord`) for each record defined in `spec.records`. Supported types: A, CNAME, MX, PTR, SRV, TXT.
- **Tags** -- system metadata tags (`resource`, `resource_name`, `resource_kind`, `organization`, `environment`) merged with user-defined `spec.tags`, with user values taking precedence on key conflict

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or Planton provider config
- **At least one VPC** to attach the zone to -- the zone is useless without a VPC attachment since records are only resolvable within attached VPCs
- **Planton CLI** installed with either Pulumi or Terraform (OpenTofu) backend

## Quick Start

Create a file `private-zone.yaml`:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudPrivateDnsZone
metadata:
  name: my-private-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AliCloudPrivateDnsZone.my-private-zone
spec:
  region: cn-hangzhou
  zoneName: internal.example.com
  vpcAttachments:
    - vpcId: vpc-abc123
  records:
    - rr: api
      type: A
      value: "10.0.1.50"
```

Deploy:

```shell
planton apply -f private-zone.yaml
```

After deployment, resources within the attached VPC can resolve `api.internal.example.com` to `10.0.1.50`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for provider initialization. Private Zone is a global service, but the provider requires a region. | Required; non-empty |
| `zoneName` | `string` | The private zone name (e.g., `internal.example.com`). This is the DNS suffix for all records in the zone. Cannot be changed after creation. | Required; 1-253 characters |
| `vpcAttachments` | `list` | VPCs to attach this zone to. At least one required. | Required; min 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `remark` | `string` | `""` | Description for the zone. Visible in the Private Zone console. |
| `resourceGroupId` | `string` | `""` | Resource group for access control and cost attribution. Cannot be changed after creation. |
| `records` | `list` | `[]` | DNS records within the zone. See record fields below. |
| `tags` | `map<string, string>` | `{}` | User-defined tags. Merged with system tags; user values win on conflict. |

### VPC Attachment Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `vpcId` | `StringValueOrRef` | -- | VPC ID to attach. Supports cross-component references to AliCloudVpc. |
| `regionId` | `string` | `""` | Region of the VPC. Defaults to `spec.region`. Set this for cross-region attachment. |

### Record Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `rr` | `string` | -- | Resource record name (e.g., `db`, `api`, `@` for zone apex). |
| `type` | `string` | -- | Record type: `A`, `CNAME`, `MX`, `PTR`, `SRV`, `TXT`. |
| `value` | `string` | -- | Record value (IP address, hostname, text content). |
| `ttl` | `int32` | `60` | Time-to-live in seconds. |
| `priority` | `int32` | `1` | Priority for MX records only (1-99). Ignored for other types. |
| `remark` | `string` | `""` | Description for the record. |

## Examples

### Internal Service Discovery

A common pattern: create a private zone for service discovery within a VPC, with A records for each service endpoint.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudPrivateDnsZone
metadata:
  name: svc-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AliCloudPrivateDnsZone.svc-zone
spec:
  region: cn-hangzhou
  zoneName: svc.internal
  vpcAttachments:
    - vpcId: vpc-app-prod
  records:
    - rr: api
      type: A
      value: "10.0.1.50"
    - rr: cache
      type: A
      value: "10.0.2.30"
    - rr: queue
      type: A
      value: "10.0.3.10"
```

### Multi-VPC Database Zone

Share database endpoints across multiple VPCs, including cross-region.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudPrivateDnsZone
metadata:
  name: db-zone
  org: my-org
  env: production
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AliCloudPrivateDnsZone.db-zone
spec:
  region: cn-hangzhou
  zoneName: db.corp
  resourceGroupId: rg-prod-123
  vpcAttachments:
    - vpcId: vpc-app-hangzhou
    - vpcId: vpc-app-shanghai
      regionId: cn-shanghai
  records:
    - rr: mysql
      type: A
      value: "10.0.10.100"
    - rr: redis
      type: A
      value: "10.0.11.50"
  tags:
    team: dba
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | `string` | The Private Zone ID assigned by Alibaba Cloud. |
| `zone_name` | `string` | The zone name as created. |
| `is_ptr` | `bool` | Whether the zone is a reverse-lookup (PTR) zone. Computed from the zone name format. |
| `record_count` | `int32` | The number of DNS records in the zone. |

## Related Components

- [AliCloudVpc](/docs/catalog/alicloud/vpc) -- VPCs that this private zone attaches to for DNS resolution
- [AliCloudDnsZone](/docs/catalog/alicloud/dns-domain) -- manages public DNS domains in Alidns (separate from private zones)
- [AliCloudDnsRecord](/docs/catalog/alicloud/dns-record) -- creates public DNS records within an Alidns domain

# AlicloudPrivateZone

Manages an Alibaba Cloud Private Zone (PVTZ) for VPC-internal DNS resolution.

## Overview

A Private Zone is a private DNS hosted zone that resolves domain names within one or more VPCs. Unlike public Alidns domains, Private Zone records are only visible to resources inside the attached VPCs -- they are never served to the public internet.

This component bundles three provider resources into a single deployable unit:

- **Private Zone** -- the hosted zone (`alicloud_pvtz_zone`)
- **VPC Attachment** -- binds the zone to one or more VPCs (`alicloud_pvtz_zone_attachment`)
- **Zone Records** -- DNS records within the zone (`alicloud_pvtz_zone_record`)

At least one VPC attachment is required. Without it, the zone has no resolver scope and records cannot be queried by any VPC resource.

### What Gets Created

- An Alibaba Cloud Private Zone with the specified zone name
- VPC attachment(s) that enable DNS resolution within attached VPCs (supports cross-region)
- Optional DNS records (A, CNAME, MX, PTR, SRV, TXT) within the zone
- System metadata tags merged with user-defined tags

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region for provider initialization (e.g., `cn-hangzhou`) |
| `zoneName` | string | The private zone name (e.g., `internal.example.com`). Cannot be changed after creation. |
| `vpcAttachments` | list | VPCs to attach. At least one required. Each entry has `vpcId` (required) and `regionId` (optional, for cross-region). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `remark` | string | `""` | Description for the zone |
| `resourceGroupId` | string | `""` | Resource group for access control. Cannot be changed after creation. |
| `records` | list | `[]` | DNS records within the zone |
| `tags` | map | `{}` | Key-value tags applied to the zone |

### Record Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `rr` | string | -- | Resource record name (e.g., `db`, `api`, `@` for apex) |
| `type` | string | -- | Record type: `A`, `CNAME`, `MX`, `PTR`, `SRV`, `TXT` |
| `value` | string | -- | Record value (IP address, hostname, etc.) |
| `ttl` | int32 | 60 | Time-to-live in seconds |
| `priority` | int32 | 1 | Priority for MX records only (1-99) |
| `remark` | string | `""` | Description for the record |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `zone_id` | The Private Zone ID assigned by Alibaba Cloud |
| `zone_name` | The zone name as created |
| `is_ptr` | Whether the zone is a reverse-lookup (PTR) zone |
| `record_count` | The number of DNS records in the zone |

## Related Components

- **AlicloudVpc** -- VPCs that this private zone attaches to
- **AlicloudDnsDomain** -- for public DNS domains (separate from private zones)
- **AlicloudDnsRecord** -- for public DNS records (separate from private zone records)

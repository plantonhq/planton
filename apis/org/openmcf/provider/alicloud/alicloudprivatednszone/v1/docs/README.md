# AlicloudPrivateDnsZone -- Research Documentation

## Alibaba Cloud Private Zone (PVTZ) Service

Alibaba Cloud Private Zone is a VPC-based DNS resolution service. It maps private domain names to IP addresses within VPCs. Unlike the public Alidns service, Private Zone records are only resolvable by resources inside attached VPCs.

### Key Characteristics

- **VPC-scoped resolution**: Records resolve only within attached VPCs. The public internet never sees private zone records.
- **Cross-region support**: A single private zone can be attached to VPCs in different Alibaba Cloud regions.
- **Multi-VPC support**: A zone can serve multiple VPCs simultaneously. Useful for shared services (database endpoints, API gateways) that span VPC boundaries.
- **Global service**: Private Zone is not region-scoped. The zone itself is global; only the VPC attachments are region-aware.
- **Record types**: Supports A, CNAME, MX, PTR, SRV, TXT. Does NOT support AAAA (IPv6) or NS records.

### Provider Resources

#### Terraform

| Resource | Purpose |
|----------|---------|
| `alicloud_pvtz_zone` | Creates the private hosted zone |
| `alicloud_pvtz_zone_attachment` | Attaches VPCs to the zone |
| `alicloud_pvtz_zone_record` | Creates DNS records within the zone |

#### Pulumi

| Resource | Package | Purpose |
|----------|---------|---------|
| `pvtz.Zone` | `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/pvtz` | Creates the zone |
| `pvtz.ZoneAttachment` | same | Attaches VPCs |
| `pvtz.ZoneRecord` | same | Creates records |

### Zone Fields (from provider)

| Field | Type | Required | ForceNew | Notes |
|-------|------|----------|----------|-------|
| `zone_name` | string | Yes | Yes | The zone's domain name |
| `remark` | string | No | No | Description |
| `resource_group_id` | string | No | Yes | Access control grouping |
| `tags` | map | No | No | Key-value tags |
| `proxy_pattern` | string | No | No | "ZONE" (default) or "RECORD". Controls DNS proxy behavior. |
| `sync_status` | string | No | No | "ON" or "OFF". For cross-account sync. |
| `user_info` | set | No | No | Cross-account sync config. |
| `is_ptr` | bool | Computed | -- | Whether zone is PTR type |
| `record_count` | int | Computed | -- | Number of records |

### Zone Attachment Fields

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `zone_id` | string | Yes | Zone to attach VPCs to |
| `vpcs` | set | Yes | Set of {vpc_id, region_id} |
| `vpc_ids` | set | -- | Simpler alternative (same-region only). Conflicts with `vpcs`. |

The `vpcs` approach is preferred because it supports cross-region attachment. The `vpc_ids` approach only works for VPCs in the same region as the provider.

### Zone Record Fields

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `zone_id` | string | Yes | -- | ForceNew |
| `rr` | string | Yes | -- | Resource record name |
| `type` | string | Yes | -- | A, CNAME, MX, PTR, SRV, TXT |
| `value` | string | Yes | -- | Record value |
| `ttl` | int | No | 60 | Time-to-live |
| `priority` | int | No | 1 | MX only (1-99) |
| `remark` | string | No | -- | Description |
| `status` | string | No | ENABLE | ENABLE or DISABLE |

### Design Decisions

1. **Multiple VPC attachments**: The T02 queue originally specified a single `vpc_id`, but the provider supports multiple VPC attachments. Private zones are commonly shared across VPCs, so `repeated vpc_attachments` is the correct design.

2. **Omitted fields**: `proxy_pattern`, `sync_status`, `user_info`, `lang`, `user_client_ip`, and record `status` are omitted to keep the spec clean. These are niche features that add complexity without broad utility.

3. **Composite bundling**: Zone + attachment + records are bundled because a zone without an attachment is non-functional, and records are natural children of the zone.

### Common Use Cases

1. **Service discovery**: Internal services register A records (e.g., `api.svc.internal` -> `10.0.1.50`) so other services can discover them by hostname.

2. **Database endpoints**: A shared `db.corp` zone attached to multiple VPCs lets applications resolve database hostnames (e.g., `mysql.db.corp`) regardless of which VPC they're in.

3. **Split-horizon DNS**: Use private zone for internal resolution of a domain that also has public records. VPC resources see the private IPs; external users see the public IPs.

4. **PTR records**: Reverse DNS for internal IP addresses (e.g., `10.0.1.100` -> `api.svc.internal`).

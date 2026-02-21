# AliCloudPrivateDnsZone -- Pulumi Architecture Overview

## Resource Graph

```
AliCloudProvider
  └── pvtz.Zone (private hosted zone)
        ├── pvtz.ZoneAttachment (VPC bindings)
        └── pvtz.ZoneRecord (one per spec.records entry)
```

## Data Flow

1. **Stack input** is loaded from the Pulumi config and deserialized into `AliCloudPrivateDnsZoneStackInput`
2. **Locals** are initialized: system tags are computed from metadata, merged with user tags
3. **Zone** is created with zone_name, remark, resource_group_id, and tags
4. **VPC attachment** is created as a child of the zone, binding all `spec.vpc_attachments` entries. Cross-region attachments use the `region_id` field.
5. **Records** are created as children of the zone, one `pvtz.ZoneRecord` per `spec.records` entry

## Key Design Choices

- The zone attachment is a single resource managing all VPCs (not one attachment per VPC). This matches the Alibaba Cloud API behavior where `alicloud_pvtz_zone_attachment` manages the full set of attached VPCs.
- Records are created as children of the zone for clean dependency tracking and deletion ordering.
- The `optionalString` helper converts empty strings to nil, avoiding unnecessary API calls for unset optional fields.

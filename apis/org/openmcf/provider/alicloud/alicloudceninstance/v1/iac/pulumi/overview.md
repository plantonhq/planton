# AlicloudCenInstance Pulumi Module Overview

## Architecture

The module creates a CEN instance and iterates over the `spec.attachments` list to create child-instance attachments. Each attachment connects a VPC, VBR, or CCN to the CEN hub.

```
AlicloudCenInstance (spec)
├── cen.Instance (CEN hub)
└── cen.InstanceAttachment[] (per attachment entry)
    ├── Attachment 0: VPC in cn-hangzhou
    ├── Attachment 1: VPC in cn-shanghai
    └── Attachment N: VPC/VBR/CCN in any region
```

## Resource Lifecycle

1. **Provider** -- created with the region from spec (for API routing)
2. **CEN Instance** -- created with name, description, protection level, tags
3. **Attachments** -- created as children of the CEN instance, each referencing a child instance by ID, type, and region

## Outputs

| Output Key | Source |
|-----------|--------|
| `cen_id` | `cenInstance.ID()` |
| `cen_instance_name` | `cenInstance.CenInstanceName` |

## Dependencies

- Attachments depend on the CEN instance (parent relationship)
- Attachments are created sequentially within the Pulumi resource graph
- External VPCs must exist before deployment (validated at apply time)

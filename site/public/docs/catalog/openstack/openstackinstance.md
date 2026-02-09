---
title: "Instance"
description: "Instance deployment documentation"
icon: "package"
order: 100
componentName: "openstackinstance"
---

# OpenStackInstance -- Research Documentation

## Terraform Provider Analysis

**Resource**: `openstack_compute_instance_v2`
**Provider version**: v3.x (terraform-provider-openstack)
**Source**: `openstack/resource_openstack_compute_instance_v2.go`

### Fields Included (15 of 30+)

| TF Field | OpenMCF Field | Type | Rationale |
|----------|--------------|------|-----------|
| `name` | `metadata.name` | string | Standard KRM pattern |
| `flavor_name` | `flavor_name` | string | Primary flavor selection (human-readable) |
| `flavor_id` | `flavor_id` | string | Alternative flavor selection (UUID) |
| `image_name` | `image_name` | string | Primary image selection (name lookup) |
| `image_id` | `image_id` | string | Alternative image selection (UUID) |
| `key_pair` | `key_pair` | StringValueOrRef | FK to OpenStackKeypair for DAG ordering |
| `network` | `networks` | repeated nested | Network attachment configuration |
| `security_groups` | `security_groups` | repeated StringValueOrRef | FK to OpenStackSecurityGroup (names, not UUIDs) |
| `block_device` | `block_device` | repeated nested | Boot-from-volume and additional storage |
| `user_data` | `user_data` | string | Cloud-init configuration |
| `metadata` | `metadata` | map | Instance metadata |
| `config_drive` | `config_drive` | optional bool | Config drive toggle |
| `scheduler_hints.group` | `server_group_id` | StringValueOrRef | FK to ServerGroup (flattened from nested block) |
| `availability_zone` | `availability_zone` | string | AZ placement |
| `tags` | `tags` | repeated string | Instance tags |
| `region` | `region` | string | Region override |

### Fields Excluded (15+ of 30+)

| TF Field | Reason |
|----------|--------|
| `personality` | Deprecated in modern Nova; conflicts with cloud-init |
| `admin_pass` | Sensitive field, not appropriate for declarative IaC |
| `network_mode` | Niche ("auto"/"none"); replaced by requiring explicit networks |
| `hypervisor_hostname` | Admin-only, conflicts with personality |
| `force_delete` | Operational escape hatch |
| `stop_before_destroy` | Operational lifecycle |
| `power_state` | Default "active" is correct; managing power via IaC is risky |
| `vendor_options` | Terraform-specific workaround block |
| `availability_zone_hints` | Niche alternative to availability_zone |
| `block_device.guest_format` | Low-level disk formatting |
| `block_device.device_type` | Low-level device control |
| `block_device.disk_bus` | Low-level bus selection |
| `block_device.multiattach` | Shared volume (niche) |
| `network.fixed_ip_v6` | IPv6 can be added later |
| `scheduler_hints.different_host` | Admin-level scheduling |
| `scheduler_hints.same_host` | Admin-level scheduling |
| `scheduler_hints.query` | Admin-level scheduling |
| `scheduler_hints.target_cell` | Admin-level scheduling |
| `scheduler_hints.additional_properties` | Escape hatch |

## Design Decisions

### 1. Singular `server_group_id` instead of nested `scheduler_hints`

The TF provider nests `group` inside `scheduler_hints`, which has 8 fields. For the 80/20 spec, only `group` is needed. A flat `server_group_id` field:
- Follows the FK naming convention
- Produces cleaner YAML
- The IaC modules map it to `scheduler_hints { group = ... }` internally

### 2. `security_groups` FK targets `status.outputs.name` (not UUID)

The Compute API (Nova) uses security group NAMES, unlike the Networking API (Neutron) which uses UUIDs. This is a well-known OpenStack API inconsistency. The FK resolves to the SecurityGroup component's `name` output.

### 3. `key_pair` FK targets `status.outputs.name`

Similar to security_groups, the Compute API uses the keypair name (not UUID). The FK provides DAG ordering in InfraCharts while resolving to the correct name value.

### 4. `image_id` and `image_name` as plain strings (not FK)

OpenStackImage (2514) is Phase 4. Most instances reference pre-existing images by name. The FK would only be useful if an InfraChart creates both an Image and an Instance from it -- an extremely rare workflow. Plain strings keep this simple; FK can be added when the Image component is created if demand warrants it.

### 5. Both `flavor_name` and `flavor_id` included

`flavor_name` is the 90% case (human-readable, what users type in Horizon). `flavor_id` serves automation workflows that reference flavors by UUID. CEL mutual exclusion keeps the spec clean.

### 6. Required networks (min 1)

Since `network_mode` is excluded, every instance needs explicit network configuration. Requiring at least one network prevents silent failures from missing network config.

### 7. Block device included despite complexity

Boot-from-volume is fundamental to production OpenStack usage (persistent root disks). 7 fields in the nested message is justified by the real-world importance of this feature.

## Pulumi SDK Mapping

| Proto Field | Pulumi Arg | Notes |
|------------|------------|-------|
| `flavor_name` | `FlavorName` | `StringPtr` |
| `flavor_id` | `FlavorId` | `StringPtr` |
| `image_name` | `ImageName` | `StringPtr` |
| `image_id` | `ImageId` | `StringPtr` |
| `key_pair` | `KeyPair` | `StringPtr` (resolved from FK) |
| `networks` | `Networks` | `InstanceNetworkArray` |
| `security_groups` | `SecurityGroups` | `StringArray` (resolved names) |
| `block_device` | `BlockDevices` | `InstanceBlockDeviceArray` |
| `user_data` | `UserData` | `StringPtr` |
| `metadata` | `Metadata` | `StringMap` |
| `config_drive` | `ConfigDrive` | `BoolPtr` |
| `server_group_id` | `SchedulerHints[].Group` | Wrapped in scheduler hints |
| `availability_zone` | `AvailabilityZone` | `StringPtr` |
| `tags` | `Tags` | `StringArray` |
| `region` | `Region` | `StringPtr` |

## ForceNew Behavior

Fields that recreate the instance on change:
- `key_pair`, `networks` (all sub-fields), `user_data`, `config_drive`
- `server_group_id` (scheduler_hints), `availability_zone`, `block_device` (all sub-fields)
- `region`

Fields that update in-place:
- `flavor_name`/`flavor_id` (resize), `security_groups`, `metadata`, `tags`
- `image_name`/`image_id` (rebuild)

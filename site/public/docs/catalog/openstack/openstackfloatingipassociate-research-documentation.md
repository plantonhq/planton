---
title: "OpenStackFloatingIpAssociate Research Documentation"
description: "OpenStackFloatingIpAssociate Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackfloatingipassociate"
---

# OpenStackFloatingIpAssociate Research Documentation

## Terraform Resource Analysis

**Resource**: `openstack_networking_floatingip_associate_v2`
**Provider**: `terraform-provider-openstack/openstack` v3.x

### Schema Analysis (4 attributes total, all selected)

| Attribute | Type | Required | ForceNew | Notes |
|-----------|------|----------|----------|-------|
| `floating_ip` | string | Yes | Yes | IP address or floating IP UUID |
| `port_id` | string | Yes | No | Port to associate with |
| `fixed_ip` | string | No | No | Specific fixed IP on multi-IP ports (Computed) |
| `region` | string | No | Yes | Region override (Computed) |

This is one of the simplest Terraform resources -- only 4 attributes, all included in the OpenMCF spec. No 80/20 filtering was needed.

### Key Design Decisions

#### FK to `address` Instead of `id`

The `floating_ip` FK annotation targets `OpenStackFloatingIp.status.outputs.address` (the IP address like "203.0.113.42"), not `floating_ip_id` (the UUID). This is because:

1. The Terraform provider's `floating_ip` attribute accepts either an IP address or a floating IP UUID
2. The Pulumi SDK's `FloatingIp` field is typed as `StringInput` (accepts either)
3. The `address` output is more human-readable and commonly used for DNS/firewall pre-configuration
4. Using `address` as the FK target is consistent with the TF provider's documentation which titles the field as "IP Address of an existing floating IP"

This is the first FK in OpenMCF that targets a non-UUID output. All previous FKs target resource IDs (UUIDs).

#### Companion Slot Pattern

This component uses enum 2526, following the "companion slot" numbering:
- 2505 / 2525 = SecurityGroup / SecurityGroupRule
- 2506 / 2526 = FloatingIp / FloatingIpAssociate

The gap between 2506 and 2526 is intentional, reserving the 2507-2524 range for primary components and the 2525+ range for companion/satellite components.

### Pulumi SDK Mapping

| Spec Field | Pulumi Type | Pulumi Field |
|------------|-------------|--------------|
| `floating_ip` | `pulumi.StringInput` | `FloatingIp` |
| `port_id` | `pulumi.StringInput` | `PortId` |
| `fixed_ip` | `pulumi.StringPtrInput` | `FixedIp` |
| `region` | `pulumi.StringPtrInput` | `Region` |

### Comparison with OpenStackFloatingIp Built-in Association

| Aspect | FloatingIpAssociate (this) | FloatingIp.port_id |
|--------|---------------------------|-------------------|
| DAG visibility | Explicit node in InfraChart | Hidden inside FloatingIp |
| Lifecycle | Separate from allocation | Coupled with allocation |
| Use case | InfraCharts, decoupled mgmt | Simple manifests |
| Resources created | 1 (associate only) | 1 (floatingip with port_id) |
| FK count | 2 required | 1 required + 1 optional |

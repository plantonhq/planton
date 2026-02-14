---
title: "OpenStackServerGroup -- Research Documentation"
description: "OpenStackServerGroup -- Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackservergroup"
---

# OpenStackServerGroup -- Research Documentation

## Terraform Provider Analysis

**Resource**: `openstack_compute_servergroup_v2`
**Provider version**: v3.x (terraform-provider-openstack)
**Source**: `openstack/resource_openstack_compute_servergroup_v2.go`

### All Schema Fields

| Field | Type | Required | Optional | Computed | ForceNew | Description |
|-------|------|----------|----------|----------|----------|-------------|
| `name` | string | Yes | - | - | Yes | Unique name for the server group |
| `policies` | list(string) | - | Yes | - | Yes | Exactly one policy (MinItems=1, MaxItems=1) |
| `rules` | list(object) | - | Yes | Yes | Yes | Rules applied to policy (only `max_server_per_host` for anti-affinity, API 2.64+) |
| `region` | string | - | Yes | Yes | Yes | OpenStack region |
| `members` | list(string) | - | - | Yes | - | Instance UUIDs in the group (read-only) |
| `value_specs` | map(string) | - | Yes | - | Yes | Additional API parameters |

### Fields Included (3 of 6)

| Field | OpenMCF Name | Rationale |
|-------|-------------|-----------|
| `name` | `metadata.name` | Standard KRM pattern |
| `policies` | `policy` (singular) | Only one policy allowed; singular is cleaner |
| `region` | `region` | Standard region override |

### Fields Excluded (3 of 6)

| Field | Reason |
|-------|--------|
| `rules` | Requires Nova API 2.64+, only for anti-affinity. Niche feature for limiting servers per host. |
| `members` | Computed read-only field. Exposed as output, not input. |
| `value_specs` | Escape hatch. Excluded from all components (consistent policy). |

## Design Decisions

### 1. Singular `policy` instead of `policies` list

The Terraform provider models policies as a list with `MinItems=1, MaxItems=1`. This is because the Nova API historically accepted a list, but has always enforced exactly one policy. Using a singular string:
- Produces cleaner YAML: `policy: anti-affinity` vs `policies: ["anti-affinity"]`
- Eliminates the need for list size validation
- Accurately reflects the semantic: a server group has ONE policy
- The IaC modules wrap this into a list when calling the API

### 2. Exclusion of `rules` / `max_server_per_host`

The `rules` field allows limiting how many instances from the same group can be placed on one host (e.g., `max_server_per_host = 2` means up to 2 members per host). This:
- Requires Nova API microversion 2.64+
- Only applies to `anti-affinity` policy
- Is useful for large clusters where strict 1:1 host mapping is impractical
- Is a niche optimization, not needed for the ARM developer environment use case

The field can be added later if demand arises. The 300-slot OpenStack enum range (2500-2799) has plenty of room for evolution.

### 3. Immutability

All fields on `openstack_compute_servergroup_v2` are `ForceNew: true`. The resource has no Update function -- any change recreates the server group. This is reflected in the proto comments.

### 4. `members` as output only

The `members` field is computed by OpenStack as instances reference the server group via scheduler hints. It cannot be set by the user. We expose it as a stack output for observability.

## Pulumi SDK Mapping

| Proto Field | Pulumi Arg | Notes |
|------------|------------|-------|
| `metadata.name` | `Name` | Server group name |
| `policy` | `Policies` | Wrapped as `[]string{policy}` |
| `region` | `Region` | Optional, `StringPtr` |

**Pulumi resource**: `compute.NewServerGroup()`
**SDK package**: `github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/compute`

## Terraform Mapping

| Proto Field | TF Attribute | Notes |
|------------|-------------|-------|
| `metadata.name` | `name` | Server group name |
| `policy` | `policies` | Wrapped as `[var.spec.policy]` |
| `region` | `region` | Conditional null for empty string |

**TF resource**: `openstack_compute_servergroup_v2`

## API Microversion Behavior

The Terraform provider automatically selects the appropriate Nova API microversion:

| Policy | Microversion |
|--------|-------------|
| `affinity` | `""` (legacy) |
| `anti-affinity` | `""` (legacy) |
| `soft-affinity` | `"2.15"` |
| `soft-anti-affinity` | `"2.15"` |
| Any + `rules` | `"2.64"` |

Since we exclude `rules`, the maximum required microversion is 2.15 (for soft policies).
Legacy policies work on all OpenStack versions.

## Production Best Practices

1. **Name server groups descriptively**: `db-anti-affinity`, `cache-affinity`, `app-soft-spread`
2. **Use anti-affinity for databases**: Ensures replicas survive host failures
3. **Use soft policies for large clusters**: Hard anti-affinity fails if there aren't enough hosts
4. **Plan for group recreation**: Changing policy recreates the group; existing instances are orphaned
5. **Monitor members**: Use the `members` output to verify instances are correctly grouped

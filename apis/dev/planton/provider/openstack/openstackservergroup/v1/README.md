# OpenStackServerGroup

An OpenStack Compute server group that controls instance placement on hypervisors.

## Overview

A server group is a placement constraint for compute instances. By assigning instances to a server group with an affinity or anti-affinity policy, you control whether instances are co-located on the same hypervisor or spread across different hypervisors.

## When to Use

- **Anti-affinity for HA**: Spread database replicas or application instances across different physical hosts to survive host failures.
- **Affinity for performance**: Co-locate tightly-coupled services on the same host to minimize network latency.
- **Soft policies**: Use `soft-affinity` or `soft-anti-affinity` when you prefer a placement strategy but don't want scheduling to fail if the constraint can't be satisfied.

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `policy` | string | Yes | Placement policy: `affinity`, `anti-affinity`, `soft-affinity`, `soft-anti-affinity` |
| `region` | string | No | Override the region from the provider config |

## Outputs

| Output | Description |
|--------|-------------|
| `server_group_id` | UUID of the server group (used by Instance's `server_group_id` FK) |
| `name` | Name of the server group |
| `members` | List of instance UUIDs in this group (computed) |
| `region` | Region where the server group was created |

## Usage

### Minimal Example

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackServerGroup
metadata:
  name: db-anti-affinity
spec:
  policy: anti-affinity
```

### With Region Override

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackServerGroup
metadata:
  name: app-affinity
  org: acme-corp
  env: production
spec:
  policy: affinity
  region: RegionOne
```

## Referencing from Instance

Instances reference a server group via the `server_group_id` field:

```yaml
# InfraChart mode (value_from reference)
apiVersion: openstack.planton.dev/v1
kind: OpenStackInstance
metadata:
  name: db-replica-1
spec:
  flavor_name: m1.large
  image_name: ubuntu-22.04
  server_group_id:
    value_from:
      name: db-anti-affinity
  networks:
    - uuid:
        value_from:
          name: app-network
```

## Policy Details

| Policy | Behavior | API Version |
|--------|----------|-------------|
| `affinity` | All members on the same hypervisor | Legacy (all versions) |
| `anti-affinity` | Members on different hypervisors | Legacy (all versions) |
| `soft-affinity` | Best-effort same hypervisor | Requires Nova API 2.15+ |
| `soft-anti-affinity` | Best-effort different hypervisors | Requires Nova API 2.15+ |

## Important Notes

- **Immutable**: All fields are ForceNew. Changing the policy recreates the server group.
- **Orphaned members**: Recreating a server group does not migrate existing member instances. They remain on their current hosts but are no longer associated with any group.
- **No name attribute in API**: The `metadata.name` is used as the OpenStack server group name. OpenStack server group names are not unique -- the UUID is the true identifier.
- **Members are computed**: The `members` output is populated by OpenStack as instances are launched with `scheduler_hints` referencing this group. It is empty at creation time.

## Related Components

- **OpenStackInstance** (2508) -- References this server group via `server_group_id` FK

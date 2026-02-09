# OpenStackSecurityGroupRule Research Documentation

## Overview

The OpenStackSecurityGroupRule component wraps the `openstack_networking_secgroup_rule_v2` Terraform resource (and `openstack.networking.SecGroupRule` Pulumi resource) as a standalone, independently managed deployment component.

## Terraform Provider Analysis

### Resource: `openstack_networking_secgroup_rule_v2`

**Source**: `terraform-provider-openstack/openstack/resource_openstack_networking_secgroup_rule_v2.go`

All fields are **ForceNew** -- any change recreates the rule. Security group rules are immutable once created.

### Full Schema (12 fields)

| Field | Type | Required | Included | Rationale |
|-------|------|----------|----------|-----------|
| `security_group_id` | string | Yes | Yes (as FK) | Core field -- which SG this rule belongs to |
| `direction` | string | Yes | Yes | "ingress" or "egress" |
| `ethertype` | string | Yes | Yes | "IPv4" or "IPv6" |
| `protocol` | string | No | Yes | tcp/udp/icmp/etc. |
| `port_range_min` | int | No | Yes | Lower port or ICMP type |
| `port_range_max` | int | No | Yes | Upper port or ICMP code |
| `remote_ip_prefix` | string | No | Yes | CIDR filter |
| `remote_group_id` | string | No | Yes (as FK) | Reference to another SG |
| `remote_address_group_id` | string | No | **No** | Niche, rarely used |
| `description` | string | No | Yes | Human-readable |
| `region` | string | No | Yes | Region override |
| `tenant_id` | string | No | **No** | Admin-only |

### RequiredWith Constraints

- `port_range_min` requires: `protocol`, `port_range_max`
- `port_range_max` requires: `protocol`, `port_range_min`

### ConflictsWith Constraints

- `remote_group_id` conflicts with: `remote_ip_prefix`, `remote_address_group_id`
- `remote_ip_prefix` conflicts with: `remote_group_id`, `remote_address_group_id`
- `remote_address_group_id` conflicts with: `remote_group_id`, `remote_ip_prefix`

## Design Decisions

### Why standalone rules exist alongside inline rules

The `OpenStackSecurityGroup` component supports inline rules via its `rules[]` field. This is convenient but has limitations:

1. **No FK resolution on `remote_group_id`**: Inline rules use plain strings for `remote_group_id`. In an InfraChart where security groups are created together, you can't reference another SG's output UUID because it doesn't exist yet.

2. **No DAG visibility**: Inline rules are invisible in the InfraChart DAG graph. They're bundled with the parent SG resource.

3. **No independent lifecycle**: All inline rules are co-managed. You can't add a rule to an existing SG in a separate InfraChart layer.

The standalone `OpenStackSecurityGroupRule` solves all three problems by making each rule its own KRM resource with full FK support.

### Two FKs to the same kind

This is the first component with two `StringValueOrRef` foreign keys pointing to the same `CloudResourceKind` (`OpenStackSecurityGroup`). Both resolve to `status.outputs.security_group_id`. This is semantically correct -- `security_group_id` identifies ownership, while `remote_group_id` identifies a traffic filter.

### Excluded fields

- **`remote_address_group_id`**: Address groups are a newer Neutron extension not widely supported across OpenStack deployments. Can be added later if ARM's OpenStack has it.
- **`tenant_id`**: Admin-only field, excluded from all OpenStack components per project design decisions.

## Relationship to OpenStackSecurityGroup

```
OpenStackSecurityGroup (2505)
├── spec.rules[] (inline rules, plain strings, no FK)
└── status.outputs.security_group_id

OpenStackSecurityGroupRule (2525)
├── spec.security_group_id (required FK -> SecurityGroup)
├── spec.remote_group_id (optional FK -> SecurityGroup)
└── status.outputs.rule_id
```

## Protocol Values

The Terraform provider accepts these protocol values:
- Empty string (any protocol)
- `tcp`, `udp`, `icmp`
- `ah`, `dccp`, `egp`, `esp`, `gre`, `igmp`
- `ipv6-encap`, `ipv6-frag`, `ipv6-icmp`, `ipv6-nonxt`, `ipv6-opts`, `ipv6-route`
- `ospf`, `pgm`, `rsvp`, `sctp`, `udplite`, `vrrp`, `ipip`
- Integer 0-255 (raw IP protocol number)

We don't constrain protocol with `in` validation because the full set is large and version-dependent. The OpenStack API validates it server-side.

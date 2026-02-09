# OpenStackSecurityGroup

An OpenMCF deployment component for managing OpenStack Neutron security groups with optional inline rules.

## Overview

A security group acts as a virtual firewall for instances and network ports, controlling ingress and egress traffic through a set of rules. This component creates the security group and optionally provisions inline rules as part of the same deployment.

## Key Features

- **Inline rules**: Define security group rules directly in the spec for self-contained deployments
- **Delete default rules**: Remove OpenStack's auto-created egress rules for zero-trust baselines
- **Stateful/stateless**: Choose between stateful (default) and stateless firewall modes
- **Stable IaC state**: Each inline rule is keyed by a user-provided `key` field, ensuring stable Terraform `for_each` / Pulumi resource naming

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | No | Human-readable description |
| `delete_default_rules` | bool | No | Remove default egress rules after creation |
| `stateful` | bool | No | Stateful (true) or stateless (false) mode |
| `rules` | SecurityGroupRule[] | No | Inline security group rules |
| `tags` | string[] | No | Tags for filtering and organization |
| `region` | string | No | Region override |

### SecurityGroupRule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `key` | string | Yes | Unique identifier for IaC state (e.g., "allow-ssh") |
| `direction` | string | Yes | "ingress" or "egress" |
| `ethertype` | string | Yes | "IPv4" or "IPv6" |
| `protocol` | string | No | "tcp", "udp", "icmp", etc. |
| `port_range_min` | int32 | No | Start port (or ICMP type) |
| `port_range_max` | int32 | No | End port (or ICMP code) |
| `remote_ip_prefix` | string | No | Source/destination CIDR |
| `remote_group_id` | string | No | Security group UUID for group-based rules |
| `description` | string | No | Per-rule description |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `security_group_id` | UUID of the security group (primary FK for downstream components) |
| `name` | Name of the security group |
| `region` | OpenStack region |

## Downstream References

This component's `security_group_id` output is referenced by:
- `OpenStackSecurityGroupRule` via `security_group_id` FK
- `OpenStackNetworkPort` via `security_group_ids[]` FK
- `OpenStackInstance` via `security_groups[]`

## Inline Rules vs Standalone Rules

| Feature | Inline Rules (this component) | Standalone OpenStackSecurityGroupRule |
|---------|-------------------------------|--------------------------------------|
| Defined in | SecurityGroup spec | Separate resource |
| FK support | Plain strings only | StringValueOrRef with FK resolution |
| DAG visibility | Not visible as separate nodes | Visible as individual DAG nodes |
| Best for | Self-contained SGs, simple setups | InfraCharts, cross-resource references |

## Validation Rules

- `port_range_min` and `port_range_max` must both be set or both unset
- Port ranges require `protocol` to be specified
- `remote_group_id` and `remote_ip_prefix` are mutually exclusive
- Inline rule `key` values must be unique within the security group
- `direction` must be "ingress" or "egress"
- `ethertype` must be "IPv4" or "IPv6"
- `tags` must be unique

## Terraform Resource

- `openstack_networking_secgroup_v2` (security group)
- `openstack_networking_secgroup_rule_v2` (inline rules, via `for_each`)

## Pulumi Resource

- `openstack.networking.SecGroup`
- `openstack.networking.SecGroupRule` (inline rules, one per rule)

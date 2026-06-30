# OpenStackSecurityGroupRule

A standalone OpenStack Neutron security group rule, managed as an independent deployment component.

## Overview

`OpenStackSecurityGroupRule` creates a single firewall rule that belongs to an existing security group. This is the standalone counterpart to inline rules in `OpenStackSecurityGroup.rules[]`.

**When to use this component** (instead of inline rules):
- When rules need to reference other security groups via `remote_group_id` with DAG-resolved FK references
- When individual rules need independent lifecycle management
- When InfraChart DAG visualization must show each rule as a visible node with dependency edges

**When to use inline rules** (in `OpenStackSecurityGroup`):
- When all rules for a security group are co-managed as a unit
- When no cross-SG references are needed
- When simplicity is preferred over granular visibility

## Foreign Key Relationships

This component has two StringValueOrRef foreign keys, both pointing to `OpenStackSecurityGroup`:

| Field | Required | Resolves To |
|-------|----------|-------------|
| `security_group_id` | Yes | `OpenStackSecurityGroup.status.outputs.security_group_id` |
| `remote_group_id` | No | `OpenStackSecurityGroup.status.outputs.security_group_id` |

## Quick Start

### Standalone usage (literal UUID)

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-ssh-ingress
spec:
  security_group_id:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  port_range_min: 22
  port_range_max: 22
  remote_ip_prefix: "0.0.0.0/0"
  description: "Allow SSH from anywhere"
```

### InfraChart usage (FK reference)

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-ssh-from-bastion
spec:
  security_group_id:
    value_from:
      name: app-sg
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  port_range_min: 22
  port_range_max: 22
  remote_group_id:
    value_from:
      name: bastion-sg
  description: "Allow SSH from bastion security group"
```

## Terraform Resource

`openstack_networking_secgroup_rule_v2`

## Pulumi Resource

`openstack.networking.SecGroupRule`

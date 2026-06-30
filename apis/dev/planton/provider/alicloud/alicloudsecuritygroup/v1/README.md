# AliCloudSecurityGroup

Manages an Alibaba Cloud Security Group with bundled security rules.

## Overview

A security group is a stateful virtual firewall that controls inbound and outbound traffic for ECS instances, ACK worker nodes, RDS instances, and other VPC-aware resources. Each security group belongs to exactly one VPC.

This component bundles the security group with its rules because a security group without rules is effectively an open door -- the rules define the access policy.

### What Gets Created

- **Security Group** -- a virtual firewall bound to a VPC
- **Security Group Rules** -- individual ingress and egress rules evaluated by priority

### How Rules Work

Rules are evaluated by priority within their direction (ingress or egress). Lower priority numbers are evaluated first (1 = highest priority, 100 = lowest). The first matching rule determines whether traffic is allowed or dropped. If no rule matches, traffic is denied by default (except intra-group traffic controlled by `innerAccessPolicy`).

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`) |
| `vpcId` | StringValueOrRef | VPC ID that this security group belongs to |
| `securityGroupName` | string | Security group name (2-128 chars) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | `""` | Human-readable description (2-256 chars) |
| `innerAccessPolicy` | string | `"Accept"` | Intra-group traffic policy: `Accept` or `Drop` |
| `resourceGroupId` | string | `""` | Resource group for organizational grouping |
| `tags` | map | `{}` | Key-value tags applied to the security group |
| `rules` | list | `[]` | Security group rules (see below) |

### Rule Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `type` | string | Yes | -- | Direction: `ingress` or `egress` |
| `ipProtocol` | string | Yes | -- | Protocol: `tcp`, `udp`, `icmp`, `gre`, `all` |
| `portRange` | string | No | `"-1/-1"` | Port range in `start/end` format (e.g., `80/80`, `1/65535`) |
| `cidrIp` | string | No | -- | IPv4 CIDR source/destination (e.g., `0.0.0.0/0`, `10.0.0.0/8`) |
| `sourceSecurityGroupId` | string | No | -- | Source SG ID for SG-to-SG rules |
| `priority` | int | No | `1` | Evaluation priority (1-100, lower = higher priority) |
| `policy` | string | No | `"accept"` | Action: `accept` or `drop` |
| `description` | string | No | -- | Rule description |

At least one of `cidrIp` or `sourceSecurityGroupId` should be specified per rule.

### Port Range Format

| Protocol | Port Range | Notes |
|----------|-----------|-------|
| tcp/udp | `"80/80"` | Single port |
| tcp/udp | `"8080/8090"` | Port range |
| tcp/udp | `"1/65535"` | All ports |
| icmp/gre/all | `"-1/-1"` | Required for non-TCP/UDP protocols |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `security_group_id` | The security group ID, referenced by downstream components via StringValueOrRef |
| `security_group_name` | The security group name as created |

## Related Components

- **AliCloudVpc** -- the VPC this security group belongs to
- **AliCloudEcsInstance** -- ECS instances associated with this security group
- **AliCloudAckManagedCluster** -- ACK worker nodes can use this security group

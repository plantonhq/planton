---
title: "Security Group Rule"
description: "Security Group Rule deployment documentation"
icon: "package"
order: 100
componentName: "openstacksecuritygrouprule"
---

# OpenStack Security Group Rule

Deploys a standalone OpenStack Neutron security group rule with configurable direction, protocol, port range, and remote source filtering. Unlike inline rules defined in OpenStackSecurityGroup `rules[]`, standalone rules support `StringValueOrRef` foreign keys on both `securityGroupId` and `remoteGroupId`, enabling cross-security-group references and explicit DAG wiring in InfraCharts.

## What Gets Created

When you deploy an OpenStackSecurityGroupRule resource, OpenMCF provisions:

- **Security Group Rule** — an `openstack_networking_secgroup_rule_v2` resource attached to the specified security group, with the configured direction, ethertype, protocol, port range, and remote source. All fields are ForceNew: any change to the rule recreates it. The rule appears as its own node in InfraChart DAG visualizations, making cross-security-group dependencies visually explicit.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **An existing security group** (UUID or OpenMCF-managed OpenStackSecurityGroup) to attach the rule to
- **A remote security group** (UUID or OpenMCF-managed OpenStackSecurityGroup) if using `remoteGroupId` instead of `remoteIpPrefix`

## Quick Start

Create a file `security-group-rule.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-ssh-ingress
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackSecurityGroupRule.allow-ssh-ingress
spec:
  securityGroupId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  portRangeMin: 22
  portRangeMax: 22
  remoteIpPrefix: "0.0.0.0/0"
```

Deploy:

```shell
openmcf apply -f security-group-rule.yaml
```

This creates an ingress rule allowing SSH (TCP port 22) from any IPv4 address on the specified security group.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `securityGroupId` | `StringValueOrRef` | UUID of the security group this rule belongs to. Can reference an OpenStackSecurityGroup resource via `valueFrom`. | Required |
| `direction` | `string` | Direction of traffic: `ingress` (incoming) or `egress` (outgoing). | Must be `ingress` or `egress` |
| `ethertype` | `string` | Layer-3 protocol type for the rule. | Must be `IPv4` or `IPv6` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `protocol` | `string` | all protocols | IP protocol for the rule. Common values: `tcp`, `udp`, `icmp`, `icmpv6`. Also accepts any IANA protocol name or number (0-255). If omitted, the rule applies to all protocols. |
| `portRangeMin` | `int32` | — | Minimum port number (0-65535) for TCP/UDP, or ICMP type (0-255). Must be set together with `portRangeMax`. Requires `protocol` to be set. |
| `portRangeMax` | `int32` | — | Maximum port number (0-65535) for TCP/UDP, or ICMP code (0-255). Must be set together with `portRangeMin`. Requires `protocol` to be set. |
| `remoteIpPrefix` | `string` | — | CIDR restricting the rule to traffic from/to a specific IP range. For ingress: source range. For egress: destination range. Example: `0.0.0.0/0`, `10.0.0.0/8`. Mutually exclusive with `remoteGroupId`. |
| `remoteGroupId` | `StringValueOrRef` | — | Restricts the rule to traffic from/to instances in another security group (or the same group for self-referencing rules). Can reference an OpenStackSecurityGroup resource via `valueFrom`. Mutually exclusive with `remoteIpPrefix`. |
| `description` | `string` | — | Human-readable description of the rule, visible in Horizon and API responses. |
| `region` | `string` | provider default | Overrides the region from the provider config for this rule. |

## Examples

### Allow SSH from Any Address

A single ingress rule permitting SSH access from all IPv4 addresses:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-ssh
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackSecurityGroupRule.allow-ssh
spec:
  securityGroupId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  portRangeMin: 22
  portRangeMax: 22
  remoteIpPrefix: "0.0.0.0/0"
  description: "Allow SSH from any IPv4 address"
```

### Allow HTTP/HTTPS from a Subnet

An ingress rule opening the standard web ports from a private subnet, demonstrating a port range:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-https-from-private
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OpenStackSecurityGroupRule.allow-https-from-private
spec:
  securityGroupId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  portRangeMin: 443
  portRangeMax: 443
  remoteIpPrefix: "10.0.0.0/8"
  description: "Allow HTTPS from private network"
```

### Cross-Security-Group Rule with Foreign Key References

A rule that allows all TCP traffic from instances in one security group to instances in another, using `valueFrom` references to OpenMCF-managed security groups. This is the primary use case for standalone rules over inline rules:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: app-to-db-tcp
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackSecurityGroupRule.app-to-db-tcp
spec:
  securityGroupId:
    valueFrom:
      kind: OpenStackSecurityGroup
      name: db-sg
      field: status.outputs.security_group_id
  direction: ingress
  ethertype: IPv4
  protocol: tcp
  portRangeMin: 5432
  portRangeMax: 5432
  remoteGroupId:
    valueFrom:
      kind: OpenStackSecurityGroup
      name: app-sg
      field: status.outputs.security_group_id
  description: "Allow PostgreSQL from app tier to database tier"
```

### Broad Egress Rule for All Protocols

An egress rule allowing all outbound IPv4 traffic from a security group, with no protocol or port restrictions:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroupRule
metadata:
  name: allow-all-egress
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackSecurityGroupRule.allow-all-egress
spec:
  securityGroupId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  direction: egress
  ethertype: IPv4
  remoteIpPrefix: "0.0.0.0/0"
  description: "Allow all outbound IPv4 traffic"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `rule_id` | `string` | UUID of the created security group rule |
| `security_group_id` | `string` | UUID of the security group this rule belongs to |
| `direction` | `string` | Direction of the rule (`ingress` or `egress`) |
| `protocol` | `string` | IP protocol of the rule. Empty string if the rule applies to all protocols. |
| `port_range_min` | `int32` | Lower bound of the port range, or ICMP type |
| `port_range_max` | `int32` | Upper bound of the port range, or ICMP code |
| `region` | `string` | OpenStack region where the rule was created |

## Related Components

- [OpenStackSecurityGroup](/docs/catalog/openstack/security-group) — the security group this rule belongs to; also supports inline rules via `rules[]` for simpler configurations
- [OpenStackInstance](/docs/catalog/openstack/instance) — compute instances that reference security groups for network access control
- [OpenStackNetwork](/docs/catalog/openstack/network) — provides the network context where security group rules take effect
- [OpenStackNetworkPort](/docs/catalog/openstack/network-port) — network ports that can have security groups applied directly

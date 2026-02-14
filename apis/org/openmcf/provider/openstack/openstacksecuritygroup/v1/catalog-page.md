# OpenStack Security Group

Deploys an OpenStack Neutron security group with optional inline firewall rules controlling ingress and egress traffic. Security groups act as virtual firewalls for instances and network ports, and this component supports both inline rules (defined in the spec) and standalone rules via the separate OpenStackSecurityGroupRule component for DAG-visible management in InfraCharts.

## What Gets Created

When you deploy an OpenStackSecurityGroup resource, OpenMCF provisions:

- **Security Group** — an `openstack_networking_secgroup_v2` resource with the configured description, stateful/stateless mode, default-rule deletion policy, and tags. The name is derived from `metadata.name`.
- **Inline Security Group Rules** — one `openstack_networking_secgroup_rule_v2` resource per entry in the `rules` array, each keyed by the rule's `key` field for stable IaC state management. Rules control traffic direction, protocol, port range, and source/destination restrictions.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **An OpenStack project** with Neutron networking enabled
- **Existing security group UUIDs** if using `remoteGroupId` in inline rules to reference other security groups

## Quick Start

Create a file `security-group.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: web-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackSecurityGroup.web-sg
spec:
  description: Security group for web servers
  rules:
    - key: allow-http
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 80
      portRangeMax: 80
      remoteIpPrefix: "0.0.0.0/0"
    - key: allow-https
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 443
      portRangeMax: 443
      remoteIpPrefix: "0.0.0.0/0"
```

Deploy:

```shell
openmcf apply -f security-group.yaml
```

This creates a security group named `web-sg` with two ingress rules allowing HTTP and HTTPS traffic from any source. OpenStack's default egress rules (allow all IPv4/IPv6 outbound) are kept.

## Configuration Reference

### Required Fields

The spec has no strictly required fields. A security group can be created with just `metadata.name`, which determines the security group name in OpenStack. When inline rules are provided, each rule requires `key`, `direction`, and `ethertype`.

### Optional Fields

**Security Group Fields**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description stored on the OpenStack resource, visible in Horizon and API responses. |
| `deleteDefaultRules` | `bool` | `false` | When `true`, deletes the two default egress rules (allow-all IPv4 and IPv6) that OpenStack creates automatically. Use for zero-trust baselines. Create-time only; cannot be changed after creation. |
| `stateful` | `bool` | deployment default (typically `true`) | Controls stateful vs stateless mode. Stateful: return traffic is automatically allowed. Stateless: return traffic must be explicitly permitted by rules. Not all OpenStack deployments support stateless security groups. |
| `rules` | `SecurityGroupRule[]` | `[]` | Inline security group rules created alongside the security group. Each entry becomes a separate `openstack_networking_secgroup_rule_v2` resource. For DAG-visible, independently managed rules, use the OpenStackSecurityGroupRule component instead. |
| `tags` | `string[]` | `[]` | Tags for filtering and organization in the OpenStack API and Horizon. Must be unique. |
| `region` | `string` | provider default | Overrides the region from the provider config for this security group and its inline rules. |

**SecurityGroupRule Sub-Fields** (each entry in the `rules` array)

| Field | Type | Required | Default | Description | Validation |
|-------|------|----------|---------|-------------|------------|
| `key` | `string` | Yes | — | Unique identifier for this rule within the security group. Used as the IaC resource name suffix. Use descriptive, kebab-case names (e.g., `allow-ssh`, `egress-all-ipv4`). | Minimum length 1; must be unique across all rules in the spec |
| `direction` | `string` | Yes | — | Traffic direction: `ingress` (incoming) or `egress` (outgoing). | Must be `ingress` or `egress` |
| `ethertype` | `string` | Yes | — | Layer-3 protocol type: `IPv4` or `IPv6`. | Must be `IPv4` or `IPv6` |
| `protocol` | `string` | No | all protocols | IP protocol for the rule. Common values: `tcp`, `udp`, `icmp`, `icmpv6`. Also accepts any IANA protocol name or number (0-255). Required when port ranges are specified. | — |
| `portRangeMin` | `int` | No | — | Minimum port number (0-65535). For ICMP: the ICMP type (0-255). Must be set together with `portRangeMax`. Requires `protocol` to be set. | Must be paired with `portRangeMax` |
| `portRangeMax` | `int` | No | — | Maximum port number (0-65535). For ICMP: the ICMP code (0-255). Must be set together with `portRangeMin`. Requires `protocol` to be set. | Must be paired with `portRangeMin` |
| `remoteIpPrefix` | `string` | No | — | CIDR to restrict traffic source (ingress) or destination (egress). Examples: `0.0.0.0/0`, `10.0.0.0/8`, `203.0.113.0/24`. Mutually exclusive with `remoteGroupId`. | — |
| `remoteGroupId` | `string` | No | — | UUID of another security group to restrict traffic to members of that group. Use for self-referencing rules or inter-group access. Mutually exclusive with `remoteIpPrefix`. | — |
| `description` | `string` | No | — | Human-readable description stored on the OpenStack rule resource. | — |

**Cross-Field Validations:**
- `portRangeMin` and `portRangeMax` must both be set or both unset
- Port ranges require `protocol` to be specified
- `remoteGroupId` and `remoteIpPrefix` are mutually exclusive

## Examples

### Basic Security Group with Common Web Rules

A security group allowing HTTP, HTTPS, and SSH access from any source:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: web-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenStackSecurityGroup.web-sg
spec:
  description: Web server security group
  rules:
    - key: allow-ssh
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 22
      portRangeMax: 22
      remoteIpPrefix: "0.0.0.0/0"
      description: Allow SSH from anywhere
    - key: allow-http
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 80
      portRangeMax: 80
      remoteIpPrefix: "0.0.0.0/0"
    - key: allow-https
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 443
      portRangeMax: 443
      remoteIpPrefix: "0.0.0.0/0"
```

### Zero-Trust Security Group with Explicit Egress

A locked-down security group that deletes the default egress rules and explicitly allows only the required outbound traffic:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: db-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackSecurityGroup.db-sg
spec:
  description: Zero-trust database security group
  deleteDefaultRules: true
  rules:
    - key: allow-postgres-from-app-subnet
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 5432
      portRangeMax: 5432
      remoteIpPrefix: "10.0.1.0/24"
      description: PostgreSQL from app subnet only
    - key: egress-dns
      direction: egress
      ethertype: IPv4
      protocol: udp
      portRangeMin: 53
      portRangeMax: 53
      remoteIpPrefix: "0.0.0.0/0"
      description: Allow DNS resolution
    - key: egress-https
      direction: egress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 443
      portRangeMax: 443
      remoteIpPrefix: "0.0.0.0/0"
      description: Allow HTTPS for package updates
  tags:
    - production
    - database
```

### Full-Featured Security Group with Mixed Rule Types

A production security group with inter-group references, ICMP rules, a port range, stateful mode, tags, and a region override:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackSecurityGroup
metadata:
  name: app-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenStackSecurityGroup.app-sg
spec:
  description: Application tier security group with mixed rule types
  stateful: true
  region: RegionOne
  tags:
    - production
    - app-tier
  rules:
    - key: allow-http-from-lb
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 8080
      portRangeMax: 8080
      remoteGroupId: a1b2c3d4-e5f6-7890-abcd-ef1234567890
      description: HTTP from load balancer security group
    - key: allow-ephemeral-from-lb
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 30000
      portRangeMax: 32767
      remoteGroupId: a1b2c3d4-e5f6-7890-abcd-ef1234567890
      description: Ephemeral port range from load balancer
    - key: allow-icmp-echo-request
      direction: ingress
      ethertype: IPv4
      protocol: icmp
      portRangeMin: 8
      portRangeMax: 0
      remoteIpPrefix: "10.0.0.0/8"
      description: ICMP echo request from internal network
    - key: allow-ssh-from-bastion
      direction: ingress
      ethertype: IPv4
      protocol: tcp
      portRangeMin: 22
      portRangeMax: 22
      remoteIpPrefix: "10.0.100.0/24"
      description: SSH from bastion subnet
    - key: egress-all-ipv4
      direction: egress
      ethertype: IPv4
      remoteIpPrefix: "0.0.0.0/0"
      description: Allow all outbound IPv4 traffic
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `security_group_id` | `string` | UUID of the created security group in OpenStack. Primary output used as a foreign key by downstream components. |
| `name` | `string` | Name of the security group, derived from `metadata.name`. |
| `region` | `string` | OpenStack region where the security group was created. |

## Related Components

- [OpenStackSecurityGroupRule](/docs/catalog/openstack/openstacksecuritygrouprule) — manages standalone security group rules with full foreign key support for DAG-visible rule management in InfraCharts
- [OpenStackInstance](/docs/catalog/openstack/openstackinstance) — references security groups via the `securityGroups` field
- [OpenStackNetworkPort](/docs/catalog/openstack/openstacknetworkport) — references security groups via the `securityGroupIds` field
- [OpenStackNetwork](/docs/catalog/openstack/openstacknetwork) — provides the network layer that security groups protect

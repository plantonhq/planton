# Scaleway Instance Security Group

Deploys a Scaleway Instance Security Group — a zonal stateful firewall that controls inbound and outbound traffic to Scaleway Instances. A single manifest creates the security group with inline rules and default policies, ready to be referenced by ScalewayInstance resources.

## What Gets Created

When you deploy a ScalewayInstanceSecurityGroup resource, OpenMCF provisions:

- **Instance Security Group** — an `instance.SecurityGroup` resource in the specified zone with configurable default policies for inbound/outbound traffic, stateful connection tracking, SMTP security settings, and ordered firewall rules
- **Inbound Rules** — inline `SecurityGroupInboundRule` entries that match traffic by protocol, port range, and source IP, applying accept or drop actions in evaluation order
- **Outbound Rules** — inline `SecurityGroupOutboundRule` entries that match traffic by protocol, port range, and destination IP, applying accept or drop actions in evaluation order

## Prerequisites

- **Scaleway credentials** configured via environment variables or OpenMCF provider config
- **A target zone** where the security group will be created (e.g., `fr-par-1`, `nl-ams-1`). The zone must match the zone of the Instances that will use this security group.

## Quick Start

Create a file `security-group.yaml`:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: web-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayInstanceSecurityGroup.web-sg
spec:
  zone: fr-par-1
  stateful: true
  inboundDefaultPolicy: drop
  outboundDefaultPolicy: accept
  inboundRules:
    - action: accept
      protocol: TCP
      portRange: "443"
      ipRange: 0.0.0.0/0
    - action: accept
      protocol: TCP
      portRange: "80"
      ipRange: 0.0.0.0/0
```

Deploy:

```shell
openmcf apply -f security-group.yaml
```

This creates a stateful security group in `fr-par-1` that drops all inbound traffic except HTTP and HTTPS from any source, while allowing all outbound traffic.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zone` | `string` | Scaleway zone for the security group (e.g., `"fr-par-1"`, `"nl-ams-1"`, `"pl-waw-1"`). Must match the zone of Instances using this group. Cannot be changed after creation. | Required |
| `inboundRules[].action` | `string` | Action when traffic matches the rule. Values: `"accept"` or `"drop"`. | Required, pattern `^(accept\|drop)$` |
| `outboundRules[].action` | `string` | Action when traffic matches the rule. Values: `"accept"` or `"drop"`. | Required, pattern `^(accept\|drop)$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the security group's purpose. |
| `stateful` | `bool` | `true` | When `true`, return traffic for accepted connections is automatically permitted. Set to `false` for stateless routing or network appliance use cases. |
| `inboundDefaultPolicy` | `string` | `"accept"` | Default policy for inbound traffic matching no rule. `"accept"` for a denylist model, `"drop"` for an allowlist model. |
| `outboundDefaultPolicy` | `string` | `"accept"` | Default policy for outbound traffic matching no rule. `"accept"` for a denylist model, `"drop"` for an allowlist model. |
| `enableDefaultSecurity` | `bool` | `true` | When `true`, blocks outbound SMTP on ports 25, 465, and 587. Set to `false` only if your Scaleway account is authorized for SMTP sending. |
| `inboundRules` | `list` | `[]` | Ordered list of inbound firewall rules. First matching rule wins. |
| `inboundRules[].protocol` | `string` | `"TCP"` | IP protocol. Values: `"TCP"`, `"UDP"`, `"ICMP"`, `"ANY"`. |
| `inboundRules[].portRange` | `string` | all ports | Single port (`"80"`) or range (`"8000-9000"`). Ignored when protocol is `"ICMP"`. |
| `inboundRules[].ipRange` | `string` | all IPs | Source IP range in CIDR notation (e.g., `"0.0.0.0/0"`, `"10.0.0.0/8"`). |
| `outboundRules` | `list` | `[]` | Ordered list of outbound firewall rules. First matching rule wins. |
| `outboundRules[].protocol` | `string` | `"TCP"` | IP protocol. Values: `"TCP"`, `"UDP"`, `"ICMP"`, `"ANY"`. |
| `outboundRules[].portRange` | `string` | all ports | Single port (`"443"`) or range (`"1024-65535"`). Ignored when protocol is `"ICMP"`. |
| `outboundRules[].ipRange` | `string` | all IPs | Destination IP range in CIDR notation (e.g., `"0.0.0.0/0"`, `"10.0.0.0/8"`). |

## Examples

### Allowlist Web Server

A security group for a web-facing Instance that drops all inbound traffic except SSH from a bastion IP, plus HTTP and HTTPS from any source:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: web-allowlist
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayInstanceSecurityGroup.web-allowlist
spec:
  zone: fr-par-1
  stateful: true
  inboundDefaultPolicy: drop
  outboundDefaultPolicy: accept
  enableDefaultSecurity: true
  inboundRules:
    - action: accept
      protocol: TCP
      portRange: "22"
      ipRange: 203.0.113.10/32
    - action: accept
      protocol: TCP
      portRange: "80"
      ipRange: 0.0.0.0/0
    - action: accept
      protocol: TCP
      portRange: "443"
      ipRange: 0.0.0.0/0
```

### Denylist with Blocked Ranges

A security group that allows all traffic by default but drops inbound connections from specific CIDR blocks and blocks outbound access to a known bad range:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: denylist-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.ScalewayInstanceSecurityGroup.denylist-sg
spec:
  zone: nl-ams-1
  stateful: true
  inboundDefaultPolicy: accept
  outboundDefaultPolicy: accept
  inboundRules:
    - action: drop
      protocol: ANY
      ipRange: 198.51.100.0/24
    - action: drop
      protocol: ANY
      ipRange: 192.0.2.0/24
  outboundRules:
    - action: drop
      protocol: ANY
      ipRange: 203.0.113.0/24
```

### Strict Egress Control

A locked-down security group for a database Instance that only permits inbound PostgreSQL connections from a private subnet and restricts outbound traffic to DNS and NTP:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayInstanceSecurityGroup
metadata:
  name: db-strict
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayInstanceSecurityGroup.db-strict
spec:
  zone: fr-par-1
  description: "Database tier -- PostgreSQL inbound from app subnet, restricted egress"
  stateful: true
  inboundDefaultPolicy: drop
  outboundDefaultPolicy: drop
  enableDefaultSecurity: true
  inboundRules:
    - action: accept
      protocol: TCP
      portRange: "5432"
      ipRange: 10.0.1.0/24
    - action: accept
      protocol: ICMP
      ipRange: 10.0.0.0/8
  outboundRules:
    - action: accept
      protocol: UDP
      portRange: "53"
      ipRange: 0.0.0.0/0
    - action: accept
      protocol: UDP
      portRange: "123"
      ipRange: 0.0.0.0/0

```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `securityGroupId` | `string` | Zoned ID of the created security group (format: `{zone}/{uuid}`, e.g., `fr-par-1/11111111-1111-...`). Referenced by ScalewayInstance resources via the `securityGroupId` field. |

## Related Components

- [ScalewayInstance](/docs/catalog/scaleway/scalewayinstance) — compute Instance that references this security group via its `securityGroupId` field
- [ScalewayPrivateNetwork](/docs/catalog/scaleway/scalewayprivatenetwork) — provides the Private Network for Instance connectivity; security groups operate independently of Private Networks
- [ScalewayPublicGateway](/docs/catalog/scaleway/scalewaypublicgateway) — manages NAT and DHCP for Instances on a Private Network, complementing security group rules for egress control

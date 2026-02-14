# Civo Firewall

Deploys a stateful firewall on Civo Cloud within a specified network (VPC), with configurable inbound and outbound rules. The component supports protocol-level filtering, CIDR-based source/destination restrictions, and instance tag selectors for automatic firewall association.

## What Gets Created

When you deploy a CivoFirewall resource, OpenMCF provisions:

- **Civo Firewall** — a `civo_firewall` resource attached to the target private network, configured with the specified ingress and egress rules
- **Ingress Rules** — one rule entry per item in `inboundRules`, controlling traffic allowed to instances behind the firewall
- **Egress Rules** — one rule entry per item in `outboundRules`, controlling traffic allowed from instances behind the firewall

## Prerequisites

- **Civo credentials** configured via environment variables or OpenMCF provider config
- **An existing Civo network** in the target region (can be created with CivoVpc)

## Quick Start

Create a file `civo-firewall.yaml`:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoFirewall
metadata:
  name: my-firewall
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoFirewall.my-firewall
spec:
  name: my-firewall
  networkId:
    value: network-uuid-here
  inboundRules:
    - protocol: tcp
      portRange: "443"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: allow-https
```

Deploy:

```shell
openmcf apply -f civo-firewall.yaml
```

This creates a firewall on the specified network that allows inbound HTTPS traffic from any source.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Name of the firewall. Must be unique per Civo account/project. | Required |
| `networkId` | `StringValueOrRef` | The network (VPC) in which to create this firewall. Can reference a CivoVpc resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `inboundRules` | `CivoFirewallInboundRule[]` | `[]` | Inbound (ingress) rules defining traffic allowed to instances. Any traffic not matching a rule is denied. |
| `outboundRules` | `CivoFirewallOutboundRule[]` | `[]` | Outbound (egress) rules defining traffic allowed from instances. If no egress rules are specified, all outbound traffic is allowed by default. |
| `tags` | `string[]` | `[]` | Instance tag selectors. Any instance in the same network with a matching tag will use this firewall automatically. |

#### CivoFirewallInboundRule

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `protocol` | `string` | — | Protocol to allow. Must be `tcp`, `udp`, or `icmp`. |
| `portRange` | `string` | — | Port or port range (e.g., `"80"`, `"443"`, `"8000-9000"`). Leave empty or use `"1-65535"` for all ports. Applies to `tcp` and `udp` only. |
| `cidrs` | `string[]` | `["0.0.0.0/0"]` | CIDR blocks from which traffic is permitted. |
| `action` | `string` | `"allow"` | Action for this rule: `"allow"` or `"deny"`. |
| `label` | `string` | `""` | Optional human-readable label for the rule. |

#### CivoFirewallOutboundRule

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `protocol` | `string` | — | Protocol to allow or deny. Must be `tcp`, `udp`, or `icmp`. |
| `portRange` | `string` | — | Port or port range (format as in inbound rules). |
| `cidrs` | `string[]` | `["0.0.0.0/0"]` | CIDR blocks to which traffic is permitted. |
| `action` | `string` | `"allow"` | Action for this rule: `"allow"` or `"deny"`. |
| `label` | `string` | `""` | Optional human-readable label for the rule. |

## Examples

### Allow SSH and HTTPS Only

A basic firewall that permits SSH and HTTPS inbound traffic:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoFirewall
metadata:
  name: web-ssh-firewall
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoFirewall.web-ssh-firewall
spec:
  name: web-ssh-firewall
  networkId:
    value: network-uuid-here
  inboundRules:
    - protocol: tcp
      portRange: "22"
      cidrs:
        - "203.0.113.0/24"
      action: allow
      label: allow-ssh-office
    - protocol: tcp
      portRange: "443"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: allow-https
```

### Restricted Egress with Tag-Based Association

A firewall that restricts both inbound and outbound traffic, auto-applied to instances tagged `backend`:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoFirewall
metadata:
  name: backend-firewall
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoFirewall.backend-firewall
spec:
  name: backend-firewall
  networkId:
    value: network-uuid-here
  inboundRules:
    - protocol: tcp
      portRange: "8080"
      cidrs:
        - "10.0.0.0/8"
      action: allow
      label: allow-internal-api
  outboundRules:
    - protocol: tcp
      portRange: "443"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: allow-https-out
    - protocol: tcp
      portRange: "5432"
      cidrs:
        - "10.0.0.0/8"
      action: allow
      label: allow-postgres-out
  tags:
    - backend
```

### Using Foreign Key References

Reference an OpenMCF-managed CivoVpc resource instead of hardcoding the network ID:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoFirewall
metadata:
  name: ref-firewall
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoFirewall.ref-firewall
spec:
  name: ref-firewall
  networkId:
    valueFrom:
      kind: CivoVpc
      name: my-network
      field: status.outputs.network_id
  inboundRules:
    - protocol: tcp
      portRange: "80"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: allow-http
    - protocol: tcp
      portRange: "443"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: allow-https
    - protocol: icmp
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: allow-ping
  tags:
    - web
    - public
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `firewallId` | `string` | Unique identifier (UUID) of the created firewall, assigned by Civo |
| `createdAtRfc3339` | `string` | Timestamp when the firewall was created, in RFC 3339 format |

## Related Components

- [CivoVpc](/docs/catalog/civo/civovpc) — provides the private network to which the firewall is attached
- [CivoComputeInstance](/docs/catalog/civo/civocomputeinstance) — compute instances protected by the firewall
- [CivoKubernetesCluster](/docs/catalog/civo/civokubernetescluster) — Kubernetes clusters that can use the firewall for network access control
- [CivoDatabase](/docs/catalog/civo/civodatabase) — managed database instances that reference the firewall for access restriction

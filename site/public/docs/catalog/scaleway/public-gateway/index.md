---
title: "Public Gateway"
description: "Public Gateway deployment documentation"
icon: "package"
order: 100
componentName: "scalewaypublicgateway"
---

# Scaleway Public Gateway

Deploys a Scaleway Public Gateway with a dedicated Flexible IP, GatewayNetwork attachment, and optional PAT (port forwarding) rules. Provides NAT masquerade for outbound internet access, SSH bastion for secure private resource access, and port-level ingress routing -- all as a single declarative resource.

## What Gets Created

When you deploy a ScalewayPublicGateway resource, Planton provisions:

- **Flexible IP** — a `network.PublicGatewayIp` resource providing a dedicated public IPv4 address for the gateway, managed independently so it survives gateway replacements
- **Public Gateway** — a `network.PublicGateway` resource, the managed network appliance that performs NAT, SSH bastion proxying, and port forwarding
- **GatewayNetwork** — a `network.GatewayNetwork` resource binding the gateway to the target Private Network with the configured masquerade setting
- **PAT Rules** — one `network.PublicGatewayPatRule` resource per entry in the `patRules` list, mapping public ports on the gateway IP to private IP:port pairs inside the attached network

## Prerequisites

- **Scaleway credentials** configured via environment variables or Planton provider config
- **A Private Network** in the target region (can be created via a ScalewayPrivateNetwork resource or referenced by UUID)
- **Zone within the Private Network's region** — Public Gateways are zonal; the zone must belong to the same region as the target Private Network (e.g., `fr-par-1` for a network in `fr-par`)

## Quick Start

Create a file `public-gateway.yaml`:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayPublicGateway
metadata:
  name: my-gateway
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayPublicGateway.my-gateway
spec:
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  zone: fr-par-1
  type: VPC-GW-S
  enableMasquerade: true
```

Deploy:

```shell
planton apply -f public-gateway.yaml
```

This creates a standard Public Gateway with NAT masquerade enabled, giving all resources in the attached Private Network outbound internet access through a single public IP.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `privateNetworkId` | `StringValueOrRef` | Private Network UUID to attach the gateway to. Can reference a ScalewayPrivateNetwork resource via `valueFrom`. | Required |
| `zone` | `string` | Scaleway zone for the gateway (e.g., `"fr-par-1"`, `"nl-ams-1"`, `"pl-waw-1"`). Must be within the same region as the target Private Network. Cannot be changed after creation. | Required |
| `type` | `string` | Gateway type determining bandwidth tier. Options: `"VPC-GW-S"` (standard), `"VPC-GW-XL"` (high-bandwidth, Paris region only). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enableMasquerade` | `bool` | `true` | Enables NAT masquerade so Private Network resources can reach the internet through the gateway's public IP. Disable only if the gateway is used solely as an SSH bastion. |
| `bastion` | `object` | — | SSH bastion configuration. When configured, the gateway acts as a jump host for SSH connections to private resources. |
| `bastion.enabled` | `bool` | `false` | Enables the SSH bastion feature on the gateway. |
| `bastion.port` | `int32` | `22` | Port the SSH bastion listens on. Change if port 22 is blocked by network policies. |
| `bastion.allowedIpRanges` | `string[]` | `[]` | CIDR ranges allowed to connect to the bastion (e.g., `"203.0.113.0/24"`). If empty, all source IPs are allowed. |
| `enableSmtp` | `bool` | `false` | Enables outbound SMTP (port 25) through the gateway. Blocked by default to prevent spam abuse. |
| `reverseDns` | `string` | — | Reverse DNS hostname for the gateway's public IP (e.g., `"gateway.example.com"`). A matching DNS A record must already exist. |
| `patRules` | `object[]` | `[]` | Port forwarding rules mapping public ports to private IP:port pairs. |
| `patRules[].privateIp` | `string` | — | Target private IP address within the attached network. Required per rule. |
| `patRules[].privatePort` | `int32` | — | Target port on the private IP. Required per rule. |
| `patRules[].publicPort` | `int32` | — | Public port on the gateway IP to listen on. Required per rule. |
| `patRules[].protocol` | `string` | `"both"` | Protocol for the rule: `"tcp"`, `"udp"`, or `"both"`. |

## Examples

### NAT Gateway for a Private Network

A standard gateway providing outbound internet access for private resources, referencing an Planton-managed Private Network:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayPublicGateway
metadata:
  name: nat-gateway
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayPublicGateway.nat-gateway
spec:
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  zone: fr-par-1
  type: VPC-GW-S
  enableMasquerade: true
```

### SSH Bastion with IP Restrictions

A gateway configured as an SSH jump host with bastion access restricted to specific CIDR ranges:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayPublicGateway
metadata:
  name: bastion-gateway
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayPublicGateway.bastion-gateway
spec:
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  zone: nl-ams-1
  type: VPC-GW-S
  enableMasquerade: true
  bastion:
    enabled: true
    port: 61022
    allowedIpRanges:
      - 203.0.113.0/24
      - 198.51.100.10/32
```

### Full Configuration with Port Forwarding and Reverse DNS

A production gateway with NAT, SSH bastion, PAT rules exposing internal services, SMTP enabled for an email relay, and reverse DNS configured:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayPublicGateway
metadata:
  name: prod-gateway
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayPublicGateway.prod-gateway
spec:
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: prod-network
      fieldPath: status.outputs.private_network_id
  zone: fr-par-1
  type: VPC-GW-S
  enableMasquerade: true
  enableSmtp: true
  reverseDns: gateway.example.com
  bastion:
    enabled: true
    port: 22
    allowedIpRanges:
      - 203.0.113.0/24
  patRules:
    - privateIp: 10.0.1.5
      privatePort: 80
      publicPort: 8080
      protocol: tcp
    - privateIp: 10.0.1.10
      privatePort: 5432
      publicPort: 15432
      protocol: tcp
    - privateIp: 10.0.1.20
      privatePort: 53
      publicPort: 5353
      protocol: both
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `gateway_id` | `string` | Zoned ID of the created Public Gateway (e.g., `"fr-par-1/xxxxxxxx-xxxx-..."`). Used to reference the gateway in other resources. |
| `public_ip_address` | `string` | Public IPv4 address assigned to the gateway. Use for DNS A records, firewall allowlists, and connectivity diagnostics. |
| `public_ip_id` | `string` | Zoned ID of the Flexible IP resource. Useful for reassigning the IP to a replacement gateway without changing the public address. |
| `gateway_network_id` | `string` | Zoned ID of the GatewayNetwork attachment binding the gateway to the Private Network. |

## Related Components

- [ScalewayPrivateNetwork](/docs/catalog/scaleway/private-network) — provides the Private Network that the gateway attaches to for NAT, bastion, and port forwarding
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/kapsule-cluster) — deploys Kubernetes clusters whose nodes use the gateway for outbound internet access
- [ScalewayRdbInstance](/docs/catalog/scaleway/rdb-instance) — deploys managed databases reachable through the gateway's port forwarding rules
- [ScalewayInstanceSecurityGroup](/docs/catalog/scaleway/instance-security-group) — controls network access for compute instances behind the gateway

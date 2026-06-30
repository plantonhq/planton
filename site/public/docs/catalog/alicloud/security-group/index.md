---
title: "Security Group"
description: "Security Group deployment documentation"
icon: "package"
order: 100
componentName: "alicloudsecuritygroup"
---

# AliCloud Security Group

Deploys an Alibaba Cloud Security Group with bundled security rules in a VPC. The component provisions the security group and its ingress/egress rules as a single atomic unit, ensuring the firewall is always created with its intended access policy.

## What Gets Created

When you deploy an AliCloudSecurityGroup resource, Planton provisions:

- **Security Group** -- an `alicloud_security_group` resource bound to the specified VPC with configurable inner access policy and tags
- **Security Group Rules** -- one `alicloud_security_group_rule` per entry in `rules`, defining ingress and egress traffic policies

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or Planton provider config
- **An Alibaba Cloud VPC** -- the security group must belong to a VPC (create one with AliCloudVpc)

## Quick Start

Create a file `security-group.yaml`:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudSecurityGroup
metadata:
  name: my-web-sg
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AliCloudSecurityGroup.my-web-sg
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: my-vpc
  securityGroupName: web-sg
  description: Security group for web-facing instances
  rules:
    - type: ingress
      ipProtocol: tcp
      portRange: "443/443"
      cidrIp: "0.0.0.0/0"
      description: Allow HTTPS from anywhere
    - type: egress
      ipProtocol: all
      portRange: "-1/-1"
      cidrIp: "0.0.0.0/0"
      description: Allow all outbound traffic
```

Deploy:

```shell
planton apply -f security-group.yaml
```

This creates a security group that allows HTTPS inbound and all outbound traffic.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for provider endpoint (e.g., `cn-hangzhou`, `us-west-1`). | Required; non-empty |
| `vpcId` | `StringValueOrRef` | VPC ID this security group belongs to. Can be a literal string or a reference to an AliCloudVpc output. | Required |
| `securityGroupName` | `string` | Security group name. Must start with a letter; can contain Unicode, digits, colons, underscores, periods, hyphens. | Required; 2-128 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the security group's purpose. |
| `innerAccessPolicy` | `string` | `"Accept"` | Controls intra-group traffic. `Accept` allows free communication between instances in the same SG. `Drop` requires explicit rules for intra-group traffic. |
| `resourceGroupId` | `string` | `""` | Resource group ID for organizational grouping. |
| `tags` | `map<string, string>` | `{}` | Tags applied to the security group. Merged with standard Planton tags. |
| `rules` | `list` | `[]` | Security group rules. Each rule creates a separate rule resource. |
| `rules[].type` | `string` | -- | `ingress` for inbound, `egress` for outbound. Required. |
| `rules[].ipProtocol` | `string` | -- | Protocol: `tcp`, `udp`, `icmp`, `gre`, `all`. Required. |
| `rules[].portRange` | `string` | `"-1/-1"` | Port range in `start/end` format. TCP/UDP: explicit ports required. ICMP/GRE/ALL: must be `"-1/-1"`. |
| `rules[].cidrIp` | `string` | `""` | IPv4 CIDR for source (ingress) or destination (egress). |
| `rules[].sourceSecurityGroupId` | `string` | `""` | Source/destination security group ID for SG-to-SG rules. |
| `rules[].priority` | `int` | `1` | Evaluation priority (1-100, lower = higher priority). |
| `rules[].policy` | `string` | `"accept"` | Action: `accept` or `drop`. |
| `rules[].description` | `string` | `""` | Rule description (only field updatable without recreation). |

## Examples

### Web Tier with HTTP/HTTPS Ingress

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudSecurityGroup
metadata:
  name: web-tier-sg
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AliCloudSecurityGroup.web-tier-sg
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: prod-vpc
  securityGroupName: web-tier
  description: Public web tier allowing HTTP/HTTPS
  tags:
    tier: web
  rules:
    - type: ingress
      ipProtocol: tcp
      portRange: "80/80"
      cidrIp: "0.0.0.0/0"
      description: Allow HTTP
    - type: ingress
      ipProtocol: tcp
      portRange: "443/443"
      cidrIp: "0.0.0.0/0"
      description: Allow HTTPS
    - type: egress
      ipProtocol: all
      portRange: "-1/-1"
      cidrIp: "0.0.0.0/0"
      description: Allow all outbound
```

### Database Tier Restricted to VPC

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudSecurityGroup
metadata:
  name: db-tier-sg
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AliCloudSecurityGroup.db-tier-sg
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: prod-vpc
  securityGroupName: db-tier
  description: Database tier restricted to VPC traffic
  innerAccessPolicy: Drop
  rules:
    - type: ingress
      ipProtocol: tcp
      portRange: "3306/3306"
      cidrIp: "10.0.0.0/8"
      description: Allow MySQL from VPC
    - type: ingress
      ipProtocol: tcp
      portRange: "5432/5432"
      cidrIp: "10.0.0.0/8"
      description: Allow PostgreSQL from VPC
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `security_group_id` | `string` | The security group ID assigned by Alibaba Cloud |
| `security_group_name` | `string` | The security group name as created |

## Related Components

- [AliCloudVpc](/docs/catalog/alicloud/vpc) -- create the VPC this security group belongs to
- [AliCloudEcsInstance](/docs/catalog/alicloud/ecsinstance) -- associate ECS instances with this security group
- [AliCloudAckManagedCluster](/docs/catalog/alicloud/alicloudackmanagedcluster) -- use this security group for ACK worker nodes

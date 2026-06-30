---
title: "VSwitch"
description: "VSwitch deployment documentation"
icon: "package"
order: 100
componentName: "alicloudvswitch"
---

# AliCloud VSwitch

Deploys an Alibaba Cloud VSwitch (subnet) within an existing VPC, bound to a single Availability Zone with a dedicated IPv4 CIDR block, optional IPv6 dual-stack support, and automatic tag management. The VSwitch is the mandatory network placement target for ECS instances, databases, Kubernetes clusters, NAT gateways, and load balancers on Alibaba Cloud.

## What Gets Created

When you deploy an AliCloudVswitch resource, Planton provisions:

- **VSwitch** — an `alicloud_vswitch` resource (Pulumi: `vpc.Switch`) with the specified VPC, Availability Zone, CIDR block, name, and optional IPv6 configuration
- **Tags** — system metadata tags (`resource`, `resource_name`, `resource_kind`, `organization`, `environment`) merged with user-defined `spec.tags`, with user values taking precedence on key conflict

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or Planton provider config
- **An existing VPC** — the VSwitch's `vpcId` must reference a valid Alibaba Cloud VPC; can be a literal ID or a `valueFrom` reference to an AliCloudVpc component
- **CIDR block planning** — the VSwitch CIDR must be a subset of the parent VPC's CIDR block, with a mask length of 16-29, and must not overlap with other VSwitches in the same VPC
- **Planton CLI** installed with either Pulumi or Terraform (OpenTofu) backend

## Quick Start

Create a file `vswitch.yaml`:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudVswitch
metadata:
  name: my-vswitch
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AliCloudVswitch.my-vswitch
spec:
  region: cn-hangzhou
  vpcId: vpc-bp1234567890abcdef
  zoneId: cn-hangzhou-a
  cidrBlock: "10.0.0.0/24"
  vswitchName: my-vswitch
```

Deploy:

```shell
planton apply -f vswitch.yaml
```

This creates a VSwitch with a `/24` CIDR block in the `cn-hangzhou-a` Availability Zone within the specified VPC.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region where the VSwitch will be created (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`). Must match the parent VPC's region. | Required; non-empty |
| `vpcId` | `StringValueOrRef` | VPC ID that this VSwitch belongs to. The VSwitch's CIDR block must fall within the VPC's CIDR range. Can reference an AliCloudVpc resource via `valueFrom`. | Required; default kind: `AliCloudVpc`, field: `status.outputs.vpc_id` |
| `zoneId` | `string` | Availability Zone for the VSwitch (e.g., `cn-hangzhou-a`, `cn-hangzhou-b`). The VSwitch is permanently bound to this zone. | Required; non-empty |
| `cidrBlock` | `string` | IPv4 CIDR block for the VSwitch. Must be a subnet of the parent VPC's CIDR block with a mask length of 16-29. Cannot be changed after creation. | Required; non-empty |
| `vswitchName` | `string` | VSwitch name. Cannot start with `http://` or `https://`. | Required; 1-128 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the VSwitch. Cannot start with `http://` or `https://`. |
| `enableIpv6` | `bool` | `false` | When `true`, allocates an IPv6 CIDR block to this VSwitch. The parent VPC must have IPv6 enabled (`enableIpv6: true` on the AliCloudVpc). |
| `ipv6CidrBlockMask` | `int32` | `0` | Selects a `/64` IPv6 segment from the parent VPC's `/56` IPv6 allocation. Only meaningful when `enableIpv6` is `true` on both the VPC and this VSwitch. | 
| `tags` | `map<string, string>` | `{}` | User-defined key-value tags applied to the VSwitch. Merged with system tags; user values take precedence on key conflict. |

## Examples

### Development Single-Zone VSwitch

A minimal VSwitch for non-production workloads with a standard `/24` CIDR block.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudVswitch
metadata:
  name: dev-vswitch
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AliCloudVswitch.dev-vswitch
spec:
  region: cn-hangzhou
  vpcId: vpc-abc123def456
  zoneId: cn-hangzhou-a
  cidrBlock: "192.168.0.0/24"
  vswitchName: dev-vswitch
```

### Production VSwitch with Tags

A production VSwitch with a large address space for Kubernetes node pools, organizational metadata, and cost-tracking tags.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudVswitch
metadata:
  name: prod-app-vswitch
  org: my-org
  env: production
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AliCloudVswitch.prod-app-vswitch
spec:
  region: cn-shanghai
  vpcId: vpc-prod-001
  zoneId: cn-shanghai-b
  cidrBlock: "10.1.0.0/20"
  vswitchName: prod-app-tier-b
  description: Application tier VSwitch in zone B for Kubernetes workers
  tags:
    team: platform
    costCenter: engineering
    tier: application
```

### VSwitch with Foreign Key Reference

References an AliCloudVpc component instead of hardcoding the VPC ID, establishing a declarative dependency between the VSwitch and its parent VPC.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudVswitch
metadata:
  name: db-vswitch
  env: staging
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.AliCloudVswitch.db-vswitch
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: my-vpc
  zoneId: cn-hangzhou-c
  cidrBlock: "10.2.0.0/24"
  vswitchName: staging-db-vswitch
  description: Database tier VSwitch for RDS and Redis instances
```

### IPv6-Enabled Dual-Stack VSwitch

A VSwitch with IPv6 support. The parent VPC must have IPv6 enabled for the IPv6 CIDR allocation to succeed.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudVswitch
metadata:
  name: ipv6-vswitch
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.AliCloudVswitch.ipv6-vswitch
spec:
  region: us-west-1
  vpcId: vpc-ipv6-enabled
  zoneId: us-west-1a
  cidrBlock: "172.16.0.0/24"
  vswitchName: ipv6-app-vswitch
  description: Dual-stack VSwitch for IPv6 workloads
  enableIpv6: true
  ipv6CidrBlockMask: 42
  tags:
    networkType: dual-stack
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vswitch_id` | `string` | The VSwitch ID assigned by Alibaba Cloud. Referenced by downstream components (AliCloudNatGateway, AliCloudEcsInstance, AliCloudAckManagedCluster, AliCloudRdsInstance, etc.) via `StringValueOrRef`. |
| `vswitch_name` | `string` | The VSwitch name as created. |
| `cidr_block` | `string` | The IPv4 CIDR block of the VSwitch. Useful for security group rules and CIDR planning. |
| `zone_id` | `string` | The Availability Zone in which the VSwitch resides. |
| `ipv6_cidr_block` | `string` | The IPv6 CIDR block allocated to the VSwitch. Only populated when IPv6 is enabled on both the parent VPC and this VSwitch. |

## Related Components

- [AliCloudVpc](/docs/catalog/alicloud/vpc) — the parent VPC that this VSwitch belongs to
- [AliCloudSecurityGroup](/docs/catalog/alicloud/security-group) — controls network traffic for resources deployed in this VSwitch
- [AliCloudNatGateway](/docs/catalog/alicloud/nat-gateway) — provides outbound internet access for resources in private VSwitches
- [AliCloudEcsInstance](/docs/catalog/alicloud/ecsinstance) — launches compute instances in this VSwitch
- [AliCloudAckManagedCluster](/docs/catalog/alicloud/alicloudackmanagedcluster) — deploys managed Kubernetes clusters using this VSwitch for node placement

# Alibaba Cloud VPC

Deploys an Alibaba Cloud Virtual Private Cloud with a configurable IPv4 CIDR block, optional IPv6 dual-stack support, resource group assignment, and automatic tag management. The VPC is the networking foundation for VSwitches, security groups, NAT gateways, load balancers, and Kubernetes clusters on Alibaba Cloud.

## What Gets Created

When you deploy an AlicloudVpc resource, OpenMCF provisions:

- **VPC** — an `alicloud_vpc` resource (Pulumi: `vpc.Network`) with the specified CIDR block, name, and optional IPv6 configuration
- **VRouter** — automatically created by Alibaba Cloud as part of VPC creation, responsible for routing traffic between VSwitches and managing route tables
- **System Route Table** — the default route table associated with the VRouter, containing system routes for intra-VPC communication
- **Tags** — system metadata tags (`resource`, `resource_name`, `resource_kind`, `organization`, `environment`) merged with user-defined `spec.tags`, with user values taking precedence on key conflict

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or OpenMCF provider config
- **CIDR block planning** — the primary IPv4 CIDR cannot be changed after creation; choose a range that accommodates future growth and avoids overlap with other VPCs if using VPC peering or CEN
- **OpenMCF CLI** installed with either Pulumi or Terraform (OpenTofu) backend

## Quick Start

Create a file `vpc.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVpc
metadata:
  name: my-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudVpc.my-vpc
spec:
  region: cn-hangzhou
  vpcName: my-vpc
  cidrBlock: "10.0.0.0/16"
```

Deploy:

```shell
openmcf apply -f vpc.yaml
```

This creates a VPC with a `/16` CIDR block in the `cn-hangzhou` region. Alibaba Cloud auto-creates a VRouter and system route table as part of the VPC.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region where the VPC will be created (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`, `ap-southeast-1`). | Required; non-empty |
| `vpcName` | `string` | VPC name. Cannot start with `http://` or `https://`. | Required; 1-128 characters |
| `cidrBlock` | `string` | Primary IPv4 CIDR block. Must be a private range (`10.0.0.0/8`, `172.16.0.0/12`, or `192.168.0.0/16`) with a mask length of 8-28. Cannot be changed after creation. | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the VPC. 1-256 characters; cannot start with `http://` or `https://`. |
| `enableIpv6` | `bool` | `false` | When `true`, Alibaba Cloud allocates a `/56` IPv6 CIDR block to the VPC. VSwitches can then be assigned IPv6 subnets. |
| `resourceGroupId` | `string` | `""` | Alibaba Cloud resource group ID for organizational grouping, access control, and cost attribution. If omitted, the VPC is placed in the account's default resource group. |
| `tags` | `map<string, string>` | `{}` | User-defined key-value tags applied to the VPC. Merged with system tags; user values take precedence on key conflict. |

## Examples

### Development VPC

A minimal VPC for non-production workloads with the smallest standard private CIDR range.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVpc
metadata:
  name: dev-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudVpc.dev-vpc
spec:
  region: cn-hangzhou
  vpcName: dev-vpc
  cidrBlock: "192.168.0.0/16"
```

### Production VPC with Tags and Resource Group

A production VPC with a large address space, resource group assignment for access control and billing, and organizational tags.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVpc
metadata:
  name: prod-vpc
  org: my-org
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AlicloudVpc.prod-vpc
spec:
  region: cn-shanghai
  vpcName: prod-platform-vpc
  cidrBlock: "10.0.0.0/8"
  description: Production VPC for platform workloads
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
```

### IPv6-Enabled VPC

A dual-stack VPC with IPv6 support enabled at creation time. Alibaba Cloud allocates a `/56` IPv6 CIDR block automatically.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVpc
metadata:
  name: ipv6-vpc
  env: staging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AlicloudVpc.ipv6-vpc
spec:
  region: us-west-1
  vpcName: ipv6-enabled-vpc
  cidrBlock: "172.16.0.0/12"
  description: Dual-stack VPC with IPv6 support
  enableIpv6: true
  resourceGroupId: rg-staging-456
  tags:
    networkType: dual-stack
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vpc_id` | `string` | The VPC ID assigned by Alibaba Cloud. Referenced by downstream components (AlicloudVswitch, AlicloudSecurityGroup, AlicloudNatGateway, etc.) via `StringValueOrRef`. |
| `vpc_name` | `string` | The VPC name as created. |
| `cidr_block` | `string` | The primary IPv4 CIDR block of the VPC. Useful for downstream VSwitch CIDR planning. |
| `router_id` | `string` | The virtual router ID automatically created with the VPC. |
| `route_table_id` | `string` | The system route table ID associated with the VPC. |

## Related Components

- [AlicloudVswitch](/docs/catalog/alicloud/alicloudvswitch) — creates VSwitches (subnets) within this VPC
- [AlicloudSecurityGroup](/docs/catalog/alicloud/alicloudsecuritygroup) — creates security groups bound to this VPC
- [AlicloudNatGateway](/docs/catalog/alicloud/alicloudnatgateway) — creates NAT gateways for outbound internet access from private VSwitches
- [AlicloudApplicationLoadBalancer](/docs/catalog/alicloud/alicloudapplicationloadbalancer) — deploys Application Load Balancers in this VPC
- [AlicloudAckManagedCluster](/docs/catalog/alicloud/alicloudackmanagedcluster) — deploys managed Kubernetes clusters in this VPC

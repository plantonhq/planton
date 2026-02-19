---
title: "VPC"
description: "VPC deployment documentation"
icon: "package"
order: 100
componentName: "awsvpc"
---

# AWS VPC

Deploys an AWS Virtual Private Cloud with automatic subnet calculation, public and private subnet creation across specified Availability Zones, an Internet Gateway for public routing, and optional NAT Gateways for private subnet internet access.

## What Gets Created

When you deploy an AwsVpc resource, OpenMCF provisions:

- **VPC** — an `ec2.Vpc` resource with the specified CIDR block, DNS support, and DNS hostname settings
- **Internet Gateway** — an `ec2.InternetGateway` attached to the VPC for public internet access
- **Public Route Table** — an `ec2.RouteTable` with a default route (`0.0.0.0/0`) pointing to the Internet Gateway
- **Public Subnets** — one or more `ec2.Subnet` resources per Availability Zone with `MapPublicIpOnLaunch` enabled, each associated with the public route table via `ec2.RouteTableAssociation`
- **Private Subnets** — one or more `ec2.Subnet` resources per Availability Zone with `MapPublicIpOnLaunch` disabled
- **Elastic IPs** (NAT only) — one `ec2.Eip` per Availability Zone, created only when `isNatGatewayEnabled` is `true`
- **NAT Gateways** (NAT only) — one `ec2.NatGateway` per Availability Zone placed in the first public subnet, created only when `isNatGatewayEnabled` is `true`
- **Private Route Tables** (NAT only) — one `ec2.RouteTable` per private subnet routing `0.0.0.0/0` through the AZ's NAT Gateway, each associated via `ec2.RouteTableAssociation`

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A valid CIDR block** for the VPC (e.g., `10.0.0.0/16`)
- **At least one Availability Zone** in the target AWS region

## Quick Start

Create a file `vpc.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsVpc.my-vpc
spec:
  region: us-west-2
  vpcCidr: "10.0.0.0/16"
  subnetsPerAvailabilityZone: 1
  subnetSize: 24
```

Deploy:

```shell
openmcf apply -f vpc.yaml
```

This creates a VPC with the `10.0.0.0/16` CIDR block. No subnets are created until you also specify `availabilityZones`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the VPC will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `vpcCidr` | `string` | CIDR block for the VPC (e.g., `10.0.0.0/16`). Defines the full IP address range. | Must be a valid CIDR notation |
| `subnetsPerAvailabilityZone` | `int32` | Number of subnets to create in each Availability Zone. Recommended default: `1`. | Required |
| `subnetSize` | `int32` | Subnet mask size (e.g., `24` for `/24` subnets with 256 addresses). Must not be larger than the VPC CIDR mask. Recommended default: `1`. | Required; must be >= VPC CIDR mask size |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `availabilityZones` | `string[]` | `[]` | List of Availability Zones to span (e.g., `["us-west-2a", "us-west-2b"]`). Public and private subnets are created in each AZ. |
| `isNatGatewayEnabled` | `bool` | `false` | When `true`, creates one NAT Gateway per AZ in the first public subnet, with an Elastic IP and private route tables for outbound internet access from private subnets. |
| `isDnsHostnamesEnabled` | `bool` | `false` | When `true`, instances with public IPs receive corresponding public DNS hostnames. |
| `isDnsSupportEnabled` | `bool` | `false` | When `true`, enables DNS resolution through the Amazon-provided DNS server within the VPC. |

## Examples

### Two-AZ VPC with Public and Private Subnets

A typical VPC spanning two Availability Zones with one public and one private subnet per AZ:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: two-az-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsVpc.two-az-vpc
spec:
  region: us-east-1
  vpcCidr: "10.0.0.0/16"
  availabilityZones:
    - us-east-1a
    - us-east-1b
  subnetsPerAvailabilityZone: 1
  subnetSize: 24
```

### VPC with NAT Gateway

Private subnets that can reach the internet through NAT Gateways, with DNS support enabled:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: nat-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsVpc.nat-vpc
spec:
  region: us-west-2
  vpcCidr: "10.1.0.0/16"
  availabilityZones:
    - us-west-2a
    - us-west-2b
  subnetsPerAvailabilityZone: 1
  subnetSize: 24
  isNatGatewayEnabled: true
  isDnsSupportEnabled: true
  isDnsHostnamesEnabled: true
```

### Three-AZ Production VPC

A production VPC spanning three Availability Zones with multiple subnets per AZ, NAT Gateways, and full DNS support:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: prod-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsVpc.prod-vpc
spec:
  region: eu-west-1
  vpcCidr: "10.10.0.0/16"
  availabilityZones:
    - eu-west-1a
    - eu-west-1b
    - eu-west-1c
  subnetsPerAvailabilityZone: 2
  subnetSize: 24
  isNatGatewayEnabled: true
  isDnsSupportEnabled: true
  isDnsHostnamesEnabled: true
```

### Large VPC with Small Subnets

A VPC with a `/20` address space divided into `/28` subnets (16 addresses each) for fine-grained network segmentation:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: segmented-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsVpc.segmented-vpc
spec:
  region: ap-southeast-1
  vpcCidr: "172.16.0.0/20"
  availabilityZones:
    - ap-southeast-1a
    - ap-southeast-1b
  subnetsPerAvailabilityZone: 3
  subnetSize: 28
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vpc_id` | `string` | ID of the created VPC |
| `internet_gateway_id` | `string` | ID of the Internet Gateway attached to the VPC |
| `vpc_cidr` | `string` | CIDR block associated with the VPC |
| `public_subnets` | `AwsVpcSubnetStackOutputs[]` | List of public subnet details (see below) |
| `private_subnets` | `AwsVpcSubnetStackOutputs[]` | List of private subnet details (see below) |

Each subnet entry in `public_subnets` and `private_subnets` contains:

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Name of the subnet (e.g., `public-subnet-us-east-1a-0`) |
| `id` | `string` | AWS subnet ID |
| `cidr` | `string` | CIDR block of the subnet |
| `nat_gateway.id` | `string` | NAT Gateway ID (public subnets only, when NAT is enabled) |
| `nat_gateway.private_ip` | `string` | NAT Gateway private IP address |
| `nat_gateway.public_ip` | `string` | NAT Gateway public IP (Elastic IP) address |

## Related Components

- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network traffic for resources deployed in the VPC
- [AwsAlb](/docs/catalog/aws/alb) — deploys an Application Load Balancer in the VPC's subnets
- [AwsEksCluster](/docs/catalog/aws/eks-cluster) — deploys a Kubernetes cluster in the VPC
- [AwsClientVpn](/docs/catalog/aws/client-vpn) — provides VPN access into the VPC
- [AwsEc2Instance](/docs/catalog/aws/ec2-instance) — launches EC2 instances in the VPC's subnets

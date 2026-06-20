# AWS Subnet

Deploy a single subnet into an AWS VPC, with its routing folded in. A subnet is an Availability-Zone-scoped slice of a VPC's IP space and the placement target for EC2, load balancers, RDS, EKS/ECS, and most other AWS workloads.

## What Gets Created

- An **EC2 subnet** in the chosen VPC and Availability Zone, with the given IPv4 CIDR (and optional IPv6 CIDR).
- When `routes` is set: a dedicated **route table** owned by the subnet, populated with your rules, plus the **route-table association** linking it to the subnet.
- When `routeTableId` is set: a **route-table association** to that existing table.
- When neither is set: nothing extra — the subnet stays on the VPC main route table.

## Prerequisites

- An existing **AwsVpc** (or a literal vpc-id). The subnet's CIDR must fit within the VPC's CIDR.
- For internet/NAT routes, the corresponding gateway must already exist; supply its id as the route `targetId`.

## Quick Start

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSubnet
metadata:
  name: app-private-usw2a
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      fieldPath: status.outputs.vpc_id
  availabilityZone: us-west-2a
  cidrBlock: 10.0.1.0/24
```

## Configuration Reference

### Required

| Field | Description |
|---|---|
| `region` | AWS region (must match the VPC's region). |
| `vpcId` | The VPC to create the subnet in. Literal id or a reference to an `AwsVpc`. |
| `availabilityZone` | The AZ the subnet lives in (e.g. `us-west-2a`). |
| `cidrBlock` | IPv4 CIDR within the VPC (e.g. `10.0.1.0/24`). |

### Optional

| Field | Description |
|---|---|
| `mapPublicIpOnLaunch` | Auto-assign a public IPv4 to instances launched here. Default `false`. |
| `assignIpv6AddressOnCreation` | Auto-assign an IPv6 address. Requires `ipv6CidrBlock`. Default `false`. |
| `ipv6CidrBlock` | IPv6 /64 for a dual-stack subnet. |
| `enableDns64` | Enable DNS64 (NAT64) for IPv6-only egress to IPv4. Default `false`. |
| `enableResourceNameDnsARecordOnLaunch` | DNS A record for instance resource names. Default `false`. |
| `enableResourceNameDnsAaaaRecordOnLaunch` | DNS AAAA record for instance resource names. Default `false`. |
| `privateDnsHostnameTypeOnLaunch` | `ip-name` or `resource-name`. |
| `routeTableId` | Associate an existing route table. Mutually exclusive with `routes`. |
| `routes` | Inline rules; creates a subnet-owned route table. Mutually exclusive with `routeTableId`. |

### Route (within `routes`)

| Field | Description |
|---|---|
| `destinationCidrBlock` / `destinationIpv6CidrBlock` / `destinationPrefixListId` | The destination — set exactly one. |
| `targetType` | `internet_gateway`, `nat_gateway`, `transit_gateway`, `vpc_peering_connection`, `vpc_endpoint`, `network_interface`, or `egress_only_internet_gateway`. |
| `targetId` | The target resource id (literal or reference). |

## Examples

Public subnet (default route to an internet gateway):

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSubnet
metadata:
  name: public-usw2a
spec:
  region: us-west-2
  vpcId:
    value: vpc-0abc123
  availabilityZone: us-west-2a
  cidrBlock: 10.0.0.0/24
  mapPublicIpOnLaunch: true
  routes:
    - destinationCidrBlock: 0.0.0.0/0
      targetType: internet_gateway
      targetId:
        value: igw-0abc123
```

Private subnet (outbound via NAT gateway):

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSubnet
metadata:
  name: private-usw2a
spec:
  region: us-west-2
  vpcId:
    value: vpc-0abc123
  availabilityZone: us-west-2a
  cidrBlock: 10.0.1.0/24
  routes:
    - destinationCidrBlock: 0.0.0.0/0
      targetType: nat_gateway
      targetId:
        value: nat-0abc123
```

## Stack Outputs

| Output | Description |
|---|---|
| `subnet_id` | The subnet's id. |
| `subnet_arn` | The subnet's ARN. |
| `availability_zone` | The AZ the subnet resides in. |
| `cidr_block` | The subnet's IPv4 CIDR. |
| `route_table_id` | The associated route table (inline-created, external, or empty for the VPC main table). |
| `region` | The region the subnet was created in. |

## Related Components

- **AwsVpc** — the network the subnet belongs to.
- **AwsElasticIp** — a stable public IP, e.g. for a NAT gateway a subnet routes to.
- **AwsTransitGateway** — a route target for inter-VPC / hybrid connectivity.

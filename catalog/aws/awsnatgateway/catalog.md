# AWS NAT Gateway

Gives instances in a private subnet outbound network access while keeping them unreachable from inbound connections -- the standard way to let private workloads reach the internet (or other private networks) without exposing them. A NAT gateway is an AWS-managed, highly-available service referenced by subnets' route tables as the target of their default route: classically one gateway per subnet/AZ (`zonal`), or -- since re:Invent 2025 -- one regional gateway spanning every AZ of a VPC. It is a first-class, independently composable building block rather than something bundled inside the VPC: create it as its own graph node, place it exactly where you intend, and reference its id from the subnets that should egress through it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **NAT gateway** -- an EC2 NAT gateway, either `public` (fronted by Elastic IPs) or `private` (no Elastic IP), placed zonally in one subnet or regionally across a whole VPC (`availabilityMode: regional`)
- **Elastic IP association** -- for a zonal public gateway, the referenced Elastic IP (`allocationId`) is bound as the gateway's stable outbound address; secondary allocations are attached when provided. A regional gateway carries its per-zone allocations in `availabilityZoneAddresses`, or lets AWS allocate and manage them (auto mode)
- **AWS Tags** -- resource-identity tags (organization, environment, resource kind, resource ID) applied to the gateway

Creating a NAT gateway does not route anything on its own. To build a working egress topology, compose the companion components listed under [Works With](#works-with).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A subnet** -- the gateway is created inside a subnet. Deploy an [AWS Subnet](/infra-catalog/aws-subnet) first, or reference an existing one by id. A public gateway needs a *public* subnet.
- **An Elastic IP** (public gateways) -- a public gateway requires an [AWS Elastic IP](/infra-catalog/aws-elastic-ip) for its outbound address. Deploy one first, or reference an existing `eipalloc-` id.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Public gateways need an internet path** -- a public NAT gateway must live in a public subnet whose route table already reaches an [AWS Internet Gateway](/infra-catalog/aws-internet-gateway) (regional gateways likewise need the IGW attached to the VPC), or it has no upstream internet path itself.
- **Zonal vs regional** -- a zonal gateway lives in one Availability Zone; for high availability run one per AZ, or deploy a single `regional` gateway and let AWS span every zone for you.
- **Region** -- the gateway is created in the specified `region`, which must match the subnet's (or VPC's) region.
- **Regional IAM** -- creating a regional gateway additionally requires the `ec2:DescribeAvailabilityZones` permission on the deploying role.

## Deploy

### Console

Open the deployment store, find **AWS NAT Gateway**, and click **Deploy**. The creation wizard walks two steps: **Placement** (public vs private connectivity, zonal vs regional mode, region, and the subnet or VPC the gateway spans) and **Addressing** (the Elastic IP for a zonal public gateway, private addressing for a private one, or the per-zone layout for a regional one). Start from a preset in the [Presets](#presets) tab -- **Public NAT Gateway (greenfield)**, **Private NAT Gateway**, or **Regional NAT Gateway**.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsNatGateway
metadata:
  name: prod-egress-nat
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  connectivityType: public
  subnetId:
    valueFrom:
      kind: AwsSubnet
      name: public-subnet-a
      fieldPath: status.outputs.subnet_id
  allocationId:
    valueFrom:
      kind: AwsElasticIp
      name: nat-eip-a
      fieldPath: status.outputs.allocation_id
```

```shell
planton apply -f nat-gateway.yaml
```

This creates a public NAT gateway in a Planton-managed public subnet, fronted by a referenced Elastic IP. An Infra Job tracks the provisioning in real time.

### InfraChart

When the gateway deploys alongside its subnet and Elastic IP in one chart, wire both references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  connectivityType: public
  subnetId:
    valueFrom:
      kind: AwsSubnet
      name: public-subnet-a
      fieldPath: status.outputs.subnet_id
  allocationId:
    valueFrom:
      kind: AwsElasticIp
      name: nat-eip-a
      fieldPath: status.outputs.allocation_id
```

The InfraPipeline resolves the dependency graph, deploys the subnet and Elastic IP first, then creates the NAT gateway on top of them.

## Key Configuration

A NAT gateway's value is in how it composes; most deployments touch only a few fields. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Connectivity type** -- `connectivityType` is required and immutable: `public` (Elastic IP, internet egress) or `private` (no Elastic IP, egress to peered/transit/VPN networks only). Changing it replaces the gateway.

**Placement mode** -- `availabilityMode` picks the placement model: `zonal` (default) builds the classic gateway inside one subnet (`subnetId`, required and immutable there); `regional` builds one AWS-managed gateway spanning every AZ of the referenced `vpcId`, either letting AWS pick zones (auto mode) or pinning them via `availabilityZoneAddresses` (switching a live gateway between auto and manual replaces it). `region` must match the subnet's (or VPC's) region.

**Elastic IP (public)** -- for a zonal public gateway `allocationId` is required and references an `AwsElasticIp` (`status.outputs.allocation_id`) rather than embedding an address, so the IP keeps its own lifecycle. Add `secondaryAllocationIds` only for very high-throughput egress that exhausts a single EIP's source ports. A regional public gateway carries per-zone allocations inside `availabilityZoneAddresses` instead -- or none at all in auto mode.

**Private addressing (private)** -- a private gateway can pin a fixed `privateIp` from the subnet's range (or let AWS choose), and add secondary private IPs either as an explicit list (`secondaryPrivateIpAddresses`) or an auto-assign count (`secondaryPrivateIpAddressCount`) -- the two are mutually exclusive.

**Placement is not routing** -- creating a gateway does nothing by itself. A private subnet gains outbound access only when its route table sends a default route (`0.0.0.0/0`) to this gateway. The egress recipe is this gateway plus an `AwsSubnet` whose route targets it (`targetType: nat_gateway`).

> **Cost note:** NAT gateways bill both an hourly charge per gateway and a per-GB data-processing charge on all traffic through them. High-volume egress -- and chatty cross-AZ traffic routed through a single-AZ gateway -- is a common surprise on an AWS bill. Traffic to S3 and DynamoDB can bypass the NAT gateway via VPC gateway endpoints.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetId` (zonal mode) | `status.outputs.subnet_id` |
| **AwsVpc** | `vpcId` (regional mode) | `status.outputs.vpc_id` |
| **AwsElasticIp** | `allocationId` | `status.outputs.allocation_id` |
| **AwsElasticIp** | `secondaryAllocationIds` | `status.outputs.allocation_id` |
| **AwsElasticIp** | `availabilityZoneAddresses[].allocationIds` (regional mode) | `status.outputs.allocation_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `nat_gateway_id` | ID of the created gateway | An `AwsSubnet` route's `targetId` (with `targetType: nat_gateway`) to give the subnet outbound egress |
| `public_ip` | The public IPv4 of a public gateway (the Elastic IP's address) | Allowlisting the egress source on third-party services |
| `private_ip` | The private IPv4 assigned within the subnet | Internal routing and diagnostics |
| `network_interface_id` | ID of the gateway's elastic network interface | Flow logs, troubleshooting |
| `subnet_id` | ID of the subnet the gateway lives in | Confirming placement |
| `region` | Region the gateway was created in | Downstream region wiring |

A NAT gateway has no ARN -- AWS exposes none, so none is published.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private-subnet internet egress** -- the canonical pattern: a public NAT gateway in a public subnet, with private subnets routing `0.0.0.0/0` to its `nat_gateway_id`. Start from the **Public NAT Gateway (greenfield)** preset.

**High-availability egress** -- a single regional gateway (start from the **Regional NAT Gateway** preset) that spans every AZ with zone-local egress; or, in the classic form, one zonal NAT gateway per Availability Zone, each in that zone's public subnet, so a zonal failure does not sever egress and cross-AZ data-processing charges are avoided.

**Private inter-network egress** -- a private NAT gateway for outbound communication to other VPCs or on-premises networks (via peering/Transit Gateway/VPN) without any internet exposure. Start from the **Private NAT Gateway** preset.

## Works With

A NAT gateway sits between the private subnets that need egress and the internet path a public subnet provides:

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- the gateway lives in one subnet (public for a public gateway), and other subnets route a default route to this gateway's `nat_gateway_id` to gain egress
- [**AWS Elastic IP**](/infra-catalog/aws-elastic-ip) -- a public gateway's stable outbound address, referenced by `status.outputs.allocation_id`
- [**AWS Internet Gateway**](/infra-catalog/aws-internet-gateway) -- the internet path the public subnet (and the NAT gateway in it) routes through
- [**AWS VPC**](/infra-catalog/aws-vpc) -- the network that contains the subnet, the gateway, and the routes that tie them together
- [**AWS Egress-Only Internet Gateway**](/infra-catalog/aws-egress-only-internet-gateway) -- the IPv6 outbound-only counterpart for dual-stack VPCs

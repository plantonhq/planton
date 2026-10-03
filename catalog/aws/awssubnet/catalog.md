# AWS Subnet

Deploys a subnet inside an AWS VPC -- a contiguous range of IP addresses pinned to a single availability zone. A subnet is not inherently "public" or "private": that identity comes entirely from its route table. This component folds routing into the subnet, so you declare a subnet's intent (public, private-with-egress, or isolated) in one place, referencing the VPC and any gateways as first-class building blocks.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Subnet** -- the IP range in the chosen VPC and availability zone: an explicit IPv4 CIDR or an IPAM-pool allocation, an optional IPv6 `/64` (explicit or IPAM), or an IPv6-only subnet (`ipv6Native`)
- **Launch behaviour** -- map-public-IP-on-launch, IPv6 auto-assignment, DNS64, and resource-name DNS A/AAAA records, applied to instances started in the subnet
- **Dedicated route table** -- created when inline `routes` and/or `propagatingVgws` are supplied; the rules (and any VGW route propagation) are written into a table owned and associated by this subnet
- **Route table association** -- created when `routeTableId` references an existing table; otherwise the subnet uses the VPC's main route table
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the subnet

To build a working network, compose the companion components listed under [Works With](#works-with).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A VPC** -- the subnet must reference a VPC. Deploy an [AWS VPC](/infra-catalog/aws-vpc) first, or reference an existing one by id.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Region and zone** -- the subnet is created in the specified `region` and `availabilityZone`, and both are permanent. Spanning zones means one subnet per zone.
- **CIDR planning** -- the IPv4 CIDR must fall within the VPC's range and not overlap any sibling subnet. AWS reserves 5 addresses per subnet, so a `/24` yields 251 usable.
- **Gateways for egress** -- a public subnet needs an internet gateway attached to the VPC; a private subnet's outbound access needs a NAT gateway (IPv4) or an egress-only internet gateway (IPv6).

## Deploy

### Console

Open the deployment store, find **AWS Subnet**, and click **Deploy**. The creation wizard walks you through placement (region, VPC, zone), addressing (IPv4 and optional IPv6), DNS and launch options, and routing. Start from a preset in the [Presets](#presets) tab -- **Public Subnet**, **Private Subnet**, or **Isolated Subnet** -- to pre-populate a typical configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSubnet
metadata:
  name: public-usw2a
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
  availabilityZone: us-west-2a
  cidrBlock: "10.0.0.0/24"
  mapPublicIpOnLaunch: true
  routes:
    - destinationCidrBlock: "0.0.0.0/0"
      targetType: internet_gateway
      targetId:
        valueFrom:
          kind: AwsInternetGateway
          name: production-igw
          fieldPath: status.outputs.internet_gateway_id
```

```shell
planton apply -f subnet.yaml
```

This creates a public subnet whose dedicated route table sends all IPv4 traffic to an internet gateway, both wired by reference to other Planton resources. An Infra Job tracks the provisioning in real time.

### InfraChart

When the subnet deploys alongside its VPC and gateways in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
  availabilityZone: us-west-2a
  cidrBlock: "10.0.0.0/24"
  mapPublicIpOnLaunch: true
  routes:
    - destinationCidrBlock: "0.0.0.0/0"
      targetType: internet_gateway
      targetId:
        valueFrom:
          kind: AwsInternetGateway
          name: production-igw
          fieldPath: status.outputs.internet_gateway_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC and internet gateway first, then creates the subnet and its route table on top of them.

## Key Configuration

These are the most important decisions when configuring a subnet. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Placement** -- `region`, `vpcId`, and the availability zone define the subnet and are immutable. Name the zone (`availabilityZone: us-west-2a`) or pin it by account-stable id (`availabilityZoneId: usw2-az1`) when coordinating layouts across AWS accounts, whose zone NAMES are shuffled per account. Reference the VPC (`valueFrom`) rather than pasting a literal id so the dependency is an explicit edge in your InfraChart graph.

**IPv4 addressing** -- declare the `cidrBlock` yourself (immutable; a `/24` per application tier is a good default, within the VPC range, non-overlapping), or allocate it from an IPAM pool (`ipv4IpamPoolId` + `ipv4NetmaskLength`, /16-/28) so your address plan owns the numbers. Exactly one method -- unless the subnet is IPv6-only.

**IPv6 and IPv6-only** -- set `ipv6CidrBlock` to a `/64` from the VPC's IPv6 range (or allocate via `ipv6IpamPoolId` + `ipv6NetmaskLength`) to make the subnet dual-stack, then optionally enable IPv6 auto-assignment and DNS64. Set `ipv6Native: true` for an IPv6-ONLY subnet -- no IPv4 at all; pair it with `privateDnsHostnameTypeOnLaunch: resource-name` (required for IPv6-only launches) and DNS64/NAT64 for reaching IPv4-only destinations.

**Routing decides identity** -- choose one: inherit the VPC main route table (isolated/default), associate an existing `routeTableId`, or supply inline `routes` (and/or `propagatingVgws`, which pulls Site-to-Site VPN / Direct Connect routes in from a virtual private gateway automatically). A default route (`0.0.0.0/0`) to an `internet_gateway` makes the subnet public; one to a `nat_gateway` makes it private with egress; an `::/0` route to an `egress_only_internet_gateway` gives private IPv6 egress. Route targets cover the provider's full set -- including `carrier_gateway` (Wavelength), `core_network` (Cloud WAN, by ARN), `local_gateway` (Outposts), and `odb_network` (Oracle Database@AWS, by ARN).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |
| **AwsInternetGateway** | `routes[].targetId` (`targetType: internet_gateway`) | `status.outputs.internet_gateway_id` |
| **AwsNatGateway** | `routes[].targetId` (`targetType: nat_gateway`) | `status.outputs.nat_gateway_id` |
| **AwsEgressOnlyInternetGateway** | `routes[].targetId` (`targetType: egress_only_internet_gateway`) | `status.outputs.egress_only_internet_gateway_id` |

`routeTableId` and the IPAM pool fields (`ipv4IpamPoolId`, `ipv6IpamPoolId`) take literal IDs -- no catalog kind produces them.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `subnet_id` | ID of the created subnet | EKS node groups, RDS subnet groups, load balancers, NAT gateways, ENIs |
| `subnet_arn` | ARN of the subnet | IAM policies, resource sharing |
| `availability_zone` | The zone the subnet lives in | Zone-aware placement of dependent resources |
| `cidr_block` | The subnet's IPv4 CIDR | Security group rules, network ACLs |
| `route_table_id` | ID of the route table the subnet owns (inline `routes`) or references (`routeTableId`); empty when the subnet rides the VPC main table | Adding routes, inspection |
| `region` | Region the subnet was created in | Downstream region wiring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public subnet** -- a default route to an internet gateway with public-IP-on-launch; home for load balancers, bastions, and NAT gateways. Start from the **Public Subnet** preset.

**Private subnet** -- a default route to a NAT gateway for outbound-only access; home for application and worker tiers. Start from the **Private Subnet** preset.

**Isolated subnet** -- no internet route at all; home for data tiers that should never reach or be reached from the internet. Start from the **Isolated Subnet** preset.

## Works With

A subnet sits between the VPC and the gateways that give it reachability:

- [**AWS VPC**](/infra-catalog/aws-vpc) -- the network the subnet draws its range from, referenced by `status.outputs.vpc_id`
- [**AWS Internet Gateway**](/infra-catalog/aws-internet-gateway) -- the default-route target that makes a subnet public
- [**AWS NAT Gateway**](/infra-catalog/aws-nat-gateway) -- the default-route target that gives a private subnet outbound IPv4 access
- [**AWS Egress-Only Internet Gateway**](/infra-catalog/aws-egress-only-internet-gateway) -- the `::/0` target for outbound-only IPv6 from a private dual-stack subnet

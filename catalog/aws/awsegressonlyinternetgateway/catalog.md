# AWS Egress-Only Internet Gateway

Attaches an egress-only internet gateway to an AWS VPC -- the IPv6 counterpart of a NAT gateway. It is a horizontally scaled, redundant, AWS-managed component that lets dual-stack instances make **outbound** IPv6 connections to the internet while AWS statefully blocks any unsolicited **inbound** ones. It is a first-class, independently composable building block: create it as its own graph node, attach it to exactly the VPC you intend, and reference its id from the subnets that should have IPv6 egress.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Egress-only internet gateway** -- an EC2 egress-only internet gateway, attached to the specified VPC at creation
- **VPC attachment** -- the gateway is attached to the VPC referenced by `vpcId`; because AWS exposes no detach/re-attach API, changing `vpcId` on a later apply **replaces** the gateway (ForceNew)
- **AWS Tags** -- resource-identity tags (organization, environment, resource kind, resource ID) applied to the gateway

Attaching a gateway does not route anything on its own. To give a subnet outbound IPv6, compose the companion components listed under [Works With](#works-with).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A dual-stack VPC** -- the gateway must attach to a VPC, and the VPC should have an IPv6 CIDR for the gateway to be useful. Deploy an [AWS VPC](/infra-catalog/aws-vpc) with IPv6 first, or reference an existing one by id.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **IPv6 only** -- an egress-only internet gateway carries IPv6 traffic only. Every IPv6 address AWS assigns is globally routable, so there is no IPv6 NAT; the egress-only gateway is what provides the outbound-but-not-inbound guarantee for IPv6.
- **Region** -- the gateway is created in the specified `region`, which must match the VPC's region.

## Deploy

### Console

Open the deployment store, find **AWS Egress-Only Internet Gateway**, and click **Deploy**. The creation wizard asks for the one decision the gateway has: the VPC to attach to and its region. Start from a preset in the [Presets](#presets) tab -- **IPv6 Egress (greenfield)** to attach by reference, or **Attach to an Existing VPC (brownfield)** to attach by literal id.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEgressOnlyInternetGateway
metadata:
  name: main-eigw
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
```

```shell
planton apply -f egress-only-internet-gateway.yaml
```

This attaches an egress-only internet gateway to a Planton-managed VPC by reference. An Infra Job tracks the provisioning in real time.

### InfraChart

When the gateway deploys alongside its VPC in one chart, wire the VPC reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC first, then attaches the gateway to it.

## Key Configuration

An egress-only internet gateway has a deliberately small surface -- the value is in how it composes, not in tuning knobs. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Region** -- `region` is required and must match the VPC's region; a gateway cannot span regions.

**VPC attachment** -- `vpcId` is required. Reference the VPC (`valueFrom`) rather than pasting a literal id so the dependency is an explicit edge in your InfraChart graph and both resources deploy in the same plan. The attachment is **immutable**: AWS has no detach/re-attach API for an egress-only gateway, so changing `vpcId` replaces the gateway.

**Attachment is not routing** -- attaching a gateway does nothing by itself. A subnet gains outbound IPv6 only when its route table sends the IPv6 default route (`::/0`) to this gateway. The recipe is this gateway plus an `AwsSubnet` whose route targets it (`targetType: egress_only_internet_gateway`).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `egress_only_internet_gateway_id` | ID of the created gateway | An `AwsSubnet` IPv6 route's `targetId` (with `targetType: egress_only_internet_gateway`) to give the subnet outbound IPv6 |
| `vpc_id` | ID of the VPC the gateway is attached to | Confirming or wiring the attachment |
| `region` | Region the gateway was created in | Downstream region wiring |

AWS exposes no ARN for an egress-only internet gateway.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dual-stack private subnets** -- attach this gateway, then give private subnets an `::/0` IPv6 route to its `egress_only_internet_gateway_id` so their IPv6 workloads can reach the internet outbound. Start from the **IPv6 Egress (greenfield)** preset.

**Cost-sensitive IPv6 egress** -- NAT gateway cost scales with hours provisioned and gigabytes processed; moving traffic that can run IPv6 end to end onto the egress-only gateway takes it off that meter.

**Brownfield attachment** -- attach a Planton-managed gateway to a dual-stack VPC created outside Planton by supplying its literal vpc-id. Start from the **Attach to an Existing VPC (brownfield)** preset.

## Works With

An egress-only internet gateway sits between a dual-stack VPC and the private subnets that need outbound IPv6:

- [**AWS VPC**](/infra-catalog/aws-vpc) -- the dual-stack network the gateway attaches to, referenced by `status.outputs.vpc_id`
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- routes an IPv6 default route (`::/0`) to this gateway's `egress_only_internet_gateway_id` for outbound IPv6
- [**AWS NAT Gateway**](/infra-catalog/aws-nat-gateway) -- the IPv4 equivalent: outbound-only access for private IPv4 subnets, with cost driven by hours and traffic volume
- [**AWS Internet Gateway**](/infra-catalog/aws-internet-gateway) -- full bidirectional internet access for a VPC, the counterpart when inbound reachability is wanted

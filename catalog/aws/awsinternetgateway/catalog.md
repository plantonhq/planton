# AWS Internet Gateway

Attaches an internet gateway to an AWS VPC -- the VPC's door to the public internet. An internet gateway is a horizontally scaled, redundant, AWS-managed component that allows bidirectional IPv4 (and, for dual-stack VPCs, IPv6) traffic between the VPC and the internet. It is a first-class, independently composable building block rather than something bundled inside the VPC: create it as its own graph node, attach it to exactly the VPC you intend, and reference its id from the subnets that should be public.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Internet gateway** -- an EC2 internet gateway, attached to the specified VPC
- **VPC attachment** -- the gateway is attached to the VPC referenced by `vpcId`; changing `vpcId` on a later apply re-attaches the gateway to the new VPC rather than recreating it
- **AWS Tags** -- resource-identity tags (organization, environment, resource kind, resource ID) applied to the gateway

Attaching a gateway does not expose anything on its own. To build a working public network, compose the companion components listed under [Works With](#works-with).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A VPC** -- the gateway must attach to a VPC. Deploy an [AWS VPC](/infra-catalog/aws-vpc) first, or reference an existing one by id.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **One gateway per VPC** -- AWS allows at most one internet gateway attached to a VPC at a time. To add public connectivity you reuse this gateway, not create a second.
- **Region** -- the gateway is created in the specified `region`, which must match the VPC's region.

## Deploy

### Console

Open the deployment store, find **AWS Internet Gateway**, and click **Deploy**. The creation wizard asks for the one decision an internet gateway has: the VPC to attach to and its region. Start from a preset in the [Presets](#presets) tab -- **Public Internet Gateway (greenfield)** to attach by reference, or **Attach to an Existing VPC (brownfield)** to attach by literal id.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsInternetGateway
metadata:
  name: main-igw
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
planton apply -f internet-gateway.yaml
```

This attaches an internet gateway to a Planton-managed VPC by reference. An Infra Job tracks the provisioning in real time.

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

An internet gateway has a deliberately small surface -- the value is in how it composes, not in tuning knobs. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Region** -- `region` is required and must match the VPC's region; a gateway cannot span regions.

**VPC attachment** -- `vpcId` is required. Reference the VPC (`valueFrom`) rather than pasting a literal id so the dependency is an explicit edge in your InfraChart graph and both resources deploy in the same plan. The attachment is **updatable**: pointing `vpcId` at a different VPC detaches the gateway from the old one and attaches it to the new one without replacing the gateway.

**Attachment is not exposure** -- attaching a gateway does nothing by itself. A subnet becomes public only when its route table sends a default route (`0.0.0.0/0`, or `::/0` for IPv6) to this gateway. The public-subnet recipe is this gateway plus an `AwsSubnet` whose route targets it (`targetType: internet_gateway`).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `internet_gateway_id` | ID of the created gateway | An `AwsSubnet` route's `targetId` (with `targetType: internet_gateway`) to make the subnet public |
| `internet_gateway_arn` | ARN of the gateway | IAM policies, resource sharing |
| `vpc_id` | ID of the VPC the gateway is attached to | Confirming or wiring the attachment |
| `region` | Region the gateway was created in | Downstream region wiring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public-facing VPC** -- attach this gateway, then give public subnets (load balancers, bastions, internet-facing services) a `0.0.0.0/0` route to its `internet_gateway_id`. Start from the **Public Internet Gateway (greenfield)** preset.

**NAT egress topology** -- the internet path that public subnets, and the NAT gateways living in them, route through. Private subnets reach the internet outbound by routing to a NAT gateway that itself sits in a public subnet routed here.

**Brownfield attachment** -- attach a Planton-managed gateway to a VPC created outside Planton by supplying its literal vpc-id. Start from the **Attach to an Existing VPC (brownfield)** preset.

## Works With

An internet gateway sits between a VPC and the subnets that need internet reachability:

- [**AWS VPC**](/infra-catalog/aws-vpc) -- the network the gateway attaches to, referenced by `status.outputs.vpc_id`
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- routes a default route to this gateway's `internet_gateway_id` to become public
- [**AWS NAT Gateway**](/infra-catalog/aws-nat-gateway) -- lives in a public subnet that routes here, giving private subnets outbound IPv4 access
- [**AWS Egress-Only Internet Gateway**](/infra-catalog/aws-egress-only-internet-gateway) -- the IPv6 outbound-only counterpart for dual-stack VPCs that need no inbound exposure

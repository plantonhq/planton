# AWS Transit Gateway VPC Attachment

Attaches a VPC to an AWS Transit Gateway — the connection that plugs one spoke VPC into the regional hub. An attachment is deliberately its own resource rather than a field on the gateway: a gateway carries many attachments, each with its own lifecycle, and the attachment ID this resource outputs is what Transit Gateway route tables associate, propagate, and route against. The gateway and VPC pair is create-time immutable -- changing either replaces the attachment and issues a new attachment ID -- while the subnet set updates in place, so Availability Zones can be added or removed without replacement.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Transit Gateway VPC Attachment** -- the link between the gateway and the VPC, with per-attachment DNS, IPv6, appliance-mode, and security-group-referencing options plus default-route-table membership dials
- **Elastic Network Interfaces** -- one per chosen subnet, provisioned by AWS; traffic between the VPC and the gateway flows through these ENIs
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A Transit Gateway** in the same region -- deployed through Planton (reference its `transit_gateway_id` output) or created elsewhere, including a gateway shared into this account via AWS RAM (paste its literal ID).
- **A VPC with subnets** in the same region. Choose one subnet per Availability Zone you want reachable through the gateway -- the AWS-recommended pattern is a dedicated /28 subnet per AZ just for attachments.
- **Non-overlapping CIDRs** with other attached VPCs if routes will propagate into shared route tables.

## Deploy

### Console

Open the deployment store, find **AWS Transit Gateway VPC Attachment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Mesh Spoke** preset in the [Presets](#presets) tab for the inherit-everything shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGatewayVpcAttachment
metadata:
  name: app-vpc-attachment
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  transitGatewayId:
    value: "tgw-0123456789abcdef0"
  vpcId:
    value: "vpc-0a1b2c3d4e5f00001"
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
```

```shell
planton apply -f attachment.yaml
```

This attaches the VPC through two Availability Zones with every behavior dial left unset — the attachment inherits all of the gateway's settings. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the attachment to the gateway and VPC deployed in the same InfraPipeline:

```yaml
spec:
  transitGatewayId:
    valueFrom:
      kind: AwsTransitGateway
      name: network-hub
      fieldPath: status.outputs.transit_gateway_id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: app-vpc
      fieldPath: status.outputs.vpc_id
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: tgw-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: tgw-subnet-b
        fieldPath: status.outputs.subnet_id
```

The InfraPipeline resolves the dependency graph, deploys the gateway, VPC, and subnets first, then provisions the attachment with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring an attachment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The immutable pair** -- `transitGatewayId` and `vpcId` are create-time immutable: changing either replaces the attachment and issues a NEW attachment ID, breaking route table associations built on the old one. The subnet set updates in place.

**Subnet coverage** -- AWS provisions one ENI per subnet, at most one per Availability Zone, all in the attached VPC. The gateway only routes to/from AZs it has an ENI in, so cover every AZ your workloads run in.

**Inherit-the-gateway dials** -- `dnsSupport`, `securityGroupReferencingSupport`, `defaultRouteTableAssociation`, and `defaultRouteTablePropagation` are three-position dials: left unset, the attachment inherits the gateway's setting; pin a position only when this one attachment must diverge. Association is exclusive (an attachment belongs to at most one route table — a default association and a custom-table association conflict at the AWS API); propagation is additive (many tables at once).

**Appliance mode** -- Enable only on the shared-services VPC hosting stateful inspection appliances (firewall, IDS/IPS). It pins a flow's return traffic to the same AZ, preserving symmetric routing.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsTransitGateway** | `transitGatewayId` | `status.outputs.transit_gateway_id` |
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |
| **AwsSubnet** | `subnetIds[]` | `status.outputs.subnet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `attachment_id` | The Transit Gateway attachment ID | The join key of the routing surface: route table associations, propagations, and static route targets |
| `attachment_arn` | ARN of the attachment | IAM policies, resource-level permissions |
| `vpc_owner_id` | AWS account ID owning the attached VPC | Cross-account (RAM-shared) topology verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Full-mesh spoke** -- Join a gateway whose default dials are on: leave every dial unset and the VPC immediately reaches every other attached VPC. Start from the **Mesh Spoke** preset.

**Segmented spoke** -- On a segmented gateway, leave the dials unset (the gateway's dials are off) and let custom `AwsTransitGatewayRouteTable` resources claim this attachment by its `attachment_id` output. Start from the **Segmented Spoke** preset.

**Inspection VPC** -- Enable appliance mode on the attachment of the shared-services VPC hosting the firewall; leave it off on every spoke. Start from the **Inspection VPC Attachment** preset.

## Works With

- [**AWS Transit Gateway**](/infra-catalog/aws-transit-gateway) -- the hub this attachment joins; provides `transit_gateway_id`
- [**AWS Transit Gateway Route Table**](/infra-catalog/aws-transit-gateway-route-table) -- associates and propagates this attachment by its `attachment_id`
- [**AWS VPC**](/infra-catalog/aws-vpc) -- the network being attached; provides `vpc_id` and subnets

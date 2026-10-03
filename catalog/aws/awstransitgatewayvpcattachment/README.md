# AWS Transit Gateway VPC Attachment

A Transit Gateway VPC attachment connects one VPC to a Transit Gateway hub. AWS provisions an elastic network interface in each chosen subnet (one per Availability Zone); traffic between the VPC and the gateway flows through those ENIs.

The attachment is its own resource because it is the unit the Transit Gateway routing surface works with: `AwsTransitGatewayRouteTable` resources associate it, accept propagations from it, and target it in static routes -- and a gateway carries one attachment per spoke VPC, each with its own lifecycle.

## When to Use

- Every VPC that should participate in a Transit Gateway topology gets exactly one attachment per gateway.
- Set `applianceModeSupport: true` only on the attachment of a shared-services VPC that hosts stateful inspection appliances (firewall, IDS/IPS).

## Prerequisites

- An `AwsTransitGateway` (or a literal gateway ID from an existing hub).
- An `AwsVpc` with one subnet per Availability Zone you want reachable through the gateway. A dedicated small subnet (/28) per AZ for the attachment is the AWS-recommended pattern.

## Spec Fields

| Field | Type | Default | Description |
|---|---|---|---|
| `region` | string | (required) | AWS region; must match the gateway and the VPC |
| `transitGatewayId` | ref | (required) | The gateway to attach to (create-time immutable). References `AwsTransitGateway.status.outputs.transit_gateway_id` |
| `vpcId` | ref | (required) | The VPC to attach (create-time immutable). References `AwsVpc.status.outputs.vpc_id` |
| `subnetIds` | ref[] | (required, min 1) | One subnet per AZ; updatable in place. References `AwsSubnet.status.outputs.subnet_id` |
| `dnsSupport` | optional bool | true | DNS resolution for this attachment (overrides the gateway dial) |
| `ipv6Support` | bool | false | IPv6 traffic over this attachment (requires IPv6 CIDRs on the VPC/subnets) |
| `applianceModeSupport` | bool | false | AZ-symmetric flows for stateful inspection VPCs |
| `securityGroupReferencingSupport` | optional bool | gateway-inherited | Cross-VPC SG referencing override for this attachment |
| `defaultRouteTableAssociation` | optional bool | gateway-inherited | Associate with the gateway's default route table. Set `false` when a custom `AwsTransitGatewayRouteTable` owns the association -- an attachment can be associated with at most ONE table |
| `defaultRouteTablePropagation` | optional bool | gateway-inherited | Propagate this VPC's CIDRs into the default route table |

## Outputs

| Output | Description |
|---|---|
| `attachment_id` | The join key referenced by route table associations, propagations, and static routes |
| `attachment_arn` | ARN for IAM policies |
| `vpc_owner_id` | Account owning the attached VPC |

## Composition

- References an `AwsTransitGateway` and an `AwsVpc` + `AwsSubnet`s.
- Referenced by `AwsTransitGatewayRouteTable.associations` / `.propagations` / `.routes[].attachmentId`.
- Remember the return path: each spoke VPC's subnets need a route with `targetType: transit_gateway` toward the gateway (on `AwsSubnet.routes`) for traffic to flow both ways.

## Example: spoke of a full-mesh hub

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGatewayVpcAttachment
metadata:
  name: app-vpc-attachment
spec:
  region: us-east-1
  transitGatewayId:
    valueFrom:
      kind: AwsTransitGateway
      name: core-hub
      fieldPath: status.outputs.transit_gateway_id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: app-vpc
      fieldPath: status.outputs.vpc_id
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: app-tgw-subnet-az1
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: app-tgw-subnet-az2
        fieldPath: status.outputs.subnet_id
```

## Example: segmented-topology spoke (custom routing domain)

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsTransitGatewayVpcAttachment
metadata:
  name: prod-vpc-attachment
spec:
  region: us-east-1
  transitGatewayId:
    valueFrom:
      kind: AwsTransitGateway
      name: segmented-hub
      fieldPath: status.outputs.transit_gateway_id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: prod-vpc
      fieldPath: status.outputs.vpc_id
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: prod-tgw-subnet-az1
        fieldPath: status.outputs.subnet_id
  defaultRouteTableAssociation: false
  defaultRouteTablePropagation: false
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
